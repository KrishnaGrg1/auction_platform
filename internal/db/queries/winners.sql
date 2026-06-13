-- name: CreateWinner :one
INSERT INTO winners (
    auction_id,
    user_id,
    final_price
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWinnerByAuctionID :one
SELECT * FROM winners
WHERE auction_id = $1;

-- name: GetWinsByUserID :many
SELECT * FROM winners
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: MarkWinnerPaid :exec
UPDATE winners
SET paid_at = NOW()
WHERE auction_id = $1;