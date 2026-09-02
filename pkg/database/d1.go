package database

import (
	"context"
	"database/sql"
	"fmt"
)

type d1DB struct {
	db *sql.DB
}

// NewD1 creates a new Cloudflare D1 database instance using the d1 driver
func NewD1(accountID, databaseID, apiToken string) (Database, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if databaseID == "" {
		return nil, fmt.Errorf("database ID is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	// Create DSN for d1 driver: "accountID/databaseID?api_token=xxx"
	dsn := fmt.Sprintf("%s/%s?api_token=%s", accountID, databaseID, apiToken)

	db, err := sql.Open("d1", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open D1 database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping D1 database: %w", err)
	}

	return &d1DB{db: db}, nil
}

func (d *d1DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *d1DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *d1DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *d1DB) Prepare(ctx context.Context, query string) (Stmt, error) {
	stmt, err := d.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	return &d1Stmt{stmt: stmt}, nil
}

func (d *d1DB) Begin(ctx context.Context) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &d1Tx{tx: tx}, nil
}

func (d *d1DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *d1DB) Close() error {
	return d.db.Close()
}

func (d *d1DB) DB() *sql.DB {
	return d.db
}

// d1Tx implements Tx interface
type d1Tx struct {
	tx *sql.Tx
}

func (t *d1Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *d1Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *d1Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *d1Tx) Prepare(ctx context.Context, query string) (Stmt, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement in transaction: %w", err)
	}
	return &d1Stmt{stmt: stmt}, nil
}

func (t *d1Tx) Commit() error {
	return t.tx.Commit()
}

func (t *d1Tx) Rollback() error {
	return t.tx.Rollback()
}

// d1Stmt implements Stmt interface
type d1Stmt struct {
	stmt *sql.Stmt
}

func (s *d1Stmt) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return s.stmt.ExecContext(ctx, args...)
}

func (s *d1Stmt) Query(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return s.stmt.QueryContext(ctx, args...)
}

func (s *d1Stmt) QueryRow(ctx context.Context, args ...interface{}) *sql.Row {
	return s.stmt.QueryRowContext(ctx, args...)
}

func (s *d1Stmt) Close() error {
	return s.stmt.Close()
}
