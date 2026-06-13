-- +goose Up
CREATE TYPE auction_type AS ENUM (
    'English',
    'Dutch'
);

CREATE TYPE auction_status AS ENUM (
    'Scheduled',
    'Active',
    'Ended',
    'Cancelled',
    'NoReserve'
);

CREATE TABLE auctions (
    id                 UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id          UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_bidder_id  UUID           REFERENCES users(id),
    -- NULL at start, updates with each bid

    title              VARCHAR(255)   NOT NULL,
    description        TEXT           NOT NULL,
    type               auction_type   NOT NULL,
    status             auction_status NOT NULL DEFAULT 'Scheduled',

    -- prices in cents
    starting_price     INTEGER        NOT NULL,
    reserved_price     INTEGER        NOT NULL DEFAULT 0,
    current_price      INTEGER        NOT NULL,
    -- snapshot of highest bid
    -- bids table is source of truth

    -- Dutch auction only (NULL for English)
    drop_amount        INTEGER,
    drop_interval      INTEGER,
    -- seconds between each price drop
    last_drop_time     TIMESTAMPTZ,

    -- anti sniping
    extend_on_bid      BOOLEAN        NOT NULL DEFAULT TRUE,
    extend_minutes     INTEGER        NOT NULL DEFAULT 5,

    start_time         TIMESTAMPTZ    NOT NULL,
    end_time           TIMESTAMPTZ    NOT NULL,
    original_end_time  TIMESTAMPTZ    NOT NULL,
    -- track if end_time was extended

    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auctions_seller_id      ON auctions(seller_id);
CREATE INDEX idx_auctions_status         ON auctions(status);
CREATE INDEX idx_auctions_end_time       ON auctions(end_time)
    WHERE status = 'Active';
-- partial index — only indexes active auctions
-- scheduler query becomes very fast

-- +goose Down
DROP TABLE IF EXISTS auctions;
DROP TYPE IF EXISTS auction_status;
DROP TYPE IF EXISTS auction_type;