-- name: GetSession :one
SELECT key, value, expires
FROM session
WHERE key = ?
LIMIT 1;

-- name: SetSession :exec
INSERT INTO session (key, value, expires)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
  value = excluded.value,
  expires = excluded.expires;

-- name: DeleteSession :exec
DELETE FROM session
WHERE key = ?;

-- name: CleanupExpiredSessions :exec
DELETE FROM session
WHERE expires < ?;

-- name: ListActiveSessions :many
SELECT key, value, expires
FROM session
WHERE expires >= ?
ORDER BY expires DESC;
