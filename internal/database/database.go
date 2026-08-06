package database

import (
	"context"
	"database/sql"

	"github.com/shurco/mycart/internal/database/sqlc"
)

// Database wraps sqlc.Queries and provides transaction support
type Database struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// New creates a new Database instance from an existing *sql.DB connection
func New(db *sql.DB) *Database {
	return &Database{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Queries returns the underlying sqlc.Queries instance
func (d *Database) Queries() *sqlc.Queries {
	return d.queries
}

// DB returns the underlying *sql.DB connection
func (d *Database) DB() *sql.DB {
	return d.db
}

// WithTx executes a function within a database transaction
func (d *Database) WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
