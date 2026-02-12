-- +goose Up
INSERT INTO prizes (name, type, value, probability_weight, is_active)
SELECT 'БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА', 'free_days', 7, 1, true
WHERE NOT EXISTS (
    SELECT 1
    FROM prizes
    WHERE LOWER(TRIM(name)) = LOWER(TRIM('БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА'))
);

-- +goose Down
DELETE FROM prizes
WHERE LOWER(TRIM(name)) = LOWER(TRIM('БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА'));
