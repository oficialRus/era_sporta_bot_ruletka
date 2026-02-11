#!/bin/bash
# Остановка всех сервисов

cd "$(dirname "$0")"

echo "============================================"
echo "  🛑 Остановка Era Sporta Bot"
echo "============================================"
echo ""

# Функция остановки сервиса
stop_service() {
    local name=$1
    local pidfile="logs/$name.pid"
    
    if [ ! -f "$pidfile" ]; then
        echo "  ⊘ $name не запущен"
        return
    fi
    
    local pid=$(cat "$pidfile")
    
    if kill -0 $pid 2>/dev/null; then
        echo "Остановка $name (PID: $pid)..."
        kill $pid
        sleep 1
        
        # Принудительная остановка если не остановился
        if kill -0 $pid 2>/dev/null; then
            echo "  Принудительная остановка..."
            kill -9 $pid 2>/dev/null || true
        fi
        
        echo "  ✓ $name остановлен"
    else
        echo "  ⊘ $name уже остановлен"
    fi
    
    rm -f "$pidfile"
}

# Остановка всех сервисов
stop_service "api"
stop_service "bot"
stop_service "web"

echo ""
echo "✓ Все сервисы остановлены"
echo ""
