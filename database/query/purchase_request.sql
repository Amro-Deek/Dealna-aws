-- name: CreatePurchaseRequest :one
INSERT INTO purchase_request (item_id, buyer_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetPurchaseRequestsByItem :many
SELECT * FROM purchase_request
WHERE item_id = $1
ORDER BY created_at ASC;

-- name: GetPurchaseRequestByID :one
SELECT * FROM purchase_request
WHERE request_id = $1;

-- name: UpdatePurchaseRequestStatus :exec
UPDATE purchase_request
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE request_id = $1;

-- name: FreezeOtherRequests :exec
UPDATE purchase_request
SET status = 'FROZEN', updated_at = CURRENT_TIMESTAMP
WHERE item_id = $1 AND request_id != $2 AND status = 'PENDING';

-- name: UnfreezeRequests :exec
UPDATE purchase_request
SET status = 'PENDING', updated_at = CURRENT_TIMESTAMP
WHERE item_id = $1 AND status = 'FROZEN';
