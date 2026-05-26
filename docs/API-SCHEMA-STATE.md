# API Schema State

- **Verification date**: 2026-05-25
- **Workflow**: `make migrate-verify` (up → version → down -all → up → version)
- **Final migration version**: 10 (clean)
- **Migration tool**: golang-migrate (manual, not auto-run at API startup)

## Application Tables (14)

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

### Attendance

| Table | Key columns | Notes |
|---|---|---|
| `attendance_sessions` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `status`, `check_in_at`, `check_out_at`, `check_in_local_date`, `check_out_local_date`, `check_in_event_id FK`, `check_out_event_id FK`, `corrected_by_event_id FK` | Status: `open`, `complete`, `corrected`. Composite unique `(tenant_id, branch_id, id)`. Partial unique index ensures one open session per child per branch. Shape check: `open` → no checkout columns; `complete`/`corrected` → checkout required. `check_out_at > check_in_at`. |
| `attendance_events` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `session_id FK`, `event_type`, `occurred_at`, `local_date`, `recorded_by_user_id FK`, `recorded_by_membership_id FK`, `request_id`, `reason_code`, `reason_note`, `details JSONB` | Event type: `check_in`, `check_out`, `correction`. Composite unique `(tenant_id, branch_id, id)`. Routine events (check-in/check-out) cannot carry reason_code or reason_note. Correction events require a reason code (`missed_check_in`, `missed_check_out`, `incorrect_time`, `duplicate_entry`, `other`). `other` requires non-empty note. |

### Funding

| Table | Key columns | Notes |
|---|---|---|
| `funding_profiles` | `id UUID PK`, `tenant_id`, `branch_id`, `child_id FK`, `billing_month DATE`, `funded_allowance_minutes INTEGER`, `created_at`, `updated_at` | Unique `(tenant_id, branch_id, child_id, billing_month)`. `billing_month` must be first day of month. Allowance bounds: 0–44640. FK to `branches(tenant_id, id)` and `children(tenant_id, branch_id, id)`. |

### Audit

| Table | Key columns | Notes |
|---|---|---|
| `audit_logs` | `id UUID PK`, `tenant_id FK`, `branch_id FK`, `action_type`, `action_entity_type`, `action_entity_id`, `actor_user_id FK`, `actor_membership_id FK`, `details JSONB`, `request_id`, `reason_code`, `reason_note`, `created_at` | `reason_code` requires `reason_note` be NULL. `reason_code = other` requires non-empty note. |

## Enum / Custom Type

| Type | Values |
|---|---|
| `lifecycle_reason_code` | `duplicate_record`, `entered_in_error`, `left_nursery`, `safeguarding_direction`, `contact_update`, `access_revoked`, `other` |

Used by: `children.left_reason_code`, `guardians.deactivation_reason_code`, `guardian_child_links.ended_reason_code`, `parent_membership_guardians.ended_reason_code`, `audit_logs.reason_code`.

## Triggers & Functions

| Trigger | Table | Function | Behavior |
|---|---|---|---|
| `parent_membership_guardians_role_check` | `parent_membership_guardians` (BEFORE INSERT OR UPDATE) | `enforce_parent_membership_guardian_role()` | Membership must have `parent` role. |
| `parent_membership_guardians_active_entity_check` | `parent_membership_guardians` (BEFORE INSERT OR UPDATE) | `enforce_parent_mapping_active_entities()` | Membership must be active, guardian must be active. |
| `memberships_role_guardian_mapping_check` | `memberships` (BEFORE UPDATE OF role) | `prevent_non_parent_with_active_guardian_mapping()` | Cannot change role away from `parent` while active guardian mapping exists. |

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
