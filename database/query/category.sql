-- name: ListCategories :many
SELECT category_id, name, description, created_at
FROM public.category
ORDER BY name ASC;
