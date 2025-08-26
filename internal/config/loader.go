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

	cfg := &Config{
		Database: DatabaseConfig{
			Type:             "postgres",
			Host:             "localhost",
			Port:             5432,
			Database:         "",
			AllowedDatabases: []string{}, // Empty means only primary database allowed
			Username:         "",
			Password:         "",
			MaxConns:         10,
			MaxIdleConns:     5,
			SSLMode:          "prefer",
		},
	}

	if err := envconfig.Process("", &cfg.Database); err != nil {
		return nil, fmt.Errorf("error processing database config: %w", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks the configuration for required fields and valid values.
// It ensures database type is supported, connection parameters are valid,
// and SSL modes are appropriate for the selected database type.
// Returns an error describing any validation failures.
func Validate(cfg *Config) error {
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
