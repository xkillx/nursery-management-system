# Post-MVP Roadmap

This document is the canonical next-work roadmap after the completed month-1 MVP baseline. Agents and contributors should start here to understand what comes next, what is already built, and what the accepted priority order is.

## Status Summary

The month-1 MVP baseline is **complete**. The system supports manager, practitioner, and parent access with attendance, absence markers, funding deduction, invoicing, parent invoice access, and Stripe payment collection. Production/UAT hardening and new feature modules are Post-MVP work.

## Completed MVP Baseline

The MVP delivered:

- Authentication: login, refresh, logout, membership switch, password reset, manager invites
- Child and guardian lifecycle management with guardian-child links and parent mappings
- Attendance: check-in, check-out, manager corrections, absence markers
- Funding v1: per-child monthly funded-hours allowance
- Invoicing: draft generation, preflight, bulk issue, parent invoice visibility
- Payments: Stripe hosted Checkout, webhook reconciliation, payment status tracking
- Audit logging, structured observability, and authorization test coverage

Historical implementation records:

- `docs/MVP-30D-API-BACKEND-BACKLOG.md` — API backend execution record
- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md` — Angular frontend execution record
- `docs/DECISION-BASELINE.md` — historical MVP decision lock

## First Gate: Pilot Readiness

The first Post-MVP gate makes the completed MVP safe and understandable for pilot operation. It covers production deployment, backup, Stripe operations, seed data, smoke testing, visual QA, accessibility, and terminology cleanup rather than new product modules.

### Backend hardening (API-29 to API-34)

| ID | Task | Status |
|---|---|---|
| API-29 | Dockerfile and production Docker Compose for single-VM deployment | Pending |
| API-30 | Backup and restore runbook for production PostgreSQL | Pending |
| API-31 | Stripe operational runbook: webhook setup, retry inspection, event replay | Pending |
| API-32 | UAT seed data for one tenant with four sites, owner, managers, staff, parents, children, and billing scenarios | Pending |
| API-33 | Backend UAT script and critical/high defect fixes | Pending |
| API-34 | Optional: reporting and CSV export for invoice/payment data | Stretch |

### Frontend hardening (FE-31 to FE-36)

| ID | Task | Status |
|---|---|---|
| FE-31 | Playwright setup for frontend smoke tests and screenshots | Pending |
| FE-32 | Playwright smoke tests: login routing, attendance, invoice run, parent invoices, Stripe redirect | Pending |
| FE-33 | Visual QA at mobile, tablet, and desktop widths | Pending |
| FE-34 | Accessibility pass for all MVP screens | Pending |
| FE-35 | Pilot copy and terminology cleanup | Pending |
| FE-36 | Final frontend production readiness pass | Pending |

## Second Lane: Owner and Four-Site Oversight

After pilot readiness, the first product expansion is owner-visible cross-site operations. The first owner release is **oversight-first**: the owner can inspect site summaries, compare finance and attendance health, review pilot readiness, and administer branch-manager access. Routine attendance, child-record edits, invoice issue, and branch workflow corrections remain branch-manager responsibilities.

| ID | Task | Status |
|---|---|---|
| API-PM-08 | Owner and four-site access model: owner role, branch switching/filtering, cross-site summaries, authorization boundaries | Done |
| FE-PM-09 | Owner and four-site navigation/access experience: owner landing view, site switcher, branch-scoped manager/staff views, parent route isolation | Done |

Reference: `CONTEXT.md` defines `User Role: Owner` and `Owner Oversight Scope (Post-MVP)`.

## Owner Site Administration

Extends the owner role beyond read-only oversight into site lifecycle management. The current owner dashboard displays site summaries and manager access but has no way to create, edit, or deactivate sites (branches) — those actions require seed scripts or direct DB access.

| ID | Task | Status |
|---|---|---|
| API-PM-10 | Owner site CRUD backend: create site, update site name/address, deactivate/reactivate site | Backlog |
| FE-PM-11 | Owner site management UI: site list, create/edit form, deactivate/reactivate controls | Backlog |

## Later Lane: Super Admin Platform Management

Super admin is a platform-level NMS operator role, separate from tenant owners. The super admin dashboard should manage the NMS SaaS control plane without becoming a hidden bypass for nursery operations. Initial scope should focus on tenant lifecycle, platform configuration, operational health, support tooling, feature flags, plan/billing metadata, and audit visibility. Any cross-tenant data access must be explicit, audited, least-privilege, and designed to preserve tenant isolation.

| ID | Task | Status |
|---|---|---|
| API-PM-09 | Super admin platform management API: tenant/site lifecycle, platform user lookup, support-safe account actions, feature flags, plan/billing metadata, platform health summaries, and cross-tenant audit controls | Backlog |
| FE-PM-10 | Super admin dashboard for NMS platform management: tenant directory, tenant health/status, owner/manager access overview, feature flag and plan controls, support workflows, and audit viewer | Backlog |

Reference: `docs/user-roles-and-permissions.md` lists Super Admin as a future role; this lane defines the first backlog scope for that role.

## Feature Sequence

After pilot readiness and owner/four-site oversight, feature work proceeds in this order:

### 1. Child Management (identity, profile, contacts, health, safeguarding, consent, funding, collection)

Direct child creation, profile, contacts, health, safeguarding, consent, funding, collection, and room placement. Child is the root identity; everything else is a sub-record of the child. The atomic-create milestone lives in API-PM-08.

| ID | Task | Status |
|---|---|---|
| API-PM-08 | Child management refactor — drop registration model, adopt direct child creation | Done |
| FE-PM-08 | Manager child-edit stepper (read-only detail + Edit button) | Done |
| API-PM-02 | Consent and acknowledgement ledger from `docs/forms/parental-consent-form.md` | Pending |

### 2. Room and Session Planning

Room setup, session templates, child bookings, capacity warnings, and eligibility checks.

| ID | Task | Status |
|---|---|---|
| API-PM-04 | Room/session planning and capacity model | Pending |
| FE-PM-05 | Room/session planning and booking UI | Pending |

### 3. Ratio Safety

Staff-to-child ratio rules, live room checks, unsafe booking blocking, and manager override flow.

| ID | Task | Status |
|---|---|---|
| API-PM-05 | Staff-to-child ratio safety engine | Pending |
| FE-PM-06 | Ratio safety dashboard and live room checks | Pending |

### 4. Safeguarding and Incidents

Incident/safeguarding records, practitioner submission, manager review, and restricted access.

| ID | Task | Status |
|---|---|---|
| API-PM-06 | Safeguarding and incident record module | Pending |
| FE-PM-07 | Safeguarding and incident record screens | Pending |

### 5. Learning Journeys

EYFS observations, practitioner drafts, manager review, and parent-visible approved entries.

| ID | Task | Status |
|---|---|---|
| API-PM-07 | Learning journey and EYFS observation module | Pending |
| FE-PM-08 | Learning journey and EYFS observation screens | Pending |

## Billing Follow-up

The adjustment invoice endpoint is a separate billing follow-up, promoted only when post-issue correction becomes operationally required.

| ID | Task | Status |
|---|---|---|
| API-BILL-01 | Manager adjustment invoice endpoint for post-issue billing corrections | Deferred |

## Site and Branch Terminology

Product-facing documentation uses **site** for a nursery location. Existing engineering, API, and database docs use **branch**. Site and branch refer to the same location boundary until a later decision separates them.

## Non-Goals

- This roadmap does not replace `docs/PRD.md`, which remains the broad strategic product vision.
- Docker deployment, Playwright, and schema changes are covered by their respective hardening or feature tasks above.
- Registration, consent, planning, ratios, safeguarding, and learning journeys are Post-MVP feature modules, not unfinished MVP work.

## Verification

When implementing Post-MVP work:

- Backend: `cd api && go test ./...` must pass
- Frontend: `cd web && npm run build` must pass
- New migrations: `make migrate-verify` with `VERIFY_DATABASE_URL` must pass
- Authorization: new routes need role/scope test coverage matching the MVP baseline
