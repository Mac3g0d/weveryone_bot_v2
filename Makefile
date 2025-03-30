# Команды
.PHONY: build run clean up down restart logs test lint env help

# Переменные
BINARY_NAME=bot
GO=go
DOCKER_COMPOSE=docker compose

# Сборка приложения локально
build:
	CGO_ENABLED=1 $(GO) build -o $(BINARY_NAME) .

# Запуск приложения локально
run: build
	./$(BINARY_NAME)

# Очистка сборки
clean:
	rm -f $(BINARY_NAME)
	rm -rf data/*.db

# Docker команды
# Собрать и запустить контейнер
up:
	$(DOCKER_COMPOSE) up --build bot

# Запустить контейнер в фоновом режиме
up-d:
	$(DOCKER_COMPOSE) up --build -d bot

# Остановить контейнеры
down:
	$(DOCKER_COMPOSE) down

# Перезапустить контейнеры
restart: down up-d

# Проверить логи
logs:
	$(DOCKER_COMPOSE) logs -f bot

# Тестирование
test:
	$(GO) test -v ./...

# Линтер
lint:
	$(GO) vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint не установлен"; \
	fi

# Создание .env файла
env:
	@if [ ! -f .env ]; then \
		echo "BOT_TOKEN='your_token_here'" > .env; \
		echo "ADMIN_ID='your_admin_id_here'" >> .env; \
		echo "Создан файл .env. Пожалуйста, заполните его правильными значениями."; \
	else \
		echo "Файл .env уже существует"; \
	fi

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build      - Собрать приложение локально"
	@echo "  make run        - Запустить приложение локально"
	@echo "  make clean      - Очистить сборку"
	@echo "  make up         - Собрать и запустить контейнер"
	@echo "  make up-d       - Собрать и запустить контейнер в фоновом режиме"
	@echo "  make down       - Остановить контейнеры"
	@echo "  make restart    - Перезапустить контейнеры" 
	@echo "  make logs       - Просмотр логов"
	@echo "  make test       - Запустить тесты"
	@echo "  make lint       - Запустить линтер"
	@echo "  make env        - Создать файл .env"
	@echo "  make help       - Показать эту справку"
