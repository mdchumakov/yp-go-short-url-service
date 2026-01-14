.PHONY: swagger build run test clean fmt fmt-check imports imports-check

# Генерация Swagger документации
swagger:
	@echo "Генерация Swagger документации..."
	@if [ -f "$(shell go env GOPATH)/bin/swag" ]; then \
		$(shell go env GOPATH)/bin/swag init -g cmd/shortener/main.go; \
	else \
		echo "Swag не найден. Устанавливаем..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		$(shell go env GOPATH)/bin/swag init -g cmd/shortener/main.go; \
	fi
	@echo "Swagger документация сгенерирована успешно!"

# Сборка проекта
build: swagger
	go build ./cmd/shortener


# Сборка с линковщиком
build-linked: swagger
	@VERSION=$${VERSION:-dev}; \
	BUILD_DATE=$${BUILD_DATE:-$$(date -u +"%Y-%m-%dT%H:%M:%SZ")}; \
	BUILD_COMMIT=$${BUILD_COMMIT:-$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}; \
	go build -ldflags "-X main.buildVersion=$$VERSION -X main.buildDate=$$BUILD_DATE -X main.buildCommit=$$BUILD_COMMIT" ./cmd/shortener


# Запуск с линковщиком
run-linked: build-linked
	./shortener

# Запуск проекта
run: build
	./shortener

# Запуск тестов
test:
	go test ./...

# Очистка
clean:
	rm -f shortener
	rm -rf docs/

# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Форматирование кода с помощью gofmt
fmt:
	@echo "Форматирование кода..."
	gofmt -s -w .
	@echo "Форматирование завершено!"

# Проверка форматирования (без изменений файлов)
fmt-check:
	@echo "Проверка форматирования кода..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "Ошибка: код не отформатирован. Запустите 'make fmt' для исправления."; \
		gofmt -s -d .; \
		exit 1; \
	else \
		echo "Код отформатирован правильно!"; \
	fi

# Форматирование и сортировка импортов с помощью goimports
imports:
	@echo "Форматирование и сортировка импортов..."
	@if [ -f "$(shell go env GOPATH)/bin/goimports" ]; then \
		$(shell go env GOPATH)/bin/goimports -w .; \
	else \
		echo "goimports не найден. Устанавливаем..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		$(shell go env GOPATH)/bin/goimports -w .; \
	fi
	@echo "Импорты отсортированы и код отформатирован!"

# Проверка форматирования и импортов (без изменений файлов)
imports-check:
	@echo "Проверка форматирования и импортов..."
	@if [ -f "$(shell go env GOPATH)/bin/goimports" ]; then \
		GOIMPORTS="$(shell go env GOPATH)/bin/goimports"; \
	else \
		echo "goimports не найден. Устанавливаем..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		GOIMPORTS="$(shell go env GOPATH)/bin/goimports"; \
	fi; \
	if [ -n "$$($$GOIMPORTS -l .)" ]; then \
		echo "Ошибка: код не отформатирован или импорты не отсортированы."; \
		echo "Запустите 'make imports' для исправления."; \
		$$GOIMPORTS -d .; \
		exit 1; \
	else \
		echo "Код отформатирован и импорты отсортированы правильно!"; \
	fi

# Миграции
migrate-up:
	@echo "Применение миграций..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" up; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi

migrate-down:
	@echo "Откат миграций..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" down; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi

migrate-force:
	@echo "Принудительная установка версии миграции..."
	@if [ -n "$(DATABASE_DSN)" ] && [ -n "$(VERSION)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" force $(VERSION); \
	else \
		echo "Ошибка: DATABASE_DSN или VERSION не установлены"; \
		exit 1; \
	fi

migrate-version:
	@echo "Текущая версия миграции..."
	@if [ -n "$(DATABASE_DSN)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate -path migrations -database "$(DATABASE_DSN)" version; \
	else \
		echo "Ошибка: DATABASE_DSN не установлен"; \
		exit 1; \
	fi

migrate-create:
	@echo "Создание новой миграции..."
	@if [ -n "$(NAME)" ]; then \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate create -ext sql -dir migrations -seq $(NAME); \
		echo "Создание миграции для SQLite..."; \
		export PATH=$$PATH:$$(go env GOPATH)/bin && migrate create -ext sql -dir migrations/sqlite -seq $(NAME); \
		echo "Миграции созданы успешно!"; \
	else \
		echo "Ошибка: NAME не установлен. Используйте: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi


# Генерация кода из proto файлов
proto:
	@echo "Генерация кода из proto файлов..."
	@if ! command -v protoc > /dev/null; then \
		echo "Ошибка: protoc не найден. Установите его: brew install protobuf"; \
		exit 1; \
	fi
	@if [ ! -f "$(shell go env GOPATH)/bin/protoc-gen-go" ]; then \
		echo "Установка protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi
	@if [ ! -f "$(shell go env GOPATH)/bin/protoc-gen-go-grpc" ]; then \
		echo "Установка protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi
	protoc \
		--go_out=internal/generated \
		--go_opt=paths=source_relative \
		--go-grpc_out=internal/generated \
		--go-grpc_opt=paths=source_relative \
		api/proto/shortener.proto

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  swagger  - Генерация Swagger документации"
	@echo "  build    - Сборка проекта"
	@echo "  run      - Запуск проекта"
	@echo "  test     - Запуск тестов"
	@echo "  clean    - Очистка проекта"
	@echo "  deps     - Установка зависимостей"
	@echo "  migrate-up     - Применить миграции (требует DATABASE_DSN)"
	@echo "  migrate-down   - Откатить миграции (требует DATABASE_DSN)"
	@echo "  migrate-force  - Принудительно установить версию (требует DATABASE_DSN и VERSION)"
	@echo "  migrate-version - Показать текущую версию миграции (требует DATABASE_DSN)"
	@echo "  migrate-create - Создать новую миграцию (требует NAME=migration_name)"
	@echo "  fmt      - Форматировать код с помощью gofmt"
	@echo "  fmt-check - Проверить форматирование кода (без изменений)"
	@echo "  imports  - Форматировать код и сортировать импорты с помощью goimports"
	@echo "  imports-check - Проверить форматирование и импорты (без изменений)"
	@echo "  help     - Показать эту справку" 