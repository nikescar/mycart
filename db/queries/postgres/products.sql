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
