# Nursery Management System

Nursery Management System is a multi-tenant MVP for UK nurseries. It focuses on core month-1 operations: attendance, funding deduction, invoicing, and parent payments.

The project currently includes:

- `api/`: Go + Gin backend with PostgreSQL, JWT auth, scoped memberships, and child/guardian lifecycle endpoints
- `web/`: Angular frontend workspace (currently bootstrapped from a dashboard template and ready for product UI work)
- `docs/`: product, architecture, and decision records

## MVP Scope

This repository is currently aligned to a 30-day MVP with:

- Attendance capture and correction model
- Funding-aware billing inputs
- Invoice generation and lifecycle groundwork
- Parent access model through explicit guardian mapping
- Role and scope based authorization (manager, practitioner, parent)

## Tech Stack

- Backend: Go, Gin, pgx, PostgreSQL
- Frontend: Angular 21, TypeScript, Tailwind CSS
- Auth/session model: JWT access tokens + refresh tokens
- Migrations: `golang-migrate` (manual workflow)

## Prerequisites

Install these before running locally:

- Go `1.25+`
- Node.js `20+` and npm
- PostgreSQL `14+`
- `migrate` CLI (`golang-migrate`)

On macOS (Homebrew), a common setup is:

```bash
brew install go node postgresql@16 golang-migrate
```

## Quick Start (Beginner Friendly)

### 1) Clone and enter the repository

```bash
git clone https://github.com/xkillx/nursery-management-system.git
cd nursery-management-system
```

### 2) Create local database

Create a local database that matches the default `DATABASE_URL` in `api/.env.example`:

```bash
createdb nursery_management
```

If your local Postgres user/password is different, update `DATABASE_URL` later in your environment.

### 3) Configure backend environment

Create a local env file from the example:

```bash
cp api/.env.example api/.env
```

Set secure values for:

- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`

### 4) Run database migrations

```bash
set -a
source api/.env
set +a

migrate -path api/db/migrations -database "$DATABASE_URL" up
```

### 5) Start the API

```bash
cd api
set -a
source .env
set +a
go run ./cmd/server
```

The API will be available at:

- `http://localhost:8080/health`
- `http://localhost:8080/api/v1/health`

### 6) Seed the first manager account

In a new terminal at repository root:

```bash
set -a
source api/.env
set +a

SEED_EMAIL=manager@pilot.local SEED_PASSWORD='ChangeThisPassword123' make seed
```

### 7) Start the web app

```bash
cd web
npm install
npm start
```

Open `http://localhost:4200`.

## Useful Commands

### Backend

```bash
cd api
go test ./...
```

### Frontend

```bash
cd web
npm test
npm run build
```

### Re-run migrations from clean state

```bash
set -a
source api/.env
set +a

migrate -path api/db/migrations -database "$DATABASE_URL" down
migrate -path api/db/migrations -database "$DATABASE_URL" up
```

## API Notes

- Base path: `/api/v1`
- Public routes: health and auth endpoints
- Protected routes: require membership-scoped JWT context
- Session model supports explicit membership selection and membership switching

## Project Structure

```text
nursery-management-system/
  api/        Go API (Gin, auth, authorization, lifecycle endpoints)
  web/        Angular frontend workspace
  docs/       PRD, ADRs, and delivery specifications
  CONTEXT.md  Domain glossary and MVP decision baseline
  Makefile    Local seed utility
```

## Documentation

Start with:

- `CONTEXT.md` for domain rules and MVP constraints
- `docs/PRD-MVP-1M.md` for product scope
- `docs/BOOTSTRAP-SPEC.md` for backend bootstrap decisions
- `docs/adr/` for architecture decisions

## Current Status

This codebase is in active MVP development. Expect iterative changes in API routes, schema, and web UI as planned phases land.
