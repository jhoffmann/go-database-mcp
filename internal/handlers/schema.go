// Package handlers provides MCP tool handlers for database operations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jhoffmann/go-database-mcp/internal/config"
	"github.com/jhoffmann/go-database-mcp/internal/database"
)

// SchemaHandler handles database schema inspection tools.
type SchemaHandler struct {
	db     database.Database
	config *config.DatabaseConfig
}

// TablesResult represents the result of listing tables.
type TablesResult struct {
	Tables []string `json:"tables"` // List of table names
	Count  int      `json:"count"`  // Number of tables
}

// DatabasesResult represents the result of listing databases.
type DatabasesResult struct {
	Databases []string `json:"databases"` // List of database names
	Count     int      `json:"count"`     // Number of databases
}

// TableSchemaResult represents the result of describing a table.
type TableSchemaResult struct {
	Schema *database.TableSchema `json:"schema"` // Complete table schema
}

// TableDataResult represents the result of getting table data.
type TableDataResult struct {
	Data *database.TableData `json:"data"` // Table data with pagination info
}

// ExplainResult represents the result of explaining a query.
type ExplainResult struct {
	Query string `json:"query"` // The original query
	Plan  string `json:"plan"`  // Query execution plan (JSON format)
}

// NewSchemaHandler creates a new SchemaHandler instance.
func NewSchemaHandler(db database.Database, config *config.DatabaseConfig) *SchemaHandler {
	return &SchemaHandler{
		db:     db,
		config: config,
	}
}

// ListTables retrieves all table names from the current database.
func (h *SchemaHandler) ListTables(ctx context.Context) (*TablesResult, error) {
	tables, err := h.db.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	return &TablesResult{
		Tables: tables,
		Count:  len(tables),
	}, nil
}

// ListDatabases retrieves all available database names on the server.
// Only returns databases that are allowed by the configuration.
func (h *SchemaHandler) ListDatabases(ctx context.Context) (*DatabasesResult, error) {
	databases, err := h.db.ListDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	// Filter databases based on allowed list
	var allowedDatabases []string
	for _, dbName := range databases {
		if h.config.IsDatabaseAllowed(dbName) {
			allowedDatabases = append(allowedDatabases, dbName)
		}
	}

	return &DatabasesResult{
		Databases: allowedDatabases,
		Count:     len(allowedDatabases),
	}, nil
}

// DescribeTable retrieves detailed schema information about a specific table.
func (h *SchemaHandler) DescribeTable(ctx context.Context, tableName string) (*TableSchemaResult, error) {
	// Validate input
	if strings.TrimSpace(tableName) == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	schema, err := h.db.DescribeTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	return &TableSchemaResult{
		Schema: schema,
	}, nil
}

// GetTableData retrieves paginated data from a specific table.
func (h *SchemaHandler) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*TableDataResult, error) {
	// Validate input
	if strings.TrimSpace(tableName) == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit cannot be negative")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}

	// Set reasonable default and maximum limits
	if limit == 0 {
		limit = 100 // Default page size
	}
	if limit > 1000 {
		limit = 1000 // Maximum page size to prevent memory issues
	}

	data, err := h.db.GetTableData(ctx, tableName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get table data for %s: %w", tableName, err)
	}

	return &TableDataResult{
		Data: data,
	}, nil
}

// ExplainQuery retrieves the execution plan for a SQL query.
func (h *SchemaHandler) ExplainQuery(ctx context.Context, query string) (*ExplainResult, error) {
	// Validate input
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	plan, err := h.db.ExplainQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to explain query: %w", err)
	}

	return &ExplainResult{
		Query: query,
		Plan:  plan,
	}, nil
}

// GetTableStatistics provides statistical information about a table (if available).
func (h *SchemaHandler) GetTableStatistics(ctx context.Context, tableName string) (map[string]any, error) {
	// Validate input
	if strings.TrimSpace(tableName) == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	// This could be extended to provide table statistics like row count, size, etc.
	// For now, we'll use a simple query to get row count
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	row := h.db.QueryRow(ctx, query)
	var rowCount int64
	if err := row.Scan(&rowCount); err != nil {
		return nil, fmt.Errorf("failed to get table statistics for %s: %w", tableName, err)
	}

	return map[string]any{
		"table_name": tableName,
		"row_count":  rowCount,
	}, nil
}

// ValidateTableName performs basic validation on table names to prevent SQL injection.
func (h *SchemaHandler) ValidateTableName(tableName string) error {
	trimmed := strings.TrimSpace(tableName)

	if trimmed == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Basic validation to prevent obvious SQL injection attempts
	dangerous := []string{";", "--", "/*", "*/", "'", "\"", "\\", "DROP", "DELETE", "UPDATE", "INSERT"}
	upper := strings.ToUpper(trimmed)

	for _, danger := range dangerous {
		if strings.Contains(upper, strings.ToUpper(danger)) {
			return fmt.Errorf("table name contains potentially dangerous characters or keywords: %s", danger)
		}
	}

	return nil
}
