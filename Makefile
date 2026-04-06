MIGRATIONS_DIR := ./migrations
DB_URL ?= postgresql://postgres:postgres@localhost:5432/taskservice?sslmode=disable
	
.PHONY: migrate-create migrate-up migrate-down migrate-rollback migrate-force

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Ошибка: Укажите имя миграции. Пример: make migrate-create name=init_schema"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) up

migrate-down:
	@read -p "ВНИМАНИЕ! Это удалит ВСЕ таблицы. Вы уверены? [y/N]: " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down; \
	else \
		echo "Отменено."; \
	fi

migrate-rollback:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down 1

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "Ошибка: Укажите версию. Пример: make migrate-force version=1"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) force $(version)