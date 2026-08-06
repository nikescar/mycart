-- name: GetSettingByKey :one
SELECT id, key, value
FROM setting
WHERE key = $1 LIMIT 1;

-- name: UpdateSetting :exec
UPDATE setting
SET value = $1
WHERE key = $2;

-- name: ListSettings :many
SELECT id, key, value
FROM setting
ORDER BY key;

-- name: CreateSetting :one
INSERT INTO setting (id, key, value)
VALUES ($1, $2, $3)
RETURNING id, key, value;

-- name: DeleteSetting :exec
DELETE FROM setting WHERE key = $1;
