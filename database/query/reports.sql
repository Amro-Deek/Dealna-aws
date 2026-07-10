-- name: CreateReport :one
INSERT INTO public.reports (
  reporter_id,
  reported_entity_id,
  entity_type,
  type,
  description,
  attachment_url
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetReport :one
SELECT * FROM public.reports
WHERE id = $1 LIMIT 1;

-- name: ListReports :many
SELECT 
  r.*,
  reporter.display_name AS reporter_name,
  CASE 
    WHEN r.entity_type = 'USER' THEN reported_user.display_name
    WHEN r.entity_type = 'ITEM' THEN reported_item.title
    ELSE NULL
  END::TEXT AS reported_entity_name
FROM public.reports r
JOIN public.profile reporter ON r.reporter_id = reporter.user_id
LEFT JOIN public.profile reported_user ON r.reported_entity_id = reported_user.user_id AND r.entity_type = 'USER'
LEFT JOIN public.item reported_item ON r.reported_entity_id = reported_item.item_id AND r.entity_type = 'ITEM'
ORDER BY r.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateReportStatus :one
UPDATE public.reports
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
