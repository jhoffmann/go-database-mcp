package config

import (
	"testing"
)

func TestParseConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		connectionStr string
		expected      *ConnectionInfo
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid PostgreSQL connection string",
			connectionStr: "postgresql://user:pass@localhost:5432/mydb",
			expected: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "Valid PostgreSQL connection string with SSL mode",
			connectionStr: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
			expected: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "require",
			},
		},
		{
			name:          "Valid MySQL connection string",
			connectionStr: "mysql://user:pass@localhost:3306/mydb",
			expected: &ConnectionInfo{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "PostgreSQL without port (uses default)",
			connectionStr: "postgresql://user:pass@localhost/mydb",
			expected: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "MySQL without port (uses default)",
			connectionStr: "mysql://user:pass@localhost/mydb",
			expected: &ConnectionInfo{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "PostgreSQL without password",
			connectionStr: "postgresql://user@localhost:5432/mydb",
			expected: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "Connection string with 'postgres' scheme",
			connectionStr: "postgres://user:pass@localhost:5432/mydb",
			expected: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "prefer",
			},
		},
		{
			name:          "Empty connection string",
			connectionStr: "",
			expectError:   true,
			errorContains: "connection string is empty",
		},
		{
			name:          "Unsupported scheme",
			connectionStr: "oracle://user:pass@localhost:1521/mydb",
			expectError:   true,
			errorContains: "unsupported database scheme: oracle",
		},
		{
			name:          "Missing hostname",
			connectionStr: "postgresql://user:pass@/mydb",
			expectError:   true,
			errorContains: "hostname is required",
		},
		{
			name:          "Missing database name",
			connectionStr: "postgresql://user:pass@localhost:5432/",
			expectError:   true,
			errorContains: "database name is required",
		},
		{
			name:          "Missing database name (no slash)",
			connectionStr: "postgresql://user:pass@localhost:5432",
			expectError:   true,
			errorContains: "database name is required",
		},
		{
			name:          "Missing username",
			connectionStr: "postgresql://:pass@localhost:5432/mydb",
			expectError:   true,
			errorContains: "username is required",
		},
		{
			name:          "No user info",
			connectionStr: "postgresql://localhost:5432/mydb",
			expectError:   true,
			errorContains: "username is required",
		},
		{
			name:          "Invalid port",
			connectionStr: "postgresql://user:pass@localhost:abc/mydb",
			expectError:   true,
			errorContains: "invalid port",
		},
		{
			name:          "Malformed URL",
			connectionStr: "not-a-url",
			expectError:   true,
			errorContains: "unsupported database scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseConnectionString(tt.connectionStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', but got: %s", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Type: expected %s, got %s", tt.expected.Type, result.Type)
			}
			if result.Host != tt.expected.Host {
				t.Errorf("Host: expected %s, got %s", tt.expected.Host, result.Host)
			}
			if result.Port != tt.expected.Port {
				t.Errorf("Port: expected %d, got %d", tt.expected.Port, result.Port)
			}
			if result.Database != tt.expected.Database {
				t.Errorf("Database: expected %s, got %s", tt.expected.Database, result.Database)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("Username: expected %s, got %s", tt.expected.Username, result.Username)
			}
			if result.Password != tt.expected.Password {
				t.Errorf("Password: expected %s, got %s", tt.expected.Password, result.Password)
			}
			if result.SSLMode != tt.expected.SSLMode {
				t.Errorf("SSLMode: expected %s, got %s", tt.expected.SSLMode, result.SSLMode)
			}
		})
	}
}

func TestConnectionInfo_ToConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		info     *ConnectionInfo
		expected string
	}{
		{
			name: "PostgreSQL with password",
			info: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "require",
			},
			expected: "postgresql://user:pass@localhost:5432/mydb?sslmode=require",
		},
		{
			name: "MySQL without password",
			info: &ConnectionInfo{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "user",
				Password: "",
				SSLMode:  "prefer",
			},
			expected: "mysql://user@localhost:3306/mydb?sslmode=prefer",
		},
		{
			name: "PostgreSQL without SSL mode",
			info: &ConnectionInfo{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				SSLMode:  "",
			},
			expected: "postgresql://user:pass@localhost:5432/mydb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.ToConnectionString()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && indexOfString(s, substr) >= 0)
}

func indexOfString(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}
