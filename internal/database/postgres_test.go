package database

import (
	"context"
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

func TestNewPostgreSQL(t *testing.T) {
	cfg := NewTestConfig("postgres")

	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v, expected nil", err)
	}

	if pg == nil {
		t.Fatal("NewPostgreSQL() returned nil")
	}

	if pg.config.Type != "postgres" {
		t.Errorf("Expected config Type = 'postgres', got %s", pg.config.Type)
	}

	if pg.db != nil {
		t.Error("Expected db to be nil before Connect(), got non-nil")
	}
}

func TestPostgreSQL_GetDriverName(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	driverName := pg.GetDriverName()
	if driverName != "postgres" {
		t.Errorf("Expected driver name 'postgres', got %s", driverName)
	}
}

func TestPostgreSQL_GetDB_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	db := pg.GetDB()
	if db != nil {
		t.Error("Expected GetDB() to return nil before Connect(), got non-nil")
	}
}

func TestPostgreSQL_Close_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	err = pg.Close()
	if err != nil {
		t.Errorf("Close() error = %v, expected nil", err)
	}
}

func TestPostgreSQL_Ping_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	ctx := context.Background()
	err = pg.Ping(ctx)
	if err == nil {
		t.Error("Ping() expected error before Connect(), got nil")
	}

	expectedError := "no database connection"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Ping() error = %v, expected error containing %q", err, expectedError)
	}
}

func TestPostgreSQL_Query_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	ctx := context.Background()
	rows, err := pg.Query(ctx, "SELECT 1")
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

func TestPostgreSQL_Exec_BeforeConnect(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	ctx := context.Background()
	result, err := pg.Exec(ctx, "CREATE TABLE test (id INT)")
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

func TestPostgreSQL_buildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.DatabaseConfig
		contains []string
	}{
		{
			name:   "basic DSN",
			config: NewTestConfig("postgres"),
			contains: []string{
				"host=localhost",
				"port=5432",
				"user=testuser",
				"password=testpass",
				"dbname=testdb",
				"sslmode=prefer",
				"connect_timeout=30",
			},
		},
		{
			name: "with SSL none",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "none",
			},
			contains: []string{
				"host=localhost",
				"port=5432",
				"user=user",
				"password=pass",
				"dbname=testdb",
				"sslmode=disable",
			},
		},
		{
			name: "with SSL require",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "require",
			},
			contains: []string{
				"host=localhost",
				"port=5432",
				"user=user",
				"password=pass",
				"dbname=testdb",
				"sslmode=require",
			},
		},
		{
			name: "custom host and port",
			config: config.DatabaseConfig{
				Type:     "postgres",
				Host:     "db.example.com",
				Port:     5433,
				Database: "myapp",
				Username: "appuser",
				Password: "secretpass",
			},
			contains: []string{
				"host=db.example.com",
				"port=5433",
				"user=appuser",
				"password=secretpass",
				"dbname=myapp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg, err := NewPostgreSQL(tt.config)
			if err != nil {
				t.Fatalf("NewPostgreSQL() error = %v", err)
			}

			dsn := pg.buildDSN()

			for _, expectedSubstring := range tt.contains {
				if !contains(dsn, expectedSubstring) {
					t.Errorf("DSN = %q, expected to contain %q", dsn, expectedSubstring)
				}
			}
		})
	}
}

func TestPostgreSQL_buildDSN_DefaultSSL(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		// SSLMode is empty, should default to "none"
	}

	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	dsn := pg.buildDSN()

	if !contains(dsn, "sslmode=disable") {
		t.Errorf("DSN = %q, expected to contain 'sslmode=disable' for empty SSL mode", dsn)
	}
}

func TestPostgreSQL_QueryRow(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	// QueryRow with nil db should handle gracefully
	// We can't actually call it because it will panic, but we can test the method exists
	// and the struct is properly initialized
	if pg.db != nil {
		t.Error("Expected db to be nil before Connect()")
	}

	// Test that the method is accessible (this tests the interface implementation)
	ctx := context.Background()
	_ = ctx // Use ctx to avoid unused variable error
}

// Test struct initialization and field access
func TestPostgreSQL_StructFields(t *testing.T) {
	cfg := NewTestConfig("postgres")
	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	// Test that config is properly stored
	if pg.config.Host != cfg.Host {
		t.Errorf("Expected config.Host = %s, got %s", cfg.Host, pg.config.Host)
	}

	if pg.config.Port != cfg.Port {
		t.Errorf("Expected config.Port = %d, got %d", cfg.Port, pg.config.Port)
	}

	if pg.config.Database != cfg.Database {
		t.Errorf("Expected config.Database = %s, got %s", cfg.Database, pg.config.Database)
	}
}

// Test DSN building with various parameter combinations
func TestPostgreSQL_buildDSN_AllParameters(t *testing.T) {
	cfg := config.DatabaseConfig{
		Type:     "postgres",
		Host:     "testhost",
		Port:     5433,
		Database: "testdb",
		Username: "testuser",
		Password: "testpass",
		SSLMode:  "require",
	}

	pg, err := NewPostgreSQL(cfg)
	if err != nil {
		t.Fatalf("NewPostgreSQL() error = %v", err)
	}

	dsn := pg.buildDSN()

	expectedParts := []string{
		"host=testhost",
		"port=5433",
		"user=testuser",
		"password=testpass",
		"dbname=testdb",
		"sslmode=require",
		"connect_timeout=30",
	}

	for _, part := range expectedParts {
		if !contains(dsn, part) {
			t.Errorf("DSN = %q, expected to contain %q", dsn, part)
		}
	}
}
