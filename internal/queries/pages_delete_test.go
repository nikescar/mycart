package queries

import (
	"context"
	"testing"

	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/pkg/security"
)

func TestDeletePage(t *testing.T) {
	runOnBothDBs(t, func(t *testing.T, db *Base, ctx context.Context) {
		// Create a test page
		content := "Test content"
		page := &models.Page{
			Core:     models.Core{ID: security.RandomString()},
			Name:     "Test Page",
			Slug:     "test-page-delete",
			Content:  &content,
			Position: "footer",
			Active:   true,
		}

		if _, err := db.AddPage(ctx, page); err != nil {
			t.Fatalf("failed to create page: %v", err)
		}

		// Delete the page
		if err := db.DeletePage(ctx, page.ID); err != nil {
			t.Fatalf("DeletePage failed: %v", err)
		}

		// Verify it's gone
		_, err := db.Page(ctx, page.Slug)
		if err == nil {
			t.Error("expected error fetching deleted page")
		}
	})
}

func TestDeletePage_NonExistent(t *testing.T) {
	runOnBothDBs(t, func(t *testing.T, db *Base, ctx context.Context) {
		// Try to delete a page that doesn't exist
		err := db.DeletePage(ctx, "non-existent-id")

		// Should not error (idempotent delete)
		if err != nil {
			t.Errorf("DeletePage for non-existent ID should not error, got: %v", err)
		}
	})
}
