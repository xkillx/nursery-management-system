# API Schema State

> **Verified as of 2026-06-11.** Latest migration: 000017.

- **Last verification date**: 2026-06-11
- **Verified migration version**: 17
- **Latest migration**: 000017 (`add_child_registration_office_checklists`)
- **Workflow**: `make migrate-verify` (up → version → down -all → up → version)
- **Migration tool**: golang-migrate (manual, not auto-run at API startup)

## Application Tables (24)

`schema_migrations` is golang-migrate metadata, not an application table.

### Tenant & Branch

| Table | Key columns | Notes |
|---|---|---|
| `tenants` | `id UUID PK`, `name TEXT` | Top-level multi-tenant entity |
| `branches` | `id UUID PK`, `tenant_id FK`, `name TEXT` | Unique `(tenant_id, name)`, composite unique `(tenant_id, id)` |

### Users & Authentication

| Table | Key columns | Notes |
|---|---|---|
| `users` | `id UUID PK`, `email`, `email_normalized UNIQUE`, `password_hash`, `is_active` | Global (not tenant-scoped) |
| `memberships` | `id UUID PK`, `tenant_id FK`, `branch_id FK`, `user_id FK`, `role TEXT`, `is_active`, `ended_at` | Unique `(tenant_id, branch_id, user_id)`. Role enum: `manager`, `practitioner`, `parent`. Active/inactive consistency check. Composite unique `(tenant_id, branch_id, id)`. |
| `refresh_tokens` | `id UUID PK`, `user_id FK`, `membership_id FK NOT NULL`, `token_hash UNIQUE`, `expires_at`, `revoked_at` | Bound to membership. |
| `password_reset_tokens` | `id UUID PK`, `user_id FK`, `token_hash UNIQUE`, `expires_at`, `used_at`, `superseded_at` | CHECK: `used_at IS NULL OR superseded_at IS NULL`. Partial index on active tokens per user. |

### Children & Guardians

| Table | Key columns | Notes |
|---|---|---|
| `children` | `id UUID PK`, `tenant_id`, `branch_id`, `full_name`, `date_of_birth`, `start_date`, `end_date`, `core_hourly_rate_minor`, `is_active`, `left_at`, `left_reason_code`, `left_reason_note` | Composite unique `(tenant_id, branch_id, id)`. Active/inactive consistency check. `core_hourly_rate_minor >= 0`. `start_date <= end_date`. |
| `guardians` | `id UUID PK`, `tenant_id`, `branch_id`, `full_name`, `relationship`, `phone`, `email`, `is_active`, `deactivated_at`, `deactivation_reason_code`, `deactivation_reason_note` | Composite unique `(tenant_id, branch_id, id)`. Active/inactive consistency check. |
| `guardian_child_links` | `id UUID PK`, `tenant_id`, `branch_id`, `guardian_id FK`, `child_id FK`, `ended_at`, `ended_reason_code`, `ended_reason_note` | Partial unique index on active `(guardian_id, child_id)`. End-reason consistency check. |
| `parent_membership_guardians` | `id UUID PK`, `tenant_id`, `branch_id`, `membership_id FK`, `guardian_id FK`, `ended_at`, `ended_reason_code`, `ended_reason_note` | Partial unique: one active mapping per membership, one active `(membership_id, guardian_id)` pair. End-reason consistency check. Triggers enforce parent role and active entities. |
| `child_registration_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, 60+ profile columns (see migration), collection password hash/metadata, 8 section review flags, `created_at`, `updated_at` | One profile per child `(tenant_id, branch_id, child_id)` UNIQUE. JSONB checks for `home_address` (object), `professional_referrals` (array). Password consistency check. |
| `child_registration_contacts` | `id UUID PK`, `tenant_id`, `branch_id`, `profile_id FK CASCADE`, `child_id FK`, `contact_type`, `sort_order`, `full_name`, relationship/contact fields, JSONB `address`/`work_address`, `has_parental_responsibility`, `created_at`, `updated_at` | Unique per `(profile_id, contact_type, sort_order)`. JSONB object checks. Sort order >= 0. FK cascades on profile delete. |
| `child_registration_office_checklists` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `deposit_status`, `deposit_paid_date`, `application_date_status`, `application_date`, `start_date_status`, `date_left`, `sessions_days_requested_status`, `sessions_days_requested`, `term_time_only_space_status`, `contract_status`, `contract_date`, `handbook_status`, `handbook_date`, `red_book_status`, `red_book_checked_date`, `birth_certificate_passport_status`, `birth_certificate_passport_checked_date`, `proof_of_address_status`, `proof_of_address_checked_date`, `notes`, `created_at`, `updated_at` | One per child `(tenant_id, branch_id, child_id)` UNIQUE. CHECK: `application_date` required when status=complete; `sessions_days_requested` required when status=complete. New enums: `registration_office_check_status`, `registration_term_time_only_status`. |

### Attendance

| Table | Key columns | Notes |
|---|---|---|
| `attendance_sessions` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `status`, `check_in_at`, `check_out_at`, `check_in_local_date`, `check_out_local_date`, `check_in_event_id FK`, `check_out_event_id FK`, `corrected_by_event_id FK` | Status: `open`, `complete`, `corrected`. Composite unique `(tenant_id, branch_id, id)`. Partial unique index ensures one open session per child per branch. Shape check: `open` → no checkout columns; `complete`/`corrected` → checkout required. `check_out_at > check_in_at`. |
| `attendance_events` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `session_id FK`, `event_type`, `occurred_at`, `local_date`, `recorded_by_user_id FK`, `recorded_by_membership_id FK`, `request_id`, `reason_code`, `reason_note`, `details JSONB` | Event type: `check_in`, `check_out`, `correction`. Composite unique `(tenant_id, branch_id, id)`. Routine events (check-in/check-out) cannot carry reason_code or reason_note. Correction events require a reason code (`missed_check_in`, `missed_check_out`, `incorrect_time`, `duplicate_entry`, `other`). `other` requires non-empty note. |

### Funding

| Table | Key columns | Notes |
|---|---|---|
| `funding_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `billing_month DATE`, `funded_allowance_minutes INTEGER`, `created_at`, `updated_at` | Unique `(tenant_id, branch_id, child_id, billing_month)`. `billing_month` must be first day of month. Allowance bounds: 0–44640. FK to `branches(tenant_id, id)` and `children(tenant_id, branch_id, id)`. |

### Invoicing

| Table | Key columns | Notes |
|---|---|---|
| `invoice_runs` | `id UUID PK`, `tenant_id`, `branch_id`, `billing_month DATE`, `run_type TEXT`, `status TEXT`, `started_at`, `completed_at`, `requested_by_user_id FK`, `requested_by_membership_id`, `request_id`, `eligible/success/blocked/failed_count`, `details JSONB` | Composite unique `(tenant_id, branch_id, id)`. `run_type IN ('draft_generation', 'issue')`. `status IN ('started', 'completed', 'completed_with_exceptions', 'failed')`. `completed_at IS NULL` only while `started`. |
| `invoices` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `billing_month DATE`, `invoice_kind TEXT`, `status TEXT`, `invoice_number TEXT`, `issued_sequence INTEGER`, `generated_run_id FK`, `issued_run_id FK`, `issued_at`, `issued_by_user_id FK`, `issued_by_membership_id`, `locked_at`, `due_at`, `currency_code CHAR(3)`, `subtotal_minor`, `funded_deduction_minor`, `total_due_minor`, `amount_paid_minor`, `paid_at`, `payment_failed_at`, `payment_status_updated_at`, `adjusts_invoice_id FK (self)`, `adjustment_reason_code`, `adjustment_reason_note`, `period_start_date`, `period_end_date`, `calculation_details JSONB` | Composite unique `(tenant_id, branch_id, id)`. Partial unique monthly `(tenant_id, branch_id, child_id, billing_month) WHERE invoice_kind = 'monthly'`. `invoice_kind IN ('monthly', 'adjustment')`. `status IN ('draft', 'issued', 'payment_failed', 'paid', 'overdue')`. Draft shape: issue fields null. Issued shape: issue fields non-null. Paid shape: `paid_at IS NOT NULL`, `amount_paid_minor = total_due_minor`. Adjustment shape: requires `adjusts_invoice_id` + non-empty reason. Monthly shape: adjustment fields null. `funded_deduction_minor` stored as positive reporting amount. |
| `invoice_lines` | `id UUID PK`, `tenant_id`, `branch_id`, `invoice_id FK`, `line_kind TEXT`, `description TEXT`, `sort_order INTEGER`, `quantity_minutes INTEGER`, `unit_amount_minor INTEGER`, `line_amount_minor INTEGER`, `raw_attended_minutes`, `rounded_attended_minutes`, `funded_allowance_minutes`, `funded_deduction_minutes`, `core_billable_minutes`, `session_count`, `details JSONB` | Composite unique `(tenant_id, branch_id, id)`. `line_kind IN ('core_childcare', 'funded_deduction', 'extra', 'adjustment')`. `core_childcare` and `extra`: `line_amount_minor >= 0`. `funded_deduction`: `line_amount_minor <= 0`. `adjustment`: signed either direction. |
| `invoice_number_sequences` | `tenant_id`, `branch_id`, `billing_year INTEGER`, `billing_month INTEGER`, `next_sequence INTEGER DEFAULT 1` | PK `(tenant_id, branch_id, billing_year, billing_month)`. `billing_year >= 2000`, `billing_month 1–12`, `next_sequence >= 1`. |

### Payments

| Table | Key columns | Notes |
|---|---|---|
| `payment_attempts` | `id UUID PK`, `tenant_id`, `branch_id`, `invoice_id FK`, `initiated_by_user_id FK`, `initiated_by_membership_id`, `request_id TEXT`, `status TEXT`, `amount_minor INTEGER`, `currency_code CHAR(3) DEFAULT 'GBP'`, `stripe_checkout_session_id TEXT`, `stripe_checkout_url TEXT`, `stripe_payment_intent_id TEXT`, `stripe_expires_at TIMESTAMPTZ`, `provider_error_code TEXT`, `provider_error_message TEXT`, `failure_reason TEXT`, `created_at`, `updated_at` | Composite unique `(tenant_id, branch_id, id)`. Composite FKs to branches, invoices, memberships. `status IN ('checkout_creation_started', 'checkout_created', 'checkout_creation_failed', 'paid', 'payment_failed', 'cancelled', 'expired')`. `amount_minor > 0`. `currency_code = 'GBP'`. CHECK: `checkout_created` requires non-null session ID and URL. Partial unique index on `stripe_checkout_session_id` where not null. Index on `(tenant_id, branch_id, invoice_id, created_at DESC)`. Partial index on open attempts. |

#### stripe_webhook_events

Raw inbox of verified Stripe webhook events.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Internal ID |
| stripe_event_id | TEXT UNIQUE | Stripe event ID |
| event_type | TEXT | Stripe event type |
| livemode | BOOLEAN | Live or test mode |
| api_version | TEXT | Stripe API version |
| provider_created_at | TIMESTAMPTZ | When Stripe created the event |
| received_at | TIMESTAMPTZ | When we received it |
| processed_at | TIMESTAMPTZ | When processing completed |
| processing_status | TEXT | `processed`, `ignored`, or `rejected` |
| processing_reason | TEXT | Machine-readable reason |
| request_id | TEXT | HTTP request ID |
| raw_payload | JSONB | Full verified raw payload |
| error_message | TEXT | Error if any |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

#### payment_reconciliation_records

Manager-facing payment timeline records.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Internal ID |
| tenant_id | UUID FK | Tenant scope |
| branch_id | UUID FK | Branch scope |
| invoice_id | UUID FK | Invoice |
| payment_attempt_id | UUID FK | Payment attempt |
| stripe_webhook_event_id | UUID FK | Triggering webhook event |
| stripe_event_id | TEXT | Stripe event ID |
| stripe_event_type | TEXT | Stripe event type |
| stripe_checkout_session_id | TEXT | Checkout session |
| stripe_payment_intent_id | TEXT | Payment intent |
| outcome | TEXT | `paid`, `payment_failed`, `expired`, `ignored`, `rejected` |
| reason_code | TEXT | Machine-readable reason |
| previous_invoice_status | TEXT | Invoice status before |
| new_invoice_status | TEXT | Invoice status after |
| attempt_previous_status | TEXT | Attempt status before |
| attempt_new_status | TEXT | Attempt status after |
| amount_minor | INTEGER | Amount in minor units |
| currency_code | CHAR(3) | Currency |
| details | JSONB | Additional details |
| created_at | TIMESTAMPTZ | |

### Audit

| Table | Key columns | Notes |
|---|---|---|
| `audit_logs` | `id UUID PK`, `tenant_id FK`, `branch_id FK`, `action_type`, `action_entity_type`, `action_entity_id`, `actor_user_id FK`, `actor_membership_id FK`, `details JSONB`, `request_id`, `reason_code`, `reason_note`, `created_at` | `reason_code` requires `reason_note` be NULL. `reason_code = other` requires non-empty note. |

## Enum / Custom Type

| Type | Values |
|---|---|
| `lifecycle_reason_code` | `duplicate_record`, `entered_in_error`, `left_nursery`, `safeguarding_direction`, `contact_update`, `access_revoked`, `other` |
| `registration_yes_no_unknown` | `unknown`, `no`, `yes` |
| `registration_immunisation_status` | `unknown`, `up_to_date`, `refused`, `partial`, `not_recorded` |
| `registration_contact_type` | `parent_carer`, `emergency_contact`, `authorised_collector` |

Used by: `children.left_reason_code`, `guardians.deactivation_reason_code`, `guardian_child_links.ended_reason_code`, `parent_membership_guardians.ended_reason_code`, `audit_logs.reason_code`.
Registration enums used by: `child_registration_profiles` and `child_registration_contacts` tables (migration 000016).

## Triggers & Functions

| Trigger | Table | Function | Behavior |
|---|---|---|---|
| `parent_membership_guardians_role_check` | `parent_membership_guardians` (BEFORE INSERT OR UPDATE) | `enforce_parent_membership_guardian_role()` | Membership must have `parent` role. |
| `parent_membership_guardians_active_entity_check` | `parent_membership_guardians` (BEFORE INSERT OR UPDATE) | `enforce_parent_mapping_active_entities()` | Membership must be active, guardian must be active. |
| `memberships_role_guardian_mapping_check` | `memberships` (BEFORE UPDATE OF role) | `prevent_non_parent_with_active_guardian_mapping()` | Cannot change role away from `parent` while active guardian mapping exists. |
| `trg_invoice_status_transition` | `invoices` (BEFORE UPDATE OF status) | `enforce_invoice_status_transition()` | Legal transitions: `draft→issued`, `issued→overdue/paid/payment_failed`, `overdue→paid/payment_failed`, `payment_failed→paid`. `paid` is terminal. No return to `draft`. `payment_failed→overdue` blocked. |
| `trg_invoice_immutability` | `invoices` (BEFORE UPDATE) | `protect_issued_invoice_immutability()` | Non-draft invoices can only change `status`, `amount_paid_minor`, `paid_at`, `payment_failed_at`, `payment_status_updated_at`, `updated_at`. |
| `trg_invoice_lines_immutability` | `invoice_lines` (BEFORE INSERT OR UPDATE OR DELETE) | `protect_issued_invoice_lines()` | Rejects line changes when parent invoice status is not `draft`. |

## Key Indexes

| Index | Table | Type | Notes |
|---|---|---|---|
| `branches_tenant_id_id_unique` | `branches` | UNIQUE btree | Scope composite key |
| `branches_tenant_id_name_key` | `branches` | UNIQUE btree | Branch name per tenant |
| `users_email_normalized_key` | `users` | UNIQUE btree | Case-insensitive email |
| `memberships_scope_id_unique` | `memberships` | UNIQUE btree | Scope composite key |
| `memberships_tenant_id_branch_id_user_id_key` | `memberships` | UNIQUE btree | One membership per user per branch |
| `children_scope_id_unique` | `children` | UNIQUE btree | Scope composite key |
| `guardians_scope_id_unique` | `guardians` | UNIQUE btree | Scope composite key |
| `idx_guardian_child_links_active_pair` | `guardian_child_links` | UNIQUE btree (partial) | One active link per guardian-child pair |
| `idx_parent_membership_guardians_active_membership` | `parent_membership_guardians` | UNIQUE btree (partial) | One active mapping per membership |
| `idx_parent_membership_guardians_active_pair` | `parent_membership_guardians` | UNIQUE btree (partial) | One active mapping per membership-guardian pair |
| `idx_attendance_sessions_one_open_child` | `attendance_sessions` | UNIQUE btree (partial) | One open session per child per branch scope |
| `attendance_events_scope_id_unique` | `attendance_events` | UNIQUE btree | Scope composite key |
| `attendance_sessions_scope_id_unique` | `attendance_sessions` | UNIQUE btree | Scope composite key |
| `idx_password_reset_tokens_user_id` | `password_reset_tokens` | btree | Lookup by user |
| `idx_password_reset_tokens_expires_at` | `password_reset_tokens` | btree | Expiry scan |
| `idx_password_reset_tokens_active_user` | `password_reset_tokens` | btree (partial) | Active tokens per user by creation desc |
| `funding_profiles_scope_child_month_unique` | `funding_profiles` | UNIQUE btree | One profile per child per billing month in scope |
| `idx_funding_profiles_scope_month` | `funding_profiles` | btree | Lookup by tenant/branch/month |
| `idx_funding_profiles_child_month` | `funding_profiles` | btree | Lookup by tenant/branch/child/month |
| `idx_invoice_runs_scope_id` | `invoice_runs` | UNIQUE btree | Scope composite key |
| `idx_invoice_runs_billing_scope` | `invoice_runs` | btree | Lookup by tenant/branch/month/type/time |
| `idx_invoice_runs_request_id` | `invoice_runs` | btree (partial) | Request ID lookup where non-null |
| `idx_invoices_scope_id` | `invoices` | UNIQUE btree | Scope composite key |
| `idx_invoices_monthly_unique` | `invoices` | UNIQUE btree (partial) | One monthly invoice per child/month where `invoice_kind = 'monthly'` |
| `idx_invoices_invoice_number_unique` | `invoices` | UNIQUE btree (partial) | Invoice number per tenant/branch where non-null |
| `idx_invoices_billing_status` | `invoices` | btree | Lookup by tenant/branch/month/status |
| `idx_invoices_child_billing` | `invoices` | btree | Lookup by tenant/branch/child/month |
| `idx_invoices_adjusts` | `invoices` | btree (partial) | Adjustment lookup where non-null |
| `idx_invoices_due_at_outstanding` | `invoices` | btree (partial) | Outstanding invoices by due date |
| `idx_invoice_lines_scope_id` | `invoice_lines` | UNIQUE btree | Scope composite key |
| `idx_invoice_lines_invoice_order` | `invoice_lines` | btree | Lines by invoice + sort order |
| `uq_payment_attempts_scoped_id` | `payment_attempts` | UNIQUE btree | Scope composite key |
| `idx_payment_attempts_invoice_created` | `payment_attempts` | btree | Attempts by invoice + created desc |
| `uq_payment_attempts_stripe_session_id` | `payment_attempts` | UNIQUE btree (partial) | Stripe session ID uniqueness where non-null |
| `idx_payment_attempts_open_attempts` | `payment_attempts` | btree (partial) | Open/created attempts per invoice |
| `uq_reconciliation_stripe_event_id` | `payment_reconciliation_records` | UNIQUE btree | Stripe event ID uniqueness on reconciliation |
| `idx_reconciliation_invoice_created` | `payment_reconciliation_records` | btree | Reconciliation records by invoice + created desc |
| `idx_reconciliation_attempt_created` | `payment_reconciliation_records` | btree | Reconciliation records by attempt + created desc |
| `idx_stripe_webhook_events_event_type` | `stripe_webhook_events` | btree | Webhook events by type + received desc |
| `idx_stripe_webhook_events_processing_status` | `stripe_webhook_events` | btree | Webhook events by processing status + received desc |
| `idx_child_registration_profiles_scope_child` | `child_registration_profiles` | UNIQUE btree | One profile per child per scope |
| `child_registration_profiles_scope_id_unique` | `child_registration_profiles` | UNIQUE btree | Scope composite key for contact FK |
| `idx_child_registration_contacts_profile_type` | `child_registration_contacts` | btree | Contacts by profile + type + sort order |
