# Context Glossary

## Pilot Nursery

The first live customer setting used to validate the MVP in production.

## Tenant

A single nursery business boundary that owns its own data and users.

## MVP Scope (Month 1)

The first 30-day release focused only on attendance, funding deduction, invoicing, and Stripe payment collection.

## User Role: Manager

A nursery staff role responsible for administration, invoicing, and operational oversight.

## User Role: Practitioner

A nursery staff role focused on day-to-day child attendance operations.

## User Role: Parent

A guardian-side role that views invoices and completes payments.

## Funding v1

A simple funded-hours deduction model used to reduce monthly billed amount per child.

## Invoice

A monthly billing statement showing gross fees, funded deduction, and final amount due.

## Parent Account Provisioning (MVP)

Parent user accounts are created by manager invitation only; public self-signup is not used in month 1.

## Attendance Record (MVP)

Attendance is captured as an event history (check-in, check-out, correction) instead of a single mutable interval row.

## Attendance Correction Authority (MVP)

Only managers can create attendance correction events.

## Invoice Source of Truth (MVP)

Monthly invoice billable hours are derived from attendance actuals.

## Funding Application Rule (MVP)

Funded-hours deduction applies only to core childcare hours; extras remain fully payable.

## Core Billing Formula (MVP)

Core due hours are calculated as `max(0, core attended hours - funded hours allowance)` before pricing is applied.

## Attendance Billing Rounding (MVP)

Each attendance session is rounded up to the nearest 15 minutes for billing.

## Billing Timezone (MVP)

Attendance day boundaries and invoice-period calculations use `Europe/London` local time.

## Incomplete Attendance Handling (MVP)

Attendance records missing check-out are excluded from automatic billing until resolved by manager correction.

## Invoice Generation Flow (MVP)

Managers manually generate monthly draft invoices before any invoice is issued.

## Invoice Issue Mode (MVP)

Managers can issue invoices one-by-one or in bulk; the default flow is bulk issue with confirmation.

## Issued Invoice Edit Policy (MVP)

Issued invoices are immutable; changes require explicit adjustment rather than direct edits.

## Payment Scope (MVP)

Parents pay invoices in full; partial payments are not supported in month 1.

## Payment Failure State (MVP)

Failed or canceled payment attempts move invoices to a `payment_failed` state.

## Invoice Numbering (MVP)

Invoice identifiers follow `INV-YYYYMM-####` sequence format.

## Invoice Granularity (MVP)

Invoices are issued per child, not combined at family level.

## Sibling Discount Policy (MVP)

Sibling discounts are deferred and not part of month 1 billing rules.

## Extras Charging Model (MVP)

Extras are added manually as invoice line items by managers.

## Attendance Capture Scope (MVP)

Attendance capture includes check-in and check-out only; room-move tracking is out of scope.

## Absence Handling (MVP)

Absences are recorded with a simple marker and are not billed automatically.

## Billing Period (MVP)

Billing runs monthly on calendar-month boundaries.

## Child Enrollment Minimum (MVP)

A child requires name, date of birth, start date, one linked guardian, and a billing rate before attendance and invoicing flows begin.

## Parent Visibility Scope (MVP)

One parent account can be linked to multiple children and can view invoices for all linked children.

## Login Identifier (MVP)

User authentication uses a unique, case-insensitive normalized email address as the login identifier.

## Pilot Tenant Bootstrap (MVP)

The first manager account for a tenant is created through a one-time administrative seed command; subsequent users are provisioned by manager invitation.

## Audit Baseline Scope (MVP)

Audit logs are mandatory for user provisioning changes, child record updates, attendance events and corrections, funding profile updates, invoice draft/issue actions, and payment-status updates.

## Cross-Month Session Allocation (MVP)

An attendance session that crosses midnight is allocated to the invoice month of its check-in date using `Europe/London` local time.

## API Response Style (MVP)

API endpoints return plain JSON resources with standard HTTP status codes instead of a global response envelope.

## API Error Contract (MVP)

API errors use a consistent JSON structure with stable error code, human-readable message, optional details, and request identifier.

## Persistence Strategy (MVP)

Data access uses `sqlc` with `pgx` and typed SQL queries rather than an ORM.

## Migration Workflow (MVP)

Schema changes are applied with manual `golang-migrate` commands against local PostgreSQL and are not auto-run on API startup.

## Session and Token Policy (MVP)

Authentication uses short-lived access JWTs plus database-backed refresh tokens that can be revoked.

## Token Transport Policy (MVP)

Access tokens are sent as bearer tokens, while refresh tokens are stored in secure HttpOnly cookies.

## Session Concurrency (MVP)

Users may hold multiple active sessions concurrently, with token revocation supported per session.

## Token Lifetime Policy (MVP)

Access tokens expire after 15 minutes; refresh tokens expire after 30 days and rotate on refresh.

## Password Reset (MVP)

Users can reset passwords through a basic email-link reset flow.

## Email Delivery Strategy (MVP)

Manager invites and password reset messages are sent through a single provider abstraction, using an SMTP sandbox in development and one transactional email provider in production.

## Branch Scope (MVP)

Records remain branch-scoped in the data model with one default branch used in the pilot.

## Core Record Deletion Policy (MVP)

Core child, attendance, and invoice records are not hard-deleted; corrections, voiding, or archival flows are used instead.

## Invoice Due Policy (MVP)

Invoices are due on receipt when issued.

## Invoice Overdue Transition (MVP)

An unpaid issued invoice transitions to `overdue` at 00:00 the next local day in `Europe/London`.

## Invoice Issue Exception Handling (MVP)

Invoice issue runs can proceed for eligible children while children with incomplete attendance are blocked and returned in an exception list for manager resolution.

## Draft Invoice Regeneration (MVP)

After attendance corrections, managers can regenerate draft invoices for individual children without rerunning the full month batch.

## Invoice Payment Retry (MVP)

Issued invoices remain payable after `payment_failed` or `overdue` states by creating a fresh Stripe checkout session for the same invoice.

## Tax Handling (MVP)

Month-1 invoicing runs in a single non-VAT mode without VAT calculation logic.

## Invoice Currency (MVP)

All invoices and Stripe checkout sessions use `GBP` only in month 1.

## Portal Delivery Model (MVP)

Managers, practitioners, and parents use one web application with role-scoped access rather than separate applications.

## Billing Visibility Scope (MVP)

Invoice amounts, statuses, and payment details are visible only to managers and parents; practitioners do not have billing access.

## Access Scope Enforcement (MVP)

All record access is constrained by both tenant and branch scope in month 1, even with a single pilot tenant and default branch.

## Tenancy Isolation Enforcement (MVP)

Tenant and branch isolation is enforced in application-layer authorization and query scoping in month 1; database RLS is deferred to post-pilot hardening.

## Test Priority (MVP)

Automated test coverage is prioritized for funding calculation logic, invoice state transitions, authorization scope checks, and Stripe webhook idempotency, with minimal UI end-to-end coverage on core happy paths.

## Observability Baseline (MVP)

Operations monitoring uses structured logs plus essential metrics for webhook outcomes, invoice-generation job health, and authentication failures.

## Async Processing Scope (MVP)

Non-immediate external operations such as email sending and retry handling run through background jobs; user-triggered Stripe checkout session creation remains synchronous.

## Retention Policy Scope (MVP)

Configurable retention/deletion policy automation is deferred; pilot records are retained with core no-hard-delete rules.

## UAT Signoff Gate (MVP)

Go-live requires explicit manager signoff on attendance, correction, invoice generation, payment, and payment-retry user acceptance checks.

## Go-Live Rollback Policy (MVP)

If a critical billing defect appears at go-live, new invoice issuance is paused and fallback procedures are used until a verified fix is deployed.

## Scope Change Rule (MVP)

During the 30-day delivery window, only changes that unblock the defined success metric are accepted; all other requests are deferred to post-MVP backlog.

## Authorization Model (MVP)

Authorization combines role checks with scope checks, including tenant scope, branch scope, and parent-child linkage where applicable.

## Attendance Edit Authority (MVP)

Historical attendance events are not edited directly by practitioners; corrections are captured as manager-only correction events.

## Guardian Unlink Visibility Rule (MVP)

When a guardian-child link is removed, that parent immediately loses access to that child's invoices, including historical invoices.

## Draft Invoice Idempotency (MVP)

Regenerating draft invoices for the same month does not create duplicates; existing eligible drafts are replaced or updated unless already issued.

## Issued Invoice Regeneration Policy (MVP)

Issued invoices are not regenerated after attendance changes; post-issue changes require an explicit adjustment flow.

## Adjustment Flow (MVP)

Post-issue billing changes are handled through a manager-created follow-up adjustment invoice linked to the original issued invoice, with a required reason.

## API Versioning (MVP)

HTTP endpoints are published under a versioned `/api/v1` route prefix.

## Deployment Model (MVP)

Production runs on a single virtual machine using Docker Compose for service orchestration.

## Entity Identifier Strategy (MVP)

Domain entities use UUID primary keys (UUIDv7 preferred, UUIDv4 acceptable).

## Invoice Explainability Persistence (MVP)

Invoice line storage preserves both intermediate billing components (attended minutes, funded minutes, billable minutes, hourly rate) and final totals.

## Parent Draft Invoice Visibility (MVP)

Draft invoices are visible only to managers; parents can view invoices only after they are issued.

## Mid-Month Leave Billing (MVP)

If a child leaves during a month, billing is automatically derived from attendance actuals up to the leave date.
