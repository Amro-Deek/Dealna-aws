-- name: JoinQueue :one
INSERT INTO queue_entry (item_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: JoinQueueAtomic :one
INSERT INTO queue_entry (item_id, user_id)
SELECT $1, $2
WHERE
  -- Condition 1: Queue not full
  (SELECT count(*) FROM queue_entry
   WHERE item_id = $1
     AND entry_status IN ('WAITING','RESERVED','CONFIRMED','HANDED_OFF')
  ) < 10
  AND
  -- Condition 2: No cooldown (no cancel/expiry in last 2 hours for this user+item)
  NOT EXISTS (
    SELECT 1 FROM queue_entry
    WHERE item_id = $1 AND user_id = $2
      AND entry_status IN ('CANCELLED', 'EXPIRED')
      AND updated_at > NOW() - INTERVAL '2 hours'
  )
RETURNING *;

-- name: GetJoinEligibility :one
SELECT
  COUNT(*) FILTER (WHERE entry_status IN ('WAITING','RESERVED','CONFIRMED','HANDED_OFF')) AS active_count,
  COALESCE(BOOL_OR(user_id = $2 AND entry_status IN ('WAITING','RESERVED','CONFIRMED','HANDED_OFF')), false)::boolean AS already_in_queue,
  COALESCE(BOOL_OR(user_id = $2 AND entry_status IN ('CANCELLED', 'EXPIRED') AND updated_at > NOW() - INTERVAL '2 hours'), false)::boolean AS in_cooldown
FROM queue_entry
WHERE item_id = $1;

-- name: GetQueueByItemID :many
SELECT 
    qe.*,
    p.display_name AS buyer_name,
    p.profile_picture_url AS buyer_pic
FROM queue_entry qe
JOIN public.profile p ON qe.user_id = p.user_id
WHERE qe.item_id = $1
ORDER BY qe.joined_at ASC;

-- name: GetQueuePosition :one
SELECT count(*) + 1
FROM queue_entry qe1
WHERE qe1.item_id = $1 AND qe1.joined_at < (
    SELECT qe2.joined_at FROM queue_entry qe2 WHERE qe2.entry_id = $2
) AND qe1.entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- name: GetFrontOfQueue :one
SELECT * FROM queue_entry
WHERE item_id = $1 AND entry_status = 'WAITING'
ORDER BY joined_at ASC
LIMIT 1;

-- name: UpdateEntryStatus :exec
UPDATE queue_entry
SET entry_status = $2, updated_at = NOW()
WHERE entry_id = $1;

-- name: SetTurnStarted :exec
UPDATE queue_entry
SET turn_started_at = CURRENT_TIMESTAMP, entry_status = 'RESERVED', updated_at = NOW()
WHERE entry_id = $1;

-- name: GetExpiredTurns :many
SELECT * FROM queue_entry
WHERE entry_status = 'RESERVED' AND turn_started_at < NOW() - INTERVAL '1 hour';

-- name: ExpireReservedEntries :many
WITH locked AS (
  SELECT entry_id, item_id, user_id FROM queue_entry
  WHERE entry_status = 'RESERVED' AND updated_at < NOW() - INTERVAL '1 hour'
  FOR UPDATE SKIP LOCKED
)
UPDATE queue_entry SET entry_status = 'EXPIRED', updated_at = NOW()
FROM locked WHERE queue_entry.entry_id = locked.entry_id
RETURNING queue_entry.*;

-- name: ExpireConfirmedEntries :many
WITH locked AS (
  SELECT entry_id, item_id, user_id FROM queue_entry
  WHERE entry_status = 'CONFIRMED' AND updated_at < NOW() - INTERVAL '7 days'
  FOR UPDATE SKIP LOCKED
)
UPDATE queue_entry SET entry_status = 'EXPIRED', updated_at = NOW()
FROM locked WHERE queue_entry.entry_id = locked.entry_id
RETURNING queue_entry.*;

-- name: AutoCompleteHandedOffEntries :many
WITH locked AS (
  SELECT entry_id, item_id, user_id FROM queue_entry
  WHERE entry_status = 'HANDED_OFF' AND updated_at < NOW() - INTERVAL '24 hours'
  FOR UPDATE SKIP LOCKED
)
UPDATE queue_entry SET entry_status = 'COMPLETED', updated_at = NOW()
FROM locked WHERE queue_entry.entry_id = locked.entry_id
RETURNING queue_entry.*;

-- name: CancelAllQueueEntries :exec
UPDATE queue_entry SET entry_status = 'CANCELLED', updated_at = NOW()
WHERE item_id = $1 AND entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- name: GetActiveEntryByItemAndUser :one
SELECT * FROM queue_entry
WHERE item_id = $1 AND user_id = $2
  AND entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- name: RemoveFromQueue :exec
DELETE FROM queue_entry
WHERE entry_id = $1;

-- name: GetQueueEntriesByUser :many
SELECT * FROM queue_entry
WHERE user_id = $1
ORDER BY joined_at DESC;

-- name: CountQueueEntries :one
SELECT count(*) FROM queue_entry
WHERE item_id = $1 AND entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- name: LeaveQueue :exec
UPDATE queue_entry SET entry_status = 'CANCELLED'
WHERE item_id = $1 AND user_id = $2 AND entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- name: GetEntryByID :one
SELECT * FROM queue_entry
WHERE entry_id = $1;

-- name: GetItemOwner :one
SELECT owner_id::text FROM item
WHERE item_id = $1;
