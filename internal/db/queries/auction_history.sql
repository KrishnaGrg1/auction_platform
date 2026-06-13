-- name: CreateAuctionHistory :one
INSERT INTO auction_history (
    auction_id,
    user_id,
    event,
    amount,
    note
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAuctionHistory :many
SELECT * FROM auction_history
WHERE auction_id = $1
ORDER BY created_at ASC;
-- full timeline of everything that happened

-- name: GetAuctionHistoryByEvent :many
SELECT * FROM auction_history
WHERE auction_id = $1
  AND event      = $2
ORDER BY created_at ASC;

-- name: GetUserBidHistory :many
SELECT
    ah.*,
    a.title        AS auction_title,
    a.current_price AS auction_current_price
FROM auction_history ah
JOIN auctions a ON a.id = ah.auction_id
WHERE ah.user_id = $1
  AND ah.event   IN ('BidPlaced', 'BidIncreased', 'BidRefunded')
ORDER BY ah.created_at DESC;
-- user's complete bidding history across all auctions