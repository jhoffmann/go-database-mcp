package handlers

import "github.com/jhoffmann/go-database-mcp/internal/config"

// createTestConfig returns a default test configuration for handler tests
func createTestConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Type:             "postgres",
		Host:             "localhost",
		Port:             5432,
		Database:         "testdb",
		AllowedDatabases: []string{}, // Empty means only primary database allowed
		Username:         "testuser",
		Password:         "testpass",
		SSLMode:          "disable",
	}
}
