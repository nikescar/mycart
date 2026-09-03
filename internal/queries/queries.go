package queries

import (
	"context"
	"database/sql"
	"embed"

	"github.com/shurco/mycart/internal/base"
	"github.com/shurco/mycart/pkg/database"
)

var db *Base

// Define the structure 'Base' that aggregates various queries related to different modules like
// settings, authentication, installation, pages, products, and cart management.
type Base struct {
	SettingQueries
	AuthQueries
	InstallQueries
	PageQueries
	ProductQueries
	CartQueries
}

// New initializes the application's database and returns an error if any occurs during the process.
// It takes a database.Database interface and an embed.FS for migrations.
// Migrations are run against the underlying *sql.DB for compatibility with goose.
func New(database database.Database, migrations embed.FS) (err error) {
	// Check database type to determine migration strategy
	dbType := database.Type()

	// Run migrations using the underlying *sql.DB
	// This is needed because goose requires *sql.DB
	underlyingDB := database.DB()
	if underlyingDB != nil {
		if dbType == "d1" {
			// For D1, use transaction-free migration approach
			if err = base.MigrateD1(database, migrations); err != nil {
				return
			}
		} else {
			// For SQLite, run migrations with goose
			if err = base.MigrateDB(underlyingDB, migrations); err != nil {
				return
			}
		}
	}

	db = &Base{
		AuthQueries:    AuthQueries{DB: database},
		InstallQueries: InstallQueries{DB: database},
		SettingQueries: SettingQueries{DB: database},
		PageQueries:    PageQueries{DB: database},
		ProductQueries: ProductQueries{DB: database},
		CartQueries:    CartQueries{DB: database},
	}
	return
}

// NewWithoutMigrations initializes the application's database without running migrations.
// Use this when the database schema is already up to date (e.g., D1 databases that don't support goose migrations).
func NewWithoutMigrations(database database.Database) (err error) {
	db = &Base{
		AuthQueries:    AuthQueries{DB: database},
		InstallQueries: InstallQueries{DB: database},
		SettingQueries: SettingQueries{DB: database},
		PageQueries:    PageQueries{DB: database},
		ProductQueries: ProductQueries{DB: database},
		CartQueries:    CartQueries{DB: database},
	}
	return
}

// NewFromDB initializes Base from an existing *sql.DB (e.g. in-memory SQLite for tests).
// This is a compatibility function for tests.
func NewFromDB(sqlite *sql.DB) {
	// Wrap the *sql.DB in our SQLite adapter
	sqliteDB := &sqliteDBWrapper{db: sqlite}

	db = &Base{
		AuthQueries:    AuthQueries{DB: sqliteDB},
		InstallQueries: InstallQueries{DB: sqliteDB},
		SettingQueries: SettingQueries{DB: sqliteDB},
		PageQueries:    PageQueries{DB: sqliteDB},
		ProductQueries: ProductQueries{DB: sqliteDB},
		CartQueries:    CartQueries{DB: sqliteDB},
	}
}

// DB returns the Base instance. If the database is not initialized, returns nil.
// Use New() to initialize the database before calling DB().
func DB() *Base {
	return db
}

// Type returns the database type ("sqlite" or "d1")
func (b *Base) Type() string {
	if b == nil || b.SettingQueries.DB == nil {
		return "unknown"
	}
	return b.SettingQueries.DB.Type()
}

// sqliteDBWrapper wraps *sql.DB to implement database.Database interface
// This is used by NewFromDB for test compatibility
type sqliteDBWrapper struct {
	db *sql.DB
}

func (w *sqliteDBWrapper) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *sqliteDBWrapper) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.db.QueryContext(ctx, query, args...)
}

func (w *sqliteDBWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

func (w *sqliteDBWrapper) Prepare(ctx context.Context, query string) (database.Stmt, error) {
	stmt, err := w.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &sqliteStmtWrapper{stmt: stmt}, nil
}

func (w *sqliteDBWrapper) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteTxWrapper{tx: tx}, nil
}

func (w *sqliteDBWrapper) Ping(ctx context.Context) error {
	return w.db.PingContext(ctx)
}

func (w *sqliteDBWrapper) Close() error {
	return w.db.Close()
}

func (w *sqliteDBWrapper) DB() *sql.DB {
	return w.db
}

func (w *sqliteDBWrapper) Type() string {
	return "sqlite"
}

// sqliteTxWrapper wraps *sql.Tx to implement database.Tx interface
type sqliteTxWrapper struct {
	tx *sql.Tx
}

func (w *sqliteTxWrapper) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w *sqliteTxWrapper) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, query, args...)
}

func (w *sqliteTxWrapper) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRowContext(ctx, query, args...)
}

func (w *sqliteTxWrapper) Prepare(ctx context.Context, query string) (database.Stmt, error) {
	stmt, err := w.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &sqliteStmtWrapper{stmt: stmt}, nil
}

func (w *sqliteTxWrapper) Commit() error {
	return w.tx.Commit()
}

func (w *sqliteTxWrapper) Rollback() error {
	return w.tx.Rollback()
}

// sqliteStmtWrapper wraps *sql.Stmt to implement database.Stmt interface
type sqliteStmtWrapper struct {
	stmt *sql.Stmt
}

func (w *sqliteStmtWrapper) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return w.stmt.ExecContext(ctx, args...)
}

func (w *sqliteStmtWrapper) Query(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return w.stmt.QueryContext(ctx, args...)
}

func (w *sqliteStmtWrapper) QueryRow(ctx context.Context, args ...interface{}) *sql.Row {
	return w.stmt.QueryRowContext(ctx, args...)
}

func (w *sqliteStmtWrapper) Close() error {
	return w.stmt.Close()
}
