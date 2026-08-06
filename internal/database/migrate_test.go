package database

import (
	"database/sql"
	"testing"

	"github.com/shurco/mycart/db/migrations"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func TestRunMigrations_SQLite(t *testing.T) {
	// Create in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationsFS := migrations.Embed()
	err = RunMigrations(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify migrations ran by checking if tables exist
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != 'goose_db_version'").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}

	if tableCount == 0 {
		t.Error("Expected at least one table to be created by migrations")
	}

	// Check migration version
	version, err := MigrationStatus(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if version == 0 {
		t.Error("Expected migration version > 0")
	}
}

func TestRunMigrations_InvalidType(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrationsFS := migrations.Embed()
	err = RunMigrations(db, "invalid", migrationsFS)
	if err == nil {
		t.Error("Expected error for invalid database type, got nil")
	}
}

func TestMigrateUpDown_SQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	migrationsFS := migrations.Embed()

	// Migrate up
	err = MigrateUp(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to migrate up: %v", err)
	}

	// Get version after up
	versionUp, err := MigrationStatus(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if versionUp == 0 {
		t.Error("Expected version > 0 after migrate up")
	}

	// Migrate down
	err = MigrateDown(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to migrate down: %v", err)
	}

	// Get version after down
	versionDown, err := MigrationStatus(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if versionDown >= versionUp {
		t.Errorf("Expected version to decrease after migrate down, got %d (was %d)", versionDown, versionUp)
	}
}

func TestMigrationStatus_NoMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	migrationsFS := migrations.Embed()

	// Get status before any migrations
	version, err := MigrationStatus(db, "sqlite", migrationsFS)
	if err != nil {
		t.Fatalf("Failed to get migration status: %v", err)
	}

	if version != 0 {
		t.Errorf("Expected version 0 before migrations, got %d", version)
	}
}
