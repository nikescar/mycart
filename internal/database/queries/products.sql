-- Product queries
-- name: GetProductByID :one
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE id = ? AND deleted = 0
LIMIT 1;

-- name: GetProductBySlug :one
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE slug = ? AND deleted = 0
LIMIT 1;

-- name: ListActiveProducts :many
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE active = 1 AND deleted = 0
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListAllProducts :many
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE deleted = 0
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountActiveProducts :one
SELECT COUNT(*) as count
FROM product
WHERE active = 1 AND deleted = 0;

-- name: CountAllProducts :one
SELECT COUNT(*) as count
FROM product
WHERE deleted = 0;

-- name: CreateProduct :exec
INSERT INTO product (id, name, desc, slug, amount, metadata, attribute, digital, active, created)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'));

-- name: UpdateProduct :exec
UPDATE product
SET name = ?,
    desc = ?,
    amount = ?,
    metadata = ?,
    attribute = ?,
    active = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: SoftDeleteProduct :exec
UPDATE product
SET deleted = 1,
    updated = datetime('now')
WHERE id = ?;

-- name: CheckProductSlugExists :one
SELECT COUNT(*) as count
FROM product
WHERE slug = ? AND id != ? AND deleted = 0;

-- Product Image queries
-- name: GetProductImages :many
SELECT id, product_id, name, ext, orig_name
FROM product_image
WHERE product_id = ?
ORDER BY id;

-- name: GetProductImageByID :one
SELECT id, product_id, name, ext, orig_name
FROM product_image
WHERE id = ?
LIMIT 1;

-- name: CreateProductImage :exec
INSERT INTO product_image (id, product_id, name, ext, orig_name)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteProductImage :exec
DELETE FROM product_image
WHERE id = ?;

-- name: DeleteProductImages :exec
DELETE FROM product_image
WHERE product_id = ?;

-- Digital File queries
-- name: GetDigitalFiles :many
SELECT id, product_id, name, ext, orig_name
FROM digital_file
WHERE product_id = ?
ORDER BY id;

-- name: CreateDigitalFile :exec
INSERT INTO digital_file (id, product_id, name, ext, orig_name)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteDigitalFile :exec
DELETE FROM digital_file
WHERE id = ?;

-- Digital Data queries
-- name: GetDigitalData :one
SELECT id, product_id, content, cart_id
FROM digital_data
WHERE product_id = ? AND (cart_id IS NULL OR cart_id = ?)
LIMIT 1;

-- name: CreateDigitalData :exec
INSERT INTO digital_data (id, product_id, content, cart_id)
VALUES (?, ?, ?, ?);

-- name: UpdateDigitalDataCart :exec
UPDATE digital_data
SET cart_id = ?
WHERE id = ?;
