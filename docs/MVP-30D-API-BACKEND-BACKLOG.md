# MVP 30-Day API Backend Backlog

Source: `docs/PRD-MVP-1M.md`  
Related baseline: `docs/DECISION-BASELINE.md`, `CONTEXT.md`, `CLAUDE.md`

## Goal and Scope

Build the Gin API backend for the month-1 nursery MVP so the pilot nursery can run:

- manager-invited staff and parent access
- child and guardian management
- daily attendance check-in, check-out, and manager corrections
- Funding v1 monthly allowance setup and deterministic deduction
- monthly invoice draft generation, issue, parent invoice visibility, and payment status tracking
- Stripe hosted Checkout payment collection with idempotent webhook processing
- backend-owned pilot operations: production env contract, Docker Compose, scheduler ownership, backup/restore checklist, webhook monitoring, and UAT support

This backlog is API/backend-only. It includes Go code, migrations, `sqlc` queries for new modules, API contracts, authorization, audit logging, background jobs, backend tests, and deployment/backend operations. Angular UI implementation is covered separately in `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`.

## Current API State

The repo already contains a working API foundation in `api/`:

- Gin bootstrap, `/api/v1` routing, health endpoints, request ID/logging/recovery middleware.
- PostgreSQL connection and manual `golang-migrate` migrations.
- Email/password login, refresh, logout, and membership switch with session-bound membership scope.
- Authenticated `/me` and authorization probe routes.
- Manager child CRUD and child mark-inactive lifecycle.
- Manager guardian CRUD, guardian deactivation/reactivation, guardian-child link create/end, and parent membership-to-guardian mapping create/end.
- Attendance daily list plus check-in, check-out, and manager correction endpoints.
- Audit writer and audit calls for existing child, guardian, link, mapping, membership switch, and attendance changes.

Known backend gaps:

- Manager invite, invite acceptance, and password-reset APIs are not implemented.
- Funding, invoicing, payments, Stripe integration, scheduler, backup, and deployment files are not implemented.
- `api/sqlc.yaml` exists, but current repositories mostly use hand-written `pgx`; `api/db/query/` is effectively empty.
- No production Docker Compose or API Dockerfile exists.

## Context Alignment Notes

- Use `check-in` and `check-out`; do not use sign-in/sign-out for child attendance.
- `Attendance Session` is the effective attendance interval used for billing.
- `Funding v1` is a simple monthly funded-hours allowance per child.
- Funded-hours deduction applies only to core childcare hours; extras remain payable.
- Monthly invoice billable hours are derived from attendance actuals.
- Each attendance session is rounded up to the nearest 15 minutes for billing.
- Attendance day boundaries and billing calculations use `Europe/London`.
- Invoices are per child, not family-combined.
- Issued invoices are immutable; direct edit APIs must not exist.
- Parents pay issued invoices in full through hosted Stripe Checkout.
- Parent invoice access uses the active parent membership -> guardian -> child link chain.
- Practitioners have attendance access only and must not receive guardian contact, billing, invoice, funding, or payment data.
- Tenant and branch scope are enforced on every protected endpoint from the active session membership.

## Decisions to Honor

- Create this as a separate API/backend backlog; do not replace `docs/MVP-30D-BACKLOG.md`.
- Keep a full API-only 30-day backlog, including already-built foundation and attendance items as verify/harden checkpoints.
- Follow the existing API architecture from `CLAUDE.md`: handler -> application/use case -> domain -> infrastructure repository.
- Use `sqlc + pgx` for all new funding, invoicing, payment, invite, password-reset, and job-related persistence.
- Do not force a risky refactor of already-working auth, child, guardian, mapping, and attendance repositories unless a task touches them for a bug or contract gap.
- Track full migration of existing hand-written query repositories to `sqlc` as post-MVP technical debt; do not make it a blocker for month-1 pilot features.
- Include backend-owned production operations in Week 4.
- Treat reporting basics and CSV export as stretch, not core.
- Treat absence marker API as stretch, not core.
- Implement issued-invoice immutability as core. Add only the schema/state hooks needed for future adjustment invoices in the core path; full adjustment endpoints are stretch unless UAT requires them.
- Public routes remain health and authentication/account-recovery routes only; all business routes require authorization guards.
- API errors use `{ code, message, details?, request_id }`.

## Recommended Libraries and Frameworks

These are preferred implementation defaults for the MVP API backlog. Keep them wrapped behind platform-layer interfaces where practical so the handler -> application/use case -> domain -> infrastructure repository flow remains intact.

| Area | Recommended choice | Backlog fit |
|---|---|---|
| HTTP framework | Keep existing Gin | Matches the current API bootstrap, `/api/v1` routing, middleware model, and PRD requirement to follow the existing architecture. |
| Gin request logging | `github.com/gin-contrib/zap` with Uber Zap, or a small custom Gin middleware around `github.com/rs/zerolog` | Supports API-01 and API-28 structured request logs, request IDs, actor/scope fields, auth denial logs, and webhook status logs. Pick one logger family and use it consistently. |
| Metrics | `github.com/prometheus/client_golang/prometheus` plus `promhttp.Handler()` mounted from Gin | Covers API-28 minimal metrics for auth failures, authorization denials, invoice generation outcomes, Stripe webhook outcomes, and basic HTTP latency/status counts. |
| DB driver and codegen | `sqlc` configured for `pgx/v5` | Required for new funding, invoicing, payments, invites, password-reset, and job persistence while avoiding risky refactors of existing hand-written repositories. |
| Migrations | Existing `github.com/golang-migrate/migrate` workflow | Verified in API-02: `make migrate-verify` (up, down -all, up) on disposable database. Use `migrate-down-all` or `migrate-reset` for full rollback. |
| Stripe | Official `github.com/stripe/stripe-go` module, pinned to the current supported major version during implementation | Fits API-22 to API-24 for hosted Checkout session creation, webhook event parsing/signature verification, payment reconciliation, and retry-safe status updates. |
| Scheduler | `github.com/robfig/cron/v3` | Fits API-20 single-instance overdue invoice transition job. Keep scheduler ownership behind an env flag so only one deployed API process runs scheduled jobs. |
| Future durable jobs | Optional post-MVP `github.com/hibiken/asynq` if Redis-backed persistence/retries become necessary | Not required for the 30-day MVP. Consider only if one-process cron is no longer sufficient. |
| Email delivery | `github.com/wneessen/go-mail` behind `api/internal/platform/email` | Provides a practical SMTP abstraction for API-05 password reset and API-06 invite emails, while keeping provider details replaceable. |
| Lightweight email fallback | Standard `net/smtp` plus a small message builder, or a gomail-style helper | Acceptable only if the MVP needs a thinner dependency and the platform email interface still owns message construction and provider behavior. |

Implementation notes:

- Do not replace Gin, `pgx`, `sqlc`, or `golang-migrate` unless a later decision document explicitly changes the backend baseline.
- Prefer `robfig/cron/v3` for MVP scheduling; defer Redis-backed queues until there is a concrete retry/durability requirement.
- Keep Stripe integration limited to hosted Checkout, PaymentIntent/payment status reads where needed, and webhook processing. Do not add direct card handling.
- Keep email, metrics, logging, scheduler, Stripe, and generated DB access under `api/internal/platform/**` or module-local infrastructure packages rather than leaking vendor APIs into domain code.
- Pin dependency versions in `api/go.mod` when implementing each backlog item, and document any non-obvious version choice in the related PR or runbook.

## Global API Definition of Done

- Endpoint is registered under `/api/v1` with plain JSON resources and standard HTTP status codes.
- Protected endpoint requires authentication and explicit role authorization.
- Every read/write query is scoped by tenant and branch from the active session membership.
- Parent-facing invoice reads validate current parent membership -> guardian -> child relationship at request time.
- Request payloads validate required fields, date/time formats, money values, month values, and enum values before use-case execution.
- Domain/application errors map to stable API error codes using the standard error shape.
- State-changing MVP actions write audit events with actor user, actor membership, tenant, branch, request id, entity, action type, and reason where required.
- New persistence is implemented with migrations plus `sqlc` query files and generated typed Go.
- Migration pair applies cleanly up/down/up on a clean local database.
- Unit tests cover domain rules and state transitions where failures affect billing, authorization, or payments.
- Authorization tests cover unauthenticated, wrong-role, wrong-scope, and allowed-role cases for each route group.
- `go test ./...` passes from `api/`.

## Week 1 - API Foundation, Auth, People, and Contracts

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| ~~API-01~~ | ~~Verify existing Gin bootstrap, config, health endpoints, request ID, structured access logs, recovery, and `/api/v1` base path. Fill only real gaps against `docs/BOOTSTRAP-SPEC.md`.~~ | ~~Existing API~~ | ~~Done. Health endpoints respond; missing critical env vars fail fast; request id appears in errors/logs.~~ |
| ~~API-02~~ | ~~Verify migration workflow and document current schema state. Ensure all current migrations apply cleanly up/down/up on a clean database.~~ | ~~API-01~~ | ~~Done. `make migrate-verify` passes (up → down -all → up), version 7 clean. Fixed migration 000006 index ordering. Schema documented in `docs/API-SCHEMA-STATE.md`.~~ |
| ~~API-03~~ | ~~Add `sqlc` generation workflow for new modules. Add Makefile/npm-equivalent command if missing, keep output at `api/internal/platform/db/sqlc`, and add first harmless query if needed to prove generation.~~ | ~~API-02~~ | ~~Done. `sqlc generate` works and generated code compiles; no existing repository refactor required.~~ |
| ~~API-04~~ | ~~Verify existing auth/session contract: login, refresh, logout, switch membership, CSRF-protected cookie session actions, single-scope auto-selection, multi-scope selection errors.~~ | ~~Existing auth~~ | ~~Done. Auth application and HTTP contract tests cover one-membership auto-selection, multi-membership selection errors, invalid selection, refresh rotation, CSRF-protected cookie session actions, switch-membership scope rotation, and logout idempotency.~~ |
| ~~API-05~~ | ~~Add password reset backend: reset request endpoint, token table, token hashing, expiry, set-new-password endpoint, email dispatch through provider abstraction.~~ **Done 2026-05-25.** | ~~API-03, email provider config~~ | ~~User can request reset and set a new password with a valid token; invalid/expired/used tokens return stable errors.~~ |
| ~~API-06~~ | ~~Add manager invite backend for practitioner and parent roles only: invite create/list/resend/revoke, token hashing, expiry, accept invite, set password, create/activate membership.~~ **Done 2026-05-25.** | ~~API-03, email provider config~~ | ~~Manager can invite non-manager users; manager role is rejected; accepted invite creates login-ready membership in selected tenant/branch. 286 tests pass (188 unit + 98 integration). Migration 000009 verified idempotent.~~ |
| ~~API-07~~ | ~~Harden existing child/guardian/link/mapping APIs against frontend contract. Confirm manager-only writes, active/default listing behavior, lifecycle reason handling, parent access revocation cascades, and no hard delete endpoints.~~ **Done 2026-05-26.** | ~~Existing people modules~~ | ~~Route tests prove role/scope boundaries and lifecycle idempotency (`go test ./internal/app/bootstrap -run TestPeople -count=1`).~~ |
| ~~API-08~~ | ~~Produce an API contract note for frontend integration covering auth, invite, password reset, child, guardian, mapping, and attendance endpoints, including the deferred relationship read endpoints for child detail screens.~~ **Done 2026-05-26.** | ~~API-04 to API-07~~ | ~~`docs/API-CONTRACT-MVP.md` covers all Week 1 routes with request/response examples, error codes, deferred contracts, and known gaps. 188 tests pass.~~ |

## Week 2 - Attendance Hardening and Funding v1

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| ~~API-09~~ | ~~Verify existing attendance daily list, check-in, check-out, and manager correction endpoints against `CONTEXT.md`.~~ | ~~Existing attendance module~~ | ~~Done 2026-05-26: route/role coverage, server-time routine capture, correction action-date semantics, correction reason validation, null routine reasons, contract docs updated.~~ |
| ~~API-10~~ | ~~Harden attendance validation and billing-readiness queries: duplicate open session, no check-out without open session, overlap detection, future correction guard, enrollment-window guard, incomplete-session discovery by month.~~ **Done 2026-05-26.** | ~~API-09, API-03~~ | ~~Internal incomplete-session period query returns open sessions with child details for a billing period; validation tests cover duplicate open, no-open-session check-out, correction overlap detection excluding target, future correction guard using server clock, and enrollment-window guard. 201 tests pass.~~ |
| ~~API-11~~ | ~~Add billing calculation package for attendance-derived minutes. Implement `Europe/London` month boundaries, check-in-month allocation for cross-midnight sessions, incomplete exclusion, and 15-minute round-up per session.~~ **Done 2026-05-26.** | ~~API-10~~ | ~~Pure calculator at `billing/domain`: London month-boundary derivation, check-in-month allocation, 15-minute ceiling rounding, incomplete session exclusions. 27 table-driven tests pass covering same-day, cross-midnight, cross-month, DST, corrected, and multiple-session cases.~~ |
| ~~API-12~~ | ~~Add Funding v1 schema and module: `funding_profiles` with monthly funded-hours allowance per child, manager read/update endpoints, audit events, tenant/branch scope, and validation.~~ **Done 2026-05-26.** | ~~API-03, API-07~~ | ~~Manager can save/read allowance; practitioner/parent are forbidden; funding update audit event is persisted. 239 tests pass. Migration 000010 verified. Funding routes: `GET/PUT /api/v1/funding/children/:child_id`.~~ |
| API-13 | Add Funding v1 calculation service: `max(0, core_attended_minutes - funded_allowance_minutes)` with extras excluded from deduction and calculation components returned for invoice generation. | API-11, API-12 | Tests prove core-only deduction, zero-floor behavior, and deterministic output. |
| API-14 | Add optional absence marker only if core attendance/funding work is complete. Keep it non-billing and manager/practitioner scoped per product decision. | API-09 | Stretch only; absence marker never changes invoice calculations. |

## Week 3 - Invoicing, Parent Billing API, and Stripe

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| API-15 | Add invoice schema with `invoices`, `invoice_lines`, `invoice_issue_runs` or equivalent run tracking, payment-ready fields, and adjustment-link columns for future follow-up invoices. Include status enum/checks for `draft`, `issued`, `payment_failed`, `paid`, `overdue`. | API-03, API-13 | Migration applies cleanly; issued invoices have DB-level protection fields needed for immutability and future adjustments. |
| API-16 | Implement invoice draft preflight endpoint for a calendar month. Return eligible children, blocked children with incomplete attendance, missing enrollment/billing/funding data, and summary totals. | API-10, API-13, API-15 | Preflight blocks only affected children and explains each exception with stable codes. |
| API-17 | Implement draft invoice generation. Generate one draft per eligible child/month from attendance actuals, funding calculation components, current child rate, and manual extras placeholder support. Make regeneration idempotent for non-issued drafts. | API-16 | Re-running generation updates/replaces draft without duplicates; issued invoices are never regenerated. |
| API-18 | Implement invoice review/list/detail endpoints for managers. Include invoice number placeholder for drafts if needed, line items, attended/funded/billable minutes, hourly rate, totals, status, due metadata, and exception references. | API-17 | Manager can inspect draft calculations without spreadsheet reconstruction. |
| API-19 | Implement invoice issue endpoints: one-by-one and bulk issue with confirmation. Assign `INV-YYYYMM-####`, set issued/due timestamps, enforce immutable issued state, and audit issue actions. | API-17 | Bulk issue succeeds for eligible drafts and returns per-child exceptions; issued invoice cannot be directly edited or regenerated. |
| API-20 | Add overdue transition job. A single scheduler instance, controlled by environment flag, marks unpaid issued invoices `overdue` at 00:00 next local day in `Europe/London`. | API-19 | Job is idempotent; disabled by default unless env enables scheduler ownership; tests cover due/overdue boundaries. |
| API-21 | Add parent invoice list/detail endpoints. Parents see issued-or-later invoices only for children authorized through current parent membership -> guardian -> active guardian-child link. | API-19 | Parent cannot see drafts, unlinked child invoices, wrong-tenant invoices, or practitioner/manager-only fields. |
| API-22 | Add Stripe Checkout session creation endpoint for issued, payment-failed, or overdue invoices. Full payment only, GBP only, hosted Checkout only, fresh session per retry. | API-19, Stripe config | Endpoint returns Checkout URL/session id; no custom card handling exists; paid invoices cannot create new checkout sessions. |
| API-23 | Add Stripe webhook endpoint: signature verification, idempotent event storage, payment reconciliation rows, invoice status updates to `paid` or `payment_failed`, and safe retry behavior. | API-22 | Duplicate webhook event is ignored safely; successful payment updates invoice once; failed/canceled payment sets `payment_failed`. |
| API-24 | Add manager payment/reconciliation endpoints: invoice payment status, payment events, checkout retry availability, and webhook processing status. | API-23 | Manager can debug paid/unpaid/failed status without direct Stripe dashboard access for routine checks. |
| API-25 | Add full adjustment invoice endpoint only if UAT or pilot operations require it. Otherwise leave schema hooks and document post-MVP work. | API-19 | Stretch only; any implemented adjustment requires manager reason and links to original issued invoice. |

## Week 4 - Hardening, Reliability, and Pilot Operations

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| API-26 | Add route-by-route authorization test matrix for all MVP endpoints. Cover unauthenticated, wrong role, wrong tenant/branch, parent relationship failure, and allowed access. | Core routes | Tests prove default-deny behavior and stable denial codes. |
| API-27 | Add billing/payment critical tests: funding formula, invoice generation, invoice state transitions, invoice numbering, draft idempotency, issued immutability, overdue job, Stripe webhook idempotency. | API-13 to API-23 | `go test ./...` covers the highest-risk money paths. |
| API-28 | Add structured logs and minimal metrics hooks for webhook outcomes, invoice-generation health, auth failures, and authorization denials. | API-20, API-23 | Logs include request id, actor/scope where available, denial/retry/status codes, and no secrets. |
| API-29 | Add API Dockerfile and production Docker Compose files for single-VM deployment with API, web, PostgreSQL, reverse proxy/HTTPS expectations, and environment file contract. | Core API stable | Compose files exist and document required secrets; no local absolute paths. |
| API-30 | Add backup and restore runbook/checklist for production PostgreSQL. Include daily backup command, restore rehearsal steps, and where backup artifacts live. | API-29 | A developer can perform and verify one local restore rehearsal from the documented steps. |
| API-31 | Add Stripe operational runbook: webhook endpoint setup, webhook secret handling, retry inspection, event replay procedure, and failure triage. | API-23 | Pilot operator can diagnose a failed or duplicated Stripe event. |
| API-32 | Add UAT seed/scenario data for one tenant, one default branch, manager, practitioner, parent, children, guardians, attendance sessions, funding profiles, draft/issued invoices, and payment states. | API-24 | Seed data supports manager/practitioner/parent UAT journeys without manual DB editing. |
| API-33 | Run backend UAT script and fix critical/high defects only. Freeze new backend feature work after this point except pilot blockers. | API-32 | UAT signoff covers attendance, correction, invoice generation, payment, and payment retry. |
| API-34 | Optional reporting and CSV export only after payment loop is stable. Keep these limited to invoice/payment exports required by pilot operations. | API-24 | Stretch only; does not delay Stripe or invoice correctness work. |

## Post-MVP API Technical Debt Backlog

These items are intentionally outside the month-1 pilot critical path. Do them after the payment loop, UAT fixes, and pilot operations are stable.

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| ~~API-TD-01~~ | ~~Migrate existing hand-written `pgx` repository queries to `sqlc`: authentication/session, children, guardians, guardian-child links, parent mappings, and attendance.~~ **Done 2025-05-25.** | ~~Core MVP API stable; route/test coverage green~~ | ~~Existing module tests pass unchanged or with equivalent coverage; `make sqlc-generate` and `go test ./...` pass; no public API behavior or authorization semantics change.~~ |
| ~~API-TD-02~~ | ~~Add env-gated PostgreSQL repository tests for the sqlc-backed authentication/session, children, guardians, guardian-child links, parent membership guardian mappings, and attendance repositories. Keep them behind a disposable database variable such as `TEST_DATABASE_URL` so normal `go test ./...` remains lightweight.~~ **Done 2026-05-25.** | ~~API-TD-01~~ | ~~Repository tests cover no-row behavior, tenant/branch scoping, lifecycle end/deactivate/reactivate writes, refresh-token rotation, attendance session/event writes, overlap detection, and nullable field mapping against a real migrated PostgreSQL database.~~ |
| API-TD-03 | Add an OpenAPI specification generated from or aligned with `docs/API-CONTRACT-MVP.md`, including implemented routes and explicitly deferred proposed contracts. | API-08; core MVP route contract stable | `docs/API-CONTRACT-MVP.openapi.yaml` validates with an OpenAPI linter, includes auth/CSRF security schemes and stable error examples, and is referenced from the Markdown contract note without changing API behavior. |

## Expected API Routes

Existing routes should be verified before adding new variants:

- `GET /api/v1/health`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/switch-membership`
- `GET /api/v1/me`
- `GET /api/v1/children`
- `GET /api/v1/children/:child_id`
- `POST /api/v1/children`
- `PATCH /api/v1/children/:child_id`
- `POST /api/v1/children/:child_id/actions/mark-inactive`
- `GET /api/v1/children/attendance`
- guardian, guardian-child-link, parent-mapping lifecycle routes currently registered by their handlers
- `POST /api/v1/attendance/check-ins`
- `POST /api/v1/attendance/check-outs`
- `POST /api/v1/attendance/corrections`

New route groups to add:

- `POST /api/v1/auth/password-reset-requests`
- `POST /api/v1/auth/password-resets`
- `POST /api/v1/invites`
- `GET /api/v1/invites`
- `POST /api/v1/invites/:invite_id/resend`
- `POST /api/v1/invites/:invite_id/revoke`
- `POST /api/v1/invites/accept`
- `GET /api/v1/funding/children/:child_id`
- `PUT /api/v1/funding/children/:child_id`
- `POST /api/v1/invoice-runs/preflight`
- `POST /api/v1/invoice-runs/drafts`
- `GET /api/v1/invoices`
- `GET /api/v1/invoices/:invoice_id`
- `POST /api/v1/invoices/:invoice_id/issue`
- `POST /api/v1/invoices/bulk-issue`
- `GET /api/v1/parent/invoices`
- `GET /api/v1/parent/invoices/:invoice_id`
- `POST /api/v1/invoices/:invoice_id/checkout-sessions`
- `POST /api/v1/stripe/webhooks`
- `GET /api/v1/payments/events`
- `GET /api/v1/invoices/:invoice_id/payments`

Exact path names may be adjusted to match existing handler naming, but the resource boundaries and role access rules must remain stable.

## Files to Create or Change

Expected backend files and folders:

- `api/db/migrations/000008_*_invites_and_password_resets.{up,down}.sql`
- `api/db/migrations/000009_*_funding_profiles.{up,down}.sql`
- `api/db/migrations/000010_*_invoices_payments_stripe.{up,down}.sql`
- `api/db/query/auth_invites.sql`
- `api/db/query/password_resets.sql`
- `api/db/query/funding.sql`
- `api/db/query/invoices.sql`
- `api/db/query/payments.sql`
- `api/internal/platform/db/sqlc/**`
- `api/internal/platform/email/**`
- `api/internal/platform/jobs/**`
- `api/internal/platform/metrics/**`
- `api/internal/modules/invites/**`
- `api/internal/modules/passwordreset/**`
- `api/internal/modules/funding/**`
- `api/internal/modules/invoicing/**`
- `api/internal/modules/payments/**`
- `api/internal/modules/stripewebhook/**`
- `api/internal/app/bootstrap/bootstrap.go`
- `api/internal/app/bootstrap/adapters.go`
- `api/internal/platform/config/config.go`
- `api/internal/platform/http/error_mapper.go`
- `api/.env.example`
- `api/Dockerfile`
- `docker-compose.prod.yml` or project-approved equivalent
- `docs/API-CONTRACT-MVP.md`
- `docs/PILOT-BACKEND-RUNBOOK.md`
- `docs/STRIPE-WEBHOOK-RUNBOOK.md`
- `docs/BACKUP-RESTORE-RUNBOOK.md`

Existing hand-written repositories may remain in place unless a task explicitly touches them.

## Verification Steps

- From `api/`, run `go test ./...`.
- From `api/`, run `sqlc generate` after query changes.
- Run migrations on a clean local database:
  - `migrate -path db/migrations -database "$DATABASE_URL" up`
  - `migrate -path db/migrations -database "$DATABASE_URL" down`
  - `migrate -path db/migrations -database "$DATABASE_URL" up`
- Smoke-test auth:
  - seeded manager login
  - refresh rotation
  - logout idempotency
  - password reset valid/expired/used token
  - invite accept valid/expired/revoked token
- Smoke-test authorization:
  - practitioner can check in/check out but cannot access funding/invoices
  - parent can see only linked issued invoices
  - manager can generate and issue invoices
  - wrong tenant/branch access is forbidden
- Smoke-test billing:
  - complete attendance produces billable rounded minutes
  - incomplete attendance appears in invoice-run exceptions
  - funded-hours allowance reduces only core childcare hours
  - draft regeneration does not duplicate invoices
  - issued invoice cannot be edited/regenerated
- Smoke-test Stripe:
  - issued invoice creates hosted Checkout session
  - failed/canceled event sets `payment_failed`
  - duplicate success webhook is idempotent
  - paid invoice cannot create another checkout session
- Verify backend operations:
  - production env variables are documented
  - Docker Compose config references required services and secrets
  - overdue scheduler can be enabled/disabled by env
  - backup and restore checklist has been rehearsed locally

## Explicit Assumptions

- The current `api/` code is the backend application for the MVP.
- The frontend will integrate with `/api/v1` plain JSON endpoints and standard error shape.
- Local development uses local PostgreSQL; Docker is not required for local API development.
- Production deployment uses a single VM with Docker Compose.
- Month 1 uses one pilot tenant and one default branch, but API scope enforcement must remain tenant/branch-aware.
- The first manager is still created by seed/admin command, not by invite.
- Manager-created invites can provision `practitioner` and `parent` roles only.
- Email delivery can be implemented through one provider abstraction with SMTP sandbox locally.
- Stripe hosted Checkout is the only month-1 payment UI.
- Currency is GBP and tax mode is non-VAT for month 1.
- Extras are manual invoice line items and are never reduced by funded-hours deduction.
- Reporting and CSV export are not required to meet the core pilot success metric.
