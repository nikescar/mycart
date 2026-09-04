package store_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/shurco/mycart/db/migrations"
	"github.com/shurco/mycart/internal/goosemigration/queries"
	"github.com/shurco/mycart/internal/store"
	"github.com/shurco/mycart/internal/store/db"
	"github.com/stretchr/testify/require"
)

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) context.Context {
	t.Helper()

	// Set up environment for SQLite
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("SQLITE_PATH", ":memory:")

	// Initialize database
	err := queries.New(migrations.Embed())
	require.NoError(t, err)

	// Initialize store layer
	err = db.Init(queries.Adapter().DB(), "sqlite")
	require.NoError(t, err)
	store.InitStore(queries.Adapter().DB())

	t.Cleanup(func() {
		queries.Close()
	})

	return context.Background()
}

func TestGetSession(t *testing.T) {
	ctx := setupTestDB(t)

	// Add a session
	key := "test_key"
	value := "test_value"
	expires := int64(9999999999)
	err := store.AddSession(ctx, key, value, expires)
	require.NoError(t, err)

	// Retrieve session
	session, err := store.GetSession(ctx, key)
	require.NoError(t, err)
	require.Equal(t, key, session.Key)
	require.True(t, session.Value.Valid)
	require.Equal(t, value, session.Value.String)
	require.True(t, session.Expires.Valid)
	require.Equal(t, expires, session.Expires.Int64)
}

func TestGetSession_NotFound(t *testing.T) {
	ctx := setupTestDB(t)

	_, err := store.GetSession(ctx, "nonexistent")
	require.Error(t, err)
	require.Equal(t, sql.ErrNoRows, err)
}
