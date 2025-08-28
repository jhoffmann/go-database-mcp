package config

import "testing"

func TestSSLMode_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		sslMode  SSLMode
		expected bool
	}{
		{"none is valid", SSLModeNone, true},
		{"prefer is valid", SSLModePrefer, true},
		{"require is valid", SSLModeRequire, true},
		{"invalid mode", SSLMode("invalid"), false},
		{"empty mode", SSLMode(""), false},
		{"case sensitive - None", SSLMode("None"), false},
		{"case sensitive - PREFER", SSLMode("PREFER"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sslMode.IsValid(); got != tt.expected {
				t.Errorf("SSLMode.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSSLMode_ToMySQLSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		sslMode     SSLMode
		expected    string
		shouldError bool
	}{
		{"none to MySQL", SSLModeNone, "false", false},
		{"prefer to MySQL", SSLModePrefer, "preferred", false},
		{"require to MySQL", SSLModeRequire, "true", false},
		{"invalid mode", SSLMode("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sslMode.ToMySQLSSLMode()

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("SSLMode.ToMySQLSSLMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSSLMode_ToPostgreSQLSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		sslMode     SSLMode
		expected    string
		shouldError bool
	}{
		{"none to PostgreSQL", SSLModeNone, "disable", false},
		{"prefer to PostgreSQL", SSLModePrefer, "prefer", false},
		{"require to PostgreSQL", SSLModeRequire, "require", false},
		{"invalid mode", SSLMode("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sslMode.ToPostgreSQLSSLMode()

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("SSLMode.ToPostgreSQLSSLMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    SSLMode
		shouldError bool
	}{
		{"parse none", "none", SSLModeNone, false},
		{"parse prefer", "prefer", SSLModePrefer, false},
		{"parse require", "require", SSLModeRequire, false},
		{"parse invalid", "invalid", "", true},
		{"parse empty", "", "", true},
		{"parse case sensitive", "None", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSSLMode(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("ParseSSLMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidSSLModes(t *testing.T) {
	modes := ValidSSLModes()

	if len(modes) != 3 {
		t.Errorf("ValidSSLModes() returned %d modes, expected 3", len(modes))
	}

	expected := []SSLMode{SSLModeNone, SSLModePrefer, SSLModeRequire}
	for i, expected := range expected {
		if i >= len(modes) || modes[i] != expected {
			t.Errorf("ValidSSLModes()[%d] = %v, want %v", i, modes[i], expected)
		}
	}
}

func TestDatabaseConfig_ValidateSSLMode(t *testing.T) {
	tests := []struct {
		name        string
		sslMode     string
		expected    SSLMode
		shouldError bool
	}{
		{"empty defaults to none", "", SSLModeNone, false},
		{"valid none", "none", SSLModeNone, false},
		{"valid prefer", "prefer", SSLModePrefer, false},
		{"valid require", "require", SSLModeRequire, false},
		{"invalid mode", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &DatabaseConfig{
				SSLMode: tt.sslMode,
			}

			got, err := cfg.ValidateSSLMode()

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.expected {
				t.Errorf("DatabaseConfig.ValidateSSLMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}
