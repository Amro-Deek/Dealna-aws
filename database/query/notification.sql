-- name: CreateNotification :one
INSERT INTO notification (user_id, type, payload)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetNotificationsForUser :many
SELECT * FROM notification
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :exec
UPDATE notification
SET is_read = true
WHERE notification_id = $1 AND user_id = $2;

-- name: CountUnreadNotifications :one
SELECT count(*) FROM notification
WHERE user_id = $1 AND is_read = false;
