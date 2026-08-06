package database

import (
	"os"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts to be 5, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay to be 1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay to be 30s, got %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier to be 2.0, got %f", cfg.Multiplier)
	}
}

func TestConnectWithRetry_SQLite(t *testing.T) {
	// Create temporary SQLite database path
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &Config{
		Type: "sqlite",
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Use fast retry config for testing
	retry := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	db, err := ConnectWithRetryConfig(cfg, retry)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	// Verify we got a SQLite adapter
	if _, ok := db.(*SQLiteAdapter); !ok {
		t.Error("Expected SQLiteAdapter, got different type")
	}

	// Test connection pool settings
	stats := db.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Errorf("Expected MaxOpenConnections to be 1 for SQLite, got %d", stats.MaxOpenConnections)
	}
}

func TestConnectWithRetry_PostgreSQL(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL connection test")
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Override with PostgreSQL config
	cfg.Type = "postgresql"

	// Use fast retry config for testing
	retry := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	db, err := ConnectWithRetryConfig(cfg, retry)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Verify we got a PostgreSQL adapter
	if _, ok := db.(*PostgresAdapter); !ok {
		t.Error("Expected PostgresAdapter, got different type")
	}

	// Test connection pool settings
	stats := db.Stats()
	if stats.MaxOpenConnections != cfg.PostgreSQL.MaxOpenConns {
		t.Errorf("Expected MaxOpenConnections to be %d, got %d", cfg.PostgreSQL.MaxOpenConns, stats.MaxOpenConnections)
	}
}

func TestConnectWithRetry_InvalidDriver(t *testing.T) {
	cfg := &Config{
		Type: "invalid",
	}

	retry := RetryConfig{
		MaxAttempts:  1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}

	_, err := ConnectWithRetryConfig(cfg, retry)
	if err == nil {
		t.Error("Expected error for invalid database type, got nil")
	}
}

func TestTestConnection_SQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	cfg := &Config{
		Type: "sqlite",
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	err := TestConnection(cfg)
	if err != nil {
		t.Errorf("TestConnection failed: %v", err)
	}
}

func TestTestConnection_InvalidPath(t *testing.T) {
	cfg := &Config{
		Type: "sqlite",
		SQLite: SQLiteConfig{
			Path: "/nonexistent/invalid/path/test.db",
		},
	}

	err := TestConnection(cfg)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
