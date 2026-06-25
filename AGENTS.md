# AGENTS.md

Multi-tenant nursery management (UK). Go 1.26 (Gin+pgx) + Angular 21 + PostgreSQL. Clean Architecture.

**Package managers:** `npm` (frontend), Go toolchain via `go` + `make`.

**Non-standard build commands:**
- `make run-api` — starts API server (auto-loads `api/.env`)
- `make sqlc-generate` — regenerates sqlc code from `db/query/*.sql` → `internal/platform/db/sqlc/`

**Cross-cutting rules (applies to every task):**
- **Forbidden imports:** `domain` → postgres/gin/http/sql; `application` → sql/http/framework types; handler → direct DB
- **Cross-module:** never import another module directly. Define interface in consumer, wire adapter in `bootstrap/adapters.go`.
- **Transactions:** always `txMgr.ExecTx(ctx, func(tx pgx.Tx) error{...})`. Never Begin/Commit/Rollback directly.
- **Auth:** actor from `tenant.ActorFromGinContext(c)`. Never parse JWT manually.
- **Error flow:** DomainError → MapDomainError() → HTTP. Auth: generic "Invalid credentials or session" only.

---

| Reference | Covers |
|---|---|
| [Architecture & Conventions](docs/agents/ARCHITECTURE.md) | Layer detail, module list, new module checklist |
| [Development Commands](docs/agents/COMMANDS.md) | All make/go/npm commands with workdir hints |
| [Testing Patterns](docs/agents/TESTING.md) | Mock vs integration, handler tests, test DB setup |
| [Git Workflow](docs/agents/GIT-WORKFLOW.md) | FF-only merge, rebase conventions |
| [Project Context](docs/agents/PROJECT-CONTEXT.md) | Post-MVP roadmap, current active work |
| [Gotchas & Setup](docs/agents/GOTCHAS.md) | Environment prerequisites, common pitfalls |
