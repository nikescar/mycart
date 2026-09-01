package database

import (
	"context"
	"database/sql"
	"fmt"
)

type d1DB struct {
	accountID  string
	databaseID string
	apiToken   string
}

// NewD1 creates a new Cloudflare D1 database instance
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

	return &d1DB{
		accountID:  accountID,
		databaseID: databaseID,
		apiToken:   apiToken,
	}, nil
}

// Exec executes a query without returning rows (stub - HTTP API deferred)
func (d *d1DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, fmt.Errorf("D1 Exec not yet implemented")
}

// Query executes a query that returns rows (stub - HTTP API deferred)
func (d *d1DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("D1 Query not yet implemented")
}

// QueryRow executes a query that returns at most one row (stub - HTTP API deferred)
func (d *d1DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// sql.Row doesn't expose errors until Scan, so we return a dummy row
	// Real implementation will use D1 HTTP API
	return &sql.Row{}
}

// Begin starts a transaction (no-op for D1)
func (d *d1DB) Begin(ctx context.Context) (Tx, error) {
	// D1 doesn't support traditional transactions via HTTP API
	// Return no-op transaction for interface compatibility
	return &d1Tx{}, nil
}

// Ping checks the database connection (stub - HTTP API deferred)
func (d *d1DB) Ping(ctx context.Context) error {
	return fmt.Errorf("D1 Ping not yet implemented")
}

// Close closes the database connection (no-op for D1)
func (d *d1DB) Close() error {
	// D1 HTTP client has no persistent connection to close
	return nil
}

// DB returns the underlying *sql.DB (not supported for D1)
func (d *d1DB) DB() *sql.DB {
	// D1 uses HTTP API, not database/sql driver
	// This will need special handling in migration code
	return nil
}

// d1Tx implements Tx interface with no-op methods
type d1Tx struct{}

// Exec executes a query in transaction (stub - HTTP API deferred)
func (t *d1Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, fmt.Errorf("D1 transaction Exec not yet implemented")
}

// Query executes a query in transaction (stub - HTTP API deferred)
func (t *d1Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("D1 transaction Query not yet implemented")
}

// QueryRow executes a query in transaction (stub - HTTP API deferred)
func (t *d1Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return &sql.Row{}
}

// Commit commits the transaction (no-op)
func (t *d1Tx) Commit() error {
	// No-op - D1 doesn't support traditional transactions
	return nil
}

// Rollback rolls back the transaction (no-op)
func (t *d1Tx) Rollback() error {
	// No-op - D1 doesn't support traditional transactions
	return nil
}
