-- name: GetProductByID :one
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product WHERE id = ? AND deleted = FALSE LIMIT 1;

-- name: GetProductBySlug :one
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product WHERE slug = ? AND deleted = FALSE LIMIT 1;

-- name: ListProducts :many
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE deleted = FALSE
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListActiveProducts :many
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE active = TRUE AND deleted = FALSE
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountProducts :one
SELECT COUNT(*) FROM product WHERE deleted = FALSE;

-- name: CreateProduct :one
INSERT INTO product (id, name, "desc", slug, amount, metadata, attribute, digital, active, created)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated;

-- name: UpdateProduct :exec
UPDATE product
SET name = ?, "desc" = ?, slug = ?, amount = ?, metadata = ?, attribute = ?, digital = ?, active = ?, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateProductActive :exec
UPDATE product
SET active = NOT active, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SoftDeleteProduct :exec
UPDATE product
SET deleted = TRUE, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteProduct :exec
DELETE FROM product WHERE id = ?;

-- name: ProductExists :one
SELECT EXISTS(SELECT 1 FROM product WHERE slug = ? AND deleted = FALSE);

-- Product Variant Queries

-- name: GetProductWithVariants :many
SELECT
    p.id as product_id,
    p.name as product_name,
    pv.id as variant_id,
    pv.sku as variant_sku,
    pv.price_surcharge as variant_price_surcharge,
    pv.quantity as variant_quantity,
    pv.option_values as variant_option_values
FROM product p
LEFT JOIN product_variant pv ON p.id = pv.product_id
WHERE p.id = ?;

-- name: CreateProductVariant :one
INSERT INTO product_variant (id, product_id, sku, price_surcharge, quantity, option_values)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, product_id, sku, price_surcharge, quantity, option_values, active, deleted, created, updated;

-- name: UpdateProductVariant :exec
UPDATE product_variant
SET sku = ?, price_surcharge = ?, quantity = ?, option_values = ?, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteProductVariant :exec
DELETE FROM product_variant WHERE id = ?;

-- name: ListProductVariantsByProduct :many
SELECT id, product_id, sku, price_surcharge, quantity, option_values, active, deleted, created, updated
FROM product_variant
WHERE product_id = ?;

-- name: GetProductOption :one
SELECT id, name, product_id, position, created
FROM product_option
WHERE id = ? LIMIT 1;

-- name: CreateProductOption :one
INSERT INTO product_option (id, name, product_id, position)
VALUES (?, ?, ?, ?)
RETURNING id, name, product_id, position, created;

-- name: DeleteProductOption :exec
DELETE FROM product_option WHERE id = ?;

-- Advanced Product Queries

-- name: BulkDeleteProductImages :exec
DELETE FROM product_image
WHERE product_id = ?;

-- name: GetProductsWithImages :many
SELECT
    p.id as product_id,
    p.name as product_name,
    pi.id as image_id,
    pi.name as image_name,
    pi.ext as image_ext
FROM product p
LEFT JOIN product_image pi ON p.id = pi.product_id
WHERE p.active = TRUE AND p.deleted = FALSE
ORDER BY p.created DESC;
