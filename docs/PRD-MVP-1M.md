# Nursery Management System - 1 Month Lean MVP PRD

## 1. MVP Goal

Deliver a live pilot for 1 UK nursery within 30 days so the nursery can run:

- daily attendance without paper/spreadsheets
- monthly invoice generation with funded-hours deduction
- parent payment collection through Stripe

## 2. Success Metric (Hard)

By day 30, one nursery can operate daily attendance and monthly invoicing without spreadsheets.

## 3. Target Customer and Users

### Customer profile

- First paying customer: single independent nursery owner (single site)

### MVP user roles

- Manager
- Practitioner
- Parent

## 4. Technical Baseline (Locked)

- Frontend: existing Angular app in `web/`.
- Backend: Gin REST API in `api/` using `/api/v1` routes.
- Persistence: PostgreSQL with `sqlc + pgx` typed queries.
- Migrations: manual `golang-migrate` workflow.
- Local development: local PostgreSQL (no Docker required).
- Production deployment: single VM with Docker Compose.
- Multi-scope model: keep `tenant_id` and `branch_id` from day 1.

## 5. Product Strategy for 30 Days

Focus on a narrow revenue-critical workflow only:

1. Child attends nursery
2. Attendance is captured reliably
3. Monthly funded hours are deducted by simple rule
4. Invoice is generated and sent
5. Parent pays via Stripe
6. Manager sees paid/unpaid status

Everything else is deferred.

## 6. In Scope (Must Build)

### 5.1 Identity and access

- Email/password login
- Role-based access: manager, practitioner, parent
- Manager-invite-only user provisioning for parent/staff users
- Access JWT (15m) + rotating refresh token (30d)
- Password reset by email link
- Concurrent sessions allowed with revocable refresh tokens

### 5.2 Child and guardian core

- Create/edit child profile (manager only)
- Create/edit guardian profile
- Link guardian to child
- Required before attendance/invoicing: child name, DOB, start date, one guardian, billing rate

### 5.3 Attendance core

- Event-based attendance (`check_in`, `check_out`, `correction`)
- Practitioner/manager sign-in and sign-out for children
- Attendance correction by manager only
- Correction reason required
- Daily attendance list
- Billing rounding: each session rounds up to nearest 15 minutes
- Incomplete sessions are excluded from billing until corrected

### 5.4 Funding v1 (simple)

- Monthly funded-hours allowance per child
- Deduction against core attended hours only (extras always payable)
- Core formula: `max(0, core attended hours - funded hours allowance)`
- Clear calculation summary on invoice

### 5.5 Invoicing

- Monthly invoice generation from attendance actuals (draft -> issued)
- Manager manually generates draft invoices by month
- Invoice issue supports one-by-one and bulk issue (bulk default)
- Line items: gross hours/fees, funded deduction, net due
- Invoice numbering: `INV-YYYYMM-####`
- Due policy: due on receipt; overdue starts 00:00 next day (`Europe/London`)
- Issued invoices are immutable (no direct edit)
- Invoice status: `draft`, `issued`, `payment_failed`, `paid`, `overdue`
- Issue run blocks only children with incomplete attendance and returns exception list
- Manager can regenerate draft invoice for one corrected child

### 5.6 Stripe payments

- Stripe hosted checkout session for invoice payment
- Full payment only (no partial payments)
- Retry payment on same issued invoice with fresh checkout session
- Payment webhook updates invoice status idempotently
- Basic reconciliation log (payment id, amount, timestamp)

### 5.7 Parent billing view

- Parent can view current/past invoices
- Parent can open Stripe payment page
- Parent sees payment status updates
- One parent account may view invoices for all linked children

### 5.8 Minimum compliance baseline

- Audit trail for attendance edits and invoice changes (who, what, when)
- No silent overwrite of attendance or invoice records
- Created/updated timestamps on core records
- Mandatory audit events include invites/roles, child updates, attendance events/corrections, funding updates, invoice draft/issue, and payment status updates

## 7. Out of Scope (Strict)

- Incident and safeguarding module (explicitly skipped)
- Ratio engine and live ratio alerts
- Messaging/chat/newsletters
- Learning journal and EYFS observations
- SEND workflows
- Multi-branch management UI (data model keeps one default branch)
- Advanced funding rules by local authority
- Payroll/HR/rotas
- Data migration/import from legacy systems

## 8. Functional Requirements (Lean)

### FR-01 Authentication

- Users can log in and log out
- Access is restricted by role
- Login identifier is unique case-insensitive normalized email
- First manager account is created by seed command, then invites are manager-led

### FR-02 Child and guardian management

- Manager can create and edit child and guardian records
- Practitioner has read-only access where needed for attendance

### FR-03 Attendance capture

- Practitioner can record check-in/check-out
- System prevents duplicate active check-in
- Manager can correct records with mandatory reason
- Sessions crossing midnight are billed to check-in month in `Europe/London`

### FR-04 Funding calculation v1

- Manager sets monthly funded-hours per child
- System calculates deduction using simple deterministic formula
- Calculation is displayed on invoice in plain language

### FR-05 Invoice lifecycle

- Manager generates monthly draft invoices
- Manager issues invoices
- Parent can view issued invoices

### FR-06 Stripe integration

- Parent can pay via Stripe
- Webhook marks invoice paid/failed
- Payment events stored for audit/reconciliation
- Failed webhook processing is retried safely with idempotent event storage

## 9. Non-Functional Requirements (MVP Level)

- Mobile-friendly UI for practitioners and parents
- Core actions complete within acceptable speed for single-site pilot
- Basic error logging for API/webhook failures
- Daily backup for production database
- HTTPS enabled in pilot environment
- Tenant and branch scope are enforced on every read/write endpoint
- API returns plain JSON resources with standard HTTP status codes

## 10. Data Model (Minimum)

- users
- children
- guardians
- child_guardian_links
- attendance_events
- funding_profiles
- invoices
- invoice_lines
- payments
- audit_logs

All tenant-owned records include `tenant_id`, `branch_id`, `created_at`, and `updated_at`.

## 11. 4-Week Delivery Plan

## Week 1 - Foundation

- Gin API bootstrap and `/api/v1` routing
- Auth and role model
- Child/guardian CRUD
- Initial DB schema and audit log table
- Basic manager and practitioner screens

## Week 2 - Attendance + Funding v1

- Attendance capture UI and API
- Attendance correction workflow with reason
- Funding profile setup per child
- Deterministic funding deduction logic

## Week 3 - Invoices + Stripe

- Invoice generation and issue workflow
- Parent invoice view
- Stripe checkout integration
- Stripe webhook handler and reconciliation status updates

## Week 4 - Pilot Hardening and Go-Live

- Bug fixes and edge-case handling
- UAT with nursery owner and 1-2 practitioners
- Pilot production setup (single VM Docker Compose) and seed data
- Go-live checklist and runbook

## 12. Acceptance Criteria (Pilot Ready)

1. Staff can complete daily attendance for all children in system.
2. Manager can generate monthly invoices with funded-hours deduction.
3. Parents can pay invoices through Stripe.
4. Invoice status updates automatically after Stripe payment event.
5. Attendance and invoice edits are auditable (actor + timestamp + change reason where required).
6. Pilot nursery runs one monthly invoice cycle without spreadsheets.

## 13. Key Risks and Mitigations

- Stripe integration delays -> use hosted Stripe flow first, avoid custom payment UI.
- Scope creep -> freeze backlog to in-scope modules only for 30 days.
- Funding ambiguity -> lock to simple funded-hours deduction rule for MVP.
- Solo developer bandwidth -> prioritize end-to-end billing flow over secondary UX polish.

## 14. Post-MVP Expansion (Not Month 1)

- Incident/safeguarding module
- Ratio engine
- Messaging and parent communications
- LA-specific funding rule packs
- Multi-branch dashboards and governance
