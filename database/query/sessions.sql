-- =========================
-- Sessions (Hybrid JWT + JTI)
-- =========================

-- name: CreateSession :exec
INSERT INTO sessions (
    user_id,
    jti,
    expires_at
) VALUES (
    $1,
    $2,
    $3
);

-- name: GetSessionByJTI :one
SELECT
    session_id,
    user_id,
    jti,
    revoked,
    expires_at,
    created_at,
    revoked_at
FROM sessions
WHERE jti = $1;

-- name: RevokeSessionByJTI :exec
UPDATE sessions
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE jti = $1;

-- name: RevokeAllSessionsForUser :exec
UPDATE sessions
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE user_id = $1
  AND revoked = FALSE;

