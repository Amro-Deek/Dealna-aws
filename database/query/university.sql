-- name: GetUniversityByDomain :one
SELECT university_id, name, domain, status
FROM university
WHERE domain = $1
LIMIT 1;
