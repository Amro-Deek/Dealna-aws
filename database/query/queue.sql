-- name: JoinQueue :one
INSERT INTO queue_entry (item_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetQueueByItemID :many
SELECT * FROM queue_entry
WHERE item_id = $1
ORDER BY joined_at ASC;

-- name: GetQueuePosition :one
SELECT count(*) + 1
FROM queue_entry qe1
WHERE qe1.item_id = $1 AND qe1.joined_at < (
    SELECT qe2.joined_at FROM queue_entry qe2 WHERE qe2.entry_id = $2
);

-- name: GetFrontOfQueue :one
SELECT * FROM queue_entry
WHERE item_id = $1 AND entry_status = 'WAITING'
ORDER BY joined_at ASC
LIMIT 1;

-- name: UpdateEntryStatus :exec
UPDATE queue_entry
SET entry_status = $2
WHERE entry_id = $1;

-- name: SetTurnStarted :exec
UPDATE queue_entry
SET turn_started_at = CURRENT_TIMESTAMP, entry_status = 'RESERVED'
WHERE entry_id = $1;

-- name: GetExpiredTurns :many
SELECT * FROM queue_entry
WHERE entry_status = 'RESERVED' AND turn_started_at < NOW() - INTERVAL '1 hour';

-- name: RemoveFromQueue :exec
DELETE FROM queue_entry
WHERE entry_id = $1;

-- name: GetQueueEntriesByUser :many
SELECT * FROM queue_entry
WHERE user_id = $1
ORDER BY joined_at DESC;

-- name: CountQueueEntries :one
SELECT count(*) FROM queue_entry
WHERE item_id = $1;

-- name: LeaveQueue :exec
DELETE FROM queue_entry
WHERE item_id = $1 AND user_id = $2;
