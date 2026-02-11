# 📝 Шпаргалка команд Era Sporta Bot

## 🚀 Запуск

```bash
cd /root/era_sporta_bot_ruletka

# Первый раз (установка)
./setup.sh

# Обычный запуск
./start.sh

# Проверка
./status.sh

# Остановка
./stop.sh
```

## 📊 Мониторинг

```bash
# Статус сервисов
./status.sh
ps aux | grep -E "api|bot|serveweb"

# Логи (живые)
tail -f logs/bot.log
tail -f logs/api.log
tail -f logs/web.log

# Последние 50 строк
tail -n 50 logs/bot.log

# Поиск ошибок
grep -i error logs/*.log
```

## 🗄️ База данных

```bash
# Подключение
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta

# Внутри psql:
\dt              # Список таблиц
\d users         # Структура таблицы
\q               # Выход
```

### Запросы

```bash
# Все пользователи
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT * FROM users"

# Последние вращения
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT * FROM spins ORDER BY created_at DESC LIMIT 10"

# Статистика
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "
SELECT 
    'Пользователей' as metric, COUNT(*)::text as value FROM users
UNION ALL
SELECT 'Вращений', COUNT(*)::text FROM spins
UNION ALL
SELECT 'Призов активных', COUNT(*)::text FROM prizes WHERE is_active = true;
"

# Топ призов
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "
SELECT p.name, COUNT(*) as wins
FROM spins s
JOIN prizes p ON s.prize_id = p.id
GROUP BY p.name
ORDER BY wins DESC;
"
```

## 🔧 Обслуживание

```bash
# Очистка данных (тестирование)
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -f scripts/reset_db.sql

# Пересоздание БД
PGPASSWORD=change_me psql -U app -h localhost -d postgres -c "DROP DATABASE era_sporta"
go run ./cmd/initdb

# Резервная копия
PGPASSWORD=change_me pg_dump -U app -h localhost era_sporta > backup_$(date +%Y%m%d).sql

# Восстановление
PGPASSWORD=change_me psql -U app -h localhost era_sporta < backup_20260209.sql
```

## 🏗️ Сборка

```bash
# Загрузка зависимостей
go mod download

# Сборка всех бинарников
mkdir -p bin
go build -o bin/api ./cmd/api
go build -o bin/bot ./cmd/bot
go build -o bin/serveweb ./cmd/serveweb

# Запуск бинарников
./bin/api &
./bin/bot &
./bin/serveweb &
```

## 🔍 Отладка

```bash
# Проверка портов
netstat -tulpn | grep -E "8080|5173"
lsof -i :8080
lsof -i :5173

# Проверка процессов
ps aux | grep go
kill <PID>          # Убить процесс
killall -9 api bot  # Убить все

# Тест API
curl http://localhost:8080/health
curl http://localhost:8080/api/prizes

# Тест БД
PGPASSWORD=change_me psql -U app -h localhost -d era_sporta -c "SELECT 1"
```

## 📝 Конфигурация

```bash
# Просмотр .env
cat .env

# Редактирование
nano .env
# или
vim .env

# Применение изменений (перезапуск)
./stop.sh && ./start.sh
```

## 🐛 Проблемы и решения

### Бот не отвечает
```bash
tail -f logs/bot.log
grep ERROR logs/bot.log
systemctl restart erasporta-bot  # если используется systemd
```

### Порт занят
```bash
lsof -i :8080
kill <PID>
./stop.sh
```

### База данных недоступна
```bash
systemctl status postgresql
PGPASSWORD=change_me psql -U app -h localhost -d postgres -c "SELECT 1"
```

### Ошибка в логах
```bash
tail -n 100 logs/bot.log
tail -n 100 logs/api.log
grep -A 5 -B 5 "error" logs/bot.log
```

## 📂 Важные файлы

```
/root/era_sporta_bot_ruletka/
├── .env                # Конфигурация (секреты)
├── logs/               # Логи сервисов
│   ├── api.log
│   ├── bot.log
│   └── web.log
├── bin/                # Бинарники
├── *.sh                # Скрипты управления
└── README.md           # Документация
```

## 🔐 Безопасность

```bash
# Права на .env
chmod 600 .env

# Проверка .gitignore
cat .gitignore | grep .env

# Просмотр без секретов
cat .env | grep -v TOKEN | grep -v PASSWORD
```

## 🌐 Доступ

- **Бот:** @era_of_sports_apsheronsk_bot
- **Канал:** https://t.me/erasporta_apsheronsk
- **Web App:** https://bot-wheel.era-sporta-apsheronsk.ru
- **API:** http://localhost:8080

## 📚 Документация

- `QUICKSTART.md` - Быстрый старт
- `README.md` - Полная документация
- `DEPLOYMENT.md` - Развертывание в продакшене
- `STATUS.md` - Текущий статус проекта
- `docs/ARCHITECTURE.md` - Архитектура

## 💡 Полезные алиасы

Добавьте в `~/.bashrc`:

```bash
alias era-start='cd /root/era_sporta_bot_ruletka && ./start.sh'
alias era-stop='cd /root/era_sporta_bot_ruletka && ./stop.sh'
alias era-status='cd /root/era_sporta_bot_ruletka && ./status.sh'
alias era-logs='tail -f /root/era_sporta_bot_ruletka/logs/*.log'
alias era-db='PGPASSWORD=change_me psql -U app -h localhost -d era_sporta'
```

После добавления: `source ~/.bashrc`
