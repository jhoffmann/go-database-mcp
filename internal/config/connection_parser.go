package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ConnectionInfo holds parsed connection string information
type ConnectionInfo struct {
	Type     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// ParseConnectionString parses a database connection string and returns ConnectionInfo.
// Supports both PostgreSQL and MySQL connection strings:
// - postgresql://[user[:password]@][host[:port]]/[dbname][?param1=value1&...]
// - mysql://[user[:password]@][host[:port]]/[dbname][?param1=value1&...]
func ParseConnectionString(connectionString string) (*ConnectionInfo, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("connection string is empty")
	}

	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	info := &ConnectionInfo{}

	// Determine database type from scheme
	switch strings.ToLower(parsedURL.Scheme) {
	case "postgres", "postgresql":
		info.Type = "postgres"
	case "mysql":
		info.Type = "mysql"
	default:
		return nil, fmt.Errorf("unsupported database scheme: %s (supported: postgresql, mysql)", parsedURL.Scheme)
	}

	// Extract hostname and port
	if parsedURL.Hostname() != "" {
		info.Host = parsedURL.Hostname()
	} else {
		return nil, fmt.Errorf("hostname is required in connection string")
	}

	if parsedURL.Port() != "" {
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %s", parsedURL.Port())
		}
		info.Port = port
	} else {
		// Set default ports
		switch info.Type {
		case "postgres":
			info.Port = 5432
		case "mysql":
			info.Port = 3306
		}
	}

	// Extract database name from path
	if len(parsedURL.Path) > 1 { // Path starts with '/'
		info.Database = parsedURL.Path[1:] // Remove leading '/'
	} else {
		return nil, fmt.Errorf("database name is required in connection string")
	}

	// Extract username and password
	if parsedURL.User != nil {
		info.Username = parsedURL.User.Username()
		if info.Username == "" {
			return nil, fmt.Errorf("username is required in connection string")
		}
		if password, hasPassword := parsedURL.User.Password(); hasPassword {
			info.Password = password
		}
	} else {
		return nil, fmt.Errorf("username is required in connection string")
	}

	// Extract SSL mode from query parameters
	queryParams := parsedURL.Query()
	if sslMode := queryParams.Get("sslmode"); sslMode != "" {
		info.SSLMode = sslMode
	} else {
		// Set default SSL modes
		switch info.Type {
		case "postgres":
			info.SSLMode = "prefer"
		case "mysql":
			info.SSLMode = "prefer"
		}
	}

	return info, nil
}

// ToConnectionString converts ConnectionInfo back to a connection string format.
// This is useful for testing and validation purposes.
func (info *ConnectionInfo) ToConnectionString() string {
	var scheme string
	switch info.Type {
	case "postgres":
		scheme = "postgresql"
	case "mysql":
		scheme = "mysql"
	default:
		scheme = info.Type
	}

	userInfo := info.Username
	if info.Password != "" {
		userInfo = fmt.Sprintf("%s:%s", info.Username, info.Password)
	}

	hostPort := info.Host
	if info.Port > 0 {
		hostPort = fmt.Sprintf("%s:%d", info.Host, info.Port)
	}

	connectionString := fmt.Sprintf("%s://%s@%s/%s", scheme, userInfo, hostPort, info.Database)

	if info.SSLMode != "" {
		connectionString += fmt.Sprintf("?sslmode=%s", info.SSLMode)
	}

	return connectionString
}
