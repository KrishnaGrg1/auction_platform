-- +goose Up
CREATE TYPE auction_event AS ENUM (
    'Created',
    'Started',
    'BidPlaced',
    'BidRefunded',
    'BidIncreased',
    -- same bidder raised their bid
    'PriceDropped',
    -- Dutch auction price drop
    'Extended',
    -- end time extended due to late bid
    'Ended',
    'Cancelled',
    'WinnerDeclared',
    'SellerPaid',
    'BidderRefundedOnCancel'
    -- refund when auction cancelled
);

CREATE TABLE auction_history (
    id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id  UUID          NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id     UUID          REFERENCES users(id),
    -- NULL for system events like PriceDropped, Started
    event       auction_event NOT NULL,
    amount      INTEGER,
    -- NULL for non money events
    note        TEXT,
    -- human readable description of what happened
    -- "Bob placed bid of $80"
    -- "Alice refunded $60 — outbid by Bob"
    -- "Price dropped from $100 to $90"
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auction_history_auction_id ON auction_history(auction_id);
CREATE INDEX idx_auction_history_user_id    ON auction_history(user_id);
CREATE INDEX idx_auction_history_event      ON auction_history(auction_id, event);

-- +goose Down
DROP TABLE IF EXISTS auction_history;
DROP TYPE IF EXISTS auction_event;