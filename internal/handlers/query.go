// Package handlers provides MCP tool handlers for database operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/jhoffmann/go-database-mcp/internal/config"
	"github.com/jhoffmann/go-database-mcp/internal/database"
	"github.com/jhoffmann/go-database-mcp/internal/security"
)

// QueryHandler handles SQL query execution tools.
type QueryHandler struct {
	db        database.Database
	validator *security.QueryValidator
}

// QueryResult represents the result of a SQL query execution.
type QueryResult struct {
	Type          string           `json:"type"`                     // Query type: select, insert, update, delete, ddl
	Columns       []string         `json:"columns,omitempty"`        // Column names for SELECT queries
	Rows          []map[string]any `json:"rows,omitempty"`           // Result rows for SELECT queries
	RowCount      int              `json:"row_count"`                // Number of rows returned (SELECT) or affected (INSERT/UPDATE/DELETE)
	RowsAffected  int64            `json:"rows_affected,omitempty"`  // Number of rows affected by the query
	LastInsertID  *int64           `json:"last_insert_id,omitempty"` // Last insert ID for INSERT queries
	ExecutionTime string           `json:"execution_time,omitempty"` // Query execution time
	Message       string           `json:"message,omitempty"`        // Success/info message
}

// NewQueryHandler creates a new QueryHandler instance.
func NewQueryHandler(db database.Database, config *config.DatabaseConfig) *QueryHandler {
	return &QueryHandler{
		db:        db,
		validator: security.NewQueryValidator(config),
	}
}

// ExecuteQuery executes a SQL query and returns formatted results.
// It supports both SELECT queries (which return data) and non-SELECT queries (INSERT, UPDATE, DELETE, DDL).
func (h *QueryHandler) ExecuteQuery(ctx context.Context, query string, args ...any) (*QueryResult, error) {
	// Security validation
	if err := h.validator.ValidateQuery(query); err != nil {
		return nil, h.validator.SanitizeErrorMessage(err)
	}

	// Validate query
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Determine query type
	queryType := h.determineQueryType(trimmedQuery)

	// Execute based on query type
	if queryType == "select" {
		return h.executeSelectQuery(ctx, query, args...)
	}

	return h.executeNonSelectQuery(ctx, query, queryType, args...)
}

// executeSelectQuery handles SELECT queries that return rows.
func (h *QueryHandler) executeSelectQuery(ctx context.Context, query string, args ...any) (*QueryResult, error) {
	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	// Process rows
	var resultRows []map[string]any
	for rows.Next() {
		// Create slice of interface{} for Scan
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan row values
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to map
		rowMap := make(map[string]any)
		for i, col := range columns {
			// Handle byte slices (common for text fields in some drivers)
			if b, ok := values[i].([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = values[i]
			}
		}
		resultRows = append(resultRows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &QueryResult{
		Type:     "select",
		Columns:  columns,
		Rows:     resultRows,
		RowCount: len(resultRows),
		Message:  fmt.Sprintf("Query executed successfully. %d rows returned.", len(resultRows)),
	}, nil
}

// executeNonSelectQuery handles INSERT, UPDATE, DELETE, and DDL queries.
func (h *QueryHandler) executeNonSelectQuery(ctx context.Context, query string, queryType string, args ...any) (*QueryResult, error) {
	result, err := h.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	queryResult := &QueryResult{
		Type:         queryType,
		RowsAffected: rowsAffected,
		RowCount:     int(rowsAffected),
	}

	// For INSERT queries, try to get the last insert ID
	if queryType == "insert" {
		if lastID, err := result.LastInsertId(); err == nil && lastID > 0 {
			queryResult.LastInsertID = &lastID
		}
	}

	// Set appropriate message
	switch queryType {
	case "insert":
		if queryResult.LastInsertID != nil {
			queryResult.Message = fmt.Sprintf("INSERT executed successfully. %d rows affected. Last insert ID: %d", rowsAffected, *queryResult.LastInsertID)
		} else {
			queryResult.Message = fmt.Sprintf("INSERT executed successfully. %d rows affected.", rowsAffected)
		}
	case "update":
		queryResult.Message = fmt.Sprintf("UPDATE executed successfully. %d rows affected.", rowsAffected)
	case "delete":
		queryResult.Message = fmt.Sprintf("DELETE executed successfully. %d rows affected.", rowsAffected)
	case "ddl":
		queryResult.Message = "DDL statement executed successfully."
	default:
		queryResult.Message = "Query executed successfully."
	}

	return queryResult, nil
}

// determineQueryType determines the type of SQL query based on its content.
func (h *QueryHandler) determineQueryType(query string) string {
	// Normalize query for analysis
	normalized := strings.ToUpper(strings.TrimSpace(query))

	// Remove leading comments and whitespace
	normalized = regexp.MustCompile(`^\s*(--[^\n]*\n\s*)*`).ReplaceAllString(normalized, "")
	normalized = regexp.MustCompile(`^\s*(/\*.*?\*/\s*)*`).ReplaceAllString(normalized, "")

	// Determine query type by first keyword
	if strings.HasPrefix(normalized, "SELECT") || strings.HasPrefix(normalized, "WITH") {
		return "select"
	}
	if strings.HasPrefix(normalized, "INSERT") {
		return "insert"
	}
	if strings.HasPrefix(normalized, "UPDATE") {
		return "update"
	}
	if strings.HasPrefix(normalized, "DELETE") {
		return "delete"
	}

	// DDL statements
	ddlKeywords := []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME"}
	for _, keyword := range ddlKeywords {
		if strings.HasPrefix(normalized, keyword) {
			return "ddl"
		}
	}

	// Default to ddl for any other statements
	return "ddl"
}

// FormatResult formats the query result in the specified format.
func (h *QueryHandler) FormatResult(result QueryResult, format string) (string, error) {
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal result to JSON: %w", err)
		}
		return string(jsonData), nil

	case "table":
		return h.formatAsTable(result)

	default:
		return "", fmt.Errorf("unsupported format: %s. Supported formats: json, table", format)
	}
}

// formatAsTable formats SELECT results as an ASCII table.
func (h *QueryHandler) formatAsTable(result QueryResult) (string, error) {
	if result.Type != "select" || len(result.Rows) == 0 {
		if result.Message != "" {
			return result.Message, nil
		}
		return fmt.Sprintf("Query executed successfully (%s). No rows to display.", result.Type), nil
	}

	var output strings.Builder
	writer := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)

	// Write headers
	fmt.Fprintln(writer, strings.Join(result.Columns, "\t"))

	// Write separator
	separator := make([]string, len(result.Columns))
	for i := range separator {
		separator[i] = strings.Repeat("-", 10) // Fixed width for simplicity
	}
	fmt.Fprintln(writer, strings.Join(separator, "\t"))

	// Write rows
	for _, row := range result.Rows {
		values := make([]string, len(result.Columns))
		for i, col := range result.Columns {
			if val := row[col]; val != nil {
				values[i] = fmt.Sprintf("%v", val)
			} else {
				values[i] = "<NULL>"
			}
		}
		fmt.Fprintln(writer, strings.Join(values, "\t"))
	}

	writer.Flush()

	// Add summary
	fmt.Fprintf(&output, "\n%d rows returned.\n", result.RowCount)

	return output.String(), nil
}

// ValidateQuery performs basic validation on SQL queries to prevent dangerous operations.
func (h *QueryHandler) ValidateQuery(query string) error {
	normalized := strings.ToUpper(strings.TrimSpace(query))

	// Check for empty query
	if normalized == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// List of potentially dangerous operations that might be restricted
	// Note: This is basic validation - production systems should implement more sophisticated checks
	dangerousPatterns := []string{
		// These might be uncommented in production for additional safety
		// "DROP DATABASE",
		// "DROP SCHEMA",
		// "TRUNCATE",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(normalized, pattern) {
			return fmt.Errorf("potentially dangerous operation detected: %s", pattern)
		}
	}

	return nil
}
