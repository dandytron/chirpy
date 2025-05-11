-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, hashed_password, email)
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1;