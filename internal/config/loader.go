package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Load reads configuration from environment variables and .env file.
// It first loads variables from .env file if present, then processes environment variables
// which take precedence over .env file values. The configuration is validated before returning.
// Returns an error if loading or validation fails.
func Load() (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Create config with minimal defaults (only for values that don't come from connection strings)
	cfg := &Config{
		Database: DatabaseConfig{
			AllowedDatabases: []string{}, // Empty means only primary database allowed
			MaxConns:         10,
			MaxIdleConns:     5,
		},
	}

	// Load environment variables first to see what's explicitly set
	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("error processing database config: %w", err)
	}

	// Apply connection string values for any fields that weren't set by env vars
	if err := cfg.Database.ApplyConnectionStringDefaults(); err != nil {
		return nil, fmt.Errorf("error processing connection string: %w", err)
	}

	// Apply final defaults for any fields that are still empty
	if cfg.Database.Type == "" {
		cfg.Database.Type = "postgres"
	}
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "prefer"
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks the configuration for required fields and valid values.
// It ensures database type is supported, connection parameters are valid,
// and SSL modes are appropriate for the selected database type.
// Supports both connection string and individual parameter configuration.
// Returns an error describing any validation failures.
func Validate(cfg *Config) error {
	// Check if we have either a connection string or individual parameters
	if cfg.Database.ConnectionString == "" {
		// Validate individual parameters approach
		if cfg.Database.Type == "" {
			return fmt.Errorf("database type is required (either via connection string or DB_TYPE)")
		}
		if cfg.Database.Host == "" {
			return fmt.Errorf("database host is required (either via connection string or DB_HOST)")
		}
		if cfg.Database.Database == "" {
			return fmt.Errorf("database name is required (either via connection string or DB_NAME)")
		}
		if cfg.Database.Username == "" {
			return fmt.Errorf("database username is required (either via connection string or DB_USER)")
		}
	}

	// Validate database type (should be populated by now)
	if cfg.Database.Type != "mysql" && cfg.Database.Type != "postgres" {
		return fmt.Errorf("database type must be 'mysql' or 'postgres', got '%s'", cfg.Database.Type)
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("database port must be between 1 and 65535, got %d", cfg.Database.Port)
	}

	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if cfg.Database.MaxConns < 1 {
		return fmt.Errorf("max connections must be at least 1, got %d", cfg.Database.MaxConns)
	}

	if cfg.Database.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative, got %d", cfg.Database.MaxIdleConns)
	}

	if cfg.Database.MaxIdleConns > cfg.Database.MaxConns {
		return fmt.Errorf("max idle connections (%d) cannot exceed max connections (%d)",
			cfg.Database.MaxIdleConns, cfg.Database.MaxConns)
	}

	if cfg.Database.Type == "postgres" {
		validSSLModes := map[string]bool{
			"disable":     true,
			"require":     true,
			"verify-ca":   true,
			"verify-full": true,
			"prefer":      true,
		}
		if !validSSLModes[cfg.Database.SSLMode] {
			return fmt.Errorf("invalid SSL mode for postgres: %s", cfg.Database.SSLMode)
		}
	}

	// Note: Primary database is always allowed by design, no validation needed

	return nil
}
