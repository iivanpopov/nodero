-- name: CreateSession :exec
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES (?, ?, ?);

-- name: UpdateSession :exec
UPDATE sessions
SET expires_at = ?, last_used_at = UTC_TIMESTAMP()
WHERE id = ?;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, expires_at, last_used_at, created_at
FROM sessions
WHERE token_hash = ? AND expires_at > UTC_TIMESTAMP();

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= UTC_TIMESTAMP();
