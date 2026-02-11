#!/bin/bash
# Полная установка и настройка проекта

set -e

cd "$(dirname "$0")"

echo "============================================"
echo "  🔧 Установка Era Sporta Bot"
echo "============================================"
echo ""

# 1. Проверка Go
echo "1. Проверка Go..."
if ! command -v go &> /dev/null; then
    echo "   ❌ Go не установлен!"
    echo "   Установите Go: https://golang.org/dl/"
    exit 1
fi
echo "   ✓ Go $(go version | awk '{print $3}')"
echo ""

# 2. Проверка PostgreSQL
echo "2. Проверка PostgreSQL..."
if ! command -v psql &> /dev/null; then
    echo "   ⚠️  psql не найден"
fi

if ! PGPASSWORD=change_me psql -U app -h localhost -d postgres -c "SELECT 1" &> /dev/null; then
    echo "   ❌ PostgreSQL недоступен!"
    echo "   Убедитесь что PostgreSQL запущен и доступен"
    exit 1
fi
echo "   ✓ PostgreSQL доступен"
echo ""

# 3. Проверка .env
echo "3. Проверка конфигурации..."
if [ ! -f .env ]; then
    echo "   ⚠️  Файл .env не найден"
    echo "   Создание из .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "   ✓ Файл .env создан"
        echo "   ⚠️  Заполните переменные в .env перед запуском!"
    else
        echo "   ❌ .env.example не найден!"
        exit 1
    fi
else
    echo "   ✓ Файл .env существует"
fi
echo ""

# 4. Установка зависимостей
echo "4. Установка Go зависимостей..."
go mod download
echo "   ✓ Зависимости установлены"
echo ""

# 5. Инициализация БД
echo "5. Инициализация базы данных..."
echo ""
go run ./cmd/initdb
echo ""

# 6. Сборка проекта (опционально)
echo "6. Сборка проекта..."
mkdir -p bin
echo "   Сборка API..."
go build -o bin/api ./cmd/api
echo "   Сборка бота..."
go build -o bin/bot ./cmd/bot
echo "   Сборка веб-сервера..."
go build -o bin/serveweb ./cmd/serveweb
echo "   ✓ Все бинарники собраны в ./bin/"
echo ""

echo "============================================"
echo "  ✓ Установка завершена!"
echo "============================================"
echo ""
echo "Следующие шаги:"
echo ""
echo "1. Проверьте настройки в .env"
echo "2. Запустите сервисы:"
echo "   ./start.sh"
echo ""
echo "3. Проверьте статус:"
echo "   ./status.sh"
echo ""
echo "4. Просмотр логов:"
echo "   tail -f logs/bot.log"
echo ""
echo "5. Остановка:"
echo "   ./stop.sh"
echo ""
