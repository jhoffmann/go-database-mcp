package security

import (
	"strings"
	"testing"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

// Test helper to create test configurations
func createTestConfig(allowedDatabases []string) *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Type:             "postgres",
		Host:             "localhost",
		Port:             5432,
		Database:         "testdb",
		AllowedDatabases: allowedDatabases,
		Username:         "testuser",
		Password:         "testpass",
		SSLMode:          "disable",
	}
}

func TestQueryValidator_ValidateBasicSafety(t *testing.T) {
	validator := NewQueryValidator(createTestConfig(nil))

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
		{
			name:    "whitespace only query",
			query:   "   \n\t  ",
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
		{
			name:    "valid select query",
			query:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "SQL comment injection attempt",
			query:   "SELECT * FROM users; -- DROP TABLE users;",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
		{
			name:    "SQL block comment injection",
			query:   "SELECT * FROM users /* DROP TABLE users */ WHERE id = 1",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
		{
			name:    "dynamic SQL execution attempt",
			query:   "EXEC('DROP TABLE users')",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
		{
			name:    "system stored procedure attempt",
			query:   "SELECT * FROM users; EXEC SP_CONFIGURE",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
		{
			name:    "file system access attempt",
			query:   "SELECT LOAD_FILE('/etc/passwd')",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
		{
			name:    "outfile injection attempt",
			query:   "SELECT * FROM users INTO OUTFILE '/tmp/users.txt'",
			wantErr: true,
			errMsg:  "potentially dangerous pattern detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateBasicSafety(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateBasicSafety() expected error but got none")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("validateBasicSafety() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateBasicSafety() error = %v, want nil", err)
			}
		})
	}
}

func TestQueryValidator_ValidateDatabaseAccess(t *testing.T) {
	tests := []struct {
		name             string
		allowedDatabases []string
		query            string
		wantErr          bool
		errMsg           string
	}{
		{
			name:             "no restrictions - only allow primary database",
			allowedDatabases: nil,
			query:            "USE testdb; SELECT * FROM users",
			wantErr:          false,
		},
		{
			name:             "USE statement - allowed database",
			allowedDatabases: []string{"testdb", "devdb"},
			query:            "USE testdb",
			wantErr:          false,
		},
		{
			name:             "USE statement - disallowed database",
			allowedDatabases: []string{"testdb", "devdb"},
			query:            "USE production",
			wantErr:          true,
			errMsg:           "access denied: database 'production' is not in allowed databases list",
		},
		{
			name:             "qualified table name - primary database always allowed",
			allowedDatabases: []string{"devdb"},
			query:            "SELECT * FROM testdb.users",
			wantErr:          false,
		},
		{
			name:             "qualified table name - disallowed database",
			allowedDatabases: []string{"testdb", "devdb"},
			query:            "SELECT * FROM production.users",
			wantErr:          true,
			errMsg:           "access denied: database 'production' is not in allowed databases list",
		},
		{
			name:             "multiple qualified names - mixed allowed/disallowed",
			allowedDatabases: []string{},
			query:            "SELECT u.name FROM testdb.users u JOIN production.orders o ON u.id = o.user_id",
			wantErr:          true,
			errMsg:           "access denied: database 'production' is not in allowed databases list",
		},
		{
			name:             "INFORMATION_SCHEMA access",
			allowedDatabases: []string{},
			query:            "SELECT * FROM INFORMATION_SCHEMA.TABLES",
			wantErr:          false,
		},
		{
			name:             "system schema should be ignored",
			allowedDatabases: []string{},
			query:            "SELECT * FROM pg_catalog.pg_tables",
			wantErr:          false,
		},
		{
			name:             "table.column references should be allowed",
			allowedDatabases: []string{},
			query:            "SELECT users.id, users.name FROM users WHERE users.active = 1",
			wantErr:          false,
		},
		{
			name:             "complex query with table.column references",
			allowedDatabases: []string{},
			query:            "SELECT ls010_proposals.id, ls010_proposals.name, ls010_proposals_cstm.proposal_id_c FROM ls010_proposals INNER JOIN ls010_proposals_cstm ON ls010_proposals.id = ls010_proposals_cstm.id_c",
			wantErr:          false,
		},
		{
			name:             "table.column in WHERE clause should be allowed",
			allowedDatabases: []string{},
			query:            "SELECT * FROM orders WHERE orders.user_id = 123 AND orders.status = 'active'",
			wantErr:          false,
		},
		{
			name:             "database.table in FROM clause should be validated",
			allowedDatabases: []string{"testdb"},
			query:            "SELECT * FROM production.users",
			wantErr:          true,
			errMsg:           "access denied: database 'production' is not in allowed databases list",
		},
		{
			name:             "database.table in JOIN clause should be validated",
			allowedDatabases: []string{"testdb"},
			query:            "SELECT * FROM testdb.users u JOIN production.orders o ON u.id = o.user_id",
			wantErr:          true,
			errMsg:           "access denied: database 'production' is not in allowed databases list",
		},
		{
			name:             "complex query with many table.column references should pass",
			allowedDatabases: []string{},
			query: `SELECT
				inventory.item_id as item_id,
				inventory.item_name as item_name,
				catalog_data.sku_code,
				catalog_data.retail_price as item_price,
				shipments.tracking_id as tracking_id,
				shipments.carrier_name as carrier_name,
				logistics_data.route_number,
				shipments.dispatch_date as ship_date,
				CASE
					WHEN catalog_data.retail_price > 0 THEN
						ROUND(((logistics_data.handling_cost -
						catalog_data.wholesale_price) / catalog_data.
						retail_price) * 100, 2)
					ELSE NULL
				END as profit_margin
			FROM inventory
			INNER JOIN catalog_data ON inventory.item_id = catalog_data.item_ref
			INNER JOIN item_shipment_map ON inventory.item_id = item_shipment_map.item_key
			INNER JOIN shipments ON item_shipment_map.shipment_ref = shipments.tracking_id
			INNER JOIN logistics_data ON shipments.tracking_id = logistics_data.shipment_ref
			WHERE
				inventory.is_active = 1
				AND shipments.is_cancelled = 0
				AND catalog_data.sale_date IS NOT NULL
				AND DATE(shipments.dispatch_date) >= catalog_data.sale_date
			ORDER BY inventory.item_id, shipments.dispatch_date ASC`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewQueryValidator(createTestConfig(tt.allowedDatabases))
			err := validator.validateDatabaseAccess(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDatabaseAccess() expected error but got none")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("validateDatabaseAccess() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateDatabaseAccess() error = %v, want nil", err)
			}
		})
	}
}

func TestQueryValidator_ValidateQueryComplexity(t *testing.T) {
	validator := NewQueryValidator(createTestConfig(nil))

	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "simple select",
			query:   "SELECT * FROM users",
			wantErr: false,
		},
		{
			name:    "acceptable number of subqueries",
			query:   "SELECT * FROM users WHERE id IN (SELECT id FROM orders WHERE total > (SELECT AVG(total) FROM orders))",
			wantErr: false,
		},
		{
			name:    "too many subqueries",
			query:   "SELECT * FROM t1 WHERE id IN (SELECT id FROM t2 WHERE id IN (SELECT id FROM t3 WHERE id IN (SELECT id FROM t4 WHERE id IN (SELECT id FROM t5 WHERE id IN (SELECT id FROM t6 WHERE id IN (SELECT id FROM t7))))))",
			wantErr: true,
			errMsg:  "query complexity limit exceeded: too many subqueries",
		},
		{
			name:    "acceptable number of joins",
			query:   "SELECT * FROM t1 JOIN t2 ON t1.id = t2.id JOIN t3 ON t2.id = t3.id",
			wantErr: false,
		},
		{
			name:    "too many joins",
			query:   "SELECT * FROM t1 JOIN t2 ON t1.id=t2.id JOIN t3 ON t2.id=t3.id JOIN t4 ON t3.id=t4.id JOIN t5 ON t4.id=t5.id JOIN t6 ON t5.id=t6.id JOIN t7 ON t6.id=t7.id JOIN t8 ON t7.id=t8.id JOIN t9 ON t8.id=t9.id JOIN t10 ON t9.id=t10.id JOIN t11 ON t10.id=t11.id JOIN t12 ON t11.id=t12.id",
			wantErr: true,
			errMsg:  "query complexity limit exceeded: too many JOINs",
		},
		{
			name:    "query too long",
			query:   generateLongQuery(60000), // > 50KB
			wantErr: true,
			errMsg:  "query complexity limit exceeded: query too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateQueryComplexity(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateQueryComplexity() expected error but got none")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("validateQueryComplexity() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateQueryComplexity() error = %v, want nil", err)
			}
		})
	}
}

func TestQueryValidator_ValidateQuery_Integration(t *testing.T) {
	tests := []struct {
		name             string
		allowedDatabases []string
		query            string
		wantErr          bool
		errMsg           string
	}{
		{
			name:             "completely valid query",
			allowedDatabases: []string{"testdb"},
			query:            "SELECT name, email FROM testdb.users WHERE active = true",
			wantErr:          false,
		},
		{
			name:             "multiple validation failures",
			allowedDatabases: []string{"testdb"},
			query:            "USE production; -- DROP TABLE users",
			wantErr:          true,
			errMsg:           "access denied", // Should catch database access violation first
		},
		{
			name:             "SQL injection attempt",
			allowedDatabases: []string{"testdb"},
			query:            "SELECT * FROM testdb.users WHERE name = 'admin'; DROP TABLE testdb.users; --",
			wantErr:          true,
			errMsg:           "potentially dangerous pattern detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewQueryValidator(createTestConfig(tt.allowedDatabases))
			err := validator.ValidateQuery(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateQuery() expected error but got none")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("ValidateQuery() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidateQuery() error = %v, want nil", err)
			}
		})
	}
}

func TestQueryValidator_SanitizeErrorMessage(t *testing.T) {
	config := &config.DatabaseConfig{
		Host:     "secret-host.com",
		Username: "secret-user",
		Password: "secret-password",
	}
	validator := NewQueryValidator(config)

	tests := []struct {
		name    string
		errMsg  string
		wantMsg string
	}{
		{
			name:    "nil error",
			errMsg:  "",
			wantMsg: "",
		},
		{
			name:    "error with password",
			errMsg:  "connection failed: password 'secret-password' is incorrect",
			wantMsg: "connection failed: password '[REDACTED]' is incorrect",
		},
		{
			name:    "error with username",
			errMsg:  "authentication failed for user secret-user",
			wantMsg: "authentication failed for user [REDACTED]",
		},
		{
			name:    "error with host",
			errMsg:  "cannot connect to secret-host.com:5432",
			wantMsg: "cannot connect to [REDACTED]:5432",
		},
		{
			name:    "error with multiple sensitive data",
			errMsg:  "user secret-user@secret-host.com authentication failed with password secret-password",
			wantMsg: "user [REDACTED]@[REDACTED] authentication failed with password [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputErr error
			if tt.errMsg != "" {
				inputErr = &testError{msg: tt.errMsg}
			}

			result := validator.SanitizeErrorMessage(inputErr)

			if tt.wantMsg == "" {
				if result != nil {
					t.Errorf("SanitizeErrorMessage() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("SanitizeErrorMessage() = nil, want %v", tt.wantMsg)
				} else if result.Error() != tt.wantMsg {
					t.Errorf("SanitizeErrorMessage() = %v, want %v", result.Error(), tt.wantMsg)
				}
			}
		})
	}
}

// Helper functions and types

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func containsSubstring(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(needle == "" || strings.Contains(haystack, needle))
}

func generateLongQuery(length int) string {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	for i := 0; i < length-10; i++ {
		builder.WriteString("a")
	}
	builder.WriteString(" FROM test")
	return builder.String()
}

// Benchmarks for performance validation

func BenchmarkQueryValidator_ValidateQuery(b *testing.B) {
	validator := NewQueryValidator(createTestConfig([]string{"testdb", "devdb"}))
	query := "SELECT u.name, o.total FROM testdb.users u JOIN testdb.orders o ON u.id = o.user_id WHERE u.active = true"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateQuery(query)
	}
}

func BenchmarkQueryValidator_ValidateDatabaseAccess(b *testing.B) {
	validator := NewQueryValidator(createTestConfig([]string{"testdb", "devdb", "staging"}))
	query := "SELECT * FROM testdb.users u JOIN devdb.profiles p ON u.id = p.user_id WHERE u.created_at > '2023-01-01'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.validateDatabaseAccess(query)
	}
}
