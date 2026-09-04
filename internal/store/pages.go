package store

import (
	"context"

	"github.com/shurco/mycart/internal/goosemigration/queries"
	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/internal/store/db"
)

// Page retrieves a page by slug for public access.
func Page(ctx context.Context, slug string) (*models.Page, error) {
	return queries.DB().PageQueries.Page(ctx, slug)
}

// ListPages retrieves a paginated list of pages.
func ListPages(ctx context.Context, private bool, limit, offset int, idList ...string) ([]models.Page, int, error) {
	return queries.DB().PageQueries.ListPages(ctx, private, limit, offset, idList...)
}

// PageByID retrieves a page by ID.
func PageByID(ctx context.Context, id string) (*models.Page, error) {
	return queries.DB().PageQueries.PageByID(ctx, id)
}

// AddPage creates a new page.
func AddPage(ctx context.Context, page *models.Page) (*models.Page, error) {
	return queries.DB().PageQueries.AddPage(ctx, page)
}

// UpdatePage updates an existing page.
func UpdatePage(ctx context.Context, page *models.Page) error {
	return queries.DB().PageQueries.UpdatePage(ctx, page)
}

// DeletePage deletes a page by ID.
func DeletePage(ctx context.Context, id string) error {
	return queries.DB().PageQueries.DeletePage(ctx, id)
}

// UpdatePageContent updates page content.
func UpdatePageContent(ctx context.Context, page *models.Page) error {
	return queries.DB().PageQueries.UpdatePageContent(ctx, page)
}

// UpdatePageActive toggles page active status.
func UpdatePageActive(ctx context.Context, id string) error {
	return queries.DB().PageQueries.UpdatePageActive(ctx, id)
}

// IsPage checks if a page exists by slug.
func IsPage(ctx context.Context, slug string) bool {
	return queries.DB().PageQueries.IsPage(ctx, slug)
}

// --- sqlc-based methods (new implementation) ---

// GetPageBySlug retrieves a page by slug using sqlc function pointer.
func GetPageBySlug(ctx context.Context, slug string) (db.Page, error) {
	return db.GetPageBySlugFunc(ctx, slug)
}

// ListPagesSqlc retrieves a paginated list of pages using sqlc function pointer.
func ListPagesSqlc(ctx context.Context, limit, offset int32) ([]db.Page, error) {
	return db.ListPagesFunc(ctx, limit, offset)
}

// CreatePage creates a new page using sqlc function pointer.
func CreatePage(ctx context.Context, params db.CreatePageParams) (db.Page, error) {
	return db.CreatePageFunc(ctx, params)
}

// UpdatePageSqlc updates a page using sqlc function pointer.
func UpdatePageSqlc(ctx context.Context, params db.UpdatePageParams) error {
	return db.UpdatePageFunc(ctx, params)
}

// DeletePageSqlc deletes a page using sqlc function pointer.
func DeletePageSqlc(ctx context.Context, id string) error {
	return db.DeletePageFunc(ctx, id)
}
