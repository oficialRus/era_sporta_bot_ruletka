#!/bin/bash
# Скрипт настройки базы данных

set -e

DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-postgres}"
DB_HOST="${DB_HOST:-localhost}"
DB_NAME="${DB_NAME:-era_sporta}"

echo "=== Настройка базы данных ==="
echo "DB: $DB_NAME на $DB_HOST"

# Создание базы данных если её нет
echo "Создание базы данных..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -c "CREATE DATABASE $DB_NAME"

echo "База данных создана или уже существует"

# Применение миграций
echo "Применение миграций..."
cd "$(dirname "$0")/.."

for migration in migrations/*.sql; do
    echo "Применяю: $migration"
    PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -f "$migration"
done

echo "=== Настройка завершена успешно! ==="
