package store

import (
	"context"

	"github.com/shurco/mycart/internal/goosemigration/queries"
	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/internal/store/db"
)

// ListProducts retrieves a list of products with optional filtering.
func ListProducts(ctx context.Context, private bool, limit, offset int, cartID string, idList ...models.CartProduct) (*models.Products, error) {
	return queries.DB().ProductQueries.ListProducts(ctx, private, limit, offset, cartID, idList...)
}

// Product retrieves a single product by ID.
func Product(ctx context.Context, private bool, productID string) (*models.Product, error) {
	return queries.DB().ProductQueries.Product(ctx, private, productID)
}

// AddProduct creates a new product without variants.
func AddProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	return queries.DB().ProductQueries.AddProduct(ctx, product)
}

// AddProductWithVariants creates a new product with variants.
func AddProductWithVariants(ctx context.Context, product *models.Product) (*models.Product, error) {
	return queries.DB().ProductQueries.AddProductWithVariants(ctx, product)
}

// UpdateProduct updates an existing product.
func UpdateProduct(ctx context.Context, product *models.Product) error {
	return queries.DB().ProductQueries.UpdateProduct(ctx, product)
}

// DeleteProduct deletes a product by ID.
func DeleteProduct(ctx context.Context, productID string) error {
	return queries.DB().ProductQueries.DeleteProduct(ctx, productID)
}

// UpdateActive toggles product active status.
func UpdateActive(ctx context.Context, productID string) error {
	return queries.DB().ProductQueries.UpdateActive(ctx, productID)
}

// ProductImages retrieves all images for a product.
func ProductImages(ctx context.Context, productID string) (*[]models.File, error) {
	return queries.DB().ProductQueries.ProductImages(ctx, productID)
}

// AddImage adds an image to a product.
func AddImage(ctx context.Context, productID, fileUUID, fileExt, fileOrigName string) (*models.File, error) {
	return queries.DB().ProductQueries.AddImage(ctx, productID, fileUUID, fileExt, fileOrigName)
}

// DeleteImage deletes an image from a product.
func DeleteImage(ctx context.Context, productID, imageID string) error {
	return queries.DB().ProductQueries.DeleteImage(ctx, productID, imageID)
}

// ProductDigital retrieves digital content for a product.
func ProductDigital(ctx context.Context, productID string) (*models.Digital, error) {
	return queries.DB().ProductQueries.ProductDigital(ctx, productID)
}

// AddDigitalFile adds a digital file to a product.
func AddDigitalFile(ctx context.Context, productID, fileUUID, fileExt, fileOrigName string) (*models.File, error) {
	return queries.DB().ProductQueries.AddDigitalFile(ctx, productID, fileUUID, fileExt, fileOrigName)
}

// AddDigitalData adds digital data (license key) to a product.
func AddDigitalData(ctx context.Context, productID, content string) (*models.Data, error) {
	return queries.DB().ProductQueries.AddDigitalData(ctx, productID, content)
}

// UpdateDigital updates digital content.
func UpdateDigital(ctx context.Context, digital *models.Data) error {
	return queries.DB().ProductQueries.UpdateDigital(ctx, digital)
}

// DeleteDigital deletes digital content from a product.
func DeleteDigital(ctx context.Context, productID, digitalID string) error {
	return queries.DB().ProductQueries.DeleteDigital(ctx, productID, digitalID)
}

// GenerateUniqueSlug generates a unique slug for a product.
func GenerateUniqueSlug(ctx context.Context, name, excludeID string) (string, error) {
	return queries.DB().ProductQueries.GenerateUniqueSlug(ctx, name, excludeID)
}

// ProductQueriesDB returns the underlying ProductQueries for CSV import operations.
func ProductQueriesDB() *queries.ProductQueries {
	return &queries.DB().ProductQueries
}

// --- sqlc-based methods (new implementation) ---

// GetProductByID retrieves a product by ID using sqlc function pointer.
func GetProductByID(ctx context.Context, id string) (db.Product, error) {
	return db.GetProductByIDFunc(ctx, id)
}

// GetProductBySlug retrieves a product by slug using sqlc function pointer.
func GetProductBySlug(ctx context.Context, slug string) (db.Product, error) {
	return db.GetProductBySlugFunc(ctx, slug)
}

// CreateProduct creates a new product using sqlc function pointer.
func CreateProduct(ctx context.Context, params db.CreateProductParams) (db.Product, error) {
	return db.CreateProductFunc(ctx, params)
}

// UpdateProductSqlc updates a product using sqlc function pointer.
func UpdateProductSqlc(ctx context.Context, params db.UpdateProductParams) error {
	return db.UpdateProductFunc(ctx, params)
}

// DeleteProductSqlc deletes a product using sqlc function pointer.
func DeleteProductSqlc(ctx context.Context, id string) error {
	return db.DeleteProductFunc(ctx, id)
}
