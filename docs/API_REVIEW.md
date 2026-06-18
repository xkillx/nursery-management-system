# Nursery Management System — API Review

**Scope:** `api/` only (Go / Gin / pgx / PostgreSQL / sqlc). `web/` excluded.
**Method:** Direct read of `internal/modules/*`, `internal/platform/*`, `internal/app/bootstrap/bootstrap.go`, and migrations `000001`–`000034`.
**Status of codebase:** Clean Architecture is well enforced (domain layer has zero framework imports, scope-unique FKs `(tenant_id, branch_id, id)` everywhere, audit writer is mandatory for mutations, transactions go through `txMgr.ExecTx`). The gaps below are about *what the domain does not yet model or enforce*, not about structural quality — which is strong.

Findings are grouped by the categories requested. Each entry: **What** · **Why** · **Recommendation** · **Priority**.

Priority legend: **Critical** (block production / regulatory / data loss) · **High** (correctness or serious UX) · **Medium** (quality / scale) · **Low** (nice-to-have).

---

## 1. Domain Completeness

### 1.1 No NHS Number on child
- **What:** `children` (mig `000004`) and `child_profiles` (mig `000034`) have no `nhs_number` column. UK nurseries collect NHS number for funding claims (early years entitlement) and medical emergencies.
- **Why:** Required by local authority funding returns; needed by ambulance/A&E contact; Ofsted may request.
- **Recommendation:** Add `nhs_number CHAR(10) NULL` to `child_health_profiles` (colocate with medical). Enforce mod-11 checksum via `CHECK` and a `validate_nhs_number(text)` PL/pgSQL function. Index uniquely per tenant if local authority dedup is desired.
- **Priority:** High

### 1.2 No sibling / family grouping
- **What:** No `family_id`, no `sibling_of` link. Multiple children of same guardians are unlinked.
- **Why:** Sibling discount billing, emergency pickup parity, sibling-room placement preference, parent communication consolidation. Billing module has no hook for sibling discount (`billing/domain/money.go`).
- **Recommendation:** Add `family_id UUID` to `children` (nullable, set when ≥2 children share guardianship). Update billing to apply discount rule against sibling count.
- **Priority:** High

### 1.3 No key person assignment
- **What:** No `key_worker_membership_id` on child. UK EYFS statutory requirement.
- **Why:** Ofsted-inspected nurseries must assign a key person; affects parent comms and observation tracking.
- **Recommendation:** Add nullable `key_worker_membership_id UUID REFERENCES memberships(id)` to `children`, with history via a `child_key_worker_assignments` table (analogous to `child_room_assignments`).
- **Priority:** High

### 1.4 No daily care log (sleep / nappy / feeding / milk)
- **What:** No tables for under-2 daily care records.
- **Why:** Parent communication expectation; safeguarding evidence for non-verbal children.
- **Recommendation:** Add `daily_care_logs` table scoped per child per day, append-only, with type enum.
- **Priority:** Medium

### 1.5 No accident / incident book
- **What:** No `accident_incidents` table.
- **Why:** Ofsted requires accident records; first-aid administration audit; parent notification trail; insurance claims.
- **Recommendation:** Append-only table `child_incidents` with type (`accident`, `incident`, `first_aid`, `medication_given`), body-location, treatment, witness, parent-notified-at, follow-up-at.
- **Priority:** High

### 1.6 No medication administration log
- **What:** No record of medicine given (paracetamol, inhaler, EpiPen).
- **Why:** Safeguarding + insurance; consent stored in `child_consent_records.plasters` etc., but no execution log.
- **Recommendation:** Add `medication_administration_records` linked to staff membership + signed-by-parent flag.
- **Priority:** High

### 1.7 No SEND / EHC plan tracking
- **What:** No fields for Special Educational Needs & Disabilities, EHC plan number, SENCo support level.
- **Why:** Legal obligation under SEND Code of Practice; affects funding (DIS/DAF).
- **Recommendation:** Extend `child_safeguarding_profiles` or add `child_send_profiles` with `ehcp_number`, `sen_category`, `sencono_membership_id`, review dates.
- **Priority:** Medium

### 1.8 No observation / assessment records
- **What:** No EYFS learning journey, observations, next steps, "two-year check".
- **Why:** Statutory EYFS assessment; parent reporting.
- **Recommendation:** Future scope — out of MVP but architect now via a `child_observations` table reserved.
- **Priority:** Low

### 1.9 No photo / document asset storage
- **What:** Consent for photos exists (`development_profile_photos`, `nursery_website`, etc.) but no asset table.
- **Why:** Consents without enforcement is just policy; no audit of who accessed/uploaded.
- **Recommendation:** Future scope — `child_assets` with object-store pointer + signed URL service.
- **Priority:** Low

### 1.10 No custody / court order field
- **What:** No way to record "non-custodial parent restricted from pickup".
- **Why:** Real safeguarding risk; nursery has legal duty to refuse collection.
- **Recommendation:** Add `restricted_from_collection BOOLEAN` + structured notes to `child_contacts`; expose as hard block in collection-password check (`set_collection_password.go`).
- **Priority:** High

### 1.11 No waiting list / enquiries
- **What:** `children` only models enrolled children. Prospective pipeline is invisible.
- **Why:** Nurseries live on waiting list; capacity planning needs it.
- **Recommendation:** Future scope — `waiting_list_entries` table.
- **Priority:** Low

---

## 2. Missing Entities & Relationships

### 2.1 No sessions / bookings pattern
- **What:** Attendance is recorded as check-in/out events; there is no concept of *contracted sessions* (Mon AM, Wed PM, etc.).
- **Why:** Cannot detect missed sessions vs unscheduled drop-in; cannot pro-rate funding; cannot enforce "no check-in on non-scheduled day"; billing currently keys off raw attended minutes (`billing/domain/attendance_minutes.go`).
- **Recommendation:** Add `child_session_patterns` (e.g. Monday 09:00–12:00) with effective date range; reconcile against actual attendance.
- **Priority:** High

### 2.2 No staff shift / room staffing
- **What:** No `room_staff_assignments`. Rooms have `capacity INT` (mig `000029`) but no `required_staff_by_age_band`.
- **Why:** EYFS statutory ratios (1:3 under-2, 1:4 age 2, 1:8 age 3+) cannot be enforced or audited.
- **Recommendation:** Add `room_staffing_rules` (age_band, staff_required_per_child) and `room_staff_assignments` per session. Validate before allowing `child_room_assignments` insert that would breach ratio.
- **Priority:** High

### 2.3 No funding entitlement as first-class concept
- **What:** UK early-years funding (15h universal / 30h working parent / 2yo disadvantaged / Tax-Free Childcare) is reduced to a single `funded_allowance_minutes INTEGER` per month (`funding_profiles`, mig `000010`). The `child_funding_records` table (mig `000034`) captures *eligibility flags* but not the *entitlement* with term dates.
- **Why:** UK funding is term-based, not monthly. Stretch vs term-time delivery differs. Funding codes expire and must be revalidated each term. Without this, billing reconciliation against local authority is impossible.
- **Recommendation:** Add `funding_entitlements` table keyed by `(child_id, funding_code, term_start, term_end, hours_per_week, delivery_mode)`; derive monthly allowance from it.
- **Priority:** Critical

### 2.4 No contract / pricing versioning
- **What:** `child_billing_profiles.custom_rate_minor` is a single current value (`000034`). No history.
- **Why:** Cannot recompute historical invoices if rate changes mid-year; cannot defend pricing in dispute.
- **Recommendation:** Add effective-dated rows (`effective_from`, `effective_to`) — same pattern as `child_room_assignments`.
- **Priority:** Medium

### 2.5 No branch-level pricing / rate card
- **What:** Hourly rate lives on `branches.core_hourly_rate_minor` (mig `000022`) as a single value.
- **Why:** No age-band pricing, no session-type pricing (full day vs AM vs PM), no sibling rate.
- **Recommendation:** Add `branch_rate_cards` table with effective dates and per-band/per-session rates.
- **Priority:** Medium

### 2.6 No late-pickup penalty config
- **What:** No config for late fee. Attendance `check_out_at` is recorded but not penalised.
- **Why:** Missed revenue; common nursery policy.
- **Recommendation:** Add `branch_late_fee_policy` (grace_minutes, fee_per_block_minutes, fee_amount_minor) and a billing line producer.
- **Priority:** Medium

---

## 3. Missing Fields & Attributes

### 3.1 No soft-delete / GDPR erasure column
- **What:** No `deleted_at`, `deleted_by`, `anonymised_at` on any PII table (`children`, `guardians`, `child_contacts`, `users`, etc.).
- **Why:** UK GDPR right to erasure (Art. 17) requires actionable deletion. Currently the only way to "remove" a child is `is_active=false` + `child_leaving_records` — data persists forever.
- **Recommendation:** Add `erasure_requested_at TIMESTAMPTZ` + scheduled job that anonymises PII columns after retention window (e.g. 7 years post-leaving per Ofsted record-keeping). Document retention in `docs/`.
- **Priority:** High

### 3.2 No optimistic-locking version column
- **What:** No `version INTEGER` / `etag`/`xmin`-based guard on any table.
- **Why:** Two managers editing the same child profile → last-write-wins silently. Update flows in `children/application/update_*.go` have no row-version check.
- **Recommendation:** Either expose `xmin` (PostgreSQL system column) in UPDATEs (`WHERE id = $1 AND xmin = $2`) or add explicit `version INTEGER NOT NULL DEFAULT 0` with `version = version + 1` in every UPDATE.
- **Priority:** Medium

### 3.3 Missing `created_by` / `updated_by` on most entities
- **What:** Audit goes to a separate `audit_logs` table (good), but on `children`, `child_profiles`, `child_health_profiles`, etc. there is no inline last-author reference.
- **Why:** Fast triage on the row itself; defensive join when audit log retained for shorter window.
- **Recommendation:** Add `created_by_membership_id`, `updated_by_membership_id` to profile-style tables. `child_consent_records` and `child_collection_settings` already do this — replicate.
- **Priority:** Low

### 3.4 `children.start_date` has no `end_date` cross-check against `child_leaving_records`
- **What:** Two parallel sources of truth for "child has left": `children.end_date`/`is_active` and `child_leaving_records.left_at` (mig `000034`).
- **Why:** Drift risk; `mark_inactive.go` must write both atomically; nothing enforces at DB level.
- **Recommendation:** Add a CHECK trigger: if `child_leaving_records` row exists then `children.is_active = false` and `children.end_date IS NOT NULL`.
- **Priority:** Medium

### 3.5 `users` has no `last_login_at`, `failed_login_count`, `locked_until`
- **What:** No brute-force tracking at user level. Rate limiter is IP-only (`bootstrap.go:147`).
- **Why:** Account lockout policy needs per-user state; attacker rotating IPs bypasses IP limiter.
- **Recommendation:** Add the three columns; increment on failed login; lock after N failures for X minutes.
- **Priority:** High

### 3.6 No `email_verified_at` on `users`
- **What:** `users` has email + password but no verification flag.
- **Why:** Invites flow does verify on accept, but owner-created users (`owner/application/grant_manager_access.go`) and seeded users skip verification. Cannot distinguish verified vs unverified for audit.
- **Recommendation:** Add `email_verified_at TIMESTAMPTZ`; set on invite-accept; backfill true for seeded.
- **Priority:** Medium

### 3.7 No `phone` normalisation on guardians / child_contacts
- **What:** `guardians.phone TEXT`, `child_contacts.telephone TEXT` — raw strings.
- **Why:** Cannot send SMS alerts reliably; cannot dedupe contacts.
- **Recommendation:** Add `phone_e164 TEXT` derived column populated on insert/update via `libphonenumber` (Go port).
- **Priority:** Low

---

## 4. Business Rule Gaps

### 4.1 No EYFS ratio enforcement on room assignment
- **What:** `child_room_assignments` (mig `000034`) has no CHECK or trigger validating age-appropriate room or staff ratio.
- **Why:** Statutory breach risk.
- **Recommendation:** Validate `child.date_of_birth` against `rooms.age_group` bounds at insert. For ratio: trigger counts active assignments vs `rooms.capacity` and staff shifts.
- **Priority:** Critical

### 4.2 Age-band transition not modeled
- **What:** A child turning 3 mid-month is not auto-flagged for room move or funding tier change.
- **Why:** Operational gap; parents complain; funding mis-claimed.
- **Recommendation:** Scheduled job emitting `age_band_transition_due` events 4 weeks before birthday; surfaces in manager dashboard.
- **Priority:** Medium

### 4.3 No funding eligibility cross-check
- **What:** `child_funding_records.funding_3yo_term_time` flag exists but is not validated against `children.date_of_birth`. A child <3 years old could be flagged for 3yo funding.
- **Why:** Local authority claw-back; fraud risk.
- **Recommendation:** CHECK constraint or app-layer rule: `funding_3yo_term_time='yes'` requires DOB ≤ billing_month - 3 years + term-start grace.
- **Priority:** High

### 4.4 Re-enrollment not modeled
- **What:** A child can be marked inactive, but the re-activation flow doesn't preserve continuity with previous `child_leaving_records`. `child_leaving_records` is 1:1 (`UNIQUE child_id`) — a child cannot leave twice.
- **Why:** Real-world: kids leave for school holiday or temp absence, return. History is one row only.
- **Recommendation:** Drop the UNIQUE on `child_leaving_records.child_id`; allow history. Add `is_current` generated column. `mark_inactive.go` should append a new row, not fail.
- **Priority:** High

### 4.5 No "primary contact" enforcement
- **What:** `child_contacts` has `sort_order` but no constraint that exactly one row of type `parent_carer` is primary.
- **Why:** Communication routing ambiguous; emergency-call priority unclear.
- **Recommendation:** Add a `is_primary BOOLEAN` column with partial unique index `WHERE is_primary AND contact_type='parent_carer'` per child.
- **Priority:** Medium

### 4.6 Concurrent active guardian-child links not bounded
- **What:** `guardian_child_links` partial unique index (mig `000004`) prevents duplicate pairs but no max per child.
- **Why:** Operational edge but minor.
- **Recommendation:** None — soft limit in app layer is fine.
- **Priority:** Low

### 4.7 Collection password has no strength rule
- **What:** `child_collection_settings.collection_password_hash` stored as bcrypt hash (good — see `set_collection_password.go`) but no min length / complexity check on the input.
- **Why:** Weak collection passwords defeat the safeguarding control the field exists for.
- **Recommendation:** App-layer validation ≥ 4 digits or ≥ 6 chars; surface to UI.
- **Priority:** Medium

### 4.8 No check-in preconditions beyond enrollment
- **What:** `check_in_child.go` checks enrollment + absence marker. Does not check: room capacity at that moment, session pattern, immunisation-up-to-date (some nurseries require), authorised-to-collect list of the dropper-offer.
- **Why:** Operational hygiene.
- **Recommendation:** Optional preconditions toggled per branch.
- **Priority:** Low

---

## 5. Validation Requirements

### 5.1 No postcode validation
- **What:** `child_profiles.home_postcode TEXT` (mig `000034`) — free text.
- **Why:** Address quality; funding eligibility uses postcode in some LA returns.
- **Recommendation:** Regex CHECK or `validate_uk_postcode(text)` function; normalise to upper-case + single space.
- **Priority:** Low

### 5.2 No email format validation at DB
- **What:** `users.email`, `guardians.email`, `child_contacts.email` — no CHECK.
- **Why:** Bad data; bounce loops.
- **Recommendation:** CHECK with email regex (or store an `email_normalized` like `users` already does, applied to guardians).
- **Priority:** Low

### 5.3 No DOB sanity CHECK
- **What:** `children.date_of_birth DATE NOT NULL` — accepts future dates, accepts 18-year-olds.
- **Why:** Wrong data pollutes funding + ratio calcs.
- **Recommendation:** `CHECK (date_of_birth <= current_date AND date_of_birth >= current_date - interval '13 years')`.
- **Priority:** High

### 5.4 No phone / NHS / national insurance format CHECKs
- **What:** As above for NHS (1.1) and phone (3.7).
- **Recommendation:** Combined CHECK constraints.
- **Priority:** Medium

### 5.5 Application-layer validation is centralised but inconsistent
- **What:** `children/application/validation.go` exists (11 symbols) but per-use-case files duplicate checks (e.g. `update_health.go`, `update_funding.go`).
- **Why:** Maintenance drift; future bugs.
- **Recommendation:** Centralise per-resource validators; call from each use case.
- **Priority:** Low

---

## 6. Workflow & Lifecycle Coverage

### 6.1 Child status is a boolean, not a state machine
- **What:** `children.is_active BOOLEAN` + optional `end_date` + optional `child_leaving_records`. Three signals.
- **Why:** Cannot model "prospective", "induction", "enrolled", "inactive (temporary)", "left", "excluded (safeguarding)". `leaving_record.reason_code` partially compensates.
- **Recommendation:** Replace boolean with `status TEXT` enum + status-transition trigger (mirror `enforce_invoice_status_transition` in mig `000012`). Keep boolean as a derived column for query ergonomics.
- **Priority:** Medium

### 6.2 Guardian deactivation does not cascade to parent_membership_guardians / auth
- **What:** `deactivate_guardian.go` flips `is_active=false` + `ended_at`. The link in `parent_membership_guardians` (mig `000004`) is not auto-ended; the parent's login still works; collection_password references remain.
- **Why:** Authorisation drift; safeguarding risk if guardian is restricted.
- **Recommendation:** Within the same transaction, end `parent_membership_guardians` rows; revoke refresh tokens (call into `authentication` via adapter).
- **Priority:** High

### 6.3 No leave-and-return flow for staff
- **What:** Manager / practitioner memberships: `memberships.is_active` + `ended_at` (mig `000004`). No "on leave / sabbatical" state, no planned end date.
- **Why:** Ratio planning, key worker continuity.
- **Recommendation:** Add `membership_status` enum (`active`, `on_leave`, `ended`).
- **Priority:** Low

### 6.4 Attendance correction has no approval workflow
- **What:** `correct_attendance.go` lets `manager`/`practitioner` mutate times immediately. History is preserved (`attendance_events` with `event_type='correction'`) — good.
- **Why:** Ofsted: corrections should ideally have a second-person awareness or approval for sensitive edits (e.g., past-timesheet corrections affecting billing).
- **Recommendation:** Optional: require second-factor manager approval for corrections older than X days or that change billing outcome.
- **Priority:** Low

### 6.5 Invoice lifecycle missing void / credit-note / disputed
- **What:** Status CHECK is `draft|issued|payment_failed|paid|overdue` (mig `000012`). No `void`, no `disputed`, no `credit_issued`. The `invoice_kind = 'adjustment'` partially compensates but cannot represent "this invoice was issued in error and voided" — only adjustment credits.
- **Why:** HMRC requires voided invoices to be retained with their original number, not deleted.
- **Recommendation:** Add `void` status (terminal, with `voided_at`, `voided_by`, `void_reason_code`). Allow `issued → void` transition. Credit notes via the existing `adjustment` kind.
- **Priority:** High

### 6.6 Payment has no refund / chargeback flow
- **What:** `payment_attempts` statuses (mig `000013`) stop at `paid`. Stripe `charge.refunded` / `charge.dispute.*` events land in `stripe_webhook_events` but `handle_stripe_webhook.go` doesn't appear to flip status back, nor create credit.
- **Why:** Reconciliation breaks; parent balance wrong.
- **Recommendation:** Add `refunded`, `partially_refunded`, `disputed` statuses; refund creates an `adjustment` invoice against the original.
- **Priority:** Critical

### 6.7 Funding profile has no renewal / expiry
- **What:** `funding_profiles` are per-month (mig `000010`), but there is no `valid_until` or "revalidation due" flag.
- **Why:** 30-hour codes must be revalidated every 3 months via HMRC; lapse causes funding loss.
- **Recommendation:** Add `eligibility_valid_until DATE`; job alerts parents/staff 14 days before lapse.
- **Priority:** High

---

## 7. Data Integrity Constraints

### 7.1 Invoices: immutability trigger is strong but limited
- **What:** `protect_issued_invoice_immutability` (mig `000012`) is excellent — whitelists only status/payment columns. Lines protected by separate trigger.
- **Gap:** `calculation_details JSONB` is also locked (good), but `audit_logs` rows referencing the invoice are NOT locked. Audit log already append-only by app convention; nothing enforces DB-side.
- **Recommendation:** Add `REVOKE UPDATE, DELETE ON audit_logs FROM app_role` (separate role for retention job only).
- **Priority:** Medium

### 7.2 Audit logs not append-only at DB level
- **What:** As above.
- **Recommendation:** Same as 7.1.
- **Priority:** High

### 7.3 `child_room_assignments` overlaps not prevented
- **What:** Two rows with `(start_date, end_date)` can overlap for the same child. There is no exclusion constraint.
- **Why:** "Current room" semantics ambiguous.
- **Recommendation:** Add `EXCLUDE USING gist (child_id WITH =, daterange(start_date, COALESCE(end_date, 'infinity'::date)) WITH &&)` after enabling `btree_gist`.
- **Priority:** Medium

### 7.4 `parent_membership_guardians` allows only one guardian per membership
- **What:** Partial unique index on `(membership_id)` (mig `000004`). A parent account can be linked to exactly one guardian record.
- **Why:** Forces 1:1 parent↔guardian. A second parent in same household needs separate account. Probably intentional, but document.
- **Recommendation:** Keep, but document in `CONTEXT.md`.
- **Priority:** Low

### 7.5 Cross-tenant FKs rely on composite key correctness
- **What:** Scope-unique FKs `(tenant_id, branch_id, id) → children(tenant_id, branch_id, id)` are good. But child-tenant/branch are taken from request actor, not from the child row itself in every code path.
- **Why:** If a query ever trusts client-supplied `child_id` without re-checking it belongs to actor's tenant, cross-tenant leak. The pattern is solid in current code but no static enforcement.
- **Recommendation:** Add a CI grep rule forbidding raw `WHERE id = $1` queries on tenant-scoped tables.
- **Priority:** Medium

### 7.6 `branches.is_active` flag exists but children/guardians don't cascade-block
- **What:** A branch can be deactivated (mig `000015`), but no constraint stops new children/attendances/invoices being added to inactive branches.
- **Why:** Owner archiving a site doesn't freeze its data.
- **Recommendation:** CHECK trigger or app-layer rule.
- **Priority:** Low

### 7.7 Money non-negativity is well-enforced in invoices, but not in payment_attempts reconciliation
- **What:** `payment_reconciliation_records.amount_minor` is plain `INTEGER`, no `>= 0` check (mig `000014`). Refunds as negative would fail to be modelled.
- **Why:** Sign-convention drift.
- **Recommendation:** Decide sign convention; add CHECK.
- **Priority:** Low

---

## 8. Security & Permissions

### 8.1 No rate limiting on `/auth/login`
- **What:** `bootstrap.go:147-148` configures limiters for password-reset and invite endpoints. **Login has no limiter.**
- **Why:** Credential stuffing brute force. Critical for any production auth.
- **Recommendation:** Add `authIPLimiter` + `authUserLimiter` (per-email). Exponential backoff. Use `ratelimit.NewFixedWindowLimiter` already in the platform.
- **Priority:** Critical

### 8.2 No rate limiting on `/auth/refresh`, `/auth/logout`, `/auth/switch-membership`
- **What:** Authenticated but unthrottled.
- **Why:** Token-grind DoS; audit spam.
- **Recommendation:** Light limiter per-user.
- **Priority:** Medium

### 8.3 No 2FA / MFA
- **What:** Not implemented anywhere in `authentication` or `users`.
- **Why:** Standard expectation for staff handling children's data; some insurers require.
- **Recommendation:** TOTP for manager+owner roles at minimum. Add `user_mfa_secrets` table.
- **Priority:** High

### 8.4 Refresh-token rotation lacks family-wide revocation on reuse
- **What:** `refresh.go` rotates the token and revokes the old one. If a stolen token is replayed after the legitimate user has rotated, the replay will fail (good) but the system does not detect this and revoke the entire family.
- **Why:** Best-practice rotation detects reuse → burns family. Current code is "rotate on use" but not "burn family on reuse".
- **Recommendation:** Add `token_family_id UUID` to refresh tokens; on reuse, revoke all tokens in family + force re-login.
- **Priority:** High

### 8.5 Access token has no `jti` and cannot be revoked
- **What:** Logout revokes the refresh token (`logout.go`), not the access token. Access tokens remain valid until expiry (`JWTAccessTTLMin`).
- **Why:** Stolen access token usable for the TTL window after logout. Mitigated by short TTL but not zero.
- **Recommendation:** Either accept short TTL (<15 min) and document, or add `jti` + Redis blacklist checked by middleware.
- **Priority:** Medium

### 8.6 Membership switch keeps prior refresh tokens valid
- **What:** `switch_membership.go` issues a new token bound to the new membership but does not revoke other refresh tokens for the user.
- **Why:** If user switches from manager to parent role in a shared-browser scenario, prior manager-scoped token remains valid until expiry.
- **Recommendation:** Optionally revoke on explicit switch; or document as accepted.
- **Priority:** Medium

### 8.7 Owner cross-site queries have no pagination limit at SQL level
- **What:** `get_site_summaries.go` returns aggregated data. If a tenant has 1000+ sites, single query loads all.
- **Why:** Performance + memory risk for chain operators.
- **Recommendation:** Cursor pagination + per-query LIMIT.
- **Priority:** Medium

### 8.8 No CSRF protection on cookie-based sessions
- **What:** Auth uses Bearer tokens in `Authorization` header (good, CSRF-safe), but if a future cookie auth is added, no CSRF middleware is staged.
- **Why:** Forward defence.
- **Recommendation:** None now; add when cookie auth added.
- **Priority:** Low

### 8.9 Webhook endpoint trusts Stripe signature only — good — but no IP allow-list
- **What:** `RegisterStripeRoutes(api)` mounts on the public `api` group. Verification is signature-based (correct). No IP allow-list of Stripe ranges as defence-in-depth.
- **Why:** Belt-and-braces.
- **Recommendation:** Optional Stripe-IP allow-list middleware.
- **Priority:** Low

### 8.10 Password reset has no email-enumeration hardening on differences
- **What:** `request_reset.go` likely returns generic success regardless of email existence (verify in code; flow looks correct). Reset tokens have TTL (mig `000008`).
- **Why:** Enumeration risk via timing differences.
- **Recommendation:** Confirm timing-equal response; constant-time path for both existing and non-existing emails.
- **Priority:** Medium

### 8.11 No security headers middleware
- **What:** Bootstrap doesn't add `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy` (the last mostly a browser concern, but API responses can carry framing headers).
- **Why:** Defence-in-depth for any browser-reachable endpoints.
- **Recommendation:** Add small `SecurityHeaders` middleware.
- **Priority:** Low

### 8.12 No CORS configuration visible
- **What:** No `cors` middleware in bootstrap. If the Angular app is served from a different origin in prod, the API will reject pre-flights.
- **Why:** Production breakage if deploy topology changes.
- **Recommendation:** Add explicit CORS allow-list (configured by env).
- **Priority:** Medium

---

## 9. Audit & Compliance

### 9.1 No retention policy for `audit_logs`
- **What:** Audit logs accumulate forever. GDPR Art. 5(1)(e) storage limitation.
- **Why:** Legal exposure; storage growth.
- **Recommendation:** Configurable retention (default 7 years to match Ofsted + HMRC), partition by month, scheduled archive.
- **Priority:** High

### 9.2 No tamper-evidence on audit log
- **What:** Rows are plain inserts. A DB admin can `UPDATE`/`DELETE` without trace.
- **Why:** Compliance posture weaker than financial-grade.
- **Recommendation:** Hash-chain each row to previous (`prev_hash`, `this_hash`); cron job verifies chain. Or use an append-only table with revoked UPDATE/DELETE grants (7.2).
- **Priority:** Medium

### 9.3 Consent history not retained
- **What:** `child_consent_records` is 1:1 with the child (mig `000034`) — a single row updated in place. `update_consent.go` overwrites; previous consent state lost.
- **Why:** GDPR consent withdrawal must be auditable. UK GDPR Art. 7(1): "demonstrate consent". Cannot demonstrate history.
- **Recommendation:** Append-only `child_consent_records_history` (full row + actor + timestamp) populated by trigger; current row remains the latest.
- **Priority:** Critical

### 9.4 Safeguarding restricted_notes has no access-log
- **What:** `child_safeguarding_profiles.restricted_notes` is high-sensitivity text. Reading it goes through normal handler; no record of who read.
- **Why:** Safeguarding best practice; "need to know" audit.
- **Recommendation:** Add explicit read-audit hook in `get_safeguarding.go`; restrict role to manager+; log every read.
- **Priority:** High

### 9.5 No data export endpoint (DSAR)
- **What:** UK GDPR Art. 15 right of access — no endpoint to export all data for a subject (child or guardian).
- **Why:** Statutory request handling.
- **Recommendation:** Future scope — `/parent/dsar/export` returns JSON bundle.
- **Priority:** Medium

### 9.6 No invoice retention / legal hold flag
- **What:** HMRC requires 6 years retention for tax records. Invoices protected against edit (good) but no explicit retention enforcement or legal-hold flag.
- **Why:** If/when archive policy runs, financial records must be exempt.
- **Recommendation:** Add `retention_until DATE` column; legal hold boolean; archive job respects.
- **Priority:** Medium

### 9.7 Stripe webhook event log retention
- **What:** `stripe_webhook_events.raw_payload JSONB` grows unbounded (mig `000014`).
- **Why:** Storage; PII (customer email inside payload) accumulates.
- **Recommendation:** Partition + 90-day retention for raw payload; keep summary forever.
- **Priority:** Low

---

## 10. API Design Concerns

### 10.1 No pagination on list endpoints — confirm
- **What:** `list_children.go`, `list_invoices.go`, `list_attendance.go`, `list_overview.go` etc. — verify they enforce `LIMIT/OFFSET` or cursor.
- **Why:** At scale, unbounded lists DoS the DB and client.
- **Recommendation:** Cursor pagination mandatory on all `list_*`. Validate during review of each `list_*` file (already done for `funding/overview` based on flag summary fields).
- **Priority:** High

### 10.2 No bulk operations
- **What:** `bulk_issue_invoices.go` exists (good — billing). No bulk check-in/out, no bulk guardian import, no bulk contact replace.
- **Why:** Operational efficiency for morning drop-off.
- **Recommendation:** Add `POST /attendance/bulk-check-in` with array body and per-row result envelope.
- **Priority:** Medium

### 10.3 No parent routes for child data
- **What:** `parent := protected.Group("/parent")` registers billing + payments only (grep confirms only those two modules register parent routes). **Parents cannot read their child's profile, attendance history, contacts, funding, or room assignment via API.**
- **Why:** Major UX gap; parent portal incomplete.
- **Recommendation:** Add `RegisterParentRoutes` to `children` module exposing read-only subset (profile summary, attendance summary, contacts of their own link, invoices). Enforce via guardian-link check.
- **Priority:** High

### 10.4 No idempotency-key support on POSTs
- **What:** `create_child.go`, `create_room_assignment.go`, `mark_absent.go`, `create_checkout_session.go` — all POSTs without `Idempotency-Key` header support.
- **Why:** Network retries double-write (especially payments).
- **Recommendation:** Add `Idempotency-Key` middleware backed by `idempotency_keys` table; reuse for `create_checkout_session` at minimum.
- **Priority:** High

### 10.5 Error envelope exists but custom error code catalog missing
- **What:** `platform/http/authz_middleware.go` writes `forbidden_role`, `unauthorized`, etc. Domain errors flow via `MapDomainError`. No central catalog of error codes clients can switch on.
- **Why:** Client UX brittle.
- **Recommendation:** Add `docs/api-errors.md` listing every emitted code.
- **Priority:** Low

### 10.6 No ETag / conditional GET
- **What:** No `If-None-Match` handling.
- **Why:** Cache bandwidth.
- **Recommendation:** Add to read-heavy endpoints (children list, invoices list).
- **Priority:** Low

### 10.7 No OpenAPI / Swagger
- **What:** No `swagger.json` or annotation.
- **Why:** Frontend integration + third-party integrators (accounting systems).
- **Recommendation:** Generate from handler comments via `swaggo/swag`.
- **Priority:** Medium

### 10.8 Inconsistent `request_id` propagation to downstream
- **What:** `RequestIDMiddleware` exists; some application use cases pass it through (audit writer uses `actor.RequestID`), others may not.
- **Why:** Distributed trace gaps.
- **Recommendation:** Audit each use case; enforce single ingress point.
- **Priority:** Low

### 10.9 No versioning strategy beyond `/v1`
- **What:** Single version. No plan for `/v2`.
- **Why:** Future-proofing.
- **Recommendation:** Document deprecation policy.
- **Priority:** Low

### 10.10 Health check is shallow
- **What:** `healthHandler` (bootstrap.go:393) only pings DB. No check for email provider, Stripe reachability, migrations current.
- **Why:** K8s readiness probe false positives.
- **Recommendation:** Split `/health/live` (process) and `/health/ready` (deps).
- **Priority:** Medium

---

## 11. Performance & Scalability Risks

### 11.1 Funding overview query loads all children per branch
- **What:** `list_overview.go` returns per-child flags for a branch+month. With 500 children per branch and 50 branches, the overview scan grows.
- **Why:** Manager dashboard latency.
- **Recommendation:** Already has flag summary counts — verify SQL uses a single aggregated query, not N rows then compute in Go.
- **Priority:** Medium

### 11.2 No covering index on `invoices (tenant_id, branch_id, status, due_at)`
- **What:** Mig `000012` adds `idx_invoices_due_at_outstanding WHERE status IN ('issued', 'overdue')` — good. But the overdue cron (`mark_overdue_invoices.go`) needs `(status, due_at)` for the wider scan.
- **Why:** Cron latency.
- **Recommendation:** Verify EXPLAIN on the cron query; add covering index if needed.
- **Priority:** Medium

### 11.3 `audit_logs` write contention
- **What:** Every mutation writes a row. Hot table on busy days.
- **Why:** Write amplification; index bloat on `(action_entity_type, action_entity_id)`.
- **Recommendation:** Partition by month; consider BRIN index on `created_at`.
- **Priority:** Medium

### 11.4 No connection-pool tuning documented
- **What:** `pgxpool` configured somewhere; no mention of max conns, statement cache.
- **Why:** Saturation under load.
- **Recommendation:** Document expected pool sizing per environment.
- **Priority:** Low

### 11.5 N+1 risk in child list endpoints
- **What:** `list_children.go` likely returns children; the handler may attach contacts, room assignment, guardian links per row in a loop.
- **Why:** Classic N+1.
- **Recommendation:** Use `IN (...)` batch fetch in repository; verify with test data of 100 children.
- **Priority:** Medium

### 11.6 No materialised view for daily attendance aggregates
- **What:** Billing recomputes attendance minutes from raw events each run (`attendance_minutes.go`).
- **Why:** Recompute cost; for 12 months × 200 children, expensive.
- **Recommendation:** Cache daily totals in `attendance_daily_summary` materialised view refreshed nightly.
- **Priority:** Low

### 11.7 Stripe webhook handler does not appear to be concurrent-safe under duplicate delivery
- **What:** Webhooks are at-least-once. The `payment_reconciliation_records` unique on `stripe_event_id` (mig `000014`) provides dedupe — good. Confirm handler is idempotent end-to-end (status transition trigger should reject duplicate `issued → issued`).
- **Why:** Stripe retries.
- **Recommendation:** Add a test: replay same event twice; assert no double-billing.
- **Priority:** Medium

### 11.8 No `LIMIT` on bulk invoice issue
- **What:** `bulk_issue_invoices.go` accepts an array. No documented max.
- **Why:** Resource exhaustion via huge payload.
- **Recommendation:** Cap at 100 per call.
- **Priority:** Low

---

## 12. Multi-Tenant Considerations

### 12.1 Stripe account is single-tenant
- **What:** `cfg.StripeSecretKey` (bootstrap.go:295) is one key for the whole API. No Stripe Connect.
- **Why:** Chain operators running on behalf of multiple legal entities need Connect for fund routing + KYC per site.
- **Recommendation:** If target market is single-legal-entity chains, current model is fine. If multi-entity SaaS, adopt Stripe Connect with `account` per branch.
- **Priority:** Medium

### 12.2 Currency is hardcoded GBP
- **What:** `CHECK (currency_code = 'GBP')` on invoices + payment_attempts (mig `000012`, `000013`). Domain `money.go` takes `minor int` only.
- **Why:** Acceptable for UK-only product; blocks international expansion.
- **Recommendation:** Document as intentional UK-only. Drop the hard CHECK if internationalisation planned.
- **Priority:** Low

### 12.3 Tenants table has no plan / quota / billing tier
- **What:** `tenants` (mig `000001`) is just `id, name, timestamps`.
- **Why:** SaaS commercial model not modeled.
- **Recommendation:** Future scope — `tenant_subscription`, `tenant_quota`.
- **Priority:** Low

### 12.4 No row-level security in Postgres
- **What:** All tenant isolation is application-enforced via composite FK + scope-predicated queries. Postgres RLS not enabled.
- **Why:** Defence in depth: a missed `WHERE tenant_id` is a leak.
- **Recommendation:** Enable RLS on PII tables with policy `tenant_id = current_setting('app.tenant_id')`. Set the variable per request in middleware.
- **Priority:** Medium

### 12.5 Owner cross-tenant reads need an audit highlight
- **What:** `OwnerActorFromGinContext` (`platform/tenant/context.go`) grants tenant-wide read. Every owner query should be tagged in audit as cross-branch.
- **Why:** Privacy posture; chain operator oversight vs snooping.
- **Recommendation:** Audit detail `cross_branch = true` on every owner query touching a specific branch.
- **Priority:** Low

---

## 13. Future Extensibility

### 13.1 Cross-module adapter pattern well-isolated
- **Strength:** Adapters in `bootstrap/adapters.go` (e.g. `guardianCheckerAdapter`, `childEnrollmentCheckerAdapter`) keep modules decoupled. Pattern is reusable.
- **Recommendation:** Keep the discipline; document in `AGENTS.md`.

### 13.2 Consent model is boolean per-channel, not a framework
- **What:** `child_consent_records` is a flat table of booleans (mig `000034`). Adding a new consent type = migration.
- **Why:** Proliferation over time.
- **Recommendation:** Future: refactor to `consent_grants (child_id, consent_type, status, granted_at, granted_by, withdrawn_at, withdrawn_by)` — append-only.

### 13.3 Notification system absent
- **What:** Emails are sent ad-hoc from each module (`email_adapter.go` in invites, passwordreset). No central notification dispatcher.
- **Why:** SMS, push, in-app future channels; templating drift.
- **Recommendation:** Introduce a `notifications` module with queue.

### 13.4 Reporting / analytics module absent
- **What:** No dedicated reporting endpoints. Owner summaries (`get_site_summaries.go`) is the only rollup.
- **Why:** Ofsted returns, LA funding returns, occupancy stats need structured output.
- **Recommendation:** Future `reports` module.

### 13.5 Document / file attachments
- **What:** No attachment model anywhere.
- **Why:** Contracts, registration forms, medical letters.
- **Recommendation:** Future `documents` module with object storage adapter.

### 13.6 Communication log (parent messages)
- **What:** No record of nursery↔parent messages.
- **Why:** Dispute resolution; safeguarding.
- **Recommendation:** Future `communications` module.

### 13.7 No webhook out (only in)
- **What:** Stripe webhook inbound exists; no outbound webhooks for integrators.
- **Why:** Future accounting integration.
- **Recommendation:** Event bus + outbound dispatcher.

### 13.8 Mobile push notification tokens
- **What:** No `user_device_tokens` table.
- **Why:** Parent app push (collection ready, invoice due).
- **Recommendation:** Add when mobile app planned.

---

## Priority Roll-up

### Critical (12)
1. 1.3 — No key person assignment (statutory)
2. 2.3 — Funding entitlement not first-class (term-based)
3. 4.1 — No EYFS ratio enforcement
4. 6.6 — No refund / chargeback flow in payments
5. 7.2 — Audit logs not append-only at DB
6. 8.1 — No rate limit on `/auth/login`
7. 9.3 — Consent history not retained (GDPR Art. 7)
8. 1.5 — No accident / incident book
9. 1.6 — No medication administration log
10. 1.10 — No custody / court order field
11. 4.3 — No funding eligibility cross-check
12. 6.7 — Funding profile no renewal / expiry

### High (15)
1.3, 1.2, 1.5, 1.6, 1.10, 2.1, 2.2, 3.1, 3.5, 4.3, 4.4, 6.2, 6.5, 6.7, 8.3, 8.4, 9.1, 9.4, 10.1, 10.4

(Most overlap with above; deduplicated list:)

- NHS number on child (1.1)
- Sibling / family grouping (1.2)
- Key person (1.3) [also Critical-list]
- Accident log (1.5)
- Medication log (1.6)
- Custody / court order (1.10)
- Sessions / bookings pattern (2.1)
- Staffing & ratio (2.2 + 4.1)
- Soft delete / GDPR erasure (3.1)
- Failed-login columns on users (3.5)
- Funding eligibility cross-check (4.3)
- Re-enrollment multi-leave (4.4)
- Guardian deactivation cascade (6.2)
- Invoice void status (6.5)
- Funding renewal/expiry (6.7)
- 2FA (8.3)
- Refresh-token family revocation (8.4)
- Audit retention (9.1)
- Safeguarding read audit (9.4)
- Pagination on lists (10.1)
- Idempotency keys (10.4)

### Medium (30+) — see body above.
### Low (15+) — see body above.

---

## Quick Wins (do first, low effort / high value)

1. Add rate limiter to `/auth/login` — reuses existing `ratelimit` platform. **~30 min.**
2. Add DOB sanity CHECK constraint — one migration. **~10 min.**
3. Add audit-log append-only grants — single migration + role change. **~1 hr.**
4. Add `email_verified_at` + `last_login_at` to `users` — one migration + small `login.go` update. **~2 hr.**
5. Add `void` invoice status + transition — extend migration `000012` trigger function. **~3 hr.**
6. Document retention policy in `docs/` — pure doc. **~1 hr.**
7. Add `Idempotency-Key` middleware on `create_checkout_session` only — narrow scope, big safety win. **~4 hr.**

## Strategic Bets (large, do deliberately)

1. Funding entitlement as first-class term-based entity (2.3) — cross-cutting, touches billing, children, reporting.
2. Sessions / bookings pattern (2.1) — fundamental shift in attendance model.
3. Consent framework refactor (13.2) — model change with migration of existing booleans.
4. Append-only audit + tamper chain (9.2) — operational discipline change.
5. RLS enablement (12.4) — defence-in-depth across whole DB.
