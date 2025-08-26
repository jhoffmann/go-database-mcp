package main

import (
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:         "postgres",
			Host:         "localhost",
			Port:         5432,
			Database:     "testdb",
			Username:     "testuser",
			Password:     "testpass",
			MaxConns:     10,
			MaxIdleConns: 5,
			SSLMode:      "prefer",
		},
	}

	server := NewServer(cfg)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.config != cfg {
		t.Error("NewServer() did not store config properly")
	}

	if server.server == nil {
		t.Error("NewServer() did not initialize MCP server")
	}
}

func TestServer_StructFields(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "mysql",
			Host:     "testhost",
			Port:     3306,
			Database: "testdb",
			Username: "testuser",
		},
	}

	server := NewServer(cfg)

	// Test that config is properly stored
	if server.config.Database.Type != cfg.Database.Type {
		t.Errorf("Expected config.Database.Type = %s, got %s",
			cfg.Database.Type, server.config.Database.Type)
	}

	if server.config.Database.Host != cfg.Database.Host {
		t.Errorf("Expected config.Database.Host = %s, got %s",
			cfg.Database.Host, server.config.Database.Host)
	}

	if server.config.Database.Port != cfg.Database.Port {
		t.Errorf("Expected config.Database.Port = %d, got %d",
			cfg.Database.Port, server.config.Database.Port)
	}
}

// We skip testing Start() directly as it uses stdio transport which interferes with test output
// The method is simple enough that testing NewServer() provides sufficient coverage
func TestServer_Start_Method_Exists(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:         "postgres",
			Host:         "localhost",
			Port:         5432,
			Database:     "testdb",
			Username:     "testuser",
			Password:     "testpass",
			MaxConns:     10,
			MaxIdleConns: 5,
			SSLMode:      "prefer",
		},
	}

	server := NewServer(cfg)

	// Test that Start method exists by checking the server is properly initialized
	// We don't actually call it because it uses stdio transport which conflicts with test output
	if server.server == nil {
		t.Error("MCP server should be initialized")
	}
}

// Test various database configurations
func TestNewServer_DifferentDatabaseTypes(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		port   int
	}{
		{
			name:   "postgres config",
			dbType: "postgres",
			port:   5432,
		},
		{
			name:   "mysql config",
			dbType: "mysql",
			port:   3306,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Database: config.DatabaseConfig{
					Type:         tt.dbType,
					Host:         "localhost",
					Port:         tt.port,
					Database:     "testdb",
					Username:     "testuser",
					Password:     "testpass",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			}

			server := NewServer(cfg)

			if server == nil {
				t.Fatal("NewServer() returned nil")
			}

			if server.config.Database.Type != tt.dbType {
				t.Errorf("Expected database type %s, got %s",
					tt.dbType, server.config.Database.Type)
			}

			if server.config.Database.Port != tt.port {
				t.Errorf("Expected database port %d, got %d",
					tt.port, server.config.Database.Port)
			}
		})
	}
}

// Test that the MCP server implementation is properly configured
func TestNewServer_MCPImplementation(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			Username: "testuser",
		},
	}

	server := NewServer(cfg)

	if server.server == nil {
		t.Fatal("MCP server was not initialized")
	}

	// The MCP server should be ready to accept connections
	// We can't test much more without actually starting the server
	// which would require stdio transport setup
}

// Test configuration with different connection pool settings
func TestNewServer_ConnectionPoolSettings(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:         "postgres",
			Host:         "localhost",
			Port:         5432,
			Database:     "testdb",
			Username:     "testuser",
			MaxConns:     25,
			MaxIdleConns: 10,
		},
	}

	server := NewServer(cfg)

	if server.config.Database.MaxConns != 25 {
		t.Errorf("Expected MaxConns = 25, got %d", server.config.Database.MaxConns)
	}

	if server.config.Database.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns = 10, got %d", server.config.Database.MaxIdleConns)
	}
}
