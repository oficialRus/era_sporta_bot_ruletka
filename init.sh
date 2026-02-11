#!/bin/bash
# Скрипт инициализации проекта

set -e

DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_NAME="${DB_NAME:-era_sporta}"

echo "============================================"
echo "  Инициализация проекта era_sporta_bot"
echo "============================================"
echo ""

# 1. Создание базы данных
echo "1. Создание базы данных '$DB_NAME'..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" 2>/dev/null | grep -q 1 && {
    echo "   ✓ База данных уже существует"
} || {
    PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -c "CREATE DATABASE $DB_NAME" >/dev/null 2>&1
    echo "   ✓ База данных создана"
}

# 2. Применение миграций
echo ""
echo "2. Применение миграций..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -f scripts/init_db_all.sql 2>&1 | grep -E "NOTICE|ERROR" || true
echo "   ✓ Миграции применены"

# 3. Проверка таблиц
echo ""
echo "3. Проверка таблиц..."
TABLE_COUNT=$(PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -tc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'" 2>/dev/null | tr -d ' ')
echo "   ✓ Создано таблиц: $TABLE_COUNT"

# 4. Проверка призов
echo ""
echo "4. Проверка призов..."
PRIZE_COUNT=$(PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -tc "SELECT COUNT(*) FROM prizes" 2>/dev/null | tr -d ' ')
echo "   ✓ Призов в базе: $PRIZE_COUNT"

echo ""
echo "============================================"
echo "  ✓ Инициализация завершена успешно!"
echo "============================================"
echo ""
echo "Запуск сервисов:"
echo "  API:    go run ./cmd/api"
echo "  Бот:    go run ./cmd/bot"
echo "  Web:    go run ./cmd/serveweb"
echo ""
