package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/jhoffmann/go-database-mcp/internal/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	config *config.Config
	server *mcp.Server
}

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

func (s *Server) Start(ctx context.Context) error {
	transport := &mcp.StdioTransport{}

	log.Printf("Starting Database MCP Server...")
	log.Printf("Database type: %s", s.config.Database.Type)
	log.Printf("Database host: %s:%d", s.config.Database.Host, s.config.Database.Port)

	return s.server.Run(ctx, transport)
}

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
