# UK Nursery Booking Types — Gap Analysis

**Date:** 2026-07-04 · **Revised:** 2026-07-05
**Scope:** Assess current system support for UK-standard booking types, session structures, and funding models defined in the UK EYFS / DfE framework. Identify gaps.

---

## 1. Executive Summary

The system implements a **single, recurring weekly booking pattern** model (`child_booking_patterns` + `child_booking_pattern_entries` × `session_types`) and a **12-month commercial term** with monthly advance-pay invoicing built on booked minutes × hourly rate, minus a monthly funded-allowance deduction.

This core cleanly handles **Full-Time, Part-Time, Sessional (AM/PM), Ad-Hoc, Blended Funding, Term-Time-Only, and Wraparound Care** cases. It does **not** handle **pure Hourly bookings**. The `term_time_only` flag on the booking pattern drives both term-time funding allowance and term-time booking-pattern exclusion via the `term_calendar` module.

**Coverage scorecard**

| # | Booking type | Status | Notes |
|---|---|---|---|
| A | Full-Time | ✅ Supported | 4–5 day-of-week entries × full-day session type |
| B | Part-Time | ✅ Supported | 1–3 day-of-week entries |
| C | Sessional (AM/PM) | ✅ Supported | Two session types (AM, PM) per day |
| D | Ad-Hoc / Flexible | ✅ Supported | `ad_hoc_bookings` module; billing integrated via `CalculateAdHocChargeMinutes` |
| E | Funded Hours (15/30) | ✅ Supported | `CalculateTermTimeFundedAllowanceMinutes` uses academic term date ranges; stretched math also works |
| F | Holiday / Term-Time Only | ✅ Supported | `academic_terms` table + `term_time_only` on booking pattern + `branch_closure_days` for inset/bank holidays |
| G | Wraparound Care | ✅ Supported | `session_types.kind` enum; no `schools` table yet |

**Session structures**

| # | Structure | Status | Notes |
|---|---|---|---|
| 1 | Full Day | ✅ Supported | Via session type (e.g. 08:00–18:00) |
| 2 | Half Day AM | ✅ Supported | Via session type |
| 3 | Half Day PM | ✅ Supported | Via session type |
| 4 | Core Session | ✅ Supported | Via session type |
| 5 | Extended Hours add-on | ⚠️ Partial | Must be modelled as separate session type on same day |
| 6 | Hourly Booking | ❌ Missing | No pure hourly surface; advance-pay is session-minute based |

---

## 2. System As-Is

### 2.1 Booking Pattern (recurring weekly)

`api/internal/modules/children/domain/child_booking_pattern.go`

- A child has **at most one current** booking pattern (`is_current` generated from `effective_to IS NULL`).
- A pattern has N **entries**, each `(day_of_week 1–7, session_type_id)`.
- Pattern carries a `term_time_only boolean` flag — when true, occurrences outside academic term date ranges are excluded from `booked_core_minutes`.
- No quantity, no end date per entry, no "skip weeks" flag, no one-off dates.
- In-term changes produce a **new pattern** + a `term_schedule_changes` audit row; increases require manager approval.

### 2.2 Session Type

`api/internal/modules/sessiontypes/domain/entities.go` · `db/migrations/000001_baseline.up.sql:595` · `db/migrations/000005_add_session_type_kind_and_flat_fee.up.sql`

```sql
CREATE TABLE session_types (
  id, tenant_id, branch_id,
  name text,
  start_time time,
  end_time time,
  is_active boolean,
  kind text NOT NULL DEFAULT 'standard'
    CHECK (kind IN ('standard','wraparound_before','wraparound_after','core','extended')),
  flat_fee_minor integer CHECK (flat_fee_minor IS NULL OR flat_fee_minor >= 0),
  ...
  CHECK (start_time < end_time)
);
```

- **`kind` field** discriminates session categories: `standard`, `wraparound_before`, `wraparound_after`, `core`, `extended`.
- **`flat_fee_minor`** (nullable) overrides hourly × minutes billing when set — used for per-session flat pricing (e.g. wraparound morning £X, afternoon £Y).
- **No age-group / school-age flag** on session types or rooms.

### 2.3 Session Templates

`api/internal/modules/sessiontemplates/domain/entities.go`

- Same shape as booking pattern entries, reusable.
- Used as wizard presets for fast pattern creation.

### 2.4 Term

`api/internal/modules/term/domain/term.go`

- 12-month commercial commitment: `term_start_date` (must be 1st of month) → `term_end_date = start + 12 months − 1 day` (DB-enforced).
- Snapshots `site_hourly_rate_minor` at creation.
- Status lifecycle: `pre_term → active → pending_renewal → ended | terminated`.
- Only `active` terms are billable.

### 2.5 Funding

`api/internal/modules/children/domain/child_funding_record.go` (entitlement) ·
`api/internal/modules/funding/domain/profile.go` (monthly allowance) ·
`api/internal/modules/billing/domain/booked_minutes.go` (term-time calculation)

- **Child funding record** captures `FundingType ∈ {none, fifteen_hours, thirty_hours, two_year_old, custom, unknown}` and `FundingModel ∈ {term_time_only, stretched, unknown}`, plus `FundedHoursPerWeek *float64`.
- **Monthly funding profile** stores `FundedAllowanceMinutes` per (child, billing_month). Calculation per `CONCEPTS.md`:
  `funded_hours_per_week × 52 × 60 / 12`
  ⚠️ **This is stretched math** (annualised over 52 weeks / 12 months).
- **Term-time funding** is computed dynamically by `CalculateTermTimeFundedAllowanceMinutes` in `billing/domain/booked_minutes.go`:
  `funded_hours_per_week × 60 × termDaysInMonth / 5`
  where `termDaysInMonth` = count of weekdays falling within any active academic term date range in the billing month. Returns 0 in holiday months.
- The `term_time_only` flag on the **booking pattern** (not the child record) drives which funding formula is used in the billing pipeline.

### 2.6 Invoice generation (advance-pay)

`api/internal/modules/billing/domain/booked_minutes.go`

```
booked_core_minutes  = Σ (occurrences_of_dow_in_month × session_type_duration_minutes)
                     + Σ ad_hoc_session_durations (for active ad-hoc bookings in month)
funded_deduction_min = min(booked_core_minutes, funded_allowance_minutes)
billable_minutes     = max(0, booked_core_minutes − funded_allowance_minutes)
subtotal             = billable_minutes × hourly_rate
```

- Pro-ration for mid-term start: `booked_minutes × (eligible_days / total_days_in_month)`.
- Ad-hoc bookings use `CalculateAdHocChargeMinutes(duration, ad_hoc_rate_multiplier)` with the branch's `ad_hoc_rate_multiplier` (default 1.50). Flat-fee sessions bypass hourly calculation entirely.
- Invoice kinds: **`monthly`** (default) and **`adjustment`**. No `ad_hoc` / `one_off` kind.
- Line kinds: `core_childcare`, `funded_deduction`, `extra`, `ad_hoc`. Extra lines are manager-added manual charges. `ad_hoc` lines represent one-off ad-hoc session charges (distinct from recurring pattern lines).

### 2.7 Attendance & Absence

- `attendance_sessions` records check-in/out timestamps. Used for safeguarding / presence tracking.
- `absence_markers` flags a child absent on a date.
- Neither drives billing under advance-pay (booked minutes do). Attendance is no longer in the billing path post-advance-pay migration.

---

## 3. Gap Analysis — Booking Type by Booking Type

### A. Full-Time Booking — ✅ SUPPORTED

Five entries: Mon–Fri × `full_day` session type. `booked_core_minutes` correctly aggregates to ~22 days/month × session duration. No changes needed.

### B. Part-Time Booking — ✅ SUPPORTED

E.g. Mon/Wed/Fri × full-day session type. Naturally produces ~12–13 days/month. No changes needed.

### C. Sessional Booking (AM-only / PM-only) — ✅ SUPPORTED

Two session types per branch (e.g. `AM 08:00–12:00`, `PM 12:00–16:00`). Booking pattern entry points at whichever applies. A parent wanting full day selects both — the system sums them. Idiomatic UK sessional usage works directly.

### D. Ad-Hoc / Flexible Booking — ✅ SUPPORTED

**Implemented:** Full `ad_hoc_bookings` module (`api/internal/modules/ad_hoc_bookings/`).

- `ad_hoc_bookings` table keyed by `(child_id, calendar_date, session_type_id)` with `status ∈ {active, cancelled}`.
- Manager API: create, list (with child/date/status filters), cancel.
- Billing integration: `ListActiveAdHocBookingsForChildInMonth` feeds into `booked_core_minutes`. `CalculateAdHocChargeMinutes(duration, multiplier)` applies the branch's `ad_hoc_rate_multiplier` (default 1.50, column on `branches`).
- Flat-fee session types are no longer supported; all ad-hoc bookings use hourly calculation.
- Angular frontend: `AdHocBookingsApiService` + `manager-ad-hoc-booking` component.

**Remaining refinements:**
- Parent self-service ad-hoc booking is not yet available (manager-only).

### E. Funded Hours Booking (15h / 30h) — ✅ SUPPORTED

**What works:**
- Funding type, eligibility code, evidence-received flag, and `FundedHoursPerWeek` are all captured.
- Blended funding is mathematically correct: booked_min − funded_min = billable_min, multiplied by the term's snapshotted rate.
- Split usage across days works for free because funding is a minutes-pool, not a per-day allocation.

**Term-time vs stretched — now enforced:**
- `CalculateTermTimeFundedAllowanceMinutes` in `billing/domain/booked_minutes.go` computes term-time allowance dynamically:
  `funded_hours_per_week × 60 × termDaysInMonth / 5`
  where `termDaysInMonth` = weekdays falling within any active academic term in the billing month. Returns 0 in holiday months.
- The `term_time_only` flag on the **booking pattern** (not the child funding record) drives which formula is used.
- Stretched math (`funded_hours_per_week × 52 × 60 / 12`) continues to apply for non-term-time patterns.

**Remaining refinements:**
- Surface funding model on the invoice's `funded_deduction` line so parents see "Term-time funding (15h × 38 weeks)" vs "Stretched funding (≈11.1h/week)".
- No funding carry-forward between months (use-it-or-lose-it per DfE headcount submission — correct behaviour but worth documenting for nursery managers).

### F. Holiday / Term-Time Only Booking — ✅ SUPPORTED

**Implemented:** `term_calendar` module (`api/internal/modules/term_calendar/`).

- `academic_terms` table with `kind ∈ {autumn, spring, summer}`, `start_date`, `end_date`, `is_active`.
- Manager API: create, list, update, archive. Angular frontend: `manager-term-calendar` component.
- `ListActiveDateRanges` returns `[]TermDateRange` for a branch and date range, used by billing.
- Booking pattern has `term_time_only boolean` — when true, `CalculateBookedCoreMinutesInMonth` excludes dates outside academic term ranges.
- `CalculateTermTimeFundedAllowanceMinutes` uses the same term date ranges for funded-hours calculation.

**Also implemented:** `branch_closures` module (`api/internal/modules/branch_closures/`).

- `branch_closure_days` table with `(tenant_id, branch_id, date, reason)` — migration `000008`.
- Manager API: create, list, delete. `ClosureDateLookup` interface in billing domain.
- `CalculateBookedCoreMinutesInMonth` and `CalculateTermTimeFundedAllowanceMinutes` both accept closure dates and exclude them from occurrence counts. `InvoiceCalculationDetails.ClosureDaysExcluded` records which dates were excluded.

### G. Wraparound Care (before/after school) — ✅ SUPPORTED

**What works:** Full wraparound support via session type extensions:

- `session_types.kind` enum includes `wraparound_before` and `wraparound_after` values (migration 000005).
- Two session types (`07:00–09:00` kind=wraparound_before, `15:00–18:00` kind=wraparound_after) can be created and attached to booking patterns. The arithmetic treats them like any other session.

**Remaining gap:** No first-class school integration. Specifically:
- No `schools` table or `children.school_id` link to the child's school (e.g. "attends St Mary's Primary").
- No school-calendar-aware scheduling (e.g. auto-adjusting for school inset days).
- No school-pickup roster or walking-bus safeguarding records.
- No age-group / school-age flag on rooms to prevent a wraparound session type being attached to a toddler.

**Remaining work:**
- Add `children.school_id` + `schools` table for school-calendar joins.

---

## 4. Gap Analysis — Session Structures

### Full Day, Half Day AM/PM, Core Session — ✅ SUPPORTED
Achievable by configuring appropriate session types. No structural change needed.

### Extended Hours add-ons — ⚠️ PARTIAL
Works only if a parent explicitly selects both the core session type **and** the extended-hours session type on the same day. The `session_types.kind` field now includes `extended` as a discriminator, but no auto-association logic exists. The system will not auto-pair them. Risk: a parent who selects only "Core 09:00–15:00" cannot stay past 15:00 without a mid-term pattern change (which triggers the schedule-change approval workflow).

### Hourly Booking — ❌ MISSING
Booking is per session type (a fixed window). There is no surface for "I need 3 hours of care on Tuesday starting whenever." The retired attendance-driven billing model was closer to this; advance-pay deliberately moved away from it. Adding hourly booking would require either:
- A new `hourly_booking` entity with date + duration, billed at hourly rate against actual clock minutes, **or**
- A "flexible" session type with a per-instance start/end entered at booking time.

Neither exists today.

---

## 5. Gap Analysis — Funding Logic

### Stretched vs Term-Time — ✅ ENFORCED
`CalculateTermTimeFundedAllowanceMinutes` uses academic term date ranges to compute per-month allowance. The `term_time_only` flag on the booking pattern drives which formula is used. See Gap E above.

### Blended Funding — ✅ SUPPORTED
`booked_core_minutes − funded_allowance_minutes = billable_minutes`. Verified in `billing/domain/funding_deduction.go:CalculateFundingDeduction`. Clean.

### Stretched-Model Boundary Case
The 52/12 monthly average means a "30h stretched" child gets ~129.2 funded minutes/month. A part-time booking pattern that yields fewer booked minutes than this leaves unused funding on the table that month — there is **no carry-forward** mechanism between months. This is the correct DfE-aligned behaviour (funding is use-it-or-lose-it per local-authority headcount submission) but worth documenting for nursery managers.

---

## 6. Prioritised Recommendations

| Priority | Gap | Effort | Why |
|---|---|---|---|
| P0 | ~~Term calendar + term-time-only enforcement (E + F)~~ | ~~Large~~ | ✅ Done — `term_calendar` module + `CalculateTermTimeFundedAllowanceMinutes` + booking pattern `term_time_only` |
| P1 | ~~Ad-hoc booking entity (D)~~ | ~~Medium~~ | ✅ Done — `ad_hoc_bookings` module with billing integration |
| P1 | ~~Session type `kind` (G + Wraparound)~~ | ~~Medium~~ | ✅ Done — `session_types.kind` columns |
| P1 | ~~`branch_closure_days` table~~ | ~~Small~~ | ✅ Done — `branch_closures` module + billing integration via `ClosureDateLookup` |
| P2 | **Hourly booking surface** | Medium | Niche for EYFS settings; defer until P1 shipped |
| P2 | **Funding-model display on invoice lines** | Small | Cheap transparency win; helps parent disputes |
| P2 | **`schools` table + `children.school_id`** | Small | Enables school-calendar-aware wraparound scheduling |
| P3 | **Funding carry-forward policy** | Small (config) | Optional commercial lever; default to current use-it-or-lose-it |
| P3 | ~~`InvoiceLineKindAdHoc`~~ | ~~Small~~ | ✅ Done — `LineKindAdHoc = "ad_hoc"` in invoice domain + DB constraint (migration 000007) |
| P3 | **Parent self-service ad-hoc booking** | Medium | Currently manager-only |

---

## 7. Files to Reference

- Booking pattern domain: `api/internal/modules/children/domain/child_booking_pattern.go`
- Booking pattern entries SQL: `api/db/migrations/000001_baseline.up.sql:608` · `api/db/query/child_booking_pattern_entries.sql`
- Session types: `api/internal/modules/sessiontypes/domain/entities.go` · `db/migrations/000001_baseline.up.sql:595` · `db/migrations/000005_add_session_type_kind_and_flat_fee.up.sql`
- Session templates: `api/internal/modules/sessiontemplates/domain/entities.go`
- Term domain: `api/internal/modules/term/domain/term.go`
- Term calendar: `api/internal/modules/term_calendar/domain/entities.go` · `api/db/migrations/000003_add_academic_terms.up.sql`
- Funding record: `api/internal/modules/children/domain/child_funding_record.go`
- Funding profile (monthly allowance): `api/internal/modules/funding/domain/profile.go`
- Booked-minutes calculation: `api/internal/modules/billing/domain/booked_minutes.go`
- Funded deduction: `api/internal/modules/billing/domain/funding_deduction.go`
- Term-time funded allowance: `api/internal/modules/billing/domain/booked_minutes.go` (`CalculateTermTimeFundedAllowanceMinutes`)
- Ad-hoc bookings: `api/internal/modules/ad_hoc_bookings/domain/entities.go` · `api/db/migrations/000004_add_ad_hoc_bookings.up.sql`
- Ad-hoc line kind: `api/internal/modules/billing/domain/invoice.go` (`LineKindAdHoc`) · `api/db/migrations/000007_add_ad_hoc_line_kind.up.sql`
- Branch closures: `api/internal/modules/branch_closures/domain/entities.go` · `api/db/migrations/000008_add_branch_closure_days.up.sql`
- Invoice domain + line kinds: `api/internal/modules/billing/domain/invoice.go`
- Invoices schema: `db/migrations/000001_baseline.up.sql` (search `CREATE TABLE invoices`)
- Absence markers: `api/internal/modules/absence/domain/marker.go`
- Domain glossary: `CONCEPTS.md` (Booking Pattern, Term, Funded Allowance Minutes, Billable Minutes)
