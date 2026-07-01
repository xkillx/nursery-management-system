# AGENTS.md

Multi-tenant nursery management (UK). Go 1.26 (Gin+pgx) + Angular 21 + PostgreSQL. Clean Architecture.

**Package managers:** `npm` (frontend), Go toolchain via `go` + `make`.

**Non-standard build commands:**
- `make run-api` — starts API server (auto-loads `api/.env`)
- `make sqlc-generate` — regenerates sqlc code from `db/query/*.sql` → `internal/platform/db/sqlc/`

**Static analysis (Go):** After any Go file change, run `go fmt ./...`, `go vet ./...`, and `go build ./...` in `api/`. Fix all `fmt` violations and `vet`/`build` warnings/errors before committing.

**Static analysis (Angular):** After any Angular file change, run `npm run lint` in `web/` and fix all lint errors. After `npm test`, run `ng build` (production) to confirm zero errors and warnings — fail the task if any build diagnostic is emitted.

**Cross-cutting rules (applies to every task):**
- **Plan first:** Before writing or modifying any code, create an implementation plan covering the approach, files affected, and any design decisions.
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
| [Documented Solutions](docs/solutions/) | Past bug fixes, best practices, and workflow patterns — search by category (e.g., `integration-issues/`) or YAML frontmatter (`module`, `tags`, `problem_type`) |
| [Shared Vocabulary](CONCEPTS.md) | Domain entities, named processes, and status concepts with project-specific meaning |
| [Design System](DESIGN.md) | UI/UX design reference — colors, typography, components, layout, form patterns, interaction rules, responsive breakpoints, design tokens |
