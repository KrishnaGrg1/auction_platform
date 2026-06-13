-- name: CreateBid :one
INSERT INTO bids (
    auction_id,
    user_id,
    amount,
    is_auto_bid
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetHighestBidByAuctionID :one
SELECT * FROM bids
WHERE auction_id = $1
  AND status     = 'Active'
ORDER BY amount DESC
LIMIT 1;

-- name: GetBidsByAuctionID :many
SELECT * FROM bids
WHERE auction_id = $1
ORDER BY amount DESC;

-- name: GetBidsByUserID :many
SELECT * FROM bids
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetActiveBidByUserAndAuction :one
SELECT * FROM bids
WHERE auction_id = $1
  AND user_id    = $2
  AND status     = 'Active'
LIMIT 1;
-- check if this user already has active bid
-- used to detect same bidder increasing bid

-- name: MarkBidOutbid :exec
UPDATE bids
SET
    status     = 'Outbid',
    updated_at = NOW()
WHERE auction_id = $1
  AND user_id    = $2
  AND status     = 'Active';
-- mark previous bid as outbid
-- when new higher bid comes in

-- name: MarkBidWon :exec
UPDATE bids
SET
    status     = 'Won',
    updated_at = NOW()
WHERE id = $1;

-- name: MarkAllLosingBids :exec
UPDATE bids
SET
    status     = 'Lost',
    updated_at = NOW()
WHERE auction_id = $1
  AND status     = 'Outbid';
-- called when auction ends
-- all outbid bids marked Lost

-- name: MarkAllBidsRefunded :exec
UPDATE bids
SET
    status     = 'Refunded',
    updated_at = NOW()
WHERE auction_id = $1
  AND status     IN ('Active', 'Outbid');
-- called when auction cancelled
-- all bids refunded