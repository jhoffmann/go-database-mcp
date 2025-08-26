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
	Type         string `json:"type" envconfig:"DB_TYPE"`                     // Database type: "mysql" or "postgres"
	Host         string `json:"host" envconfig:"DB_HOST"`                     // Database server hostname
	Port         int    `json:"port" envconfig:"DB_PORT"`                     // Database server port
	Database     string `json:"database" envconfig:"DB_NAME"`                 // Database name to connect to
	Username     string `json:"username" envconfig:"DB_USER"`                 // Database username
	Password     string `json:"password" envconfig:"DB_PASSWORD"`             // Database password
	MaxConns     int    `json:"max_conns" envconfig:"DB_MAX_CONNS"`           // Maximum number of open connections
	MaxIdleConns int    `json:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS"` // Maximum number of idle connections
	SSLMode      string `json:"ssl_mode" envconfig:"DB_SSL_MODE"`             // SSL/TLS mode for connection
}
