-- name: CreateUserWarning :one
INSERT INTO public.user_warnings (
  user_id,
  admin_id,
  reason
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserWarnings :many
SELECT * FROM public.user_warnings
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountUserWarnings :one
SELECT COUNT(*) FROM public.user_warnings
WHERE user_id = $1;
