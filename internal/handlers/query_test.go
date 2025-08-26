package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jhoffmann/go-database-mcp/internal/database"
)

// MockDatabase implements database.Database for testing
type MockDatabase struct {
	queryFunc         func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execFunc          func(ctx context.Context, query string, args ...any) (sql.Result, error)
	queryRowFunc      func(ctx context.Context, query string, args ...any) *sql.Row
	driver            string
	shouldReturnError bool
	errorMessage      string
}

func (m *MockDatabase) Connect(ctx context.Context) error                   { return nil }
func (m *MockDatabase) Close() error                                        { return nil }
func (m *MockDatabase) Ping(ctx context.Context) error                      { return nil }
func (m *MockDatabase) GetDB() *sql.DB                                      { return nil }
func (m *MockDatabase) GetDriverName() string                               { return m.driver }
func (m *MockDatabase) ListTables(ctx context.Context) ([]string, error)    { return nil, nil }
func (m *MockDatabase) ListDatabases(ctx context.Context) ([]string, error) { return nil, nil }
func (m *MockDatabase) DescribeTable(ctx context.Context, tableName string) (*database.TableSchema, error) {
	return nil, nil
}
func (m *MockDatabase) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*database.TableData, error) {
	return nil, nil
}
func (m *MockDatabase) ExplainQuery(ctx context.Context, query string) (string, error) {
	return "", nil
}

func (m *MockDatabase) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, args...)
	}
	return nil, errors.New("mock not configured")
}

func (m *MockDatabase) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, query, args...)
	}
	return nil
}

func (m *MockDatabase) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}
	if m.execFunc != nil {
		return m.execFunc(ctx, query, args...)
	}
	return &MockResult{rowsAffected: 1}, nil
}

// MockResult implements sql.Result
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
	err          error
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, m.err
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, m.err
}

func TestNewQueryHandler(t *testing.T) {
	mockDB := &MockDatabase{driver: "postgres"}

	handler := NewQueryHandler(mockDB, createTestConfig())

	if handler == nil {
		t.Fatal("NewQueryHandler returned nil")
	}

	if handler.db != mockDB {
		t.Error("QueryHandler database not set correctly")
	}
}

func TestQueryHandler_DetermineQueryType(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"SELECT * FROM users", "select"},
		{"  select id from table", "select"},
		{"INSERT INTO users VALUES (1, 'name')", "insert"},
		{"UPDATE users SET name = 'new'", "update"},
		{"DELETE FROM users WHERE id = 1", "delete"},
		{"CREATE TABLE test (id INT)", "ddl"},
		{"DROP TABLE test", "ddl"},
		{"ALTER TABLE users ADD COLUMN age INT", "ddl"},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", "select"},
		{"/* comment */ SELECT 1", "select"},
		{"-- comment\nSELECT 1", "select"},
	}

	handler := &QueryHandler{}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := handler.determineQueryType(tt.query)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for query: %s", tt.expected, result, tt.query)
			}
		})
	}
}

func TestQueryHandler_ExecuteQuery_NonSelect(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		args         []any
		rowsAffected int64
		lastInsertID int64
		wantType     string
		wantErr      bool
	}{
		{
			name:         "insert query",
			query:        "INSERT INTO users (name, email) VALUES (?, ?)",
			args:         []any{"John", "john@example.com"},
			rowsAffected: 1,
			lastInsertID: 42,
			wantType:     "insert",
			wantErr:      false,
		},
		{
			name:         "update query",
			query:        "UPDATE users SET email = ? WHERE id = ?",
			args:         []any{"newemail@example.com", 1},
			rowsAffected: 1,
			wantType:     "update",
			wantErr:      false,
		},
		{
			name:         "delete query",
			query:        "DELETE FROM users WHERE active = ?",
			args:         []any{false},
			rowsAffected: 3,
			wantType:     "delete",
			wantErr:      false,
		},
		{
			name:         "create table",
			query:        "CREATE TABLE test (id INT PRIMARY KEY)",
			args:         []any{},
			rowsAffected: 0,
			wantType:     "ddl",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResult := &MockResult{
				lastInsertID: tt.lastInsertID,
				rowsAffected: tt.rowsAffected,
			}
			mockDB := &MockDatabase{
				execFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
					return mockResult, nil
				},
				driver: "postgres",
			}

			handler := NewQueryHandler(mockDB, createTestConfig())
			result, err := handler.ExecuteQuery(context.Background(), tt.query, tt.args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Type != tt.wantType {
					t.Errorf("Expected type %s, got %s", tt.wantType, result.Type)
				}
				if result.RowsAffected != tt.rowsAffected {
					t.Errorf("Expected %d rows affected, got %d", tt.rowsAffected, result.RowsAffected)
				}
				if tt.wantType == "insert" && tt.lastInsertID > 0 {
					if result.LastInsertID == nil || *result.LastInsertID != tt.lastInsertID {
						t.Errorf("Expected last insert ID %d, got %v", tt.lastInsertID, result.LastInsertID)
					}
				}
				if result.Message == "" {
					t.Error("Expected non-empty message")
				}
			}
		})
	}
}

func TestQueryHandler_ExecuteQuery_Errors(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantErr   bool
		errString string
		setupMock func() *MockDatabase
	}{
		{
			name:      "empty query",
			query:     "",
			wantErr:   true,
			errString: "query cannot be empty",
			setupMock: func() *MockDatabase {
				return &MockDatabase{driver: "postgres"}
			},
		},
		{
			name:      "whitespace only query",
			query:     "   \n\t  ",
			wantErr:   true,
			errString: "query cannot be empty",
			setupMock: func() *MockDatabase {
				return &MockDatabase{driver: "postgres"}
			},
		},
		{
			name:      "database error",
			query:     "INSERT INTO users (id) VALUES (?)",
			wantErr:   true,
			errString: "query execution failed",
			setupMock: func() *MockDatabase {
				return &MockDatabase{
					shouldReturnError: true,
					errorMessage:      "duplicate key violation",
					driver:            "postgres",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.setupMock()

			handler := NewQueryHandler(mockDB, createTestConfig())
			_, err := handler.ExecuteQuery(context.Background(), tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !containsString(err.Error(), tt.errString) {
					t.Errorf("Expected error to contain %q, got %q", tt.errString, err.Error())
				}
			}
		})
	}
}

func TestQueryHandler_FormatResult_JSON(t *testing.T) {
	result := &QueryResult{
		Type:     "select",
		Columns:  []string{"id", "name"},
		RowCount: 2,
		Message:  "Test message",
	}

	handler := &QueryHandler{}
	formatted, err := handler.FormatResult(*result, "json")

	if err != nil {
		t.Fatalf("FormatResult() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(formatted), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	if parsed["type"] != "select" {
		t.Errorf("Expected type 'select', got %v", parsed["type"])
	}

	if parsed["row_count"] != float64(2) {
		t.Errorf("Expected row_count 2, got %v", parsed["row_count"])
	}
}

func TestQueryHandler_FormatResult_Table(t *testing.T) {
	result := &QueryResult{
		Type:    "select",
		Columns: []string{"id", "name"},
		Rows: []map[string]any{
			{"id": int64(1), "name": "Alice"},
			{"id": int64(2), "name": "Bob"},
		},
		RowCount: 2,
	}

	handler := &QueryHandler{}
	formatted, err := handler.FormatResult(*result, "table")

	if err != nil {
		t.Fatalf("FormatResult() error = %v", err)
	}

	if !containsString(formatted, "Alice") || !containsString(formatted, "Bob") {
		t.Errorf("Table format should contain row data")
	}

	if !containsString(formatted, "id") || !containsString(formatted, "name") {
		t.Errorf("Table format should contain column headers")
	}
}

func TestQueryHandler_FormatResult_NonSelectTable(t *testing.T) {
	result := &QueryResult{
		Type:    "insert",
		Message: "INSERT executed successfully",
	}

	handler := &QueryHandler{}
	formatted, err := handler.FormatResult(*result, "table")

	if err != nil {
		t.Fatalf("FormatResult() error = %v", err)
	}

	if !containsString(formatted, "INSERT executed successfully") {
		t.Errorf("Table format should contain message for non-SELECT queries")
	}
}

func TestQueryHandler_FormatResult_InvalidFormat(t *testing.T) {
	result := &QueryResult{
		Type:     "select",
		RowCount: 0,
	}

	handler := &QueryHandler{}
	_, err := handler.FormatResult(*result, "invalid")

	if err == nil {
		t.Error("Expected error for invalid format")
	}

	if !containsString(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got %v", err)
	}
}

func TestQueryHandler_Context_Timeout(t *testing.T) {
	// Test that query execution respects context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond) // Ensure context is expired

	mockDB := &MockDatabase{
		execFunc: func(ctx context.Context, query string, args ...any) (sql.Result, error) {
			// Check if context is cancelled
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return &MockResult{rowsAffected: 1}, nil
		},
		driver: "postgres",
	}
	handler := NewQueryHandler(mockDB, createTestConfig())

	_, err := handler.ExecuteQuery(ctx, "INSERT INTO test VALUES (1)")

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestQueryHandler_ValidateQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "valid select query",
			query:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "valid insert query",
			query:   "INSERT INTO users (name) VALUES ('test')",
			wantErr: false,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only query",
			query:   "   \n\t  ",
			wantErr: true,
		},
	}

	handler := &QueryHandler{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) <= len(s) && s[:len(substr)] == substr) ||
		(len(substr) <= len(s) && s[len(s)-len(substr):] == substr) ||
		(len(substr) < len(s) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}
