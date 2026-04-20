-- name: FollowUser :exec
INSERT INTO follow (follower_profile_id, following_profile_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UpdateFollowerCount :exec
UPDATE profile SET follower_count = follower_count + $2 WHERE profile_id = $1;

-- name: UpdateFollowingCount :exec
UPDATE profile SET following_count = following_count + $2 WHERE profile_id = $1;

-- name: UnfollowUser :exec
DELETE FROM follow
WHERE follower_profile_id = $1 AND following_profile_id = $2;

-- name: IsFollowing :one
SELECT COUNT(*) > 0 AS is_following
FROM follow
WHERE follower_profile_id = $1 AND following_profile_id = $2;

-- name: GetFollowers :many
SELECT f.follower_profile_id, f.following_profile_id, f.followed_at,
       p.display_name, p.profile_picture_url
FROM follow f
JOIN profile p ON p.profile_id = f.follower_profile_id
WHERE f.following_profile_id = $1
ORDER BY f.followed_at DESC;

-- name: GetFollowing :many
SELECT f.follower_profile_id, f.following_profile_id, f.followed_at,
       p.display_name, p.profile_picture_url
FROM follow f
JOIN profile p ON p.profile_id = f.following_profile_id
WHERE f.follower_profile_id = $1
ORDER BY f.followed_at DESC;

-- name: GetProfileIDByUserID :one
SELECT profile_id FROM profile WHERE user_id = $1 LIMIT 1;
