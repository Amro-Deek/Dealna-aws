-- name: CreateTransaction :one
INSERT INTO transaction (item_id, buyer_id, seller_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transaction
WHERE transaction_id = $1;

-- name: GetTransactionByItem :one
SELECT * FROM transaction
WHERE item_id = $1;

-- name: ConfirmSeller :exec
UPDATE transaction
SET seller_confirmed = true, updated_at = CURRENT_TIMESTAMP
WHERE transaction_id = $1;

-- name: ConfirmBuyer :exec
UPDATE transaction
SET buyer_confirmed = true, updated_at = CURRENT_TIMESTAMP
WHERE transaction_id = $1;

-- name: CompleteTransaction :exec
UPDATE transaction
SET status = 'COMPLETED', updated_at = CURRENT_TIMESTAMP
WHERE transaction_id = $1;

-- name: CancelTransaction :exec
UPDATE transaction
SET status = 'CANCELLED', updated_at = CURRENT_TIMESTAMP
WHERE transaction_id = $1;
