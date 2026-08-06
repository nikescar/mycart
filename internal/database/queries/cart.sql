-- Cart queries
-- name: GetCartByID :one
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE id = ?
LIMIT 1;

-- name: ListCarts :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountCarts :one
SELECT COUNT(*) as count
FROM cart;

-- name: CreateCart :exec
INSERT INTO cart (id, email, amount_total, currency, payment_system, created)
VALUES (?, ?, ?, ?, ?, datetime('now'));

-- name: UpdateCartPayment :exec
UPDATE cart
SET payment_id = ?,
    payment_status = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: UpdateCartTotal :exec
UPDATE cart
SET amount_total = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: DeleteCart :exec
DELETE FROM cart
WHERE id = ?;

-- Cart Product queries
-- name: GetCartProducts :many
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE cart_id = ?
ORDER BY id;

-- name: GetCartProductByID :one
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE id = ?
LIMIT 1;

-- name: CreateCartProduct :exec
INSERT INTO cart_product (id, cart_id, product_id, quantity, amount)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateCartProductQuantity :exec
UPDATE cart_product
SET quantity = ?,
    amount = ?
WHERE id = ?;

-- name: DeleteCartProduct :exec
DELETE FROM cart_product
WHERE id = ?;

-- name: DeleteCartProducts :exec
DELETE FROM cart_product
WHERE cart_id = ?;

-- name: GetCartProductByCartAndProduct :one
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE cart_id = ? AND product_id = ?
LIMIT 1;

-- Payment queries
-- name: ListCartsWithPaymentStatus :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE payment_status = ?
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListCartsByPaymentSystem :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE payment_system = ?
ORDER BY created DESC
LIMIT ? OFFSET ?;
