package database

import (
	"context"
	"database/sql"
)

// Database is the unified interface for database operations.
// Both SQLite and PostgreSQL adapters implement this interface.
type Database interface {
	// DB returns the underlying *sql.DB connection
	DB() *sql.DB

	// Close closes the database connection
	Close() error

	// BeginTx starts a database transaction
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Queries returns a database-specific queries object
	// The type will be *sqlite.Queries or *postgres.Queries
	Queries() interface{}

	// Ping verifies the connection to the database is still alive
	Ping(ctx context.Context) error

	// Stats returns database statistics
	Stats() sql.DBStats
}
