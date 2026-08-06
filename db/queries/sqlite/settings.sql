-- name: GetSettingByKey :one
SELECT id, key, value
FROM setting
WHERE key = ? LIMIT 1;

-- name: UpdateSetting :exec
UPDATE setting
SET value = ?
WHERE key = ?;

-- name: ListSettings :many
SELECT id, key, value
FROM setting
ORDER BY key;

-- name: CreateSetting :one
INSERT INTO setting (id, key, value)
VALUES (?, ?, ?)
RETURNING id, key, value;

-- name: DeleteSetting :exec
DELETE FROM setting WHERE key = ?;
