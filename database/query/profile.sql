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

-- name: UpdateProfile :exec
UPDATE profile
SET 
  display_name = COALESCE(sqlc.narg('display_name'), display_name),
  bio = COALESCE(sqlc.narg('bio'), bio),
  profile_picture_url = COALESCE(sqlc.narg('profile_picture_url'), profile_picture_url),
  display_name_last_changed_at = COALESCE(sqlc.narg('display_name_last_changed_at'), display_name_last_changed_at)
WHERE user_id = $1;