# Development Commands

All `make` commands run from the project root. Commands annotated with `(workdir: ...)` should be run with `workdir` set to that directory.

```bash
# API (workdir: root)
make run-api                  # auto-loads api/.env

# Repo tests (workdir: root, needs TEST_DATABASE_URL)
make test-api-repositories

# Specific test (workdir: api)
go test ./internal/modules/attendance/application/ -run TestCheckIn -v

# Migrations (workdir: root, reads DATABASE_URL from api/.env)
make migrate-up / migrate-down / migrate-reset
make migrate-create name=add_foo   # creates up/down pair in api/db/migrations/
make migrate-verify                 # needs VERIFY_DATABASE_URL (different DB)

# Seed (workdir: api, idempotent — updates existing users)
set -a && source .env && set +a && go run ./cmd/seed \
  -email o@x.com -password 'X' -local \
  -manager-email m@x.com \
  -staff-email s@x.com \
  -parent-email p@x.com

# Codegen (workdir: root)
make sqlc-generate             # go tool sqlc generate from db/query/*.sql → internal/platform/db/sqlc/

# Frontend (workdir: web)
npm install && npm start       # :4200, proxies /api → localhost:8080 (proxy.conf.json)
npm test                       # Karma
```

## sqlc

- `go tool sqlc generate` reads from `db/query/*.sql` and outputs to `internal/platform/db/sqlc/`
- Requires sqlc listed in `go.mod` tool directives
