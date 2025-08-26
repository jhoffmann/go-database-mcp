package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/database"
)

// MockSchemaDatabase extends MockDatabase for schema operations
type MockSchemaDatabase struct {
	MockDatabase
	tables        []string
	databases     []string
	tableSchema   *database.TableSchema
	tableData     *database.TableData
	explainResult string
	listTablesErr error
	listDBErr     error
	describeErr   error
	tableDataErr  error
	explainErr    error
}

func (m *MockSchemaDatabase) ListTables(ctx context.Context) ([]string, error) {
	return m.tables, m.listTablesErr
}

func (m *MockSchemaDatabase) ListDatabases(ctx context.Context) ([]string, error) {
	return m.databases, m.listDBErr
}

func (m *MockSchemaDatabase) DescribeTable(ctx context.Context, tableName string) (*database.TableSchema, error) {
	return m.tableSchema, m.describeErr
}

func (m *MockSchemaDatabase) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*database.TableData, error) {
	return m.tableData, m.tableDataErr
}

func (m *MockSchemaDatabase) ExplainQuery(ctx context.Context, query string) (string, error) {
	return m.explainResult, m.explainErr
}

func TestNewSchemaHandler(t *testing.T) {
	mockDB := &MockSchemaDatabase{}

	handler := NewSchemaHandler(mockDB, createTestConfig())

	if handler == nil {
		t.Fatal("NewSchemaHandler returned nil")
	}

	if handler.db != mockDB {
		t.Error("SchemaHandler database not set correctly")
	}
}

func TestSchemaHandler_ListTables(t *testing.T) {
	tests := []struct {
		name      string
		tables    []string
		error     error
		wantErr   bool
		wantCount int
	}{
		{
			name:      "successful list with tables",
			tables:    []string{"users", "products", "orders"},
			error:     nil,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:      "empty database",
			tables:    []string{},
			error:     nil,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:      "database error",
			tables:    nil,
			error:     errors.New("database connection failed"),
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockSchemaDatabase{
				tables:        tt.tables,
				listTablesErr: tt.error,
			}
			mockDB.driver = "postgres"

			handler := NewSchemaHandler(mockDB, createTestConfig())
			result, err := handler.ListTables(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListTables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Tables) != tt.wantCount {
					t.Errorf("Expected %d tables, got %d", tt.wantCount, len(result.Tables))
				}

				for i, expectedTable := range tt.tables {
					if i < len(result.Tables) && result.Tables[i] != expectedTable {
						t.Errorf("Expected table %s, got %s", expectedTable, result.Tables[i])
					}
				}

				if result.Count != tt.wantCount {
					t.Errorf("Expected count %d, got %d", tt.wantCount, result.Count)
				}
			}
		})
	}
}

func TestSchemaHandler_ListDatabases(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		error     error
		wantErr   bool
		wantCount int
	}{
		{
			name:      "successful list with databases",
			databases: []string{"myapp", "test", "admin"},
			error:     nil,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:      "single database",
			databases: []string{"myapp"},
			error:     nil,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "database error",
			databases: nil,
			error:     errors.New("insufficient privileges"),
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockSchemaDatabase{
				databases: tt.databases,
				listDBErr: tt.error,
			}
			mockDB.driver = "postgres"

			// Create config that allows all test databases
			testConfig := createTestConfig()
			if len(tt.databases) > 0 {
				testConfig.Database = tt.databases[0]          // Set primary database to first test database
				testConfig.AllowedDatabases = tt.databases[1:] // Allow remaining databases
			}
			handler := NewSchemaHandler(mockDB, testConfig)
			result, err := handler.ListDatabases(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListDatabases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Databases) != tt.wantCount {
					t.Errorf("Expected %d databases, got %d", tt.wantCount, len(result.Databases))
				}

				for i, expectedDB := range tt.databases {
					if i < len(result.Databases) && result.Databases[i] != expectedDB {
						t.Errorf("Expected database %s, got %s", expectedDB, result.Databases[i])
					}
				}

				if result.Count != tt.wantCount {
					t.Errorf("Expected count %d, got %d", tt.wantCount, result.Count)
				}
			}
		})
	}
}

func TestSchemaHandler_DescribeTable(t *testing.T) {
	sampleSchema := &database.TableSchema{
		TableName: "users",
		Columns: []database.ColumnInfo{
			{
				Name:            "id",
				Type:            "INTEGER",
				IsNullable:      false,
				IsPrimaryKey:    true,
				IsAutoIncrement: true,
			},
			{
				Name:       "name",
				Type:       "VARCHAR",
				IsNullable: false,
				MaxLength:  ptr(255),
			},
			{
				Name:         "email",
				Type:         "VARCHAR",
				IsNullable:   true,
				DefaultValue: nil,
				MaxLength:    ptr(255),
			},
		},
		Indexes: []database.IndexInfo{
			{
				Name:      "PRIMARY",
				Columns:   []string{"id"},
				IsUnique:  true,
				IsPrimary: true,
			},
			{
				Name:     "idx_email",
				Columns:  []string{"email"},
				IsUnique: true,
			},
		},
	}

	tests := []struct {
		name        string
		tableName   string
		schema      *database.TableSchema
		error       error
		wantErr     bool
		wantColumns int
		wantIndexes int
	}{
		{
			name:        "successful describe",
			tableName:   "users",
			schema:      sampleSchema,
			error:       nil,
			wantErr:     false,
			wantColumns: 3,
			wantIndexes: 2,
		},
		{
			name:        "table not found",
			tableName:   "nonexistent",
			schema:      nil,
			error:       errors.New("table does not exist"),
			wantErr:     true,
			wantColumns: 0,
			wantIndexes: 0,
		},
		{
			name:        "empty table name",
			tableName:   "",
			schema:      nil,
			error:       nil,
			wantErr:     true,
			wantColumns: 0,
			wantIndexes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockSchemaDatabase{
				tableSchema: tt.schema,
				describeErr: tt.error,
			}
			mockDB.driver = "postgres"

			handler := NewSchemaHandler(mockDB, createTestConfig())
			result, err := handler.DescribeTable(context.Background(), tt.tableName)

			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Schema == nil {
					t.Fatal("Expected non-nil schema")
				}

				if len(result.Schema.Columns) != tt.wantColumns {
					t.Errorf("Expected %d columns, got %d", tt.wantColumns, len(result.Schema.Columns))
				}

				if len(result.Schema.Indexes) != tt.wantIndexes {
					t.Errorf("Expected %d indexes, got %d", tt.wantIndexes, len(result.Schema.Indexes))
				}

				if result.Schema.TableName != tt.tableName {
					t.Errorf("Expected table name %s, got %s", tt.tableName, result.Schema.TableName)
				}
			}
		})
	}
}

func TestSchemaHandler_GetTableData(t *testing.T) {
	sampleData := &database.TableData{
		TableName: "users",
		Columns:   []string{"id", "name", "email"},
		Rows: []map[string]any{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
		},
		Total:  100,
		Limit:  2,
		Offset: 0,
	}

	tests := []struct {
		name      string
		tableName string
		limit     int
		offset    int
		data      *database.TableData
		error     error
		wantErr   bool
		wantRows  int
	}{
		{
			name:      "successful get data",
			tableName: "users",
			limit:     2,
			offset:    0,
			data:      sampleData,
			error:     nil,
			wantErr:   false,
			wantRows:  2,
		},
		{
			name:      "empty result",
			tableName: "empty_table",
			limit:     10,
			offset:    0,
			data: &database.TableData{
				TableName: "empty_table",
				Columns:   []string{"id"},
				Rows:      []map[string]any{},
				Total:     0,
				Limit:     10,
				Offset:    0,
			},
			error:    nil,
			wantErr:  false,
			wantRows: 0,
		},
		{
			name:      "invalid table",
			tableName: "nonexistent",
			limit:     10,
			offset:    0,
			data:      nil,
			error:     errors.New("table does not exist"),
			wantErr:   true,
			wantRows:  0,
		},
		{
			name:      "invalid limit",
			tableName: "users",
			limit:     -1,
			offset:    0,
			data:      nil,
			error:     nil,
			wantErr:   true,
			wantRows:  0,
		},
		{
			name:      "invalid offset",
			tableName: "users",
			limit:     10,
			offset:    -1,
			data:      nil,
			error:     nil,
			wantErr:   true,
			wantRows:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockSchemaDatabase{
				tableData:    tt.data,
				tableDataErr: tt.error,
			}
			mockDB.driver = "postgres"

			handler := NewSchemaHandler(mockDB, createTestConfig())
			result, err := handler.GetTableData(context.Background(), tt.tableName, tt.limit, tt.offset)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTableData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Data == nil {
					t.Fatal("Expected non-nil data")
				}

				if len(result.Data.Rows) != tt.wantRows {
					t.Errorf("Expected %d rows, got %d", tt.wantRows, len(result.Data.Rows))
				}

				if result.Data.TableName != tt.tableName {
					t.Errorf("Expected table name %s, got %s", tt.tableName, result.Data.TableName)
				}

				if result.Data.Limit != tt.limit {
					t.Errorf("Expected limit %d, got %d", tt.limit, result.Data.Limit)
				}

				if result.Data.Offset != tt.offset {
					t.Errorf("Expected offset %d, got %d", tt.offset, result.Data.Offset)
				}
			}
		})
	}
}

func TestSchemaHandler_ExplainQuery(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		explainResult string
		error         error
		wantErr       bool
	}{
		{
			name:          "successful explain",
			query:         "SELECT * FROM users WHERE id = 1",
			explainResult: `{"Plan": {"Node Type": "Index Scan", "Relation Name": "users"}}`,
			error:         nil,
			wantErr:       false,
		},
		{
			name:          "complex query explain",
			query:         "SELECT u.name, COUNT(o.id) FROM users u JOIN orders o ON u.id = o.user_id GROUP BY u.name",
			explainResult: `{"Plan": {"Node Type": "HashAggregate", "Plans": [{"Node Type": "Hash Join"}]}}`,
			error:         nil,
			wantErr:       false,
		},
		{
			name:          "invalid query",
			query:         "SELECT * FROM nonexistent",
			explainResult: "",
			error:         errors.New("relation does not exist"),
			wantErr:       true,
		},
		{
			name:          "empty query",
			query:         "",
			explainResult: "",
			error:         nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockSchemaDatabase{
				explainResult: tt.explainResult,
				explainErr:    tt.error,
			}
			mockDB.driver = "postgres"

			handler := NewSchemaHandler(mockDB, createTestConfig())
			result, err := handler.ExplainQuery(context.Background(), tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExplainQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.Query != tt.query {
					t.Errorf("Expected query %s, got %s", tt.query, result.Query)
				}

				if result.Plan != tt.explainResult {
					t.Errorf("Expected plan %s, got %s", tt.explainResult, result.Plan)
				}
			}
		})
	}
}

// Helper function for creating pointers
func ptr[T any](v T) *T {
	return &v
}

func TestSchemaHandler_Validation(t *testing.T) {
	mockDB := &MockSchemaDatabase{}
	mockDB.driver = "postgres"
	handler := NewSchemaHandler(mockDB, createTestConfig())

	// Test table name validation
	_, err := handler.DescribeTable(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty table name")
	}

	// Test pagination validation
	_, err = handler.GetTableData(context.Background(), "users", -1, 0)
	if err == nil {
		t.Error("Expected error for negative limit")
	}

	_, err = handler.GetTableData(context.Background(), "users", 10, -1)
	if err == nil {
		t.Error("Expected error for negative offset")
	}

	// Test query validation
	_, err = handler.ExplainQuery(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty query")
	}
}
