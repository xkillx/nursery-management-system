# Context Glossary

## Pilot Nursery

The first live customer setting used to validate the initial release in production.

## Tenant

A single nursery business boundary that owns its own data and users.

## Nursery Site

A physical nursery location within a tenant. Product-facing docs and user-facing language should prefer site, while existing engineering and API contracts may continue to use branch for the same location boundary unless a later decision separates the concepts.

## Branch Scope

The existing technical and API scope for a nursery site. Branch scope appears in session membership, authorization, API contracts, and database records, and should be treated as the same location boundary as nursery site in roadmap planning.

## Release Roadmap

The shipped operational baseline covers attendance, absence markers, funding deduction, invoicing, parent invoice access, Stripe payment collection, and manager/practitioner/parent access, and is followed by a stabilization track that extends it with owner and multi-site operations, registration and consent, room/session planning, ratio safety, safeguarding and incidents, learning journeys, and production hardening. The default next workstream is pilot readiness and owner-visible four-site operations, after which feature work proceeds through registration and consent, room/session planning, ratio safety, safeguarding and incidents, then learning journeys; registration and consent come first because child profile and compliance data feed later planning and safety workflows.

## Child Identity Source of Truth

The existing child record remains the source of truth for a child's full name and date of birth. The registration/enrolment profile may display those values for context but does not keep a second editable copy of them.

## Collection Password

A sensitive collection secret recorded for emergency child collection when the nursery has prior notice that someone outside the usual listed contacts will collect the child. Normal registration profile reads expose only password presence and last-updated metadata, not the stored secret itself.

## User Role: Owner

The business owner or delegated owner-level operator responsible for oversight across the nursery business and its sites. A tenant may have multiple owner-role users; the pilot may start with one owner account.

## Owner Oversight Scope

The first owner release is cross-site read and oversight first. The owner may inspect site summaries, compare finance and attendance health, review pilot readiness, and administer branch-manager access, while routine attendance, child-record edits, invoice issue, and branch workflow corrections remain branch-manager responsibilities unless a later owner-administration decision expands them.

## Owner Access Mode

The owner works from a business-wide owner view across all nursery sites, with site filters and drill-downs where needed. Owner access is distinct from switching into or impersonating a branch-scoped manager session.

## Owner Manager Access Administration

In the first owner release, the owner may invite, activate or deactivate, and assign manager access for each nursery site. The owner may grant site-manager access to a new email address or to an existing active user account without changing that user's other memberships. Practitioner and parent access administration remains a site-manager responsibility.

## Owner Role Provisioning

Owner-role users are provisioned through administrative or bootstrap paths in the first owner release. Owners may manage site-manager access, but they do not invite, assign, or deactivate other owner-role users.

## Owner Cross-Site Summary

The owner-facing summary compares nursery sites using aggregate health metrics and exception counts. The first owner landing view is exception-led: sites without active managers, overdue or failed payments, outstanding balances, incomplete attendance, and funding readiness flags are more prominent than routine detail. It does not grant routine drill-down into named child, parent, invoice, or attendance workflows in the first owner release.

## Owner Summary Export Boundary

The first owner release provides on-screen cross-site summaries only. Exportable owner reporting is deferred until reporting requirements are clearer.

## Owner Site Health Metrics

The first owner summary covers attendance today, active enrolled children, invoice and payment health, funding readiness exceptions, and manager access status per site. Invoice and payment health includes invoice counts by status plus financial totals such as issued, paid, outstanding, overdue outstanding, and failed-payment count. Staff costs, Ofsted readiness, room capacity, safeguarding, and learning journey metrics are deferred until those product areas exist.

## Owner Site Setup Visibility

Owner summaries include active sites even when setup or current-period data is missing. A site without an active manager is an owner action item; missing finance, funding, or attendance data is shown as zero or not ready according to the metric rather than hiding the site.

## Owner Site Billing Setup

Owner site overview is the first owner workflow for setting each nursery site's core hourly rate. A missing site core hourly rate is a site setup exception because future draft invoice generation for that site is not billing-ready.

## Owner Summary Time Windows

Owner attendance summaries use today's site attendance. Owner invoice, payment, and funding readiness summaries default to the current calendar billing month, while active enrolled children and manager access status are current snapshots.

## Owner Site Coverage

Owner access applies to all active nursery sites within the owner's tenant. The current pilot has four sites, but four-site access is a pilot shape rather than a permanent product limit.

## Owner Site Filter Boundary

Filtering or focusing an owner view to one nursery site changes what the owner is inspecting, not what the owner is allowed to do. Avoid using site switcher to imply that the owner enters or impersonates a branch-scoped manager or practitioner session. Owners remain blocked from branch-scoped manager write actions such as child or contact edits, attendance corrections, invoice generation or issue, funding profile management, and practitioner attendance actions.

## User Role: Manager

A nursery staff role responsible for administration, invoicing, and operational oversight.

## Multi-Site Manager Access

A manager assigned to more than one nursery site works through separate manager memberships and an active site scope. Multiple manager memberships do not grant owner-style cross-site summaries or tenant-wide site filters.

## User Role: Practitioner

A nursery staff role focused on day-to-day child attendance operations.

## User Role: Parent

A parent-side role that views invoices and completes payments. The relationship between the parent membership and the children they can act on is recorded directly on parent_membership_children, not through a separate guardian entity.

## Manager Operations Dashboard

The manager-facing landing surface for operational oversight across attendance, invoicing, and payment status. Avoid generic dashboard or ecommerce dashboard.

## Membership

A user's role-bearing participation in one tenant and one branch. A user may have multiple active memberships, but each authenticated session acts through exactly one membership.

## Role Capability Inheritance

Manager permissions include practitioner attendance actions within the same active session scope.

## Funding v1

A simple funded-hours deduction model used to reduce monthly billed amount per child.

## Funding Profile

A child's funded-hours allowance for a single billing month.

## Funded-Hours Allowance

The amount of core childcare time covered by funding for one child in one billing month. Managers discuss and enter this as hours and minutes, while a missing allowance remains different from an explicit zero allowance.

## Missing Funding Profile

A child-month without a recorded funding profile. This is distinct from a funding profile with a zero-minute allowance.

## Missing Funding Profile Invoice Block

A missing funding profile blocks invoice draft preflight for that child-month; an explicit zero-minute allowance does not.

## Funding Overview

A manager-facing triage surface for a selected billing month that highlights child-month funding profiles needing review. It supports funding readiness checks but does not replace invoice preflight or change invoice blocking rules.

## Funding Overview Triage Flag

A warning shown in the funding overview for a missing funding profile, an explicit zero-minute allowance, an allowance under one hour, or an allowance above 160 hours. These flags are advisory and do not block saving funding profiles or generating invoices.

## Funding Profile Enrollment Scope

Funding profiles are valid for billing months that overlap the child's enrollment window, including historical months after the child has left.

## Funding Profile Partial-Month Rule

A funding profile may exist when the billing month overlaps the child's enrollment window by at least one day.

## Funding Profile Audit Scope

Funding profile creation and allowance changes are persisted as audit events, while unchanged idempotent saves are not.

## Funding Profile Invoice Snapshot Boundary

Changing a funding profile after an invoice is issued does not change that issued invoice; issued invoice lines preserve the funded allowance applied at issue time.

## Funding v1 Metadata Scope

Funding v1 records only the month-specific allowance; funding source, entitlement type, evidence, stretched funding, term-time rules, and notes are deferred.

## Invoice

A monthly billing statement showing gross fees, funded deduction, and final amount due.

## Monthly Invoice

An invoice for one child and one calendar billing month. A child has at most one monthly invoice for the same billing month within a nursery branch.

## Invoice Line

A charge, deduction, extra, or adjustment entry that explains how an invoice total was formed. Invoice lines preserve the billing calculation used for the invoice they belong to.

## Parent Account Provisioning

Parent user accounts are created by manager invitation only; public self-signup is not used in the current release.

## Password Reset Request Privacy

Password reset requests do not reveal whether an email address belongs to an account; the user sees the same accepted outcome either way.

## Password Reset Link State

Invalid, expired, and already-used password reset links are unusable link states; the user must request a fresh reset link or return to sign-in.

## Manager Invite Access Grant Timing

A manager invitation does not grant usable access by itself; invited access becomes usable only after the invitee accepts the invitation and sets a password.

## Manager Invite Membership Activation Scope

Accepting a manager invitation creates a new active membership for a new login identity; invitation acceptance does not reactivate existing inactive users or inactive memberships in the current release.

## Manager Invite Session Start

Accepting a manager invitation does not automatically start an authenticated session; invitees sign in through the normal login flow after acceptance.

## Parent Membership Child Mapping

A relationship showing that a parent-role membership in a given tenant+branch is attached to a specific child. A parent membership may have multiple active child mappings, and a child may have multiple parent memberships attached. Active mappings grant parent portal access for that child; ended mappings do not.

## Parent Mapping Change Flow

A single parent membership can be mapped to many children. Each (membership_id, child_id) pair is independent; mapping the same membership to a different child while one mapping is already active is allowed and does not error.

## Parent Membership End Cascade

Ending a parent membership ends all of its active child mappings in the same action so no dangling active mapping remains. The cascade is implemented as a database trigger on the memberships table.

## Parent Membership End Cascade Reason Attribution

When parent membership end cascades to end an active child mapping, the mapping stores an explicit system cascade reason code so automatic effects are distinguishable from direct manager-initiated end actions.

## Parent Mapping End Visibility Rule

When a parent-membership-to-child mapping is ended, that parent immediately loses access to that child's invoices, including historical invoices. Access is authorized against current active mappings at request time.

## Parent Mapping Idempotent Create

Creating a parent-membership-to-child mapping for a pair that is already active is treated as idempotent success, returning the existing row; no conflict is raised.

## Parent Mapping Active-Entity Requirement

Creating an active parent-membership-to-child mapping requires the parent membership to be active and the child to exist in the same tenant+branch scope. Role must be parent.

## Contact Detail Scope

The child record now carries the full parent/contact set directly, on child_contacts rows of type parent_carer, emergency_contact, or authorised_collector. There is no longer a separate guardians table; rich profile/contact modeling on the child is deferred until after pilot validation.

## Child Parent Carer Contact Requirement Enforcement

A child is enrollment-complete when the child record has at least one child_contacts row of type parent_carer in the same tenant+branch. The blocker is 'no parent_carer contact' rather than the previous 'no active guardian link'. Site-level billing setup is checked separately through the nursery site's core hourly rate.

## Child and Contact Management Lifecycle

A child may have many contact records (parent_carers, emergency contacts, authorised collectors). The parent_carer subset is required for enrollment completeness; the other contact types are operational details.

## Relationship End Reason Requirement

Ending a parent-membership-to-child mapping requires a reason code from the lifecycle_reason_code vocabulary, with a free-text note required when reason_code is 'other'.

## Relationship End Reason Shape

The end action writes ended_at, ended_reason_code, and ended_reason_note; the check constraint enforces the same shape used by every other ended relationship in the system.

## Lifecycle Reason Vocabulary

The shared reason codes (duplicate_record, entered_in_error, left_nursery, safeguarding_direction, contact_update, access_revoked, other) apply to child mappings as to every other lifecycle-ending relationship. No new values are added.

## Lifecycle Other-Reason Note Requirement

When the reason code is 'other', a non-empty trimmed reason note is required. The constraint is enforced at the database level.

## Relationship End Terminology

'End' (not 'delete') is the term used for the lifecycle action on a parent-membership-to-child mapping; the row is preserved with ended_at set, and a partial unique index keeps the active-set membership_id+child_id uniqueness invariant.

## Site Core Hourly Rate

The GBP hourly fee for standard core childcare attendance at one nursery site before funded-hours deduction. It applies to every child in that site for this feature; child-specific discounts, exceptions, and special fees are outside the site core hourly rate.

## Site Billing Setup Authority

Owner-level users set the site core hourly rate for each nursery site. Site managers may view the rate used for billing in their site, but pricing policy belongs to owner-level setup for this feature.

## Site Core Hourly Rate Introduction

When site core hourly rates are introduced, a nursery site with one consistent existing child hourly rate may use that value as its initial site rate. A site with no existing rate or multiple existing child rates requires owner setup before future draft invoice generation; existing issued invoices keep their saved billing snapshots.

## Current Site Core Hourly Rate

For this feature, a nursery site has one current core hourly rate rather than an effective-dated rate schedule. Draft invoice generation and regeneration use the site's current positive rate at the time of generation, while issued invoices keep their saved calculation snapshot.

## Site Core Hourly Rate Readiness

A nursery site is billing-ready only when its site core hourly rate is set to a positive GBP amount. Missing, invalid, or zero site rates are setup exceptions and block future draft invoice generation.

## Site Rate Change Draft Boundary

Changing a nursery site's core hourly rate does not automatically recalculate existing draft invoices. Managers must run invoice generation or regeneration again for draft invoices to pick up the changed site rate; issued invoices remain locked to their saved snapshots.

## Site Core Hourly Rate Audit

Owner changes to a nursery site's core hourly rate are audit-significant because they affect future billing. The first release records rate changes for accountability but does not require an owner-facing rate history screen.

## Child Detail/Enrollment Surface

A manager-facing child-focused view for inspecting whether a child has the minimum data needed for attendance and child-linked invoicing setup. It separates child enrollment completeness from site-level billing setup and month-specific funding profile readiness.

## Child-Specific Core Rate Boundary

Managers do not enter or maintain core hourly rates on individual child records for this feature. Child create, edit, registration-linked setup, and enrollment completeness should not treat a child-specific hourly rate as required, while manager-facing billing context may show the nursery site's applicable core hourly rate.

## No Hard-Delete Core Records

Child, attendance, and invoice records are retained during the current release; manager workflows may end, deactivate, correct, or supersede records, but do not permanently delete them.

## Manager Provisioning Authority

Manager role assignment is reserved to administrative bootstrap flows in the current release; manager-invited users are limited to non-manager roles.

## Attendance Record

Attendance is captured as an event history (check-in, check-out, correction) instead of a single mutable interval row. Use check-in and check-out rather than sign-in or sign-out.

## Attendance Session

An attendance session is one continuous period of a child's attendance, beginning with check-in and ending with check-out.

## Attendance Open Session Rule

A child may have multiple attendance sessions on the same local day, but may have only one open attendance session at a time. Avoid the phrase active check-in for this rule.

## Session Type

A reusable, site-scoped named time block that defines when a planned childcare session runs, such as Morning 08:00–13:00 or Full Day 08:00–18:00. It is configured per nursery site and selected into a child's booking pattern, and it is distinct from an Attendance Session, which records actual attendance.
_Avoid_: Session (bare), Session Template

## Booking Pattern

A child's planned weekly attendance expressed as day-of-week and session-type entries, valid over an effective date range. Only one booking pattern is active for a child on any given date, historical patterns are retained rather than overwritten, and it records expected attendance rather than actual attendance.
_Avoid_: Session Pattern (tolerated alias only)

## Session Pattern (user-facing label)

**Session pattern** is the user-facing label for the Booking Pattern concept. The API, DB, code, and documentation use the canonical internal name **booking pattern**; the wizard and standalone pages surface the alias "Session pattern" with a short subtitle clarifying the relationship. A child can never be created via the registration wizard without a non-empty pattern.

## Booked Session

One day-of-week plus session-type entry within a booking pattern, representing that a child is expected to attend a specific session type on a specific weekday. A child may have multiple booked sessions on the same day.
_Avoid_: session slot, pattern item

## Session Template

A named, per-site, reusable week of booked sessions (a set of day-of-week and session-type entries). Templates are reference data, not enrolled patterns: a child booking pattern **copies** the template's entries at creation time and is independent thereafter. Editing or archiving a template never alters historical patterns.

## Booking Pattern Billing Boundary

Booking patterns record expected attendance only and do not drive billing in the current release; monthly invoices continue to derive billable minutes from attendance actuals, and funding, invoicing, daily-register, and occupancy consumption of booking patterns is deferred to later work.

## Booking Pattern Enrollment Independence

A new child created via the registration wizard must have a non-empty booking pattern; the wizard's Session Pattern step is mandatory. Existing children created before this rule remain enrollment-complete and are not backfilled; they are surfaced on the child list and detail with a "No session pattern" badge and a one-click link to the standalone pattern page. Attendance capture and invoicing are not blocked by the absence of a booking pattern for an existing child.

## Booking Pattern Creation Endpoint

The wizard's two-step submit calls `POST /children` first and then `POST /children/:child_id/booking-patterns` for the same actor. If the pattern POST fails after the child POST succeeds, the Review step shows a Retry button and a "Save child without pattern" fallback that routes to the child detail screen; the standalone `:childId/booking-pattern` page can then be used to finish setup. The API does not expose a single-transaction child-with-pattern create.

## Session Type Management Authority

Session types are site-scoped reference data managed like rooms: managers and owners may create, update, archive, and reactivate them, while practitioners may read them. Session type management is not a parent-facing workflow.

## Session Template Management Authority

Session templates share the session-type authority model: managers and owners may create, update, archive, and reactivate them per site, while practitioners may read them. Templates are not a parent-facing workflow.

## Booking Pattern Management Authority

Booking pattern assignment and edits are manager-only child-record writes; practitioners may read a child's booking pattern, and owner-level users do not get direct per-child booking-pattern create, edit, or delete access in the current release.

## Booking Pattern History Immutability

Closed or past booking patterns are read-only and are never edited or deleted; only the current open pattern whose effective date is today or later may be edited. Changing an already-started plan is done by appending a new pattern rather than editing history.

## Booking Pattern Effective Date Rules

A new booking pattern's effective-from date must be today or later (no backdating) and must not overlap any existing pattern for the same child. Creating a new pattern closes the previous active pattern by setting its effective-to to the day before the new pattern's effective-from, producing an adjacent, gap-free, non-overlapping timeline. Mid-history insertion of a pattern between existing patterns is not supported.

## Attendance Daily List Scope

The attendance-facing child list shows children for the current `Europe/London` local day: active children and any child with an open attendance session that still needs resolution. Avoid treating this as a historical attendance report.

## Attendance Correction Target

An attendance correction applies to the effective attendance session for one child rather than rewriting the original check-in or check-out event. A correction may adjust an existing session or record a missed session that should have existed.

## Attendance Correction Scope

Attendance correction changes or establishes the full effective check-in and check-out interval for a session; voiding or excluding a session from billing is a separate deferred concept.

## Duplicate Entry Attendance Correction

A duplicate-entry attendance correction still establishes a valid effective attendance interval; it does not void, delete, or exclude a session from billing in the current release.

## Attendance Correction History

A session may receive more than one attendance correction; each correction remains part of the historical trail while the latest correction determines the current effective interval. Manager correction history presents original attendance events and each correction in chronological order; prior events are read-only.

## Attendance Correction Recorded Time

The time and local day a correction is recorded are the manager action time and action day, distinct from the corrected attendance interval.

## Attendance Correction Authority

Only managers can create attendance correction events.

## Attendance Correction Child Selector Scope

Manager correction workflows may target any child in the manager's branch, including inactive children, because historical corrections may be needed after a child has left. Correction validity is still constrained by the child's enrollment window.

## Attendance Correction Reason Vocabulary

Only attendance corrections carry attendance reason codes; routine check-in and check-out events do not. Attendance corrections use attendance-specific reason codes rather than lifecycle reason codes; the starter set is `missed_check_in`, `missed_check_out`, `incorrect_time`, `duplicate_entry`, and `other`. When `other` is selected, managers must provide a note; notes are optional for the standard codes.

## Attendance Correction Audit Reason Semantics

Attendance correction reason codes are distinct from lifecycle reason codes even when an audit trail records the correction.

## Invoice Source of Truth

Monthly invoice billable minutes are derived from attendance actuals.

## Funding Application Rule

Funded-hours deduction applies only to core childcare hours; extras remain fully payable.

## Core Attended Minutes

Core attended minutes means rounded core attendance minutes for billing and funding deduction, not raw elapsed attendance minutes.

## Funded Deduction Minutes

Funded deduction minutes are the portion of core attended minutes covered by the child's funded allowance for the billing month.

## Core Billable Minutes

Core billable minutes are the remaining core attended minutes after funded deduction minutes are applied.

## Core Billing Formula

Core due hours are calculated as `max(0, rounded core attendance minutes - funded hours allowance)` before pricing is applied. Rounded core attendance minutes are the sum of per-session billable minutes rather than raw elapsed minutes.

## Core Billing Price Rounding

Core childcare amounts are converted from minutes and hourly minor-unit rates by rounding any fractional minor unit up to the next minor unit.

## Attendance Billing Rounding

Each attendance session is rounded up to the nearest 15 minutes for billing. Any positive elapsed-time remainder beyond a 15-minute boundary rounds up; exact 15-minute boundaries do not add another block.

## Billing Timezone

Attendance day boundaries and invoice-period calculations use `Europe/London` local time.

## Attendance Timestamp Semantics

Attendance event times are captured as absolute instants while attendance day grouping is derived from `Europe/London` local dates.

## Incomplete Attendance Handling

Attendance records missing check-out are excluded from automatic billing until a manager correction establishes the full effective attendance interval.

## Incomplete Attendance Triage

Manager-facing review of unresolved attendance sessions for the current billing month, with today's unresolved sessions surfaced first. This is billing readiness triage rather than a historical attendance report.

## Invoice Generation Flow

Managers manually generate monthly draft invoices before any invoice is issued.

## Invoice Draft Preflight

A manager-facing readiness preview for one billing month before draft invoice generation. It identifies which child-months are eligible or blocked without being an invoice run.

## Invoice Draft Preflight Population

Invoice draft preflight considers child-months whose enrollment window overlaps the billing month, even if the child is no longer currently active.

## Invoice Draft Preflight Existing Invoice Rule

An existing draft monthly invoice does not block invoice draft preflight for that child-month, while an issued-or-later monthly invoice does.

## Invoice Draft Preflight Totals

Invoice draft preflight totals are estimated aggregate invoice amounts for eligible child-months before draft invoice generation.

## Zero-Attendance Invoice Eligibility

A child-month with no completed attendance sessions can still pass invoice draft preflight when enrollment, billing, funding, and attendance-completeness data are present.

## Zero-Total Draft Invoice

An eligible child-month can produce a draft monthly invoice with zero amount due; zero-total drafts still represent the monthly billing statement for that child-month.

## Zero-Total Invoice Issue

A zero-total draft invoice can be issued like any other eligible draft invoice and receives an invoice number.

## Payable Invoice

A parent-visible issued-or-later invoice with a positive outstanding balance. Zero-total issued invoices and paid invoices are not payable.

## Parent Billing View

A parent-facing surface where a parent can view current and past issued-or-later invoices for linked children, see payment status, and start payment for payable invoices. Avoid: family account statement, billing dashboard.

## Parent Portal Surface

The parent-facing area of the product remains separate from staff and owner operational surfaces. Parent navigation exposes only parent-safe workflows for linked children and must not show staff or owner navigation.

## Parent Invoice Attention Order

Parent-facing invoice lists surface invoices needing payment attention before paid or zero-balance history. Overdue invoices have the highest payment attention, followed by payment-failed invoices and then payable issued invoices.

## Parent Invoice Child Group

A parent-facing grouping of invoice history for one linked child. Child groups organize non-urgent current and past invoices without hiding cross-child payment attention.

## Parent Invoice Detail View

A parent-facing view of one parent-visible invoice showing the billing period, totals, funded deduction, calculation summary, and invoice lines. It must remain parent-safe and must not expose manager-only invoice generation or audit details.

## Unpaid Invoice Label

A manager-facing label for issued or overdue invoices with no payment collected. It is not a separate invoice lifecycle state, and payment-failed invoices remain separate because they need failure-specific follow-up.

## Payment Follow-up Queue

Manager-facing work queue for invoices that need payment attention, sorted by urgency: overdue invoices, payment-failed invoices, then issued unpaid invoices. Avoid splitting these into unrelated dashboard concerns.

## Invoice Issue Mode

Managers can issue invoices one-by-one or in bulk; the default flow is bulk issue with confirmation.

## Invoice Issue Confirmation

A manager's explicit approval that selected draft invoices should become immutable issued invoices. Confirmation applies to both one-by-one and bulk issue actions.

## Invoice Issue Result Summary

A manager-facing summary shown after invoice issue that identifies which invoices became issued, the invoice numbers assigned, and which drafts were skipped or failed and still need follow-up.

## Invoice Issue Time

The business instant when a draft invoice becomes an issued invoice. In the current release, the invoice is also locked and due at that same instant.

## Invoice Issue Validation Boundary

Issuing validates that the invoice is an existing draft invoice in the manager's billing scope. It does not recalculate billing readiness; managers regenerate drafts before issue when source data changes.

## Bulk Invoice Issue

A manager-triggered issue action for all draft monthly invoices in one billing month or a manager-selected subset of those drafts. Bulk issue requires explicit manager confirmation before invoices become issued.

## Bulk Invoice Issue Default Selection

Manager-facing bulk issue starts with all ready draft invoices selected for the billing month. Managers may remove individual drafts before confirmation, while one-by-one issue remains a fallback action from draft review.

## Invoice Run

A manager-triggered monthly billing operation that prepares or issues per-child invoices as a batch. An invoice run may include successful invoices and child-specific exceptions.

## Invoice Run Exception

A child-specific blocker reported during invoice preflight, draft generation, or issue. Exceptions do not stop unaffected children from being drafted or issued in the same billing month.

## Invoice Run Status

Manager-facing current billing-month readiness and progress for invoice generation and issue. It summarizes eligible, blocked, draft, and issued child-months, with the latest run time where one exists.

## Single Invoice Issue Run

A manager-triggered invoice run that issues exactly one draft invoice. It is still recorded as an invoice run so issued invoice history is consistent across single and bulk issue paths.

## Issued Invoice Edit Policy

Issued invoices are immutable; changes require explicit adjustment rather than direct edits.

## Adjustment Invoice

A follow-up invoice that corrects or offsets a previously issued invoice. An adjustment invoice must be linked to the issued invoice it adjusts and must carry a manager-provided reason.

## Payment Scope

Parents pay invoices in full; partial payments are not supported in the current release.

## Payment Attempt

A single try to collect a payable invoice through the payment provider. Each retry creates a new payment attempt for the same invoice.

## Checkout Retry Availability

A manager-visible indication that a parent can currently start a fresh payment attempt for a payable invoice. It is diagnostic only and is not itself checkout session creation.

## Payment Outcome Authority

Invoice payment state changes only after a payment provider-confirmed outcome. Browser return or cancel navigation after payment initiation is not itself a payment outcome.

## Payment Canceled Return

A parent-facing message shown when a parent returns from hosted checkout without completing payment. It is a temporary return outcome, not a durable invoice payment state; the invoice remains payable unless a payment provider-confirmed outcome says otherwise.

## Payment Reconciliation Record

A manager-facing record of a handled payment outcome for one invoice payment attempt. It explains whether the attempt paid, failed, expired, or was ignored without relying on raw provider webhook payloads as the operational timeline.

## Payment Event

A manager-facing payment timeline item backed by a payment reconciliation record. Avoid using payment event to mean the raw payment provider webhook payload.

## Webhook Processing Status

The local result of processing a verified payment provider webhook event, such as processed, ignored, or rejected. It is distinct from the provider's webhook delivery status.

## Payment Failure State

Failed or canceled payment attempts move invoices to a `payment_failed` state.

## Payment-State Transition Telemetry

Payment-state transition telemetry describes operational monitoring of local payment attempt and invoice status changes caused by checkout creation, verified payment-provider webhook outcomes, or overdue scheduler transitions. It is distinct from payment reconciliation records, which remain the manager-facing history.

## Invoice Numbering

Invoice identifiers follow `INV-YYYYMM-####` sequence format.

## Invoice Numbering Month

The `YYYYMM` segment of an invoice number is the invoice billing month, not the calendar month when the manager issues it.

## Bulk Invoice Issue Sequence Order

Bulk-issued invoices receive invoice numbers in deterministic child-name order, with invoice identity used only as a tie-breaker.

## Invoice Granularity

Invoices are issued per child, not combined at family level.

## Sibling Discount Policy

Sibling discounts are deferred and not part of current billing rules.

## Extras Charging Model

Extras are added manually as invoice line items by managers.

## Draft Invoice Extras Placeholder

Draft invoice generation does not create payable extras automatically; it preserves an explicit zero-value placeholder so later manager-entered extras remain outside attendance and funding calculations.

## Draft Invoice Manual Extras Preservation

Regenerating a draft monthly invoice preserves existing manual extra lines while recalculating attendance-derived core childcare and funded-hours deduction lines.

## Attendance Capture Scope

Attendance capture includes check-in and check-out only; room-move tracking is out of scope.

## Attendance Capture Time Authority

Routine practitioner check-in and check-out capture uses the current server time; historical or custom attendance times are manager-controlled corrections.

## Practitioner Attendance Scope

Practitioners can perform check-in and check-out for any child within the active session branch.

## Practitioner Contact Visibility

Practitioner attendance workflows expose only attendance-facing child information and do not expose parent_carer contact details such as email or phone in the current release.

## Child and Contact Write Authority

Child/contact and child-mapping write actions (create, update, mapping changes) are manager-only in the current release, while practitioner access remains read-only for attendance-facing child views.

## Absence Marker

An absence marker records that one child was absent for one `Europe/London` local date. It is separate from attendance sessions and attendance corrections, and it does not create, edit, void, or exclude any attendance session; a child-day cannot have both an absence marker and attendance.

## Clear Absence Marker

Clearing an absence marker means removing the active absent state for a child-day while preserving the historical record that the marker existed.

## Absence Marker Authority

Managers and practitioners may create and clear absence markers. Parents do not have access to absence marker workflows.

## Absence Marker Detail Scope

Absence markers do not carry reason codes, notes, charge policy, or funding policy in the current release.

## Absence Marker Date Scope

Absence markers are used only for the current `Europe/London` local day in the current release; future absence scheduling is out of scope.

## Billing Period

Billing runs monthly on calendar-month boundaries.

## Billing Month

A calendar month identified as `YYYY-MM` for funding and invoice workflows.

## Invoice Run Default Billing Month

Manager-facing invoice run workflows default to the most recent completed billing month. Managers may still select the current billing month or older billing months intentionally.

## Billing Quantity Unit

Attendance-derived billing quantities use integer minutes as the canonical unit. Hours are a display format only.

## Funded Allowance Unit

Funded-hours allowances are stored and exchanged as integer minutes. Hours are a display format only.

## Funded Allowance Bounds

A funded-hours allowance may be zero but must not exceed the number of minutes in a 31-day month.

## Child Enrollment Minimum

A child requires name, date of birth, start date, one parent_carer contact, and a billing rate before attendance and invoicing flows begin.

## Enrollment Gate Scope

Enrollment minimum checks gate all new routine attendance capture and invoicing actions, while manager-only historical attendance corrections remain allowed under the post-enrollment correction policy.

## Enrollment-Incomplete Attendance Visibility

Enrollment-incomplete children remain visible in the attendance-facing child list, but routine check-in remains blocked until enrollment is complete.

## Child Creation Flow

Managers may create a child record before adding a parent_carer contact, but attendance and invoicing remain blocked until all child enrollment minimum requirements are satisfied.

## Child Billing Rate Source

Each child has one current core billing rate in enrollment data, while issued invoices preserve the applied rate in invoice lines for historical explainability.

## Draft Invoice Rate Source

Draft invoice generation uses the child's current stored core billing rate at generation time, including when generating drafts for historical billing months.

## Zero Core Billing Rate

A core billing rate of zero is an explicit valid rate, not missing billing data.

## Child Enrollment Lifecycle

Child records remain retained when enrollment ends; the child can be marked inactive/left while attendance and billing history stays intact.

## Child Lifecycle Reason Requirement

Manager-initiated child lifecycle transitions such as marking a child inactive/left require a stable reason code with an optional note.

## Child Lifecycle vs Enrollment Date Separation

Child lifecycle transitions (active/inactive-left) are managed independently from enrollment date edits; enrollment boundary dates are updated through explicit date-change flows with enrollment window integrity checks.

## Child Reactivation Deferral

Current lifecycle APIs do not include explicit reactivation of ended parent-child mappings, while explicit child reactivation endpoints are deferred unless pilot operations require them.

## Child Default Listing Scope

Manager child and child listings default to active records only, with an explicit option to include inactive or ended records (child-level only) for historical administration.

## Child Identity Uniqueness

Child identity is anchored by UUID entity identifiers, and matching name/date-of-birth combinations are not enforced as a hard uniqueness rule.

## Post-Enrollment Attendance Correction

After enrollment ends, managers may still record corrections for historical attendance sessions so billing derived from attendance actuals remains accurate. Historical corrections require the child to exist in scope but do not require current active enrollment, while the corrected attendance interval remains constrained to the child's enrollment dates.

## Enrollment Date Semantics

Enrollment boundaries use date-only fields (start and optional end date), while billing calculations continue to derive from attendance timestamps in `Europe/London`.

## Enrollment Window Integrity

Enrollment date updates are rejected when they would place existing attendance records outside the child's enrollment window.

## Parent-Child Mapping Cardinality

Parent-Child mappings are many-to-many between parent memberships and children: a parent membership may map to many children, and a child may be mapped to by many parent memberships..

## Parent Visibility Scope

One parent account can be linked to multiple children and can view invoices for linked children within the active session scope.

## Parent Portal

The parent-side product surface where a parent-role user views issued invoices for linked children and initiates full invoice payment. It exposes only parent-relevant billing navigation and account/session controls, not staff operational or admin surfaces.

## Parent Invoice Payment Initiation

Only parent-role users initiate payment for parent-visible payable invoices in the current release. Managers can review invoice and payment state but do not initiate parent payment.

## Manager Payment Retry Visibility

A manager-facing indication that a parent can retry payment for a payable invoice, without letting the manager launch Stripe Checkout or complete payment on the parent's behalf.

## Manager Payment Timeline

A lean chronological manager-facing view of payment attempts and webhook/payment outcomes for a single invoice. It explains the outcome, amount, timestamp, and reason in nursery operations language, while provider identifiers remain secondary diagnostics and raw webhook payloads are not shown.

## Manager Invoice Payment Review

The invoice-detail anchored manager workflow for understanding payment status, retry availability, and reconciliation history for one invoice. The invoice list may show lightweight payment cues, but FE-23 does not introduce a separate manager Payments area.

## Manager Payment Pending State

A manager-facing payment review state for a checkout attempt that has been created or is still awaiting provider confirmation. Managers see the attempt time and amount, and retry is not presented while the API reports that retry is unavailable.

## Login Identifier

User authentication uses a unique, case-insensitive normalized email address as the login identifier.

## Password Credential Policy

User passwords must be at least 8 characters in the current release; additional composition rules are not enforced.

## Pilot Tenant Bootstrap

The first manager account for a tenant is created through a one-time administrative seed command; subsequent users are provisioned by manager invitation.

## Audit Baseline Scope

Audit logs are mandatory for user provisioning changes, child record updates, attendance events and corrections, funding profile updates, invoice draft/issue actions, and payment-status updates.

## Audit Request Correlation

Persisted audit events include request identifier correlation so domain changes can be traced to individual API requests.

## Request Correlation Context

A shared operational correlation context connects API request handling, audit-visible domain changes, payment-provider webhook handling, and scheduler telemetry for one troubleshooting flow. In the current release, this means correlation identifiers in logs and headers, low-cardinality metrics for correlated operations, and audit-linked request identifiers rather than full distributed tracing spans.

## Audit Actor Semantics

Audit events include actor membership identity for user-initiated actions and may omit actor identity only for system-initiated actions, while tenant and branch scope plus action metadata remain required.

## Manager Invite Audit Scope

Manager invitation creation, resend, revocation, and acceptance are persisted as audit events because they change user provisioning state.

## Manager Invite Acceptance Actor Semantics

Invitation acceptance audit events are attributed to the newly provisioned user and membership created by that acceptance, while manager-initiated invitation actions are attributed to the manager's active session actor.

## Cross-Month Session Allocation

An attendance session that crosses midnight is allocated to the invoice month of its check-in date using `Europe/London` local time.

## API Response Style

API endpoints return plain JSON resources with standard HTTP status codes instead of a global response envelope.

## API Error Contract

API errors use a consistent JSON structure with stable error code, human-readable message, optional details, and request identifier.

## Lifecycle Error Code Baseline

Child, child-contacts, and parent-child-mapping lifecycle endpoints use a stable domain error-code set so clients can handle expected lifecycle failures deterministically.

## Authorization Error Status Policy

Authorization failures return `403 Forbidden` with stable authorization-specific error codes for explicit role, scope-selection, and relationship failures, while opaque resource identifiers outside the active tenant-branch scope are treated as not found so resource existence is not disclosed across scopes.

## Authorization Error Disclosure Policy

Authorization errors expose stable machine-readable denial codes while keeping human-readable messages generic.

## Authorization Denial Code Baseline

The initial stable authorization denial code set includes role, scope, unknown-role, parent-child-link, and scope-selection denial variants alongside unauthorized authentication failure.

## Scope Selection Error Semantics

Authentication-time membership selection does not create a session until a valid active membership is selected. Missing selection after valid credentials returns a membership selection challenge when multiple active memberships are available; malformed selection remains a validation error; an unavailable selected membership returns the challenge again when alternatives remain and a generic sign-in failure when none remain. Authenticated session actions with a well-formed selection outside the user's allowed memberships return an authorization error.

## Persistence Strategy

Data access uses `sqlc` with `pgx` and typed SQL queries rather than an ORM.

## Migration Workflow

Schema changes are applied with manual `golang-migrate` commands against local PostgreSQL and are not auto-run on API startup.

## Session and Token Policy

Authentication uses short-lived access JWTs plus database-backed refresh tokens that can be revoked.

## Session Scope Mode

Each active session is bound to exactly one membership scope (tenant, branch, and role) rather than selecting scope per request.

## Session Scope Selection Rule

When a user has multiple memberships, the active session scope is chosen explicitly by the user at login and must belong to that user.

## Session Single-Scope Auto-Selection

Authentication auto-selects the active membership scope when exactly one active membership is available.

## Login Membership Selection Challenge

A pre-session sign-in state reached after a login identity is verified but before an authenticated session exists, when more than one active membership is available and the user must choose one.

## Membership Choice Label

A user-facing label that helps a person recognize an available membership by nursery, branch, and role. Raw membership, tenant, and branch identifiers are not suitable labels for human selection.

## Membership Selection Retry

The continuation of the same sign-in flow after a membership selection challenge, where the user chooses one available membership without re-entering credentials. Editing the login identity or password starts a new sign-in flow.

## Unavailable Membership Choice

A membership choice that was offered during sign-in but can no longer be selected by the time the user retries. The user should be guided back to another available choice when possible, without creating an authenticated session.

## Session Scope Selection Identifier

Session scope selection uses the membership identifier as the canonical selector instead of a tenant-branch-role tuple.

## Session Scope Requirement

Authentication does not create a session unless the user has at least one active membership scope available.

## Membership Breadth

A user may hold memberships across multiple tenants and branches, but each active session is bound to exactly one selected membership scope.

## Single Role Per Scope

Within a single tenant-branch scope, a user holds exactly one membership role; concurrent dual-role memberships in the same scope are not used in the current release.

## Registered Email Invite Deferral

Manager invitations in the current release are used only for new login identities; inviting an email address that already belongs to a user is deferred until an authenticated existing-user membership invite flow exists.

## Membership Activity State

Authentication and authorization consider only active memberships; inactive memberships cannot be selected for sessions.

## Session Scope Visibility

Authentication responses include both the active session scope and the user's available scopes so clients can present explicit scope switching.

## Session Membership Binding

Each refresh-token-backed session is bound to one membership scope and cannot change scope unless an explicit membership switch request is validated.

## Session Refresh Scope Integrity

Token refresh succeeds only when the session's bound membership remains active; refresh does not auto-fallback to a different membership.

## Session Claim Shape

Access tokens include subject user identifier, membership identifier, tenant identifier, branch identifier, role, and standard token timestamps so each request can be unambiguously authorized against one active membership.

## Session Scope Switch Flow

Changing the active membership scope is an explicit session action and is distinct from routine token refresh.

## Session Scope Switch Authentication

Membership scope switching is performed through a refresh-token-backed session action that rotates session tokens.

## Session Action CSRF Policy

Refresh-token-cookie-backed session actions, including refresh, logout, and membership switch, require CSRF protection in the current release.

## Logout Idempotency

Session logout is idempotent and returns success even when no active refresh-token session is present.

## Session CSRF Mechanism

Cookie-backed session actions use a double-submit CSRF token pattern with a client-readable CSRF token echoed in a request header and a trusted origin or referer check.

## Session Scope Switch Auditing

Membership scope switch actions are persisted as audit events with actor, previous scope, new scope, and request identifier.

## Authentication Event Persistence Scope

Login, refresh, logout, and password reset activity are treated as authentication/security telemetry in structured logs and metrics rather than persisted as audit-log domain events in the current release.

## Authentication Failure Telemetry

Authentication failure telemetry covers failed identity, credential, session-token, and reset-token validation. It is distinct from ordinary request validation errors and from authorization denials after a user or session scope is known.

## Token Transport Policy

Access tokens are sent as bearer tokens, while refresh tokens are stored in secure HttpOnly cookies.

## Session Concurrency

Users may hold multiple active sessions concurrently, with token revocation supported per session.

## Session Persistence Model

Session state is persisted as refresh-token-backed, revocable session records; a second parallel session artifact is not introduced in the current release.

## Membership Change Revocation Policy

When membership scope or role is changed, refresh tokens for affected sessions are revoked immediately, while already-issued access tokens remain valid only until their normal expiry.

## Password Reset Session Revocation Policy

After a successful password reset, all refresh-token-backed sessions for that user are revoked immediately, while already-issued access tokens remain valid only until their normal expiry.

## Token Lifetime Policy

Access tokens expire after 15 minutes; refresh tokens expire after 30 days and rotate on refresh.

## Password Reset

Users can reset passwords through a basic email-link reset flow.

## Password Reset Account Scope

Password reset applies to the global user login identity rather than to an individual membership, tenant, or branch scope.

## Password Reset Request Disclosure Policy

Password reset requests do not reveal whether a login identifier belongs to an active user account; eligible active users receive reset email while inactive or unknown users do not, and the requester sees a generic outcome.

## Password Reset Link Supersession

Only the newest unused password reset link for a user remains valid; creating a new reset request invalidates earlier unused reset links for that user.

## Password Reset Link Lifetime

Password reset links are short-lived recovery links that remain valid for 60 minutes from creation.

## Password Reset Link Issuance

A password reset link is considered issued only when the reset message is accepted for delivery; failed delivery does not leave a usable reset link behind.

## Password Reset Token Exposure Policy

Raw password reset tokens are delivered only through the reset message and are never returned by API responses.

## Manager Invite Token Exposure Policy

Raw manager invitation tokens are delivered only through the invitation message and are never returned by API responses.

## Password Reset Request Throttling

Password reset requests are rate-limited by login identifier and client source so the public recovery flow cannot be used for unbounded email sending.

## Manager Invite Acceptance Throttling

Public invitation acceptance attempts are rate-limited by client source to reduce token-guessing risk.

## Password Reset Token Error Semantics

Invalid, expired, and already-consumed password reset links return distinct stable error codes with generic human-readable messages; superseded reset links are treated as already consumed.

## Manager Invite Link Lifetime

Manager invitation links expire after seven days from issuance.

## Manager Invite Link Supersession

Resending a pending manager invitation issues a new invitation link and invalidates the prior unused link for that invitation.

## Manager Invite Link Issuance

A manager invitation link is considered issued only when the invitation message is accepted for delivery; failed delivery does not leave a newly generated usable invite link behind.

## Manager Invite Idempotent Create

Creating a manager invitation for the same normalized email, role, tenant, and branch as an existing pending invitation refreshes that pending invitation instead of creating a duplicate; creating one for the same email and scope with a different role is a conflict.

## Manager Invite Historical Reissue

Creating a new manager invitation after an earlier invitation for the same email, role, and scope was revoked or expired creates a new invitation record rather than reopening the historical invitation.

## Manager Invite Minimum Data

Creating a manager invitation requires only the invitee login email and invited role in the current release; staff profiles and display names are not collected by the invitation flow.

## Manager Invite Scope Source

Manager invitations target the manager's active tenant-branch session scope; clients do not choose invitation tenant or branch in the request body.

## Manager Invite List Visibility

Manager invitation lists default to currently pending invitations in the active tenant-branch scope, while accepted, revoked, expired, and all-invite history are available only through explicit status filtering. Status-changing manager actions are available only for currently pending invitations; expired invitations are treated as historical in the manager UI.

## Manager Invite Expiry State

Invitation expiry is derived from the invitation expiry time rather than a separate manager action or background lifecycle transition.

## Manager Invite Revocation

Managers may revoke pending invitations, including pending invitations whose links have already expired; revoking an already revoked invitation is idempotent, while accepted invitations cannot be revoked.

## Manager Invite Revocation Reason Scope

Manager invitation revocation does not require a reason code or note in the current release; the audit event records the actor and invitation details without using lifecycle reason fields.

## Manager Invite Token Error Semantics

Invalid, expired, revoked, and already-accepted manager invitation links return distinct stable error codes with generic human-readable messages.

## Already-Accepted Manager Invitation Link

A manager invitation link whose invitation has already been accepted by an invitee. It cannot provision another user or membership; avoid calling it an already-used invitation link.

## Manager Invite Acceptance Idempotency

Invitation acceptance is single-use; after one successful acceptance, later attempts with the same invitation link return an already-accepted result and do not create additional users or memberships.

## Email Delivery Strategy

Manager invites and password reset messages are sent through a single provider abstraction, using an SMTP sandbox in development and one transactional email provider in production.

## Branch Scope

Records remain branch-scoped in the data model with one default branch used in the pilot.

## Core Record Deletion Policy

Core child, attendance, and invoice records are not hard-deleted; corrections, voiding, or archival flows are used instead.

## Child and Contact Management Lifecycle

Manager child/contact management supports create and update plus lifecycle transitions (child inactive/left, parent-membership-to-child mapping ended) rather than hard delete operations. There is no longer a separate guardian entity to manage; parent contacts are stored on child_contacts rows of type parent_carer.

## Parent-Child Mapping Recreate Policy

Managers may recreate a parent-membership-to-child mapping that was previously ended, restoring the parent's portal access for that child. The system does not auto-restore; the manager must create a new active mapping. The new mapping is independent of the prior ended row's audit history.

## Relationship End Reason Requirement

Manager-initiated actions that end a parent-membership-to-child mapping require an explicit reason for audit explainability.

## Relationship End Reason Shape

Relationship end reasons use a stable machine-readable reason code with an optional human note so audit trails stay both queryable and explainable.

## Lifecycle Reason Vocabulary

Lifecycle transitions across child and parent-membership-to-child mapping use a shared controlled vocabulary of reason codes with scoped subsets per action.

## Lifecycle Reason Starter Set

The initial shared reason-code set for the current release is `duplicate_record`, `entered_in_error`, `left_nursery`, `safeguarding_direction`, `contact_update`, `access_revoked`, and `other`, with scoped usage per lifecycle action.

## Lifecycle Other-Reason Note Requirement

When lifecycle `reason_code` is `other`, a non-empty `reason_note` is required so the audit trail remains explainable.

## Relationship End Terminology

Lifecycle actions that stop active links or mappings use the canonical verb "end"; terms implying hard deletion (such as delete/remove) are not used for these actions.

## Child Parent Carer Contact Requirement Enforcement

Removing a child's last parent_carer contact is allowed, but that child immediately becomes enrollment-incomplete and is blocked from attendance and invoicing until at least one parent_carer contact is restored.

## Parent-Child Mapping Idempotent Create

Creating a parent-membership-to-child mapping for a pair that already has an active mapping is treated as an idempotent success and returns the existing active row.

## Parent-Child Mapping Re-Linking

A parent-membership-to-child mapping can be recreated for a pair that previously had one; the new mapping is independent of the prior ended row's audit history. The partial unique index ensures only one active mapping per pair at a time.

## Invoice Due Policy

Invoices are due on receipt when issued.

## Unpaid Issued Invoice

An issued invoice with an outstanding positive balance. A zero-total issued invoice is not considered unpaid for overdue transition purposes.

## Invoice Overdue Transition

An unpaid issued invoice transitions to `overdue` at 00:00 the next local day in `Europe/London`.

## Payment Failed Overdue Boundary

A `payment_failed` invoice remains payable but does not transition to `overdue`.

## Invoice Issue Exception Handling

Invoice issue runs can proceed for eligible draft invoices while invoices that cannot be issued are blocked and returned in an exception list for manager resolution.

## Guided Invoice Run Default Path

Manager-facing invoice runs keep eligible child-months actionable when other child-months have exceptions. The default manager path is to continue with ready invoices and resolve blocked child-months separately.

## Draft Invoice Generation Exception Handling

Draft invoice generation can proceed for eligible child-months while blocked child-months are skipped and returned as exceptions for manager resolution.

## Draft Generation Transaction Boundary

Expected child-month blockers produce generation exceptions, while unexpected system failures leave no partial invoice generation result behind.

## Invoice Issue Transaction Boundary

Expected invoice issue blockers produce issue exceptions, while unexpected system failures leave no partial invoice issue result behind.

## Empty Draft Generation Run

A draft invoice generation run can complete without creating invoices when no child-months are eligible; this is a valid billing outcome rather than a request failure.

## Empty Invoice Issue Run

An invoice issue run can complete without issuing invoices when no draft invoices are eligible; this is a valid billing outcome rather than a request failure.

## Selected Draft Generation Child Exception

When a selected child-month cannot be found in the active billing scope or does not overlap the billing month, draft generation treats that child-month as an exception rather than failing the whole run.

## Selected Draft Generation Uniqueness

Selected-child draft generation operates on unique child-months; duplicate selected child identifiers do not create duplicate invoice work.

## Selected Invoice Issue Uniqueness

Selected-invoice issue operates on unique invoices; duplicate selected invoice identifiers do not create duplicate issue work.

## Invoice Run Blocked Child

A child excluded from a specific monthly invoice run because billing readiness checks found a resolvable issue, such as incomplete attendance. This is a run/month-specific billing state, not a child lifecycle state.

## Invoice Exception Reference

A manager-visible reference from invoice review to child-month blockers recorded on the invoice run that generated or processed related invoices. Exception references provide run context and do not represent invoice lines, adjustment invoices, or invoices for blocked child-months.

## Invoice Preflight Blocker

A child-month readiness exception returned by invoice draft preflight. A blocked child may have multiple preflight blockers, and each blocker should be exposed with a stable code.

## Invoice Run Exception Resolution

Manager-facing invoice run exceptions identify the blocked child-month, the blocking reason, and the next workflow that can resolve it. The invoice run workflow does not edit attendance, enrollment, billing-rate, or funding records inline.

## Invoice Preflight Enrollment Blocker Codes

Invoice draft preflight explains enrollment incompleteness with granular stable blocker codes rather than only a generic enrollment-incomplete label.

## Draft Invoice Regeneration

After attendance corrections, managers can regenerate draft invoices for individual children without rerunning the full month batch.

## Draft Invoice Generation Scope

Draft invoice generation can target all eligible child-months in a billing month or a manager-selected subset of child-months.

## Draft Invoice Generation Outcome

Draft invoice generation reports the invoice run outcome and affected invoice references; detailed invoice review remains a separate manager billing view.

## Manager Invoice Review

A manager-facing billing view for inspecting generated invoice headers, line items, calculation quantities, status, and due/payment metadata across invoice statuses. Draft invoices are included so managers can review calculations before issue without reconstructing totals elsewhere; invoice line editing is a separate manager workflow, not part of review itself.

## Parent Invoice View

A parent-facing billing view for issued-or-later invoices belonging to children reachable through that parent's current membership-to-child mappings. It explains payable invoice identity, child, period, status, totals, payment state, and parent-readable line items.

## Parent Invoice Detail Disclosure

Parent invoice detail supports understanding and paying an issued invoice, not reviewing manager billing operations. Manager-only run context, invoice exceptions, lock metadata, adjustment internals, and attendance source-session snapshots are outside the parent disclosure boundary.

## Issued-or-Later Invoice

An invoice whose lifecycle has moved beyond draft into `issued`, `payment_failed`, `paid`, or `overdue`. Parents can see issued-or-later invoices for linked children, while draft invoices remain manager-only.

## Invoice Payment Retry

Issued invoices remain payable after `payment_failed` or `overdue` states by creating a fresh Stripe checkout session for the same invoice.

## Tax Handling

Current invoicing runs in a single non-VAT mode without VAT calculation logic.

## Invoice Currency

All invoices and Stripe checkout sessions use `GBP` only in the current release.

## Portal Delivery Model

Managers, practitioners, and parents use one web application with role-scoped access rather than separate applications.

## Billing Visibility Scope

Invoice amounts, statuses, and payment details are visible only to managers and parents; practitioners do not have billing access.

## Access Scope Enforcement

All record access is constrained by both tenant and branch scope in the current release, even with a single pilot tenant and default branch.

## Parent-Child Mapping Scope Consistency

Parent-membership-to-child mappings are valid only when the parent membership and the child belong to the same tenant and branch scope; the database trigger `enforce_parent_membership_child_scope` rejects any insert that violates this rule.

## Tenancy Isolation Enforcement

Tenant and branch isolation is enforced in application-layer authorization and query scoping in the current release; database RLS is deferred to post-pilot hardening.

## Test Priority

Automated test coverage is prioritized for funding calculation logic, invoice state transitions, authorization scope checks, and Stripe webhook idempotency, with minimal UI end-to-end coverage on core happy paths.

## Authorization Verification Baseline

Authorization acceptance requires a route-by-route role and scope test matrix that proves unauthenticated requests are rejected, wrong-role/wrong-tenant/wrong-branch requests are forbidden, and correctly scoped allowed-role requests succeed.

## Observability Baseline

Operations monitoring uses structured logs plus essential metrics for webhook outcomes, invoice-generation job health, and authentication failures.

## Authorization Denial Logging

Authorization denials are captured in structured logs with request identifier, actor, scope context, and denial reason code.

## Authorization Denial Persistence Scope

Authorization denials are treated as operational telemetry and are recorded in structured logs and metrics rather than persisted as audit-log domain events in the current release.

## Authorization Denial Metrics

Authorization denials emit metrics tagged by stable denial reason code to support operational monitoring.

## Async Processing Scope

Non-immediate external operations such as email sending and retry handling run through background jobs; user-triggered Stripe checkout session creation remains synchronous.

## Nursery Room

An operational room within a nursery site that houses children of a specific age group. A room belongs to exactly one site, and a site may have many rooms. The same room name (e.g. "Baby Room") may exist in different sites within the same tenant, but must be unique among active rooms within a single site. Rooms are archived (is_active = false) rather than hard-deleted because future attendance, child assignment, ratio, and reporting features may reference them.

## Room Age Group

A classification for a nursery room: baby, toddler, preschool, or mixed. "Mixed" means the room intentionally serves children across multiple age bands. Age group is a constrained value rather than free text.

## Room Capacity

The maximum number of children a room is licensed or configured to hold. Capacity is a positive integer and may be changed over time as room configuration or Ofsted registration changes.

## Room Archive

Setting a room to inactive (is_active = false). Archiving is blocked while active children remain assigned to the room. Archived rooms may be reactivated later. This follows the same no-hard-delete policy as other core records.

## Room Reactivation

Restoring an archived room to active status. Reactivation does not restore any previous child assignments or alter the room's name, capacity, or age group.

## Primary Room

The single room a child is operationally attached to for day-to-day care. One child has at most one primary room at any time. The primary room is the source of truth for room occupancy: a room's `assigned_count` is the number of children whose `primary_room_id` points at the room. Editing a child's primary room re-points them; no effective-dated history is kept. Replaces prior language of "assigned to a room".

A child's primary room is set at intake and is required to proceed through the guided registration stepper. After intake, the primary room remains editable from the section editor (and from the inline child edit) via `PATCH /api/v1/children/:id`.

## Unassigned Child

A child whose `primary_room_id IS NULL`. Unassigned children do not contribute to any room's occupancy and do not block archiving the rooms they are not attached to. Existing children begin as unassigned after the column is introduced; managers assign them from the Children form.

## Retention Policy Scope

Configurable retention/deletion policy automation is deferred; pilot records are retained with core no-hard-delete rules.

## UAT Signoff Gate

Go-live requires explicit manager signoff on attendance, correction, invoice generation, payment, and payment-retry user acceptance checks.

## Go-Live Rollback Policy

If a critical billing defect appears at go-live, new invoice issuance is paused and fallback procedures are used until a verified fix is deployed.

## Scope Change Rule

During the 30-day delivery window, only changes that unblock the defined success metric are accepted; all other requests are deferred to the post-release backlog.

## Authorization Model

Authorization combines role checks with scope checks, including tenant scope, branch scope, and parent-child linkage where applicable.

## Authorization Guards

Authorization guards are the combined enforcement layer for role, scope, and relationship checks; RBAC is one component of this guard model.

## Authorization Check Taxonomy

Scope checks validate tenant and branch boundaries from the active session membership, while relationship checks validate dynamic record linkage such as parent-child access.

## Authorization Evaluation Rule

Session-bound scope and role claims are used for baseline access checks, while dynamic relationship checks such as parent-child linkage are validated against current records.

## Parent Relationship Check Freshness

Parent access to child-linked resources is authorized against current parent-membership-to-child mappings at request time.

## Authorization Route Policy

Protected endpoints follow default-deny behavior, with explicit role and scope requirements declared per route.

## Route Role Declaration Style

Current protected endpoints declare explicit allowed role lists per route; capability abstraction is deferred.

## Unknown Role Handling

If a session presents an unknown or unsupported role claim, authorization fails closed with a forbidden response and a stable role-specific error code.

## Authorization Denial Precedence

When multiple authorization checks fail, guards return one deterministic primary denial code based on a fixed evaluation order.

## Authorization Check Order

Authorization evaluation order is authentication validity first, then role checks, then tenant/branch scope checks, then relationship checks.

## Authorization Layer Boundary

Middleware enforces token validity and coarse route-level role and scope checks, while domain logic enforces resource-specific relationship authorization.

## Authorization Context Propagation

Authorization middleware constructs a normalized request authorization context (user, membership, role, tenant, branch, request identifier) for downstream handlers and services.

## Scope Source of Truth

Effective tenant and branch scope is derived from the active session membership, and client-supplied scope fields are rejected when they conflict.

## Session Scope Isolation Rule

Operations cannot cross into another branch or tenant within the same session; users must switch membership scope before acting in a different scope.

## Attendance Edit Authority

Historical attendance events are not edited directly by practitioners; corrections are captured as manager-only correction events.

## Parent-Child Mapping End Visibility Rule

When a parent-membership-to-child mapping is ended, that parent immediately loses access to that child's invoices, including historical invoices.

## Parent-Child Relink Visibility Rule

When a parent-membership-to-child mapping is recreated for a pair that previously had one, parent visibility for that child's invoices is restored based on the current active-mapping relationship check.

## Draft Invoice Idempotency

Regenerating draft invoices for the same month does not create duplicates; existing eligible drafts are replaced or updated unless already issued.

## Draft Invoice Regeneration Identity

Regenerating an existing draft monthly invoice preserves the draft invoice identity while replacing its draft calculation contents.

## Draft Invoice Generation Audit Scope

Draft invoice generation records audit history for each draft invoice that is created or recalculated; blocked child-months are represented as invoice run exceptions because no invoice changes for those child-months.

## Invoice Issue Audit Scope

Invoice issue records audit history for each invoice that becomes issued; blocked invoices are represented as invoice run exceptions because no invoice changes for those invoices.

## Invoice Run History

Repeated draft generation requests create separate invoice run history entries even when they update the same child-month draft invoice.

## Issued Invoice Regeneration Policy

Issued invoices are not regenerated after attendance changes; post-issue changes require an explicit adjustment flow.

## Draft Regeneration Issue Race Rule

If a draft monthly invoice becomes issued while draft regeneration is being prepared, issued status takes precedence and the invoice is skipped rather than recalculated.

## Stale Draft Blocked Regeneration

When an existing draft monthly invoice is no longer eligible for regeneration, the draft is left in place and the child-month is returned as blocked rather than silently removed.

## Adjustment Flow

Post-issue billing changes are represented by a manager-created follow-up adjustment invoice linked to the original issued invoice, with a required reason. Current operation may defer creating adjustment invoices until a validated pilot need exists.

## API Versioning

HTTP endpoints are published under a versioned `/api/v1` route prefix.

## Public Route Allowlist

Only health and authentication endpoints are public in the current release; all other API routes require authorization guards.

## Invite Acceptance Route Visibility

Invitation acceptance is a public provisioning route because invitees do not have a session before acceptance; all manager invitation management routes remain manager-protected.

## Deployment Model

Production runs on a single virtual machine using Docker Compose for service orchestration.

## Entity Identifier Strategy

Domain entities use UUID primary keys (UUIDv7 preferred, UUIDv4 acceptable).

## Invoice Explainability Persistence

Invoice line storage preserves both intermediate billing components (core attended minutes, funded deduction minutes, core billable minutes, hourly rate) and final totals.

## Invoice Attendance Source Snapshot

A generated draft invoice preserves a compact snapshot of the attendance sessions used to calculate it so managers can explain how the billed minutes were derived.

## Draft Invoice Calculation Lines

Generated draft monthly invoices use consistent explanatory lines for core childcare and funded-hours deduction, including a zero-value funded deduction line when no deduction amount is applied.

## Parent Draft Invoice Visibility

Draft invoices are visible only to managers; parents can view invoices only after they are issued.

## Mid-Month Leave Billing

If a child leaves during a month, billing is automatically derived from attendance actuals up to the leave date.

## Post-Leave Invoice Artifact Policy

Marking a child inactive/left does not trigger automatic voiding of unrelated invoice artifacts; future billing is naturally blocked by attendance-derived invoice generation.

## Child Management

Manager maintains the child record and surrounding profile in a guided stepper after a parent or carer has completed the physical form. A child is the root identity; everything else (profile, contacts, health, safeguarding, consent, funding, collection, room placement, billing) is a sub-record of that child. There is no separate registration workflow — the create flow runs as a single transaction.

## Child Management Atomic Create

Manager creates a child in a single transaction that includes identity, profile, contacts, health, safeguarding, consent, funding, collection settings, room placement, and billing profile. All sections are optional except identity, room, and the safeguarding acknowledgement. The stepper in the UI keeps its "answer required to proceed" rule, but the API does not enforce that — partial submission is allowed and the manager can fill in the rest later through the per-resource PATCH/PUT endpoints.

## Child Room Placement History

`child_room_assignments` is 1:many per child; one row is current (`end_date IS NULL`). Moving a child to a different room closes the current row and inserts a new one in the same transaction. The room id lives on the assignment, not on the child.

## Child Leaving Record

Written when a child is marked inactive via `POST /api/v1/children/:id/actions/mark-inactive`; the reason code is constrained to `duplicate_record | entered_in_error | left_nursery | safeguarding_direction | contact_update | access_revoked | other`. The reason note is required when `reason_code = 'other'`. The `children` table no longer carries a `left_at` or `left_reason_code` column; the mark-inactive use case writes a `child_leaving_records` row and updates `children.is_active = false` in one transaction.

## Child Funding Record

A per-child eligibility record (`child_funding_records`) capturing 15/30h, 2yo, tax-free childcare, benefits, and support notes. It is distinct from `funding_profiles` (per-billing-month funded allowance minutes used by invoicing). Both tables coexist; the funding record does not change invoice generation.

## Child Collection Settings

The collection password hash and acknowledgement (`child_collection_settings`) are 1:1 with the child. The API only exposes password presence and last-updated metadata on read; the stored secret is never returned. `PUT /api/v1/children/:id/collection-settings` accepts an optional password (hashed with bcrypt on the server) and an `over_18_collection_acknowledged` flag.

## Child Profile Audit Redaction

Per-sub-resource audit records prove who changed which section and when. They do not copy sensitive values (medical notes, safeguarding notes, contact details, collection password material) into `Details` on the audit log row.

## Consent Decision Tier

The product classification of a single consent item on Step 4 of the manager child-registration stepper. Each item is one of **Required**, **Required-acknowledged**, or **Optional**, and the tier governs both the validation rules and the visual badge the manager sees.

## Required Consent

A consent item the parent must explicitly grant (tick Yes) for the child's registration to be considered complete. In the current release this is GDPR data processing consent and the information truthfulness declaration. A Required item that is not Yes is a blocking completion issue.

## Required-acknowledged Consent

A consent item the parent must explicitly answer Yes or No for the child's registration to be considered complete. In the current release this is safeguarding reporting acknowledgement, information sharing consent, urgent medical treatment, and first aid/plasters. A No answer is recorded truthfully, does not block completion, is surfaced in the review summary with a clear warning, and captures an optional free-text reason in the audit trail.

## Optional Consent

A consent item the parent may grant or decline at their discretion. The child's registration can be completed without recording a decision. In the current release this covers professional liaison (Area SENCO, Health Visitor, Transition documents), activities (local outings, face painting, parent-supplied sun cream, parent-supplied nappy cream), and photographs & social media (development records, display boards, promotional literature, website, staff/student coursework, social media).

## Consent Record Audit Trail

The evidence captured on the consent record itself in addition to the system audit log: **signed_by** (the parent/carer full name who gave the consent) and **date_signed** (when the consent was given, date-only). Both fields are required for the consent record to be considered complete and are distinct from the manager's session identity recorded in the audit log.

## Consent Review Completion

The state where a child's consent record has every Required item = Yes, every Required-acknowledged item touched (Yes or No), and the consent record audit trail (signed_by + date_signed) is filled. Optional items do not contribute to the completion check. The review-and-create step in new-registration mode and the Save Changes action in edit mode both surface this state.

## Consent Reason for Change

A one-line free-text note a manager may add when saving an edit to a consent record that changes a value from the previously saved state. The note is passed to the API and recorded as part of the audit event for the change. Reason for change is optional and is only requested when at least one consent value differs from the previously saved value; it is not requested for first-time saves during new registration.

