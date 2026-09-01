package database

import (
	"context"
	"database/sql"
)

// Database defines the interface for database operations
type Database interface {
	// Core operations
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row

	// Transaction support
	Begin(ctx context.Context) (Tx, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error

	// Migration support - returns underlying *sql.DB for goose
	DB() *sql.DB
}

// Tx defines transaction interface
type Tx interface {
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}

// Config holds database configuration
type Config struct {
	Type string // "sqlite" or "d1"

	// SQLite config
	Path string

	// D1 config
	AccountID  string
	DatabaseID string
	APIToken   string
}
