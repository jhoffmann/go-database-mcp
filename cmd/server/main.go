// Package main provides the entry point for the Database MCP Server.
// It implements a Model Context Protocol server that provides database connectivity
// for MySQL and PostgreSQL databases via stdio transport.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/jhoffmann/go-database-mcp/internal/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the Database MCP Server instance.
// It wraps the MCP server implementation with database-specific configuration
// and provides lifecycle management.
type Server struct {
	config *config.Config // Database configuration
	server *mcp.Server    // MCP server instance
}

// NewServer creates a new Database MCP Server instance with the given configuration.
// It initializes the MCP server implementation with database-specific tools and handlers.
func NewServer(cfg *config.Config) *Server {
	impl := &mcp.Implementation{
		Name:    "database-mcp",
		Version: "1.0.0",
	}

	mcpServer := mcp.NewServer(impl, nil)

	return &Server{
		config: cfg,
		server: mcpServer,
	}
}

// Start begins serving MCP requests using stdio transport.
// It establishes database connections and starts the MCP server to handle client requests.
// The server will run until the context is cancelled or an error occurs.
func (s *Server) Start(ctx context.Context) error {
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

	server := NewServer(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
