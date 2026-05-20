# Decision Baseline - MVP Month 1

This document consolidates locked implementation decisions for the 30-day pilot MVP.
If a future task conflicts with this baseline, treat this file as the source of truth unless a new explicit decision supersedes it.

## 1. Outcome and Scope Lock

- Pilot target: 1 live UK nursery, single site.
- Success metric: nursery can run daily attendance and monthly invoicing without spreadsheets.
- In scope: attendance, funding v1, invoicing, Stripe payments, parent invoice view.
- Out of scope: safeguarding, ratio engine, messaging, learning journal, SEND, payroll/HR/rotas, advanced funding packs.
- Change rule: only scope changes that unblock the success metric are allowed during month 1.

## 2. Users and Access

- Roles: `manager`, `practitioner`, `parent`.
- Parent provisioning: manager-invite only.
- First manager bootstrap: one-time seed/admin command.
- Login identifier: unique case-insensitive normalized email.
- Billing visibility: manager + parent only; practitioner has no billing access.
- Parent visibility: one parent account can view invoices for all linked children.
- Guardian unlink rule: parent loses access to that child invoices immediately, including historical invoices.

## 3. Architecture and Stack

- Frontend: existing Angular app in `web/`.
- Backend: Gin REST API in `api/`.
- API prefix: `/api/v1`.
- API style: plain JSON resources + HTTP status codes.
- API errors: `{ code, message, details?, request_id }`.
- Data access: `sqlc + pgx` (no ORM).
- Local dev DB: local PostgreSQL (no Docker required).
- Production deployment: single VM with Docker Compose.

## 4. Tenancy, Branch, and Authorization

- Data model includes `tenant_id` and `branch_id` from day 1.
- Pilot uses one default branch but remains branch-scoped.
- Authorization model: role checks + scope checks (tenant, branch, parent-child linkage).
- Isolation enforcement in MVP: application-layer scoping only; PostgreSQL RLS deferred post-pilot.

## 5. Auth and Session Policy

- No mandatory MFA in month 1.
- Session model: access JWT + DB-backed refresh token.
- Token transport: bearer access token + HttpOnly secure refresh cookie.
- Token lifetimes: access 15 minutes, refresh 30 days, refresh token rotates.
- Concurrent sessions: allowed with per-session revocation.
- Password reset: basic email-link flow.

## 6. Attendance Policy

- Model: event-based attendance (`check_in`, `check_out`, `correction`).
- Attendance scope: check-in/check-out only (no room moves).
- Correction authority: manager only; practitioner cannot directly edit historical events.
- Billing rounding: each session rounds up to nearest 15 minutes.
- Timezone and boundaries: `Europe/London`, calendar day boundaries.
- Cross-month session allocation: assigned to month of check-in date.
- Incomplete sessions: excluded from billing until manager correction.
- Absence handling: simple absence marker, not auto-billed.

## 7. Funding and Billing Rules

- Funding model: simple funded-hours deduction (`Funding v1`).
- Funding applies to core childcare hours only; extras always payable.
- Core formula: `max(0, core_attended_hours - funded_hours_allowance)`.
- Child minimum before billing/attendance: name, DOB, start date, one guardian, billing rate.
- Billing period: monthly calendar month only.

## 8. Invoicing Policy

- Invoice source of truth: attendance actuals.
- Draft generation: manager manual monthly trigger.
- Parent visibility: draft invoices are manager-only; parents see issued and later states only.
- Issue mode: one-by-one and bulk; default bulk with confirmation.
- Immutable issued invoices: no direct edit after issue.
- Invoice numbering: `INV-YYYYMM-####`.
- Invoice granularity: per child (not family-combined).
- Currency: GBP only.
- Tax mode: non-VAT only for month 1.
- Due policy: due on receipt.
- Overdue transition: unpaid issued invoice becomes `overdue` at next local midnight.
- Incomplete attendance on issue: block only affected children and return exception list.
- Draft regeneration: allowed per child after correction; no duplicates for same month.
- Issued regeneration: not allowed; use adjustment flow.
- Adjustment flow: manager creates linked follow-up adjustment invoice with required reason.
- Child leaves mid-month: billing is automatically derived from attendance actuals up to leave date.

## 9. Payments and Stripe

- Payment model: full payment only (no partials).
- Retry model: same issued invoice can create fresh Stripe checkout session.
- Failure state: `payment_failed` explicit invoice status.
- Webhook requirements: signature verification, idempotent event handling, safe retries.

## 10. Audit, Deletion, and Retention

- Core records (`child`, `attendance`, `invoice`) are never hard-deleted in MVP.
- Entity identifiers: UUID primary keys (`UUIDv7` preferred, `UUIDv4` acceptable).
- Invoice explainability: persist intermediate components (attended/funded/billable minutes, hourly rate) and final totals.
- Audit mandatory for: user provisioning/role changes, child updates, attendance events/corrections, funding updates, invoice draft/issue, payment status updates.
- Retention engine: deferred; pilot keeps records with no-hard-delete baseline.

## 11. Operations, Testing, and Reliability

- Observability baseline: structured logs + metrics for webhook outcomes, invoice-generation health, and auth failures.
- Async scope: background jobs for non-immediate external work (email and retries); checkout session creation remains synchronous.
- Scheduler ownership in MVP: single scheduler instance controlled by environment flag.
- Test priority: funding logic, invoice state transitions, authorization scope checks, webhook idempotency.
- UI E2E: minimal happy-path coverage only.

## 12. Go-Live Governance

- UAT gate: explicit manager signoff required for attendance, correction, invoice generation, payment, payment retry.
- Rollback policy: if critical billing defect appears, pause new invoice issuance and use fallback procedure until verified fix.
