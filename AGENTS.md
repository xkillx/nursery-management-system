# AGENTS.md

## TL;DR

Multi-tenant nursery mgmt (UK). Go 1.26 (Gin+pgx) + Angular 21 + PostgreSQL. Clean Architecture. MVP done; Post-MVP roadmap in [`docs/POST-MVP-ROADMAP.md`](docs/POST-MVP-ROADMAP.md). Domain glossary: [`CONTEXT.md`](CONTEXT.md) (200+ entries).

## Commands

```bash
# API
make run-api                  # auto-loads api/.env
make test-api-repositories    # needs TEST_DATABASE_URL (disposable DB)
cd api && go test ./internal/modules/attendance/application/ -run TestCheckIn -v

# Migrations (reads DATABASE_URL from api/.env)
make migrate-up / migrate-down / migrate-reset
make migrate-create name=add_foo   # creates up/down pair in api/db/migrations/
make migrate-verify                 # needs VERIFY_DATABASE_URL (different DB)

# Seed
cd api && set -a && source .env && set +a && go run ./cmd/seed -email o@x.com -password 'X' -local -manager-email m@x.com -staff-email s@x.com -parent-email p@x.com

# Codegen
make sqlc-generate             # go tool sqlc generate from api/db/query/*.sql → internal/platform/db/sqlc/

# Frontend
cd web && npm install && npm start   # :4200, proxies /api → localhost:8080 (proxy.conf.json)
cd web && npm test                    # Karma
```

## Architecture

```
Handler → Application (Use Cases) → Domain → Infrastructure (postgres/)
api/internal/
  app/bootstrap/         # wiring only — no globals, no init(), no service locator
  modules/<name>/        # domain/ application/ infrastructure/postgres/ interfaces/http/
  platform/              # audit config db email errors http lifecycle ratelimit tenant transaction uid
```

**Forbidden imports:** domain→postgres/gin/http/sql, application→sql/http/framework types, handler→direct DB.

**Cross-module:** never import another module directly. Define interface in consumer, wire adapter in `bootstrap/adapters.go`.

**Transactions:** always `txMgr.ExecTx(ctx, func(tx pgx.Tx) error{...})`. Never Begin/Commit/Rollback directly.

**Auth:** actor from `tenant.ActorFromGinContext(c)`. Never parse JWT manually.

**Error flow:** DomainError → MapDomainError() → HTTP. Auth: generic "Invalid credentials or session" only.

**New module checklist:** domain (entities+repo interface, zero framework imports) → application (one use case per file) → postgres repo → handler → wire in `bootstrap.go` → add route.

## Modules (16)

- `authentication`
- `children`
- `parentchildmappings`
- `attendance`
- `absence`
- `funding`
- `billing`
- `payments`
- `invites`
- `passwordreset`
- `owner`
- `rooms`
- `sessiontypes`
- `sessiontemplates`
- `term`
- `invoicerun` (scheduler/batch only, no HTTP routes)

Routes are registered per-module in [`api/internal/app/bootstrap/bootstrap.go`](api/internal/app/bootstrap/bootstrap.go).

## Post-MVP Context

Owner cross-site oversight done (API-PM-08). Pilot readiness gate next. Registration/consent is current active work. See [`docs/POST-MVP-ROADMAP.md`](docs/POST-MVP-ROADMAP.md) and [`CONTEXT.md`](CONTEXT.md) (domain glossary, 200+ entries).

## Testing

- Domain/Application: mock repos
- Integration: real PostgreSQL
- Handler: httptest + gin context
- Repo tests: require `TEST_DATABASE_URL` pointing to a disposable test DB

## Git

- **FF-only merge:** always `git merge --ff-only`. Rebase feature branches onto target before merging. No merge commits.

## Gotchas

- PostgreSQL 14+ needed; `brew services start postgresql@18` (version varies by install)
- `api/.env` must have `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `PASSWORD_RESET_TOKEN_SECRET`
- Frontend API calls work via proxy.conf.json (`/api`→`:8080`); direct curl at `:8080/api/v1`
- Seed idempotent (updates existing users)
- Queries must be tenant + branch scoped
- sqlc gen reads from `db/query/*.sql`, outputs to `internal/platform/db/sqlc/`
- `go tool sqlc generate` requires sqlc listed in go.mod tool directives
