-- name: CreateProviderPreRegistration :one
INSERT INTO provider_pre_registration (email, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProviderPreRegistrationByToken :one
SELECT * FROM provider_pre_registration
WHERE token = $1;

-- name: GetProviderPreRegistrationByEmail :one
SELECT * FROM provider_pre_registration
WHERE email = $1;

-- name: UpdateProviderPreRegistration :exec
UPDATE provider_pre_registration
SET token = $2,
    expires_at = $3,
    used_at = $4,
    resend_count = $5,
    resend_window_start = $6,
    verified_at = $7
WHERE id = $1;
