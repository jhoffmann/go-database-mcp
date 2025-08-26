// Package security provides security validation and access control for database operations.
package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jhoffmann/go-database-mcp/internal/config"
)

// QueryValidator provides security validation for SQL queries.
type QueryValidator struct {
	config *config.DatabaseConfig
}

// NewQueryValidator creates a new QueryValidator instance.
func NewQueryValidator(config *config.DatabaseConfig) *QueryValidator {
	return &QueryValidator{
		config: config,
	}
}

// ValidateQuery performs comprehensive security validation on a SQL query.
func (v *QueryValidator) ValidateQuery(query string) error {
	// Database access validation (check first for access control)
	if err := v.validateDatabaseAccess(query); err != nil {
		return err
	}

	// Basic validation
	if err := v.validateBasicSafety(query); err != nil {
		return err
	}

	// Query complexity validation
	if err := v.validateQueryComplexity(query); err != nil {
		return err
	}

	return nil
}

// validateBasicSafety performs basic SQL injection and dangerous operation checks.
func (v *QueryValidator) validateBasicSafety(query string) error {
	normalized := strings.ToUpper(strings.TrimSpace(query))

	if normalized == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []struct {
		pattern     string
		description string
	}{
		{"--", "SQL comments"},
		{";--", "SQL injection attempts"},
		{"/*", "SQL block comments"},
		{"*/", "SQL block comments"},
		{"EXEC(", "dynamic SQL execution"},
		{"EXECUTE(", "dynamic SQL execution"},
		{"SP_", "system stored procedures"},
		{"XP_", "extended stored procedures"},
		{"LOAD_FILE", "file system access"},
		{"INTO OUTFILE", "file system access"},
		{"INTO DUMPFILE", "file system access"},
	}

	for _, dangerous := range dangerousPatterns {
		if strings.Contains(normalized, dangerous.pattern) {
			return fmt.Errorf("potentially dangerous pattern detected (%s): %s", dangerous.description, dangerous.pattern)
		}
	}

	return nil
}

// validateDatabaseAccess validates that queries only access allowed databases.
func (v *QueryValidator) validateDatabaseAccess(query string) error {
	// Always validate database access - if AllowedDatabases is empty,
	// only the primary database should be allowed

	normalized := strings.ToUpper(strings.TrimSpace(query))

	// Check for USE statements (match at beginning of query or after semicolon)
	usePattern := regexp.MustCompile(`(?:^|\s*;\s*)USE\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:;|$|\s)`)
	if matches := usePattern.FindStringSubmatch(normalized); len(matches) > 1 {
		databaseName := strings.ToLower(matches[1])
		if !v.config.IsDatabaseAllowed(databaseName) {
			return fmt.Errorf("access denied: database '%s' is not in allowed databases list", databaseName)
		}
	}

	// Check for fully qualified table names (database.table)
	// This regex looks for word.word patterns that could be database.table
	qualifiedTablePattern := regexp.MustCompile(`(?i)\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\.\s*[a-zA-Z_][a-zA-Z0-9_]*(?:\s|$|,|\)|;)`)
	matches := qualifiedTablePattern.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if len(match) > 1 {
			databaseName := strings.ToLower(match[1])
			// Skip common SQL keywords and aliases that aren't database names
			if v.isSystemKeyword(databaseName) || v.isCommonAlias(databaseName) {
				continue
			}
			if !v.config.IsDatabaseAllowed(databaseName) {
				return fmt.Errorf("access denied: database '%s' is not in allowed databases list", databaseName)
			}
		}
	}

	// Check for INFORMATION_SCHEMA access restrictions
	if strings.Contains(normalized, "INFORMATION_SCHEMA") {
		// Allow INFORMATION_SCHEMA queries but validate they don't access restricted databases
		return v.validateInformationSchemaQuery(query)
	}

	return nil
}

// validateQueryComplexity checks for overly complex queries that might cause performance issues.
func (v *QueryValidator) validateQueryComplexity(query string) error {
	normalized := strings.ToUpper(strings.TrimSpace(query))

	// Limit on number of SELECT statements (including subqueries)
	selectCount := strings.Count(normalized, "SELECT")
	subqueryCount := selectCount - 1 // Subtract 1 for main query
	if subqueryCount > 5 {
		return fmt.Errorf("query complexity limit exceeded: too many subqueries (%d > 5)", subqueryCount)
	}

	// Limit on number of JOINs
	joinCount := strings.Count(normalized, "JOIN")
	if joinCount > 10 {
		return fmt.Errorf("query complexity limit exceeded: too many JOINs (%d > 10)", joinCount)
	}

	// Limit query length
	if len(query) > 50000 { // 50KB limit
		return fmt.Errorf("query complexity limit exceeded: query too long (%d characters > 50000)", len(query))
	}

	return nil
}

// validateInformationSchemaQuery validates queries against INFORMATION_SCHEMA.
func (v *QueryValidator) validateInformationSchemaQuery(query string) error {
	// For now, allow INFORMATION_SCHEMA queries
	// In production, you might want to add more restrictions
	return nil
}

// isSystemKeyword checks if a word is a common SQL system keyword.
func (v *QueryValidator) isSystemKeyword(word string) bool {
	keywords := map[string]bool{
		"INFORMATION_SCHEMA": true,
		"PERFORMANCE_SCHEMA": true,
		"SYS":                true,
		"MYSQL":              true,
		"PG_CATALOG":         true,
		"PUBLIC":             true,
	}
	return keywords[strings.ToUpper(word)]
}

// isCommonAlias checks if a word is commonly used as a table alias.
func (v *QueryValidator) isCommonAlias(word string) bool {
	aliases := map[string]bool{
		"u":  true,
		"o":  true,
		"p":  true,
		"t":  true,
		"t1": true,
		"t2": true,
		"t3": true,
		"a":  true,
		"b":  true,
		"c":  true,
	}
	return aliases[strings.ToLower(word)]
}

// SanitizeErrorMessage removes sensitive information from error messages.
func (v *QueryValidator) SanitizeErrorMessage(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()

	// Remove potential credential information
	sensitivePatterns := []string{
		v.config.Password,
		v.config.Username,
		v.config.Host,
	}

	for _, pattern := range sensitivePatterns {
		if pattern != "" {
			message = strings.ReplaceAll(message, pattern, "[REDACTED]")
		}
	}

	return fmt.Errorf("%s", message)
}
