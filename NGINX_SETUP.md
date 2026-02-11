# 🔧 Настройка Nginx для бота-рулетки

## Текущая ситуация

**Статус:**
- ✅ Бот запущен
- ✅ API работает на порту 8080
- ✅ Web сервер работает на порту 5173
- ❌ Nginx конфигурация для bot-wheel.era-sporta-apsheronsk.ru не активирована

## Что нужно сделать

### 1. Активировать Nginx конфигурацию

Конфигурация уже создана в `/etc/nginx/sites-available/bot-wheel-era-sporta.conf`

Выполните команды:

```bash
# Создать символическую ссылку
sudo ln -sf /etc/nginx/sites-available/bot-wheel-era-sporta.conf /etc/nginx/sites-enabled/

# Проверить конфигурацию
sudo nginx -t

# Перезагрузить Nginx
sudo systemctl reload nginx
```

### 2. Настроить SSL (Let's Encrypt)

```bash
# Установить certbot если нет
sudo apt install certbot python3-certbot-nginx

# Получить SSL сертификат
sudo certbot --nginx -d bot-wheel.era-sporta-apsheronsk.ru
```

Certbot автоматически:
- Получит сертификат
- Обновит конфигурацию Nginx
- Настроит автообновление

### 3. Проверка работы

После настройки проверьте:

```bash
# Проверить API
curl https://bot-wheel.era-sporta-apsheronsk.ru/api/prizes

# Проверить веб-приложение
curl https://bot-wheel.era-sporta-apsheronsk.ru/
```

## Конфигурация Nginx

Файл `/etc/nginx/sites-available/bot-wheel-era-sporta.conf`:

```nginx
server {
    listen 80;
    server_name bot-wheel.era-sporta-apsheronsk.ru;

    # API проксирование на localhost:8080
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS' always;
        add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,X-Telegram-Init-Data' always;
    }

    # Web App (Mini App) на localhost:5173
    location / {
        proxy_pass http://localhost:5173;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Как это работает

```
Telegram Bot
    ↓
Кнопка "Открыть приложение"
    ↓
https://bot-wheel.era-sporta-apsheronsk.ru
    ↓
Nginx (порт 443/80)
    ├── / → localhost:5173 (Web App)
    └── /api/ → localhost:8080 (API)
```

## Текущие запущенные сервисы

```bash
# Проверить статус
ps aux | grep -E "api|serveweb|bot" | grep go

# Порты
# 8080 - API (Go)
# 5173 - Web сервер (Go)
```

## Команды для управления

```bash
# Статус Nginx
sudo systemctl status nginx

# Перезагрузка Nginx
sudo systemctl reload nginx

# Логи Nginx
sudo tail -f /var/log/nginx/bot-wheel-era-sporta-access.log
sudo tail -f /var/log/nginx/bot-wheel-era-sporta-error.log

# Логи приложения
tail -f /root/era_sporta_bot_ruletka/logs/api.log
tail -f /root/era_sporta_bot_ruletka/logs/web.log
tail -f /root/era_sporta_bot_ruletka/logs/bot.log
```

## После настройки

1. Откройте бота в Telegram: @era_of_sports_apsheronsk_bot
2. Отправьте `/start`
3. Поделитесь номером телефона
4. Нажмите "Открыть приложение"
5. Должна открыться рулетка и работать!

## Проблемы и решения

### Ошибка 502 Bad Gateway
```bash
# Проверить, запущены ли сервисы
ps aux | grep -E "api|serveweb"

# Перезапустить если нужно
cd /root/era_sporta_bot_ruletka
./stop.sh
./start.sh
```

### API не отвечает
```bash
# Проверить логи
tail -f logs/api.log

# Проверить порт
netstat -tulpn | grep 8080
```

### Веб не открывается
```bash
# Проверить логи
tail -f logs/web.log

# Проверить порт
netstat -tulpn | grep 5173
```

## DNS настройки

Убедитесь, что `bot-wheel.era-sporta-apsheronsk.ru` указывает на IP сервера:

```bash
# Проверить DNS
dig bot-wheel.era-sporta-apsheronsk.ru
nslookup bot-wheel.era-sporta-apsheronsk.ru

# Текущий IP сервера
curl -s ifconfig.me
```

---

**После выполнения этих шагов приложение будет полностью работать!**
