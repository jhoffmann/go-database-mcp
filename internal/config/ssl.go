// Package config provides SSL/TLS configuration mapping for different database drivers.
package config

import "fmt"

// SSLMode represents the common SSL/TLS configuration options that work across
// different database types. These values are mapped to database-specific SSL modes.
type SSLMode string

const (
	// SSLModeNone disables SSL/TLS encryption entirely
	SSLModeNone SSLMode = "none"

	// SSLModePrefer attempts to use SSL/TLS but falls back to unencrypted if unavailable
	SSLModePrefer SSLMode = "prefer"

	// SSLModeRequire mandates SSL/TLS encryption and fails if unavailable
	SSLModeRequire SSLMode = "require"
)

// ValidSSLModes returns a list of all valid SSL mode values
func ValidSSLModes() []SSLMode {
	return []SSLMode{SSLModeNone, SSLModePrefer, SSLModeRequire}
}

// IsValid checks if the given SSL mode string is valid
func (s SSLMode) IsValid() bool {
	switch s {
	case SSLModeNone, SSLModePrefer, SSLModeRequire:
		return true
	default:
		return false
	}
}

// ToMySQLSSLMode converts a common SSL mode to MySQL-specific SSL configuration
func (s SSLMode) ToMySQLSSLMode() (string, error) {
	switch s {
	case SSLModeNone:
		return "false", nil
	case SSLModePrefer:
		return "preferred", nil
	case SSLModeRequire:
		return "true", nil
	default:
		return "", fmt.Errorf("invalid SSL mode: %s", s)
	}
}

// ToPostgreSQLSSLMode converts a common SSL mode to PostgreSQL-specific SSL configuration
func (s SSLMode) ToPostgreSQLSSLMode() (string, error) {
	switch s {
	case SSLModeNone:
		return "disable", nil
	case SSLModePrefer:
		return "prefer", nil
	case SSLModeRequire:
		return "require", nil
	default:
		return "", fmt.Errorf("invalid SSL mode: %s", s)
	}
}

// ParseSSLMode parses a string into an SSLMode, returning an error if invalid
func ParseSSLMode(mode string) (SSLMode, error) {
	sslMode := SSLMode(mode)
	if !sslMode.IsValid() {
		return "", fmt.Errorf("invalid SSL mode '%s', valid options are: none, prefer, require", mode)
	}
	return sslMode, nil
}
