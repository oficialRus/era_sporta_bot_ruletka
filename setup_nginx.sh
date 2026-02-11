#!/bin/bash
# Скрипт настройки Nginx для бота-рулетки
# Запуск: sudo ./setup_nginx.sh

set -e

echo "============================================"
echo "  🔧 Настройка Nginx для бота-рулетки"
echo "============================================"
echo ""

# Проверка прав root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Запустите скрипт с sudo:"
    echo "   sudo ./setup_nginx.sh"
    exit 1
fi

DOMAIN="bot-wheel.era-sporta-apsheronsk.ru"
CONFIG_FILE="/etc/nginx/sites-available/bot-wheel-era-sporta.conf"
ENABLED_LINK="/etc/nginx/sites-enabled/bot-wheel-era-sporta.conf"

echo "1. Проверка конфигурации..."
if [ ! -f "$CONFIG_FILE" ]; then
    echo "   ❌ Файл $CONFIG_FILE не найден!"
    exit 1
fi
echo "   ✓ Конфигурация найдена"
echo ""

echo "2. Активация конфигурации..."
if [ -L "$ENABLED_LINK" ]; then
    echo "   ⚠️  Конфигурация уже активирована"
else
    ln -sf "$CONFIG_FILE" "$ENABLED_LINK"
    echo "   ✓ Конфигурация активирована"
fi
echo ""

echo "3. Проверка синтаксиса Nginx..."
if nginx -t 2>&1 | grep -q "syntax is ok"; then
    echo "   ✓ Синтаксис корректен"
else
    echo "   ❌ Ошибка в конфигурации Nginx"
    nginx -t
    exit 1
fi
echo ""

echo "4. Перезагрузка Nginx..."
systemctl reload nginx
echo "   ✓ Nginx перезагружен"
echo ""

echo "5. Проверка статуса Nginx..."
if systemctl is-active --quiet nginx; then
    echo "   ✓ Nginx работает"
else
    echo "   ❌ Nginx не запущен"
    systemctl status nginx
    exit 1
fi
echo ""

echo "============================================"
echo "  ✓ Nginx настроен!"
echo "============================================"
echo ""
echo "Следующий шаг - SSL сертификат:"
echo ""
echo "  sudo certbot --nginx -d $DOMAIN"
echo ""
echo "После установки SSL проверьте:"
echo "  curl https://$DOMAIN/api/prizes"
echo ""
echo "Логи Nginx:"
echo "  sudo tail -f /var/log/nginx/bot-wheel-era-sporta-access.log"
echo "  sudo tail -f /var/log/nginx/bot-wheel-era-sporta-error.log"
echo ""
