// Package config provides configuration structures and loading functionality
// for the database MCP server.
package config

import (
	"fmt"
	"slices"
)

// Config represents the complete configuration for the database MCP server.
type Config struct {
	Database DatabaseConfig `json:"database"` // Database connection configuration
}

// DatabaseConfig contains all settings required to connect to a database.
// It supports both MySQL and PostgreSQL databases with SSL/TLS configuration.
// Supports either individual connection parameters or a single connection string.
type DatabaseConfig struct {
	// Connection string approach (preferred)
	ConnectionString string `json:"connection_string" envconfig:"DB_CONNECTION_STRING"` // Full database connection string (postgresql:// or mysql://)

	// Legacy individual parameters (deprecated, but maintained for backwards compatibility)
	Type     string `json:"type" envconfig:"DB_TYPE"`         // Database type: "mysql" or "postgres"
	Host     string `json:"host" envconfig:"DB_HOST"`         // Database server hostname
	Port     int    `json:"port" envconfig:"DB_PORT"`         // Database server port
	Database string `json:"database" envconfig:"DB_NAME"`     // Primary database name to connect to
	Username string `json:"username" envconfig:"DB_USER"`     // Database username
	Password string `json:"password" envconfig:"DB_PASSWORD"` // Database password
	SSLMode  string `json:"ssl_mode" envconfig:"DB_SSL_MODE"` // SSL/TLS mode: "none", "prefer", or "require"

	// Additional configuration (applies to both approaches)
	AllowedDatabases []string `json:"allowed_databases" envconfig:"DB_ALLOWED_NAMES"` // List of allowed database names (empty means all allowed)
	MaxConns         int      `json:"max_conns" envconfig:"DB_MAX_CONNS"`             // Maximum number of open connections
	MaxIdleConns     int      `json:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS"`   // Maximum number of idle connections
}

// IsDatabaseAllowed checks if a database name is allowed to be accessed.
// If AllowedDatabases is empty, only the primary database (DB_NAME) is allowed.
// If AllowedDatabases is specified, only those databases plus the primary database are allowed.
func (cfg *DatabaseConfig) IsDatabaseAllowed(databaseName string) bool {
	// Always allow the primary database
	if databaseName == cfg.Database {
		return true
	}

	// If no additional databases specified, only allow primary database
	if len(cfg.AllowedDatabases) == 0 {
		return false
	}

	// Check if database is in additional allowed list
	return slices.Contains(cfg.AllowedDatabases, databaseName)
}

// ValidateSSLMode checks if the configured SSL mode is valid and returns
// the parsed SSLMode. If no SSL mode is configured, it returns SSLModePrefer as default.
func (cfg *DatabaseConfig) ValidateSSLMode() (SSLMode, error) {
	if cfg.SSLMode == "" {
		return SSLModePrefer, nil
	}

	return ParseSSLMode(cfg.SSLMode)
}

// ApplyConnectionStringDefaults parses the connection string and uses it to populate
// any individual configuration fields that are still at their default values.
// Individual parameters take precedence over connection string values when both are provided.
// This function should be called before environment variable processing to ensure
// env vars can override connection string parameters.
func (cfg *DatabaseConfig) ApplyConnectionStringDefaults() error {
	if cfg.ConnectionString == "" {
		return nil // No connection string provided
	}

	connInfo, err := ParseConnectionString(cfg.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Only apply connection string values for fields that are empty (not set by env vars)
	if cfg.Type == "" {
		cfg.Type = connInfo.Type
	}
	if cfg.Host == "" {
		cfg.Host = connInfo.Host
	}
	if cfg.Port == 0 {
		cfg.Port = connInfo.Port
	}
	if cfg.Database == "" {
		cfg.Database = connInfo.Database
	}
	if cfg.Username == "" {
		cfg.Username = connInfo.Username
	}
	if cfg.Password == "" {
		cfg.Password = connInfo.Password
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = connInfo.SSLMode
	}

	return nil
}
