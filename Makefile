.PHONY: run-api run-web build-api migrate-up migrate-down migrate-down-all migrate-reset migrate-version migrate-create migrate-verify sqlc-generate test-api-repositories swagger-generate swagger-validate

API_DIR := api
WEB_DIR := web
API_ENV := $(API_DIR)/.env
MIGRATIONS_DIR := $(API_DIR)/db/migrations
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X nursery-management-system/api/internal/platform/version.Version=$(VERSION) \
           -X nursery-management-system/api/internal/platform/version.Commit=$(COMMIT) \
           -X nursery-management-system/api/internal/platform/version.Date=$(BUILD_DATE)

run-api:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; cd "$(API_DIR)" && go run ./cmd/server

build-api:
	@cd "$(API_DIR)" && go build -ldflags "$(LDFLAGS)" -o ../bin/server ./cmd/server

.PHONY: debug-api
debug-api:
	@if [ -f "$(API_ENV)" ]; then set -a; . "$(API_ENV)"; set +a; fi; LOG_LEVEL=debug cd "$(API_DIR)" && go run ./cmd/server
	@echo ""
	@echo "Tip: pipe output to a log file:"
	@echo "  make debug-api 2>&1 | tee tmp/api.log"
	@echo "Then search with:"
	@echo "  rg '<request_id>' tmp/api.log"

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

wire-generate:
	@cd "$(API_DIR)" && go tool wire gen ./internal/app/bootstrap/...

.PHONY: generate
generate: sqlc-generate wire-generate

test-api-repositories:
	@test -n "$$TEST_DATABASE_URL" || (echo "TEST_DATABASE_URL is required (use a disposable database)" && exit 1)
	cd "$(API_DIR)" && go test \
		./internal/modules/authentication/infrastructure/postgres/ \
		./internal/modules/children/infrastructure/postgres/ \
		./internal/modules/guardians/infrastructure/postgres/ \
		./internal/modules/guardianlinks/infrastructure/postgres/ \
		./internal/modules/parentmappings/infrastructure/postgres/ \
		./internal/modules/attendance/infrastructure/postgres/ \
		./internal/modules/funding/infrastructure/postgres/ \
		./internal/modules/payments/infrastructure/postgres/

swagger-generate:
	@cd "$(API_DIR)" && go tool swag init -g cmd/server/main.go --parseInternal --parseDependency -o ./docs

swagger-validate:
	@cd "$(API_DIR)" && go tool swag fmt --dir cmd,internal
	@echo "Swagger annotations formatted."
