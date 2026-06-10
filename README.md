# Nursery Management System

A multi-tenant nursery management MVP for UK childcare providers. The system focuses on the core month-1 workflows a nursery needs to operate: attendance, absence tracking, funding deductions, invoice preparation, parent billing, and scoped access for managers, practitioners, and parents.

This repository contains a Go API, an Angular web application, PostgreSQL migrations, and product/architecture documentation.

## Contents

- [Project Status](#project-status)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Run the Project](#run-the-project)
- [Useful Commands](#useful-commands)
- [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)

## Project Status

This project is in active MVP development. The backend contains the main domain flows and the frontend is an Angular workspace built from a dashboard foundation, with nursery-specific screens being added iteratively.

Expect API routes, database schema, and UI screens to evolve as the MVP is completed.

## Features

- Multi-tenant nursery and branch model
- JWT authentication with refresh-token sessions
- Membership-scoped access control
- Manager, practitioner, and parent role boundaries
- Child and guardian lifecycle management
- Guardian-child links and parent access mappings
- Attendance check-in/check-out flows
- Absence markers and correction workflows
- Funding profiles and funding-aware billing inputs
- Draft invoice generation and invoice issue flow
- Parent invoice access and payment-session groundwork
- Stripe webhook reconciliation groundwork
- API health checks, metrics support, and migration tooling

## Tech Stack

| Area | Technology |
| --- | --- |
| Backend | Go, Gin, pgx |
| Database | PostgreSQL |
| Migrations | golang-migrate |
| Frontend | Angular, TypeScript, Tailwind CSS |
| Authentication | JWT access tokens and refresh-token cookies |
| Payments | Stripe integration groundwork |
| Code generation | sqlc |

## Project Structure

```text
nursery-management-system/
  api/                 Go API, migrations, sqlc queries, seed command
  web/                 Angular frontend workspace
  docs/                Product docs, API contract, ADRs, delivery notes
  docs/forms/          Nursery form templates
  CONTEXT.md           Domain glossary and MVP decision baseline
  Makefile             Common local development commands
  README.md            Project setup and usage guide
```

## Prerequisites

Install these tools before running the project locally:

- Git
- Go 1.26 or newer
- Node.js 20 or newer
- npm
- PostgreSQL 14 or newer
- golang-migrate CLI

On macOS with Homebrew:

```bash
brew install git go node postgresql@16 golang-migrate
brew services start postgresql@16
```

Check your installed versions:

```bash
git --version
go version
node --version
npm --version
psql --version
migrate -version
```

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/xkillx/nursery-management-system.git
cd nursery-management-system
```

If you fork this project, replace the GitHub URL with your fork URL.

### 2. Create the local database

The default local database name is `nursery_management`.

```bash
createdb nursery_management
```

If your PostgreSQL setup uses the default `postgres` user and password from `api/.env.example`, the default database URL will work:

```text
postgres://postgres:postgres@localhost:5432/nursery_management?sslmode=disable
```

If your machine uses your macOS/Linux username instead, update `DATABASE_URL` in `api/.env` after the next step. For example:

```text
postgres://your_username@localhost:5432/nursery_management?sslmode=disable
```

### 3. Create the API environment file

```bash
cp api/.env.example api/.env
```

Open `api/.env` and update at least these values:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/nursery_management?sslmode=disable
JWT_ACCESS_SECRET=replace-with-a-long-random-value
JWT_REFRESH_SECRET=replace-with-a-long-random-value
PASSWORD_RESET_TOKEN_SECRET=replace-with-a-long-random-value
INVITE_TOKEN_SECRET=replace-with-a-long-random-value
WEB_BASE_URL=http://localhost:4200
```

You can generate local random secrets with:

```bash
openssl rand -hex 32
```

Do not commit `api/.env`. It is ignored by Git.

### 4. Install frontend dependencies

```bash
cd web
npm ci
cd ..
```

Use `npm install` instead if you intentionally want to update dependency versions.

### 5. Run database migrations

From the repository root:

```bash
make migrate-up
```

This command loads `api/.env` and applies all migrations in `api/db/migrations`.

### 6. Seed the first manager account

Run this once after migrations. The command is safe to re-run; it updates the same user and ensures the manager membership exists.

```bash
cd api
set -a
source .env
set +a
go run ./cmd/seed \
  -email manager@pilot.local \
  -password "ChangeThisPassword123" \
  -tenant "Pilot Nursery" \
  -branch "Main"
cd ..
```

Use this account to sign in locally:

```text
Email: manager@pilot.local
Password: ChangeThisPassword123
```

Use a stronger password for shared or deployed environments.

## Run the Project

Open two terminal windows from the repository root.

### Terminal 1: start the API

```bash
make run-api
```

The API runs on:

```text
http://localhost:8080
```

Check the health endpoint:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/health
```

### Terminal 2: start the web app

```bash
make run-web
```

The Angular dev server runs on:

```text
http://localhost:4200
```

Open `http://localhost:4200` in your browser.

Important local frontend note: the Angular app currently uses a same-origin API base path of `/api/v1`. That is suitable behind a same-origin reverse proxy, but the repository does not currently include an Angular dev proxy configuration. If browser API calls from `localhost:4200` return `404` or hit the Angular dev server instead of the Go API, the backend is still available directly at `http://localhost:8080/api/v1`.

### Test login through the API

You can verify the seeded manager account directly:

```bash
curl -i \
  -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"manager@pilot.local","password":"ChangeThisPassword123"}'
```

## Useful Commands

### Backend

```bash
make run-api
cd api && go test ./...
```

### Frontend

```bash
make run-web
cd web && npm test
cd web && npm run build
```

### Database migrations

```bash
make migrate-up          # Apply pending migrations
make migrate-version     # Show current migration version
make migrate-down        # Roll back one migration
make migrate-down-all    # Roll back all application migrations
make migrate-reset       # Roll back all migrations, then apply them again
```

Create a new migration pair:

```bash
make migrate-create name=add_example_table
```

Verify migrations against a disposable database:

```bash
export VERIFY_DATABASE_URL="postgres://user:pass@localhost:5432/disposable_db?sslmode=disable"
make migrate-verify
```

Never point `VERIFY_DATABASE_URL` at a database you care about. The verification command runs destructive rollback steps.

### Code generation

Regenerate sqlc code after changing files in `api/db/query`:

```bash
make sqlc-generate
```

### Repository tests that need a database

Some repository tests require a disposable test database:

```bash
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/nursery_management_test?sslmode=disable"
make test-api-repositories
```

## Environment Variables

The API loads configuration from environment variables. For local development, copy `api/.env.example` to `api/.env`.

| Variable | Required | Notes |
| --- | --- | --- |
| `APP_ENV` | Yes | Use `local`, `staging`, or `prod`. |
| `API_PORT` | Yes | Defaults to `8080`. |
| `API_BASE_PATH` | Yes | Must be `/api/v1`. |
| `DATABASE_URL` | Yes | PostgreSQL connection string. |
| `JWT_ACCESS_SECRET` | Yes | Secret used to sign access tokens. |
| `JWT_REFRESH_SECRET` | Yes | Secret used for refresh-token handling. |
| `WEB_BASE_URL` | Yes | Local default is `http://localhost:4200`. |
| `SMTP_HOST` / `SMTP_PORT` | Yes | Defaults expect SMTP on `localhost:1025`. |
| `SMTP_FROM` | Yes | Sender address for local email flows. |
| `PASSWORD_RESET_TOKEN_SECRET` | Yes | Secret for password reset tokens. |
| `INVITE_TOKEN_SECRET` | Yes | Secret for invite tokens. |
| `STRIPE_SECRET_KEY` | No locally | Required outside `local` when payment flows are enabled. |
| `STRIPE_WEBHOOK_SECRET` | No locally | Required outside `local` when Stripe webhooks are configured. |
| `SCHEDULER_OWNER` | No | Set `true` for the process that should run scheduled jobs. |
| `LOG_LEVEL` | No | One of `debug`, `info`, `warn`, or `error`. |
| `METRICS_ENABLED` | No | Defaults on for `local` and `staging`. |

For local email testing, a tool such as Mailpit can provide SMTP on port `1025`:

```bash
brew install mailpit
mailpit
```

Mailpit's browser UI usually runs on `http://localhost:8025`.

## Troubleshooting

### `DATABASE_URL is required`

Make sure `api/.env` exists and contains `DATABASE_URL`, then run commands from the repository root with the provided Makefile targets.

```bash
cp api/.env.example api/.env
make migrate-up
```

### `failed to connect postgres`

Start PostgreSQL and confirm that the username, password, host, port, and database name in `DATABASE_URL` are correct.

```bash
brew services start postgresql@16
psql "postgres://postgres:postgres@localhost:5432/nursery_management?sslmode=disable"
```

If your local PostgreSQL user is different, update `DATABASE_URL` in `api/.env`.

### `migrate: command not found`

Install the golang-migrate CLI:

```bash
brew install golang-migrate
```

### `relation does not exist`

The database exists, but migrations have not been applied.

```bash
make migrate-up
```

### `JWT_ACCESS_SECRET is required`

Set all required secrets in `api/.env`.

```bash
openssl rand -hex 32
```

### Frontend API calls return `404`

The Angular app uses `/api/v1` as a same-origin API path. The Go API runs on `localhost:8080`, while Angular runs on `localhost:4200`. Until a local dev proxy or reverse proxy is added, test API endpoints directly on `http://localhost:8080/api/v1`.

### Port already in use

The API defaults to port `8080` and the web app defaults to port `4200`.

For the API, change `API_PORT` in `api/.env`.

For Angular, run:

```bash
cd web
npm start -- --port 4300
```

## Documentation

Start with these files:

- `CONTEXT.md`: domain glossary and MVP constraints
- `docs/PRD-MVP-1M.md`: month-1 MVP product scope
- `docs/BOOTSTRAP-SPEC.md`: backend bootstrap decisions
- `docs/API-CONTRACT-MVP.openapi.yaml`: API contract
- `docs/API-SCHEMA-STATE.md`: current API/schema state
- `docs/adr/`: architecture decision records
- `docs/forms/`: nursery form templates

## License

No license file is currently included. Add a license before publishing or accepting external contributions.
