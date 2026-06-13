-- +goose Up
CREATE TABLE winners (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id  UUID        UNIQUE NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    -- UNIQUE ensures only one winner per auction
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    final_price INTEGER     NOT NULL,
    paid_at     TIMESTAMPTZ,
    -- NULL until payment confirmed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_winners_auction_id ON winners(auction_id);
CREATE INDEX idx_winners_user_id    ON winners(user_id);

-- +goose Down
DROP TABLE IF EXISTS winners;