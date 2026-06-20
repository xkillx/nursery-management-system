# Decision Baseline - MVP Month 1

This document consolidates locked implementation decisions for the 30-day pilot MVP.
If a future task conflicts with this baseline, treat this file as the source of truth unless a new explicit decision supersedes it.

## Post-MVP Supersession

The one-site MVP lock below remains valid as a **historical baseline**. Post-MVP planning now proceeds with the following accepted expansions that supersede the one-site constraint for future work:

- Owner and four-site oversight is the first Post-MVP expansion lane after pilot readiness.
- The first owner release is oversight-first: cross-site read and administration, not branch-scoped operational writes.
- The Post-MVP feature sequence is registration/consent, room/session planning, ratio safety, safeguarding/incidents, then learning journeys.
- See `docs/POST-MVP-ROADMAP.md` for the canonical next-work roadmap.

Do not treat the one-site lock below as blocking accepted Post-MVP scope.

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
- Funding applies to booked core childcare minutes only; extras always payable.
- Core formula: `max(0, booked_core_minutes - funded_minutes_allowance)`.
- Child minimum before billing/attendance: name, DOB, start date, one guardian, billing rate, an active Term, a booking pattern, and a Funding Profile for the billing month.
- Billing period: monthly calendar month only.

## 8. Invoicing Policy (Post-MVP: booking-based advance-pay)

- Invoice source of truth: the active Term's Booking Pattern. Attendance is no longer a billing input. See `docs/adr/0006-booking-pattern-billing-source.md`.
- Commercial model: 12-month fixed-term advance-pay. See `docs/adr/0007-12-month-fixed-term-contract.md`.
- Draft generation: scheduler-driven on the 25th of the prior month at 00:05 Europe/London; the manager can also trigger a full-month regeneration explicitly.
- Issue: auto-issue on the 25th. The invoice is parent-visible and payable the same day.
- Parent visibility: issued (and later) invoices are parent-visible; drafts are manager-only.
- Issue mode: bulk (the scheduled run); one-by-one and bulk-issue endpoints remain for the "regenerate and re-issue" path.
- Immutable issued invoices: no direct edit after issue.
- Invoice numbering: `INV-YYYYMM-####` (unchanged).
- Invoice granularity: per child (not family-combined).
- Currency: GBP only.
- Tax mode: non-VAT only for month 1.
- Due policy: `due_at = billing_month_start` (1st of the month at 00:00 UTC). The system issues on the 25th of the prior month; the parent owes from the 1st.
- Overdue transition: unpaid issued invoice becomes `overdue` at 00:00 on the 8th of the billing month in Europe/London (1-week grace). A `payment_failed` invoice does not transition to `overdue`.
- Attendance block: practitioners see a per-child "attendance blocked — invoice overdue" warning on the daily list when the child's billing month is past the 8th and the invoice is unpaid; managers can override per child for the current billing month.
- Booking-pattern changes: in-term changes follow a 1-month-notice rule. Decreases are auto-approved and effective on the 1st of the month at least one calendar month after the request. Increases require manager approval (room capacity + staff:child ratios).
- Early termination: 1-month notice. The parent pays for the notice month only; no further liability; no refunds of any pre-generated drafts (none will exist for the cancelled term going forward under the rolling-generation rule).
- 52-week billing: the booked pattern is billed for all 52 weeks of the year, including weeks the nursery is closed; no closure calendar, no pro-rating.
- Extras: manual invoice line items added by managers (unchanged from the existing `Extras Charging Model`).
- Term renewal: T-45 soft warning, T-30 renewal prompt, T+1 transition to `ended` if no renewal recorded. The child becomes `inactive` on the day after `term_end_date` when no renewal is on file.
- Adjustment flow: manager creates linked follow-up adjustment invoice with required reason (unchanged).
- Settling-in window: free, no invoice; attendance is captured for the register only.

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
