# Phase 1 MVP Implementation Report — UK Nursery Management System

**Prepared:** June 2026
**Audience:** Engineering, Product, Nursery Stakeholders, Investors
**Version:** 1.0

---

## 1. Executive Summary

### Current System Maturity

The platform is in **advanced MVP stage**. Seven of ten Phase 1 modules have substantial implementation covering database schema, domain logic, API endpoints, and frontend pages. The system uses Go 1.26 (Gin + pgx) with Clean Architecture on the backend, and Angular 21 on the frontend, with multi-tenant PKI-scoped isolation baked into every table.

### Overall MVP Readiness

| Dimension | Status |
|---|---|
| Backend domain logic | ~80% complete |
| Database schema | ~85% complete |
| REST API coverage | ~75% complete |
| Frontend UI coverage | ~65% complete |
| Tests | Good coverage in core modules (funding, billing, attendance, sessions) |

### Business Capability Coverage

| Capability | Status |
|---|---|
| Child onboarding & records | Complete |
| Session/booking management | Complete |
| Room management & assignment | Complete |
| Daily attendance register | Complete |
| Billing & invoicing | Complete |
| Government funding | Complete |
| Payment collection (Stripe) | Complete |
| Parent portal (invoices) | Complete |
| **Admissions & enrolment workflow** | **Not implemented** |
| **Parent communication** | **Not implemented** |
| **Subscription/billing plans** | **Not implemented** |

### Biggest Implementation Risks

1. **Missing admissions pipeline** — nurseries cannot manage enquiries, waiting lists, or tour scheduling. This is a front-door business requirement.
2. **No parent communication** — no in-app messaging, announcements, or push notifications. Parents cannot communicate with staff through the platform.
3. **Subscription & tenant billing** — no mechanism to charge nursery operators for platform usage. The multi-tenant infrastructure exists but has no revenue model implemented.
4. **No document/attachment storage** — child records lack document upload capability (registration forms, consent forms, medical letters).

---

## 2. Current Implementation Assessment

### 2.1 Child Management

**Status: Completed**

**Implemented Features:**
- Child profile (name, DOB, start/end dates, active/inactive status)
- Parent/guardian contacts (supports parent_carer, emergency_contact, authorised_collector types)
- Medical information (conditions, medication, immunisations, dietary requirements)
- GP and health visitor details (stored in child_health_profiles)
- SEND information (safeguarding profile with developmental concerns, social services, professional referrals)
- Consent records (17 consent fields covering medical treatment, GDPR, safeguarding, photos, outings, social media)
- Collection settings (password-protected collection, over-18 acknowledgement)
- Leaving records (reason-coded departure tracking with lifecycle auditing)
- Room assignment history (current via generated column)
- Safeguarding profile (social services, developmental concerns, professional referrals)
- Billing profile (site rate or custom rate)
- Full multi-step registration wizard on frontend
- Child list with search and filtering on frontend

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Documents & attachments | Cannot upload registration forms, medical letters, consent PDFs | High |
| Child photo/avatar | Profile identification | Low |

**Business Impact:** Core child management is fully operational. The documents gap requires workaround (paper/manual upload outside system).

---

### 2.2 Admissions & Enrolment

**Status: Not Implemented**

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Parent enquiries | No lead capture from website/parent enquiries | Critical |
| Waiting list management | Cannot manage oversubscription or priority lists | Critical |
| Tour scheduling | No tour booking or calendar management | High |
| Registration workflow | No formal staged registration (enquiry → tour → offer → accept → enrol) | Critical |
| Child enrolment | No enrolment pipeline — children are created directly, bypassing admissions | Critical |

**Business Impact:** The admissions pipeline is the primary customer acquisition channel for nurseries. Without it, nurseries cannot:
- Track prospective families
- Manage waiting lists (required by Ofsted for oversubscribed settings)
- Schedule show-rounds
- Convert enquiries into enrolled children
The existing `manager_invites` and `parentchildmappings` modules handle post-enrolment setup (inviting parents, linking to children), but the pre-enrolment funnel is entirely absent.

---

### 2.3 Session Pattern Management

**Status: Completed**

**Implemented Features:**
- Session types (name, start/end time, active/inactive)
- Full-day, half-day, morning, afternoon sessions (configurable via session types)
- Child booking patterns (weekly recurring schedules with day-of-week + session type entries)
- Session templates (reusable schedule templates for quick assignment)
- Pattern versioning (effective_from/effective_to with generated is_current column)
- Frontend UI for managing session types, templates, and individual child booking patterns

**Missing Features:** None identified.

**Business Impact:** Fully operational. Nurseries can define session offerings and assign recurring weekly schedules to children.

---

### 2.4 Room Management

**Status: Completed**

**Implemented Features:**
- Room CRUD (name, description, age group, capacity)
- Room activation/archiving (is_active flag with archive/reactivate use cases)
- Capacity management (capacity field with partial enforcement)
- Age range configuration (age_group text field)
- Child room assignment (start/end dated assignments with generated is_current column)
- Frontend: Room list + form (both manager and owner roles)
- Frontend: Room assignments visible on child detail page

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Capacity validation on assignment | Over-capacity assignments possible | Medium |
| Staff-to-room ratio tracking | Required for EYFS compliance | Medium |

**Business Impact:** Room management is functional. Capacity ratio enforcement would prevent overbooking.

---

### 2.5 Attendance & Registers

**Status: Completed**

**Implemented Features:**
- Check-in/check-out with timestamp and local date tracking
- Open session enforcement (one open session per child at a time)
- Attendance correction workflow (reason-coded corrections with audit trail)
- Absence marking (date-level absence markers with clear workflow)
- Late collection tracking (via check-out timestamps + correction reasons)
- Practitioner attendance view (list children for check-in/check-out)
- Manager attendance corrections view
- Audit trail via attendance_events table

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Daily attendance register PDF | Required for Ofsted inspections | Medium |
| Bulk check-in/check-out | Efficiency for room-based check-in | Low |

**Business Impact:** Core attendance tracking is robust. The correction workflow with reason codes is well-designed for UK nursery compliance.

---

### 2.6 Billing & Invoicing

**Status: Completed**

**Implemented Features:**
- Draft invoice generation with preflight validation
- Invoice lifecycle management (draft → issued → paid/overdue/payment_failed)
- Invoice state machine enforced at database level (trigger-based)
- Issued invoice immutability (lock after issue)
- Invoice lines (core_childcare, funded_deduction, extra, adjustment)
- Monthly billing with per-child invoices
- Bulk invoice generation and issue runs
- Invoice run tracking (generation + issue runs with status/outcome)
- Invoice number sequencing (per branch/month/year)
- Adjustment invoices (credit notes via adjusts_invoice_id)
- Attendance-based billing (actual attended minutes vs booked minutes)
- Site-level hourly rates with per-child custom rate override
- Scheduler for automated invoice runs
- Frontend: Invoice run page, invoice list, invoice detail, parent invoice view

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Fee plans (tuition tiers) | Cannot define standardised fee plans — rates are per-branch or per-child | Low |
| Discounts (sibling, referral, staff) | No discount rules engine | Medium |
| Credit notes workflow | Adjustment invoices exist but no standalone credit note UI | Low |
| Billing reports | No revenue reports or aging summaries | Medium |

**Business Impact:** The billing engine is sophisticated. Actual-minus-funded-minute calculation with funding deduction is the strongest part of the system. A discounts module would reduce manual adjustments.

---

### 2.7 Government Funding

**Status: Completed**

**Implemented Features:**
- Funding eligibility per child (funding_type: fifteen_hours, thirty_hours, two_year_old)
- Funding model (term_time_only or stretched)
- Eligibility code tracking with validation status (e.g., 30-hour code)
- Benefits checklist (benefits_status with benefit_notes, plus benefit checklist migration)
- Monthly funding allowance (funding_profiles table with per-child-month allowance in minutes)
- Funding deduction in invoice calculation (funded_deduction line items on invoices)
- Funding overview page (list by month with totals)
- Evidence received tracking

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| DfE eligibility API integration | Must manually validate eligibility codes | High |
| Automated funding entitlement calculation | Funding allowance is manually entered per month | Medium |
| 2-year-old deprivation check automation | Eligibility checking is manual | Medium |

**Business Impact:** Funding tracking is functional but requires manual data entry for allowance amounts. DfE API integration would reduce administrative burden.

---

### 2.8 Parent Communication

**Status: Not Implemented**

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| In-app messaging (parent ↔ staff) | No parent-staff communication channel | Critical |
| Push notifications | Cannot send real-time alerts (check-in, invoice, announcement) | High |
| Announcements | No broadcast to all parents | High |
| Email notifications | Only invite emails exist — no invoice/attendance notifications | High |

**Business Impact:** Parent communication is a core nursery requirement. Parents expect:
- Check-in/check-out notifications
- Invoice reminders
- Absence reporting (tell staff child is absent today)
- General announcements
Daily nursery operations rely on staff-parent communication.

---

### 2.9 User Management & RBAC

**Status: Partially Completed**

**Implemented Features:**
- User registration and authentication (email + password with JWT tokens)
- Role-based memberships (owner, manager, practitioner, parent)
- Role-based route protection on backend (RequireRoles middleware)
- Role-based route guards on frontend (authGuard + roleGuard + roleDefaultRedirectGuard)
- Role-scoped dashboards (owner overview, manager dashboard, practitioner attendance, parent invoices)
- Invite system (manager invites for practitioners and parents)
- Password reset with token
- Refresh token rotation
- Audit logging (action_entity_type, action_entity_id, actor tracking)

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Roles & permissions CRUD UI | Roles are hardcoded — cannot customise permissions | Medium |
| User profile management | No "my account" or change password page | Medium |
| Staff management | No staff list or role assignment UI (only invite flow exists) | High |
| Session management | No "active sessions" or force logout | Low |
| Permission model | All role checks are hardcoded — no Permission entity | Low |

**Business Impact:** Authentication and basic RBAC work. Staff management needs a dedicated UI (view staff, change roles, deactivate accounts). The existing invite flow covers creation but not ongoing management.

---

### 2.10 Multi-Tenant Platform

**Status: Partially Completed**

**Implemented Features:**
- Tenant table (id, name)
- Branch table (tenant_id, name, is_active, core_hourly_rate_minor)
- Tenant-scoped isolation on every table (tenant_id + branch_id composite keys)
- Tenant extraction from JWT via middleware
- Owner role for tenant-wide management
- Branch CRUD

**Missing Features:**
| Feature | Impact | Priority |
|---|---|---|
| Subscription management | No way to bill tenants for platform usage | Critical |
| Tenant provisioning | No self-service sign-up or tenant creation flow | High |
| Tenant configuration | No configurable tenant settings (branding, locale) | Medium |
| Usage metering | No per-tenant usage tracking | Medium |

**Business Impact:** The tenancy model is architecturally sound but lacks the commercial layer. Subscriptions are needed to monetise the platform.

---

## 3. Functional Gap Analysis

| Module | Current Capability | Missing Capability | Impact | Recommendation |
|---|---|---|---|---|
| Child Management | Complete child profile, contacts, medical, consent, safeguarding | Document/attachment upload | Cannot digitise paper forms | Add file storage module (S3/S3-compatible) with child-scoped document management |
| Admissions & Enrolment | None | Enquiry → Waiting list → Tour → Offer → Enrol pipeline | No customer acquisition funnel | Build admissions module with status pipeline, tour scheduling, and offer management |
| Session Patterns | Full session types, booking patterns, templates | None | — | — |
| Room Management | Room CRUD, capacity, age groups, assignments | Capacity enforcement, staff ratios | Risk of overbooking | Add capacity validation on room assignment; add staff-to-child ratio tracking |
| Attendance & Registers | Check-in/out, corrections, absence, late collection | Daily register PDF export | Ofsted inspection evidence | Add PDF export of daily attendance register |
| Billing & Invoicing | Full invoice lifecycle, funding deduction, payment integration | Fee plan templates, discount rules, credit note UI | Manual billing adjustments | Add fee plan templates (standardised tiers); add sibling/staff discount rules |
| Government Funding | Funding eligibility, monthly allowance, invoice deduction | DfE API integration, automated entitlement calc | Manual code validation | Plan DfE API integration as Phase 2; automate allowance by session type × funding hours |
| Parent Communication | None | Messaging, notifications, announcements | No parent engagement channel | Build in-app messaging + push/email notification system |
| User Management & RBAC | Auth, JWT, roles, invites, audit | Staff management UI, user profile, permission customisation | Limited staff administration | Add staff management page; extend RBAC with granular permissions if business requires |
| Multi-Tenant | Tenant/branch isolation, owner role | Subscriptions, provisioning, tenant config | Cannot monetise | Add subscription tier model + Stripe billing for tenant plans |

---

## 4. Domain Model Analysis

### 4.1 Child Management

**Main Entities:** Child, ChildProfile, ChildContact, ChildHealthProfile, ChildSafeguardingProfile, ChildConsentRecord, ChildFundingRecord, ChildBillingProfile, ChildCollectionSetting, ChildRoomAssignment, ChildLeavingRecord

**Relationships:**
```
Child (1) → (1) ChildProfile
Child (1) → (N) ChildContact (type: parent_carer, emergency_contact, authorised_collector)
Child (1) → (1) ChildHealthProfile (GP, medical, dietary)
Child (1) → (1) ChildSafeguardingProfile (SEND, social services, developmental concerns)
Child (1) → (1) ChildConsentRecord (17 consent fields)
Child (1) → (1) ChildFundingRecord (funding eligibility)
Child (1) → (1) ChildBillingProfile (rate configuration)
Child (1) → (1) ChildCollectionSetting (collection password)
Child (1) → (N) ChildRoomAssignment (dated, only one current)
Child (1) → (1) ChildLeavingRecord (departure reason)
```

**Business Rules:**
- Child must have a unique identity per tenant/branch scope
- Child cannot be modified after leaving record created
- Consent must be signed and dated
- Room assignment dates must not overlap
- Child must have active enrolment to check in
- Contacts ordered by sort_order per contact_type

### 4.2 Admissions & Enrolment

**Status: Not Implemented — proposed model**

**Proposed Entities:** Enquiry, WaitingListEntry, TourBooking, AdmissionOffer, Enrolment

**Proposed Relationships:**
```
Enquiry → WaitingListEntry → TourBooking → AdmissionOffer → Enrolment → Child
```

**Proposed Business Rules:**
- Enquiry must capture parent name, contact, child DOB, required sessions
- Waiting list can be priority-ordered (date, sibling, staff child)
- Tour slots must respect staff availability
- Offer has expiry date
- Enrolment creates Child + initial booking pattern + room assignment

### 4.3 Session Pattern Management

**Main Entities:** SessionType, ChildBookingPattern, ChildBookingPatternEntry, SessionTemplate, SessionTemplateEntry

**Relationships:**
```
Branch (1) → (N) SessionType
Child (1) → (N) ChildBookingPattern (only one current)
ChildBookingPattern (1) → (N) ChildBookingPatternEntry (day_of_week + session_type_id)
SessionTemplate (1) → (N) SessionTemplateEntry (day_of_week + session_type_id)
```

**Business Rules:**
- Only one active booking pattern per child at a time (is_current = true)
- Session type times must not overlap within a day (enforced at application layer)
- Session type start < end time
- Day of week 1-7 (Monday-Sunday)
- One entry per day_of_week + session_type per pattern

### 4.4 Room Management

**Main Entities:** Room, ChildRoomAssignment

**Relationships:**
```
Branch (1) → (N) Room
Room (1) → (N) ChildRoomAssignment
Child (1) → (N) ChildRoomAssignment
```

**Business Rules:**
- Room name unique per branch (active rooms only)
- Capacity is advisory (not enforced at DB level)
- One current room assignment per child (is_current = true)
- Assignment dates must not overlap for same child

### 4.5 Attendance & Registers

**Main Entities:** AttendanceSession, AttendanceEvent, AbsenceMarker

**Relationships:**
```
Child (1) → (N) AttendanceSession (one per check-in)
AttendanceSession (1) → (N) AttendanceEvent (check_in, check_out, correction)
Child (1) → (N) AbsenceMarker (one per date when marked)
```

**Business Rules:**
- One open session per child at any time (unique partial index)
- Check-out must be after check-in
- Corrections require reason code (missed_check_in, missed_check_out, incorrect_time, duplicate_entry, other)
- Absence markers are per-child per-date, unique when active
- Sessions are child-scoped, not room-scoped

### 4.6 Billing & Invoicing

**Main Entities:** Invoice, InvoiceLine, InvoiceRun, InvoiceNumberSequence, PaymentAttempt, PaymentReconciliationRecord

**Relationships:**
```
Child (1) → (N) Invoice (one per month, kind: monthly/adjustment)
Invoice (1) → (N) InvoiceLine (core_childcare, funded_deduction, extra, adjustment)
InvoiceRun (1) → (N) Invoice (generated or issued)
Invoice (N) → (N) PaymentAttempt (1) → (N) PaymentReconciliationRecord
Invoice (0..1) → Invoice (adjustment chain)
```

**Business Rules:**
- Invoice state machine: draft → issued → paid/overdue/payment_failed (enforced by trigger)
- Issued invoices are immutable (header fields cannot change; lines cannot be inserted/updated/deleted)
- One monthly invoice per child per month
- Invoice amounts in GBP minor units (pence)
- Paid invoices cannot transition
- Credit notes are adjustment invoices linked to original
- Payment reconciliation records stripe webhook events

### 4.7 Government Funding

**Main Entities:** ChildFundingRecord, FundingProfile

**Relationships:**
```
Child (1) → (1) ChildFundingRecord (eligibility)
Child (1) → (N) FundingProfile (monthly allowance)
```

**Business Rules:**
- Funding types: none, fifteen_hours, thirty_hours, two_year_old, custom
- Funding models: term_time_only, stretched
- Funding allowance capped at 44,640 minutes/month (31 days × 24 hrs × 60 min)
- One funding profile per child per billing month

### 4.8 Parent Communication

**Status: Not Implemented — proposed model**

**Proposed Entities:** Conversation, Message, Notification, NotificationPreference

**Proposed Relationships:**
```
Branch (1) → (N) Conversation (staff-parent thread)
Conversation (1) → (N) Message
Child (N) → (N) Conversation
User (1) → (N) Notification
User (1) → (1) NotificationPreference
```

### 4.9 User Management & RBAC

**Main Entities:** User, Membership, RefreshToken, PasswordResetToken, ManagerInvite

**Relationships:**
```
User (1) → (N) Membership (per tenant/branch with role)
User (1) → (N) RefreshToken (per membership)
User (1) → (N) PasswordResetToken
Branch (1) → (N) ManagerInvite
Membership (1) → (N) ParentMembershipChild (parent → children mappings)
```

**Business Rules:**
- Owner membership has no branch_id (tenant-wide)
- Manager/practitioner/parent memberships require branch_id
- One owner per tenant
- One active membership per user per branch
- Active memberships cannot have ended_at set
- Invite flows: manager creates invite → user accepts → membership created

### 4.10 Multi-Tenant Platform

**Main Entities:** Tenant, Branch

**Relationships:**
```
Tenant (1) → (N) Branch
Tenant (1) → (N) Membership (owner, managers, practitioners, parents)
```

**Business Rules:**
- Every data row is tenant-scoped via tenant_id
- Branch names are unique per tenant
- Branches can be deactivated (is_active flag)
- Cross-tenant data access is structurally impossible (tenant_id in every FK)

---

## 5. Core Business Workflow Validation

### Workflow: Child → Session Pattern → Room Assignment → Attendance → Funding Allocation → Invoice Generation → Payment Collection

```
Child
  │
  ├── Step 1: Create child record
  │     Module: Child Management ✅
  │
  ├── Step 2: Define session pattern
  │     Module: Session Pattern Management ✅
  │     Dependency: Session types must exist (configured per branch)
  │
  ├── Step 3: Assign room
  │     Module: Room Management ✅
  │     Dependency: Room must exist and be active
  │
  ├── Step 4: Track attendance
  │     Module: Attendance & Registers ✅
  │     Dependency: Valid enrolment (active child, start_date ≤ today ≤ end_date)
  │     Failure point: Check-in does not validate room assignment or session pattern
  │
  ├── Step 5: Allocate funding
  │     Module: Government Funding ✅
  │     Dependency: ChildFundingRecord must exist with funding_enabled = true
  │     Failure point: Funding allowance is manually set per month in funding_profiles
  │
  ├── Step 6: Generate invoice
  │     Module: Billing & Invoicing ✅
  │     Dependency: Active term (booking pattern + rate + dates)
  │     Dependency: Attendance records for the billing month
  │     Dependency: Funding profiles for the billing month
  │     Failure point: If no term exists, child is skipped in invoice run
  │
  └── Step 7: Collect payment
        Module: Billing & Invoicing (invoice) + Payments (Stripe) ✅
        Dependency: Invoice must be issued (not draft)
        Dependency: Stripe account must be configured
        Failure point: No retry logic for failed payments beyond manual reissue
```

### Missing Workflow Dependencies

| Dependency | Missing? | Risk |
|---|---|---|
| Enquiry → Child creation | Yes (no admissions module) | Children are created without a formal application pipeline |
| Session pattern validation on check-in | Partial | Check-in does not verify child has a valid session pattern for the day |
| Room capacity check on assignment | Yes | Over-capacity not prevented |
| Funding eligibility auto-calculation | Yes | Must manually enter funding allowance minutes |
| Payment retry/reminder automation | No | No dunning or automated payment retry |
| Invoice notification to parents | Yes | No notification when invoice is issued |

---

## 6. Technical Architecture Considerations

### 6.1 Backend Architecture

**Domain Boundaries (current):**

| Module | Domain | Application | Infrastructure | HTTP |
|---|---|---|---|---|
| children | ✅ | ✅ | ✅ | ✅ |
| attendance | ✅ | ✅ | ✅ | ✅ |
| absence | ✅ | ✅ | ✅ | ✅ |
| billing | ✅ | ✅ | ✅ | ✅ |
| funding | ✅ | ✅ | ✅ | ✅ |
| payments | ✅ | ✅ | ✅ | ✅ |
| rooms | ✅ | ✅ | ✅ | ✅ |
| sessiontypes | ✅ | ✅ | ✅ | ✅ |
| sessiontemplates | ✅ | ✅ | ✅ | ✅ |
| term | ✅ | ✅ | ✅ | ✅ |
| authentication | ✅ | ✅ | ✅ | ✅ |
| invites | ✅ | ✅ | ✅ | ✅ |
| parentchildmappings | ✅ | ✅ | ✅ | ✅ |
| owner | ✅ | ✅ | ✅ | ✅ |
| passwordreset | ✅ | ✅ | ✅ | ✅ |
| invoicerun | — | — | — | — (scheduler only) |

**Service Responsibilities:**
- Children module is the largest — responsible for child, profile, contacts, health, safeguarding, consent, funding, billing, collection settings, room assignments, booking patterns, leaving records
- This creates a broad domain boundary; consider splitting if the module grows further

**API Requirements:**
- RESTful with JSON
- All endpoints scoped to tenant/branch via JWT claims
- Standardised error responses via domain errors mapped to HTTP
- Audit logging via middleware and explicit audit writer calls

**Database Considerations:**
- PostgreSQL with pgx driver
- Composite PKI pattern (tenant_id, branch_id, id) for every scoped table
- Generated columns for computed state (is_current on room assignments, booking patterns)
- Trigger-based invoice state machine enforcement
- Partial unique indexes for active-only constraints
- JSONB for flexible fields (address, referrals, calculation_details)

### 6.2 Database Schema

**Main Tables by Domain:**

| Domain | Tables |
|---|---|
| Multi-Tenant | tenants, branches |
| Auth/RBAC | users, memberships, refresh_tokens, password_reset_tokens, manager_invites |
| Child | children, child_profiles, child_contacts, child_health_profiles, child_safeguarding_profiles, child_consent_records, child_funding_records, child_billing_profiles, child_collection_settings, child_leaving_records |
| Rooms | rooms, child_room_assignments |
| Sessions | session_types, child_booking_patterns, child_booking_pattern_entries, session_templates, session_template_entries |
| Term | term, term_schedule_change |
| Attendance | attendance_sessions, attendance_events, absence_markers |
| Billing | invoices, invoice_lines, invoice_number_sequences, invoice_runs, invoice_run_advance |
| Payments | payment_attempts, payment_reconciliation_records, stripe_webhook_events |
| Audit | audit_logs |
| Parent | parent_membership_children |

**Missing Tables:**

| Table | Domain | Purpose |
|---|---|---|
| child_documents | Child | Document/attachment storage |
| enquiries | Admissions | Parent enquiry capture |
| waiting_list_entries | Admissions | Waiting list management |
| tour_bookings | Admissions | Tour scheduling |
| admission_offers | Admissions | Offer management |
| conversations | Communication | Staff-parent messaging |
| messages | Communication | Individual messages |
| notifications | Communication | Push/email notifications |
| notification_preferences | Communication | Per-user notification settings |
| subscriptions | Multi-Tenant | Tenant billing plans |
| tenant_configurations | Multi-Tenant | Per-tenant settings |
| fee_plans | Billing | Standardised fee templates |
| discounts | Billing | Discount rules |

**Key Constraints:**
- Scoped uniqueness via composite indexes (tenant_id, branch_id, ...)
- CHECK constraints for status enums and business rules
- FOREIGN KEY constraints cascade through scoped references
- Invoice immutability enforced via trigger functions

### 6.3 Security

**RBAC:**
- Roles: owner, manager, practitioner, parent
- Enforced at API gateway via middleware (RequireRoles)
- Frontend route guards match backend roles
- Owner has tenant-wide access (no branch_id constraint)
- All other roles are branch-scoped
- One owner per tenant enforced via unique partial index

**Tenant Isolation:**
- Every table includes tenant_id (and where applicable branch_id)
- All foreign keys are scoped (tenant_id, branch_id, local_id)
- JWT contains tenant_id, branch_id, membership_id
- Data access is structurally restricted — no cross-tenant leaks possible
- Isolation is enforced at database level, not just application level

**Audit Logging:**
- Central audit_logs table with action_type, entity_type, entity_id, actor
- Lifecycle reason codes for entity lifecycle events
- Request ID tracking for end-to-end traceability
- Triggered from application use cases (not DB triggers)

**Data Privacy (UK Considerations):**
- GDPR: consent_records include information_truthfulness_declaration, signer_name, signed_date
- Child data protection: all tables include audit for data access/modification
- Personal data: contacts store email, telephone, address — all audited
- Data retention: leaving records track child departure with reason codes
- Parental responsibility: child_contacts track has_parental_responsibility flag

### 6.4 UK Compliance Considerations

**Current Compliance Coverage:**
- ✓ EYFS: Safeguarding profile with social services, developmental concerns tracking
- ✓ EYFS: Staff-to-child ratios (accessible via rooms.capacity, not enforced)
- ✓ GDPR: Consent records with explicit consent fields for data processing
- ✓ GDPR: Information-sharing consent tracking
- ✓ Ofsted: Attendance records with check-in/out timestamps
- ✓ Ofsted: Absence tracking with clearance workflow
- ✓ DfE: Early years funding — fifteen_hours, thirty_hours, two_year_old
- ✓ DfE: Funding eligibility codes with benefits checking
- ✓ GDPR: Leaving records with audit trail

**Compliance Gaps:**
- ⚠ EYFS: Ratio enforcement needs staff-to-room assignment
- ⚠ EYFS: Outdoor play and risk assessment tracking
- ⚠ Safeguarding: Child protection case management
- ⚠ GDPR: Data retention policy enforcement (automated purging)
- ⚠ Ofsted: Self-Evaluation Form (SEF) tracking
- ⚠ Ofsted: Incident/accident reporting

---

## 7. Implementation Roadmap

### Sprint 1 — Foundation (Weeks 1-2)

**Scope:** Revenue-enabling + compliance-gap modules.

| Module | Items | Dependencies |
|---|---|---|
| Admissions | Enquiry, waiting list, tour booking, offer, enrolment pipeline | None |
| Parent Communication | Conversations, messages, notifications (email + push) | None |
| Multi-Tenant | Subscription table + Stripe billing for tenants | Branches |

**Risks:**
- Admissions module is the largest new build — may slip
- Subscription billing requires Stripe Connect or similar

**Expected Outcome:**
- Nurseries can manage the full parent journey from enquiry to enrolment
- Parents can receive notifications
- Platform operator can charge for usage

### Sprint 2 — Child Management Enhancement (Weeks 3-4)

**Scope:** Remaining child features.

| Module | Items | Dependencies |
|---|---|---|
| Child Management | Document/attachment upload (S3-compatible) | Sprint 1 |
| Admissions & Enrolment | Enrolment creates full child profile + initial term | Sprint 1 |

**Risks:**
- File upload introduces storage costs and security considerations
- Document virus scanning requirements

**Expected Outcome:**
- Complete digital child file including uploaded documents
- Admission-to-active-child pipeline is end-to-end

### Sprint 3 — Staff & Operations Enhancement (Week 5)

**Scope:** Staff management, compliance, operations.

| Module | Items | Dependencies |
|---|---|---|
| User Management | Staff list, profile editing, role changes, deactivation | None |
| Room Management | Capacity enforcement on assignment, staff assignment | Sprint 1 |
| Attendance | Daily register PDF export | None |

**Risks:**
- Low technical risk — incremental enhancements

**Expected Outcome:**
- Operational staff management without invites-only flow
- Ofsted-ready attendance register export

### Sprint 4 — Financial Enhancement (Week 6)

**Scope:** Revenue features and reporting.

| Module | Items | Dependencies |
|---|---|---|
| Billing | Fee plan templates, discount rules, credit note UI | None |
| Government Funding | DfE eligibility API integration (scoped) | Sprint 1 |
| Multi-Tenant | Tenant configuration (branding, locale) | Sprint 1 |
| Reporting | Basic financial reports (revenue, aged receivables) | None |

**Risks:**
- DfE API integration requires DfE registration and accreditation
- Credit notes need careful accounting treatment

**Expected Outcome:**
- Standardised fee plans reduce manual rate configuration
- Discount rules automate common adjustments
- Basic financial reporting for nursery management

### Sprint 5 — Polish & Testing (Weeks 7-8)

**Scope:** Quality, performance, documentation.

| Item | Description |
|---|---|
| End-to-end testing | Full workflow from enquiry to payment |
| Performance testing | Multi-tenancy under load |
| API documentation | OpenAPI spec completeness |
| User documentation | Nursery staff guides |
| Security audit | Penetration testing, data privacy review |

---

## 8. MVP Completion Criteria

### Phase 1 MVP is complete when a nursery can:

| Capability | Current Status | Target Sprint |
|---|---|---|
| ✓ Register a nursery (tenant + branch) | ✅ Done | — |
| ✓ Add staff members (manager, practitioner) | ✅ Done (invite flow) | — |
| ✓ Create child records with full profile | ✅ Done | — |
| ✓ Record parent/guardian information | ✅ Done | — |
| ✓ Record medical and allergy information | ✅ Done | — |
| ✓ Record consent and GDPR permissions | ✅ Done | — |
| ✓ Upload child documents and consent forms | ❌ Missing | Sprint 2 |
| ✓ Accept parent enquiries and manage waiting list | ❌ Missing | Sprint 1 |
| ✓ Schedule show-round tours | ❌ Missing | Sprint 1 |
| ✓ Enrol children through a defined pipeline | ❌ Missing | Sprint 1 |
| ✓ Define session types (full-day, half-day, etc.) | ✅ Done | — |
| ✓ Assign recurring weekly schedules to children | ✅ Done | — |
| ✓ Set up rooms with age ranges and capacity | ✅ Done | — |
| ✓ Assign children to rooms | ✅ Done | — |
| ✓ Check children in and out daily | ✅ Done | — |
| ✓ Track absences | ✅ Done | — |
| ✓ Correct attendance records | ✅ Done | — |
| ✓ Set up government funding eligibility | ✅ Done | — |
| ✓ Allocate funded hours per month | ✅ Done | — |
| ✓ Generate monthly invoices | ✅ Done | — |
| ✓ Apply funding deductions to invoices | ✅ Done | — |
| ✓ Issue invoices to parents | ✅ Done | — |
| ✓ Collect payments via Stripe | ✅ Done | — |
| ✓ View invoices in parent portal | ✅ Done | — |
| ✓ Communicate with parents via the platform | ❌ Missing | Sprint 1 |
| ✓ Notify parents of invoices and check-outs | ❌ Missing | Sprint 1 |
| ✓ Manage staff accounts and roles | ⚠ Partial | Sprint 3 |
| ✓ Maintain audit trail of all changes | ✅ Done | — |

### Formal Definition of Done

All of the above capabilities exist as:
1. Database schema with appropriate constraints
2. Domain logic with business rule enforcement
3. REST API endpoints with authentication and authorisation
4. Frontend UI pages for the relevant roles
5. Tests covering core use cases
6. API documentation via OpenAPI spec

---

## 9. Risks & Recommendations

### 9.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Children module becoming a monolith | Medium | Maintainability | Monitor aggregate boundaries; split if children module exceeds ~50 files |
| File storage costs | Low | Operational | Use S3-compatible storage with presigned URLs; enforce file size limits |
| Stripe integration complexity | Low | Payment flow | Already well-structured with webhook reconciliation |
| Invoice performance at scale | Medium | Billing month-end | Invoice generation is already async via invoice runs; optimise queries if needed |
| DfE API integration complexity | Medium | Funding automation | DfE API is government-grade; budget for integration overhead |

### 9.2 Business Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Admissions pipeline gap delays go-live | High | Customer acquisition | Prioritise admissions module in Sprint 1 |
| No parent communication reduces stickiness | High | Churn | Build MVP communication (email notifications) before in-app messaging |
| Subscription billing not implemented | High | Revenue | Implement Stripe-based tenant billing early |
| Competitors offer funding API integration | Medium | Competitive disadvantage | Scope DfE integration as high-priority Phase 2 |
| No mobile app | Medium | Market expectation | Phase 2 consideration; responsive web app covers near term |

### 9.3 Data Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Child data breach | Low | Critical | All data is tenant-scoped; audit trail on all modifications; encryption at rest |
| Data loss on migration | Low | Critical | Migration scripts are idempotent; test in staging first |
| GDPR data purging not automated | Medium | Compliance | Add data retention scheduler |
| UK GDPR right to erasure compliance | Medium | Legal | Child deletion workflow (with audit trail, not hard delete) exists via leaving records |

### 9.4 Scaling Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Invoice generation as single tenant scales (500+ children) | Medium | Performance | Invoice generation is batch-run per branch; horizontal scaling via stateless API |
| Concurrent check-in at peak time | Low | Performance | Unique partial index prevents duplicate open sessions |
| Multi-region data residency (Welsh/Scottish nurseries) | Low | Legal | Current single-region DB; evaluate if non-English settings are targeted |

### 9.5 Strategic Recommendations

1. **Complete the admissions pipeline before any other feature.** Without it, nurseries cannot use the system for customer acquisition. This is the highest business-impact gap.

2. **Add email notifications early.** Parent communication doesn't need a full messaging UI initially. Invoice notifications, check-in/out notifications, and absence alerts can be transactional emails triggered from existing use cases.

3. **Implement Stripe billing for tenant subscriptions.** The platform needs a revenue model. Use the existing Stripe integration patterns for tenant subscription management.

4. **Defer DfE API integration to Phase 2.** The manual eligibility code validation works for MVP. Government API integration is high-effort and can wait.

5. **Invest in capacity enforcement.** Room capacity and staff ratio enforcement are Ofsted-relevant compliance features that reduce manual oversight.

6. **Consider module splitting for the children package.** With ~15 domain entities and ~30 application use cases, the children module is approaching a size where separation into sub-domains (child-core, child-health, child-funding, child-billing) may improve maintainability.

7. **Avoid over-engineering RBAC.** The current hardcoded role checks work well for 4 roles. Only build a dynamic permission system if nursery-specific permission requirements emerge.

---

## Appendix A: Complete File Inventory

| Layer | Files | Coverage |
|---|---|---|
| DB Migrations | 3 up/down pairs | Full schema |
| DB Queries (sqlc) | 33 .sql files | Generated Go code |
| API Modules | 15 Go modules | Domain + App + Infra + HTTP |
| API Bootstrap | 471 lines | Full wiring |
| Frontend Pages | ~40 components | Role-based routing |
| Frontend Guards | 4 guards | Auth + role enforcement |

## Appendix B: Database Object Count

| Object Type | Count |
|---|---|
| Tables | 31 |
| Indexes | ~60 |
| Foreign Keys | ~55 |
| Triggers | 6 |
| Functions | 6 |
| Enums | 2 |
