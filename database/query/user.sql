-- name: GetUserByID :one
SELECT
    user_id,
    role,
    email
FROM "User"
WHERE user_id = $1;



-- name: CreateUser :one
INSERT INTO "User" (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE "User"
SET last_login_at = NOW()
WHERE user_id = $1;

-- name: GetUserForLogin :one
SELECT
    user_id,
    email,
    password_hash,
    role
FROM "User"
WHERE email = $1;

-- name: GetUserByEmail :one
SELECT
    user_id,
    email,
    password_hash,
    role
FROM "User"
WHERE email = $1
LIMIT 1;

