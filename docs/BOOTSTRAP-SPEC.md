# Bootstrap Spec - Day 1 Setup

Source decisions: `docs/PRD-MVP-1M.md`, `docs/MVP-30D-BACKLOG.md`, `CONTEXT.md`

## 1. Objective

Create a working API foundation for the 30-day MVP with:

- Gin backend in `api/`
- local PostgreSQL for development
- `golang-migrate` migration workflow
- config and secrets via environment variables
- production-ready path to single-VM Docker Compose deployment

## 2. Locked Technical Choices

- Framework: Gin
- API versioning: `/api/v1`
- DB: PostgreSQL
- Data access: `sqlc + pgx`
- Migrations: `golang-migrate` (manual commands)
- Local dev: no Docker requirement
- Production: Docker Compose on single VM

## 3. Required Environment Variables

## API core

- `APP_ENV` (`local`, `staging`, `prod`)
- `API_PORT` (e.g., `8080`)
- `API_BASE_PATH` (fixed to `/api/v1`)

## Database

- `DATABASE_URL` (Postgres DSN)

Local setup (example):

```bash
psql postgres -c "CREATE ROLE nursery_app WITH LOGIN PASSWORD 'nursery_app';"
psql postgres -c "CREATE DATABASE nursery_management OWNER nursery_app;"
```

## Auth

- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`
- `JWT_ACCESS_TTL_MINUTES` (default `15`)
- `JWT_REFRESH_TTL_HOURS` (default `720`)

## Email

- `EMAIL_PROVIDER` (`smtp` for local)
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USER`
- `SMTP_PASS`
- `SMTP_FROM`

## Stripe

- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_PUBLISHABLE_KEY`

## App URLs

- `WEB_BASE_URL` (used in invite/reset links)

## 4. Folder and Module Skeleton

Recommended initial layout:

```text
api/
  cmd/server/
    main.go
  internal/
    platform/
      config/
      db/
      http/
      logger/
    auth/
    users/
    children/
    attendance/
    funding/
    invoicing/
    payments/
    audit/
  db/
    migrations/
    query/
  sqlc.yaml
  go.mod
```

Notes:
- Keep module boundaries now; avoid generic `controllers/services` dumping ground.
- Implement handlers per domain under `/api/v1` routes.

## 5. Day-1 Build Checklist

1. Initialize Go module in `api/`.
2. Add Gin, pgx, sqlc tooling, and migration tooling.
3. Create config loader from env vars.
4. Add DB connection + ping on startup.
5. Add router with `/health` and `/api/v1/health`.
6. Add structured error handling middleware.
7. Add request logging middleware.
8. Create initial migration files:
   - `tenants`, `branches`, `users`, `memberships`, `audit_logs`
9. Run migrations against local Postgres.
10. Boot server and verify health endpoint.

## 6. Migration Workflow (Local)

Rules:
- Migrations are explicit, manual, and versioned.
- API startup does not auto-run migrations.
- Every schema change must have `up` and `down` migration.

Checklist:
1. Create migration file pair for change.
2. Run `migrate up` locally.
3. Verify schema + run smoke checks.
4. Commit migration files with related code.

## 7. Day-1 Acceptance Criteria

- `api/` compiles and runs locally.
- Health endpoints return `200`.
- API can connect to local Postgres.
- Initial migrations apply on clean local DB.
- `/api/v1` base path is active.
- Fail-fast behavior exists for missing critical env vars.

## 8. Seed First Manager (Local)

Use the seed command once to create the initial manager user. It is safe to re-run; it will upsert the user password and ensure a manager membership.

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5432/nursery_management?sslmode=disable \
go run ./cmd/seed \
  --email manager@pilot.local \
  --password "ReplaceWithAStrongPassword" \
  --tenant "Pilot Nursery" \
  --branch "Main"
```

Required flags:

- `--email`
- `--password`

Optional flags:

- `--database-url` (defaults to `DATABASE_URL` env var)
- `--tenant` (defaults to `Pilot Nursery`)
- `--branch` (defaults to `Main`)

## 9. Deployment Readiness Hook (For Later in Month)

Even if deployment is not done on Day 1, prepare these now:

- Dockerfile placeholder for API
- Compose file placeholder for prod services
- env var contract documented and stable
- no local absolute paths in code/config

This keeps local-first development aligned with end-of-month deployment.
