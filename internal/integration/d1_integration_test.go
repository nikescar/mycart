package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/internal/queries"
	"github.com/shurco/mycart/internal/testutil"
)

// TestD1Integration tests the application against real Cloudflare D1 database
// when USE_CLOUDFLARE_TESTS=1 is set. Falls back to SQLite otherwise.
//
// This test verifies:
// - Database connection and migrations
// - Basic CRUD operations
// - Query functionality
// - Integration with real Cloudflare infrastructure
//
// Run with: USE_CLOUDFLARE_TESTS=1 go test ./internal/integration -v
func TestD1Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// SetupTestEnv auto-detects environment:
	// - With CLOUDFLARE_TEST_ACCOUNT_ID: creates real D1 database via API
	// - Without: uses in-memory SQLite
	cleanup := testutil.SetupTestEnv(t)
	defer cleanup()

	db := queries.DB()
	if db == nil {
		t.Fatal("queries.DB() returned nil after setup")
	}

	// Log which database backend we're testing against
	dbType := db.Type()
	t.Logf("Testing against database type: %s", dbType)
	if os.Getenv("USE_CLOUDFLARE_TESTS") == "1" {
		if dbType != "d1" {
			t.Errorf("Expected D1 database with USE_CLOUDFLARE_TESTS=1, got %s", dbType)
		}
		t.Log("✓ Successfully connected to real Cloudflare D1 database")
	} else {
		if dbType != "sqlite" {
			t.Errorf("Expected SQLite database in local mode, got %s", dbType)
		}
		t.Log("✓ Running against local SQLite (fast mode)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Verify migrations ran successfully
	t.Run("migrations_applied", func(t *testing.T) {
		setting, err := db.GetSettingByKey(ctx, "installed")
		if err != nil {
			t.Fatalf("GetSettingByKey failed: %v", err)
		}
		if _, ok := setting["installed"]; !ok {
			t.Error("installed key not found - migrations may not have run")
		}
		t.Log("✓ Migrations applied successfully")
	})

	// Test 2: CRUD operations on pages
	t.Run("page_crud_operations", func(t *testing.T) {
		// Create
		page, err := db.AddPage(ctx, &models.Page{
			Name:     "Integration Test Page",
			Slug:     "integration-test",
			Position: "footer",
		})
		if err != nil {
			t.Fatalf("AddPage failed: %v", err)
		}
		if page.ID == "" {
			t.Error("AddPage returned page with empty ID")
		}
		t.Logf("✓ Created page with ID: %s", page.ID)

		// Read
		pages, _, err := db.ListPages(ctx, true, 10, 0)
		if err != nil {
			t.Fatalf("ListPages failed: %v", err)
		}
		if len(pages) == 0 {
			t.Error("ListPages returned no pages after creation")
		}
		t.Logf("✓ Listed %d pages", len(pages))

		// Update
		content := "Test content for integration"
		err = db.UpdatePageContent(ctx, &models.Page{
			Core:    models.Core{ID: page.ID},
			Content: &content,
		})
		if err != nil {
			t.Fatalf("UpdatePageContent failed: %v", err)
		}
		t.Log("✓ Updated page content")

		// Delete
		err = db.DeletePage(ctx, page.ID)
		if err != nil {
			t.Fatalf("DeletePage failed: %v", err)
		}
		t.Log("✓ Deleted page")
	})

	// Test 3: Settings operations
	t.Run("settings_operations", func(t *testing.T) {
		// Update a setting
		err := db.UpdateSettingByKey(ctx, &models.SettingName{
			Key:   "site_name",
			Value: "D1 Integration Test Site",
		})
		if err != nil {
			t.Fatalf("UpdateSettingByKey failed: %v", err)
		}

		// Read it back
		settings, err := db.GetSettingByKey(ctx, "site_name")
		if err != nil {
			t.Fatalf("GetSettingByKey failed: %v", err)
		}

		siteName, ok := settings["site_name"]
		if !ok {
			t.Error("site_name not found in settings")
		}
		if siteName.Value != "D1 Integration Test Site" {
			t.Errorf("site_name = %v, want 'D1 Integration Test Site'", siteName.Value)
		}
		t.Log("✓ Settings read/write working")
	})

	// Final summary
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if dbType == "d1" {
		t.Log("✅ D1 INTEGRATION TEST PASSED")
		t.Log("   Real Cloudflare D1 database operations verified")
	} else {
		t.Log("✅ INTEGRATION TEST PASSED (SQLite mode)")
		t.Log("   Run with USE_CLOUDFLARE_TESTS=1 to test D1")
	}
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// TestD1Performance benchmarks basic operations against D1/SQLite
func TestD1Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cleanup := testutil.SetupTestEnv(t)
	defer cleanup()

	db := queries.DB()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dbType := db.Type()
	t.Logf("Performance testing against: %s", dbType)

	// Benchmark page creation
	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := db.AddPage(ctx, &models.Page{
			Name:     "Perf Test",
			Slug:     fmt.Sprintf("perf-test-%d-%d", time.Now().UnixNano(), i),
			Position: "footer",
		})
		if err != nil {
			t.Fatalf("AddPage iteration %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	t.Logf("Created 10 pages in %v (avg: %v per page)", elapsed, elapsed/10)

	if dbType == "d1" {
		t.Log("Note: D1 performance includes network latency to Cloudflare")
	}
}
