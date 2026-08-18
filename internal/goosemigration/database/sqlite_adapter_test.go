package database

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewSQLiteAdapter(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	if adapter == nil {
		t.Fatal("NewSQLiteAdapter returned nil")
	}

	if adapter.DB() != db {
		t.Error("DB() did not return the correct database instance")
	}

	if adapter.Queries() == nil {
		t.Error("Queries() returned nil")
	}
}

func TestSQLiteAdapter_Ping(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	ctx := context.Background()

	if err := adapter.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestSQLiteAdapter_Stats(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	stats := adapter.Stats()

	if stats.MaxOpenConnections < 0 {
		t.Error("Stats returned invalid MaxOpenConnections")
	}
}

func TestSQLiteAdapter_Close(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	adapter := NewSQLiteAdapter(db)
	if err := adapter.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
