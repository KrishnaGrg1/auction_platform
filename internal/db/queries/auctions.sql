-- name: CreateAuction :one
INSERT INTO auctions (
    seller_id,
    title,
    description,
    type,
    starting_price,
    reserved_price,
    current_price,
    drop_amount,
    drop_interval,
    extend_on_bid,
    extend_minutes,
    start_time,
    end_time,
    original_end_time
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
-- original_end_time = end_time at creation
RETURNING *;

-- name: GetAuctionByID :one
SELECT * FROM auctions
WHERE id = $1;

-- name: LockAuctionByID :one
SELECT * FROM auctions
WHERE id = $1
FOR UPDATE;
-- always use this inside bid transaction
-- never use GetAuctionByID inside transaction

-- name: GetActiveAuctions :many
SELECT * FROM auctions
WHERE status = 'Active'
ORDER BY end_time ASC;

-- name: GetAuctionsBySellerID :many
SELECT * FROM auctions
WHERE seller_id = $1
ORDER BY created_at DESC;

-- name: GetScheduledAuctionsReadyToStart :many
SELECT * FROM auctions
WHERE status     = 'Scheduled'
  AND start_time <= NOW();
-- scheduler uses this to activate auctions

-- name: GetExpiredActiveAuctions :many
SELECT * FROM auctions
WHERE status   = 'Active'
  AND end_time < NOW();
-- scheduler uses this to end auctions

-- name: GetActiveDutchAuctions :many
SELECT * FROM auctions
WHERE status = 'Active'
  AND type   = 'Dutch';
-- scheduler uses this to drop prices

-- name: UpdateAuctionAfterBid :one
UPDATE auctions
SET
    current_price     = $1,
    current_bidder_id = $2,
    updated_at        = NOW()
WHERE id = $3
RETURNING *;

-- name: ExtendAuctionEndTime :one
UPDATE auctions
SET
    end_time   = end_time + ($1 * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $2
RETURNING *;
-- $1 = extend_minutes from auction row

-- name: DropDutchPrice :one
UPDATE auctions
SET
    current_price  = current_price - drop_amount,
    last_drop_time = NOW(),
    updated_at     = NOW()
WHERE id           = $1
  AND current_price > drop_amount
-- safety — never drop below zero
RETURNING *;

-- name: UpdateAuctionStatus :one
UPDATE auctions
SET
    status     = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: CancelAuction :one
UPDATE auctions
SET
    status     = 'Cancelled',
    updated_at = NOW()
WHERE id        = $1
  AND seller_id = $2
-- only seller can cancel their own auction
  AND status    = 'Scheduled'
-- can only cancel before it starts
RETURNING *;

-- name: GetAuctionsList :many
SELECT *
FROM auctions
WHERE status = $1
  AND type = $2
ORDER BY end_time DESC
LIMIT $3 OFFSET $4;