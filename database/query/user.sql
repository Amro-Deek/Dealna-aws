-- name: GetUserByEmail :one
SELECT
    user_id,
    email,
    role,
    keycloak_sub
FROM "User"
WHERE email = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByKeycloakSub :one
SELECT
    user_id,
    email,
    role,
    keycloak_sub
FROM "User"
WHERE keycloak_sub = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByID :one
SELECT
    user_id,
    email,
    role,
    keycloak_sub
FROM "User"
WHERE user_id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: CreateStudentUser :one
INSERT INTO "User" (
    email,
    role,
    university_id,
    keycloak_sub,
    account_status,
    email_verified
) VALUES (
    $1, 'VERIFIED_STUDENT', $2, $3, 'ACTIVE', true
) RETURNING user_id, email, role, keycloak_sub;