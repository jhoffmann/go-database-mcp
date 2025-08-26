// Package main provides the entry point for the Database MCP Server.
// It implements a Model Context Protocol server that provides database connectivity
// for MySQL and PostgreSQL databases via stdio transport.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/jhoffmann/go-database-mcp/internal/config"
	"github.com/jhoffmann/go-database-mcp/internal/database"
	"github.com/jhoffmann/go-database-mcp/internal/handlers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the Database MCP Server instance.
// It wraps the MCP server implementation with database-specific configuration
// and provides lifecycle management.
type Server struct {
	config    *config.Config    // Database configuration
	server    *mcp.Server       // MCP server instance
	dbManager *database.Manager // Database manager
}

// NewServer creates a new Database MCP Server instance with the given configuration.
// It initializes the MCP server implementation with database-specific tools and handlers.
func NewServer(cfg *config.Config) (*Server, error) {
	impl := &mcp.Implementation{
		Name:    "database-mcp",
		Version: "1.0.0",
	}

	mcpServer := mcp.NewServer(impl, nil)

	// Create database manager
	dbManager, err := database.NewManager(cfg.Database)
	if err != nil {
		return nil, err
	}

	server := &Server{
		config:    cfg,
		server:    mcpServer,
		dbManager: dbManager,
	}

	// Register MCP tools
	server.registerTools()

	return server, nil
}

// registerTools registers all MCP tools with the server.
func (s *Server) registerTools() {
	// Query tool - Execute SQL queries with result formatting
	type QueryArgs struct {
		Query  string `json:"query" jsonschema:"the SQL query to execute"`
		Args   []any  `json:"args,omitempty" jsonschema:"parameters for the query"`
		Format string `json:"format,omitempty" jsonschema:"output format (json or table)"`
	}

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "query",
		Description: "Execute SQL queries with parameter binding and result formatting",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewQueryHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.ExecuteQuery(ctx, args.Query, args.Args...)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		format := args.Format
		if format == "" {
			format = "json"
		}

		formatted, err := handler.FormatResult(*result, format)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error formatting result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: formatted},
			},
		}, result, nil
	})

	// List tables tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_tables",
		Description: "List all tables in the current database",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewSchemaHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.ListTables(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d tables: %v", result.Count, result.Tables)},
			},
		}, result, nil
	})

	// List databases tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_databases",
		Description: "List all available databases on the server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewSchemaHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.ListDatabases(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Found %d databases: %v", result.Count, result.Databases)},
			},
		}, result, nil
	})

	// Describe table tool
	type DescribeTableArgs struct {
		TableName string `json:"table_name" jsonschema:"name of the table to describe"`
	}

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "describe_table",
		Description: "Get detailed schema information about a specific table",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DescribeTableArgs) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewSchemaHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.DescribeTable(ctx, args.TableName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Table %s has %d columns and %d indexes",
					result.Schema.TableName, len(result.Schema.Columns), len(result.Schema.Indexes))},
			},
		}, result, nil
	})

	// Get table data tool
	type GetTableDataArgs struct {
		TableName string `json:"table_name" jsonschema:"name of the table to get data from"`
		Limit     int    `json:"limit,omitempty" jsonschema:"maximum number of rows to return"`
		Offset    int    `json:"offset,omitempty" jsonschema:"number of rows to skip"`
	}

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "get_table_data",
		Description: "Retrieve paginated data from a specific table",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetTableDataArgs) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewSchemaHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.GetTableData(ctx, args.TableName, args.Limit, args.Offset)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Retrieved %d rows from %s (total: %d)",
					len(result.Data.Rows), result.Data.TableName, result.Data.Total)},
			},
		}, result, nil
	})

	// Explain query tool
	type ExplainQueryArgs struct {
		Query string `json:"query" jsonschema:"SQL query to explain"`
	}

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "explain_query",
		Description: "Get the execution plan for a SQL query",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ExplainQueryArgs) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewSchemaHandler(s.dbManager.GetDatabase(), &s.config.Database)
		result, err := handler.ExplainQuery(ctx, args.Query)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Execution plan for query:\n%s", result.Plan)},
			},
		}, result, nil
	})

	// Connection info tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "connection_info",
		Description: "Get information about the current database connection",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
		if s.dbManager.GetDatabase() == nil {
			return nil, nil, fmt.Errorf("database not connected")
		}

		handler := handlers.NewAdminHandler(s.dbManager.GetDatabase())
		result, err := handler.GetConnectionInfo(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Driver: %s, Connected: %v, Ping: %s",
					result.Driver, result.Connected, result.PingTime)},
			},
		}, result, nil
	})
}

// Start begins serving MCP requests using stdio transport.
// It establishes database connections and starts the MCP server to handle client requests.
// The server will run until the context is cancelled or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	// Connect to database
	log.Printf("Connecting to database...")
	if err := s.dbManager.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database connected successfully")

	transport := &mcp.StdioTransport{}

	log.Printf("Starting Database MCP Server...")
	log.Printf("Database type: %s", s.config.Database.Type)
	log.Printf("Database host: %s:%d", s.config.Database.Host, s.config.Database.Port)

	return s.server.Run(ctx, transport)
}

// main is the entry point for the Database MCP Server.
// It loads configuration, initializes the server, and handles graceful shutdown
// on SIGINT and SIGTERM signals.
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Database MCP Server...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Database type: %s", cfg.Database.Type)
	log.Printf("Database host: %s:%d", cfg.Database.Host, cfg.Database.Port)

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
