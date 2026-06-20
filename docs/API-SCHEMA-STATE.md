# API Schema State

> **Verified as of 2026-06-21.** Latest migration: 000005 (session templates).

- **Last verification date**: 2026-06-21
- **Verified migration version**: 5
- **Latest migration**: 000005 (`session_templates`) — adds `session_templates` and `session_template_entries` for named, per-site, reusable weeks of booked sessions. Captured at registration; no backfill of existing children (see ADR-0009).
- **Workflow**: `make migrate-verify` (up → version → down -all → up → version)
- **Migration tool**: golang-migrate (manual, not auto-run at API startup)

## Application Tables (35)

`schema_migrations` is golang-migrate metadata, not an application table.

### Tenant & Branch

| Table | Key columns | Notes |
|---|---|---|
| `tenants` | `id UUID PK`, `name TEXT` | Top-level multi-tenant entity |
| `branches` | `id UUID PK`, `tenant_id FK`, `name TEXT`, `core_hourly_rate_minor INTEGER` | Unique `(tenant_id, name)`, composite unique `(tenant_id, id)`. `core_hourly_rate_minor` nullable; must be positive if set (CHECK). Authoritative site rate. |

### Users & Authentication

| Table | Key columns | Notes |
|---|---|---|
| `users` | `id UUID PK`, `email`, `email_normalized UNIQUE`, `password_hash`, `is_active` | Global (not tenant-scoped) |
| `memberships` | `id UUID PK`, `tenant_id FK`, `branch_id FK`, `user_id FK`, `role TEXT`, `is_active`, `ended_at` | Unique `(tenant_id, branch_id, user_id)`. Role enum: `manager`, `practitioner`, `parent`. Active/inactive consistency check. Composite unique `(tenant_id, branch_id, id)`. |
| `refresh_tokens` | `id UUID PK`, `user_id FK`, `membership_id FK NOT NULL`, `token_hash UNIQUE`, `expires_at`, `revoked_at` | Bound to membership. |
| `password_reset_tokens` | `id UUID PK`, `user_id FK`, `token_hash UNIQUE`, `expires_at`, `used_at`, `superseded_at` | CHECK: `used_at IS NULL OR superseded_at IS NULL`. Partial index on active tokens per user. |

### Children & Sub-Records

The child bounded context is split into a small `children` identity table and 1:1 sub-record tables, plus a 1:many `child_contacts` and `child_room_assignments` history. This replaces the legacy `child_registration_*` four-table model and the operational columns formerly on `children`.

| Table | Key columns | Notes |
|---|---|---|
| `children` | `id UUID PK`, `tenant_id`, `branch_id`, `first_name`, nullable `middle_name`, nullable `last_name`, `date_of_birth`, `start_date`, `end_date`, `is_active` | Composite unique `(tenant_id, branch_id, id)`. Active/inactive consistency check. `first_name` is required and not blank. `middle_name` and `last_name` are nullable. `start_date <= end_date`. Room placement and billing rate live in dedicated sub-record tables; the child row itself is identity only. |
| `child_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, demographics (sex, religion, ethnic_origin, first_language, other_languages), `home_address JSONB`, `home_postcode`, `home_telephone`, disability / access, routine care, GDPR declaration metadata, `registration_date`, 7 section review booleans | 1:1 with `children`. JSONB `home_address` object check. Per-child demographic / consent-declaration row. |
| `child_contacts` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `contact_type child_contact_type`, `sort_order`, `full_name`, relationship, JSONB `address`/`work_address`, telephone/email, `has_parental_responsibility` | 1:many per child. `contact_type IN ('parent_carer','emergency_contact','authorised_collector')`. UNIQUE `(tenant_id, branch_id, child_id, contact_type, sort_order)`. JSONB object checks. `sort_order >= 0`. |
| `child_health_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, medical/allergy/dietary status fields, immunisation status/country, doctor + health visitor contact | 1:1 with `children`. `medical_conditions_status`, `prescribed_medication_status`, `dietary_requirements_status` ∈ `('unknown','no','yes')`. `immunisation_status` ∈ `('unknown','up_to_date','refused','partial','not_recorded')`. |
| `child_safeguarding_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, social services status/notes + social worker contact, six concern flags (walking, speech_language, hearing, sight, emotional_wellbeing, behaviour), `professional_referrals JSONB`, `restricted_notes` | 1:1 with `children`. All six concern flags ∈ `('unknown','no','yes')`. JSONB array check on `professional_referrals`. |
| `child_consent_records` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, 18 boolean consent flags (`urgent_medical_treatment`, `plasters`, `safeguarding_reporting_acknowledgement`, `information_sharing_consent`, `gdpr_data_processing_consent`, `area_senco_liaison`, `health_visitor_liaison`, `transition_documents`, `local_outings`, `face_painting`, `parent_supplied_sun_cream`, `parent_supplied_nappy_cream`, `development_profile_photos`, `nursery_display_boards`, `promotional_literature`, `nursery_website`, `staff_student_coursework`, `social_media`), `social_media_channel_notes`, `notes_exceptions`, `signer_name`, `signed_date`, `paper_form_on_file`, `entered_by_user_id`/`entered_by_membership_id` | 1:1 with `children`. Single current row per child (no version column). PUT-style update — the new write replaces the current row. |
| `child_funding_records` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, six eligibility tri-state text fields (benefits_contribute_to_fees, working_tax_credit, college_uni_paid_to_parent, college_uni_paid_to_nursery, funding_3yo_term_time, funding_2yo_term_time), `funding_support_notes`, `funding_support_reviewed` | 1:1 with `children`. Each tri-state ∈ `('unknown','no','yes')`. |
| `child_collection_settings` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, `over_18_collection_acknowledged`, `collection_password_hash`, `collection_password_updated_at`, `collection_password_updated_by_user_id`/`_membership_id` | 1:1 with `children`. CHECK: all four password-metadata columns are null together, or all non-null together. |
| `child_room_assignments` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `room_id FK rooms`, `start_date`, nullable `end_date`, generated `is_current BOOLEAN` (= `end_date IS NULL`) | 1:many per child — full room-placement history. `end_date >= start_date`. Partial indexes on the `is_current` rows power capacity / current-room queries. |
| `child_billing_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, `billing_basis` ∈ `('site_rate','custom')`, `custom_rate_minor` (nullable, positive when present), `effective_from` | 1:1 with `children`. CHECK: `billing_basis='site_rate'` ⇔ `custom_rate_minor IS NULL`; `billing_basis='custom'` ⇔ `custom_rate_minor > 0`. |
| `child_leaving_records` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK UNIQUE`, `left_at`, `reason_code` ∈ `lifecycle_reason_code`, `reason_note` | 1:1 with `children`. Created when a child is marked inactive; carries the lifecycle reason and optional note. |
| `child_booking_patterns` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `effective_from DATE`, nullable `effective_to DATE`, generated `is_current BOOLEAN` (= `effective_to IS NULL`) | 1:many per child — full pattern history. CHECK: `effective_to IS NULL OR effective_to >= effective_from`. Partial unique index on the `is_current` rows enforces one open pattern per child. Replacing an active pattern closes it adjacently (`effective_to = new.effective_from - 1 day`). |
| `child_booking_pattern_entries` | `id UUID PK`, `tenant_id`, `branch_id`, `pattern_id FK`, `day_of_week INTEGER` ∈ `1..7` (ISO Monday=1), `session_type_id FK` | 1:many per pattern. UNIQUE `(tenant_id, branch_id, pattern_id, day_of_week, session_type_id)`. FK to `session_types(tenant_id, branch_id, id)`. Multiple entries on the same day (e.g. AM + PM) are allowed. |

### Session Types and Templates

| Table | Key columns | Notes |
|---|---|---|
| `session_types` | `id UUID PK`, `tenant_id`, `branch_id`, `name`, `start_time TIME`, `end_time TIME`, `is_active` | Per-site reference data. CHECK: `start_time < end_time`. Partial unique on `(tenant_id, branch_id, name) WHERE is_active = true` enforces one active name per site. Soft-deletable via `is_active`. |
| `session_templates` | `id UUID PK`, `tenant_id`, `branch_id`, `name`, nullable `description`, `is_active` | Per-site reference data. Partial unique on `(tenant_id, branch_id, name) WHERE is_active = true`. Soft-deletable via `is_active`. |
| `session_template_entries` | `id UUID PK`, `tenant_id`, `branch_id`, `template_id FK`, `day_of_week INTEGER` ∈ `1..7`, `session_type_id FK` | 1:many per template. UNIQUE `(tenant_id, branch_id, template_id, day_of_week, session_type_id)`. FK to `session_types`. Templates **copy** entries into a booking pattern at creation time — no FK from `child_booking_pattern_entries` to `session_templates` (history integrity over live update; ADR-0009). |

### Guardians

| Table | Key columns | Notes |
|---|---|---|
| `guardians` | `id UUID PK`, `tenant_id`, `branch_id`, `full_name`, `relationship`, `phone`, `email`, `is_active`, `deactivated_at`, `deactivation_reason_code`, `deactivation_reason_note` | Composite unique `(tenant_id, branch_id, id)`. Active/inactive consistency check. |
| `guardian_child_links` | `id UUID PK`, `tenant_id`, `branch_id`, `guardian_id FK`, `child_id FK`, `ended_at`, `ended_reason_code`, `ended_reason_note` | Partial unique index on active `(guardian_id, child_id)`. End-reason consistency check. |
| `parent_membership_guardians` | `id UUID PK`, `tenant_id`, `branch_id`, `membership_id FK`, `guardian_id FK`, `ended_at`, `ended_reason_code`, `ended_reason_note` | Partial unique: one active mapping per membership, one active `(membership_id, guardian_id)` pair. End-reason consistency check. Triggers enforce parent role and active entities. |

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
| `child_contact_type` | `parent_carer`, `emergency_contact`, `authorised_collector` |

Used by: `child_leaving_records.reason_code`, `guardians.deactivation_reason_code`, `guardian_child_links.ended_reason_code`, `parent_membership_guardians.ended_reason_code`, `audit_logs.reason_code`.
`child_contact_type` is used by `child_contacts.contact_type`.

Tri-state text columns (no enum) on child sub-records: `child_profiles.disability_status`, `child_health_profiles.medical_conditions_status`/`prescribed_medication_status`/`dietary_requirements_status`, `child_safeguarding_profiles.social_services_status` + the six concern flags, `child_funding_records` six eligibility fields — all `('unknown','no','yes')` (immunisation_status is a separate five-value set).

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
| `idx_branches_core_hourly_rate` | `branches` | btree | Lookup by core hourly rate |
| `users_email_normalized_key` | `users` | UNIQUE btree | Case-insensitive email |
| `memberships_scope_id_unique` | `memberships` | UNIQUE btree | Scope composite key |
| `memberships_tenant_id_branch_id_user_id_key` | `memberships` | UNIQUE btree | One membership per user per branch |
| `children_scope_id_unique` | `children` | UNIQUE btree | Scope composite key |
| `idx_parent_membership_children_active_pair` | `parent_membership_children` | UNIQUE btree (partial) | One active mapping per parent-membership-child pair |
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
| `idx_child_profiles_child` | `child_profiles` | btree | Lookup by tenant/branch/child |
| `idx_child_contacts_child_type` | `child_contacts` | btree | Contacts by child + type + sort order (composite UNIQUE already covers scope lookup) |
| `idx_child_health_profiles_child` | `child_health_profiles` | btree | Lookup by tenant/branch/child |
| `idx_child_safeguarding_profiles_child` | `child_safeguarding_profiles` | btree | Lookup by tenant/branch/child |
| `idx_child_consent_records_child` | `child_consent_records` | btree | Lookup by tenant/branch/child |
| `idx_child_funding_records_child` | `child_funding_records` | btree | Lookup by tenant/branch/child |
| `idx_child_collection_settings_child` | `child_collection_settings` | btree | Lookup by tenant/branch/child |
| `idx_child_room_assignments_child_current` | `child_room_assignments` | btree (partial) | Current room per child |
| `idx_child_room_assignments_room_current` | `child_room_assignments` | btree (partial) | Current occupants per room (capacity) |
| `idx_child_billing_profiles_child` | `child_billing_profiles` | btree | Lookup by tenant/branch/child |
| `idx_child_leaving_records_child` | `child_leaving_records` | btree | Lookup by tenant/branch/child |
