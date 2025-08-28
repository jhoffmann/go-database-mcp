package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

// PostgreSQL implements the Database interface for PostgreSQL database connections.
// It provides PostgreSQL-specific implementations of database operations including
// schema introspection, data access, and query execution with SSL support.
type PostgreSQL struct {
	db     *sql.DB               // The underlying database connection
	config config.DatabaseConfig // Configuration settings for the connection
}

// NewPostgreSQL creates a new PostgreSQL database instance with the given configuration.
// The connection is not established until Connect() is called.
func NewPostgreSQL(cfg config.DatabaseConfig) (*PostgreSQL, error) {
	return &PostgreSQL{
		config: cfg,
	}, nil
}

// Connect establishes a connection to the PostgreSQL database.
// It builds the DSN from configuration, opens the connection, configures the connection pool,
// and verifies connectivity with a ping. Returns an error if any step fails.
func (p *PostgreSQL) Connect(ctx context.Context) error {
	dsn := p.buildDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	configureConnectionPool(db, p.config)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	p.db = db
	return nil
}

// Close closes the PostgreSQL database connection and releases associated resources.
// It's safe to call even if no connection has been established.
func (p *PostgreSQL) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping verifies that the PostgreSQL database connection is still alive and accessible.
// Returns an error if no connection exists or if the database is unreachable.
func (p *PostgreSQL) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("no database connection")
	}
	return p.db.PingContext(ctx)
}

// Query executes a SQL query that returns rows, typically a SELECT statement.
// It supports parameter binding to prevent SQL injection attacks.
func (p *PostgreSQL) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if p.db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SQL query that is expected to return at most one row.
// It supports parameter binding to prevent SQL injection attacks.
func (p *PostgreSQL) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a SQL statement that doesn't return rows, such as INSERT, UPDATE, or DELETE.
// It supports parameter binding to prevent SQL injection attacks.
// Returns a Result containing information about the execution.
func (p *PostgreSQL) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if p.db == nil {
		return nil, fmt.Errorf("no database connection")
	}
	return p.db.ExecContext(ctx, query, args...)
}

// ListTables returns a list of all table names in the current PostgreSQL database.
// Queries the information_schema.tables view for tables in the 'public' schema.
func (p *PostgreSQL) ListTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`

	rows, err := p.Query(ctx, query)
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

// ListDatabases returns a list of all available database names on the PostgreSQL server.
// Queries the pg_database system catalog, excluding template databases.
func (p *PostgreSQL) ListDatabases(ctx context.Context) ([]string, error) {
	query := `
		SELECT datname 
		FROM pg_database 
		WHERE datistemplate = false
		ORDER BY datname`

	rows, err := p.Query(ctx, query)
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

// DescribeTable returns detailed schema information about the specified PostgreSQL table.
// It retrieves column definitions, data types, constraints, and index information
// using the information_schema views and system catalogs.
func (p *PostgreSQL) DescribeTable(ctx context.Context, tableName string) (*TableSchema, error) {
	schema := &TableSchema{
		TableName: tableName,
		Columns:   []ColumnInfo{},
		Indexes:   []IndexInfo{},
		Metadata:  make(map[string]any),
	}

	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
			CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END as is_auto_increment
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT k.column_name
			FROM information_schema.table_constraints t
			JOIN information_schema.key_column_usage k ON t.constraint_name = k.constraint_name
			WHERE t.constraint_type = 'PRIMARY KEY' 
				AND t.table_name = $1 AND k.table_name = $1
		) pk ON c.column_name = pk.column_name
		WHERE c.table_name = $1 AND c.table_schema = 'public'
		ORDER BY c.ordinal_position`

	rows, err := p.Query(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var column ColumnInfo
		var nullable string
		var defaultValue, maxLength sql.NullString
		var isPrimaryKey, isAutoIncrement bool

		err := rows.Scan(
			&column.Name,
			&column.Type,
			&nullable,
			&defaultValue,
			&maxLength,
			&isPrimaryKey,
			&isAutoIncrement,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		column.IsNullable = nullable == "YES"
		column.IsPrimaryKey = isPrimaryKey
		column.IsAutoIncrement = isAutoIncrement

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
			i.relname as index_name,
			array_agg(a.attname ORDER BY a.attnum) as column_names,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND t.relkind = 'r'
		GROUP BY i.relname, ix.indisunique, ix.indisprimary`

	indexRows, err := p.Query(ctx, indexQuery, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get index info: %w", err)
	}
	defer indexRows.Close()

	for indexRows.Next() {
		var index IndexInfo
		var columnArray string

		err := indexRows.Scan(&index.Name, &columnArray, &index.IsUnique, &index.IsPrimary)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index info: %w", err)
		}

		columnArray = strings.Trim(columnArray, "{}")
		index.Columns = strings.Split(columnArray, ",")

		schema.Indexes = append(schema.Indexes, index)
	}

	return schema, nil
}

// GetTableData retrieves data from the specified PostgreSQL table with pagination support.
// If limit is 0 or negative, it defaults to 100 rows. The method also returns
// the total row count for pagination purposes.
func (p *PostgreSQL) GetTableData(ctx context.Context, tableName string, limit int, offset int) (*TableData, error) {
	if limit <= 0 {
		limit = 100
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", tableName)
	var total int
	err := p.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	query := fmt.Sprintf("SELECT * FROM \"%s\" LIMIT $1 OFFSET $2", tableName)
	rows, err := p.Query(ctx, query, limit, offset)
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
// Uses PostgreSQL's EXPLAIN (FORMAT JSON) command to provide detailed query analysis.
func (p *PostgreSQL) ExplainQuery(ctx context.Context, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query)
	var result string
	err := p.QueryRow(ctx, explainQuery).Scan(&result)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}
	return result, nil
}

// GetDB returns the underlying *sql.DB instance for direct database operations.
// Returns nil if no connection has been established.
func (p *PostgreSQL) GetDB() *sql.DB {
	return p.db
}

// GetDriverName returns the name of the database driver.
// Always returns "postgres" for PostgreSQL connections.
func (p *PostgreSQL) GetDriverName() string {
	return "postgres"
}

// buildDSN constructs a PostgreSQL connection string from the configuration.
// It includes SSL configuration, timeout settings, and other connection parameters
// required for establishing a secure and reliable PostgreSQL connection.
func (p *PostgreSQL) buildDSN() string {
	var params []string

	params = append(params, fmt.Sprintf("host=%s", p.config.Host))
	params = append(params, fmt.Sprintf("port=%d", p.config.Port))
	params = append(params, fmt.Sprintf("user=%s", p.config.Username))
	params = append(params, fmt.Sprintf("password=%s", p.config.Password))
	params = append(params, fmt.Sprintf("dbname=%s", p.config.Database))

	// Handle SSL mode using common SSL configuration
	sslMode, err := p.config.ValidateSSLMode()
	if err != nil {
		// Default to none mode if invalid
		sslMode = config.SSLModeNone
	}

	postgresSSLMode, _ := sslMode.ToPostgreSQLSSLMode()
	params = append(params, fmt.Sprintf("sslmode=%s", postgresSSLMode))

	params = append(params, "connect_timeout=30")

	return strings.Join(params, " ")
}
