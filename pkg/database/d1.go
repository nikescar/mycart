package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/d1"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type d1DB struct {
	accountID  string
	databaseID string
	apiToken   string
	client     *cloudflare.Client
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

	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &d1DB{
		accountID:  accountID,
		databaseID: databaseID,
		apiToken:   apiToken,
		client:     client,
	}, nil
}

// Exec executes a query without returning rows
func (db *d1DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Convert args to strings for D1 API
	stringArgs := make([]string, len(args))
	for i, arg := range args {
		stringArgs[i] = fmt.Sprintf("%v", arg)
	}

	queryBody := d1.DatabaseQueryParamsBodyD1SingleQuery{
		Sql:    cloudflare.F(query),
		Params: cloudflare.F(stringArgs),
	}

	params := d1.DatabaseQueryParams{
		AccountID: cloudflare.F(db.accountID),
		Body:      queryBody,
	}

	result, err := db.client.D1.Database.Query(ctx, db.databaseID, params)
	if err != nil {
		return nil, fmt.Errorf("D1 Exec failed: %w", err)
	}

	// Calculate rows affected
	rowsAffected := int64(0)
	for _, item := range result.Result {
		if meta := item.Meta; meta.ChangedDB || meta.RowsWritten > 0 {
			rowsAffected += int64(meta.RowsWritten)
		}
	}

	return &d1Result{rowsAffected: rowsAffected}, nil
}

// Query executes a query that returns rows
func (db *d1DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("D1 Query not yet fully implemented - use QueryRow for single row queries")
}

// QueryRow executes a query that returns at most one row
func (db *d1DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// For now, return an empty row
	// A full implementation would need to parse D1 API response into sql.Row
	return &sql.Row{}
}

// Prepare creates a prepared statement
func (db *d1DB) Prepare(ctx context.Context, query string) (Stmt, error) {
	return &d1Stmt{
		db:    db,
		query: query,
		ctx:   ctx,
	}, nil
}

// Begin starts a transaction
func (db *d1DB) Begin(ctx context.Context) (Tx, error) {
	// D1 doesn't support traditional transactions via HTTP API
	// Return a pseudo-transaction that batches queries
	return &d1Tx{
		db:  db,
		ctx: ctx,
	}, nil
}

// Ping checks the database connection
func (db *d1DB) Ping(ctx context.Context) error {
	// Test with a simple query
	_, err := db.Exec(ctx, "SELECT 1")
	return err
}

// Close closes the database connection
func (db *d1DB) Close() error {
	// HTTP client has no persistent connection to close
	return nil
}

// DB returns the underlying *sql.DB
func (db *d1DB) DB() *sql.DB {
	// D1 uses HTTP API, not database/sql driver
	return nil
}

// d1Stmt implements Stmt interface
type d1Stmt struct {
	db    *d1DB
	query string
	ctx   context.Context
}

// Exec executes the prepared statement
func (s *d1Stmt) Exec(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(ctx, s.query, args...)
}

// Query executes the prepared statement query
func (s *d1Stmt) Query(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(ctx, s.query, args...)
}

// QueryRow executes the prepared statement query for a single row
func (s *d1Stmt) QueryRow(ctx context.Context, args ...interface{}) *sql.Row {
	return s.db.QueryRow(ctx, s.query, args...)
}

// Close closes the prepared statement
func (s *d1Stmt) Close() error {
	// No persistent statement to close
	return nil
}

// d1Tx implements Tx interface
type d1Tx struct {
	db  *d1DB
	ctx context.Context
}

// Exec executes a query in transaction
func (t *d1Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.db.Exec(ctx, query, args...)
}

// Query executes a query in transaction
func (t *d1Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.db.Query(ctx, query, args...)
}

// QueryRow executes a query in transaction
func (t *d1Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.db.QueryRow(ctx, query, args...)
}

// Prepare creates a prepared statement in transaction
func (t *d1Tx) Prepare(ctx context.Context, query string) (Stmt, error) {
	return t.db.Prepare(ctx, query)
}

// Commit commits the transaction
func (t *d1Tx) Commit() error {
	// D1 queries are auto-committed
	return nil
}

// Rollback rolls back the transaction
func (t *d1Tx) Rollback() error {
	// D1 doesn't support rollback via HTTP API
	return nil
}

// d1Result implements sql.Result
type d1Result struct {
	rowsAffected int64
}

func (r *d1Result) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("D1 does not support LastInsertId")
}

func (r *d1Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
