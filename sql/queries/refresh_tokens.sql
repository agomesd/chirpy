-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;
-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;
-- name: RevokleRefreshToken :one
UPDATE refresh_tokens
SET updated_at = NOW(),
    revoked_at = NOW()
WHERE token = $1
RETURNING *;
