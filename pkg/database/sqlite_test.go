package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLite_New(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()
}

func TestSQLite_Exec(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Create table
	_, err = db.Exec(ctx, `CREATE TABLE users (id TEXT PRIMARY KEY, email TEXT NOT NULL)`)
	require.NoError(t, err)

	// Insert data
	result, err := db.Exec(ctx, "INSERT INTO users (id, email) VALUES (?, ?)", "1", "test@example.com")
	require.NoError(t, err)

	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)
}

func TestSQLite_Query(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Setup
	_, err = db.Exec(ctx, `CREATE TABLE users (id TEXT, email TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO users VALUES ('1', 'test@example.com')`)
	require.NoError(t, err)

	// Test Query
	rows, err := db.Query(ctx, "SELECT id, email FROM users WHERE id = ?", "1")
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())

	var id, email string
	err = rows.Scan(&id, &email)
	require.NoError(t, err)
	require.Equal(t, "1", id)
	require.Equal(t, "test@example.com", email)
}

func TestSQLite_QueryRow(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Setup
	_, err = db.Exec(ctx, `CREATE TABLE users (id TEXT, email TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `INSERT INTO users VALUES ('1', 'test@example.com')`)
	require.NoError(t, err)

	// Test QueryRow
	var email string
	err = db.QueryRow(ctx, "SELECT email FROM users WHERE id = ?", "1").Scan(&email)
	require.NoError(t, err)
	require.Equal(t, "test@example.com", email)
}

func TestSQLite_Transaction(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Setup
	_, err = db.Exec(ctx, `CREATE TABLE users (id TEXT, email TEXT)`)
	require.NoError(t, err)

	// Begin transaction
	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Insert in transaction
	_, err = tx.Exec(ctx, "INSERT INTO users VALUES (?, ?)", "1", "test@example.com")
	require.NoError(t, err)

	// Commit
	err = tx.Commit()
	require.NoError(t, err)

	// Verify data persisted
	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestSQLite_Transaction_Rollback(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Setup
	_, err = db.Exec(ctx, `CREATE TABLE users (id TEXT, email TEXT)`)
	require.NoError(t, err)

	// Begin transaction
	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Insert in transaction
	_, err = tx.Exec(ctx, "INSERT INTO users VALUES (?, ?)", "1", "test@example.com")
	require.NoError(t, err)

	// Rollback
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify data NOT persisted
	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestSQLite_Ping(t *testing.T) {
	t.Parallel()

	db, err := NewSQLite(":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	require.NoError(t, err)
}
