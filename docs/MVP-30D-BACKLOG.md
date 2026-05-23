# MVP 30-Day Execution Backlog

Source: `docs/PRD-MVP-1M.md`

## Delivery Rules

- Scope lock: no new modules beyond attendance, funding v1, invoicing, Stripe, parent billing view.
- Build vertical slices early and demo every week.
- Every feature must include basic audit logging and role checks.
- API style: `/api/v1` routes with plain JSON and HTTP status codes.
- Persistence: `sqlc + pgx`, no ORM.
- Done each day means: code merged locally, smoke-tested manually, notes updated.

## Definition of Done (Global)

- `manager`, `practitioner`, and `parent` role boundaries enforced.
- Attendance and invoice edits captured in `audit_logs` with actor and timestamp.
- Stripe webhook flow is idempotent and updates invoice status reliably.
- Pilot nursery can run one real monthly invoice cycle without spreadsheets.
- Tenant and branch scoping enforced on all read/write operations.

## Day-by-Day Plan

## Week 1 - Foundations and Core Data

| Day | Focus | Tasks | Done check |
|---|---|---|---|
| 1 | Project bootstrap | Initialize Gin API skeleton, `/api/v1` router, env config, local Postgres connection, golang-migrate wiring | Done (health endpoints respond) |
| 2 | Auth base | Login/logout, JWT access (15m), refresh token store (30d), password hash and verify | Done (seeded user login/refresh/logout verified) |
| 3 | Roles and guards | Authorization guards middleware (role, scope, relationship) + scoped query enforcement | Done (protected routes reject wrong role/scope/relationship with matrix tests) |
| 4 | Core schema | Create tables: users, memberships, children, guardians, links, sessions, audit_logs | Done (migrations up/down/up + seed scenarios) |
| 5 | Child management API | Manager CRUD for child and guardian + relationship linking | Done (manager child/guardian APIs + relationship lifecycle actions wired) |
| 6 | Staff UI basics | Manager/practitioner screens; manager edit only, practitioner read only | Permissions match policy in UI/API |
| 7 | Week 1 gate | Hardening + smoke test + backup check + demo prep | Week 1 demo completed |

## Week 2 - Attendance and Funding v1

| Day | Focus | Tasks | Done check |
|---|---|---|---|
| 8 | Attendance schema/API | Create event model (`check_in`,`check_out`,`correction`) + endpoints | Done (check-in/check-out API + attendance daily list endpoint) |
| 9 | Attendance validations | Prevent duplicate active check-in, enforce correction reason, timezone policy | Done (duplicate session blocked, enrollment gate, correction reason contracts, clock injection) |
| 10 | Attendance UI | Practitioner flow: today list, check-in, check-out | Done (operational list with check-in/check-out, search, filters, row errors, enrollment gating) |
| 11 | Attendance corrections | Manager-only correction endpoint + audit event + incomplete session flagging | Correction stores reason and actor |
| 12 | Funding schema | Create funding_profiles + monthly allowance settings per child | Manager can save allowance |
| 13 | Funding calc v1 | Implement formula + 15-min round-up + core-only deduction + extras payable | Inputs produce expected net due |
| 14 | Week 2 gate | E2E dry run: attendance -> monthly totals -> funding result | Demo shows full operational chain |

## Week 3 - Invoicing and Stripe

| Day | Focus | Tasks | Done check |
|---|---|---|---|
| 15 | Invoice schema | Create invoices, invoice_lines, payments tables | Migration and seed data valid |
| 16 | Invoice generation | Manual month draft generation from attendance actuals + exception list | Drafts generated for eligible children |
| 17 | Invoice lifecycle | Issue one/bulk, immutable issued invoices, status rules (`payment_failed`, `overdue`) | Status transitions are enforced |
| 18 | Parent invoice view | Parent portal page for current and past invoices | Parent sees only own invoices |
| 19 | Stripe checkout | Create Stripe checkout/payment-link flow from issued invoice | Parent can open Stripe payment page |
| 20 | Stripe webhook | Verify signature, process events, idempotent event store, retry processing | Paid event updates invoice once |
| 21 | Week 3 gate | End-to-end billing flow demo + failed payment scenario | Success and failure flows both pass |

## Week 4 - Hardening and Pilot Go-Live

| Day | Focus | Tasks | Done check |
|---|---|---|---|
| 22 | Reconciliation view | Manager screen for payment status, webhook history, retry checkout action | Manager can audit payment events |
| 23 | Reporting basics | Add minimal dashboard: unpaid invoices, today attendance, monthly totals | Dashboard loads with live data |
| 24 | CSV export | Export invoices and payments for manager | CSV matches UI totals |
| 25 | Security pass | Secret handling, HTTPS config, role/access review, basic rate limits | No critical auth/access gaps |
| 26 | Data safety pass | Daily backup job and restore test checklist | Backup and restore steps verified |
| 27 | UAT script prep | Write pilot test script for manager/practitioner/parent journeys | UAT checklist approved by owner |
| 28 | Pilot UAT day | Run real user testing in nursery with fixes list | Critical blockers identified |
| 29 | Fix day | Resolve critical/high UAT issues only, freeze scope | All critical issues closed |
| 30 | Go-live | Deploy on single VM via Docker Compose, monitoring checks, first invoice cycle support | Pilot live and operational |

## Weekly Demo Checklist

- Week 1: auth + roles + child/guardian CRUD
- Week 2: daily attendance + correction audit + funding v1 preview
- Week 3: invoice issue + Stripe payment + webhook reconciliation
- Week 4: production readiness + UAT signoff + go-live

## Backlog Parking Lot (Do Not Start in Month 1)

- Incident/safeguarding workflows
- Ratio engine and alerts
- Messaging and notifications
- Learning journal
- SEND features
- Multi-branch support
- Advanced local authority funding rule packs

## Operational Runbook (Day 30 to Day 37)

- Daily check Stripe webhook failures and retry queue.
- Daily check unpaid invoices and parent payment completion.
- Track defects in a single prioritized list: critical/high/medium.
- Do not add new features during first pilot week.
