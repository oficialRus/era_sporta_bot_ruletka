# 🚀 Руководство по развертыванию

## Быстрый старт на сервере

### 1. Клонирование и настройка

```bash
cd /root/era_sporta_bot_ruletka

# Проверка файлов
ls -la
```

### 2. Полная установка

```bash
# Запустить автоматическую установку
./setup.sh
```

Скрипт setup.sh выполнит:
- ✓ Проверку Go и PostgreSQL
- ✓ Установку зависимостей
- ✓ Создание базы данных
- ✓ Применение миграций
- ✓ Сборку всех бинарников

### 3. Запуск сервисов

```bash
# Запуск всех сервисов в фоне
./start.sh

# Проверка статуса
./status.sh

# Просмотр логов
tail -f logs/bot.log
tail -f logs/api.log
tail -f logs/web.log

# Остановка всех сервисов
./stop.sh
```

## Ручной запуск

### Вариант 1: Go run (для разработки)

```bash
# Терминал 1
go run ./cmd/api

# Терминал 2
go run ./cmd/bot

# Терминал 3
go run ./cmd/serveweb
```

### Вариант 2: Бинарники (для продакшена)

```bash
# Сборка
go build -o bin/api ./cmd/api
go build -o bin/bot ./cmd/bot
go build -o bin/serveweb ./cmd/serveweb

# Запуск
./bin/api &
./bin/bot &
./bin/serveweb &
```

## Systemd сервисы (рекомендуется для продакшена)

### Создание сервисов

**1. API сервис** `/etc/systemd/system/erasporta-api.service`:

```ini
[Unit]
Description=Era Sporta API Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/era_sporta_bot_ruletka
ExecStart=/root/era_sporta_bot_ruletka/bin/api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**2. Bot сервис** `/etc/systemd/system/erasporta-bot.service`:

```ini
[Unit]
Description=Era Sporta Telegram Bot
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/era_sporta_bot_ruletka
ExecStart=/root/era_sporta_bot_ruletka/bin/bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**3. Web сервис** `/etc/systemd/system/erasporta-web.service`:

```ini
[Unit]
Description=Era Sporta Web Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/era_sporta_bot_ruletka
ExecStart=/root/era_sporta_bot_ruletka/bin/serveweb
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Управление сервисами

```bash
# Перезагрузка systemd
systemctl daemon-reload

# Включение автозапуска
systemctl enable erasporta-api
systemctl enable erasporta-bot
systemctl enable erasporta-web

# Запуск
systemctl start erasporta-api
systemctl start erasporta-bot
systemctl start erasporta-web

# Статус
systemctl status erasporta-api
systemctl status erasporta-bot
systemctl status erasporta-web

# Остановка
systemctl stop erasporta-api
systemctl stop erasporta-bot
systemctl stop erasporta-web

# Перезапуск
systemctl restart erasporta-api
systemctl restart erasporta-bot
systemctl restart erasporta-web

# Логи
journalctl -u erasporta-api -f
journalctl -u erasporta-bot -f
journalctl -u erasporta-web -f
```

## Nginx конфигурация

Для проксирования API и веб-приложения:

```nginx
# /etc/nginx/sites-available/erasporta-bot

server {
    listen 80;
    server_name bot-wheel.era-sporta-apsheronsk.ru;

    # API
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
    }

    # Web App
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

Включение конфигурации:

```bash
ln -s /etc/nginx/sites-available/erasporta-bot /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## SSL сертификат (Let's Encrypt)

```bash
apt install certbot python3-certbot-nginx
certbot --nginx -d bot-wheel.era-sporta-apsheronsk.ru
```

## Мониторинг и обслуживание

### Проверка работы

```bash
# Статус через скрипт
./status.sh

# Проверка процессов
ps aux | grep -E "api|bot|serveweb"

# Проверка портов
netstat -tulpn | grep -E "8080|5173"

# Проверка логов
tail -f logs/*.log
```

### Резервное копирование БД

```bash
# Создание бэкапа
PGPASSWORD=change_me pg_dump -U app -h localhost era_sporta > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановление из бэкапа
PGPASSWORD=change_me psql -U app -h localhost era_sporta < backup_20260209_230000.sql
```

### Очистка данных

```bash
# Очистить только данные (таблицы останутся)
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -f scripts/reset_db.sql

# Полное пересоздание БД
PGPASSWORD=change_me psql -U app -h localhost -d postgres -c "DROP DATABASE IF EXISTS era_sporta"
go run ./cmd/initdb
```

## Обновление кода

```bash
# Остановка сервисов
./stop.sh
# или
systemctl stop erasporta-api erasporta-bot erasporta-web

# Получение изменений
git pull

# Установка зависимостей
go mod download

# Применение миграций (если есть новые)
# Добавьте новые миграции в migrations/ и запустите:
# PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -f migrations/004_new_migration.sql

# Пересборка
go build -o bin/api ./cmd/api
go build -o bin/bot ./cmd/bot
go build -o bin/serveweb ./cmd/serveweb

# Запуск
./start.sh
# или
systemctl start erasporta-api erasporta-bot erasporta-web
```

## Проверка работоспособности

### 1. API

```bash
curl http://localhost:8080/health
# или
curl https://bot-wheel.era-sporta-apsheronsk.ru/api/health
```

### 2. База данных

```bash
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "
SELECT 
    (SELECT COUNT(*) FROM users) as users,
    (SELECT COUNT(*) FROM spins) as spins,
    (SELECT COUNT(*) FROM prizes) as prizes;
"
```

### 3. Бот

Отправьте `/start` боту в Telegram: @era_of_sports_apsheronsk_bot

## Устранение неполадок

### Бот не отвечает

```bash
# Проверить логи
tail -f logs/bot.log
journalctl -u erasporta-bot -n 50

# Проверить токен
grep BOT_TOKEN .env

# Перезапуск
systemctl restart erasporta-bot
```

### API недоступен

```bash
# Проверить порт
netstat -tulpn | grep 8080

# Проверить логи
tail -f logs/api.log

# Проверить Nginx
nginx -t
systemctl status nginx
```

### Ошибки БД

```bash
# Проверить подключение
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT 1"

# Проверить таблицы
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "\dt"

# Пересоздать БД
go run ./cmd/initdb
```

## Контакты

- Канал: https://t.me/erasporta_apsheronsk
- Бот: @era_of_sports_apsheronsk_bot
