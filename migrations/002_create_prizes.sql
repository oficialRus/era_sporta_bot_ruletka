-- +goose Up
CREATE TABLE IF NOT EXISTS prizes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    value DECIMAL(10,2) NOT NULL DEFAULT 0,
    probability_weight INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO prizes (name, type, value, probability_weight) VALUES
    ('Скидка 10%', 'discount', 10, 30),
    ('Скидка 20%', 'discount', 20, 15),
    ('1 месяц бесплатно', 'free_month', 1, 5),
    ('Скидка 5%', 'discount', 5, 50);

-- +goose Down
DROP TABLE IF EXISTS prizes;
