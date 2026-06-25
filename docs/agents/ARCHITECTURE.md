# Architecture & Conventions

## Layer Diagram

```
Handler → Application (Use Cases) → Domain → Infrastructure (postgres/)
api/internal/
  app/bootstrap/         # wiring only — no globals, no init(), no service locator
  modules/<name>/        # domain/ application/ infrastructure/postgres/ interfaces/http/
  platform/              # audit config db email errors http lifecycle ratelimit tenant transaction uid
```

## Forbidden Imports

- `domain` — must not import postgres/gin/http/sql
- `application` — must not import sql/http/framework types
- handler — must not import DB directly

## Cross-Module Rule

Never import another module directly. Define the interface in the consumer, wire the adapter in `bootstrap/adapters.go`.

## Transaction Pattern

```go
txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
    // repo calls here
})
```

Never use Begin/Commit/Rollback directly. `txMgr.ExecTx` is called only from repository/infrastructure code — the application layer never sees `pgx.Tx`.

## Auth Pattern

```go
actor := tenant.ActorFromGinContext(c)
```

Never parse JWT tokens manually.

## Error Flow

```
DomainError → MapDomainError() → HTTP response
```

Auth errors must use the generic message: "Invalid credentials or session" only.

## Tenant + Branch Scoping

All SQL queries must be scoped by tenant and branch — never query across tenants.

## Modules (16)

- `authentication`, `children`, `parentchildmappings`, `attendance`, `absence`
- `funding`, `billing`, `payments`, `invites`, `passwordreset`, `owner`
- `rooms`, `sessiontypes`, `sessiontemplates`, `term`
- `invoicerun` (scheduler/batch only, no HTTP routes)

Routes are registered per-module in `api/internal/app/bootstrap/bootstrap.go`.

## New Module Checklist

1. `domain/` — entities + repository interface (zero framework imports)
2. `application/` — one use case per file
3. `infrastructure/postgres/` — repository implementation
4. `interfaces/http/` — handler
5. Wire in `bootstrap.go` (bootstrap/)
6. Add route
