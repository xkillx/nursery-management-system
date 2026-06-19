# Implementation Plan: Booking-Based Advance-Pay Invoicing (12-Month Term)

## 1. Summary

Replace the MVP's attendance-actuals-based monthly invoicing with a 12-month fixed-term advance-pay model driven by the child's `Booking Pattern`. Each child is enrolled onto a 12-month term that fixes a weekly schedule of session types. The system generates and issues one invoice per child per calendar month, in advance, on the 25th of the prior month, due on the 1st of the billing month. In-term schedule changes are bounded by a 1-month notice. The term ends hard at 12 months and must be explicitly renewed. Funded-hours deduction still applies, but to booked core minutes rather than attended minutes. Attendance is still captured for child safety and operational reporting, but is **not** a billing input.

This is a **replacement**, not a coexistence model: there is no production user base, the MVP can be rewritten cleanly, and there is no migration of historical invoices. The existing `attendance`-to-`billing` wiring is removed; the existing `booking pattern` data model and `session_types` module are reused as the contract.

## 2. Product Decisions (recap)

| # | Decision | Choice |
|---|---|---|
| 1 | Scope of change | **Replacement** (no coexistence, no migration of old invoices) |
| 2 | Commitment shape | **12-month fixed term**, monthly invoiced in advance |
| 3 | Booking unit | **Specific weekly schedule** — reuses existing `Booking Pattern` (day-of-week × `SessionType`) |
| 4 | Attendance role | **Captured for operations only; no billing effect** (no refund for under-attendance, no auto-charge for over-attendance) |
| 5 | Invoice generation cadence | **Rolling monthly, system-generated, auto-issued** (manager does not manually issue each month) |
| 6 | Invoice issue date | **25th of the prior month** |
| 7 | Invoice due date | **1st of the billing month** |
| 8 | Overdue policy | **1-week grace**: invoice transitions to `overdue` on the 8th; child's attendance may be blocked from the 8th; manager can override the block per-child |
| 9 | In-term schedule adjustment | **1 calendar month notice**; decreases are free; increases require manager approval (room capacity + staff:child ratios) |
| 10 | Early termination | **1 calendar month notice**; parent pays for the notice month only; no further liability; no refunds of future pre-generated drafts (none will exist for the cancelled term going forward) |
| 11 | End of term | **Hard end at `term_end_date`**. **T-45 soft warning** to manager; **T-30 renewal prompt** to manager; parent must explicitly sign a new 12-month term to continue; otherwise the child becomes `inactive` at `term_end_date` |
| 12 | Funded-hours deduction | **Apply to booked core hours** (not attended hours). Existing `Funding Profile` semantics are preserved; only the upstream input changes |
| 13 | Term start | **`term_start_date` is the agreed formal start**; settling-in is the gap (0+ days) between `enrollment_start_date` and `term_start_date` |
| 14 | Settling-in | **Free**, no invoice; attendance is captured for operational/registers only |
| 15 | Term creation flow | **Combined with child creation**; term is required at child creation (term details are part of the existing `Child Management Atomic Create` transaction) |
| 16 | `term_start_date` constraint | **Always the 1st of a calendar month**; no mid-month starts, no pro-rata math |
| 17 | Nursery holidays / closure weeks | **Always bill the full booked pattern for 52 weeks**; no closure calendar; no pro-rating |
| 18 | Extras | **Manual invoice line items added by managers** (unchanged from current `Extras Charging Model`) |
| 19 | Parent signoff | **None in the system**; the manager records the term; the parent sees it reflected on the child detail and the first invoice |
| 20 | Term auditability | **Auditable** (actor, time, term contents, schedule changes, renewals, terminations) — under existing `Audit Baseline Scope` |
| 21 | Invoice numbering | **Unchanged**: `INV-YYYYMM-####` where `YYYYMM` is the billing month |

## 3. Glossary Updates (CONTEXT.md)

### 3.1 New terms to add

- **Term** — A 12-month commercial commitment between a nursery site and a parent for one child, fixing a weekly schedule of session types and an end date. One child has at most one active Term at a time; historical Terms are retained.
- **`term_start_date`** — The first day of a Term; always the 1st of a calendar month.
- **`term_end_date`** — The last day of a Term; equals `term_start_date + 12 months − 1 day`.
- **Term Status** — Derived from dates and lifecycle actions: `pre_term` (created, `term_start_date` is in the future, settling-in window), `active` (`term_start_date` ≤ today ≤ `term_end_date`), `pending_renewal` (within 30 days of `term_end_date`, manager has not yet created a renewal), `ended` (term reached `term_end_date` with no renewal recorded), `terminated` (ended early via notice).
- **Settling-in Window** — The optional period between a child's `enrollment_start_date` and their first `term_start_date`; attendance is captured, no invoice is generated.
- **Renewal Prompt** — A T-30 calendar-day manager-facing notification that a Term is approaching its `term_end_date` and a renewal decision is required.
- **Renewal Soft Warning** — A T-45 calendar-day earlier heads-up to the manager, ahead of the formal renewal prompt.
- **Advance-Pay Invoice** — A monthly invoice generated and issued by the system on the 25th of the prior month, due on the 1st of the billing month.
- **In-Term Schedule Adjustment** — A child-level booking-pattern change inside an active Term, subject to the 1-month-notice rule, with separate handling for decreases and increases.
- **Notice Month (Termination)** — The single calendar month the parent is liable for when they terminate a Term with 1 month notice; the child may leave earlier in that month but the parent pays the full month.

### 3.2 Existing terms to supersede (replace with new wording or remove)

- **Invoice Source of Truth** — *was*: "Monthly invoice billable minutes are derived from attendance actuals." **Replace with**: "Monthly invoice billable core minutes are derived from the active Term's Booking Pattern for the billing month. Funded-hours deduction is applied to those booked minutes. Attendance is not a billing input."
- **Booking Pattern Billing Boundary** — *was*: "Booking patterns record expected attendance only and do not drive billing..." **Replace with**: "Booking Patterns are the source of booked core minutes for monthly invoicing in this release. A Booking Pattern attached to an active Term drives the monthly invoice's core line; a Booking Pattern outside any Term has no billing effect."
- **Mid-Month Leave Billing** — **Retire** (replaced by the 1-month notice + notice-month policy).
- **Post-Leave Invoice Artifact Policy** — **Retire** (no longer relevant under advance-pay; future drafts stop on termination).
- **Invoice Attendance Source Snapshot** — **Retire** (no attendance snapshot on issued invoices under the booking-driven model).
- **Invoice Explainability Persistence** — *was*: "Invoice line storage preserves both intermediate billing components (core attended minutes, funded deduction minutes, core billable minutes, hourly rate) and final totals." **Replace with**: "Invoice line storage preserves intermediate billing components (booked core minutes per session, funded deduction minutes, core billable minutes, site hourly rate at term creation) and final totals."
- **Invoice Due Policy** — *was*: "Invoices are due on receipt when issued." **Replace with**: "Invoices are issued on the 25th of the prior month and due on the 1st of the billing month."
- **Invoice Overdue Transition** — *was*: "An unpaid issued invoice transitions to overdue at 00:00 the next local day in Europe/London." **Replace with**: "An unpaid issued invoice transitions to overdue at 00:00 on the 8th of the billing month in Europe/London (1-week grace). A payment_failed invoice does not transition to overdue."
- **Draft Invoice Calculation Lines** — Keep, but reword: "Generated draft monthly invoices use consistent explanatory lines for core childcare (per session, summed to monthly booked core minutes × site hourly rate) and funded-hours deduction, including a zero-value funded deduction line when no deduction amount is applied."
- **Zero-Attendance Invoice Eligibility** — **Retire** (no longer meaningful; eligibility is about booking coverage, not attendance completeness).
- **Incomplete Attendance Handling** — *was*: about excluding from billing. **Replace with**: "Attendance records missing check-out remain visible on operational triage surfaces for manager action; they no longer affect billing."
- **Incomplete Attendance Triage** — Keep, but reframe as operational-only (no longer a billing readiness step).
- **Attendance Correction Scope** — Keep the operational scope; explicitly note that corrections do not move money.
- **Cross-Month Session Allocation** — **Retire** (allocation was attendance-driven; the new model books whole months).
- **Child Funding Record** — Keep as-is; still distinct from `Funding Profile`.
- **Child Management Atomic Create** — *was*: "Manager creates a child in a single transaction that includes identity, profile, contacts, health, safeguarding, consent, funding, collection settings, room placement, and billing profile." **Replace with**: same list plus **"and the child's first Term"**. The first Term is created in the same transaction as the child.

## 4. ADR Candidates

Two ADRs are warranted (all three ADR criteria — hard to reverse, surprising without context, real trade-off — apply to each).

1. **`docs/adr/0006-booking-pattern-billing-source.md`** — Supersedes ADR-0005. States that Booking Patterns are now the invoice source of truth, replacing attendance actuals. Captures the trade-off: bookings are a parent commitment and don't refund for under-attendance, but they do match UK nursery commercial reality and make the advance-pay model coherent.

2. **`docs/adr/0007-12-month-fixed-term-contract.md`** — New. Captures the Term entity and its lifecycle (12-month fixed term, hard end, 1-month notice, 30-day renewal prompt, advance-pay monthly invoicing). Records the trade-off: 12-month terms are stricter than rolling subscriptions, but match UK nursery norm and make advance invoicing tractable.

## 5. Domain Model

### 5.1 New entity: `term`

```
term
  id                uuid PK
  tenant_id         uuid FK -> tenants
  branch_id         uuid FK -> branches
  child_id          uuid FK -> children
  term_start_date   date NOT NULL              -- always 1st of a month
  term_end_date     date NOT NULL              -- term_start_date + 12 months - 1 day
  booking_pattern_id uuid FK -> child_booking_patterns NOT NULL
  site_hourly_rate_minor int NOT NULL         -- rate snapshotted at term creation (minor units, GBP)
  status            text NOT NULL              -- pre_term | active | pending_renewal | ended | terminated
  termination_reason_code text NULL           -- lifecycle reason code (per existing vocabulary) when terminated
  termination_reason_note  text NULL
  terminated_at     timestamptz NULL
  created_at        timestamptz NOT NULL
  created_by_membership_id uuid NOT NULL      -- audit
  updated_at        timestamptz NOT NULL
  UNIQUE (child_id) WHERE status IN ('pre_term','active','pending_renewal')   -- at most one non-historical Term per child
  CHECK (term_start_date = date_trunc('month', term_start_date))
  CHECK (term_end_date   = (term_start_date + interval '12 months' - interval '1 day')::date)
```

Historical (`ended`, `terminated`) Terms are retained for audit. The unique partial index guarantees the "at most one active Term per child" invariant.

### 5.2 New entity: `term_schedule_change` (audit trail for in-term adjustments)

```
term_schedule_change
  id                uuid PK
  term_id           uuid FK -> term
  previous_booking_pattern_id uuid FK -> child_booking_patterns NOT NULL
  new_booking_pattern_id      uuid FK -> child_booking_patterns NOT NULL
  change_kind       text NOT NULL              -- 'decrease' | 'increase'
  requested_at      timestamptz NOT NULL
  effective_from    date NOT NULL              -- 1st of the month, >= today + 1 month
  approved_by_membership_id uuid NULL          -- NULL for decreases (auto-approved); required for increases
  approval_decision text NULL                  -- 'approved' | 'rejected' | NULL
  created_at        timestamptz NOT NULL
  request_id        text NOT NULL              -- request correlation
```

### 5.3 New entity: `invoice_run_advance` (the scheduled monthly generation log)

```
invoice_run_advance
  id                uuid PK
  tenant_id         uuid
  branch_id         uuid
  billing_month     date NOT NULL              -- 1st of the month
  generated_at      timestamptz NOT NULL
  generated_invoice_count   int NOT NULL
  skipped_term_count        int NOT NULL       -- terms that hadn't started or were already terminated
  exception_count   int NOT NULL               -- e.g., missing funding profile
  triggered_by      text NOT NULL              -- 'scheduler' | 'manager_regenerate'
  request_id        text NULL
```

### 5.4 Entities to modify

- **`invoices`** — no schema change required. The data shape stays the same; the calculation inputs change (booking pattern minutes instead of attended minutes). Drop the `attendance_source_sessions` snapshot column if it exists, or stop populating it.
- **`invoice_lines`** — keep the explainability line shape. The "core minutes" line now stores the per-session booked minutes (snapshot), summed to the monthly total.
- **`funding_profiles`** — no schema change. Still per-child per-billing-month minutes. The funded deduction is now applied to booked core minutes for the month.
- **`child_booking_patterns`** — no schema change. Reuse the existing table. The pattern referenced by a Term must be effective-dated to span (or be open past) the Term's `term_end_date`.
- **`children`** — add `current_term_id` denormalisation for fast lookups (optional, but consistent with existing denormalisation patterns).
- **`attendance_*` tables** — no schema change. Stop wiring them into billing.

## 6. Module Structure

### 6.1 New module: `api/internal/modules/term/`

Standard clean-architecture layout: `domain/`, `application/`, `infrastructure/postgres/`, `interfaces/http/`.

**Domain**
- `term.go` — `Term` entity, `TermStatus` enum, invariants (dates, one active per child), constructor functions.
- `repository.go` — interface: `Create`, `GetByID`, `GetActiveForChild`, `List`, `UpdateStatus`, `Terminate`, `ListExpiringWithin`.
- `errors.go` — `ErrTermAlreadyExists`, `ErrTermNotActive`, `ErrInvalidDateRange`, `ErrInvalidStartDay`, `ErrTermOverlap`, etc.

**Application** (one use case per file)
- `create_term.go` — invoked as part of child atomic create.
- `get_term.go`, `list_terms.go`, `list_expiring.go` — manager reads.
- `request_schedule_change.go` — initiate a decrease or increase; effective-from is computed from the 1-month-notice rule.
- `approve_schedule_change.go` — manager approval for increases only.
- `reject_schedule_change.go` — manager rejection for increases only.
- `terminate_term.go` — early termination with notice; sets `status=terminated`, records reason code/note.
- `mark_pending_renewal.go` — invoked by the daily scheduler; flips status when within 30 days of `term_end_date`.
- `mark_ended.go` — invoked by the daily scheduler on the day after `term_end_date` if no renewal term exists.

**Infrastructure (postgres)**
- `term_repository.go` — sqlc-backed implementation.
- `term_schedule_change_repository.go`.

**Interfaces (http)**
- `term_handler.go` — REST routes: `POST /api/v1/children/:id/terms`, `GET /api/v1/children/:id/terms`, `GET /api/v1/children/:id/terms/current`, `POST /api/v1/terms/:id/schedule-changes`, `POST /api/v1/terms/:id/schedule-changes/:changeId/approve|reject`, `POST /api/v1/terms/:id/actions/terminate`, `GET /api/v1/terms?expiring_within_days=30`.
- `dto.go` — request/response shapes.

### 6.2 New module: `api/internal/modules/invoicerun/`

Hosts the scheduled monthly invoice generation use case. Splitting this out of `billing` keeps the `billing` module focused on the live invoice lifecycle.

**Application**
- `generate_advance_invoices.go` — for a given billing month (default = next month from "today"), iterate active Terms in the branch, build one draft invoice per term per the calculation rule below, run preflight (including `Missing Funding Profile Invoice Block`), persist drafts in a transaction, then issue them in the same transaction so they are immediately parent-visible and payable. Skip (and log as an exception) any term whose `term_start_date > billing_month_end` or `term_end_date < billing_month_start`. Idempotent: a second run for the same billing month is a no-op (the unique constraint on `(child_id, billing_month)` on `invoices` already enforces this, but the run also checks and short-circuits).
- `mark_overdue_advance_invoices.go` — daily scheduler; transitions `issued` invoices to `overdue` where due_date < today and the grace has passed (today >= due_date + 7 days) and state is still `issued`.
- `expire_terms.go` — daily scheduler; flips terms to `pending_renewal` at T-30, to `ended` at `term_end_date + 1 day` if no renewal exists, and writes the renewal prompt + soft warning at T-45 / T-30.

### 6.3 Modified module: `api/internal/modules/billing/`

- `application/generate_draft_invoices.go` — **replaced** with a thin wrapper around `invoicerun.generate_advance_invoices.go` for the rare manager-triggered "regenerate now" path. The previous "attendance actuals → billable minutes" logic is removed.
- `application/issue_invoices.go` — **replaced** with auto-issue. Bulk issue and single-issue manager flows still exist for the "regenerate and re-issue" path but the default flow is auto-issue on the 25th.
- `application/preflight_draft_invoices.go` — **rewritten** to preflight on Term coverage for the month (active Term whose `term_start_date ≤ month_end` and `term_end_date ≥ month_start`) and `Funding Profile` presence, not on attendance.
- `application/mark_overdue_invoices.go` — **rewritten** to use the 8th-of-month rule (1-week grace).
- `application/calculate_line_minutes.go` — **new**. The per-month calculation:
  - `booked_core_minutes_in_month(child, billing_month) = Σ over sessions of the active Booking Pattern s.t. session overlaps billing_month: session.duration_minutes × count_of_occurrences_in_month(billing_month, s.day_of_week, s.session_type)`.
  - `core_due_minor = ceil(booked_core_minutes_in_month / 60 × site_hourly_rate_minor)`.
  - `funded_deduction_minor = min(funded_allowance_minutes_in_month, booked_core_minutes_in_month) / 60 × site_hourly_rate_minor` (rounded per existing `Core Billing Price Rounding`).
  - `core_billable_minor = max(0, core_due_minor − funded_deduction_minor)`.
  - Plus manual `extras` lines (existing `Extras Charging Model`).
- `infrastructure/postgres/` — the `invoices` and `invoice_lines` repos stay; the calculation input queries are rewritten to read from the Term + Booking Pattern instead of attendance sessions.

### 6.4 Modified module: `api/internal/modules/funding/`

- `application/` — no semantic change to funded-hours deduction; only the upstream call changes from "give me the attended core minutes for this child-month" to "give me the booked core minutes for this child-month" (the new billing module passes the booked minutes in). The preflight `Missing Funding Profile Invoice Block` is preserved.

### 6.5 Modified module: `api/internal/modules/attendance/`

- `application/`, `infrastructure/` — no schema change, no functional change. **The wiring into `billing` is removed** by deleting the cross-module call. `correct_attendance.go` and `check_in_child.go` continue to write attendance events and update session effective intervals; the absence of these from billing is now correct, not a bug. Operational triage (`Incomplete Attendance Triage`) remains but is reframed as operational-only (no "blocks invoice" warning).

### 6.6 Modified module: `api/internal/modules/children/`

- `application/create_child.go` — the atomic create transaction is extended to also create the child's first `Term` (with `term_start_date`, `booking_pattern_id`, etc.) and any required `Booking Pattern` rows. If the manager opts for a settling-in gap, `enrollment_start_date < term_start_date` is allowed; both are stored on the child/term.
- The `Child Management Atomic Create` glossary entry is updated (see §3.2).

### 6.7 Modified: `api/internal/app/bootstrap/`

- `bootstrap.go` — wire the new `term` and `invoicerun` modules; register the new HTTP routes; register the scheduled job runner for the three daily jobs (`generate_advance_invoices`, `mark_overdue_advance_invoices`, `expire_terms`).
- `adapters.go` — no new cross-module adapter needed; the term module reads from `sessiontypes` and `children` via interfaces defined in the term domain (Clean Architecture rule: consumer owns the interface).

## 7. Database Migrations

### 7.1 New migration: `000003_terms.up.sql` / `000003_terms.down.sql`

Create `term`, `term_schedule_change`, `invoice_run_advance` per §5. Add `current_term_id` denormalisation column on `children` (nullable, FK to `term.id`). Add unique partial index enforcing "at most one active Term per child". Add the `CHECK` constraints on `term_start_date` and `term_end_date`. Backfill: not required (no production data — this is a clean replacement; the migration runs against an empty or about-to-be-wiped dev DB).

### 7.2 No changes to existing migrations

Migration `000001` (baseline) and `000002` (session types + booking patterns) stay as-is. The new model builds on `child_booking_patterns` from `000002`.

## 8. Phased Delivery

Each phase ends with all related tests green (`make test-api-repositories`, module-level unit tests, handler tests), `golangci-lint run ./...` clean, and `go test ./...` green.

### Phase 1 — Term domain, application, infrastructure, HTTP (no billing wiring yet)
- Migration `000003` up + down.
- `modules/term/` complete: domain entity, repository interface, use cases, postgres repo, HTTP handlers, DTOs.
- Manager can create a Term for an existing child, list terms, view a current term, request a schedule change, approve/reject (increases only), terminate with reason.
- Unit tests for domain invariants and use cases.
- Repository tests against `TEST_DATABASE_URL`.
- Handler tests with `gin` context and the existing `tenant.ActorFromGinContext` auth.
- Wire the new routes in `bootstrap.go`.
- **Done when**: a manager can run the full Term CRUD + schedule-change + terminate lifecycle against a real Postgres test DB, all tests pass.

### Phase 2 — Replace billing calculation: booking-driven core minutes
- `modules/billing/application/calculate_line_minutes.go` — new calculation per §6.3.
- `modules/billing/application/generate_draft_invoices.go` — rewritten to use the new calculator.
- `modules/billing/application/preflight_draft_invoices.go` — rewritten to preflight on Term coverage + Funding Profile.
- `modules/billing/infrastructure/postgres/` — replace attendance-reading queries with Term/Booking Pattern queries.
- Delete the cross-module call from `modules/attendance/` into `modules/billing/`.
- Keep `modules/attendance/` functionally unchanged (writes still happen; no billing effect).
- Unit tests: per-child monthly calculation, funding deduction edge cases (booked < funded, booked > funded, missing funding profile).
- Repository tests: end-to-end draft generation for a billing month against a known set of terms + booking patterns.
- **Done when**: a draft invoice for an active Term is generated, with correct `booked_core_minutes`, `funded_deduction_minutes`, `core_billable_minutes`, and line items — verified by repo test against a fixture of 2-3 child/term/pattern combinations. Attendance-derived input is not read anywhere in billing.

### Phase 3 — Issue, due, overdue, attendance block
- `modules/billing/application/issue_invoices.go` — auto-issue on generation.
- `modules/billing/application/mark_overdue_invoices.go` — 8th-of-month rule.
- New HTTP route: `POST /api/v1/billing/invoices/:id/override-attendance-block` — manager override per child for the current billing month.
- Update `invoices.due_date` semantics: `due_date = billing_month_start` (was: `issue_time`).
- `invoices.overdue_at` rule: `overdue_at = billing_month_start + 7 days` (was: `issue_time + 1 day`).
- Attendance block: practitioner daily-list surfaces an "attendance blocked — invoice overdue" warning when the child's billing month is past the 8th and the invoice is unpaid; manager override hides the warning.
- Unit + handler tests.
- **Done when**: an issued invoice is due on the 1st, becomes overdue on the 8th, and the attendance block is applied; manager override clears it.

### Phase 4 — Scheduler + invoicerun module
- New module `modules/invoicerun/` (§6.2).
- Scheduled jobs (cron or equivalent background runner) at:
  - **Daily 02:00 Europe/London** — `expire_terms` (mark pending_renewal, mark ended, write soft-warning + renewal-prompt audit events).
  - **Daily 02:00 Europe/London** — `mark_overdue_advance_invoices` (8th-of-month check).
  - **Monthly 25th 00:05 Europe/London** — `generate_advance_invoices` (one per active Term in the branch, for the next billing month). **Idempotent**: re-running for the same month is a no-op.
- The 25th and 8th run on the calendar day; if the 25th or 8th is a Saturday/Sunday/bank holiday, the run still executes on the calendar day (no business-day shifting) — keep it simple, predictable, and aligned with `INV-YYYYMM-####` semantics.
- The 1st-of-month "due" is also a calendar day, no shifting.
- Logging + metrics per existing `Observability Baseline`.
- **Done when**: with a clock set to 25th of any month, running the scheduler produces the correct set of advance invoices for the following month; with the clock set to the 8th, the correct invoices transition to `overdue`.

### Phase 5 — Renewal flow
- Manager-facing `GET /api/v1/terms?expiring_within_days=30` already shipped in Phase 1.
- Add a manager UI surface (web) showing terms in `pending_renewal` with a one-click "Create renewal term" action.
- "Create renewal term" = `POST /api/v1/children/:id/terms` with `term_start_date = previous_term_end_date + 1 day`. The new term is `pre_term` (no settling-in for back-to-back renewals; both terms are full 12-month terms).
- If no renewal is recorded by `term_end_date + 1 day`, `expire_terms` flips the child to `inactive` (the existing `Child Enrollment Lifecycle` flow).
- **Done when**: a child whose term ends without renewal transitions to `inactive` on `term_end_date + 1`; a renewed child has a new active Term starting the day after the previous one.

### Phase 6 — Glossary + ADR + docs
- Update `CONTEXT.md` per §3.1 and §3.2.
- Write `docs/adr/0006-booking-pattern-billing-source.md` (supersedes ADR-0005).
- Write `docs/adr/0007-12-month-fixed-term-contract.md`.
- Update `docs/POST-MVP-ROADMAP.md` (or successor) if it exists; the Booking Pattern entry is no longer "capture-only".
- Update `docs/API-CONTRACT.openapi.yaml` for new term + schedule-change + advance-invoice-run endpoints.

## 9. Acceptance Criteria

- A child can be created with a Term in a single atomic transaction. The child has `enrollment_complete` AND `has_active_term` immediately after creation (or, if `term_start_date` is in the future, after that date).
- With a child whose Term is active, the scheduled run on the 25th of the prior month produces exactly one `issued` invoice for the upcoming month. The invoice's `due_date` is the 1st of the billing month. The `core_minutes` line equals `Σ over the active Booking Pattern's sessions in the month: session.duration_minutes × count_of_occurrences_in_month`. The `funded_deduction_minutes` line is `min(funding_profile.allowance_minutes_in_month, core_minutes)` and the `core_billable_minor` reflects that deduction.
- On the 8th of the billing month, an unpaid issued invoice transitions to `overdue`. A `payment_failed` invoice does not.
- An in-term schedule decrease requested on day X (1st ≤ X ≤ 25th) takes effect for the invoice generated for month X+2 (so the parent sees one full unchanged month of invoice after requesting). A schedule increase requires manager approval and is rejected unless the manager approves it.
- Terminating a Term with 1 month notice sets `status=terminated`, ends the child-month's billing, and produces no invoices for any month after the notice month. The parent pays for the notice month (full month) only.
- A Term reaching `term_end_date` without a renewal is flipped to `ended`; the child is flipped to `inactive` the following day. A renewed child has a new active Term with `term_start_date = previous.term_end_date + 1 day` and `term_end_date = new.term_start_date + 12 months − 1 day`.
- Attendance events are still captured by practitioners, and a child missing check-out is still surfaced on the daily register. **No** attendance event changes an invoice.
- 52 weeks of the year are billed on the booked pattern, including weeks the nursery is closed.
- Audit events are persisted for: Term creation, schedule change request, schedule change approval/rejection, Term termination, renewal creation, scheduler-driven status transitions.

## 10. Open Implementation Decisions (agent-owned)

These are not product questions; the implementing agent decides the cleanest local option:

- **Scheduler implementation**: existing background job runner vs. a new one. If a cron-style runner already exists in the platform, use it; otherwise introduce the simplest one (a `time.Ticker` in a goroutine started by `lifecycle.Run` is acceptable for the pilot).
- **Audit event names**: new event types (`term_created`, `term_schedule_change_requested`, `term_schedule_change_approved`, `term_schedule_change_rejected`, `term_terminated`, `term_renewal_created`, `term_status_transitioned`) following existing audit event naming conventions in the codebase.
- **`current_term_id` denormalisation**: optional; include only if the read pattern justifies it.
- **Invoice number assignment** on auto-issue: keep the existing `INV-YYYYMM-####` and `Bulk Invoice Issue Sequence Order` (deterministic child-name order) — this works for auto-issue with no manager selection.
- **Manager override of attendance block** storage: a per-child-per-month record on the child or on the invoice; pick whichever is simpler given the existing data model. The semantics is "for this child, for this billing month, the attendance block is cleared even if the invoice is overdue."
- **Cancellation of the first pre-generated invoice if the term is terminated before the month starts**: not possible under the rolling generation rule (no drafts exist yet for the notice month), so nothing to do.
- **Manager UX for the renewal flow**: list page, form, and renewal prompt badge follow existing manager UI patterns.
- **Frontend changes**: add the Term management surfaces to the manager web app; the parent portal continues to show invoices only (no term/schedule view in the first release).
- **Test data wipe**: since this is a clean replacement with no production users, the dev DB can be wiped before Phase 1. Document this in the PR description.
- **Existing module test files**: many of the existing billing test files assert attendance-driven behaviour. These tests will need to be updated or replaced as part of Phase 2; the agent updates them in the same commits as the calculation rewrite.

## 11. Out of Scope (for this plan)

- Mid-month Term starts (already ruled out: `term_start_date` is always the 1st).
- Pro-rated first / last month invoices (no pro-rata math under advance-pay).
- Sibling discount (per existing `Sibling Discount Policy` — deferred).
- Per-child bespoke rates, discounts, or session exceptions (per existing `Child-Specific Core Rate Boundary`).
- Nursery closure calendar with pro-rating (per Q14: always bill 52 weeks).
- Parent-side e-signature of the Term (manager records the term; parent sees it on the child detail and the first invoice).
- Refunds for under-attendance or credits for nursery closures (per Q4 and Q14).
- Auto-charge for over-attendance (per Q4).
- Adjustments to issued invoices for "missed sessions" (per Q4: irrelevant in the booking model; adjustments remain a manager-driven flow for genuine billing errors, per existing `Adjustment Flow`).
- Tax / VAT (per existing `Tax Handling`).
- Multi-currency (per existing `Invoice Currency`).
- Day-of-month edge cases for the 25th and 8th falling on a weekend/bank holiday (the run executes on the calendar day; no business-day shifting).
- Migration of any existing data, including the existing `funding_profiles`, `invoices`, `attendance_*` rows in the dev DB (clean wipe; no production users).
- New product surface for "scheduled vs actual" attendance reporting to managers or parents (operational reporting is out of scope for this plan; the data is captured and retained, the surface can come later).
