package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	_ "modernc.org/sqlite"
)

type sqliteDB struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite database instance
func NewSQLite(path string) (Database, error) {
	// For :memory: databases, use a unique name with shared cache to ensure:
	// 1. All connections from THIS db instance see the same schema (required for migrations)
	// 2. Different db instances are isolated from each other (required for parallel tests)
	dsn := path
	if path == ":memory:" {
		// Generate a unique name for this in-memory database
		uniqueID := generateUniqueID()
		dsn = fmt.Sprintf("file:mem-%s?mode=memory&cache=shared", uniqueID)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Enable foreign key constraints
	// This is set per-connection, so we set it once on the initial connection
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &sqliteDB{db: db}, nil
}

// generateUniqueID generates a unique identifier for in-memory databases
func generateUniqueID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback if random fails (shouldn't happen)
		return "fallback"
	}
	return hex.EncodeToString(b)
}


func (s *sqliteDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

func (s *sqliteDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *sqliteDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *sqliteDB) Prepare(ctx context.Context, query string) (Stmt, error) {
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return &sqliteStmt{stmt: stmt}, nil
}

func (s *sqliteDB) Begin(ctx context.Context) (Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &sqliteTx{tx: tx}, nil
}

func (s *sqliteDB) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqliteDB) Close() error {
	return s.db.Close()
}

func (s *sqliteDB) DB() *sql.DB {
	return s.db
}

func (s *sqliteDB) Type() string {
	return "sqlite"
}

// sqliteTx implements Tx interface
type sqliteTx struct {
	tx *sql.Tx
}

func (t *sqliteTx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *sqliteTx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *sqliteTx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *sqliteTx) Prepare(ctx context.Context, query string) (Stmt, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement in transaction: %w", err)
	}
	return &sqliteStmt{stmt: stmt}, nil
}

func (t *sqliteTx) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTx) Rollback() error {
	return t.tx.Rollback()
}

// sqliteStmt implements Stmt interface
type sqliteStmt struct {
	stmt *sql.Stmt
}

func (s *sqliteStmt) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return s.stmt.ExecContext(ctx, args...)
}

func (s *sqliteStmt) Query(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return s.stmt.QueryContext(ctx, args...)
}

func (s *sqliteStmt) QueryRow(ctx context.Context, args ...interface{}) *sql.Row {
	return s.stmt.QueryRowContext(ctx, args...)
}

func (s *sqliteStmt) Close() error {
	return s.stmt.Close()
}
