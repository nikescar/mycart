-- name: GetProductByID :one
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product WHERE id = $1 AND deleted = FALSE LIMIT 1;

-- name: GetProductBySlug :one
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product WHERE slug = $1 AND deleted = FALSE LIMIT 1;

-- name: ListProducts :many
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE deleted = FALSE
ORDER BY created DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveProducts :many
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE active = TRUE AND deleted = FALSE
ORDER BY created DESC
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM product WHERE deleted = FALSE;

-- name: CreateProduct :one
INSERT INTO product (id, name, "desc", slug, amount, metadata, attribute, digital, active, created)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
RETURNING id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated;

-- name: UpdateProduct :exec
UPDATE product
SET name = $1, "desc" = $2, slug = $3, amount = $4, metadata = $5, attribute = $6, digital = $7, active = $8, updated = NOW()
WHERE id = $9;

-- name: UpdateProductActive :exec
UPDATE product
SET active = NOT active, updated = NOW()
WHERE id = $1;

-- name: SoftDeleteProduct :exec
UPDATE product
SET deleted = TRUE, updated = NOW()
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM product WHERE id = $1;

-- name: ProductExists :one
SELECT EXISTS(SELECT 1 FROM product WHERE slug = $1 AND deleted = FALSE);

-- name: ListAllProducts :many
SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product ORDER BY created;

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
WHERE p.id = $1;

-- name: CreateProductVariant :one
INSERT INTO product_variant (id, product_id, sku, price_surcharge, quantity, option_values)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, product_id, sku, price_surcharge, quantity, option_values, active, deleted, created, updated;

-- name: UpdateProductVariant :exec
UPDATE product_variant
SET sku = $2, price_surcharge = $3, quantity = $4, option_values = $5, updated = NOW()
WHERE id = $1;

-- name: DeleteProductVariant :exec
DELETE FROM product_variant WHERE id = $1;

-- name: ListProductVariantsByProduct :many
SELECT id, product_id, sku, price_surcharge, quantity, option_values, active, deleted, created, updated
FROM product_variant
WHERE product_id = $1;

-- name: GetProductOption :one
SELECT id, name, product_id, position, created
FROM product_option
WHERE id = $1 LIMIT 1;

-- name: CreateProductOption :one
INSERT INTO product_option (id, name, product_id, position)
VALUES ($1, $2, $3, $4)
RETURNING id, name, product_id, position, created;

-- name: DeleteProductOption :exec
DELETE FROM product_option WHERE id = $1;

-- Advanced Product Queries

-- name: BulkDeleteProductImages :exec
DELETE FROM product_image
WHERE product_id = $1;

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
