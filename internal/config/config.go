// Package config provides configuration structures and loading functionality
// for the database MCP server.
package config

// Config represents the complete configuration for the database MCP server.
type Config struct {
	Database DatabaseConfig `json:"database"` // Database connection configuration
}

// DatabaseConfig contains all settings required to connect to a database.
// It supports both MySQL and PostgreSQL databases with SSL/TLS configuration.
type DatabaseConfig struct {
	Type             string   `json:"type" envconfig:"DB_TYPE"`                       // Database type: "mysql" or "postgres"
	Host             string   `json:"host" envconfig:"DB_HOST"`                       // Database server hostname
	Port             int      `json:"port" envconfig:"DB_PORT"`                       // Database server port
	Database         string   `json:"database" envconfig:"DB_NAME"`                   // Primary database name to connect to
	AllowedDatabases []string `json:"allowed_databases" envconfig:"DB_ALLOWED_NAMES"` // List of allowed database names (empty means all allowed)
	Username         string   `json:"username" envconfig:"DB_USER"`                   // Database username
	Password         string   `json:"password" envconfig:"DB_PASSWORD"`               // Database password
	MaxConns         int      `json:"max_conns" envconfig:"DB_MAX_CONNS"`             // Maximum number of open connections
	MaxIdleConns     int      `json:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS"`   // Maximum number of idle connections
	SSLMode          string   `json:"ssl_mode" envconfig:"DB_SSL_MODE"`               // SSL/TLS mode for connection
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
	for _, allowed := range cfg.AllowedDatabases {
		if allowed == databaseName {
			return true
		}
	}

	return false
}
