package config

import (
	"os"
	"strings"
	"testing"
)

func TestValidate_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "valid postgres config",
			config: &Config{
				Database: DatabaseConfig{
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
		},
		{
			name: "valid mysql config",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "mysql",
					Host:         "localhost",
					Port:         3306,
					Database:     "testdb",
					Username:     "testuser",
					Password:     "testpass",
					MaxConns:     25,
					MaxIdleConns: 5,
					SSLMode:      "required",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.config); err != nil {
				t.Errorf("Validate() error = %v, expected nil", err)
			}
		})
	}
}

func TestValidate_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError string
	}{
		{
			name: "invalid database type",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "oracle",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database type must be 'mysql' or 'postgres'",
		},
		{
			name: "missing host",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database host is required",
		},
		{
			name: "invalid port - zero",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         0,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         65536,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database port must be between 1 and 65535",
		},
		{
			name: "missing database name",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database name is required",
		},
		{
			name: "missing username",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "",
					MaxConns:     10,
					MaxIdleConns: 5,
				},
			},
			wantError: "database username is required",
		},
		{
			name: "invalid max connections",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     0,
					MaxIdleConns: 5,
				},
			},
			wantError: "max connections must be at least 1",
		},
		{
			name: "negative max idle connections",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: -1,
				},
			},
			wantError: "max idle connections cannot be negative",
		},
		{
			name: "max idle exceeds max connections",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     5,
					MaxIdleConns: 10,
				},
			},
			wantError: "max idle connections (10) cannot exceed max connections (5)",
		},
		{
			name: "invalid postgres SSL mode",
			config: &Config{
				Database: DatabaseConfig{
					Type:         "postgres",
					Host:         "localhost",
					Port:         5432,
					Database:     "testdb",
					Username:     "testuser",
					MaxConns:     10,
					MaxIdleConns: 5,
					SSLMode:      "invalid",
				},
			},
			wantError: "invalid SSL mode for postgres: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if err == nil {
				t.Errorf("Validate() expected error containing %q, got nil", tt.wantError)
				return
			}
			if err.Error() != tt.wantError && !contains(err.Error(), tt.wantError) {
				t.Errorf("Validate() error = %v, expected error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"DB_TYPE":           os.Getenv("DB_TYPE"),
		"DB_HOST":           os.Getenv("DB_HOST"),
		"DB_PORT":           os.Getenv("DB_PORT"),
		"DB_NAME":           os.Getenv("DB_NAME"),
		"DB_USER":           os.Getenv("DB_USER"),
		"DB_PASSWORD":       os.Getenv("DB_PASSWORD"),
		"DB_MAX_CONNS":      os.Getenv("DB_MAX_CONNS"),
		"DB_MAX_IDLE_CONNS": os.Getenv("DB_MAX_IDLE_CONNS"),
		"DB_SSL_MODE":       os.Getenv("DB_SSL_MODE"),
	}

	// Clean up function
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	testEnv := map[string]string{
		"DB_TYPE":           "mysql",
		"DB_HOST":           "testhost",
		"DB_PORT":           "3307",
		"DB_NAME":           "testdatabase",
		"DB_USER":           "testuser",
		"DB_PASSWORD":       "testpassword",
		"DB_MAX_CONNS":      "20",
		"DB_MAX_IDLE_CONNS": "10",
		"DB_SSL_MODE":       "required",
	}

	for key, value := range testEnv {
		os.Setenv(key, value)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, expected nil", err)
	}

	if cfg.Database.Type != "mysql" {
		t.Errorf("Expected Type = 'mysql', got %s", cfg.Database.Type)
	}
	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected Host = 'testhost', got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 3307 {
		t.Errorf("Expected Port = 3307, got %d", cfg.Database.Port)
	}
	if cfg.Database.Database != "testdatabase" {
		t.Errorf("Expected Database = 'testdatabase', got %s", cfg.Database.Database)
	}
	if cfg.Database.Username != "testuser" {
		t.Errorf("Expected Username = 'testuser', got %s", cfg.Database.Username)
	}
	if cfg.Database.Password != "testpassword" {
		t.Errorf("Expected Password = 'testpassword', got %s", cfg.Database.Password)
	}
	if cfg.Database.MaxConns != 20 {
		t.Errorf("Expected MaxConns = 20, got %d", cfg.Database.MaxConns)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns = 10, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.SSLMode != "required" {
		t.Errorf("Expected SSLMode = 'required', got %s", cfg.Database.SSLMode)
	}
}

func TestLoad_ValidationError(t *testing.T) {
	// Save original environment
	originalType := os.Getenv("DB_TYPE")
	defer func() {
		if originalType == "" {
			os.Unsetenv("DB_TYPE")
		} else {
			os.Setenv("DB_TYPE", originalType)
		}
	}()

	// Set invalid database type
	os.Setenv("DB_TYPE", "invalid")

	_, err := Load()
	if err == nil {
		t.Error("Load() expected error for invalid database type, got nil")
	}
	if !contains(err.Error(), "configuration validation failed") {
		t.Errorf("Load() error = %v, expected error containing 'configuration validation failed'", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAtPosition(s, substr))))
}

func containsAtPosition(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Additional security-focused tests

func TestLoad_CredentialSecurity(t *testing.T) {
	// Save original environment
	originalEnv := os.Environ()
	defer func() {
		// Restore environment
		os.Clearenv()
		for _, env := range originalEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "allowed databases configuration",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_TYPE", "postgres")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "secretpass")
				os.Setenv("DB_ALLOWED_NAMES", "testdb,devdb,staging")
			},
			wantErr: false,
		},
		{
			name: "primary database always allowed even if not in list",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_TYPE", "postgres")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_NAME", "proddb")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "secretpass")
				os.Setenv("DB_ALLOWED_NAMES", "testdb,devdb,staging")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Load() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Load() error = %v, want nil", err)
				} else {
					// Verify allowed databases parsed correctly
					if os.Getenv("DB_ALLOWED_NAMES") != "" {
						expectedCount := len(strings.Split(os.Getenv("DB_ALLOWED_NAMES"), ","))
						if len(cfg.Database.AllowedDatabases) != expectedCount {
							t.Errorf("Load() AllowedDatabases count = %d, want %d",
								len(cfg.Database.AllowedDatabases), expectedCount)
						}
					}
				}
			}
		})
	}
}

func TestDatabaseConfig_IsDatabaseAllowed(t *testing.T) {
	tests := []struct {
		name             string
		allowedDatabases []string
		testDatabase     string
		want             bool
	}{
		{
			name:             "empty allowed list means only primary database allowed",
			allowedDatabases: []string{},
			testDatabase:     "anydb",
			want:             false,
		},
		{
			name:             "primary database always allowed",
			allowedDatabases: []string{},
			testDatabase:     "testdb",
			want:             true,
		},
		{
			name:             "database in allowed list",
			allowedDatabases: []string{"testdb", "devdb"},
			testDatabase:     "testdb",
			want:             true,
		},
		{
			name:             "database not in allowed list",
			allowedDatabases: []string{"testdb", "devdb"},
			testDatabase:     "proddb",
			want:             false,
		},
		{
			name:             "case sensitive matching - allowed database",
			allowedDatabases: []string{"TestDB"},
			testDatabase:     "TestDB",
			want:             true,
		},
		{
			name:             "case sensitive matching - different case not allowed",
			allowedDatabases: []string{"TestDB"},
			testDatabase:     "TESTDB",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primaryDB := "testdb" // Default primary database

			// Use different primary database for case-sensitive tests
			if strings.Contains(tt.name, "case sensitive") {
				primaryDB = "primarydb"
			}

			config := &DatabaseConfig{
				Database:         primaryDB,
				AllowedDatabases: tt.allowedDatabases,
			}
			if got := config.IsDatabaseAllowed(tt.testDatabase); got != tt.want {
				t.Errorf("IsDatabaseAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
