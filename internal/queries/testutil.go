package queries

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/shurco/mycart/db/migrations"
)

// runOnBothDBs runs the test function against both SQLite and PostgreSQL.
// PostgreSQL tests are skipped in short mode.
func runOnBothDBs(t *testing.T, testFunc func(*testing.T, *Base, context.Context)) {
	t.Run("SQLite", func(t *testing.T) {
		db, ctx := bootstrapSQLite(t)
		testFunc(t, db, ctx)
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping PostgreSQL test in short mode")
		}
		db, ctx := bootstrapPostgreSQL(t)
		testFunc(t, db, ctx)
	})
}

// bootstrapSQLite creates an in-memory SQLite database for testing.
func bootstrapSQLite(t *testing.T) (*Base, context.Context) {
	t.Helper()

	// Set up in-memory SQLite
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("SQLITE_PATH", ":memory:")

	// Initialize
	if err := New(migrations.Embed()); err != nil {
		t.Fatalf("failed to initialize SQLite: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(func() {
		cancel()
		Close()
	})

	return DB(), ctx
}

// bootstrapPostgreSQL creates a test PostgreSQL database connection.
func bootstrapPostgreSQL(t *testing.T) (*Base, context.Context) {
	t.Helper()

	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL tests")
	}

	// Set up PostgreSQL
	os.Setenv("DB_TYPE", "postgresql")
	os.Setenv("DATABASE_URL", testDBURL)

	// Initialize
	if err := New(migrations.Embed()); err != nil {
		t.Fatalf("failed to initialize PostgreSQL: %v", err)
	}

	// Clean up test data before each test
	baseDB := DB()
	cleanTestData(t, baseDB.SettingQueries.DB)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(func() {
		cancel()
		Close()
	})

	return DB(), ctx
}

// cleanTestData removes all data from test database (keeps schema).
func cleanTestData(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"cart_item", "cart", "product_variant_option", "product_variant_image",
		"product_variant", "product_option_value", "product_option",
		"product_image", "digital_data", "digital_file", "product",
		"page", "session", "subdomain", "setting",
	}

	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			t.Logf("warning: failed to clean %s: %v", table, err)
		}
	}
}
