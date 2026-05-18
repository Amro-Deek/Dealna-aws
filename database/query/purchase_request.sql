-- name: CreatePurchaseRequest :one
INSERT INTO purchase_request (item_id, buyer_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetPurchaseRequestsByItem :many
SELECT 
  pr.*,
  p.display_name AS buyer_name,
  p.profile_picture_url AS buyer_pic
FROM purchase_request pr
INNER JOIN profile p ON p.user_id = pr.buyer_id
WHERE pr.item_id = $1
ORDER BY pr.created_at ASC;

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

-- name: GetPurchaseRequestsByBuyer :many
SELECT * FROM purchase_request
WHERE buyer_id = $1
ORDER BY updated_at DESC;
