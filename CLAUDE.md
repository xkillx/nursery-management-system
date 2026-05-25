# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

```bash
# API
cd api && go run ./cmd/server          # start API (needs .env loaded)
make run-api                           # same, auto-loads api/.env
cd api && go test ./...                # all tests
cd api && go test ./internal/modules/authentication/...  # single module tests
cd api && go test ./internal/modules/authentication/application/ -run TestLogin -v  # single test

# Migrations
make migrate-up                        # apply all pending
make migrate-down                      # rollback one
make migrate-down-all                  # rollback all
make migrate-reset                     # down-all then up
make migrate-create name=add_xxx       # create new migration pair
make migrate-verify                    # idempotency check (needs VERIFY_DATABASE_URL)

# Code generation
make sqlc-generate                     # regenerate sqlc from db/query/*.sql

# Frontend
cd web && npm install && npm start     # dev server on :4200
cd web && npm run build                # production build
cd web && npm test                     # karma tests

# Seed
cd api && set -a && source .env && set +a && SEED_EMAIL=x@y.local SEED_PASSWORD='Pass123' go run ./cmd/seed
```

Environment: copy `api/.env.example` to `api/.env`. Set `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `PASSWORD_RESET_TOKEN_SECRET`. PostgreSQL 14+ required.

## Project Overview

Multi-tenant nursery management MVP (UK). Go API + Angular frontend + PostgreSQL.

- `api/` — Go backend (Gin, pgx, JWT auth, scoped memberships)
- `web/` — Angular 21 + Tailwind 4 frontend (TailAdmin template)
- `docs/` — PRD, ADRs, backlog specs
- `CONTEXT.md` — domain glossary and MVP decision baseline

## Architecture

Clean Architecture / Hexagonal / DDD. Dependencies point inward only.

```text
HTTP Handler → Application (Use Cases) → Domain → Infrastructure (implements interfaces)
```

### Forbidden imports

```text
domain → postgres/gin/http/sql
application → sql/http/framework types
handler → direct database access
```

### API structure

```text
api/internal/
  app/bootstrap/        # wiring only: bootstrap.go, adapters.go, token_parser.go
  modules/
    authentication/     # login, refresh, logout, membership switch
    children/           # CRUD, mark-inactive, attendance list
    guardians/          # CRUD, deactivate/reactivate (cascades links+mappings)
    guardianlinks/      # create (idempotent), end guardian-child links
    parentmappings/     # create (idempotent), end parent-membership-guardian mappings
    passwordreset/      # request token, reset password
    attendance/         # (in progress)

    <module>/
      domain/           # entities, errors, repository interface (zero framework imports)
      application/      # one use case per file, pure logic, signature: (ctx, actor, ...) → (Result, error)
      infrastructure/
        postgres/       # repository impl, SQL lives here only
        tokens/         # authentication module only
      interfaces/http/  # thin handler + DTO

  platform/
    audit/              # shared audit writer (pool or tx)
    config/             # env config
    db/                 # postgres pool + sqlc generated code
    email/              # smtp + fake sender
    errors/             # domain error types
    http/               # authz middleware, error mapper, request middleware
    lifecycle/          # shared reason codes and validation
    ratelimit/          # rate limiter
    tenant/             # ActorContext, AuthorizationContext
    transaction/        # ExecTx manager
    uid/                # NewUUID (v7), NewCSRFToken
```

### Frontend structure

```text
web/src/app/
  core/       config, constants, errors, guards, http, models, services, utils
  features/   staff (child-form, guardian-form, manager-children, manager-guardians, practitioner-attendance-children)
  pages/      auth-pages, dashboard, forms, invoices, charts, calendar, other-page
  shared/     components, data, models
```

## Key Patterns

### Dependency injection

All wiring in `app/bootstrap/bootstrap.go`. Repository → UseCase → Handler. No globals, no `init()`, no service locator.

### Transactions

Always use `txMgr.ExecTx(ctx, func(tx pgx.Tx) error { ... })`. Never call `Begin/Commit/Rollback` directly.

### Cross-module communication

Never import another module directly. Define interface in consumer, wire adapter in `bootstrap/adapters.go`.

Current adapters: `guardianCheckerAdapter`, `childCheckerAdapter`, `membershipCheckerAdapter`.

### Authorization

Get actor from `tenant.ActorFromGinContext(c)`. Never parse JWT manually. Middleware handles JWT validation + role checks. `TokenParser` interface returns `tenant.AuthorizationContext`.

### Error flow

Domain/Application → `DomainError` → `MapDomainError()` → HTTP response. Auth handler returns generic "Invalid credentials or session" to prevent info leakage.

### Audit

All modules use `audit.Writer`. Works with pool or tx:
```go
auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{...})
auditWriter.Write(ctx, pool, actor, audit.WriteParams{...})
```

## Routes

| Method | Path | Roles | Module |
|--------|------|-------|--------|
| POST | /auth/login | public | authentication |
| POST | /auth/refresh | public (cookie) | authentication |
| POST | /auth/logout | public (cookie) | authentication |
| POST | /auth/switch-membership | public (cookie) | authentication |
| POST | /auth/password-reset-requests | public | passwordreset |
| POST | /auth/password-resets | public | passwordreset |
| GET | /children/attendance | manager, practitioner | children |
| GET/POST/PATCH | /children[/:id] | manager | children |
| POST | /children/:id/actions/mark-inactive | manager | children |
| GET/POST/PATCH | /guardians[/:id] | manager | guardians |
| POST | /guardians/:id/actions/deactivate | manager | guardians |
| POST | /guardians/:id/actions/reactivate | manager | guardians |
| POST | /guardian-child-links | manager | guardianlinks |
| POST | /guardian-child-links/:id/actions/end | manager | guardianlinks |
| POST | /parent-membership-guardian-mappings | manager | parentmappings |
| POST | /parent-membership-guardian-mappings/:id/actions/end | manager | parentmappings |

## Database

- Migrations: `api/db/migrations/` (golang-migrate, sequential numbering)
- SQL queries: only inside `infrastructure/postgres/`
- sqlc: `api/db/query/*.sql` → generated Go in `internal/platform/db/sqlc/`
- All queries must be tenant + branch scoped

## Testing Strategy

- Domain/Application: mock repositories
- Integration: real PostgreSQL
- Handler: httptest + gin context

## Adding New Module

1. Create `modules/<name>/` with domain/application/infrastructure/interfaces layers
2. Define domain entities + repository interface (zero framework imports)
3. Create use cases (one per file)
4. Implement postgres repository
5. Create handler + DTO
6. Wire in `bootstrap.go`
7. Add route in handler setup
