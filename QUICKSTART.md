# ⚡ Быстрый старт

## Первый запуск

```bash
cd /root/era_sporta_bot_ruletka

# 1. Полная установка (один раз)
./setup.sh

# 2. Запуск всех сервисов
./start.sh

# 3. Проверка статуса
./status.sh
```

## Ежедневное использование

```bash
# Запуск
./start.sh

# Статус
./status.sh

# Логи
tail -f logs/bot.log      # Telegram бот
tail -f logs/api.log      # API сервер
tail -f logs/web.log      # Веб-сервер

# Остановка
./stop.sh
```

## Проверка работы

### 1. Бот Telegram
- Найти: @era_of_sports_apsheronsk_bot
- Отправить: `/start`
- Должен ответить и запросить номер телефона

### 2. API
```bash
curl http://localhost:8080/health
```

### 3. База данных
```bash
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT COUNT(*) FROM users"
```

## Полезные команды

```bash
# Просмотр всех пользователей
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT * FROM users"

# Просмотр всех вращений
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT * FROM spins ORDER BY created_at DESC LIMIT 10"

# Просмотр призов
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT * FROM prizes"

# Статистика
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "
SELECT 
    (SELECT COUNT(*) FROM users) as total_users,
    (SELECT COUNT(*) FROM spins) as total_spins,
    (SELECT COUNT(*) FROM prizes WHERE is_active = true) as active_prizes;
"
```

## Очистка данных (для тестирования)

```bash
# Очистить все данные пользователей и вращений
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -f scripts/reset_db.sql

# Перезапустить бота после очистки
./stop.sh && ./start.sh
```

## Проблемы?

1. **Бот не запускается** → Проверьте `BOT_TOKEN` в `.env`
2. **Ошибка БД** → Запустите `./setup.sh` заново
3. **Порт занят** → Остановите старые процессы: `./stop.sh`

## Важные файлы

- `.env` - Конфигурация (токены, пароли)
- `logs/` - Логи всех сервисов
- `bin/` - Собранные бинарники

## Структура проекта

```
era_sporta_bot_ruletka/
├── cmd/           # Точки входа
│   ├── api/       # HTTP API
│   ├── bot/       # Telegram бот
│   ├── serveweb/  # Веб-сервер
│   └── initdb/    # Инициализация БД
├── internal/      # Внутренняя логика
├── migrations/    # SQL миграции
├── logs/          # Логи (создается автоматически)
├── .env           # Конфигурация
└── *.sh           # Скрипты управления
```

## Ссылки

- 📖 Полная документация: `README.md`
- 🚀 Развертывание: `DEPLOYMENT.md`
- 🏗️ Архитектура: `docs/ARCHITECTURE.md`
- 📱 Канал: https://t.me/erasporta_apsheronsk
- 🤖 Бот: @era_of_sports_apsheronsk_bot
