package database

import (
	"context"
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

func TestNewManager_ValidConfig(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v, expected nil", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	if manager.config.Type != cfg.Type {
		t.Errorf("Expected manager config Type = %s, got %s", cfg.Type, manager.config.Type)
	}
}

func TestNewManager_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    config.DatabaseConfig
		wantError string
	}{
		{
			name: "missing database type",
			config: config.DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "testdb",
				Username:     "testuser",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "database type is required",
		},
		{
			name: "unsupported database type",
			config: config.DatabaseConfig{
				Type:         "oracle",
				Host:         "localhost",
				Port:         5432,
				Database:     "testdb",
				Username:     "testuser",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "unsupported database type: oracle",
		},
		{
			name: "missing host",
			config: config.DatabaseConfig{
				Type:         "postgres",
				Port:         5432,
				Database:     "testdb",
				Username:     "testuser",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "database host is required",
		},
		{
			name: "missing port",
			config: config.DatabaseConfig{
				Type:         "postgres",
				Host:         "localhost",
				Database:     "testdb",
				Username:     "testuser",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "database port is required",
		},
		{
			name: "missing database name",
			config: config.DatabaseConfig{
				Type:         "postgres",
				Host:         "localhost",
				Port:         5432,
				Username:     "testuser",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "database name is required",
		},
		{
			name: "missing username",
			config: config.DatabaseConfig{
				Type:         "postgres",
				Host:         "localhost",
				Port:         5432,
				Database:     "testdb",
				MaxConns:     10,
				MaxIdleConns: 5,
			},
			wantError: "database username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewManager(tt.config)
			if err == nil {
				t.Errorf("NewManager() expected error containing %q, got nil", tt.wantError)
				return
			}
			if !contains(err.Error(), tt.wantError) {
				t.Errorf("NewManager() error = %v, expected error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestManager_GetDatabase_BeforeConnect(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v, expected nil", err)
	}

	db := manager.GetDatabase()
	if db != nil {
		t.Error("GetDatabase() expected nil before Connect(), got non-nil")
	}
}

func TestManager_Close_BeforeConnect(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v, expected nil", err)
	}

	// Should not panic or error when closing before connecting
	err = manager.Close()
	if err != nil {
		t.Errorf("Close() error = %v, expected nil", err)
	}
}

func TestManager_Ping_BeforeConnect(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}

	manager, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("NewManager() error = %v, expected nil", err)
	}

	ctx := context.Background()
	err = manager.Ping(ctx)
	if err == nil {
		t.Error("Ping() expected error before Connect(), got nil")
	}
	if !contains(err.Error(), "no database connection established") {
		t.Errorf("Ping() error = %v, expected error containing 'no database connection established'", err)
	}
}

func TestValidateConfig_AllValid(t *testing.T) {
	tests := []struct {
		name   string
		config config.DatabaseConfig
	}{
		{
			name: "valid postgres config",
			config: config.DatabaseConfig{
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
		},
		{
			name: "valid mysql config",
			config: config.DatabaseConfig{
				Type:         "mysql",
				Host:         "localhost",
				Port:         3306,
				Database:     "testdb",
				Username:     "testuser",
				Password:     "testpass",
				MaxConns:     25,
				MaxIdleConns: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if err != nil {
				t.Errorf("validateConfig() error = %v, expected nil", err)
			}
		})
	}
}

func TestValidateConfig_AllInvalid(t *testing.T) {
	tests := []struct {
		name      string
		config    config.DatabaseConfig
		wantError string
	}{
		{
			name: "empty database type",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			wantError: "database type is required",
		},
		{
			name: "unsupported database type",
			config: config.DatabaseConfig{
				Type:     "sqlite",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			wantError: "unsupported database type: sqlite",
		},
		{
			name: "empty host",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
			},
			wantError: "database host is required",
		},
		{
			name: "zero port",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     0,
				Database: "testdb",
				Username: "testuser",
			},
			wantError: "database port is required",
		},
		{
			name: "empty database name",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "testuser",
			},
			wantError: "database name is required",
		},
		{
			name: "empty username",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantError: "database username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if err == nil {
				t.Errorf("validateConfig() expected error containing %q, got nil", tt.wantError)
				return
			}
			if !contains(err.Error(), tt.wantError) {
				t.Errorf("validateConfig() error = %v, expected error containing %q", err, tt.wantError)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
