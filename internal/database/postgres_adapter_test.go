package database

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// setupTestPostgres creates a test PostgreSQL connection.
// Skips the test if DATABASE_URL environment variable is not set.
func setupTestPostgres(t *testing.T) *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL tests")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func TestNewPostgresAdapter(t *testing.T) {
	db := setupTestPostgres(t)
	defer db.Close()

	adapter := NewPostgresAdapter(db)
	if adapter == nil {
		t.Fatal("NewPostgresAdapter returned nil")
	}

	if adapter.DB() != db {
		t.Error("DB() did not return the correct database instance")
	}

	if adapter.Queries() == nil {
		t.Error("Queries() returned nil")
	}
}

func TestPostgresAdapter_Ping(t *testing.T) {
	db := setupTestPostgres(t)
	defer db.Close()

	adapter := NewPostgresAdapter(db)
	ctx := context.Background()

	if err := adapter.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestPostgresAdapter_Stats(t *testing.T) {
	db := setupTestPostgres(t)
	defer db.Close()

	adapter := NewPostgresAdapter(db)
	stats := adapter.Stats()

	if stats.MaxOpenConnections < 0 {
		t.Error("Stats returned invalid MaxOpenConnections")
	}
}

func TestPostgresAdapter_Close(t *testing.T) {
	db := setupTestPostgres(t)

	adapter := NewPostgresAdapter(db)
	if err := adapter.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
