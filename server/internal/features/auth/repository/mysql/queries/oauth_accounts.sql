-- name: GetOAuthAccount :one
SELECT id, user_id, provider, provider_user_id, created_at
FROM oauth_accounts
WHERE provider = ? AND provider_user_id = ?;

-- name: CreateOAuthAccount :exec
INSERT INTO oauth_accounts (user_id, provider, provider_user_id)
VALUES (?, ?, ?);
