#!/bin/bash
# Запуск всех сервисов проекта Era Sporta Bot

set -e

cd "$(dirname "$0")"

# Проверка .env
if [ ! -f .env ]; then
    echo "❌ Файл .env не найден!"
    echo "Скопируйте .env.example в .env и заполните переменные"
    exit 1
fi

# Загрузка переменных окружения
source .env

echo "============================================"
echo "  🚀 Запуск Era Sporta Bot"
echo "============================================"
echo ""

# Проверка Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен!"
    exit 1
fi

# Проверка подключения к БД (не блокируем запуск)
echo "Проверка подключения к базе данных..."
if PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT 1" &> /dev/null; then
    echo "✓ База данных доступна"
else
    echo "⚠️  База данных недоступна (запуск продолжится; при ошибках выполните: go run ./cmd/initdb)"
fi
echo ""

# Создание директории для логов
mkdir -p logs

# Функция для запуска сервиса
start_service() {
    local name=$1
    local cmd=$2
    local port=$3
    
    echo "Запуск $name..."
    nohup $cmd > logs/$name.log 2>&1 &
    local pid=$!
    echo $pid > logs/$name.pid
    echo "  ✓ $name запущен (PID: $pid, лог: logs/$name.log)"
    
    if [ ! -z "$port" ]; then
        echo "     Порт: $port"
    fi
}

# Проверка, не запущены ли уже сервисы
if [ -f logs/api.pid ] && kill -0 $(cat logs/api.pid) 2>/dev/null; then
    echo "⚠️  API уже запущен (PID: $(cat logs/api.pid))"
    echo "   Для перезапуска выполните: ./stop.sh && ./start.sh"
    exit 1
fi

# Запуск сервисов
echo "Запуск сервисов..."
echo ""
start_service "api" "go run ./cmd/api" ":$API_PORT"
sleep 1
start_service "bot" "go run ./cmd/bot"
sleep 1
start_service "web" "go run ./cmd/serveweb"

echo ""
echo "============================================"
echo "  ✓ Все сервисы запущены!"
echo "============================================"
echo ""
echo "API:  http://localhost:$API_PORT"
echo "Bot:  @${BOT_TOKEN%%:*}"
echo "Web:  $WEBAPP_URL"
echo ""
echo "Логи:"
echo "  tail -f logs/api.log"
echo "  tail -f logs/bot.log"
echo "  tail -f logs/web.log"
echo ""
echo "Остановка: ./stop.sh"
echo ""
