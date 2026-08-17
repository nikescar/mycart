-- name: GetCart :one
SELECT id, email, amount_total, currency, payment_id, payment_status, CAST(cart AS TEXT) as cart, payment_system, created, updated
FROM cart WHERE id = ? LIMIT 1;

-- name: ListCarts :many
SELECT id, email, amount_total, currency, payment_id, payment_status, CAST(cart AS TEXT) as cart, payment_system, created, updated
FROM cart
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountCarts :one
SELECT COUNT(*) FROM cart;

-- name: CreateCart :one
INSERT INTO cart (id, email, amount_total, currency, payment_id, payment_status, cart, payment_system, created)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING id, email, amount_total, currency, payment_id, payment_status, CAST(cart AS TEXT) as cart, payment_system, created, updated;

-- name: UpdateCart :exec
UPDATE cart
SET email = ?, amount_total = ?, currency = ?, payment_id = ?, payment_status = ?, cart = ?, payment_system = ?, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateCartPaymentStatus :exec
UPDATE cart
SET payment_status = ?, updated = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteCart :exec
DELETE FROM cart WHERE id = ?;

-- name: ListAllCarts :many
SELECT id, email, amount_total, currency, payment_id, payment_status, CAST(cart AS TEXT) as cart, payment_system, created, updated
FROM cart ORDER BY created;
