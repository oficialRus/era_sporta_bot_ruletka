-- Era Sporta Bot (ruletka) - schema + seed dump
--
-- Это не "pg_dump" существующих данных пользователей/вращений,
-- а воспроизводимый SQL-дамп структуры БД + дефолтных призов из миграций проекта.
--
-- Применение:
--   PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -f db_dump_schema_and_seed.sql
--
-- (Опционально) если БД ещё не создана:
--   PGPASSWORD=change_me psql -U app -h localhost -d postgres -c "CREATE DATABASE era_sporta"
--
-- Дата генерации: 2026-02-16

BEGIN;

-- Migration 001: Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    username VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_user_id ON users(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Migration 002: Create prizes table
CREATE TABLE IF NOT EXISTS prizes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    value DECIMAL(10,2) NOT NULL DEFAULT 0,
    probability_weight INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default prizes if table is empty
INSERT INTO prizes (name, type, value, probability_weight)
SELECT * FROM (VALUES
    ('Скидка 10%', 'discount', 10, 30),
    ('Скидка 20%', 'discount', 20, 15),
    ('1 месяц бесплатно', 'free_month', 1, 5),
    ('Скидка 5%', 'discount', 5, 50)
) AS v(name, type, value, probability_weight)
WHERE NOT EXISTS (SELECT 1 FROM prizes LIMIT 1);

-- Migration 004: Add first-spin fixed prize (idempotent)
INSERT INTO prizes (name, type, value, probability_weight, is_active)
SELECT 'БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА', 'free_days', 7, 1, true
WHERE NOT EXISTS (
    SELECT 1
    FROM prizes
    WHERE LOWER(TRIM(name)) = LOWER(TRIM('БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА'))
);

-- Migration 003: Create spins table
CREATE TABLE IF NOT EXISTS spins (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    prize_id INT NOT NULL REFERENCES prizes(id),
    result_value DECIMAL(10,2) NOT NULL,
    ip_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_spins_user_id ON spins(user_id);
CREATE INDEX IF NOT EXISTS idx_spins_created_at ON spins(created_at);
CREATE INDEX IF NOT EXISTS idx_spins_user_created ON spins(user_id, created_at);

COMMIT;

