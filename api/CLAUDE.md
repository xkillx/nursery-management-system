# API — CLAUDE.md

Architecture reference for the Go backend. Read this before modifying `api/`.

## Architecture

Clean Architecture + Hexagonal + DDD. Incremental migration — slices land module-by-module.

### Layering

```
interfaces/http   →  application/use-case  →  domain
infrastructure    →  implements domain/application ports
app/bootstrap     →  composition root (DI wiring)
```

**Dependency direction**: imports only point inward. `domain` has zero framework imports.

### Forbidden Dependencies

```
domain          → postgres/gin/http/sql
application     → postgres/sql/http types
handler         → db pool direct usage
```

## Module Layout

```
api/internal/
  app/bootstrap/           Composition root — all DI wiring lives here
  modules/
    authentication/        Login, refresh, logout, membership switch
      domain/              Entities, errors, repository interfaces
      application/         Use cases (LoginUseCase, RefreshUseCase, etc.)
      infrastructure/
        postgres/          SQL implementation of repository interfaces
        tokens/            JWT TokenManager (HS256, access + refresh)
      interfaces/http/     Thin HTTP handlers, cookie/CSRF management
    children/              CRUD, mark-inactive, attendance list
    guardians/             CRUD, deactivate/reactivate (cascades links+mappings)
    guardianlinks/         Create (idempotent), end guardian-child links
    parentmappings/        Create (idempotent), end parent-membership-guardian mappings
  platform/
    audit/                 Shared audit log writer (works with pool or tx)
    config/                Environment config (Load, validate)
    db/                    PostgreSQL pool setup
    errors/                Domain error types (NotFound, Validation, etc.)
    http/
      authz_middleware.go  JWT authn + role-based authz, TokenParser interface
      errors.go            ErrorResponse struct, writeError/writeInternalError
      error_mapper.go      MapDomainError — domain errors → HTTP responses
      middleware.go         Request ID, access log, panic recovery
    lifecycle/             Shared reason code constants and validation
    tenant/                ActorContext, AuthorizationContext, context extraction
    transaction/           Transaction manager (ExecTx abstraction)
    uid/                   NewUUID (v7), NewCSRFToken
```

## Module Structure (per module)

Each module under `modules/` follows this layout:

```
module_name/
  domain/
    entities.go        Pure structs, no framework imports
    errors.go          Sentinel errors
    repository.go      Interfaces only — no SQL, no pgx
  application/
    <use_case>.go      One file per use case, pure logic
  infrastructure/
    postgres/
      repository.go    SQL queries, implements domain interfaces
  interfaces/http/
    handler.go         Thin HTTP layer — parse request, call use case, respond
    dto.go             Request/response structs
```

## Key Patterns

### Dependency Injection

All wiring in `app/bootstrap/bootstrap.go`. Construct repos → use cases → handlers. No globals, no service locators, no init().

### Use Case Signature

Use cases take `tenant.ActorContext` as first param after `ctx`. Returns `(Result, error)`. No framework types in signatures.

### Error Flow

1. Domain/application returns `*errors.DomainError` (from `platform/errors`)
2. HTTP handler calls `httpserver.MapDomainError(err, requestID)` → `(status, ErrorResponse)`
3. Specific error types handled individually where needed (e.g., auth hides credential details)

### Audit Writing

All modules use `audit.Writer` from `platform/audit`. Works with both pool and tx:

```go
auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{...})
auditWriter.Write(ctx, pool, actor, audit.WriteParams{...})
```

### Transaction Pattern

Use `transaction.Manager.ExecTx` — no manual begin/rollback/commit:

```go
txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
    // all DB ops use tx
    return nil  // commit on success, rollback on error
})
```

### Cross-Module Checks

Modules that need data from other modules receive checker interfaces in their constructors (defined in their own `domain/`). Adapters in `app/bootstrap/adapters.go` bridge them.

Example: `guardianlinks` needs to check guardian active status → defines `GuardianActiveChecker` interface → bootstrap wires `guardianCheckerAdapter` wrapping `GuardianRepository`.

### Authorization Context

Middleware sets `tenant.AuthorizationContext` on gin context via `tenant.AuthContextKey`. Handlers extract via `tenant.ActorFromGinContext(c)` which returns `tenant.ActorContext` with parsed UUIDs.

### Middleware

- `httpserver.AuthnMiddleware(tokenParser)` — JWT validation, sets auth context
- `httpserver.RequireRoles(roles...)` — role-based guard
- Middleware defines `TokenParser` interface returning `tenant.AuthorizationContext` — no module imports needed

## Routes

All routes preserve the original API contract:

| Method | Path | Roles | Module |
|--------|------|-------|--------|
| POST | /auth/login | public | authentication |
| POST | /auth/refresh | public (cookie) | authentication |
| POST | /auth/logout | public (cookie) | authentication |
| POST | /auth/switch-membership | public (cookie) | authentication |
| GET | /children/attendance | manager, practitioner | children |
| GET | /children | manager | children |
| GET | /children/:id | manager | children |
| POST | /children | manager | children |
| PATCH | /children/:id | manager | children |
| POST | /children/:id/actions/mark-inactive | manager | children |
| GET | /guardians | manager | guardians |
| GET | /guardians/:id | manager | guardians |
| POST | /guardians | manager | guardians |
| PATCH | /guardians/:id | manager | guardians |
| POST | /guardians/:id/actions/deactivate | manager | guardians |
| POST | /guardians/:id/actions/reactivate | manager | guardians |
| POST | /guardian-child-links | manager | guardianlinks |
| POST | /guardian-child-links/:id/actions/end | manager | guardianlinks |
| POST | /parent-membership-guardian-mappings | manager | parentmappings |
| POST | /parent-membership-guardian-mappings/:id/actions/end | manager | parentmappings |

## Adding a New Module

1. Create directory structure under `modules/<name>/`
2. Define domain entities, errors, repository interfaces (no SQL)
3. Write use cases in `application/`
4. Implement repository in `infrastructure/postgres/`
5. Write thin HTTP handler in `interfaces/http/`
6. Wire in `app/bootstrap/bootstrap.go`

## Database

- Migrations: `api/db/migrations/` — manual `golang-migrate`
- SQL lives only in `infrastructure/postgres/` packages
- All queries are tenant+branch scoped
- `sqlc` configured but not yet used for code generation

## Testing

```bash
cd api && go test ./...
```

- Domain/application tests: mock repository interfaces
- Integration tests: hit real PostgreSQL
- HTTP handler tests: use `httptest.NewRecorder` with gin test context
