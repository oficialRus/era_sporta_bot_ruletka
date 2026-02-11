.PHONY: help init-db reset-db run-api run-bot run-web run-all build clean

# Переменные для базы данных
DB_USER ?= postgres
DB_PASS ?= postgres
DB_HOST ?= localhost
DB_NAME ?= era_sporta

help:
	@echo "Доступные команды:"
	@echo "  make init-db     - Создать БД и применить миграции"
	@echo "  make reset-db    - Очистить данные в БД"
	@echo "  make run-api     - Запустить API сервер"
	@echo "  make run-bot     - Запустить Telegram бота"
	@echo "  make run-web     - Запустить веб-сервер"
	@echo "  make build       - Собрать все бинарники"
	@echo "  make clean       - Удалить собранные файлы"

# Создание базы данных
init-db:
	@echo "=== Создание базы данных ==="
	@PGPASSWORD=$(DB_PASS) psql -U $(DB_USER) -h $(DB_HOST) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
		PGPASSWORD=$(DB_PASS) psql -U $(DB_USER) -h $(DB_HOST) -c "CREATE DATABASE $(DB_NAME)"
	@echo "База данных создана или уже существует"
	@echo "=== Применение миграций ==="
	@PGPASSWORD=$(DB_PASS) psql -U $(DB_USER) -h $(DB_HOST) -d $(DB_NAME) -f scripts/init_db_all.sql
	@echo "=== База данных готова! ==="

# Сброс данных
reset-db:
	@echo "=== Очистка данных ==="
	@PGPASSWORD=$(DB_PASS) psql -U $(DB_USER) -h $(DB_HOST) -d $(DB_NAME) -f scripts/reset_db.sql
	@echo "=== Данные очищены ==="

# Запуск API
run-api:
	@echo "=== Запуск API на порту 8080 ==="
	go run ./cmd/api

# Запуск бота
run-bot:
	@echo "=== Запуск Telegram бота ==="
	go run ./cmd/bot

# Запуск веб-сервера
run-web:
	@echo "=== Запуск веб-сервера ==="
	go run ./cmd/serveweb

# Сборка всех бинарников
build:
	@echo "=== Сборка проекта ==="
	@mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/bot ./cmd/bot
	go build -o bin/serveweb ./cmd/serveweb
	@echo "=== Сборка завершена! Бинарники в ./bin/ ==="

# Очистка
clean:
	@echo "=== Удаление собранных файлов ==="
	rm -rf bin
	@echo "=== Очистка завершена ==="
