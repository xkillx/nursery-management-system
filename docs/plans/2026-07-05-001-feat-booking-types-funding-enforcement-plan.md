---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
title: "feat: Booking Types and Funding Enforcement"
type: feat
depth: deep
created: 2026-07-05
---

# Booking Types and Funding Enforcement

## Execution Progress

Committed on `feat/booking-types-funding-enforcement` as `786f19d`.

| U-ID | Unit | Status |
|------|------|--------|
| U1 | Academic Term Calendar | ✅ Go backend complete ✅ Angular frontend |
| U2 | Booking Pattern Term-Time Flag | ✅ Go backend complete ✅ Angular frontend (model updated) |
| U3 | Term-Time Billing Calculation | ✅ Domain signature + math complete ✅ TermDateLookup adapter wired into `generate_term_invoices.go` |
| U4 | Term-Time Funding Allowance | ✅ Domain function added ✅ Pipeline wiring + funding_model label on invoice details |
| U5 | Ad-Hoc Booking Module | ✅ Go backend complete ✅ Angular frontend |
| U6 | Ad-Hoc Invoice Integration | ✅ `LineKindAdHoc` constant + `CalculateAdHocChargeMinutes` ✅ Invoice line generation in `generate_term_invoices.go` ✅ CONCEPTS.md update |
| U7 | Session Type Kind + Flat Fee | ✅ Go backend complete ✅ Flat-fee branching in billing pipeline ✅ Angular frontend |

**Backend gates passed:** `go fmt`, `go vet`, `go build ./...` — zero warnings.
**Frontend gates passed:** `ng build` — zero errors, zero warnings.

**Remaining backend work:**
- Implement `termDateLookupAdapter` in bootstrap (consuming `term_calendar` repo)
- Inject it into `generate_term_invoices.go`; when `termRow.TermTimeOnly`, call adapter and pass result into `CalculateBookedCoreMinutesInMonth`
- Use `CalculateTermTimeFundedAllowanceMinutes` when `funding_model = term_time_only` and no explicit monthly profile exists
- Query `AdHocBookingLookup`, produce `ad_hoc` invoice lines per booking (flat fee vs multiplier branch), `SortOrder ≥ 3`
- Add `funding_model` + term-dates-explainability keys to `calculation_details` and `funded_deduction` line JSON

## Summary

Implements P0 and P1 gaps from `docs/reports/BOOKING-TYPES-GAP-ANALYSIS.md`: term calendar with term-time-only enforcement, ad-hoc booking entity, session type `kind` classification and flat-fee pricing, ad-hoc line display on invoices. P2/P3 (hourly booking, carry-forward, funding model invoice display) deferred to follow-up. Covers Go backend (domain, application, infrastructure, HTTP), Angular frontend, and billing pipeline changes. Product Contract unchanged.

---

## Goal Capsule

UK nurseries need academic term calendars, term-time-only contracts, one-off ad-hoc bookings, session type categorisation (standard/wraparound/extended), per-session-type flat pricing, and term-time-aware funding allowance computation. This plan enables those features while preserving the existing advance-pay 52-week billing for standard (non-term-time) patterns.

---

## Problem Frame

The system currently models a single recurring weekly booking pattern and always bills 52 weeks per year (ADR-0006, DECISION-BASELINE). This covers full-time, part-time, and sessional (AM/PM) patterns. However:

1. **No academic term calendar.** Term-time-only nurseries (the majority of wraparound/sessional providers) bill across 52 weeks when they should only bill during school terms (~38 weeks).
2. **Term-time-only funding flag stored, not enforced.** `child_funding_records.funding_model` captures `term_time_only` but the billing pipeline always uses stretched math (`funded_hours_per_week × 52 × 60 / 12`). A 15h term-time-only entitlement should be ~22.8h/month during term months and 0h during holidays.
3. **No ad-hoc booking surface.** One-off sessions (backup childcare, inset-day extras) cannot be booked or invoiced through the system. Nurseries use off-system workarounds or distort booking patterns.
4. **No session type classification.** Wraparound care (07:00–09:00, 15:00–18:00) is priced hourly like all other sessions. UK nurseries typically price wraparound as flat per-session fees (£X per morning, £Y per afternoon). There is no `kind` discriminator or flat-fee column.
5. **No ad-hoc invoice line.** Invoices cannot distinguish recurring charges from one-off bookings.

Gap analysis priority: term calendar + term-time enforcement (P0), ad-hoc booking + session type kind/flat fee (P1). Hourly booking (P2) and carry-forward (P3) deferred.

---

## Requirements Contract

| R-ID | Requirement | Source |
|------|-------------|--------|
| R1 | Academic term calendar per branch with start/end dates and kind (autumn, spring, summer). | Gap F |
| R2 | Booking pattern optional `term_time_only` flag. When true, billed minutes exclude non-term dates. | Gap F |
| R3 | Term-time-only funding model: allowance computed from term calendar during term months, zero during holidays. Stretched model unchanged. | Gap E |
| R4 | Ad-hoc booking entity: child + calendar date + session type. Billed at per-session rate. | Gap D |
| R5 | Session type `kind` enum (standard, wraparound_before, wraparound_after, core, extended). | Gap G |
| R6 | Optional flat fee per session type. When set, billed as flat amount instead of hourly × minutes. | Gap G |
| R7 | Invoice lines distinguish recurring vs ad-hoc bookings. Funding model labeled on funded deduction lines. | Gap E |
| R8 | All existing full-time, part-time, sessional patterns continue to bill 52 weeks unchanged. | ADR-0006 |

### Success Criteria

- Term-time-only child invoices zero in a non-term month.
- Non-term-time patterns bill identically to current behavior (52-week rule).
- Ad-hoc bookings appear as distinct invoice lines with configurable pricing.
- Session types with flat fee produce flat invoice line amounts regardless of duration.
- Manager can create, edit, and archive academic terms and session type kind/flat fee through the UI.

---

## Key Technical Decisions

### KTD-1: Term-time-only flag on booking pattern, not child record

**Decision:** `term_time_only boolean` column on `child_booking_patterns`, default false.

**Rationale:** Term-time-only describes how a booking pattern is billed, not an intrinsic child attribute. A child could have a term-time-only pattern for one period and a year-round pattern for a different period (e.g., after moving schools). Placing the flag on the pattern keeps billing logic localized to the pattern-reading code path. The gap analysis report explicitly proposes this location (`docs/reports/BOOKING-TYPES-GAP-ANALYSIS.md:171`).

**Alternative considered:** Flag on `children` table. Rejected — conflates child identity with billing contract. If a child changes from term-time to year-round, the child record would need historical tracking, duplicating what `child_booking_patterns` already provides via pattern versioning.

### KTD-2: Academic terms at branch level, owned by manager

**Decision:** `academic_terms` table scoped to `(tenant_id, branch_id)`. CRUD managed by manager/owner via HTTP API.

**Rationale:** Different branches in a group may serve different catchment areas with different school calendars. Placing terms at the branch level matches the existing session types scope (`session_types` is per-branch). Branch managers already configure session types and rates; academic terms are the same category of operational configuration.

### KTD-3: Term-time funding uses term calendar dates, not manual override

**Decision:** When `child_funding_records.funding_model = term_time_only` AND the branch has active academic terms, the system auto-computes `funded_allowance_minutes` per month during invoice generation:

```
term_days_in_month = count of dates within academic terms falling in billing month
total_days_in_month = days in billing month
funded_allowance = funded_hours_per_week × term_days_in_month × 60 / 5
```

When funding_model is `stretched` or no academic terms exist, current formula applies.

**Rationale:** ADR-0011 (registration funding is guidance-only) deliberately decouples `child_funding_records` from `funding_profiles`. This plan narrows that rule: for term-time-only children, the term calendar provides an authoritative computation source. Managers can still override via the existing `upsert_profile` endpoint. The auto-computation runs during invoice generation as a fallback when no explicit monthly profile exists, not as an automatic upsert into `funding_profiles`. This preserves ADR-0011's manager-control intent while reducing manual work for the common case.

**Alternative considered:** Fully automatic upsert into `funding_profiles` from registration. Rejected — violates ADR-0011 deliberately. Managers should see and approve funding numbers before they appear on invoices.

### KTD-4: Ad-hoc bookings as separate entity, not pattern date extension

**Decision:** New `ad_hoc_bookings` table keyed by `(child_id, calendar_date, session_type_id)`. Independent from `child_booking_pattern_entries`.

**Rationale:** Booking pattern entries are `(day_of_week, session_type_id)` — weekly recurring with no date field. Adding a `calendar_date` column to pattern entries would blur the semantic distinction between recurring and one-off bookings and complicate the "one current pattern" invariant. A separate entity keeps the recurring-pattern model clean and allows ad-hoc pricing to diverge (ad-hoc multiplier).

### KTD-5: Ad-hoc pricing via branch-level multiplier

**Decision:** `branches.ad_hoc_rate_multiplier numeric(4,2) DEFAULT 1.50` applied to the site hourly rate × session duration for ad-hoc bookings. No per-session-type flat fee for ad-hoc specifically.

**Rationale:** Ad-hoc bookings are premium (backup care, last-minute). A branch-level multiplier keeps configuration simple (one number per branch) and scales with whatever session type is selected. Session types get their own flat fee via KTD-6 for structured wraparound pricing. The two pricing mechanisms are orthogonal: ad-hoc multiplier for one-off demand pricing, flat fee for fixed-fee session categories.

### KTD-6: Session type `kind` enum + nullable flat fee

**Decision:** Add `kind text` and `flat_fee_minor int` (nullable) columns to `session_types`. Kind values: `standard`, `wraparound_before`, `wraparound_after`, `core`, `extended`.

**Rationale:** `kind` enables UI filtering and billing differentiation. Flat fee per session type replaces hourly × minutes calculation when set. The default `standard` with null `flat_fee_minor` preserves existing behavior for all current session types. This is backward-compatible: no migration of existing data needed beyond column additions with defaults.

### KTD-7: Invoice line kind `ad_hoc` for one-off bookings

**Decision:** Add `ad_hoc` to `invoice_lines.line_kind` enum. Ad-hoc bookings produce separate lines from `core_childcare` recurring lines.

**Rationale:** Gap D notes the lack of distinction between recurring and one-off charges. A dedicated line kind enables clear parent-facing invoices and supports different pricing. The existing constraint (`funded_deduction line → amount <= 0; core_childcare → amount >= 0`) does not restrict `ad_hoc`, so it accepts any positive amount naturally.

---

## High-Level Technical Design

### New Billing Pipeline Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Invoice Generation (monthly)                  │
└───────────────────────┬─────────────────────────────────────────┘
                        │
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
   ┌─────────────┐ ┌─────────┐ ┌──────────────┐
   │ Booking     │ │ Ad-Hoc  │ │ Academic     │
   │ Pattern     │ │ Bookings│ │ Terms        │
   │ Entries     │ │         │ │ (new)        │
   └──────┬──────┘ └────┬────┘ └──────┬───────┘
          │             │             │
          │ term_time   │             │ term
          │ _only? ─────┼── filter ───┘ dates
          │             │             │
          ▼             ▼             │
   ┌─────────────────────────────┐    │
   │ CalculateBookedCoreMinutes  │◀───┘
   │ InMonth (modified)          │
   │  - recurring: filtered by   │
   │    term dates if TTO        │
   │  - ad-hoc: summed per date  │
   └──────────────┬──────────────┘
                  │
                  ▼
   ┌─────────────────────────────┐
   │ ComputeFundedDeduction      │
   │ (modified: term-time math)  │
   │  - TTO: term-based formula  │
   │  - Stretched: 52/12 formula │
   └──────────────┬──────────────┘
                  │
                  ▼
   ┌─────────────────────────────┐
   │ CreateInvoiceLines          │
   │  - core_childcare (recur)   │
   │  - ad_hoc (one-off)         │
   │  - funded_deduction         │
   └─────────────────────────────┘
```

### Session Type Billing Branch

```
SessionType.kind
    │
    ├── standard ─┬── flat_fee_minor set?
    │             │   ├── YES → charge flat_fee_minor
    │             │   └── NO  → charge hourly × minutes (current)
    │             │
    ├── wraparound_before/after
    │             └── flat_fee_minor REQUIRED → charge flat fee
    │
    ├── core ─┬── flat_fee_minor set?
    │         │   ├── YES → charge flat_fee_minor
    │         │   └── NO  → charge hourly × minutes
    │
    └── extended ─┴── same as core
```

### Academic Term Entity Model

```
AcademicTerm
  id            uuid
  tenant_id     uuid
  branch_id     uuid
  name          text        "Autumn 2026"
  kind          text        autumn | spring | summer
  start_date    date
  end_date      date
  is_active     boolean

  Constraints:
  - start_date < end_date
  - UNIQUE(tenant_id, branch_id, name) WHERE is_active
  - Overlapping date ranges permitted (e.g. "half-term" within "Autumn 2026")
```

---

## Scope Boundaries

### In Scope

- Academic term calendar: CRUD, validation, management UI.
- Term-time-only flag on booking patterns: schema, domain, API, UI toggle.
- Term-time-aware booking minutes calculation in billing pipeline.
- Term-time-aware funding allowance computation during invoice generation.
- Ad-hoc booking entity: CRUD, validation, API, manager UI.
- Ad-hoc invoice integration: new line kind, branch multiplier pricing.
- Session type `kind` enum and `flat_fee_minor`: schema, validation, API, UI.
- Flat fee billing calculation in invoice generation.
- Invoice line display of funding model on funded deduction lines.

### Out of Scope

- Hourly booking surface (P2). Conflicts architecturally with advance-pay model. Needs separate design pass.
- Funding carry-forward between months (P3). Optional commercial lever; default use-it-or-lose-it continues.
- School pickup/dropoff logistics, school-age group flags on rooms, school-calendar join via `schools` table. Gap G mentions these; they are product-level features beyond billing primitives.
- Auto-creation of academic terms from government data. Manual entry by managers.
- Adjustment of existing invoices when term dates change mid-cycle. Invoice generation runs prospectively; historical invoices are immutable (DB trigger).
- 12-month term duration flexibility (noted in invoice analysis, separate concern).

### Deferred to Follow-Up Work

- **Hourly booking surface (P2):** New `hourly_bookings` entity with date + duration, or flexible session type with per-instance start/end. Requires separate planning due to architectural conflict with advance-pay.
- **Funding carry-forward (P3):** `funding_carry_forward` table tracking unused funded minutes per month. Optional commercial lever — nurseries may or may not want it.
- **Advanced term management:** Half-term auto-split, bank holiday auto-exclusion, import from local authority calendars.
- **School-age wraparound flow:** Linked school record on child, school-calendar-aware scheduling, roster management. Gap G scope beyond billing primitives.

---

## Open Questions

### Deferred to Implementation

- **Q1:** When academic term dates change after invoices have already been generated for future months, should the system regenerate affected draft invoices automatically, or require manual intervention? Default to manual (draft regeneration already exists as a separate action).
- **Q2:** Ad-hoc booking conflicts: should the system prevent an ad-hoc booking for a child who already has a recurring booking pattern entry on that exact date + session type, or allow both (with the ad-hoc booking treated as an extra)? Default: allow — the ad-hoc is a deliberate extra booking.
- **Q3:** Session type flat fee: should there be a minimum duration validation (e.g., flat fee only allowed on session types ≥ 2 hours) to prevent accidental misuse? Deferred — no constraint initially; manager discretion.

---

## Alternative Approaches Considered

### Term-time enforcement: system-wide calendar vs branch-level terms

**Chosen:** Branch-level academic terms (KTD-2).

**Alternative:** System-wide academic term calendar shared across all branches. Rejected because different branches in a multi-site group serve different school catchment areas with different term dates. Branch-level aligns with existing session types scope.

### Term-time flag on child vs booking pattern

**Chosen:** On `child_booking_patterns` (KTD-1).

**Alternative:** Flag on `children` table (`attendance_type` column). Rejected because `children` lacks temporal versioning of this attribute. A child transitioning from term-time (age 3, preschool) to year-round (age 4, school-age wraparound) would require manual cleanup. Placing the flag on the pattern (which is already versioned with effective_from/effective_to) preserves per-period billing mode naturally.

### Funding allowance: auto-upsert vs inline computation

**Chosen:** Inline computation during invoice generation (KTD-3).

**Alternative:** Auto-upsert funding_profiles from term calendar when funding_model = term_time_only. This would keep funding_profiles as the single source. Rejected because it requires a separate background job or event to keep profiles in sync with term date changes, violating ADR-0011's manager-control principle. The inline fallback (use explicit profile; if none, compute from term calendar) is simpler and keeps manager overrides authoritative.

### Ad-hoc booking: extension of pattern entries vs separate entity

**Chosen:** Separate `ad_hoc_bookings` entity (KTD-4).

**Alternative:** Add optional `calendar_date` column to `child_booking_pattern_entries`. Rejected because it would break the unique constraint `(pattern_id, day_of_week)` — two ad-hoc bookings on the same Monday would collide. The constraint could be relaxed, but that would allow duplicate recurring entries accidentally. A separate entity is cleaner and allows ad-hoc-specific pricing logic.

### Flat fee: per-session-type vs per-branch pricing table

**Chosen:** `flat_fee_minor` column on `session_types` (KTD-6).

**Alternative:** A separate `session_type_pricing` table supporting multiple pricing models (hourly, flat, per-session, per-term). Rejected as over-engineering for the current use case. The single nullable column handles the binary case (hourly or flat). A pricing table could be introduced later when pricing models proliferate.

---

## Implementation Units

### Phase 1: Term Calendar Foundation

#### U1. Academic Term Calendar Module

**Goal:** New `term_calendar` module providing academic term CRUD per branch. Foundation for term-time billing.

**Status:** ✅ Go backend shipped (migrations `000003_*`, domain, app 4 use cases, postgres repo, HTTP handler, bootstrap wiring). ⬜ Angular frontend (`academic-terms-api.service.ts`, models, `manager-term-calendar` page). ⬜ Unit + integration tests.

**Requirements:** R1

**Dependencies:** None

**Files:**
- `api/internal/modules/term_calendar/domain/entities.go`
- `api/internal/modules/term_calendar/domain/errors.go`
- `api/internal/modules/term_calendar/domain/repository.go`
- `api/internal/modules/term_calendar/application/create_term.go`
- `api/internal/modules/term_calendar/application/list_terms.go`
- `api/internal/modules/term_calendar/application/update_term.go`
- `api/internal/modules/term_calendar/application/delete_term.go`
- `api/internal/modules/term_calendar/infrastructure/postgres/repository.go`
- `api/internal/modules/term_calendar/infrastructure/postgres/dto.go`
- `api/internal/modules/term_calendar/interfaces/http/handler.go`
- `api/internal/modules/term_calendar/interfaces/http/dto.go`
- `api/db/migrations/00000X_add_academic_terms.up.sql`
- `api/db/migrations/00000X_add_academic_terms.down.sql`
- `api/db/query/academic_terms.sql`
- `api/internal/app/bootstrap/bootstrap.go`
- `api/internal/app/bootstrap/adapters.go`
- `api/internal/modules/term_calendar/application/create_term_test.go`
- `api/internal/modules/term_calendar/application/list_terms_test.go`
- `api/internal/modules/term_calendar/infrastructure/postgres/repository_test.go`
- `web/src/app/features/staff/data/academic-terms-api.service.ts`
- `web/src/app/features/staff/models/academic-term.models.ts`
- `web/src/app/features/staff/pages/manager-term-calendar/manager-term-calendar.component.ts`
- `web/src/app/features/staff/pages/manager-term-calendar/manager-term-calendar.component.html`

**Approach:**
- Create `term_calendar` module following standard layer pattern: domain → application → postgres → http.
- `AcademicTerm` entity with: ID, TenantID, BranchID, Name, Kind, StartDate, EndDate, IsActive, timestamps.
- Kind enum: `autumn`, `spring`, `summer` (stored as text, validated in application layer).
- Domain validation: start_date < end_date, name required, kind required.
- Application layer: CreateTerm, ListTerms (active + optionally archived), UpdateTerm, DeleteTerm (soft deactivate).
- Repository interface in domain, implementation in `infrastructure/postgres/` using sqlc.
- HTTP routes (manager + owner only):
  - `GET /sites/:site_id/academic-terms` — list
  - `POST /sites/:site_id/academic-terms` — create
  - `PATCH /sites/:site_id/academic-terms/:id` — update
  - `DELETE /sites/:site_id/academic-terms/:id` — deactivate
- Wire adapters in `bootstrap/adapters.go`: term lookup adapter for billing module consumption.
- Angular: service with typed models, management page under `/manager/site-settings/term-calendar`.

**Patterns to follow:**
- Module structure mirrors `sessiontypes/` module exactly.
- Repository pattern: flat interface in domain, sqlc DTOs in infra layer.
- HTTP handler pattern: Gin + actor-based role checks (manager, owner).
- `type Tx = any` in domain repository interface. No infrastructure imports.
- Angular service: snake_case models matching Go JSON tags per `booking-pattern-api-snake-case-mismatch` learning.

**Test scenarios:**
- Happy path: create term with valid dates and kind → persisted correctly.
- Happy path: list terms returns only active terms by default.
- Happy path: update term name → changes reflected.
- Happy path: delete term → is_active becomes false.
- Edge case: overlapping date ranges between terms permitted (half-term within autumn).
- Edge case: list across multiple academic years returns correct results.
- Error: start_date >= end_date → domain error rejected.
- Error: empty name → domain error rejected.
- Error: invalid kind value → domain error rejected.
- Error: duplicate name at same site where both active → rejected.
- Integration: repository create then list round-trips correctly against real Postgres.

**Verification:** `go build ./...` and `go vet ./...` pass in `api/`. `npm run lint` and `ng build` pass in `web/`. New API endpoints return 201 on create, 200 on list, 422 on invalid input. Manager can create and list terms via UI.

---

### Phase 2: Term-Time Only Enforcement

#### U2. Booking Pattern Term-Time Flag

**Goal:** Add optional `term_time_only boolean` to `child_booking_patterns` and propagate through the booking pattern lifecycle.

**Requirements:** R2, R8

**Dependencies:** U1

**Files:**
- `api/db/migrations/00000X_add_term_time_only_flag.up.sql`
- `api/db/migrations/00000X_add_term_time_only_flag.down.sql`
- `api/internal/modules/children/domain/child_booking_pattern.go`
- `api/db/query/child_booking_patterns.sql`
- `api/internal/modules/children/infrastructure/postgres/booking_pattern_repository.go`
- `api/internal/modules/children/application/booking_patterns.go`
- `api/internal/modules/children/application/booking_pattern_helpers.go`
- `api/internal/modules/children/interfaces/http/booking_pattern_dto.go`
- `web/src/app/features/staff/models/booking-pattern.models.ts`
- `web/src/app/features/staff/pages/manager-booking-pattern/manager-booking-pattern.component.ts`
- `web/src/app/features/staff/pages/manager-booking-pattern/manager-booking-pattern.component.html`
- `api/internal/modules/children/application/booking_patterns_test.go`

**Approach:**
- Migration: `ALTER TABLE child_booking_patterns ADD COLUMN term_time_only boolean NOT NULL DEFAULT false`. Safe on existing rows — all default to year-round.
- Domain: add `TermTimeOnly bool` to `BookingPattern` struct. No domain validation needed (boolean is always valid).
- sqlc: update `ChildBookingPatternsListByChild`, `GetByID`, `GetCurrentOpenByChild`, `Insert` to include the new column.
- Application: `CreateBookingPattern` and `UpdateBookingPattern` accept `term_time_only` in input params, pass through to repository.
- HTTP DTO: add `term_time_only` to request/response DTOs.
- Angular: checkbox toggle on booking pattern creation/edit form. Model uses `term_time_only: boolean`.
- Existing patterns: default `false` — no behavior change.

**Patterns to follow:**
- Migration pattern: non-nullable with default for safe roll-forward.
- DTO mapping: json tag `term_time_only` matching Angular snake_case.

**Test scenarios:**
- Happy path: create booking pattern with `term_time_only = true` → persisted and returned correctly.
- Happy path: create pattern without specifying flag → defaults to `false`.
- Happy path: update existing pattern to set `term_time_only = true` → flag updated.
- Edge case: multiple patterns for same child, some term-time, some year-round → each retains own flag.
- Backward compatibility: existing patterns loaded with new code → `term_time_only` reads as `false`.

**Verification:** API accepts `term_time_only` in booking pattern create/update requests. Existing patterns return `term_time_only: false`. Angular checkbox persists correctly. `go test ./internal/modules/children/application/...` passes.

#### U3. Term-Time Billing Calculation

**Goal:** Modify `CalculateBookedCoreMinutesInMonth` to filter non-term dates when `term_time_only` is true.

**Requirements:** R2, R8

**Dependencies:** U1, U2

**Files:**
- `api/internal/modules/billing/domain/booked_minutes.go`
- `api/internal/modules/billing/application/generate_term_invoices.go`
- `api/internal/modules/billing/infrastructure/postgres/repository.go`
- `api/db/query/invoices.sql`
- `api/internal/modules/term_calendar/domain/repository.go` (lookup adapter)
- `api/internal/app/bootstrap/adapters.go`
- `api/internal/modules/billing/domain/booked_minutes_test.go`
- `api/internal/modules/billing/application/generate_term_invoices_test.go`

**Approach:**
- Modify `CalculateBookedCoreMinutesInMonth` signature to accept optional `termDates []TermDateRange` parameter (or `nil` for year-round).
  - `TermDateRange` struct: `{StartDate, EndDate time.Time}` — provided via adapter from `term_calendar` module.
- When `termDates` is nil or empty: behavior identical to current (count all day-of-week occurrences in month).
- When `termDates` is populated: during the day-iteration loop, skip dates that do NOT fall within any term date range.
- `generate_term_invoices.go`: after loading booking pattern entries, check `TermTimeOnly` flag. If true, call term calendar adapter to get active term date ranges for the billing month's year. Pass ranges as `termDates`.
- Add new `TermDateLookup` interface in `billing/domain/`: `GetTermDateRangesForBranchAndMonth(ctx, tenantID, branchID, month time.Time) ([]TermDateRange, error)`. Wiring via bootstrap adapter consuming `term_calendar` module.
- Invoice `calculation_details` JSON: add `term_time_only: true` and `term_dates_used: [...]` for explainability.
- ADR-0006 preservation: for non-term-time patterns (the current default), the calculation is identical to existing behavior. The 52-week rule holds unless explicitly opted into term-time-only.

**Patterns to follow:**
- Cross-module dependency: interface defined in `billing/domain/`, implemented as adapter in bootstrap consuming `term_calendar` module's list use case.
- No changes to `BookedPatternEntry` or `BookedSession` structs — filtering happens inside the calculation function.
- Date comparison logic: `!date.After(start_date) && !date.Before(end_date)` for inclusive range check.

**Test scenarios:**
- Covers Gap F.
- Happy path: `term_time_only = false`, 5 Mondays → all 5 counted (existing behavior preserved).
- Happy path: `term_time_only = true`, 5 Mondays, autumn term covers 3 Mondays → only 3 counted.
- Happy path: `term_time_only = true`, billing month entirely outside any term → 0 minutes.
- Happy path: `term_time_only = true`, multiple terms overlapping month (e.g., autumn ends Oct 25, spring starts Oct 26) → dates from both terms counted.
- Edge case: term_dates is nil (no terms configured for branch) → falls back to counting all dates.
- Edge case: term range partially overlaps month → only dates within range counted.
- Backward compatibility: calling with nil termDates returns same result as current implementation.
- Integration: end-to-end generation for term-time-only child produces correct booked minutes.

**Verification:** Unit tests pass for all term filtering scenarios. Generate invoices for a term-time-only pattern and non-term-time pattern in the same month; billed minutes differ correctly. `go test ./internal/modules/billing/...` passes.

#### U4. Term-Time Funding Allowance

**Goal:** Compute `funded_allowance_minutes` from term calendar during invoice generation when funding model is `term_time_only`.

**Requirements:** R3, R8

**Dependencies:** U1, U3

**Files:**
- `api/internal/modules/billing/domain/booked_minutes.go`
- `api/internal/modules/billing/application/generate_term_invoices.go`
- `api/internal/modules/children/domain/child_funding_record.go`
- `api/internal/modules/funding/domain/profile.go`
- `api/internal/modules/term_calendar/domain/repository.go` (adapter)
- `CONCEPTS.md`
- `api/internal/modules/billing/domain/booked_minutes_test.go`

**Approach:**
- Add `CalculateTermTimeFundedAllowanceMinutes(fundedHoursPerWeek float64, termDateRanges []TermDateRange, billingMonth time.Time) int` to `billing/domain/booked_minutes.go`:
  ```
  termDaysInMonth = count dates within any term date range that fall in billing month
  totalWorkDaysInMonth = count of weekdays (Mon-Fri) in billing month
  if termDaysInMonth == 0: return 0
  allowance = fundedHoursPerWeek × 60 × termDaysInMonth / 5
  return int(allowance)
  ```
- This formula allocates weekly funded hours proportionally to actual term days in the month. A 15h/week entitlement in a month with 15 term weekdays out of 21 total weekdays gets `15 × 60 × 15 / 5 = 2700 minutes`.
- In `generate_term_invoices.go`: after determining `funded_allowance_minutes` from the funding profile (if explicit), check if child funding record has `funding_model = term_time_only`. If so AND no explicit profile exists for this month, compute allowance via term formula instead of stretched formula.
- **Stretched model unchanged:** when `funding_model = stretched` (or `unknown`), continue using `funded_hours_per_week × 52 × 60 / 12` stored in `funding_profiles`.
- Add `funding_model` to invoice line details JSON on the `funded_deduction` line for display (R7).
- Update CONCEPTS.md: clarify Funded Allowance Minutes has two computation paths (term-time vs stretched).

**Patterns to follow:**
- No mutation of `funding_profiles` — computation is inline during generation, consistent with ADR-0011's guidance-only principle.
- Cross-module funding record lookup: interface in billing domain, adapter in bootstrap reading `children` module's funding record repository.

**Test scenarios:**
- Covers Gap E.
- Happy path: 15h TTO entitlement, 15 term weekdays in month → 2700 minutes.
- Happy path: 30h TTO entitlement, 0 term days in month (summer holiday) → 0 minutes.
- Happy path: stretched funding → uses existing formula, term calendar ignored.
- Edge case: funding model is `unknown` → falls back to stretched formula.
- Edge case: explicit funding profile exists for month → use explicit value regardless of funding model.
- Edge case: funded hours per week is 0 → 0 minutes regardless of term dates.
- Domain test: `TestCalculateTermTimeFundedAllowanceMinutes` with various month/term combinations.

**Verification:** Funding deduction lines show correct amounts for TTO vs stretched children. Non-term months produce zero funded allowance for TTO children. `go test ./internal/modules/billing/domain/...` passes.

---

### Phase 3: Ad-Hoc Bookings

#### U5. Ad-Hoc Booking Module

**Goal:** New `ad_hoc_bookings` module for one-off session bookings per child per date.

**Requirements:** R4

**Dependencies:** None (parallel with Phase 2 possible)

**Files:**
- `api/internal/modules/ad_hoc_bookings/domain/entities.go`
- `api/internal/modules/ad_hoc_bookings/domain/errors.go`
- `api/internal/modules/ad_hoc_bookings/domain/repository.go`
- `api/internal/modules/ad_hoc_bookings/application/create_booking.go`
- `api/internal/modules/ad_hoc_bookings/application/list_bookings.go`
- `api/internal/modules/ad_hoc_bookings/application/cancel_booking.go`
- `api/internal/modules/ad_hoc_bookings/infrastructure/postgres/repository.go`
- `api/internal/modules/ad_hoc_bookings/infrastructure/postgres/dto.go`
- `api/internal/modules/ad_hoc_bookings/interfaces/http/handler.go`
- `api/internal/modules/ad_hoc_bookings/interfaces/http/dto.go`
- `api/db/migrations/00000X_add_ad_hoc_bookings.up.sql`
- `api/db/migrations/00000X_add_ad_hoc_bookings.down.sql`
- `api/db/query/ad_hoc_bookings.sql`
- `api/internal/app/bootstrap/bootstrap.go`
- `api/internal/app/bootstrap/adapters.go`
- `api/internal/modules/ad_hoc_bookings/application/create_booking_test.go`
- `api/internal/modules/ad_hoc_bookings/application/list_bookings_test.go`
- `api/internal/modules/ad_hoc_bookings/infrastructure/postgres/repository_test.go`
- `api/internal/modules/ad_hoc_bookings/interfaces/http/handler_test.go`
- `web/src/app/features/staff/data/ad-hoc-bookings-api.service.ts`
- `web/src/app/features/staff/models/ad-hoc-booking.models.ts`
- `web/src/app/features/staff/pages/manager-ad-hoc-booking/manager-ad-hoc-booking.component.ts`
- `web/src/app/features/staff/pages/manager-ad-hoc-booking/manager-ad-hoc-booking.component.html`
- `api/db/migrations/00000X_add_ad_hoc_multiplier.up.sql`
- `api/db/migrations/00000X_add_ad_hoc_multiplier.down.sql`

**Approach:**
- `AdHocBooking` entity: ID, TenantID, BranchID, ChildID, CalendarDate, SessionTypeID, BookedByMembershipID, Status (`active`, `cancelled`), timestamps.
- Migration: `ad_hoc_bookings` table with FKs to `children`, `session_types`. Index on `(tenant_id, branch_id, child_id, calendar_date)`.
- Migration: `branches` table add `ad_hoc_rate_multiplier numeric(4,2) DEFAULT 1.50`.
- Domain validation: calendar_date must be future or today, session_type must be active, child must be active.
- Application: `CreateAdHocBooking`, `ListAdHocBookings` (by child + date range), `CancelAdHocBooking`.
- Pricing is NOT computed at booking time — computed during invoice generation (U6), which reads the branch multiplier at that point.
- HTTP routes (manager + owner):
  - `GET /sites/:site_id/ad-hoc-bookings?child_id=&from=&to=` — list
  - `POST /sites/:site_id/ad-hoc-bookings` — create
  - `POST /sites/:site_id/ad-hoc-bookings/:id/cancel` — cancel
- Angular: simple form to select child, date, session type. Displays upcoming ad-hoc bookings with cancel option.

**Patterns to follow:**
- Module structure mirrors `sessiontypes/` module.
- Status uses text column with domain constants (not DB enum), matching `invoice.go` pattern.
- Cancel is a state transition, not a DELETE — preserves audit trail.
- `type Tx = any` in domain repository interface.

**Test scenarios:**
- Happy path: create ad-hoc booking for future date → persisted with `active` status.
- Happy path: list bookings for child in date range → returns matching bookings.
- Happy path: cancel booking → status becomes `cancelled`.
- Edge case: create booking for date where child already has recurring pattern entry → permitted (treated as extra).
- Edge case: create booking with past date → domain error rejected.
- Edge case: create booking with inactive session type → domain error rejected.
- Error: child not found → domain error rejected.
- Integration: repository create + list round-trip against real Postgres.

**Verification:** CRUD API functional. Angular form creates and displays bookings. `go test ./internal/modules/ad_hoc_bookings/...` passes.

#### U6. Ad-Hoc Booking Invoice Integration

**Goal:** Include ad-hoc bookings in invoice generation as separate `ad_hoc` line items with multiplied pricing.

**Requirements:** R4, R7

**Dependencies:** U3, U5

**Files:**
- `api/internal/modules/billing/domain/invoice.go`
- `api/internal/modules/billing/domain/booked_minutes.go`
- `api/internal/modules/billing/application/generate_term_invoices.go`
- `api/internal/modules/ad_hoc_bookings/domain/repository.go` (lookup adapter)
- `api/db/query/invoices.sql`
- `api/db/migrations/00000X_add_ad_hoc_line_kind.up.sql`
- `api/db/migrations/00000X_add_ad_hoc_line_kind.down.sql`
- `web/src/app/features/staff/data/manager-invoices-api.service.ts`
- `web/src/app/shared/components/ecommerce/billing/billing-invoice-table/billing-invoice-table.component.ts`
- `web/src/app/shared/components/ecommerce/billing/billing-invoice-table/billing-invoice-table.component.html`
- `CONCEPTS.md`
- `api/internal/modules/billing/application/generate_term_invoices_test.go`
- `api/internal/modules/billing/domain/booked_minutes_test.go`

**Approach:**
- Migration: add `ad_hoc` to `invoice_lines.line_kind` CHECK constraint.
- Add `LineKindAdHoc = "ad_hoc"` constant to `billing/domain/invoice.go`.
- Add `CalculateAdHocChargeMinutes(sessionDurationMinutes int, adHocRateMultiplier float64) int` to `billing/domain/booked_minutes.go`. Returns `ceil(sessionDuration × multiplier)`. For a 240-min session with 1.5x multiplier → 360 chargeable minutes.
- Add `CalculateAdHocFlatAmount(sessionType, branchRateMinor, adHocRateMultiplier) Money` for when ad-hoc uses a flat fee (alternative to multiplier). Defers to session type's `flat_fee_minor` if set, otherwise computes `hourly × charged_minutes`.
- In `generate_term_invoices.go`: after computing recurring booked minutes, query `AdHocBookingLookup` for active ad-hoc bookings in the billing month. For each booking:
  - If the session type has `flat_fee_minor` set → use flat fee directly.
  - Otherwise → compute via multiplier: `session_duration × ad_hoc_rate_multiplier × branch_hourly_rate`.
  - Create invoice line with `line_kind = ad_hoc`. Description: "Ad-hoc session: [session_type_name] on [date]".
- Ad-hoc lines sort after `core_childcare` (SortOrder=3+).
- Funding deduction applies to `core_childcare` only, NOT to ad-hoc lines. Ad-hoc is always fully billable.
- Angular: invoice table component renders `ad_hoc` lines with distinct styling (italic or "ad-hoc" badge).
- Update CONCEPTS.md: add Ad-Hoc Booking definition.

**Patterns to follow:**
- Cross-module lookup: `AdHocBookingLookup` interface in `billing/domain/`, adapter in bootstrap.
- Line creation follows existing pattern from `generate_term_invoices.go`.
- Ad-hoc bookings do NOT contribute to `booked_core_minutes` for funding deduction purposes — they are supplementary revenue.

**Test scenarios:**
- Covers Gap D.
- Happy path: child with 2 ad-hoc bookings in billing month → 2 `ad_hoc` invoice lines created.
- Happy path: ad-hoc line uses branch multiplier 1.50 → 240min session charged as 360min.
- Happy path: ad-hoc booking for session type with flat_fee_minor=5000 → charged £50.00 regardless of duration.
- Happy path: no ad-hoc bookings in month → no `ad_hoc` lines generated.
- Edge case: ad-hoc booking on date outside any term (term-time child) → still billed (ad-hoc is not term-filtered).
- Edge case: ad-hoc booking uses same session type as recurring pattern → separate lines (recurring + ad_hoc).
- Edge case: ad-hoc booking cancelled before generation → not included in invoice.
- Funding: funded deduction unchanged regardless of ad-hoc bookings present.
- Integration: end-to-end generation with mixed recurring + ad-hoc produces correct line items.

**Verification:** Invoice generation for month with ad-hoc bookings produces distinct `ad_hoc` lines. Amounts reflect multiplier correctly. Funded deduction not affected by ad-hoc bookings. `go test ./internal/modules/billing/...` passes.

---

### Phase 4: Session Type Enhancements

#### U7. Session Type Kind and Flat Fee

**Goal:** Add `kind` classification and optional `flat_fee_minor` to session types. Modify billing to use flat fee when set.

**Requirements:** R5, R6

**Dependencies:** None (parallel with Phase 1-3 possible)

**Files:**
- `api/db/migrations/00000X_add_session_type_kind_and_flat_fee.up.sql`
- `api/db/migrations/00000X_add_session_type_kind_and_flat_fee.down.sql`
- `api/internal/modules/sessiontypes/domain/entities.go`
- `api/db/query/session_types.sql`
- `api/internal/modules/sessiontypes/application/create_session_type.go`
- `api/internal/modules/sessiontypes/application/update_session_type.go`
- `api/internal/modules/sessiontypes/infrastructure/postgres/repository.go`
- `api/internal/modules/sessiontypes/interfaces/http/handler.go`
- `api/internal/modules/sessiontypes/interfaces/http/dto.go`
- `api/internal/modules/billing/application/generate_term_invoices.go`
- `api/internal/modules/billing/domain/booked_minutes.go`
- `web/src/app/features/staff/data/session-types-api.service.ts`
- `web/src/app/features/staff/models/session-type.models.ts`
- `web/src/app/features/staff/pages/manager-session-type-form/manager-session-type-form.component.ts`
- `web/src/app/features/staff/pages/manager-session-type-form/manager-session-type-form.component.html`
- `web/src/app/features/owner/pages/owner-session-type-form/owner-session-type-form.component.ts`
- `api/internal/modules/sessiontypes/application/application_test.go`
- `api/internal/modules/sessiontypes/infrastructure/postgres/repository_test.go`
- `api/internal/modules/billing/domain/booked_minutes_test.go`

**Approach:**
- Migration: `ALTER TABLE session_types ADD COLUMN kind text NOT NULL DEFAULT 'standard'`. Add CHECK constraint: `kind IN ('standard', 'wraparound_before', 'wraparound_after', 'core', 'extended')`.
- Migration: `ALTER TABLE session_types ADD COLUMN flat_fee_minor int`. Nullable — null means use standard hourly calculation. CHECK: `flat_fee_minor >= 0` when not null.
- Domain: add `Kind string` and `FlatFeeMinor *int` to `SessionType` struct. Add domain validation: kind must be valid enum value.
- Application: `create_session_type` and `update_session_type` accept `kind` and `flat_fee_minor`. Validate flat_fee_minor is non-negative when provided.
- HTTP DTO: add `kind` and `flat_fee_minor` to request/response JSON.
- Billing integration: in `generate_term_invoices.go`, when building invoice lines for each `BookedSession`, check the session type's `flat_fee_minor`:
  - If set → line amount = `flat_fee_minor × occurrences_in_month` (not hourly × minutes).
  - If null → use existing hourly × minutes calculation.
- `BookedSessionType` struct in `billing/domain/booked_minutes.go`: add `Kind string` and `FlatFeeMinor *int` fields so the billing pipeline can branch.
- Angular form: dropdown for `kind` (default 'standard'). Number input for `flat_fee_minor` (shown only when kind is not 'standard', or always available as an override). Format as GBP on display.
- Wraparound validation: when kind is `wraparound_before` or `wraparound_after`, flat_fee_minor is RECOMMENDED but not required (UI warning, not API error).

**Patterns to follow:**
- Enum stored as text with application-level validation (matching existing pattern for status fields — no DB enum types).
- Nullable int: `*int` in Go, `number | null` in TypeScript.
- Backward-compatible default: `DEFAULT 'standard'` ensures existing session types work unchanged.
- `BookedSessionType` projection extended — no new entity needed.

**Test scenarios:**
- Covers Gap G.
- Happy path: create session type with kind='wraparound_before' and flat_fee_minor=1500 → (£15.00 flat fee).
- Happy path: create session type with kind='standard' and flat_fee_minor=null → standard hourly billing.
- Happy path: update existing session type from standard to core → changes reflected.
- Edge case: update flat_fee_minor on session type already used in historical invoices → new invoices use new value, old invoices unchanged (immutability trigger).
- Edge case: flat_fee_minor=0 → valid (free session). Line amount = 0.
- Error: invalid kind value → domain error rejected.
- Error: flat_fee_minor < 0 → domain error rejected.
- Billing: session type with flat_fee_minor produces flat line amount regardless of session duration.
- Billing: wraparound session type without flat_fee_minor falls back to hourly × minutes.
- Backward compatibility: existing session types loaded as kind='standard', flat_fee_minor=null.

**Verification:** Session type CRUD includes kind and flat fee fields. Invoice generation uses flat fee when set. Kind dropdown works in Angular. `go test ./internal/modules/sessiontypes/...` and `go test ./internal/modules/billing/...` pass.

---

## Verification Contract

### Unit Test Gates

| Module | Test File | Key Assertions |
|--------|-----------|----------------|
| `term_calendar/application` | `create_term_test.go` | Valid creation, date validation, kind validation |
| `billing/domain` | `booked_minutes_test.go` | Term-time filtering, term-time funding math, ad-hoc charge calculation, flat fee branching |
| `ad_hoc_bookings/application` | `create_booking_test.go` | Date validation, session type validation, child validation |
| `sessiontypes/application` | `application_test.go` | Kind validation, flat fee validation |

### Integration Test Gates

| Scope | Test File | Key Assertions |
|-------|-----------|----------------|
| `term_calendar/postgres` | `repository_test.go` | CRUD round-trips, unique name constraint |
| `ad_hoc_bookings/postgres` | `repository_test.go` | Create + list + cancel round-trips |
| `billing/application` | `generate_term_invoices_test.go` | Term-time child invoice, ad-hoc inclusion, flat fee application |

### Acceptance Verification

| Scenario | Expected Outcome |
|----------|-----------------|
| Term-time-only child, September billing (in-term) | Billed for term days only. Funding allowance proportionally allocated. |
| Term-time-only child, August billing (holiday) | Billed 0 core minutes. Funding allowance = 0. |
| Year-round child, any month | Billed 52 weeks / 12 months as before. Funding uses stretched formula. |
| Create ad-hoc booking, then generate invoice | Invoice contains `ad_hoc` line with multiplied/flat-fee amount. |
| Session type with flat_fee_minor=2000, 20 occurrences | Invoice line amount = £20.00 × 20 = £400.00. |
| Standard session type, no flat fee | Invoice line amount = hourly × minutes (unchanged). |
| Funded deduction line on invoice | Includes `funding_model` label in details JSON. |

---

## Definition of Done

1. Migration scripts apply and roll back cleanly: `make migrate-up && make migrate-down`.
2. `go fmt ./...`, `go vet ./...`, `go build ./...` pass in `api/` with zero warnings.
3. `npm run lint` passes in `web/` with zero errors.
4. `ng build` (production) passes with zero diagnostics.
5. All unit test scenarios enumerated for each implementation unit pass.
6. All integration test scenarios enumerated pass (when `TEST_DATABASE_URL` provided).
7. CONCEPTS.md updated with new domain terms: Academic Term, Ad-Hoc Booking, Session Type Kind.
8. API endpoints functional and tested via handler tests.
9. Angular management UIs render and persist data correctly.
10. Existing full-time, part-time, sessional patterns produce unchanged invoices.

---

## System-Wide Impact

### Affected Modules

- **billing:** Core calculation changes (term-time filtering, flat fee branching, ad-hoc inclusion). Highest change density.
- **children:** Booking pattern entity extended (term_time_only flag).
- **funding:** Allowance computation context enriched (term-time awareness).
- **sessiontypes:** Entity extended (kind, flat_fee_minor).
- **term_calendar:** New module (CRUD for academic terms).
- **ad_hoc_bookings:** New module (CRUD for one-off bookings).
- **bootstrap:** New adapters for cross-module term lookup and ad-hoc booking lookup.

### Affected User Roles

- **Manager:** New term calendar management UI, booking pattern toggle, ad-hoc booking form, session type form upgrades.
- **Owner:** Read access to term calendar and session type changes.
- **Parent:** Invoice display changes (ad-hoc lines, funding model label). No new parent-facing actions.

### Data Migration Impact

- All migrations are additive (new tables, new columns with defaults). No data migration scripts needed.
- Existing session types: default to `kind='standard'`, `flat_fee_minor=null` — behavior unchanged.
- Existing booking patterns: default to `term_time_only=false` — behavior unchanged.
- Existing branches: default to `ad_hoc_rate_multiplier=1.50` — sensible starting point.

---

## Risks and Dependencies

### Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Term-time calculation narrows ADR-0006 | High | Explicit documentation in CONCEPTS.md and ADR update. Default behavior (term_time_only=false) is identical to current. Test coverage spans both paths. |
| Academic term dates entered incorrectly by managers | Medium | Validation: start < end, no empty dates, reasonable range. Future enhancement: half-term auto-split or calendar import. |
| Ad-hoc bookings not cleaned up if child is deactivated | Low | Booking creation validates child is active. Cancel endpoint available. No FK cascade (by design — preserves audit). |
| Flat fee set to 0 unintentionally | Low | UI warning for flat_fee_minor=0. Domain permits (could be complimentary session). |
| Invoice generation performance with ad-hoc lookups | Low | Ad-hoc query scoped to (tenant, branch, child, month). Index on those columns. Single query per child during generation. |

### Dependencies

| Dependency | Status |
|------------|--------|
| Phase 2 depends on Phase 1 (term calendar for date lookups) | Enforced via U-ID deps |
| U6 depends on U3 (term-time billing) and U5 (ad-hoc entity) | Cross-phase |
| U7 (session types) is independent of term work | Parallelizable |
| All Go work requires `make sqlc-generate` after migration creation | Build process |
| Angular work depends on Go API completion | Sequential per phase |

---

## Sources and Research

### Internal Sources

- `docs/reports/BOOKING-TYPES-GAP-ANALYSIS.md` — origin document, all gap definitions and priorities
- `docs/adr/0006-booking-pattern-billing-source.md` — booking patterns are billing truth (supersedes ADR-0005)
- `docs/adr/0007-12-month-fixed-term-contract.md` — term structure and lifecycle
- `docs/adr/0009-session-templates-named-reusable-weeks.md` — copy semantics for templates
- `docs/adr/0011-registration-funding-guidance-only.md` — funding profile manual control
- `docs/DECISION-BASELINE.md` — 52-week billing rule, funding formula, invoice immutability
- `docs/solutions/invoice-implementation-analysis.md` — billing pipeline gaps and patterns
- `docs/solutions/architecture-patterns/clean-architecture-dependency-rule.md` — Tx alias pattern
- `docs/agents/ARCHITECTURE.md` — module structure, forbidden imports, new module checklist
- `docs/agents/TESTING.md` — test conventions (unit vs integration)

### Code References

- `api/internal/modules/billing/domain/booked_minutes.go` — core calculation functions
- `api/internal/modules/billing/application/generate_term_invoices.go` — advance-pay generation pipeline
- `api/internal/modules/billing/domain/invoice.go` — invoice constants and line kinds
- `api/internal/modules/term/domain/term.go` — term entity and lifecycle
- `api/internal/modules/funding/domain/profile.go` — funding profile entity
- `api/internal/modules/children/domain/child_booking_pattern.go` — booking pattern entity
- `api/internal/modules/sessiontypes/domain/entities.go` — session type entity
- `api/internal/app/bootstrap/adapters.go` — cross-module adapter wiring
- `api/db/migrations/000001_baseline.up.sql` — existing schema reference

---

## Phased Delivery

### Phase 1: Term Calendar Foundation (U1)
New module. No dependencies. Establishes the data primitive that Phase 2 consumes.

### Phase 2: Term-Time Only Enforcement (U2, U3, U4)
Depends on Phase 1. Changes billing pipeline. Highest-risk phase due to ADR-0006 narrowing. Requires careful testing of both term-time and year-round paths.

### Phase 3: Ad-Hoc Bookings (U5, U6)
U5 is independent of Phases 1-2 and can be parallelized. U6 depends on Phase 2's billing changes being in place. Adds new entity and invoice integration.

### Phase 4: Session Type Enhancements (U7)
Independent of all other phases. Can be built and shipped separately. Modifies existing entity with backward-compatible defaults.

### Post-Release Follow-Up
- P2 hourly booking surface (separate plan needed)
- P3 funding carry-forward policy
- P2 funding model display on invoice detail view (data present in line details; UI enhancement)
- Advanced term management (half-term auto-split, bank holiday exclusion)
- School-age wraparound flow (school records, roster management)
- **ADR-0006 update:** Create ADR-0012 documenting the term-time-only exception to the "always bill 52 weeks" rule (narrowing, not revocation)
- **Missing scheduler job:** Monthly advance generation cron (noted in `docs/solutions/invoice-implementation-analysis.md`) not addressed here; orthogonal to new features
