package config

import (
	"testing"
)

func TestDatabaseConfig_DefaultValues(t *testing.T) {
	cfg := DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		MaxConns:     10,
		MaxIdleConns: 5,
		SSLMode:      "prefer",
	}

	if cfg.Type != "postgres" {
		t.Errorf("Expected Type to be 'postgres', got %s", cfg.Type)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected Port to be 5432, got %d", cfg.Port)
	}
	if cfg.Database != "testdb" {
		t.Errorf("Expected Database to be 'testdb', got %s", cfg.Database)
	}
	if cfg.Username != "testuser" {
		t.Errorf("Expected Username to be 'testuser', got %s", cfg.Username)
	}
	if cfg.Password != "testpass" {
		t.Errorf("Expected Password to be 'testpass', got %s", cfg.Password)
	}
	if cfg.MaxConns != 10 {
		t.Errorf("Expected MaxConns to be 10, got %d", cfg.MaxConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("Expected MaxIdleConns to be 5, got %d", cfg.MaxIdleConns)
	}
	if cfg.SSLMode != "prefer" {
		t.Errorf("Expected SSLMode to be 'prefer', got %s", cfg.SSLMode)
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Type: "mysql",
			Host: "db.example.com",
			Port: 3306,
		},
	}

	if cfg.Database.Type != "mysql" {
		t.Errorf("Expected Database.Type to be 'mysql', got %s", cfg.Database.Type)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected Database.Host to be 'db.example.com', got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("Expected Database.Port to be 3306, got %d", cfg.Database.Port)
	}
}
