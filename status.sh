#!/bin/bash
# Проверка статуса сервисов

cd "$(dirname "$0")"

echo "============================================"
echo "  📊 Статус Era Sporta Bot"
echo "============================================"
echo ""

# Функция проверки статуса
check_service() {
    local name=$1
    local pidfile="logs/$name.pid"
    local port=$2
    
    if [ ! -f "$pidfile" ]; then
        echo "  $name: ⊘ не запущен"
        return
    fi
    
    local pid=$(cat "$pidfile")
    
    if kill -0 $pid 2>/dev/null; then
        echo "  $name: ✓ работает (PID: $pid)"
        
        if [ ! -z "$port" ]; then
            if nc -z localhost $port 2>/dev/null || curl -s http://localhost:$port >/dev/null 2>&1; then
                echo "         Порт $port: ✓ доступен"
            else
                echo "         Порт $port: ⚠️  недоступен"
            fi
        fi
        
        # Показать последние строки лога
        if [ -f "logs/$name.log" ]; then
            local last_line=$(tail -n 1 "logs/$name.log")
            if [ ! -z "$last_line" ]; then
                echo "         Последняя запись: ${last_line:0:60}..."
            fi
        fi
    else
        echo "  $name: ❌ процесс не найден"
        rm -f "$pidfile"
    fi
}

# Загрузка переменных окружения
if [ -f .env ]; then
    source .env
fi

# Проверка сервисов
echo "Сервисы:"
check_service "api" ${API_PORT:-8080}
check_service "bot"
check_service "web"

echo ""
echo "База данных:"
if PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT COUNT(*) FROM users" >/dev/null 2>&1; then
    user_count=$(PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -tc "SELECT COUNT(*) FROM users" | tr -d ' ')
    spin_count=$(PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -tc "SELECT COUNT(*) FROM spins" | tr -d ' ')
    echo "  ✓ Подключение активно"
    echo "    Пользователей: $user_count"
    echo "    Вращений: $spin_count"
else
    echo "  ❌ Недоступна"
fi

echo ""
echo "Логи:"
echo "  tail -f logs/api.log"
echo "  tail -f logs/bot.log"
echo "  tail -f logs/web.log"
echo ""
