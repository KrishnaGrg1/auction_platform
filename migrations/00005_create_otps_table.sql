-- +goose Up
CREATE TABLE otps (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       VARCHAR(6)   NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otps_user_id ON otps(user_id);
CREATE INDEX idx_otps_expires_at ON otps(expires_at);

-- +goose Down
DROP TABLE IF EXISTS otps;