-- name: CreateStudent :exec
INSERT INTO student (
    user_id,
    student_id,
    major,
    academic_year
)
VALUES ($1, $2, $3, $4);
