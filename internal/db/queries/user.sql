-- name: CreateUser :one
INSERT INTO users (
    first_name,
    last_name,
    email,
    password
)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: LockUserByID :one
SELECT * FROM users
WHERE id = $1
FOR UPDATE;
-- use inside transaction before touching balance

-- name: VerifyUser :one
UPDATE users
SET
    is_verified = TRUE,
    updated_at  = NOW()
WHERE email = $1
RETURNING *;

-- name: ChangePassword :one
UPDATE users
SET
    password   = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: HoldBidAmount :one
-- move money from available to held
-- called when user places a new bid
UPDATE users
SET
    available_balance = available_balance - $1,
    held_balance      = held_balance      + $1,
    updated_at        = NOW()
WHERE id                 = $2
  AND available_balance >= $1
-- safety check — cannot hold more than available
RETURNING *;

-- name: ReleaseBidAmount :one
-- move money from held back to available
-- called when user gets outbid
UPDATE users
SET
    held_balance      = held_balance      - $1,
    available_balance = available_balance + $1,
    updated_at        = NOW()
WHERE id             = $2
  AND held_balance  >= $1
RETURNING *;

-- name: IncreaseHeldByDifference :one
-- same bidder raising their own bid
-- only charge the extra amount
UPDATE users
SET
    available_balance = available_balance - $1,
    held_balance      = held_balance      + $1,
    updated_at        = NOW()
WHERE id                 = $2
  AND available_balance >= $1
-- $1 = new_amount - old_amount
RETURNING *;

-- name: TransferHeldToAvailable :one
-- winner's held money goes to seller available
-- two separate queries, both in same transaction
-- Step 1: remove from winner
UPDATE users
SET
    held_balance = held_balance - $1,
    updated_at   = NOW()
WHERE id            = $2
  AND held_balance >= $1
RETURNING *;

-- name: CreditAvailableBalance :one
-- Step 2: add to seller
UPDATE users
SET
    available_balance = available_balance + $1,
    updated_at        = NOW()
WHERE id = $2
RETURNING *;

-- name: Deposit :one
-- user deposits money into platform
UPDATE users
SET
    available_balance = available_balance + $1,
    updated_at        = NOW()
WHERE id = $2
RETURNING *;

-- name: Withdraw :one
-- user withdraws money from platform
UPDATE users
SET
    available_balance = available_balance - $1,
    updated_at        = NOW()
WHERE id                 = $2
  AND available_balance >= $1
RETURNING *;