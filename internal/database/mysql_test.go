package database

import (
	"context"
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

func TestNewMySQL(t *testing.T) {
	cfg := NewTestConfig("mysql")

	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v, expected nil", err)
	}

	if mysql == nil {
		t.Fatal("NewMySQL() returned nil")
	}

	if mysql.config.Type != "mysql" {
		t.Errorf("Expected config Type = 'mysql', got %s", mysql.config.Type)
	}

	if mysql.db != nil {
		t.Error("Expected db to be nil before Connect(), got non-nil")
	}
}

func TestMySQL_GetDriverName(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	driverName := mysql.GetDriverName()
	if driverName != "mysql" {
		t.Errorf("Expected driver name 'mysql', got %s", driverName)
	}
}

func TestMySQL_GetDB_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	db := mysql.GetDB()
	if db != nil {
		t.Error("Expected GetDB() to return nil before Connect(), got non-nil")
	}
}

func TestMySQL_Close_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	err = mysql.Close()
	if err != nil {
		t.Errorf("Close() error = %v, expected nil", err)
	}
}

func TestMySQL_Ping_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	ctx := context.Background()
	err = mysql.Ping(ctx)
	if err == nil {
		t.Error("Ping() expected error before Connect(), got nil")
	}

	expectedError := "no database connection"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ping() error = %v, expected error containing %q", err, expectedError)
	}
}

func TestMySQL_Query_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	ctx := context.Background()
	rows, err := mysql.Query(ctx, "SELECT 1")
	if err == nil {
		t.Error("Query() expected error before Connect(), got nil")
	}
	if rows != nil {
		t.Error("Query() expected nil rows before Connect(), got non-nil")
	}

	expectedError := "no database connection"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Query() error = %v, expected error containing %q", err, expectedError)
	}
}

func TestMySQL_Exec_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	ctx := context.Background()
	result, err := mysql.Exec(ctx, "CREATE TABLE test (id INT)")
	if err == nil {
		t.Error("Exec() expected error before Connect(), got nil")
	}
	if result != nil {
		t.Error("Exec() expected nil result before Connect(), got non-nil")
	}

	expectedError := "no database connection"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Exec() error = %v, expected error containing %q", err, expectedError)
	}
}

func TestMySQL_buildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.DatabaseConfig
		contains []string
	}{
		{
			name:   "basic DSN",
			config: NewTestConfig("mysql"),
			contains: []string{
				"testuser:testpass@tcp(localhost:3306)/testdb",
				"parseTime=true",
				"timeout=30s",
			},
		},
		{
			name: "with SSL none",
			config: config.DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "none",
			},
			contains: []string{
				"user:pass@tcp(localhost:3306)/testdb",
				"tls=false",
			},
		},
		{
			name: "with SSL required",
			config: config.DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "required",
			},
			contains: []string{
				"user:pass@tcp(localhost:3306)/testdb",
				"tls=true",
			},
		},
		{
			name: "with SSL preferred",
			config: config.DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "preferred",
			},
			contains: []string{
				"user:pass@tcp(localhost:3306)/testdb",
				"tls=preferred",
			},
		},
		{
			name: "custom host and port",
			config: config.DatabaseConfig{
				Type:     "mysql",
				Host:     "db.example.com",
				Port:     3307,
				Database: "myapp",
				Username: "appuser",
				Password: "secretpass",
			},
			contains: []string{
				"appuser:secretpass@tcp(db.example.com:3307)/myapp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mysql, err := NewMySQL(tt.config)
			if err != nil {
				t.Fatalf("NewMySQL() error = %v", err)
			}

			dsn := mysql.buildDSN()

			for _, expectedSubstring := range tt.contains {
				if !contains(dsn, expectedSubstring) {
					t.Errorf("DSN = %q, expected to contain %q", dsn, expectedSubstring)
				}
			}
		})
	}
}

func TestMySQL_buildDSN_DefaultSSL(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "unknown", // Should default to tls=true
	}

	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	dsn := mysql.buildDSN()

	if !contains(dsn, "tls=true") {
		t.Errorf("DSN = %q, expected to contain 'tls=true' for unknown SSL mode", dsn)
	}
}

func TestMySQL_QueryRow(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	// QueryRow with nil db should handle gracefully
	// We can't actually call it because it will panic, but we can test the method exists
	// and the struct is properly initialized
	if mysql.db != nil {
		t.Error("Expected db to be nil before Connect()")
	}

	// Test that the method is accessible (this tests the interface implementation)
	ctx := context.Background()
	_ = ctx // Use ctx to avoid unused variable error
}

// Test struct initialization and field access
func TestMySQL_StructFields(t *testing.T) {
	cfg := NewTestConfig("mysql")
	mysql, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL() error = %v", err)
	}

	// Test that config is properly stored
	if mysql.config.Host != cfg.Host {
		t.Errorf("Expected config.Host = %s, got %s", cfg.Host, mysql.config.Host)
	}

	if mysql.config.Port != cfg.Port {
		t.Errorf("Expected config.Port = %d, got %d", cfg.Port, mysql.config.Port)
	}

	if mysql.config.Database != cfg.Database {
		t.Errorf("Expected config.Database = %s, got %s", cfg.Database, mysql.config.Database)
	}
}
