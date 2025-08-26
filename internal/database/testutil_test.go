package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	ConnectFunc       func(ctx context.Context) error
	CloseFunc         func() error
	PingFunc          func(ctx context.Context) error
	QueryFunc         func(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowFunc      func(ctx context.Context, query string, args ...any) *sql.Row
	ExecFunc          func(ctx context.Context, query string, args ...any) (sql.Result, error)
	ListTablesFunc    func(ctx context.Context) ([]string, error)
	ListDatabasesFunc func(ctx context.Context) ([]string, error)
	DescribeTableFunc func(ctx context.Context, tableName string) (*TableSchema, error)
	GetTableDataFunc  func(ctx context.Context, tableName string, limit int, offset int) (*TableData, error)
	ExplainQueryFunc  func(ctx context.Context, query string) (string, error)
	GetDBFunc         func() *sql.DB
	GetDriverNameFunc func() string

	// State tracking
	Connected  bool
	Closed     bool
	PingCount  int
	QueryCount int
	ExecCount  int
}

func (m *MockDatabase) Connect(ctx context.Context) error {
	if m.ConnectFunc != nil {
		err := m.ConnectFunc(ctx)
		if err == nil {
			m.Connected = true
		}
		return err
	}
	m.Connected = true
	return nil
}

func (m *MockDatabase) Close() error {
	if m.CloseFunc != nil {
		err := m.CloseFunc()
		if err == nil {
			m.Closed = true
		}
		return err
	}
	m.Closed = true
	return nil
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	m.PingCount++
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	if !m.Connected {
		return fmt.Errorf("no database connection")
	}
	return nil
}

func (m *MockDatabase) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	m.QueryCount++
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return nil, fmt.Errorf("mock query not implemented")
}

func (m *MockDatabase) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	// Return a mock Row that will return an error when scanned
	return &sql.Row{}
}

func (m *MockDatabase) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	m.ExecCount++
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return &MockResult{RowsAffectedValue: 1}, nil
}

func (m *MockDatabase) ListTables(ctx context.Context) ([]string, error) {
	if m.ListTablesFunc != nil {
		return m.ListTablesFunc(ctx)
	}
	return []string{"table1", "table2"}, nil
}

func (m *MockDatabase) ListDatabases(ctx context.Context) ([]string, error) {
	if m.ListDatabasesFunc != nil {
		return m.ListDatabasesFunc(ctx)
	}
	return []string{"db1", "db2"}, nil
}

func (m *MockDatabase) DescribeTable(ctx context.Context, tableName string) (*TableSchema, error) {
	if m.DescribeTableFunc != nil {
		return m.DescribeTableFunc(ctx, tableName)
	}
	return &TableSchema{
		TableName: tableName,
		Columns: []ColumnInfo{
			{Name: "id", Type: "INTEGER", IsPrimaryKey: true},
			{Name: "name", Type: "VARCHAR", IsNullable: true},
		},
	}, nil
}

func (m *MockDatabase) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*TableData, error) {
	if m.GetTableDataFunc != nil {
		return m.GetTableDataFunc(ctx, tableName, limit, offset)
	}
	return &TableData{
		TableName: tableName,
		Columns:   []string{"id", "name"},
		Rows: []map[string]any{
			{"id": 1, "name": "test1"},
			{"id": 2, "name": "test2"},
		},
		Total:  2,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (m *MockDatabase) ExplainQuery(ctx context.Context, query string) (string, error) {
	if m.ExplainQueryFunc != nil {
		return m.ExplainQueryFunc(ctx, query)
	}
	return `{"query_plan": "mock"}`, nil
}

func (m *MockDatabase) GetDB() *sql.DB {
	if m.GetDBFunc != nil {
		return m.GetDBFunc()
	}
	return nil
}

func (m *MockDatabase) GetDriverName() string {
	if m.GetDriverNameFunc != nil {
		return m.GetDriverNameFunc()
	}
	return "mock"
}

// MockResult implements sql.Result for testing
type MockResult struct {
	LastInsertIdValue int64
	RowsAffectedValue int64
	Error             error
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.LastInsertIdValue, m.Error
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.RowsAffectedValue, m.Error
}

// MockDriver implements driver.Driver for testing
type MockDriver struct {
	OpenFunc func(name string) (driver.Conn, error)
}

func (m *MockDriver) Open(name string) (driver.Conn, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc(name)
	}
	return &MockConn{}, nil
}

// MockConn implements driver.Conn for testing
type MockConn struct {
	PrepareFunc func(query string) (driver.Stmt, error)
	CloseFunc   func() error
	BeginFunc   func() (driver.Tx, error)
}

func (m *MockConn) Prepare(query string) (driver.Stmt, error) {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(query)
	}
	return &MockStmt{}, nil
}

func (m *MockConn) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockConn) Begin() (driver.Tx, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc()
	}
	return &MockTx{}, nil
}

// MockStmt implements driver.Stmt for testing
type MockStmt struct{}

func (m *MockStmt) Close() error                                    { return nil }
func (m *MockStmt) NumInput() int                                   { return 0 }
func (m *MockStmt) Exec(args []driver.Value) (driver.Result, error) { return &MockDriverResult{}, nil }
func (m *MockStmt) Query(args []driver.Value) (driver.Rows, error)  { return &MockRows{}, nil }

// MockTx implements driver.Tx for testing
type MockTx struct{}

func (m *MockTx) Commit() error   { return nil }
func (m *MockTx) Rollback() error { return nil }

// MockDriverResult implements driver.Result for testing
type MockDriverResult struct{}

func (m *MockDriverResult) LastInsertId() (int64, error) { return 1, nil }
func (m *MockDriverResult) RowsAffected() (int64, error) { return 1, nil }

// MockRows implements driver.Rows for testing
type MockRows struct {
	closed bool
}

func (m *MockRows) Columns() []string              { return []string{"id", "name"} }
func (m *MockRows) Close() error                   { m.closed = true; return nil }
func (m *MockRows) Next(dest []driver.Value) error { return fmt.Errorf("no more rows") }

// NewTestConfig returns a valid test configuration
func NewTestConfig(dbType string) config.DatabaseConfig {
	port := 5432
	if dbType == "mysql" {
		port = 3306
	}

	return config.DatabaseConfig{
		Type:         dbType,
		Host:         "localhost",
		Port:         port,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}
}
