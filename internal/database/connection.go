package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jhoffmann/go-database-mcp/internal/config"
	_ "github.com/lib/pq"
)

// Manager handles database connections and provides a factory for creating database instances.
// It supports both MySQL and PostgreSQL databases with connection pooling and SSL configuration.
type Manager struct {
	config   config.DatabaseConfig // Database configuration settings
	database Database              // Active database connection instance
}

// NewManager creates a new database manager with the given configuration.
// It validates the configuration but does not establish a connection until Connect is called.
// Returns an error if the configuration is invalid.
func NewManager(cfg config.DatabaseConfig) (*Manager, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	return &Manager{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the database based on the configured database type.
// It creates the appropriate database instance (MySQL or PostgreSQL) and connects to it.
// Returns an error if the database type is unsupported or if the connection fails.
func (m *Manager) Connect(ctx context.Context) error {
	var db Database
	var err error

	switch m.config.Type {
	case "mysql":
		db, err = NewMySQL(m.config)
	case "postgres":
		db, err = NewPostgreSQL(m.config)
	default:
		return fmt.Errorf("unsupported database type: %s", m.config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create database instance: %w", err)
	}

	if err := db.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	m.database = db
	return nil
}

// GetDatabase returns the active database connection instance.
// Returns nil if no connection has been established yet.
func (m *Manager) GetDatabase() Database {
	return m.database
}

// Close closes the database connection and releases associated resources.
// It's safe to call even if no connection has been established.
func (m *Manager) Close() error {
	if m.database != nil {
		return m.database.Close()
	}
	return nil
}

// Ping verifies the database connection is still alive and accessible.
// Returns an error if no connection has been established or if the database is unreachable.
func (m *Manager) Ping(ctx context.Context) error {
	if m.database == nil {
		return fmt.Errorf("no database connection established")
	}
	return m.database.Ping(ctx)
}

// validateConfig validates the database configuration settings.
// It checks that all required fields are present and that the database type is supported.
// Returns an error describing any validation failures.
func validateConfig(cfg config.DatabaseConfig) error {
	if cfg.Type == "" {
		return fmt.Errorf("database type is required")
	}
	if cfg.Type != "mysql" && cfg.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
	if cfg.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("database port is required")
	}
	if cfg.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Username == "" {
		return fmt.Errorf("database username is required")
	}

	return nil
}

// configureConnectionPool sets up connection pooling parameters for the database connection.
// It uses configuration values if provided, otherwise applies sensible defaults:
// - MaxOpenConns: 25 connections
// - MaxIdleConns: 5 connections
// - ConnMaxLifetime: 5 minutes
// - ConnMaxIdleTime: 30 seconds
func configureConnectionPool(db *sql.DB, cfg config.DatabaseConfig) {
	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(cfg.MaxConns)
	} else {
		db.SetMaxOpenConns(25)
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(5)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)
}
