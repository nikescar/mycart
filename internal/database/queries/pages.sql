-- name: ListPages :many
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE active = ?
ORDER BY name;

-- name: ListPagesByPosition :many
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE position = ? AND active = ?
ORDER BY name;

-- name: GetPageBySlug :one
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE slug = ?
LIMIT 1;

-- name: GetPageByID :one
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE id = ?
LIMIT 1;

-- name: CreatePage :exec
INSERT INTO page (id, name, slug, content, position, active, created)
VALUES (?, ?, ?, ?, ?, ?, datetime('now'));

-- name: UpdatePage :exec
UPDATE page
SET name = ?,
    content = ?,
    position = ?,
    active = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: DeletePage :exec
DELETE FROM page
WHERE id = ?;

-- name: CheckSlugExists :one
SELECT COUNT(*) as count
FROM page
WHERE slug = ? AND id != ?;
