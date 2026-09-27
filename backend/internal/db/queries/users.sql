-- name: UpsertUserByEmail :one
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;
