package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

// RunMigrations executes database migrations using goose.
func RunMigrations(db *sql.DB, dbType string, migrationsFS embed.FS) error {
	// Determine the migration path based on database type
	var migrationPath string
	switch dbType {
	case "sqlite":
		migrationPath = "sqlite"
	case "postgresql", "postgres":
		migrationPath = "postgres"
	default:
		return fmt.Errorf("unsupported database type for migrations: %s", dbType)
	}

	// Get the subdirectory from the embedded filesystem
	migrationsSubFS, err := fs.Sub(migrationsFS, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to access migrations directory %s: %w", migrationPath, err)
	}

	// Set the appropriate goose dialect
	var dialect string
	if dbType == "sqlite" {
		dialect = "sqlite3"
	} else {
		dialect = "postgres"
	}

	// Set goose to use the correct dialect
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect to %s: %w", dialect, err)
	}

	// Run migrations
	goose.SetBaseFS(migrationsSubFS)
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// MigrateUp runs all pending migrations.
func MigrateUp(db *sql.DB, dbType string, migrationsFS embed.FS) error {
	return RunMigrations(db, dbType, migrationsFS)
}

// MigrateDown rolls back the last migration.
func MigrateDown(db *sql.DB, dbType string, migrationsFS embed.FS) error {
	var migrationPath string
	switch dbType {
	case "sqlite":
		migrationPath = "sqlite"
	case "postgresql", "postgres":
		migrationPath = "postgres"
	default:
		return fmt.Errorf("unsupported database type for migrations: %s", dbType)
	}

	migrationsSubFS, err := fs.Sub(migrationsFS, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to access migrations directory %s: %w", migrationPath, err)
	}

	var dialect string
	if dbType == "sqlite" {
		dialect = "sqlite3"
	} else {
		dialect = "postgres"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("failed to set goose dialect to %s: %w", dialect, err)
	}

	goose.SetBaseFS(migrationsSubFS)
	if err := goose.Down(db, "."); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// MigrationStatus returns the current migration version.
func MigrationStatus(db *sql.DB, dbType string, migrationsFS embed.FS) (int64, error) {
	var migrationPath string
	switch dbType {
	case "sqlite":
		migrationPath = "sqlite"
	case "postgresql", "postgres":
		migrationPath = "postgres"
	default:
		return 0, fmt.Errorf("unsupported database type for migrations: %s", dbType)
	}

	migrationsSubFS, err := fs.Sub(migrationsFS, migrationPath)
	if err != nil {
		return 0, fmt.Errorf("failed to access migrations directory %s: %w", migrationPath, err)
	}

	var dialect string
	if dbType == "sqlite" {
		dialect = "sqlite3"
	} else {
		dialect = "postgres"
	}

	if err := goose.SetDialect(dialect); err != nil {
		return 0, fmt.Errorf("failed to set goose dialect to %s: %w", dialect, err)
	}

	goose.SetBaseFS(migrationsSubFS)
	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, nil
}
