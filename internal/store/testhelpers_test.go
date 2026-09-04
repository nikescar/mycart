package store_test

import (
	"context"
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
