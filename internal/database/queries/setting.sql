-- name: GetSettingByKey :one
SELECT id, key, value
FROM setting
WHERE key = ?
LIMIT 1;

-- name: GetSettingByID :one
SELECT id, key, value
FROM setting
WHERE id = ?
LIMIT 1;

-- name: UpdateSettingByKey :exec
UPDATE setting
SET value = ?
WHERE key = ?;

-- name: UpdateSettingByID :exec
UPDATE setting
SET value = ?
WHERE id = ?;

-- name: ListAllSettings :many
SELECT id, key, value
FROM setting
ORDER BY key;

-- name: CreateSetting :exec
INSERT INTO setting (id, key, value)
VALUES (?, ?, ?);

-- name: DeleteSettingByKey :exec
DELETE FROM setting
WHERE key = ?;

-- name: BulkUpdateSettings :exec
UPDATE setting
SET value = CASE key
  WHEN ? THEN ?
  WHEN ? THEN ?
  WHEN ? THEN ?
  ELSE value
END
WHERE key IN (?, ?, ?);
