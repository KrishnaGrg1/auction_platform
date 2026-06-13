-- +goose Up
CREATE TYPE bid_status AS ENUM (
    'Active',
    -- currently highest bid
    'Outbid',
    -- someone bid higher
    'Won',
    -- this bid won the auction
    'Lost',
    -- auction ended, did not win
    'Refunded'
    -- bid was refunded on cancellation
);

CREATE TABLE bids (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id  UUID        NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    amount      INTEGER     NOT NULL,
    status      bid_status  NOT NULL DEFAULT 'Active',
    is_auto_bid BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_auction_id          ON bids(auction_id);
CREATE INDEX idx_bids_user_id             ON bids(user_id);
CREATE INDEX idx_bids_auction_amount      ON bids(auction_id, amount DESC);
CREATE INDEX idx_bids_auction_active      ON bids(auction_id)
    WHERE status = 'Active';
-- partial index — only active bids
-- "who is winning right now?" is very fast

-- +goose Down
DROP TABLE IF EXISTS bids;
DROP TYPE IF EXISTS bid_status;