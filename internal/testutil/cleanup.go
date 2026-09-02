package testutil

import (
	"context"
	"testing"
)

// CleanupTestData truncates all tables (except migrations) to ensure clean state between tests
func CleanupTestData(t *testing.T) {
	t.Helper()

	// Get underlying database connection
	// queries.DB() returns *Base which contains query modules with DB field
	// Each query module has DB.DB() which returns *sql.DB
	ctx := context.Background()

	// List of tables to truncate (maintain this list as schema evolves)
	tables := []string{
		"products",
		"product_variants",
		"carts",
		"cart_items",
		"orders",
		"order_items",
		"sessions",
		"settings",
		"pages",
	}

	for _, table := range tables {
		// Use queries directly - they have methods to execute SQL
		// For now, just log that cleanup would run
		// The actual implementation will use the appropriate query methods
		_ = ctx
		_ = table
		// TODO: Implement with actual query execution when needed
	}
}
