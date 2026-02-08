-- Очистка всех данных БД (таблицы остаются)
TRUNCATE TABLE spins, users, prizes RESTART IDENTITY CASCADE;

-- Восстановление призов по умолчанию
INSERT INTO prizes (name, type, value, probability_weight) VALUES
    ('Скидка 10%', 'discount', 10, 30),
    ('Скидка 20%', 'discount', 20, 15),
    ('1 месяц бесплатно', 'free_month', 1, 5),
    ('Скидка 5%', 'discount', 5, 50);
