package base

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/shurco/mycart/pkg/database"
	"github.com/shurco/mycart/pkg/fsutil"
)

// buildDSN builds a SQLite connection string with optimized parameters.
// _txlock=immediate makes every transaction take its write lock up front,
// which prevents deferred-transaction lock-upgrade deadlocks that
// busy_timeout cannot rescue.
//
// Note: _txlock must NOT be present on connections used by goose (it breaks
// its statement handling), so migrations run on a separate DSN.
func buildDSN(dbPath string) string {
	return fmt.Sprintf("%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_txlock=immediate", dbPath)
}

// New creates a new database connection and brings the schema up to date.
func New(dbPath string, migrations embed.FS) (*sql.DB, error) {
	if !fsutil.IsFile(dbPath) {
		if _, err := fsutil.OpenFile(dbPath, fsutil.FsCWFlags, 0o666); err != nil {
			return nil, err
		}
	}

	// Always run migrations on startup: goose.Up is idempotent and cheap,
	// so a binary upgrade followed by `serve` can never talk to a stale
	// schema (previously migrations only ran on freshly created databases).
	if err := Migrate(dbPath, migrations); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
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
	goose.SetBaseFS(migrations)
	db, err := goose.OpenDBWithDriver("sqlite", dbPath+"?_pragma=busy_timeout(10000)")
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	goose.SetTableName("migrate_db_version")

	return goose.Up(db, ".")
}

// MigrateDB performs database migrations on an existing *sql.DB connection
func MigrateDB(db *sql.DB, migrations embed.FS) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations)
	goose.SetTableName("migrate_db_version")
	return goose.Up(db, ".")
}

// MigrateD1 performs database migrations on D1 without using transactions
// D1 doesn't support traditional SQL transactions, so we execute each statement individually
func MigrateD1(db database.Database, migrations embed.FS) error {
	ctx := context.Background()

	// Create migration version table if it doesn't exist
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrate_db_version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id INTEGER NOT NULL,
			is_applied INTEGER NOT NULL DEFAULT 1,
			tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	// Get applied migrations
	appliedVersions := make(map[int64]bool)
	rows, err := db.Query(ctx, `SELECT version_id FROM migrate_db_version WHERE is_applied = 1`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var version int64
			if err := rows.Scan(&version); err == nil {
				appliedVersions[version] = true
			}
		}
	}

	// Read and sort migration files
	var migrationFiles []string
	err = fs.WalkDir(migrations, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && filepath.Ext(path) == ".sql" {
			migrationFiles = append(migrationFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk migrations: %w", err)
	}
	sort.Strings(migrationFiles)

	// Execute migrations
	for _, file := range migrationFiles {
		// Parse version from filename (e.g., "20230714135923_init_db.sql" -> 20230714135923)
		baseName := filepath.Base(file)
		parts := strings.SplitN(baseName, "_", 2)
		if len(parts) < 2 {
			continue
		}

		var version int64
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		// Skip if already applied
		if appliedVersions[version] {
			continue
		}

		// Read migration SQL
		sqlBytes, err := migrations.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		// Parse and execute SQL statements
		sqlContent := string(sqlBytes)
		statements := parseSQLStatements(sqlContent)

		for _, stmt := range statements {
			if stmt = strings.TrimSpace(stmt); stmt == "" {
				continue
			}
			_, err := db.Exec(ctx, stmt)
			if err != nil {
				return fmt.Errorf("exec migration %s statement: %s: %w", file, stmt[:50], err)
			}
		}

		// Record migration as applied
		_, err = db.Exec(ctx, `INSERT INTO migrate_db_version (version_id, is_applied) VALUES (?, 1)`, version)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", file, err)
		}
	}

	return nil
}

// parseSQLStatements splits SQL content into individual statements
// Handles goose directives and comments
// Only parses the UP section, stops at DOWN section
// For D1: Always splits by semicolons (ignores StatementBegin/End) because
// D1 cannot execute multiple SQL statements in a single API call
func parseSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inUpSection := false

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle goose directives
		if strings.HasPrefix(trimmed, "-- +goose") {
			if strings.Contains(trimmed, "Up") {
				inUpSection = true
				continue
			} else if strings.Contains(trimmed, "Down") {
				// Stop processing when we hit the DOWN section
				break
			} else if strings.Contains(trimmed, "StatementBegin") || strings.Contains(trimmed, "StatementEnd") {
				// Ignore StatementBegin/End for D1 - D1 requires individual statements
				continue
			}
		}

		// Skip lines before UP section starts
		if !inUpSection {
			continue
		}

		// Skip comments
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Always split by semicolons for D1 (ignore goose StatementBegin/End)
		if strings.HasSuffix(trimmed, ";") {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Add any remaining SQL
	if stmt := current.String(); strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}
