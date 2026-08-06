-- name: GetAuthCredentials :one
SELECT key, value
FROM setting
WHERE key IN ('email', 'password')
LIMIT 2;

-- name: GetEmailSetting :one
SELECT id, key, value
FROM setting
WHERE key = 'email'
LIMIT 1;

-- name: GetPasswordSetting :one
SELECT id, key, value
FROM setting
WHERE key = 'password'
LIMIT 1;

-- name: UpdatePassword :exec
UPDATE setting
SET value = ?
WHERE key = 'password';
