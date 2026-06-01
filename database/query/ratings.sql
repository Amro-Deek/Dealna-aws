-- name: CreateRating :one
INSERT INTO rating (
    rater_id,
    rated_user_id,
    transaction_id,
    stars,
    comment,
    is_frozen
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetPendingRatings :many
-- Returns completed transactions where the user was the buyer and hasn't rated yet, and the transaction is >= 14 days old.
SELECT 
    t.transaction_id,
    i.item_id,
    i.title AS item_title,
    t.seller_id,
    p.display_name AS seller_name,
    EXTRACT(DAY FROM (CURRENT_TIMESTAMP - t.completed_at))::INTEGER AS days_since_completion
FROM transaction t
JOIN item i ON t.item_id = i.item_id
JOIN profile p ON t.seller_id = p.user_id
LEFT JOIN rating r ON r.transaction_id = t.transaction_id AND r.rater_id = t.buyer_id
WHERE t.buyer_id = $1
  AND t.transaction_status = 'COMPLETED'
  AND r.rating_id IS NULL
  AND t.completed_at <= CURRENT_TIMESTAMP - INTERVAL '14 days';

-- name: GetTransactionsToRemind :many
-- Transactions completed exactly X days ago where the buyer hasn't rated.
SELECT 
    t.transaction_id,
    t.buyer_id,
    t.seller_id,
    i.title AS item_title
FROM transaction t
JOIN item i ON t.item_id = i.item_id
LEFT JOIN rating r ON r.transaction_id = t.transaction_id AND r.rater_id = t.buyer_id
WHERE t.transaction_status = 'COMPLETED'
  AND r.rating_id IS NULL
  AND t.completed_at >= CURRENT_TIMESTAMP - ($1::int * INTERVAL '1 day') - INTERVAL '1 day'
  AND t.completed_at < CURRENT_TIMESTAMP - ($1::int * INTERVAL '1 day');

-- name: CountRatingsBetweenUsers :one
SELECT COUNT(*)
FROM transaction t
JOIN rating r ON r.transaction_id = t.transaction_id
WHERE t.transaction_status = 'COMPLETED'
  AND ((t.buyer_id = $1 AND t.seller_id = $2) OR (t.buyer_id = $2 AND t.seller_id = $1))
  AND t.completed_at >= CURRENT_TIMESTAMP - INTERVAL '30 days';

-- name: UpdateUserRating :one
UPDATE "User"
SET 
    total_ratings = $2,
    sum_ratings = $3,
    bayesian_rating = $4
WHERE user_id = $1
RETURNING *;

-- name: GetGlobalRatingAverage :one
SELECT 
    COALESCE(SUM(stars)::float / NULLIF(COUNT(stars), 0), 4.0)::float AS global_avg,
    COUNT(stars)::int AS total_count
FROM rating
WHERE is_frozen = false;

-- name: UpdateSysConfig :exec
INSERT INTO sys_config (key, value, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET 
    value = EXCLUDED.value,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetSysConfig :one
SELECT value FROM sys_config WHERE key = $1;
