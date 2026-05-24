.PHONY: run-api run-web migrate-up migrate-down migrate-down-all migrate-reset migrate-version migrate-create migrate-verify sqlc-generate

API_DIR := api
WEB_DIR := web
API_ENV := $(API_DIR)/.env
MIGRATIONS_DIR := $(API_DIR)/db/migrations

run-api:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; cd "$(API_DIR)" && go run ./cmd/server

run-web:
	@cd "$(WEB_DIR)" && npm start

migrate-up:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1); migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" up

migrate-down:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1); migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" down 1

migrate-down-all:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1); migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" down -all

migrate-reset:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1); migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" down -all && migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" up

migrate-version:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1); migrate -path "$(MIGRATIONS_DIR)" -database "$$DATABASE_URL" version

migrate-verify:
	@test -n "$$VERIFY_DATABASE_URL" || (echo "VERIFY_DATABASE_URL is required (use a disposable database)" && exit 1)
	@echo "==> migrate up"
	@migrate -path "$(MIGRATIONS_DIR)" -database "$$VERIFY_DATABASE_URL" up
	@echo "==> version after first up"
	@migrate -path "$(MIGRATIONS_DIR)" -database "$$VERIFY_DATABASE_URL" version
	@echo "==> migrate down -all"
	@migrate -path "$(MIGRATIONS_DIR)" -database "$$VERIFY_DATABASE_URL" down -all
	@echo "==> migrate up again"
	@migrate -path "$(MIGRATIONS_DIR)" -database "$$VERIFY_DATABASE_URL" up
	@echo "==> version after second up"
	@migrate -path "$(MIGRATIONS_DIR)" -database "$$VERIFY_DATABASE_URL" version
	@echo "==> verification complete"

migrate-create:
	@test -n "$(name)" || (echo "name is required, usage: make migrate-create name=add_table_name" && exit 1)
	@migrate create -ext sql -dir "$(MIGRATIONS_DIR)" -seq "$(name)"

sqlc-generate:
	@cd "$(API_DIR)" && go tool sqlc generate
