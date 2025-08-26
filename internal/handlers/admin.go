// Package handlers provides MCP tool handlers for database operations.
package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/jhoffmann/go-database-mcp/internal/database"
)

// AdminHandler handles database administrative operations.
type AdminHandler struct {
	db database.Database
}

// ConnectionInfo represents database connection information.
type ConnectionInfo struct {
	Driver    string `json:"driver"`    // Database driver name
	Connected bool   `json:"connected"` // Whether currently connected
	PingTime  string `json:"ping_time"` // Time taken to ping database
}

// NewAdminHandler creates a new AdminHandler instance.
func NewAdminHandler(db database.Database) *AdminHandler {
	return &AdminHandler{
		db: db,
	}
}

// GetConnectionInfo retrieves information about the current database connection.
func (h *AdminHandler) GetConnectionInfo(ctx context.Context) (*ConnectionInfo, error) {
	start := time.Now()
	err := h.db.Ping(ctx)
	pingDuration := time.Since(start)

	return &ConnectionInfo{
		Driver:    h.db.GetDriverName(),
		Connected: err == nil,
		PingTime:  fmt.Sprintf("%.2fms", float64(pingDuration.Nanoseconds())/1e6),
	}, nil
}
