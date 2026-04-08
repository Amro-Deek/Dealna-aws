-- name: CreateStudent :exec
INSERT INTO student (
    user_id,
    student_id,
    major,
    academic_year
)
VALUES ($1, $2, $3, $4);

-- name: UpdateStudent :exec
UPDATE student
SET 
  major = COALESCE(sqlc.narg('major'), major),
  academic_year = COALESCE(sqlc.narg('academic_year'), academic_year)
WHERE user_id = $1;

-- name: GetStudentByUserID :one
SELECT *
FROM student
WHERE user_id = $1
LIMIT 1;
