// Package database provides a unified interface for interacting with MySQL and PostgreSQL databases.
package database

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations that must be implemented by all database drivers.
// It provides a unified API for connecting to, querying, and inspecting database schemas.
type Database interface {
	// Connect establishes a connection to the database using the provided context.
	// It returns an error if the connection cannot be established.
	Connect(ctx context.Context) error

	// Close closes the database connection and releases associated resources.
	// It returns an error if the connection cannot be closed properly.
	Close() error

	// Ping verifies the database connection is still alive and accessible.
	// It returns an error if the database is unreachable.
	Ping(ctx context.Context) error

	// Query executes a SQL query that returns rows, typically a SELECT statement.
	// It accepts a query string and optional arguments for parameter binding.
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRow executes a SQL query that is expected to return at most one row.
	// It accepts a query string and optional arguments for parameter binding.
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row

	// Exec executes a SQL statement that doesn't return rows, such as INSERT, UPDATE, or DELETE.
	// It returns a Result containing information about the execution.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// ListTables returns a list of all table names in the current database.
	ListTables(ctx context.Context) ([]string, error)

	// ListDatabases returns a list of all available database names on the server.
	ListDatabases(ctx context.Context) ([]string, error)

	// DescribeTable returns detailed schema information about the specified table,
	// including column definitions, indexes, and metadata.
	DescribeTable(ctx context.Context, tableName string) (*TableSchema, error)

	// GetTableData retrieves data from the specified table with pagination support.
	// The limit parameter controls how many rows to return, and offset specifies how many rows to skip.
	GetTableData(ctx context.Context, tableName string, limit int, offset int) (*TableData, error)

	// ExplainQuery returns the execution plan for the given SQL query in JSON format.
	ExplainQuery(ctx context.Context, query string) (string, error)

	// GetDB returns the underlying *sql.DB instance for direct database operations.
	GetDB() *sql.DB

	// GetDriverName returns the name of the database driver (e.g., "mysql", "postgres").
	GetDriverName() string
}

// TableSchema represents the complete schema definition of a database table.
type TableSchema struct {
	TableName string         `json:"table_name"`         // Name of the table
	Columns   []ColumnInfo   `json:"columns"`            // List of column definitions
	Indexes   []IndexInfo    `json:"indexes,omitempty"`  // List of indexes on the table
	Metadata  map[string]any `json:"metadata,omitempty"` // Additional metadata about the table
}

// ColumnInfo represents detailed information about a database table column.
type ColumnInfo struct {
	Name            string  `json:"name"`                 // Column name
	Type            string  `json:"type"`                 // Data type (e.g., "VARCHAR", "INT")
	IsNullable      bool    `json:"is_nullable"`          // Whether the column allows NULL values
	DefaultValue    *string `json:"default_value"`        // Default value for the column, if any
	IsPrimaryKey    bool    `json:"is_primary_key"`       // Whether this column is part of the primary key
	IsAutoIncrement bool    `json:"is_auto_increment"`    // Whether this column auto-increments
	MaxLength       *int    `json:"max_length,omitempty"` // Maximum length for string types
}

// IndexInfo represents information about a database table index.
type IndexInfo struct {
	Name      string   `json:"name"`       // Index name
	Columns   []string `json:"columns"`    // List of columns that make up the index
	IsUnique  bool     `json:"is_unique"`  // Whether the index enforces uniqueness
	IsPrimary bool     `json:"is_primary"` // Whether this is the primary key index
}

// TableData represents paginated data from a database table.
type TableData struct {
	TableName string           `json:"table_name"` // Name of the table
	Columns   []string         `json:"columns"`    // Column names in the result set
	Rows      []map[string]any `json:"rows"`       // Actual row data as key-value pairs
	Total     int              `json:"total"`      // Total number of rows in the table
	Limit     int              `json:"limit"`      // Number of rows returned in this batch
	Offset    int              `json:"offset"`     // Number of rows skipped from the beginning
}
