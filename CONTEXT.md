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

## Membership

A user's role-bearing participation in one tenant and one branch. A user may have multiple active memberships, but each authenticated session acts through exactly one membership.

## Role Capability Inheritance (MVP)

Manager permissions include practitioner attendance actions within the same active session scope.

## Funding v1

A simple funded-hours deduction model used to reduce monthly billed amount per child.

## Funding Profile (MVP)

A child's funded-hours allowance for a single billing month.

## Missing Funding Profile (MVP)

A child-month without a recorded funding profile. This is distinct from a funding profile with a zero-minute allowance.

## Missing Funding Profile Invoice Block (MVP)

A missing funding profile blocks invoice draft preflight for that child-month; an explicit zero-minute allowance does not.

## Funding Profile Enrollment Scope (MVP)

Funding profiles are valid for billing months that overlap the child's enrollment window, including historical months after the child has left.

## Funding Profile Partial-Month Rule (MVP)

A funding profile may exist when the billing month overlaps the child's enrollment window by at least one day.

## Funding Profile Audit Scope (MVP)

Funding profile creation and allowance changes are persisted as audit events, while unchanged idempotent saves are not.

## Funding Profile Invoice Snapshot Boundary (MVP)

Changing a funding profile after an invoice is issued does not change that issued invoice; issued invoice lines preserve the funded allowance applied at issue time.

## Funding v1 Metadata Scope (MVP)

Funding v1 records only the month-specific allowance; funding source, entitlement type, evidence, stretched funding, term-time rules, and notes are deferred.

## Invoice

A monthly billing statement showing gross fees, funded deduction, and final amount due.

## Monthly Invoice (MVP)

An invoice for one child and one calendar billing month. A child has at most one monthly invoice for the same billing month within a nursery branch.

## Invoice Line (MVP)

A charge, deduction, extra, or adjustment entry that explains how an invoice total was formed. Invoice lines preserve the billing calculation used for the invoice they belong to.

## Parent Account Provisioning (MVP)

Parent user accounts are created by manager invitation only; public self-signup is not used in month 1.

## Manager Invite Access Grant Timing (MVP)

A manager invitation does not grant usable access by itself; invited access becomes usable only after the invitee accepts the invitation and sets a password.

## Manager Invite Membership Activation Scope (MVP)

Accepting a manager invitation creates a new active membership for a new login identity; invitation acceptance does not reactivate existing inactive users or inactive memberships in month 1.

## Manager Invite Session Start (MVP)

Accepting a manager invitation does not automatically start an authenticated session; invitees sign in through the normal login flow after acceptance.

## Guardian Identity Separation (MVP)

Guardian records store child relationship and contact data independently from authentication users; portal access exists only when a guardian is linked to a parent-role user membership.

## Parent Membership Guardian Mapping (MVP)

Within a tenant-branch scope, a parent membership maps to at most one guardian record.

## Parent Mapping Change Flow (MVP)

Changing an active parent membership-to-guardian mapping requires explicitly ending the current mapping (with reason) before creating a new active mapping; implicit in-place replacement is not used.

## Parent Membership End Cascade (MVP)

Ending a parent membership ends any active parent-membership-to-guardian mapping in the same action so no dangling active mapping remains.

## Parent Membership End Cascade Reason Attribution (MVP)

When parent membership end cascades to end an active parent-membership-to-guardian mapping, the mapping stores an explicit system cascade reason code so automatic effects are distinguishable from direct manager-initiated end actions.

## Parent Mapping End Visibility Rule (MVP)

Ending an active parent-membership-to-guardian mapping immediately removes that parent membership's access to child-linked resources reachable through that guardian relationship.

## Parent Mapping Idempotent Create (MVP)

Creating a parent-membership-to-guardian mapping for a pair that is already active is treated as idempotent success, while attempts to map that membership to a different active guardian require explicit end-then-remap flow.

## Parent Mapping Active-Entity Requirement (MVP)

Creating an active parent-membership-to-guardian mapping requires both the membership and guardian to be active at mapping time.

## Guardian Contact vs Login (MVP)

Guardian contact records may exist without an email address, while user login identity requires a unique normalized email on the user account.

## Parent-Guardian Email Independence (MVP)

Parent membership-to-guardian mapping does not require the parent user's login email to match the guardian contact email.

## Guardian Email Auto-Link Policy (MVP)

Entering or editing a guardian contact email does not automatically link that guardian to any user login; parent portal access is granted only through explicit invitation and membership-to-guardian mapping.

## Parent Invite Mapping Separation (MVP)

Accepting a parent invitation creates login-ready parent membership access but does not automatically map that membership to a guardian; guardian access still requires an explicit parent-membership-to-guardian mapping.

## Contact Detail Scope (MVP)

Child and guardian records carry only minimal operational contact details in month 1; richer profile/contact modeling is deferred until after pilot validation.

## Guardian Creation Minimum Data (MVP)

Guardian creation requires only full name in month 1; email and phone are optional contact details.

## Guardian-Child Link (MVP)

A relationship showing that a guardian record is connected to a child within the same nursery scope. A child may have multiple guardian-child links, and active links contribute to parent access and enrollment completeness; ended links no longer grant access.

## Manager Provisioning Authority (MVP)

Manager role assignment is reserved to administrative bootstrap flows in month 1; manager-invited users are limited to non-manager roles.

## Attendance Record (MVP)

Attendance is captured as an event history (check-in, check-out, correction) instead of a single mutable interval row. Use check-in and check-out rather than sign-in or sign-out.

## Attendance Session (MVP)

An attendance session is one continuous period of a child's attendance, beginning with check-in and ending with check-out.

## Attendance Open Session Rule (MVP)

A child may have multiple attendance sessions on the same local day, but may have only one open attendance session at a time. Avoid the phrase active check-in for this rule.

## Attendance Daily List Scope (MVP)

The attendance-facing child list shows children for the current `Europe/London` local day: active children and any child with an open attendance session that still needs resolution. Avoid treating this as a historical attendance report.

## Attendance Correction Target (MVP)

An attendance correction applies to the effective attendance session for one child rather than rewriting the original check-in or check-out event. A correction may adjust an existing session or record a missed session that should have existed.

## Attendance Correction Scope (MVP)

Attendance correction changes or establishes the full effective check-in and check-out interval for a session; voiding or excluding a session from billing is a separate deferred concept.

## Duplicate Entry Attendance Correction (MVP)

A duplicate-entry attendance correction still establishes a valid effective attendance interval; it does not void, delete, or exclude a session from billing in month 1.

## Attendance Correction History (MVP)

A session may receive more than one attendance correction; each correction remains part of the historical trail while the latest correction determines the current effective interval.

## Attendance Correction Recorded Time (MVP)

The time and local day a correction is recorded are the manager action time and action day, distinct from the corrected attendance interval.

## Attendance Correction Authority (MVP)

Only managers can create attendance correction events.

## Attendance Correction Reason Vocabulary (MVP)

Only attendance corrections carry attendance reason codes; routine check-in and check-out events do not. Attendance corrections use attendance-specific reason codes rather than lifecycle reason codes; the starter set is `missed_check_in`, `missed_check_out`, `incorrect_time`, `duplicate_entry`, and `other`.

## Attendance Correction Audit Reason Semantics (MVP)

Attendance correction reason codes are distinct from lifecycle reason codes even when an audit trail records the correction.

## Invoice Source of Truth (MVP)

Monthly invoice billable minutes are derived from attendance actuals.

## Funding Application Rule (MVP)

Funded-hours deduction applies only to core childcare hours; extras remain fully payable.

## Core Attended Minutes (MVP)

Core attended minutes means rounded core attendance minutes for billing and funding deduction, not raw elapsed attendance minutes.

## Funded Deduction Minutes (MVP)

Funded deduction minutes are the portion of core attended minutes covered by the child's funded allowance for the billing month.

## Core Billable Minutes (MVP)

Core billable minutes are the remaining core attended minutes after funded deduction minutes are applied.

## Core Billing Formula (MVP)

Core due hours are calculated as `max(0, rounded core attendance minutes - funded hours allowance)` before pricing is applied. Rounded core attendance minutes are the sum of per-session billable minutes rather than raw elapsed minutes.

## Core Billing Price Rounding (MVP)

Core childcare amounts are converted from minutes and hourly minor-unit rates by rounding any fractional minor unit up to the next minor unit.

## Attendance Billing Rounding (MVP)

Each attendance session is rounded up to the nearest 15 minutes for billing. Any positive elapsed-time remainder beyond a 15-minute boundary rounds up; exact 15-minute boundaries do not add another block.

## Billing Timezone (MVP)

Attendance day boundaries and invoice-period calculations use `Europe/London` local time.

## Attendance Timestamp Semantics (MVP)

Attendance event times are captured as absolute instants while attendance day grouping is derived from `Europe/London` local dates.

## Incomplete Attendance Handling (MVP)

Attendance records missing check-out are excluded from automatic billing until a manager correction establishes the full effective attendance interval.

## Invoice Generation Flow (MVP)

Managers manually generate monthly draft invoices before any invoice is issued.

## Invoice Draft Preflight (MVP)

A manager-facing readiness preview for one billing month before draft invoice generation. It identifies which child-months are eligible or blocked without being an invoice run.

## Invoice Draft Preflight Population (MVP)

Invoice draft preflight considers child-months whose enrollment window overlaps the billing month, even if the child is no longer currently active.

## Invoice Draft Preflight Existing Invoice Rule (MVP)

An existing draft monthly invoice does not block invoice draft preflight for that child-month, while an issued-or-later monthly invoice does.

## Invoice Draft Preflight Totals (MVP)

Invoice draft preflight totals are estimated aggregate invoice amounts for eligible child-months before draft invoice generation.

## Zero-Attendance Invoice Eligibility (MVP)

A child-month with no completed attendance sessions can still pass invoice draft preflight when enrollment, billing, funding, and attendance-completeness data are present.

## Zero-Total Draft Invoice (MVP)

An eligible child-month can produce a draft monthly invoice with zero amount due; zero-total drafts still represent the monthly billing statement for that child-month.

## Zero-Total Invoice Issue (MVP)

A zero-total draft invoice can be issued like any other eligible draft invoice and receives an invoice number.

## Payable Invoice (MVP)

A parent-visible issued-or-later invoice with a positive outstanding balance. Zero-total issued invoices and paid invoices are not payable.

## Invoice Issue Mode (MVP)

Managers can issue invoices one-by-one or in bulk; the default flow is bulk issue with confirmation.

## Invoice Issue Confirmation (MVP)

A manager's explicit approval that selected draft invoices should become immutable issued invoices. Confirmation applies to both one-by-one and bulk issue actions.

## Invoice Issue Time (MVP)

The business instant when a draft invoice becomes an issued invoice. In month 1, the invoice is also locked and due at that same instant.

## Invoice Issue Validation Boundary (MVP)

Issuing validates that the invoice is an existing draft invoice in the manager's billing scope. It does not recalculate billing readiness; managers regenerate drafts before issue when source data changes.

## Bulk Invoice Issue (MVP)

A manager-triggered issue action for all draft monthly invoices in one billing month or a manager-selected subset of those drafts. Bulk issue requires explicit manager confirmation before invoices become issued.

## Invoice Run (MVP)

A manager-triggered monthly billing operation that prepares or issues per-child invoices as a batch. An invoice run may include successful invoices and child-specific exceptions.

## Single Invoice Issue Run (MVP)

A manager-triggered invoice run that issues exactly one draft invoice. It is still recorded as an invoice run so issued invoice history is consistent across single and bulk issue paths.

## Issued Invoice Edit Policy (MVP)

Issued invoices are immutable; changes require explicit adjustment rather than direct edits.

## Adjustment Invoice (MVP)

A follow-up invoice that corrects or offsets a previously issued invoice. An adjustment invoice must be linked to the issued invoice it adjusts and must carry a manager-provided reason.

## Payment Scope (MVP)

Parents pay invoices in full; partial payments are not supported in month 1.

## Payment Attempt (MVP)

A single try to collect a payable invoice through the payment provider. Each retry creates a new payment attempt for the same invoice.

## Checkout Retry Availability (MVP)

A manager-visible indication that a parent can currently start a fresh payment attempt for a payable invoice. It is diagnostic only and is not itself checkout session creation.

## Payment Outcome Authority (MVP)

Invoice payment state changes only after a payment provider-confirmed outcome. Browser return or cancel navigation after payment initiation is not itself a payment outcome.

## Payment Reconciliation Record (MVP)

A manager-facing record of a handled payment outcome for one invoice payment attempt. It explains whether the attempt paid, failed, expired, or was ignored without relying on raw provider webhook payloads as the operational timeline.

## Payment Event (MVP)

A manager-facing payment timeline item backed by a payment reconciliation record. Avoid using payment event to mean the raw payment provider webhook payload.

## Webhook Processing Status (MVP)

The local result of processing a verified payment provider webhook event, such as processed, ignored, or rejected. It is distinct from the provider's webhook delivery status.

## Payment Failure State (MVP)

Failed or canceled payment attempts move invoices to a `payment_failed` state.

## Invoice Numbering (MVP)

Invoice identifiers follow `INV-YYYYMM-####` sequence format.

## Invoice Numbering Month (MVP)

The `YYYYMM` segment of an invoice number is the invoice billing month, not the calendar month when the manager issues it.

## Bulk Invoice Issue Sequence Order (MVP)

Bulk-issued invoices receive invoice numbers in deterministic child-name order, with invoice identity used only as a tie-breaker.

## Invoice Granularity (MVP)

Invoices are issued per child, not combined at family level.

## Sibling Discount Policy (MVP)

Sibling discounts are deferred and not part of month 1 billing rules.

## Extras Charging Model (MVP)

Extras are added manually as invoice line items by managers.

## Draft Invoice Extras Placeholder (MVP)

Draft invoice generation does not create payable extras automatically; it preserves an explicit zero-value placeholder so later manager-entered extras remain outside attendance and funding calculations.

## Draft Invoice Manual Extras Preservation (MVP)

Regenerating a draft monthly invoice preserves existing manual extra lines while recalculating attendance-derived core childcare and funded-hours deduction lines.

## Attendance Capture Scope (MVP)

Attendance capture includes check-in and check-out only; room-move tracking is out of scope.

## Attendance Capture Time Authority (MVP)

Routine practitioner check-in and check-out capture uses the current server time; historical or custom attendance times are manager-controlled corrections.

## Practitioner Attendance Scope (MVP)

Practitioners can perform check-in and check-out for any child within the active session branch.

## Practitioner Contact Visibility (MVP)

Practitioner attendance workflows expose only attendance-facing child information and do not expose guardian contact details such as email or phone in month 1.

## Child and Guardian Write Authority (MVP)

Child/guardian and relationship write actions (create, update, deactivate, link, unlink, mapping changes) are manager-only in month 1, while practitioner access remains read-only for attendance-facing child views.

## Absence Marker (MVP)

An absence marker records that one child was absent for one `Europe/London` local date. It is separate from attendance sessions and attendance corrections, and it does not create, edit, void, or exclude any attendance session; a child-day cannot have both an absence marker and attendance.

## Clear Absence Marker (MVP)

Clearing an absence marker means removing the active absent state for a child-day while preserving the historical record that the marker existed.

## Absence Marker Authority (MVP)

Managers and practitioners may create and clear absence markers. Parents do not have access to absence marker workflows.

## Absence Marker Detail Scope (MVP)

Absence markers do not carry reason codes, notes, charge policy, or funding policy in month 1.

## Absence Marker Date Scope (MVP)

Absence markers are used only for the current `Europe/London` local day in month 1; future absence scheduling is out of scope.

## Billing Period (MVP)

Billing runs monthly on calendar-month boundaries.

## Billing Month (MVP)

A calendar month identified as `YYYY-MM` for funding and invoice workflows.

## Billing Quantity Unit (MVP)

Attendance-derived billing quantities use integer minutes as the canonical unit. Hours are a display format only.

## Funded Allowance Unit (MVP)

Funded-hours allowances are stored and exchanged as integer minutes. Hours are a display format only.

## Funded Allowance Bounds (MVP)

A funded-hours allowance may be zero but must not exceed the number of minutes in a 31-day month.

## Child Enrollment Minimum (MVP)

A child requires name, date of birth, start date, one linked guardian, and a billing rate before attendance and invoicing flows begin.

## Enrollment Gate Scope (MVP)

Enrollment minimum checks gate all new routine attendance capture and invoicing actions, while manager-only historical attendance corrections remain allowed under the post-enrollment correction policy.

## Child Creation Flow (MVP)

Managers may create a child record before linking a guardian, but attendance and invoicing remain blocked until all child enrollment minimum requirements are satisfied.

## Child Billing Rate Source (MVP)

Each child has one current core billing rate in enrollment data, while issued invoices preserve the applied rate in invoice lines for historical explainability.

## Draft Invoice Rate Source (MVP)

Draft invoice generation uses the child's current stored core billing rate at generation time, including when generating drafts for historical billing months.

## Zero Core Billing Rate (MVP)

A core billing rate of zero is an explicit valid rate, not missing billing data.

## Child Enrollment Lifecycle (MVP)

Child records remain retained when enrollment ends; the child can be marked inactive/left while attendance and billing history stays intact.

## Child Lifecycle Reason Requirement (MVP)

Manager-initiated child lifecycle transitions such as marking a child inactive/left require a stable reason code with an optional note.

## Child Lifecycle vs Enrollment Date Separation (MVP)

Child lifecycle transitions (active/inactive-left) are managed independently from enrollment date edits; enrollment boundary dates are updated through explicit date-change flows with enrollment window integrity checks.

## Child Reactivation Deferral (MVP)

Month-1 lifecycle APIs include explicit guardian reactivation, while explicit child reactivation endpoints are deferred unless pilot operations require them.

## Child and Guardian Default Listing Scope (MVP)

Manager child and guardian listings default to active records only, with an explicit option to include inactive or ended records for historical administration.

## Child Identity Uniqueness (MVP)

Child identity is anchored by UUID entity identifiers, and matching name/date-of-birth combinations are not enforced as a hard uniqueness rule.

## Post-Enrollment Attendance Correction (MVP)

After enrollment ends, managers may still record corrections for historical attendance sessions so billing derived from attendance actuals remains accurate. Historical corrections require the child to exist in scope but do not require current active enrollment, while the corrected attendance interval remains constrained to the child's enrollment dates.

## Enrollment Date Semantics (MVP)

Enrollment boundaries use date-only fields (start and optional end date), while billing calculations continue to derive from attendance timestamps in `Europe/London`.

## Enrollment Window Integrity (MVP)

Enrollment date updates are rejected when they would place existing attendance records outside the child's enrollment window.

## Guardian-Child Link Cardinality (MVP)

Guardian-child relationships are many-to-many: a child may link to multiple guardians, and a guardian may link to multiple children.

## Parent Visibility Scope (MVP)

One parent account can be linked to multiple children and can view invoices for linked children within the active session scope.

## Parent Portal (MVP)

The parent-side product surface where a parent-role user views issued invoices for linked children and initiates full invoice payment.

## Parent Invoice Payment Initiation (MVP)

Only parent-role users initiate payment for parent-visible payable invoices in month 1. Managers can review invoice and payment state but do not initiate parent payment.

## Login Identifier (MVP)

User authentication uses a unique, case-insensitive normalized email address as the login identifier.

## Password Credential Policy (MVP)

User passwords must be at least 8 characters in month 1; additional composition rules are not enforced.

## Pilot Tenant Bootstrap (MVP)

The first manager account for a tenant is created through a one-time administrative seed command; subsequent users are provisioned by manager invitation.

## Audit Baseline Scope (MVP)

Audit logs are mandatory for user provisioning changes, child record updates, attendance events and corrections, funding profile updates, invoice draft/issue actions, and payment-status updates.

## Audit Request Correlation (MVP)

Persisted audit events include request identifier correlation so domain changes can be traced to individual API requests.

## Audit Actor Semantics (MVP)

Audit events include actor membership identity for user-initiated actions and may omit actor identity only for system-initiated actions, while tenant and branch scope plus action metadata remain required.

## Manager Invite Audit Scope (MVP)

Manager invitation creation, resend, revocation, and acceptance are persisted as audit events because they change user provisioning state.

## Manager Invite Acceptance Actor Semantics (MVP)

Invitation acceptance audit events are attributed to the newly provisioned user and membership created by that acceptance, while manager-initiated invitation actions are attributed to the manager's active session actor.

## Cross-Month Session Allocation (MVP)

An attendance session that crosses midnight is allocated to the invoice month of its check-in date using `Europe/London` local time.

## API Response Style (MVP)

API endpoints return plain JSON resources with standard HTTP status codes instead of a global response envelope.

## API Error Contract (MVP)

API errors use a consistent JSON structure with stable error code, human-readable message, optional details, and request identifier.

## Lifecycle Error Code Baseline (MVP)

Child, guardian, guardian-child-link, and parent-mapping lifecycle endpoints use a stable domain error-code set so clients can handle expected lifecycle failures deterministically.

## Authorization Error Status Policy (MVP)

Authorization failures return `403 Forbidden` with stable authorization-specific error codes for explicit role, scope-selection, and relationship failures, while opaque resource identifiers outside the active tenant-branch scope are treated as not found so resource existence is not disclosed across scopes.

## Authorization Error Disclosure Policy (MVP)

Authorization errors expose stable machine-readable denial codes while keeping human-readable messages generic.

## Authorization Denial Code Baseline (MVP)

The initial stable authorization denial code set includes role, scope, unknown-role, parent-child-link, and scope-selection denial variants alongside unauthorized authentication failure.

## Scope Selection Error Semantics (MVP)

Missing or malformed membership selection at authentication time returns a validation error, while a well-formed selection outside the user's allowed memberships returns an authorization error.

## Persistence Strategy (MVP)

Data access uses `sqlc` with `pgx` and typed SQL queries rather than an ORM.

## Migration Workflow (MVP)

Schema changes are applied with manual `golang-migrate` commands against local PostgreSQL and are not auto-run on API startup.

## Session and Token Policy (MVP)

Authentication uses short-lived access JWTs plus database-backed refresh tokens that can be revoked.

## Session Scope Mode (MVP)

Each active session is bound to exactly one membership scope (tenant, branch, and role) rather than selecting scope per request.

## Session Scope Selection Rule (MVP)

When a user has multiple memberships, the active session scope is chosen explicitly by the user at login and must belong to that user.

## Session Single-Scope Auto-Selection (MVP)

Authentication auto-selects the active membership scope when exactly one active membership is available.

## Session Scope Selection Identifier (MVP)

Session scope selection uses the membership identifier as the canonical selector instead of a tenant-branch-role tuple.

## Session Scope Requirement (MVP)

Authentication does not create a session unless the user has at least one active membership scope available.

## Membership Breadth (MVP)

A user may hold memberships across multiple tenants and branches, but each active session is bound to exactly one selected membership scope.

## Single Role Per Scope (MVP)

Within a single tenant-branch scope, a user holds exactly one membership role; concurrent dual-role memberships in the same scope are not used in month 1.

## Registered Email Invite Deferral (MVP)

Manager invitations in month 1 are used only for new login identities; inviting an email address that already belongs to a user is deferred until an authenticated existing-user membership invite flow exists.

## Membership Activity State (MVP)

Authentication and authorization consider only active memberships; inactive memberships cannot be selected for sessions.

## Session Scope Visibility (MVP)

Authentication responses include both the active session scope and the user's available scopes so clients can present explicit scope switching.

## Session Membership Binding (MVP)

Each refresh-token-backed session is bound to one membership scope and cannot change scope unless an explicit membership switch request is validated.

## Session Refresh Scope Integrity (MVP)

Token refresh succeeds only when the session's bound membership remains active; refresh does not auto-fallback to a different membership.

## Session Claim Shape (MVP)

Access tokens include subject user identifier, membership identifier, tenant identifier, branch identifier, role, and standard token timestamps so each request can be unambiguously authorized against one active membership.

## Session Scope Switch Flow (MVP)

Changing the active membership scope is an explicit session action and is distinct from routine token refresh.

## Session Scope Switch Authentication (MVP)

Membership scope switching is performed through a refresh-token-backed session action that rotates session tokens.

## Session Action CSRF Policy (MVP)

Refresh-token-cookie-backed session actions, including refresh, logout, and membership switch, require CSRF protection in month 1.

## Logout Idempotency (MVP)

Session logout is idempotent and returns success even when no active refresh-token session is present.

## Session CSRF Mechanism (MVP)

Cookie-backed session actions use a double-submit CSRF token pattern with a client-readable CSRF token echoed in a request header and a trusted origin or referer check.

## Session Scope Switch Auditing (MVP)

Membership scope switch actions are persisted as audit events with actor, previous scope, new scope, and request identifier.

## Authentication Event Persistence Scope (MVP)

Login, refresh, logout, and password reset activity are treated as authentication/security telemetry in structured logs and metrics rather than persisted as audit-log domain events in month 1.

## Token Transport Policy (MVP)

Access tokens are sent as bearer tokens, while refresh tokens are stored in secure HttpOnly cookies.

## Session Concurrency (MVP)

Users may hold multiple active sessions concurrently, with token revocation supported per session.

## Session Persistence Model (MVP)

Session state is persisted as refresh-token-backed, revocable session records; a second parallel session artifact is not introduced in month 1.

## Membership Change Revocation Policy (MVP)

When membership scope or role is changed, refresh tokens for affected sessions are revoked immediately, while already-issued access tokens remain valid only until their normal expiry.

## Password Reset Session Revocation Policy (MVP)

After a successful password reset, all refresh-token-backed sessions for that user are revoked immediately, while already-issued access tokens remain valid only until their normal expiry.

## Token Lifetime Policy (MVP)

Access tokens expire after 15 minutes; refresh tokens expire after 30 days and rotate on refresh.

## Password Reset (MVP)

Users can reset passwords through a basic email-link reset flow.

## Password Reset Account Scope (MVP)

Password reset applies to the global user login identity rather than to an individual membership, tenant, or branch scope.

## Password Reset Request Disclosure Policy (MVP)

Password reset requests do not reveal whether a login identifier belongs to an active user account; eligible active users receive reset email while inactive or unknown users do not, and the requester sees a generic outcome.

## Password Reset Link Supersession (MVP)

Only the newest unused password reset link for a user remains valid; creating a new reset request invalidates earlier unused reset links for that user.

## Password Reset Link Lifetime (MVP)

Password reset links are short-lived recovery links that remain valid for 60 minutes from creation.

## Password Reset Link Issuance (MVP)

A password reset link is considered issued only when the reset message is accepted for delivery; failed delivery does not leave a usable reset link behind.

## Password Reset Token Exposure Policy (MVP)

Raw password reset tokens are delivered only through the reset message and are never returned by API responses.

## Manager Invite Token Exposure Policy (MVP)

Raw manager invitation tokens are delivered only through the invitation message and are never returned by API responses.

## Password Reset Request Throttling (MVP)

Password reset requests are rate-limited by login identifier and client source so the public recovery flow cannot be used for unbounded email sending.

## Manager Invite Acceptance Throttling (MVP)

Public invitation acceptance attempts are rate-limited by client source to reduce token-guessing risk.

## Password Reset Token Error Semantics (MVP)

Invalid, expired, and already-consumed password reset links return distinct stable error codes with generic human-readable messages; superseded reset links are treated as already consumed.

## Manager Invite Link Lifetime (MVP)

Manager invitation links expire after seven days from issuance.

## Manager Invite Link Supersession (MVP)

Resending a pending manager invitation issues a new invitation link and invalidates the prior unused link for that invitation.

## Manager Invite Link Issuance (MVP)

A manager invitation link is considered issued only when the invitation message is accepted for delivery; failed delivery does not leave a newly generated usable invite link behind.

## Manager Invite Idempotent Create (MVP)

Creating a manager invitation for the same normalized email, role, tenant, and branch as an existing pending invitation refreshes that pending invitation instead of creating a duplicate; creating one for the same email and scope with a different role is a conflict.

## Manager Invite Historical Reissue (MVP)

Creating a new manager invitation after an earlier invitation for the same email, role, and scope was revoked or expired creates a new invitation record rather than reopening the historical invitation.

## Manager Invite Minimum Data (MVP)

Creating a manager invitation requires only the invitee login email and invited role in month 1; staff profiles, display names, and guardian linkage are not collected by the invitation flow.

## Manager Invite Scope Source (MVP)

Manager invitations target the manager's active tenant-branch session scope; clients do not choose invitation tenant or branch in the request body.

## Manager Invite List Visibility (MVP)

Manager invitation lists default to currently pending invitations in the active tenant-branch scope, while accepted, revoked, expired, and all-invite history are available only through explicit status filtering.

## Manager Invite Expiry State (MVP)

Invitation expiry is derived from the invitation expiry time rather than a separate manager action or background lifecycle transition.

## Manager Invite Revocation (MVP)

Managers may revoke pending invitations, including pending invitations whose links have already expired; revoking an already revoked invitation is idempotent, while accepted invitations cannot be revoked.

## Manager Invite Revocation Reason Scope (MVP)

Manager invitation revocation does not require a reason code or note in month 1; the audit event records the actor and invitation details without using lifecycle reason fields.

## Manager Invite Token Error Semantics (MVP)

Invalid, expired, revoked, and already-accepted manager invitation links return distinct stable error codes with generic human-readable messages.

## Manager Invite Acceptance Idempotency (MVP)

Invitation acceptance is single-use; after one successful acceptance, later attempts with the same invitation link return an already-accepted result and do not create additional users or memberships.

## Email Delivery Strategy (MVP)

Manager invites and password reset messages are sent through a single provider abstraction, using an SMTP sandbox in development and one transactional email provider in production.

## Branch Scope (MVP)

Records remain branch-scoped in the data model with one default branch used in the pilot.

## Core Record Deletion Policy (MVP)

Core child, attendance, and invoice records are not hard-deleted; corrections, voiding, or archival flows are used instead.

## Child and Guardian Management Lifecycle (MVP)

Manager child/guardian management supports create and update plus lifecycle transitions (child inactive/left, guardian deactivated, guardian-child link ended/relinked) rather than hard delete operations.

## Guardian Link Lifecycle (MVP)

Guardian records and guardian-child links are deactivated or ended rather than hard-deleted so history remains explainable while access can be removed immediately.

## Guardian Deactivation Cascade (MVP)

Deactivating a guardian ends that guardian's active guardian-child links and active parent-membership mapping in the same action so parent access is revoked immediately with no partially active relationship state.

## Guardian Deactivation Idempotency (MVP)

Deactivating a guardian is idempotent; repeating the same deactivation request for an already inactive guardian returns success without introducing additional state changes.

## Guardian Lifecycle Timestamp Semantics (MVP)

Guardian entity lifecycle uses deactivate/reactivate terminology with deactivation-specific timestamps, while "end" timestamps are reserved for relationship records such as guardian-child links and parent-membership mappings.

## Guardian Deactivation Reason Requirement (MVP)

Manager-initiated guardian deactivation requires a stable reason code with an optional note.

## Deactivation Cascade Reason Attribution (MVP)

When guardian deactivation cascades to end active guardian-child links or parent-membership mapping, dependent records persist an explicit cascade end reason so automatic effects are distinguishable from direct manager-initiated end actions.

## Guardian Reactivation Policy (MVP)

Managers may reactivate the same guardian record when deactivation was mistaken, but previously ended guardian-child links and parent-membership mapping are not auto-restored and must be re-linked explicitly.

## Relationship End Reason Requirement (MVP)

Manager-initiated actions that end guardian-child links or parent-membership mapping require an explicit reason for audit explainability.

## Relationship End Reason Shape (MVP)

Relationship end reasons use a stable machine-readable reason code with an optional human note so audit trails stay both queryable and explainable.

## Lifecycle Reason Vocabulary (MVP)

Lifecycle transitions across child, guardian, guardian-child link, and parent-membership mapping use a shared controlled vocabulary of reason codes with scoped subsets per action.

## Lifecycle Reason Starter Set (MVP)

The initial shared reason-code set for month 1 is `duplicate_record`, `entered_in_error`, `left_nursery`, `safeguarding_direction`, `contact_update`, `access_revoked`, and `other`, with scoped usage per lifecycle action.

## Lifecycle Other-Reason Note Requirement (MVP)

When lifecycle `reason_code` is `other`, a non-empty `reason_note` is required so the audit trail remains explainable.

## Relationship End Terminology (MVP)

Lifecycle actions that stop active links or mappings use the canonical verb "end"; terms implying hard deletion (such as delete/remove) are not used for these actions.

## Child Guardian Link Requirement Enforcement (MVP)

Ending a child's last active guardian-child link is allowed, but that child immediately becomes enrollment-incomplete and is blocked from attendance and invoicing until an active guardian link is restored.

## Guardian Link Reactivation (MVP)

Only one active guardian-child link may exist per pair at a time, while historical ended links are retained so the same pair can be linked again later.

## Guardian Link Idempotent Create (MVP)

Creating a guardian-child link for a pair that already has an active link is treated as an idempotent success and does not create a duplicate active link.

## Invoice Due Policy (MVP)

Invoices are due on receipt when issued.

## Unpaid Issued Invoice (MVP)

An issued invoice with an outstanding positive balance. A zero-total issued invoice is not considered unpaid for overdue transition purposes.

## Invoice Overdue Transition (MVP)

An unpaid issued invoice transitions to `overdue` at 00:00 the next local day in `Europe/London`.

## Payment Failed Overdue Boundary (MVP)

A `payment_failed` invoice remains payable but does not transition to `overdue`.

## Invoice Issue Exception Handling (MVP)

Invoice issue runs can proceed for eligible draft invoices while invoices that cannot be issued are blocked and returned in an exception list for manager resolution.

## Draft Invoice Generation Exception Handling (MVP)

Draft invoice generation can proceed for eligible child-months while blocked child-months are skipped and returned as exceptions for manager resolution.

## Draft Generation Transaction Boundary (MVP)

Expected child-month blockers produce generation exceptions, while unexpected system failures leave no partial invoice generation result behind.

## Invoice Issue Transaction Boundary (MVP)

Expected invoice issue blockers produce issue exceptions, while unexpected system failures leave no partial invoice issue result behind.

## Empty Draft Generation Run (MVP)

A draft invoice generation run can complete without creating invoices when no child-months are eligible; this is a valid billing outcome rather than a request failure.

## Empty Invoice Issue Run (MVP)

An invoice issue run can complete without issuing invoices when no draft invoices are eligible; this is a valid billing outcome rather than a request failure.

## Selected Draft Generation Child Exception (MVP)

When a selected child-month cannot be found in the active billing scope or does not overlap the billing month, draft generation treats that child-month as an exception rather than failing the whole run.

## Selected Draft Generation Uniqueness (MVP)

Selected-child draft generation operates on unique child-months; duplicate selected child identifiers do not create duplicate invoice work.

## Selected Invoice Issue Uniqueness (MVP)

Selected-invoice issue operates on unique invoices; duplicate selected invoice identifiers do not create duplicate issue work.

## Invoice Run Blocked Child (MVP)

A child excluded from a specific monthly invoice run because billing readiness checks found a resolvable issue, such as incomplete attendance. This is a run/month-specific billing state, not a child lifecycle state.

## Invoice Exception Reference (MVP)

A manager-visible reference from invoice review to child-month blockers recorded on the invoice run that generated or processed related invoices. Exception references provide run context and do not represent invoice lines, adjustment invoices, or invoices for blocked child-months.

## Invoice Preflight Blocker (MVP)

A child-month readiness exception returned by invoice draft preflight. A blocked child may have multiple preflight blockers, and each blocker should be exposed with a stable code.

## Invoice Preflight Enrollment Blocker Codes (MVP)

Invoice draft preflight explains enrollment incompleteness with granular stable blocker codes rather than only a generic enrollment-incomplete label.

## Draft Invoice Regeneration (MVP)

After attendance corrections, managers can regenerate draft invoices for individual children without rerunning the full month batch.

## Draft Invoice Generation Scope (MVP)

Draft invoice generation can target all eligible child-months in a billing month or a manager-selected subset of child-months.

## Draft Invoice Generation Outcome (MVP)

Draft invoice generation reports the invoice run outcome and affected invoice references; detailed invoice review remains a separate manager billing view.

## Manager Invoice Review (MVP)

A manager-facing billing view for inspecting generated invoice headers, line items, calculation quantities, status, and due/payment metadata across invoice statuses. Draft invoices are included so managers can review calculations before issue.

## Parent Invoice View (MVP)

A parent-facing billing view for issued-or-later invoices belonging to children reachable through that parent's current membership-to-guardian access path. It explains payable invoice identity, child, period, status, totals, payment state, and parent-readable line items.

## Parent Invoice Detail Disclosure (MVP)

Parent invoice detail supports understanding and paying an issued invoice, not reviewing manager billing operations. Manager-only run context, invoice exceptions, lock metadata, adjustment internals, and attendance source-session snapshots are outside the parent disclosure boundary.

## Issued-or-Later Invoice (MVP)

An invoice whose lifecycle has moved beyond draft into `issued`, `payment_failed`, `paid`, or `overdue`. Parents can see issued-or-later invoices for linked children, while draft invoices remain manager-only.

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

## Guardian Link Scope Consistency (MVP)

Guardian-child links are valid only when guardian and child belong to the same tenant and branch scope.

## Tenancy Isolation Enforcement (MVP)

Tenant and branch isolation is enforced in application-layer authorization and query scoping in month 1; database RLS is deferred to post-pilot hardening.

## Test Priority (MVP)

Automated test coverage is prioritized for funding calculation logic, invoice state transitions, authorization scope checks, and Stripe webhook idempotency, with minimal UI end-to-end coverage on core happy paths.

## Authorization Verification Baseline (MVP)

Authorization acceptance requires a route-by-route role and scope test matrix that proves unauthenticated requests are rejected, wrong-role/wrong-tenant/wrong-branch requests are forbidden, and correctly scoped allowed-role requests succeed.

## Observability Baseline (MVP)

Operations monitoring uses structured logs plus essential metrics for webhook outcomes, invoice-generation job health, and authentication failures.

## Authorization Denial Logging (MVP)

Authorization denials are captured in structured logs with request identifier, actor, scope context, and denial reason code.

## Authorization Denial Persistence Scope (MVP)

Authorization denials are treated as operational telemetry and are recorded in structured logs and metrics rather than persisted as audit-log domain events in month 1.

## Authorization Denial Metrics (MVP)

Authorization denials emit metrics tagged by stable denial reason code to support operational monitoring.

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

## Authorization Guards (MVP)

Authorization guards are the combined enforcement layer for role, scope, and relationship checks; RBAC is one component of this guard model.

## Authorization Check Taxonomy (MVP)

Scope checks validate tenant and branch boundaries from the active session membership, while relationship checks validate dynamic record linkage such as parent-child access.

## Authorization Evaluation Rule (MVP)

Session-bound scope and role claims are used for baseline access checks, while dynamic relationship checks such as parent-child linkage are validated against current records.

## Parent Relationship Check Freshness (MVP)

Parent access to child-linked resources is authorized against current guardian-child links at request time.

## Authorization Route Policy (MVP)

Protected endpoints follow default-deny behavior, with explicit role and scope requirements declared per route.

## Route Role Declaration Style (MVP)

Month-1 protected endpoints declare explicit allowed role lists per route; capability abstraction is deferred.

## Unknown Role Handling (MVP)

If a session presents an unknown or unsupported role claim, authorization fails closed with a forbidden response and a stable role-specific error code.

## Authorization Denial Precedence (MVP)

When multiple authorization checks fail, guards return one deterministic primary denial code based on a fixed evaluation order.

## Authorization Check Order (MVP)

Authorization evaluation order is authentication validity first, then role checks, then tenant/branch scope checks, then relationship checks.

## Authorization Layer Boundary (MVP)

Middleware enforces token validity and coarse route-level role and scope checks, while domain logic enforces resource-specific relationship authorization.

## Authorization Context Propagation (MVP)

Authorization middleware constructs a normalized request authorization context (user, membership, role, tenant, branch, request identifier) for downstream handlers and services.

## Scope Source of Truth (MVP)

Effective tenant and branch scope is derived from the active session membership, and client-supplied scope fields are rejected when they conflict.

## Session Scope Isolation Rule (MVP)

Operations cannot cross into another branch or tenant within the same session; users must switch membership scope before acting in a different scope.

## Attendance Edit Authority (MVP)

Historical attendance events are not edited directly by practitioners; corrections are captured as manager-only correction events.

## Guardian Link End Visibility Rule (MVP)

When a guardian-child link is ended, that parent immediately loses access to that child's invoices, including historical invoices.

## Guardian Relink Visibility Rule (MVP)

When a guardian-child relationship is re-linked with an active guardian-child link, parent visibility for that child's invoices is restored based on the current active-link relationship check.

## Draft Invoice Idempotency (MVP)

Regenerating draft invoices for the same month does not create duplicates; existing eligible drafts are replaced or updated unless already issued.

## Draft Invoice Regeneration Identity (MVP)

Regenerating an existing draft monthly invoice preserves the draft invoice identity while replacing its draft calculation contents.

## Draft Invoice Generation Audit Scope (MVP)

Draft invoice generation records audit history for each draft invoice that is created or recalculated; blocked child-months are represented as invoice run exceptions because no invoice changes for those child-months.

## Invoice Issue Audit Scope (MVP)

Invoice issue records audit history for each invoice that becomes issued; blocked invoices are represented as invoice run exceptions because no invoice changes for those invoices.

## Invoice Run History (MVP)

Repeated draft generation requests create separate invoice run history entries even when they update the same child-month draft invoice.

## Issued Invoice Regeneration Policy (MVP)

Issued invoices are not regenerated after attendance changes; post-issue changes require an explicit adjustment flow.

## Draft Regeneration Issue Race Rule (MVP)

If a draft monthly invoice becomes issued while draft regeneration is being prepared, issued status takes precedence and the invoice is skipped rather than recalculated.

## Stale Draft Blocked Regeneration (MVP)

When an existing draft monthly invoice is no longer eligible for regeneration, the draft is left in place and the child-month is returned as blocked rather than silently removed.

## Adjustment Flow (MVP)

Post-issue billing changes are handled through a manager-created follow-up adjustment invoice linked to the original issued invoice, with a required reason.

## API Versioning (MVP)

HTTP endpoints are published under a versioned `/api/v1` route prefix.

## Public Route Allowlist (MVP)

Only health and authentication endpoints are public in month 1; all other API routes require authorization guards.

## Invite Acceptance Route Visibility (MVP)

Invitation acceptance is a public provisioning route because invitees do not have a session before acceptance; all manager invitation management routes remain manager-protected.

## Deployment Model (MVP)

Production runs on a single virtual machine using Docker Compose for service orchestration.

## Entity Identifier Strategy (MVP)

Domain entities use UUID primary keys (UUIDv7 preferred, UUIDv4 acceptable).

## Invoice Explainability Persistence (MVP)

Invoice line storage preserves both intermediate billing components (core attended minutes, funded deduction minutes, core billable minutes, hourly rate) and final totals.

## Invoice Attendance Source Snapshot (MVP)

A generated draft invoice preserves a compact snapshot of the attendance sessions used to calculate it so managers can explain how the billed minutes were derived.

## Draft Invoice Calculation Lines (MVP)

Generated draft monthly invoices use consistent explanatory lines for core childcare and funded-hours deduction, including a zero-value funded deduction line when no deduction amount is applied.

## Parent Draft Invoice Visibility (MVP)

Draft invoices are visible only to managers; parents can view invoices only after they are issued.

## Mid-Month Leave Billing (MVP)

If a child leaves during a month, billing is automatically derived from attendance actuals up to the leave date.

## Post-Leave Invoice Artifact Policy (MVP)

Marking a child inactive/left does not trigger automatic voiding of unrelated invoice artifacts; future billing is naturally blocked by attendance-derived invoice generation.
