-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
VALUES (gen_random_uuid(), $1, $2)
RETURNING *;
-- name: GetChirps :many
SELECT *
FROM chirps
ORDER BY created_at;
-- name: GetChirp :one
SELECT *
FROM chirps
WHERE id = $1;
-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;
