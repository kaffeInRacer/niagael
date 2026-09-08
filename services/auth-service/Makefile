CONFIG_FILE ?= config.yaml
MIGRATIONS_DIR ?= ./migrations

DB_URL := $(shell yq -r '.postgres.migration_url' $(CONFIG_FILE))
MIGRATIONS_SOURCE := file://$(abspath $(MIGRATIONS_DIR))

.PHONY: migrate-up migrate-down migrate-clean build run

build:
	go build -o ./tmp/auth-service ./cmd/*.go

run: build
	./tmp/auth-service

migrate-up:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" up

migrate-down:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" down 1

migrate-clean:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" drop -f
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" up

migrate-status:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" version
