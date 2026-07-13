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
    keycloak_sub,
    total_ratings,
    sum_ratings,
    bayesian_rating
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

-- name: GetUserUniversityID :one
SELECT university_id
FROM public."User"
WHERE user_id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: CreateApplicantUser :one
INSERT INTO "User" (
    email,
    role,
    keycloak_sub,
    account_status,
    email_verified,
    university_id
) VALUES (
    $1, 'APPLICANT', $2, 'PENDING', false, (SELECT university_id FROM university WHERE domain = 'birzeit.edu' LIMIT 1)
) RETURNING user_id, email, role, keycloak_sub;

-- name: UpdateUserRole :exec
UPDATE "User"
SET role = $2
WHERE user_id = $1;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset (email, token, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (email) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at, created_at = CURRENT_TIMESTAMP;

-- name: GetPasswordResetToken :one
SELECT email, token, expires_at, created_at
FROM password_reset
WHERE email = $1 AND token = $2
LIMIT 1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset
WHERE email = $1;