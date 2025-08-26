package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

// MySQL implements the Database interface for MySQL database connections.
// It provides MySQL-specific implementations of database operations including
// schema introspection, data access, and query execution with SSL support.
type MySQL struct {
	db     *sql.DB               // The underlying database connection
	config config.DatabaseConfig // Configuration settings for the connection
}

// NewMySQL creates a new MySQL database instance with the given configuration.
// The connection is not established until Connect() is called.
func NewMySQL(cfg config.DatabaseConfig) (*MySQL, error) {
	return &MySQL{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the MySQL database.
// It builds the DSN from configuration, opens the connection, configures the connection pool,
// and verifies connectivity with a ping. Returns an error if any step fails.
func (m *MySQL) Connect(ctx context.Context) error {
	dsn := m.buildDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	configureConnectionPool(db, m.config)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	m.db = db
	return nil
}

// Close closes the MySQL database connection and releases associated resources.
// It's safe to call even if no connection has been established.
func (m *MySQL) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// Ping verifies that the MySQL database connection is still alive and accessible.
// Returns an error if no connection exists or if the database is unreachable.
func (m *MySQL) Ping(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("no database connection")
	}
	return m.db.PingContext(ctx)
}

// Query executes a SQL query that returns rows, typically a SELECT statement.
// It supports parameter binding to prevent SQL injection attacks.
func (m *MySQL) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if m.db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return m.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SQL query that is expected to return at most one row.
// It supports parameter binding to prevent SQL injection attacks.
func (m *MySQL) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a SQL statement that doesn't return rows, such as INSERT, UPDATE, or DELETE.
// It supports parameter binding to prevent SQL injection attacks.
// Returns a Result containing information about the execution.
func (m *MySQL) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if m.db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return m.db.ExecContext(ctx, query, args...)
}

// ListTables returns a list of all table names in the current MySQL database.
// Uses the SHOW TABLES command to retrieve table names.
func (m *MySQL) ListTables(ctx context.Context) ([]string, error) {
	query := "SHOW TABLES"
	rows, err := m.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// ListDatabases returns a list of all available database names on the MySQL server.
// Uses the SHOW DATABASES command to retrieve database names.
func (m *MySQL) ListDatabases(ctx context.Context) ([]string, error) {
	query := "SHOW DATABASES"
	rows, err := m.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %w", err)
		}
		databases = append(databases, dbName)
	}

	return databases, rows.Err()
}

// DescribeTable returns detailed schema information about the specified MySQL table.
// It retrieves column definitions, data types, constraints, and index information
// using the INFORMATION_SCHEMA tables.
func (m *MySQL) DescribeTable(ctx context.Context, tableName string) (*TableSchema, error) {
	schema := &TableSchema{
		TableName: tableName,
		Columns:   []ColumnInfo{},
		Indexes:   []IndexInfo{},
		Metadata:  make(map[string]any),
	}

	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY,
			EXTRA,
			CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := m.Query(ctx, query, m.config.Database, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var column ColumnInfo
		var nullable, columnKey, extra string
		var defaultValue, maxLength sql.NullString

		err := rows.Scan(
			&column.Name,
			&column.Type,
			&nullable,
			&defaultValue,
			&columnKey,
			&extra,
			&maxLength,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		column.IsNullable = nullable == "YES"
		column.IsPrimaryKey = columnKey == "PRI"
		column.IsAutoIncrement = strings.Contains(extra, "auto_increment")

		if defaultValue.Valid {
			column.DefaultValue = &defaultValue.String
		}

		if maxLength.Valid {
			if length, err := strconv.Atoi(maxLength.String); err == nil {
				column.MaxLength = &length
			}
		}

		schema.Columns = append(schema.Columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading column data: %w", err)
	}

	indexQuery := `
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE
		FROM INFORMATION_SCHEMA.STATISTICS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX`

	indexRows, err := m.Query(ctx, indexQuery, m.config.Database, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get index info: %w", err)
	}
	defer indexRows.Close()

	indexMap := make(map[string]*IndexInfo)
	for indexRows.Next() {
		var indexName, columnName string
		var nonUnique int

		err := indexRows.Scan(&indexName, &columnName, &nonUnique)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index info: %w", err)
		}

		if index, exists := indexMap[indexName]; exists {
			index.Columns = append(index.Columns, columnName)
		} else {
			indexMap[indexName] = &IndexInfo{
				Name:      indexName,
				Columns:   []string{columnName},
				IsUnique:  nonUnique == 0,
				IsPrimary: indexName == "PRIMARY",
			}
		}
	}

	for _, index := range indexMap {
		schema.Indexes = append(schema.Indexes, *index)
	}

	return schema, nil
}

// GetTableData retrieves data from the specified MySQL table with pagination support.
// If limit is 0 or negative, it defaults to 100 rows. The method also returns
// the total row count for pagination purposes.
func (m *MySQL) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*TableData, error) {
	if limit <= 0 {
		limit = 100
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	var total int
	err := m.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	query := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", tableName)
	rows, err := m.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	data := &TableData{
		TableName: tableName,
		Columns:   columns,
		Rows:      []map[string]any{},
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			if values[i] != nil {
				row[col] = values[i]
			} else {
				row[col] = nil
			}
		}
		data.Rows = append(data.Rows, row)
	}

	return data, rows.Err()
}

// ExplainQuery returns the execution plan for the given SQL query in JSON format.
// Uses MySQL's EXPLAIN FORMAT=JSON command to provide detailed query analysis.
func (m *MySQL) ExplainQuery(ctx context.Context, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
	var result string
	err := m.QueryRow(ctx, explainQuery).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}
	return result, nil
}

// GetDB returns the underlying *sql.DB instance for direct database operations.
// Returns nil if no connection has been established.
func (m *MySQL) GetDB() *sql.DB {
	return m.db
}

// GetDriverName returns the name of the database driver.
// Always returns "mysql" for MySQL connections.
func (m *MySQL) GetDriverName() string {
	return "mysql"
}

// buildDSN constructs a MySQL Data Source Name (DSN) from the configuration.
// It includes SSL configuration, timeout settings, and other connection parameters
// required for establishing a secure and reliable MySQL connection.
func (m *MySQL) buildDSN() string {
	var params []string

	if m.config.SSLMode != "" {
		switch m.config.SSLMode {
		case "none":
			params = append(params, "tls=false")
		case "required":
			params = append(params, "tls=true")
		case "preferred":
			params = append(params, "tls=preferred")
		default:
			params = append(params, "tls=true")
		}
	}

	params = append(params, "parseTime=true")
	params = append(params, "timeout=30s")
	params = append(params, "readTimeout=30s")
	params = append(params, "writeTimeout=30s")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		m.config.Username,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Database,
	)

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}
