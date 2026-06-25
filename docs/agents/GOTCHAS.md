# Gotchas & Setup

## Environment Prerequisites

- PostgreSQL 14+ required
- Start PostgreSQL: `brew services start postgresql@18` (version varies by install)
- `api/.env` must contain: `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `PASSWORD_RESET_TOKEN_SECRET`

## Common Pitfalls

- **Proxy config:** Frontend API calls go through `proxy.conf.json` (`/api` → `:8080`). Direct `curl` at `:8080/api/v1`.
- **Seed idempotent:** Running seed multiple times updates existing users instead of failing.
- **sqlc gen paths:** Reads from `db/query/*.sql`, outputs to `internal/platform/db/sqlc/`.
- **sqlc tool requirement:** `go tool sqlc generate` requires sqlc listed in `go.mod` tool directives.
- **Tenant + branch scoping:** All SQL queries must filter by tenant and branch — never query across tenants.
