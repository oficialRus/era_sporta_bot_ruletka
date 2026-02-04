-- +goose Up
CREATE TABLE IF NOT EXISTS spins (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    prize_id INT NOT NULL REFERENCES prizes(id),
    result_value DECIMAL(10,2) NOT NULL,
    ip_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_spins_user_id ON spins(user_id);
CREATE INDEX idx_spins_created_at ON spins(created_at);
CREATE INDEX idx_spins_user_created ON spins(user_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS spins;
