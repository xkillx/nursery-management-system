# Phase 1 MVP Gap Analysis - UK Nursery Management System

Date: 2026-06-19

Scope: functional completeness only. This audit intentionally excludes code quality, style, refactoring, and architecture-pattern commentary.

Evidence reviewed: `api/internal/modules/*`, `api/internal/app/bootstrap/bootstrap.go`, migrations `000001`-`000034`, `api/db/query/*.sql`, `docs/PRD.md`, `docs/API-SCHEMA-STATE.md`, `docs/API-CONTRACT.openapi.yaml`, `docs/API_REVIEW.md`, `docs/DECISION-BASELINE.md`, `docs/user-roles-and-permissions.md`, frontend routes and screens under `web/src/app`.

Important baseline conflict: `docs/DECISION-BASELINE.md` defines a narrower pilot baseline: attendance, funding v1, invoicing, Stripe payments, and parent invoice view; it explicitly defers safeguarding, ratio engine, messaging, SEND, and advanced funding. This report assesses the repository against the broader Phase 1 MVP scope supplied in the audit request.

## Executive Summary

Overall MVP Completion: **43%**

Modules Complete: **0/10**

Risk Level: **High**

Critical Missing Features:
1. Admissions pipeline is not implemented: no enquiries, waiting list, tour scheduling, prospect conversion workflow, or admission lifecycle.
2. Session pattern/bookings are not implemented: no full-day/half-day/session catalogue, child booking pattern, expected attendance, or schedule-driven billing.
3. UK funding is not production-grade: no entitlement codes, term/reconfirmation lifecycle, 15/30-hour/2-year-old funding rules, stretched delivery, or local-authority claim evidence.
4. Parent communication is effectively absent: no messaging, notifications, announcements, or parent operational portal beyond invoices.
5. Compliance-critical operational records are missing: accident/incident log, medication administration, SEND/EHCP plans, key person, documents/attachments, register closure, late collection, ratio evidence.

Recommended Priority Order:
1. Implement session patterns/bookings because admissions, attendance, funding, billing, occupancy, and absence all depend on expected sessions.
2. Implement admissions pipeline from enquiry to enrolment, including waiting list, tour, registration status, and conversion.
3. Replace funding v1 with UK entitlement lifecycle and allocation model.
4. Add compliance-critical child records: SEND, accident/incident, medication, documents, custody restrictions, key person.
5. Add parent communication and parent self-service flows for messages, announcements, absence reporting, and document/consent updates.

## Level 1 - Module Coverage

| MVP Module | Implemented % | Status | Confidence |
|---|---:|---|---|
| Child Management | 58% | Partial | High |
| Admissions & Enrolment | 20% | Prototype Only | High |
| Session Pattern Management | 0% | Not Started | High |
| Room Management | 45% | Partial | High |
| Attendance & Registers | 55% | Partial | High |
| Billing & Invoicing | 55% | Partial | High |
| Government Funding | 35% | Partial | High |
| Parent Communication | 5% | Not Ready | High |
| User Management & RBAC | 55% | Partial | High |
| Multi-Tenant Platform | 50% | Partial | Medium-High |

## Module-by-Module Analysis

### 1. Child Management

Status: **Partial**

Completion: **58%**

Confidence: **High**

Implemented:
- Child identity records with first/middle/last name, date of birth, start/end date, active status.
- Guardian records and child-guardian links.
- Parent membership to guardian mapping for parent invoice access.
- Child profile subrecord with demographics, home address, disability status, access requirements, routine care, registration date, GDPR declaration metadata.
- Child contacts for parent/carer, emergency contact, and authorised collector.
- Health profile with medical conditions, prescribed medication status, dietary requirements, immunisation status, GP/doctor contact, and health visitor contact.
- Safeguarding-style profile with social services status, social worker contact, concern flags, professional referrals, and restricted notes.
- Consent record with multiple operational consents and paper-form signer metadata.
- Collection settings with collection-password hash and over-18 acknowledgement.
- Room assignment history, billing profile, funding record, and leaving record.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Child profile | Partial | Identity plus profile subrecord. | NHS number, key person, primary contact, custody/court-order restrictions, family grouping, preferred name, photo, enrolment status lifecycle. | Incomplete operational and compliance record. |
| Parent/guardian information | Partial | Guardians and child contacts. | Primary payer, billing responsibility split, parental responsibility enforcement, court restrictions, phone normalisation, family grouping. | Incorrect communications, billing ambiguity, collection risk. |
| Emergency contacts | Partial | `child_contacts` includes emergency contact type. | Priority ordering semantics beyond `sort_order`, collection restrictions, authorised pickup validation flow. | Staff may not have legally safe collection guidance. |
| Medical information | Partial | Medical, medication, immunisation, dietary, GP/health visitor fields. | Medication administration log, care plans, allergy severity/action plan, asthma/inhaler/EpiPen specifics, emergency medical plan. | High safeguarding and insurance risk. |
| Allergies | Partial | Dietary notes and side-effects fields. | Dedicated allergy entity, allergen type, severity, treatment, review date, kitchen visibility. | Operational risk in meals and emergency response. |
| GP information | Partial | Doctor name/address/phone. | NHS number, practice identifier, emergency treatment workflow linkage. | Less reliable emergency and funding evidence. |
| SEND information | Partial | Disability status, access requirements, developmental concern flags, referrals. | SEND profile, EHCP number, SENCo owner, support plan, review dates, DAF/DIS linkage. | SEND obligations and funding evidence are not covered. |
| Consent records | Partial | Single current consent row. | Consent history/ledger, withdrawal effective date, document evidence, parent self-service signature flow. | Cannot prove consent changes over time. |
| Documents and attachments | Missing | No asset/document table found. | Upload, storage, metadata, access controls, expiry/review, signed URLs. | Paper records remain outside system; inspection evidence incomplete. |

Business Impact:
- The child record is useful for internal manager entry but is not yet a complete UK nursery child file.
- Missing medication, incident, SEND, key person, custody, and documents are serious live-operation blockers.

Recommendation:
- Add child document storage, SEND/EHCP profile, key-person assignment history, medication/accident logs, allergy action plans, custody restrictions, and consent ledger before production.

### 2. Admissions & Enrolment

Status: **Prototype Only**

Completion: **20%**

Confidence: **High**

Implemented:
- Manager can create a child with full child-management subrecords.
- Registration/intake UI exists as a child edit flow.
- Child can be linked to guardians, assigned to a room, given funding/billing profile, and marked inactive.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Parent enquiries | Missing | No enquiry entity or route found. | Lead source, enquiry status, requested start, child age, parent details, follow-up tasks. | Nursery cannot manage sales pipeline. |
| Waiting list | Missing | No waiting-list entity found. | Priority, desired sessions, age-band demand, room capacity forecast. | Occupancy planning cannot be run. |
| Tour scheduling | Missing | No tour entity or calendar flow. | Tour slots, reminders, outcome, no-show, conversion. | Admissions admin remains manual. |
| Registration | Partial | Child-management intake captures a registered child profile. | Parent-facing digital registration, draft registration, approval workflow, status transitions, document collection. | Duplicate entry and incomplete onboarding risk. |
| Child enrolment | Partial | Start date, guardian link, room assignment, billing/funding subrecords. | Enrolment state machine, offer acceptance, session pattern, deposit/fees, readiness checklist. | A child can exist without being operationally ready. |

Business Impact:
- The system starts at “create child,” not at real admissions.
- No visibility into prospects, conversion, capacity demand, or incomplete registration.

Recommendation:
- Add `enquiries`, `waiting_list_entries`, `tours`, `registration_drafts`, and `enrolments` with explicit lifecycle states from enquiry to enrolled.

### 3. Session Pattern Management

Status: **Not Started**

Completion: **0%**

Confidence: **High**

Implemented:
- No evidence of session-pattern, booking, or contracted-attendance entities.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Full-day sessions | Missing | None. | Session catalogue/rate card. | Cannot sell or bill booked full-day care. |
| Half-day sessions | Missing | None. | Session type and time bands. | Cannot model common nursery packages. |
| Morning sessions | Missing | None. | AM session template and booking pattern. | Cannot produce expected register. |
| Afternoon sessions | Missing | None. | PM session template and booking pattern. | Cannot produce expected register. |
| Custom attendance schedules | Missing | None. | Effective-dated child booking pattern, ad hoc sessions, holiday-only schedules. | Attendance, funding, billing, occupancy, and absence cannot be reconciled to bookings. |

Business Impact:
- This is the largest functional gap because it breaks expected attendance, absence detection, funded-hours allocation, room capacity forecasting, fee plans, and session-rate billing.

Recommendation:
- Implement effective-dated `session_templates`, `child_session_patterns`, `child_bookings`, and ad hoc session adjustments before expanding billing/funding.

### 4. Room Management

Status: **Partial**

Completion: **45%**

Confidence: **High**

Implemented:
- Room CRUD, archive/reactivate, name, description, age group, capacity, active status.
- Child room assignment history with current assignment indexes.
- Owner and manager room screens.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Room management | Partial | Rooms can be created, edited, archived, reactivated, listed. | Room moves during day, temporary room occupancy, room leader/key person assignment. | Room admin exists but daily room operations are incomplete. |
| Capacity management | Partial | Room capacity field and assigned-child counts. | Capacity enforcement, occupancy by booked sessions, waiting-list demand, over-capacity prevention. | Nursery can overbook operationally. |
| Age range configuration | Partial | Fixed age groups: baby/toddler/preschool/mixed. | Min/max age months, ratio band, room transition rules. | Cannot reliably enforce age-appropriate placement. |
| Child room assignment | Partial | Effective-dated room assignment history. | Validation against capacity, age range, session pattern, and ratio. | Assignment may be administratively invalid. |

Business Impact:
- Useful for static room setup, not enough for real occupancy/ratio control.

Recommendation:
- Add age-range config in months, capacity enforcement by day/session, room move events, and ratio rules linked to staff presence.

### 5. Attendance & Registers

Status: **Partial**

Completion: **55%**

Confidence: **High**

Implemented:
- Check-in, check-out, one open session per child, attendance events, correction workflow.
- Practitioner attendance list.
- Absence marker and clear marker.
- Incomplete sessions block invoice generation.
- Manager attendance corrections and history.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Check-in | Partial | Child check-in with enrollment and absence checks. | Expected-session validation, room selection, authorised drop-off capture, signature/photo, late arrival reason. | Register is actual-only, not schedule-aware. |
| Check-out | Partial | Completes open session. | Collector identity, collection password verification event, late collection fee trigger, handover notes. | Safeguarding and late-fee evidence missing. |
| Attendance registers | Partial | Current attendance list and event history. | Daily register by room, expected vs actual, register close/sign-off, emergency roll call. | Register may not meet operational audit expectations. |
| Absence tracking | Partial | Simple absence marker. | Parent-reported absence, absence reason/categories, booked-session linkage, illness exclusion dates. | Absence remains shallow and not schedule-aware. |
| Late collection tracking | Missing | Check-out time exists only as attendance data. | Due pickup time, grace period, late flag, reason, fee, parent notification. | Lost revenue and weak safeguarding evidence. |

Business Impact:
- Attendance can capture presence, but it is not a complete statutory register or operational room register.

Recommendation:
- Add expected-session registers, room registers, closure/sign-off, collection identity/password events, emergency roll call, and late collection policy.

### 6. Billing & Invoicing

Status: **Partial**

Completion: **55%**

Confidence: **High**

Implemented:
- Branch core hourly rate and per-child custom hourly rate.
- Attendance-minute calculation with 15-minute rounding.
- Draft invoice preflight/generation, invoice runs, invoice lines, issue/bulk issue.
- Invoice immutability after issue.
- Parent invoice view and Stripe checkout payment attempts.
- Stripe webhook reconciliation and manager payment status/events.
- Adjustment invoice schema hooks exist.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Fee plans | Missing | Hourly rate only. | Contract plans, age-band plans, monthly plans, registration/deposit fees. | Cannot model common nursery pricing. |
| Hourly rates | Partial | Branch and per-child hourly rate. | Effective-dated rate cards, age bands, historical pricing, family/sibling pricing. | Disputes and historical recomputation risk. |
| Session rates | Missing | No session templates/rates. | Full-day/AM/PM/session price model. | Cannot invoice booked packages. |
| Invoice generation | Partial | Draft/issue from attendance actuals. | Booking-based charges, extras/consumables, credits, discounts, family invoices, tax/statement exports. | Billing is too narrow for real nursery finance. |
| Payment tracking | Partial | Stripe full-payment attempts and webhook status. | Manual payments, bank transfer/cash/card terminal recording, refunds, partial payments, payment plans. | Finance team cannot reconcile all payment methods. |
| Credits | Partial/Missing | Adjustment schema exists. | Manager UI/API workflow for credit notes and application to balances. | Cannot handle overpayment or disputes cleanly. |
| Discounts | Missing | No discount entity/rule. | Sibling, staff, promotional, hardship, funded consumable discounts. | Revenue/pricing model incomplete. |

Business Impact:
- Current billing can support a very simple hourly-attendance pilot, not real UK nursery fee structures.

Recommendation:
- Build rate cards, session rates, discounts/credits, manual payments/refunds, and booking-based invoice generation.

### 7. Government Funding

Status: **Partial**

Completion: **35%**

Confidence: **High**

Implemented:
- Child funding record with eligibility/intake flags.
- Monthly funding profile with `funded_allowance_minutes`.
- Manager funding overview with flags for missing/zero/unusual monthly allowance.
- Invoice funding deduction from monthly funded minutes.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| 15-hour funding | Partial | Can enter funded minutes. | Universal entitlement model, term dates, age checks, stretched/term-time patterns. | Risk of incorrect claims and parent invoices. |
| 30-hour funding | Partial | Can enter funded minutes. | Eligibility code, reconfirmation date, grace period, working-parent checks. | Local authority clawback risk. |
| 2-year-old funding | Partial | Intake flag exists. | 2YO criteria, eligibility evidence, start term rules. | Compliance and claim evidence incomplete. |
| Funding allocation | Partial | Monthly minutes deducted from invoice. | Allocation to booked sessions/days, term/stretched calculation, unused/carryover rules. | Funding cannot be explained robustly. |
| Funding eligibility tracking | Partial | Basic yes/no/unknown flags. | Code validation status, evidence documents, reconfirmation alerts, local authority claim status. | High operational and revenue risk. |

Business Impact:
- Funding v1 is a manual allowance deduction, not a UK nursery funding engine.

Recommendation:
- Add entitlement records with funding type, eligibility code, term, delivery mode, reconfirmation, evidence, and allocation to booked sessions.

### 8. Parent Communication

Status: **Not Ready**

Completion: **5%**

Confidence: **High**

Implemented:
- Invite and password reset emails.
- Parent can view issued invoices and start Stripe checkout.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Messaging | Missing | No messaging entity/API/screen. | Parent-staff conversations, attachments, audit, read receipts. | Parent engagement is not viable. |
| Notifications | Missing | No notification delivery model. | Event notifications, channels, preferences, delivery/read state. | Parents will not receive operational updates. |
| Announcements | Missing | No announcement entity/API/screen. | Room/site broadcasts, scheduling, audience targeting. | Nursery still needs external comms tools. |

Business Impact:
- Parent portal is finance-only. This is not a usable parent communication MVP.

Recommendation:
- Add message threads, announcements, notification preferences, and delivery tracking. Start with site/room announcements and invoice/absence notifications.

### 9. User Management & RBAC

Status: **Partial**

Completion: **55%**

Confidence: **High**

Implemented:
- Authentication, refresh tokens, password reset, invite acceptance.
- Roles: owner, manager, practitioner, parent.
- Role-protected routes and tenant/branch/parent-child scope checks.
- Audit logs for key mutations.
- Owner manager-access management.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Authentication | Partial | Login, refresh, logout, password reset, invite accept. | MFA, account lockout, email verification, session/device management UI. | Security posture incomplete for production. |
| Authorization | Partial | Coarse role gates and scope checks. | Permission catalogue, custom roles, least-privilege staff variants, DSL/finance/admin roles. | Overbroad or underbroad operational access. |
| Role-based permissions | Partial | Owner/manager/practitioner/parent route groups. | Platform administrator role, finance admin, room leader, DSL, granular permissions. | Real nursery staff structure cannot be represented. |
| Audit logs | Partial | Audit table and writer used for many mutations. | Access logs for sensitive records, audit reporting UI, export, retention policy. | Compliance evidence hard to retrieve. |

Business Impact:
- Good enough for a simple branch pilot, not enough for a live multi-role nursery or SaaS platform.

Recommendation:
- Add platform administrator, permission model, staff profile/role assignments, sensitive-record access logging, MFA/lockout/email verification.

### 10. Multi-Tenant Platform

Status: **Partial**

Completion: **50%**

Confidence: **Medium-High**

Implemented:
- Tenants, branches/sites, branch-scoped records, memberships.
- Owner cross-site summary and manager-access management.
- Tenant and branch IDs across core data.
- App-layer isolation and tests for scope behavior.

Feature Gap Analysis:

| Feature | Status | What Exists | What Is Missing | Business Impact |
|---|---|---|---|---|
| Tenant isolation | Partial | Tenant/branch columns, scoped queries, authorization tests. | PostgreSQL RLS, cross-tenant admin support workflows, tenant data export/deletion. | SaaS isolation depends on app-layer correctness only. |
| Branch management | Partial | Branches exist; owner can see sites; rooms are site-scoped. | Full site CRUD/config, opening hours, holidays, policies, local authority settings. | Multi-site operations are incomplete. |
| Subscription management | Missing | No subscription/plan tables found. | Plans, billing status, tenant limits, payment provider subscription linkage. | SaaS commercial operation impossible. |
| Tenant configuration | Partial/Missing | Branch core hourly rate. | Funding settings, session templates, late fee policy, comms preferences, retention policy, feature flags. | Each nursery cannot be configured safely. |

Business Impact:
- The product has multi-tenant foundations but lacks SaaS platform administration and tenant configuration.

Recommendation:
- Add platform administrator, subscription/plan model, tenant/site configuration, RLS or equivalent defense-in-depth, and support-safe tenant tooling.

## Entity & Data Model Gaps

| Entity | Priority | Reason |
|---|---|---|
| `enquiries` | Critical | Required to manage admissions before a child is enrolled. |
| `waiting_list_entries` | Critical | Required for occupancy and demand management. |
| `tour_bookings` | High | Required for admissions conversion workflow. |
| `registration_drafts` | High | Required for parent-facing incomplete registration before child creation. |
| `enrolments` | Critical | Required to separate prospect/registered/enrolled/started states. |
| `session_templates` | Critical | Required for full-day, half-day, AM, PM, and custom sessions. |
| `child_session_patterns` | Critical | Required for expected attendance, funding allocation, and billing. |
| `child_booking_exceptions` | High | Required for ad hoc sessions, holidays, swaps, and temporary changes. |
| `branch_rate_cards` | High | Required for age/session pricing and historical rates. |
| `funding_entitlements` | Critical | Required for 15h/30h/2yo entitlement logic and evidence. |
| `funding_eligibility_checks` | Critical | Required for codes, reconfirmation, grace periods, and claim evidence. |
| `funding_allocations` | Critical | Required to allocate funded hours to booked sessions/invoices. |
| `invoice_adjustments` / credit note workflow | High | Required for credits and corrections after issue. |
| `discount_rules` | High | Required for sibling/staff/promotional discounts. |
| `payment_receipts` / manual payments | High | Required for non-Stripe payment tracking. |
| `late_collection_events` | High | Required for safeguarding evidence and late fees. |
| `register_closures` | Critical | Required for signed daily register evidence. |
| `room_movement_events` | High | Required for room-level actual occupancy. |
| `staff_profiles` | Critical | Required for real practitioner records, DSLs, room leaders, finance admins. |
| `staff_room_assignments` / shifts | Critical | Required for ratio compliance. |
| `child_key_person_assignments` | Critical | Required by EYFS key-person expectations. |
| `child_send_profiles` | High | Required for SEND/EHCP/support-plan tracking. |
| `medication_administration_records` | Critical | Required for safe medication handling. |
| `child_incidents` / accident book | Critical | Required for accidents, first aid, parent notification, insurance. |
| `child_documents` / attachments | Critical | Required for forms, evidence, funding documents, court orders. |
| `message_threads` and `messages` | High | Required for parent communication. |
| `notifications` | High | Required for operational alerts and parent updates. |
| `announcements` | Medium-High | Required for room/site broadcasts. |
| `tenant_subscriptions` | High | Required for SaaS operation. |
| `tenant_settings` / `branch_settings` | High | Required for configuration, policies, local authority settings. |

Additional data-model gaps:
- Missing value objects: NHS number, phone number, address, funding code, money/rate-card version, session time band, age band in months, entitlement term.
- Missing lifecycle states: enquiry, waiting-listed, tour-booked, tour-completed, offered, registered, enrolled, started, left, re-enrolled.
- Missing audit fields: most profile rows lack inline `created_by`/`updated_by`; audit log exists but sensitive read/access auditing is not evident.
- Missing constraints: no capacity/age validation on room assignment, no session-overlap validation for bookings because bookings do not exist, no funding eligibility age/code checks, no exact-one-primary-contact rule.
- Missing indexes likely needed once added: booking lookup by branch/date/session, funding code lookup, entitlement term lookup, register closure by branch/date/room, notification unread by user.
- Missing UK-specific compliance data: NHS number, key person, SEND/EHCP, DBS/staff suitability, staff qualifications, safeguarding chronology, accident book, medication logs, funding evidence, local authority claim identifiers, court-order/custody restrictions.

## Workflow Gap Analysis

| Workflow | Severity | Description |
|---|---|---|
| Admissions: Enquiry -> Waiting List -> Tour -> Registration -> Enrolment | Critical | Only child creation/registration-style intake exists. Enquiry, waiting list, tour, offer, and admissions lifecycle transitions are missing. |
| Child Lifecycle: Create Child -> Assign Room -> Define Session Pattern -> Start Attendance | Critical | Child creation and room assignment exist. Session pattern is missing, so attendance starts without expected booking context. |
| Attendance: Check-In -> Attendance -> Check-Out -> Register Closure | Critical | Check-in/out exist. Expected register, room register, closure/sign-off, collector identity, and late collection tracking are missing. |
| Funding: Eligibility -> Allocation -> Invoice Impact | Critical | Monthly allowance impacts invoice. Eligibility code/term/reconfirmation and allocation to sessions are missing. |
| Billing: Attendance -> Charges -> Invoice -> Payment -> Balance | High | Attendance-driven invoice and Stripe full payment exist. Fee plans, session rates, discounts, credits, manual payments, refunds, and family balances are missing. |
| Parent Absence Reporting | High | Staff can mark absence; parent absence reporting is missing. |
| Document Collection | Critical | No document upload/review/expiry workflow. |
| SEND Support Review | High | No SEND support plan or review workflow. |
| Medication Administration | Critical | No medication permission/administer/sign-off workflow. |
| Incident/Accident Reporting | Critical | No accident/incident recording, parent notification, or first-aid evidence workflow. |
| Ratio Monitoring | Critical | No staff presence, ratio calculation, or live room safety workflow. |
| Subscription Lifecycle | High | No tenant plan/trial/suspension/payment lifecycle. |

Missing transitions:
- Prospect to child.
- Waiting-list offer to accepted/rejected.
- Registered child to enrolled child.
- Enrolled child to active attendance eligibility based on start date plus session pattern.
- Attendance day to closed register.
- Funding eligibility to validated entitlement.
- Issued invoice to credit/refund/adjusted balance.
- Child left to re-enrolled.

Missing validations:
- Child age against room age range.
- Room capacity against booked and current occupants.
- Check-in against expected booked session.
- Check-out against authorised collector and collection password.
- Funding type against child age and term eligibility.
- Rate selection against effective date.
- Discount eligibility against family/sibling grouping.
- Parent access against active guardian link for all parent-facing child data, not just invoices.

Missing automation:
- Age-band transition alerts.
- Funding reconfirmation alerts.
- Waiting-list capacity matching.
- Tour reminders.
- Register close reminders.
- Incomplete attendance escalation.
- Late collection fee generation.
- Overdue invoice scheduler evidence.
- Consent/document expiry reminders.

## Permission Gap Analysis

| Role | Missing Permission | Severity |
|---|---|---|
| Platform Administrator | Entire role missing from runtime model. | Critical |
| Platform Administrator | Tenant lifecycle, subscription, feature flags, support-safe account actions, cross-tenant audit. | Critical |
| Nursery Owner | Full branch/site configuration and tenant settings management. | High |
| Nursery Owner | Cross-site financial, occupancy, funding, and compliance reports beyond current summaries. | High |
| Nursery Owner | Subscription/billing plan management. | High |
| Nursery Manager | Admissions pipeline permissions: enquiry, waiting list, tour, offer, enrolment approval. | Critical |
| Nursery Manager | Session pattern and booking management. | Critical |
| Nursery Manager | Funding entitlement/code verification and claim evidence. | Critical |
| Nursery Manager | Discounts, credits, manual payment, refund, and adjustment workflows. | High |
| Nursery Manager | Staff/room ratio, key-person, incident, medication, SEND, document management. | Critical |
| Nursery Staff | Room-scoped child access based on assignment. Current practitioner attendance is branch-wide. | High |
| Nursery Staff | Daily care, incident, medication, parent update, and room register actions. | High |
| Nursery Staff | Restrictions on sensitive safeguarding/SEND records. | Critical |
| Parent | View/update own child profile, emergency contacts, medical info, consents, documents. | High |
| Parent | Report absence and view booked sessions. | High |
| Parent | Messaging, notifications, announcements. | High |
| Parent | View payment history beyond invoice/checkout. | Medium |

Missing restrictions:
- No platform administrator vs tenant user separation.
- No DSL-only safeguarding permission.
- No finance-admin role separate from manager.
- No room-leader/practitioner room scoping.
- No per-field restrictions for sensitive child health/SEND/safeguarding data.
- Owner is intentionally barred from branch operational writes, but this may conflict with requested Phase 1 owner capabilities unless formalized.

Tenant and branch isolation gaps:
- App-layer scoping exists, but PostgreSQL row-level security is deferred.
- Parent-child ownership checks are evidenced for invoices; equivalent parent portal checks for profile/communication are not implemented because those parent features are missing.
- Support/platform cross-tenant access model is not implemented.

## Compliance Risks

| Risk | Severity | Recommendation |
|---|---|---|
| No expected session register or register closure/sign-off. | Critical | Add daily/room registers from bookings with closure, corrections, audit, and export. |
| No key person assignment. | Critical | Add effective-dated key-person assignments and surface in child and room views. |
| No accident/incident book. | Critical | Add append-only incident/accident/first-aid records with parent notification and signatures. |
| No medication administration log. | Critical | Add medication consent, administration, dosage, staff witness, and parent acknowledgement workflow. |
| No staff presence/qualification/ratio engine. | Critical | Add staff profiles, room staffing, qualifications, and ratio snapshots. |
| No SEND/EHCP support plan. | High | Add SEND profile, SENCo ownership, support plans, reviews, DAF/DIS evidence. |
| No safeguarding case chronology. | Critical | Add restricted safeguarding concern/case model with access logging and DSL workflow. |
| Consent record is single-current only. | High | Add immutable consent ledger with effective dates and signer evidence. |
| No documents/attachments. | Critical | Add secure document storage with metadata, expiry, review, access audit. |
| Funding evidence is not modelled. | Critical | Add entitlement code, term, reconfirmation, evidence documents, claim status. |
| No NHS number. | High | Add validated NHS number where collected and protect as sensitive health data. |
| No custody/court-order restriction model. | Critical | Add restricted collection/contact rules and hard warnings in checkout/collection. |
| No data retention/erasure engine. | High | Add retention policy, anonymisation, SAR/export, and purge workflow. |
| No sensitive read/access audit for health/SEND/safeguarding. | High | Log access to high-risk child records and exports. |
| No parent communication archive. | High | Add auditable messaging/announcement records with delivery/read status. |

## MVP Readiness Assessment

| Module | Readiness |
|---|---|
| Child Management | Beta Ready |
| Admissions & Enrolment | Prototype Only |
| Session Pattern Management | Not Ready |
| Room Management | Prototype Only |
| Attendance & Registers | Beta Ready for simple actual attendance; Prototype Only for statutory registers |
| Billing & Invoicing | Beta Ready for simple hourly attendance billing; Prototype Only for nursery billing |
| Government Funding | Prototype Only |
| Parent Communication | Not Ready |
| User Management & RBAC | Beta Ready for simple pilot roles; Prototype Only for full Phase 1 |
| Multi-Tenant Platform | Prototype Only |

Overall MVP Completion: **43%**

Overall Readiness: **Prototype/Beta hybrid, not production-ready for a real UK nursery next month.**

Primary production blockers:
- No admissions pipeline.
- No booked sessions/session patterns.
- No complete register lifecycle.
- No production-grade funding model.
- No parent communication.
- Missing compliance-critical records: incidents, medication, SEND, documents, key person, safeguarding chronology.
- No full tenant subscription/platform administration.

## Top 20 Highest Priority Items To Implement Next

| Priority | Item | Reason |
|---:|---|---|
| 1 | Implement session templates and child session patterns. | Dependency for attendance, registers, funding, billing, occupancy, and absence. |
| 2 | Implement expected daily/room registers with closure/sign-off. | Compliance-critical attendance evidence. |
| 3 | Implement admissions pipeline: enquiry, waiting list, tour, offer, registration, enrolment. | Required to operate nursery admissions and occupancy. |
| 4 | Implement UK funding entitlements with 15h/30h/2yo types, codes, terms, reconfirmation, grace periods. | Highest revenue and clawback risk. |
| 5 | Implement funding allocation to booked sessions and invoice lines. | Makes funding explainable and auditable. |
| 6 | Implement fee plans, session rates, and effective-dated rate cards. | Required for real nursery billing. |
| 7 | Implement accident/incident/first-aid records. | Compliance and insurance blocker. |
| 8 | Implement medication administration records. | Child safety and audit blocker. |
| 9 | Implement document/attachment storage with secure access and expiry/review metadata. | Required for registration, funding, consent, court orders, inspection evidence. |
| 10 | Implement key-person assignment history. | EYFS operational requirement. |
| 11 | Implement SEND/EHCP/support-plan profile. | UK compliance and funding support gap. |
| 12 | Implement safeguarding concern/case chronology with restricted access. | Critical confidentiality and safeguarding gap. |
| 13 | Implement late collection tracking and fee policy. | Operational necessity and revenue protection. |
| 14 | Implement parent absence reporting and booked-session visibility. | Parent self-service and attendance accuracy. |
| 15 | Implement parent messaging, announcements, and notification delivery/read state. | Parent communication MVP blocker. |
| 16 | Implement discounts, credits, manual payments, refunds, and adjustment workflows. | Required for real finance operations. |
| 17 | Implement staff profiles, room staffing, qualifications, and ratio snapshots. | EYFS ratio and room safety blocker. |
| 18 | Implement custody/court-order collection restrictions and authorised-collector verification. | High safeguarding risk. |
| 19 | Implement platform administrator, tenant subscription, and tenant configuration. | SaaS operation blocker. |
| 20 | Implement retention/SAR/anonymisation and sensitive-record access audit. | GDPR and regulated-customer trust requirement. |

## Final Auditor Conclusion

The repository contains solid foundations for a narrow pilot: child records, guardians, actual attendance, simple absence, basic funding allowance, hourly invoice generation, Stripe payment, owner overview, and coarse RBAC.

Against the requested Phase 1 UK Nursery Management System MVP, it is not ready for production. The largest blockers are not minor missing fields; they are whole operational domains: admissions, session patterns, statutory registers, parent communication, production-grade funding, compliance records, and platform administration.

For a real UK nursery going live next month, this implementation should be treated as a pilot core for attendance and simple invoicing only, not a complete Phase 1 MVP.
