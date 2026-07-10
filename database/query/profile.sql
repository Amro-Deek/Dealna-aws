-- name: CreateProfile :exec
INSERT INTO profile (
    user_id,
    display_name
) VALUES (
    $1, $2
);

-- name: GetProfileByUserID :one
SELECT *
FROM profile
WHERE user_id = $1
LIMIT 1;

-- name: GetProfileByProfileID :one
SELECT *
FROM profile
WHERE profile_id = $1
LIMIT 1;

-- name: UpdateProfile :exec
UPDATE profile
SET 
  display_name = COALESCE(sqlc.narg('display_name'), display_name),
  bio = COALESCE(sqlc.narg('bio'), bio),
  profile_picture_url = COALESCE(sqlc.narg('profile_picture_url'), profile_picture_url),
  display_name_last_changed_at = COALESCE(sqlc.narg('display_name_last_changed_at'), display_name_last_changed_at)
WHERE user_id = $1;

-- name: UpdateDeviceToken :exec
UPDATE profile
SET device_token = $2
WHERE user_id = $1;

-- name: GetAdminUserProfileStats :one
SELECT 
    (SELECT COUNT(*) FROM public.reports WHERE reported_entity_id = $1 AND entity_type = 'USER')::int AS reports_received,
    (SELECT COUNT(*) FROM public.user_warnings WHERE target_user_id = $1)::int AS warnings_received,
    (SELECT COUNT(*) FROM public.item WHERE owner_id = $1)::int AS total_posts;