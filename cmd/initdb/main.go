package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL не установлен")
	}

	ctx := context.Background()

	fmt.Println("============================================")
	fmt.Println("  Инициализация базы данных")
	fmt.Println("============================================")
	fmt.Println()

	// Сначала подключаемся к postgres для создания базы
	fmt.Println("Проверка и создание базы данных era_sporta...")
	postgresURL := "postgres://app:change_me@localhost:5432/postgres?sslmode=disable"
	connMaster, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к postgres: %v", err)
	}

	// Создаем базу если её нет
	var exists bool
	err = connMaster.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'era_sporta')").Scan(&exists)
	if err != nil {
		log.Fatalf("Не удалось проверить существование БД: %v", err)
	}

	if !exists {
		fmt.Println("  Создание базы данных era_sporta...")
		_, err = connMaster.Exec(ctx, "CREATE DATABASE era_sporta")
		if err != nil {
			log.Fatalf("Не удалось создать БД: %v", err)
		}
		fmt.Println("  ✓ База данных создана")
	} else {
		fmt.Println("  ✓ База данных уже существует")
	}
	connMaster.Close(ctx)

	// Подключение к созданной БД
	fmt.Println()
	fmt.Println("Подключение к базе данных era_sporta...")
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Println("✓ Подключение установлено")
	fmt.Println()

	// Применение миграций
	migrations := []string{
		`-- Migration 001: Create users table
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
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);`,

		`-- Migration 002: Create prizes table
CREATE TABLE IF NOT EXISTS prizes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    value DECIMAL(10,2) NOT NULL DEFAULT 0,
    probability_weight INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`,

		`-- Insert default prizes if table is empty
INSERT INTO prizes (name, type, value, probability_weight, is_active)
SELECT * FROM (VALUES
    ('БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА',                           'free_days',  7,  20, true),
    ('БЕСПЛАТНЫЕ 7 ДНЕЙ ФИТНЕСА',                           'free_days',  7,  20, true),
    ('ЗАРЯЖЕННЫЙ ФИТНЕС-ИНТЕНСИВ 🔥',                       'bonus',      1,  25, true),
    ('ШЕЙПИНГ — ГРУППОВАЯ ТРЕНИРОВКА ДЛЯ ФОРМЫ И РЕЛЬЕФА', 'bonus',      1,  25, true),
    ('БЕЗЛИМИТ ПОСЕЩЕНИЙ НА 1 МЕСЯЦ',                       'free_month', 1,   1, true),
    ('1 ДЕНЬ В ЭРА СПОРТА + МИНИ-ПРОГРАММА ТРЕНИРОВОК',     'free_days',  1,  25, true),
    ('СКИДКА НА ГОДОВОЙ АБОНЕМЕНТ',                         'discount',   1,  15, true),
    ('10% НА МАССАЖ / ВОССТАНОВЛЕНИЕ',                      'discount',   10, 25, true)
) AS v(name, type, value, probability_weight, is_active)
WHERE NOT EXISTS (SELECT 1 FROM prizes LIMIT 1);`,

		`-- Migration 003: Create spins table
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
CREATE INDEX IF NOT EXISTS idx_spins_user_created ON spins(user_id, created_at);`,
	}

	fmt.Println("Применение миграций...")
	for i, migration := range migrations {
		fmt.Printf("  Миграция %d/%d...\n", i+1, len(migrations))
		if _, err := conn.Exec(ctx, migration); err != nil {
			log.Fatalf("Ошибка при выполнении миграции %d: %v", i+1, err)
		}
	}

	fmt.Println("✓ Все миграции применены")
	fmt.Println()

	// Проверка таблиц
	var tableCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err != nil {
		log.Printf("Не удалось проверить таблицы: %v", err)
	} else {
		fmt.Printf("✓ Создано таблиц: %d\n", tableCount)
	}

	// Проверка призов
	var prizeCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM prizes").Scan(&prizeCount)
	if err != nil {
		log.Printf("Не удалось проверить призы: %v", err)
	} else {
		fmt.Printf("✓ Призов в базе: %d\n", prizeCount)
	}

	fmt.Println()
	fmt.Println("============================================")
	fmt.Println("  ✓ Инициализация завершена успешно!")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Println("Запуск сервисов:")
	fmt.Println("  API:    go run ./cmd/api")
	fmt.Println("  Бот:    go run ./cmd/bot")
	fmt.Println("  Web:    go run ./cmd/serveweb")
	fmt.Println()
}
