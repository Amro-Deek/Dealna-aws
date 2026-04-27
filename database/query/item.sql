-- name: InsertItem :one
INSERT INTO public.item (
    owner_id,
    category_id,
    title,
    description,
    price,
    pickup_location,
    item_status
) VALUES (
    $1, $2, $3, $4, $5, $6, 'AVAILABLE'
) RETURNING *;

-- name: GetItemDetails :one
SELECT 
    i.item_id, i.owner_id, i.category_id, i.title, i.description, i.price, i.pickup_location, i.item_status, i.created_at, i.updated_at,
    p.display_name AS owner_display_name,
    p.profile_picture_url AS owner_profile_picture_url,
    u.university_id,
    c.name AS category_name
FROM public.item i
JOIN public."User" u ON i.owner_id = u.user_id
JOIN public.profile p ON u.user_id = p.user_id
LEFT JOIN public.category c ON i.category_id = c.category_id
WHERE i.item_id = $1 AND i.deleted_at IS NULL;

-- name: GetFeedItems :many
SELECT 
    i.item_id, i.owner_id, i.category_id, i.title, i.description, i.price, i.pickup_location, i.item_status, i.created_at,
    p.display_name AS owner_display_name,
    p.profile_picture_url AS owner_profile_picture_url,
    c.name AS category_name,
    COALESCE((
        SELECT a.file_path 
        FROM public.attachment a 
        WHERE a.item_id = i.item_id 
        ORDER BY a.uploaded_at ASC 
        LIMIT 1
    ), '') AS thumbnail_url
FROM public.item i
JOIN public."User" u ON i.owner_id = u.user_id
JOIN public.profile p ON u.user_id = p.user_id
LEFT JOIN public.category c ON i.category_id = c.category_id
WHERE u.university_id = $1 
  AND i.item_status = 'AVAILABLE'
  AND i.deleted_at IS NULL
  AND ($2::uuid IS NULL OR i.category_id = $2)
  AND ($3::numeric IS NULL OR i.price >= $3)
  AND ($4::numeric IS NULL OR i.price <= $4)
  AND ($5::text IS NULL OR i.title ILIKE '%' || $5 || '%')
  AND i.owner_id != $8
ORDER BY i.created_at DESC
LIMIT $7 OFFSET $6;

-- name: GetDailyItemCount :one
SELECT COUNT(*)
FROM public.item
WHERE owner_id = $1 
  AND created_at >= NOW() - INTERVAL '1 day'
  AND deleted_at IS NULL;

-- name: UpdateItemStatus :exec
UPDATE public.item
SET item_status = $2, updated_at = CURRENT_TIMESTAMP
WHERE item_id = $1 AND deleted_at IS NULL;

-- name: DeleteItem :exec
UPDATE public.item
SET deleted_at = CURRENT_TIMESTAMP
WHERE item_id = $1;

-- name: InsertAttachment :one
INSERT INTO public.attachment (item_id, file_path)
VALUES ($1, $2) RETURNING *;

-- name: GetAttachmentsByItem :many
SELECT attachment_id, item_id, file_path, uploaded_at
FROM public.attachment
WHERE item_id = $1
ORDER BY uploaded_at ASC;

-- name: GetUserStorefront :many
SELECT 
    i.item_id, i.title, i.price, i.item_status, i.created_at,
    c.name AS category_name,
    COALESCE((
        SELECT a.file_path 
        FROM public.attachment a 
        WHERE a.item_id = i.item_id 
        ORDER BY a.uploaded_at ASC 
        LIMIT 1
    ), '') AS thumbnail_url
FROM public.item i
LEFT JOIN public.category c ON i.category_id = c.category_id
WHERE i.owner_id = $1 
  AND i.deleted_at IS NULL
ORDER BY i.created_at DESC
LIMIT $3 OFFSET $2;
