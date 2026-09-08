CONFIG_FILE ?= config.yaml
MIGRATIONS_DIR ?= ./migrations

DB_URL := $(shell yq -r '.postgres.migration_url' $(CONFIG_FILE))
MIGRATIONS_SOURCE := file://$(abspath $(MIGRATIONS_DIR))

.PHONY: migrate-up migrate-down migrate-clean migrate-add proto-gen proto-gen-product proto-gen-pricing

proto-gen:
	cd proto && protoc --go_out=$(file) --go_opt=paths=source_relative \
		--go-grpc_out=$(file) --go-grpc_opt=paths=source_relative \
		$(file).proto

migrate-up:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" up

migrate-down:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" down 1

migrate-clean:
	migrate -source "$(MIGRATIONS_SOURCE)" -database "$(DB_URL)" drop -f

migrate-add:
	@test -n "$(name)" || (printf '%s\n' 'usage: make migrate-add name=create_products_table'; exit 1)
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" -seq "$(name)"
