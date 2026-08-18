package store

import (
	"context"

	"github.com/shurco/mycart/internal/goosemigration/queries"
	"github.com/shurco/mycart/internal/models"
)

// Carts retrieves a paginated list of carts.
func Carts(ctx context.Context, limit, offset int) ([]*models.Cart, int, error) {
	return queries.DB().CartQueries.Carts(ctx, limit, offset)
}

// Cart retrieves a cart by ID with full details.
func Cart(ctx context.Context, cartID string) (*models.Cart, error) {
	return queries.DB().CartQueries.Cart(ctx, cartID)
}

// BuildCartItems builds cart items from cart and products data.
// This is a utility function for combining cart and product information.
func BuildCartItems(cart *models.Cart, products *models.Products) []map[string]any {
	return queries.BuildCartItems(cart, products)
}
