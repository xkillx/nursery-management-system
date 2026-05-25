# API Coding Rules

## Architecture

Use:

- Clean Architecture
- Hexagonal Architecture
- DDD

Dependency flow:

```text
HTTP/Handler
    ↓
Application (Use Cases)
    ↓
Domain
    ↓
Infrastructure implements interfaces
```

Rules:

- Dependencies only point inward
- `domain` must have zero framework imports
- No globals
- No `init()`
- No service locator

---

## Forbidden

Never do:

```text
domain → postgres/gin/http/sql
application → sql/http/framework types
handler → direct database access
```

---

## Project Structure

```text
api/internal/

app/bootstrap/
    bootstrap.go       # dependency wiring only
    token_parser.go    # adapter: TokenManager → TokenParser interface
    adapters.go        # cross-module checker adapters

modules/
    authentication/    # login, refresh, logout, membership switch
    children/          # CRUD, mark-inactive, attendance list
    guardians/         # CRUD, deactivate/reactivate (cascades links+mappings)
    guardianlinks/     # create (idempotent), end guardian-child links
    parentmappings/    # create (idempotent), end parent-membership-guardian mappings

    <module>/
        domain/
            entities.go
            errors.go
            repository.go

        application/
            <usecase>.go

        infrastructure/
            postgres/
                repository.go
            tokens/               # authentication only
                tokens.go

        interfaces/http/
            handler.go
            dto.go

platform/
    audit/               # shared audit writer (pool or tx)
    config/              # environment config
    db/                  # postgres pool setup
    errors/              # domain error types
    http/
        authz_middleware  # JWT authn + role authz, TokenParser interface
        errors.go         # ErrorResponse struct
        error_mapper.go   # MapDomainError
        middleware.go      # request ID, access log, recovery
    lifecycle/           # shared reason codes and validation
    tenant/              # ActorContext, AuthorizationContext
    transaction/         # ExecTx manager
    uid/                 # NewUUID (v7), NewCSRFToken
```

---

## Domain Rules

- Pure business logic only
- No framework imports
- No SQL
- No HTTP
- Repository = interface only

Example:

```go
type UserRepository interface {}
```

---

## Use Case Rules

- One use case per file
- Pure business logic
- No HTTP objects
- No DB objects

Signature:

```go
func(ctx context.Context, actor tenant.ActorContext, ...)
```

Return:

```go
(Result, error)
```

---

## Handler Rules

Handlers must be thin:

Allowed:

- Parse request
- Validate input
- Call use case
- Return response

Forbidden:

- SQL query
- Business logic
- Direct DB access

---

## Repository Rules

Inside:

```text
infrastructure/postgres/
```

Allowed:

- SQL queries
- Database operations
- Implement repository interfaces

Forbidden:

- Business logic

---

## Dependency Injection

All dependency wiring:

```text
app/bootstrap/bootstrap.go
```

Flow:

```go
Repository
    ↓
UseCase
    ↓
Handler
```

Never:

- Global variables
- init()
- Service locator

---

## Transaction Rules

Always:

```go
txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
    return nil
})
```

Never:

```go
Begin()
Commit()
Rollback()
```

---

## Error Rules

Flow:

```text
Domain/Application
    ↓
DomainError
    ↓
MapDomainError()
    ↓
HTTP response
```

Auth handler is special: returns generic "Invalid credentials or session" to prevent info leakage. Auth errors bypass MapDomainError.

---

## Cross Module Communication

Do not directly import another module.

Use interfaces:

```go
type GuardianChecker interface {}
```

Wire adapters in:

```text
app/bootstrap/adapters.go
```

Current adapters:

- `guardianCheckerAdapter` — guardianlinks/parentmappings check guardian active status
- `childCheckerAdapter` — guardianlinks check child exists in scope
- `membershipCheckerAdapter` — parentmappings check membership role/active status

---

## Authorization

Get actor only from:

```go
tenant.ActorFromGinContext(c)
```

Do not manually parse JWT.

Middleware handles:

- JWT validation
- Authorization context
- Role checks

Middleware defines `TokenParser` interface returning `tenant.AuthorizationContext` — no module imports needed. Adapter in `bootstrap/token_parser.go` wraps `tokens.TokenManager`.

---

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

---

## Audit

All modules use `audit.Writer`. Works with pool or tx:

```go
auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{...})
auditWriter.Write(ctx, pool, actor, audit.WriteParams{...})
```

---

## Database Rules

- Migrations → `db/migrations/`
- SQL only inside `infrastructure/postgres/`
- All queries must be tenant + branch scoped
- `sqlc` configured but not yet used for code generation

---

## Testing

Run:

```bash
cd api && go test ./...
```

Test strategy:

- Domain/Application → mock repositories
- Integration → real PostgreSQL
- Handler → httptest + gin context

---

## Adding New Module

1. Create module folder under `modules/<name>/`
2. Create domain entities + interfaces
3. Create use cases
4. Implement repository
5. Create handler
6. Wire in bootstrap
