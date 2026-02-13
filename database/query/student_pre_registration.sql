-- name: CreateStudentPreRegistration :exec
INSERT INTO student_pre_registration (
    email,
    token,
    expires_at,
    resend_count,
    resend_window_start
)
VALUES ($1, $2, $3, 0, NOW());


-- name: GetStudentPreRegistrationByToken :one
SELECT *
FROM student_pre_registration
WHERE token = $1
LIMIT 1;


-- name: GetStudentPreRegistrationByEmail :one
SELECT *
FROM student_pre_registration
WHERE email = $1
LIMIT 1;


-- name: UpdateStudentPreRegistration :exec
UPDATE student_pre_registration
SET
    token = $2,
    expires_at = $3,
    used_at = $4,
    resend_count = $5,
    resend_window_start = $6,
    verified_at = $7
WHERE id = $1;
