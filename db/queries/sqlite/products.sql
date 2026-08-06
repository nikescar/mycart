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
