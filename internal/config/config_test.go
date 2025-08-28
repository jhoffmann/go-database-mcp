package config

import (
	"testing"
)

func TestDatabaseConfig_DefaultValues(t *testing.T) {
	cfg := DatabaseConfig{
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

	if cfg.Type != "postgres" {
		t.Errorf("Expected Type to be 'postgres', got %s", cfg.Type)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected Port to be 5432, got %d", cfg.Port)
	}
	if cfg.Database != "testdb" {
		t.Errorf("Expected Database to be 'testdb', got %s", cfg.Database)
	}
	if cfg.Username != "testuser" {
		t.Errorf("Expected Username to be 'testuser', got %s", cfg.Username)
	}
	if cfg.Password != "testpass" {
		t.Errorf("Expected Password to be 'testpass', got %s", cfg.Password)
	}
	if cfg.MaxConns != 10 {
		t.Errorf("Expected MaxConns to be 10, got %d", cfg.MaxConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("Expected MaxIdleConns to be 5, got %d", cfg.MaxIdleConns)
	}
	if cfg.SSLMode != "prefer" {
		t.Errorf("Expected SSLMode to be 'prefer', got %s", cfg.SSLMode)
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Type: "mysql",
			Host: "db.example.com",
			Port: 3306,
		},
	}

	if cfg.Database.Type != "mysql" {
		t.Errorf("Expected Database.Type to be 'mysql', got %s", cfg.Database.Type)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected Database.Host to be 'db.example.com', got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("Expected Database.Port to be 3306, got %d", cfg.Database.Port)
	}
}

func TestDatabaseConfig_ApplyConnectionStringDefaults(t *testing.T) {
	tests := []struct {
		name        string
		config      DatabaseConfig
		expectError bool
		expected    DatabaseConfig
	}{
		{
			name: "PostgreSQL connection string populates all fields",
			config: DatabaseConfig{
				ConnectionString: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
				// All individual fields start empty
			},
			expected: DatabaseConfig{
				ConnectionString: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
				Type:             "postgres",
				Host:             "localhost",
				Port:             5432,
				Database:         "mydb",
				Username:         "user",
				Password:         "pass",
				SSLMode:          "require",
			},
		},
		{
			name: "MySQL connection string populates all fields",
			config: DatabaseConfig{
				ConnectionString: "mysql://user:pass@dbhost:3306/mydb",
				// All individual fields start empty
			},
			expected: DatabaseConfig{
				ConnectionString: "mysql://user:pass@dbhost:3306/mydb",
				Type:             "mysql",
				Host:             "dbhost",
				Port:             3306,
				Database:         "mydb",
				Username:         "user",
				Password:         "pass",
				SSLMode:          "prefer",
			},
		},
		{
			name: "Individual parameters take precedence over connection string",
			config: DatabaseConfig{
				ConnectionString: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
				Type:             "mysql",    // Individual parameter overrides connection string
				Host:             "custom",   // Individual parameter overrides connection string
				Port:             3306,       // Individual parameter overrides connection string
				Database:         "customdb", // Individual parameter overrides connection string
				Username:         "admin",    // Individual parameter overrides connection string
				Password:         "secret",   // Individual parameter overrides connection string
				SSLMode:          "none",     // Individual parameter overrides connection string
			},
			expected: DatabaseConfig{
				ConnectionString: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
				Type:             "mysql",    // From individual parameter
				Host:             "custom",   // From individual parameter
				Port:             3306,       // From individual parameter
				Database:         "customdb", // From individual parameter
				Username:         "admin",    // From individual parameter
				Password:         "secret",   // From individual parameter
				SSLMode:          "none",     // From individual parameter
			},
		},
		{
			name: "No connection string does nothing",
			config: DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "test",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
			expected: DatabaseConfig{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "test",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name: "Invalid connection string returns error",
			config: DatabaseConfig{
				ConnectionString: "invalid://connection/string",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ApplyConnectionStringDefaults()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Compare the results
			if tt.config.Type != tt.expected.Type {
				t.Errorf("Type: expected %s, got %s", tt.expected.Type, tt.config.Type)
			}
			if tt.config.Host != tt.expected.Host {
				t.Errorf("Host: expected %s, got %s", tt.expected.Host, tt.config.Host)
			}
			if tt.config.Port != tt.expected.Port {
				t.Errorf("Port: expected %d, got %d", tt.expected.Port, tt.config.Port)
			}
			if tt.config.Database != tt.expected.Database {
				t.Errorf("Database: expected %s, got %s", tt.expected.Database, tt.config.Database)
			}
			if tt.config.Username != tt.expected.Username {
				t.Errorf("Username: expected %s, got %s", tt.expected.Username, tt.config.Username)
			}
			if tt.config.Password != tt.expected.Password {
				t.Errorf("Password: expected %s, got %s", tt.expected.Password, tt.config.Password)
			}
			if tt.config.SSLMode != tt.expected.SSLMode {
				t.Errorf("SSLMode: expected %s, got %s", tt.expected.SSLMode, tt.config.SSLMode)
			}
		})
	}
}
