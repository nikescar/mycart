package base

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/shurco/mycart/pkg/fsutil"
)

// buildDSN builds a SQLite connection string with optimized parameters
func buildDSN(dbPath string) string {
	return fmt.Sprintf("%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)
}

// New creates a new database connection and performs migrations if necessary
func New(dbPath string, migrations embed.FS) (*sql.DB, error) {
	if !fsutil.IsFile(dbPath) {
		if _, err := fsutil.OpenFile(dbPath, fsutil.FsCWFlags, 0o666); err != nil {
			return nil, err
		}

		if err := Migrate(dbPath, migrations); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", buildDSN(dbPath))
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA auto_vacuum"); err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate performs database migrations
func Migrate(dbPath string, migrations embed.FS) error {
	// Extract sqlite subdirectory from embedded filesystem
	migrationsSubFS, err := fs.Sub(migrations, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to access sqlite migrations: %w", err)
	}

	// Set goose dialect
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	goose.SetBaseFS(migrationsSubFS)
	db, err := goose.OpenDBWithDriver("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	goose.SetTableName("migrate_db_version")

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
