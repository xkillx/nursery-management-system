---
title: "Booking Pattern Consolidation — Recurring Booking as Single Source of Truth"
type: feat
date: 2026-07-22
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-plan-bootstrap
execution: code
---

## Goal Capsule

**Objective:** Eliminate `child_booking_patterns` and `child_booking_pattern_entries` as separate concepts. Make the recurring booking the single source of truth for a child's weekly schedule. Billing reads booking entries via a new `BookingEntriesLookup` interface. The `/manager/bookings` page becomes the sole hub for booking management.

**Authority hierarchy:** This plan's decisions were resolved during a 16-question design interview. The plan is the authority for implementation.

**Stop conditions:**
- Recurring booking stores per-day session entries, `term_time_only`, and effective dates
- Billing calculates invoices from recurring booking entries (not patterns)
- `child_booking_patterns` and `child_booking_pattern_entries` tables are dropped
- All booking pattern frontend code and API endpoints are removed
- Child profile shows read-only "Current Booking" summary
- `go vet ./...`, `go build ./...`, `npx ng lint` pass clean

**Execution profile:** Code implementation. Three-phase sequential execution.

**Tail ownership:** `ce-work` or `/goal` executor.

---

## Product Contract

### Summary

Consolidate booking patterns into recurring bookings. Today the system maintains two overlapping concepts: `child_booking_patterns` (weekly schedule template with per-day session types, effective date ranges, term-time-only flag) and `bookings` (operational booking record). Both store "which days does this child attend and what session?" — in two different places with two different schemas. This creates confusion about the source of truth and forces billing to read from a separate pattern table.

After consolidation, a recurring booking IS the child's weekly schedule. One concept, one place.

### Problem Frame

- **Duplicate concepts:** Booking patterns and recurring bookings store the same data (day-of-week + session type per day, date range). Two sources of truth for one fact.
- **Billing coupling:** Billing reads from `child_booking_pattern_entries` via `ListBookingPatternEntries`. It should read from bookings.
- **Term coupling:** The term module has `BookingPatternLookup` interface pointing at pattern IDs. It should resolve bookings by child+date at calculation time.
- **Scattered UX:** A manager must visit a child's profile to set a booking pattern, then visit `/manager/bookings` to create the actual booking. These should be one flow.

### Requirements

**Recurring Booking as Source of Truth**

- R1. A recurring booking stores per-day session type entries (`session_entries` JSONB), `term_time_only` (boolean), `effective_start_date`, `effective_end_date`.
- R2. One active recurring booking per child. No versioning — edit in place. To schedule a future change: cancel current + create new with future start date.
- R3. `days_of_week` and `session_template_id` columns are removed from the `bookings` table. `session_entries` is the sole source of truth. `days_of_week` is derived from entries for display.
- R4. The existing `SessionGridComponent` (day × session type grid) is the UI for selecting entries.

**Billing Integration**

- R5. Billing defines a `BookingEntriesLookup` interface: `GetEntriesForChildInMonth(ctx, tenantID, branchID, childID, billingMonth) ([]BookedPatternEntry, error)`. Returns the active recurring booking's entries for a given month.
- R6. The bookings module implements `BookingEntriesLookup` by querying `bookings` where `child_id = X AND status = 'active' AND effective_start_date <= monthEnd AND (effective_end_date IS NULL OR effective_end_date >= monthStart)`.
- R7. Wired in `bootstrap/adapters.go`. Billing no longer reads from `child_booking_pattern_entries`.

**Term Module**

- R8. Replace `BookingPatternLookup` with `BookingEntriesLookup`. The term does not store a booking reference. At invoice generation time, billing resolves the active booking for the child+month.

**Frontend**

- R9. Remove booking pattern page (`/manager/children/{id}/booking-pattern`), component, and service methods.
- R10. Remove booking pattern API endpoints from the children module.
- R11. Add `term_time_only` toggle to the create-recurring booking page.
- R12. Child profile page gets a read-only "Current Booking" section showing the child's active recurring booking (days, sessions, dates). Links to `/manager/bookings` to edit.
- R13. Keep all three create pages: `/manager/bookings/new/recurring`, `/manager/bookings/ad-hoc/new`, `/manager/bookings/hourly/new`.

**Schema**

- R14. Add `term_time_only BOOLEAN NOT NULL DEFAULT false` to `bookings` table.
- R15. Drop `child_booking_patterns` and `child_booking_pattern_entries` tables.
- R16. Remove `days_of_week` and `session_template_id` columns from `bookings`.

### Scope Boundaries

**In scope:**
- Schema migration (add column, drop tables, remove columns)
- `BookingEntriesLookup` interface + bookings module implementation
- Billing rewiring to use new interface
- Term module interface replacement
- Booking pattern code removal (backend + frontend)
- Child profile "Current Booking" read-only section
- `term_time_only` on create-recurring page
- Update booking response DTOs to derive `days_of_week` from entries

**Out of scope:**
- Ad-hoc and hourly booking changes (unchanged)
- Invoice PDF changes
- Parent portal booking view changes
- Booking wizard drawer component (can be removed later as cleanup)
- Data migration (no live data exists)

---

## Key Technical Decisions

### KTD-1: `session_entries` as sole booking structure

**Decision:** Remove `days_of_week` (int array) and `session_template_id` (UUID) from `bookings`. Keep only `session_entries` JSONB.

**Rationale:** `session_entries` already stores per-day session types and is the richer structure. `days_of_week` is derivable. `session_template_id` is unused in the per-day model. Removing redundant columns prevents dual-source-of-truth bugs.

**Alternative considered:** Keep `days_of_week` as denormalized convenience column. Rejected — adds write-time sync burden for no real benefit.

### KTD-2: `BookingEntriesLookup` resolves by child+date, not booking ID

**Decision:** The interface takes `(childID, billingMonth)` and returns entries. No booking ID stored on terms.

**Rationale:** Decouples terms from specific booking records. If a booking is replaced, billing automatically picks up the new one. More resilient than foreign-key references.

**Alternative considered:** Store `booking_id` on the term. Rejected — creates broken-reference risk and requires term updates when bookings change.

### KTD-3: One active recurring booking per child, no versioning

**Decision:** A child has at most one active recurring booking. Edit in place. No effective-dated version chain.

**Rationale:** Versioning via multiple pattern rows was a workaround for "draft future changes." A simpler model (one booking, edit directly) covers 95% of cases. Future changes can use pause + create with future start date.

**Alternative considered:** Full versioning (multiple bookings with overlapping dates, one "current"). Rejected — adds complexity for a rare use case.

### KTD-4: Three-phase execution

**Decision:** Phase 1 (additive: schema + interface), Phase 2 (remove pattern code), Phase 3 (frontend cleanup).

**Rationale:** Each phase leaves the system working. Phase 1 is purely additive. Phase 2 removes old code only after billing works on the new path. Phase 3 cleans up the frontend. If any phase has issues, you can pause.

---

## Implementation Units

### Phase 1: Additive — Schema, Interface, Billing Rewiring

#### IU-1.1: Add `term_time_only` to bookings table

**Files:**
- `api/internal/platform/db/migrations/` (new migration)
- `api/internal/platform/db/sqlc/bookings.sql.go` (regenerated)
- `api/internal/modules/bookings/domain/entities.go` (add field)
- `api/internal/modules/bookings/interfaces/http/dto.go` (add to request/response)
- `api/internal/modules/bookings/interfaces/http/handler.go` (accept in create/update)

**Changes:**
- Migration: `ALTER TABLE bookings ADD COLUMN term_time_only BOOLEAN NOT NULL DEFAULT false;`
- Add `TermTimeOnly bool` to `domain.Booking` struct
- Add `term_time_only` to `createBookingRequest` and `updateBookingRequest` DTOs
- Pass through in create/update handlers
- Update `BookingsCreate` and `BookingsUpdate` sqlc queries

**Test scenarios:**
- Create recurring booking with `term_time_only: true` → persisted and returned in response
- Create recurring booking without `term_time_only` → defaults to `false`
- Update booking to set `term_time_only`

#### IU-1.2: Define `BookingEntriesLookup` interface in billing module

**Files:**
- `api/internal/modules/billing/domain/booked_minutes.go` (or new file `booking_lookup.go`)

**Changes:**
- Define interface:
  ```go
  type BookingEntriesLookup interface {
      GetEntriesForChildInMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) ([]BookedPatternEntry, error)
  }
  ```
- `BookedPatternEntry` and `BookedSessionType` stay as-is (they're pure value objects used by billing calculation — no rename needed)

**Test scenarios:**
- Interface compiles and is mockable

#### IU-1.3: Implement `BookingEntriesLookup` in bookings module

**Files:**
- `api/internal/modules/bookings/application/booking_entries_lookup.go` (new)
- `api/internal/modules/bookings/domain/repository.go` (add query method if needed)

**Changes:**
- Implement `GetEntriesForChildInMonth`: query bookings where `child_id = X AND status = 'active' AND effective_start_date <= monthEnd AND (effective_end_date IS NULL OR effective_end_date >= monthStart)`
- Parse `session_entries` JSONB into `[]billing.BookedPatternEntry`
- Resolve session type details (name, start/end minutes) from `session_types` table join

**Test scenarios:**
- Child with active recurring booking → returns entries
- Child with no booking → returns empty
- Child with cancelled booking → returns empty
- Booking with `effective_end_date` before month → returns empty
- Booking starting mid-month → returns entries (billing filters by occurrence dates)

#### IU-1.4: Wire `BookingEntriesLookup` in bootstrap

**Files:**
- `api/internal/app/bootstrap/adapters.go`

**Changes:**
- Add adapter struct implementing `billing.BookingEntriesLookup`
- Wire into billing module dependencies
- Remove old `ListBookingPatternEntries` wiring if present

**Test scenarios:**
- Integration: billing prefill uses booking entries, not pattern entries

#### IU-1.5: Update billing to use `BookingEntriesLookup`

**Files:**
- `api/internal/modules/billing/application/compute_invoice_prefill.go`
- `api/internal/modules/billing/application/` (any file using `ListBookingPatternEntries`)

**Changes:**
- Replace `repo.ListBookingPatternEntries(ctx, tx, tenantID, branchID, patternID)` with `bookingLookup.GetEntriesForChildInMonth(ctx, tenantID, branchID, childID, billingMonth)`
- Remove `BookingPatternID` from `InvoicePrefillParams` (or keep as optional metadata)
- Update `AdvancePayTermRow` to not require `BookingPatternID`

**Test scenarios:**
- Invoice prefill produces identical calculation using booking entries
- Funded child with term_time_only booking → correct allowance calculation

#### IU-1.6: Replace `BookingPatternLookup` in term module

**Files:**
- `api/internal/modules/term/application/create_term.go`
- `api/internal/modules/term/application/request_schedule_change.go`
- `api/internal/app/bootstrap/` (wiring)

**Changes:**
- Remove `BookingPatternLookup` interface from term module
- Term creation no longer requires a booking pattern reference
- If term stored `booking_pattern_id`, remove it (terms reference child+month, not pattern)

**Test scenarios:**
- Term creation works without booking pattern reference
- Invoice generation resolves booking at calculation time

#### IU-1.7: Update booking response DTOs

**Files:**
- `api/internal/modules/bookings/interfaces/http/dto.go`

**Changes:**
- Remove `days_of_week` and `session_template_id` from `bookingResponse`
- Keep `session_entries` as the primary field
- If any consumer needs `days_of_week`, derive it from entries in the response mapper

**Test scenarios:**
- Booking response contains `session_entries` but not `days_of_week` or `session_template_id`

#### IU-1.8: Remove `days_of_week` and `session_template_id` from bookings table

**Files:**
- `api/internal/platform/db/migrations/` (new migration)
- `api/internal/platform/db/sqlc/bookings.sql.go` (regenerated)
- `api/internal/modules/bookings/domain/entities.go`

**Changes:**
- Migration: `ALTER TABLE bookings DROP COLUMN days_of_week, DROP COLUMN session_template_id;`
- Remove fields from `Booking` struct
- Update sqlc queries that reference these columns

**Test scenarios:**
- All booking CRUD operations work without these columns
- Booking list/detail responses are correct

---

### Phase 2: Remove Booking Pattern Backend Code

#### IU-2.1: Remove booking pattern API endpoints

**Files:**
- `api/internal/modules/children/interfaces/http/handler.go` (remove pattern routes)
- `api/internal/app/bootstrap/` (unregister routes)

**Changes:**
- Remove all `/children/{id}/booking-patterns` routes
- Remove handler methods for pattern CRUD

**Test scenarios:**
- API returns 404 for pattern endpoints
- No compilation errors

#### IU-2.2: Remove booking pattern use cases

**Files:**
- `api/internal/modules/children/application/booking_patterns.go` (delete)

**Changes:**
- Delete `ListBookingPatterns`, `GetBookingPattern`, `GetCurrentBookingPattern`, `CreateBookingPattern`, `UpdateBookingPattern`
- Delete `resolveBookingPatternEntries`, `createBookingPatternInTx` helpers

**Test scenarios:**
- `go build ./...` passes

#### IU-2.3: Remove booking pattern repository methods

**Files:**
- `api/internal/modules/children/domain/repository.go` (remove pattern methods from interface)
- `api/internal/modules/children/infrastructure/postgres/repository.go` (remove implementations)

**Changes:**
- Remove `ListByChild`, `GetPatternByID`, `GetActiveForDate`, `InsertPattern`, `CloseCurrentPattern`, `ClosePatternByID`, `ReplaceEntries`, `UpdateEffectiveFrom`, `UpdateTermTimeOnly`, `GetPreviousClosedByChild`, `ExistsInScope` (pattern-specific)
- Remove `entriesForPattern`, `entriesForPatternTx`, `mapBookingPatternRow` helpers

**Test scenarios:**
- `go build ./...` passes
- Children module tests pass (non-pattern tests)

#### IU-2.4: Drop booking pattern tables

**Files:**
- `api/internal/platform/db/migrations/` (new migration)

**Changes:**
- `DROP TABLE child_booking_pattern_entries;`
- `DROP TABLE child_booking_patterns;`

**Test scenarios:**
- Migration applies cleanly
- No code references the dropped tables

#### IU-2.5: Remove generated sqlc code for patterns

**Files:**
- `api/internal/platform/db/sqlc/child_booking_patterns.sql.go` (delete)
- `api/internal/platform/db/sqlc/child_booking_pattern_entries.sql.go` (delete)
- `api/internal/platform/db/query/child_booking_patterns.sql` (delete)
- `api/internal/platform/db/query/child_booking_pattern_entries.sql` (delete)

**Changes:**
- Delete the sqlc query files and regenerated Go files

**Test scenarios:**
- `make sqlc-generate` passes
- `go build ./...` passes

#### IU-2.6: Remove booking pattern types from children domain

**Files:**
- `api/internal/modules/children/domain/` (entities)

**Changes:**
- Remove `BookingPattern`, `BookingPatternEntry`, `EntrySessionType` types if they exist in the children domain
- Remove `SessionTypeLookup`, `SessionTypeInfo` if only used by patterns

**Test scenarios:**
- `go build ./...` passes

---

### Phase 3: Frontend Cleanup

#### IU-3.1: Remove booking pattern page and component

**Files:**
- `web/src/app/features/staff/pages/manager-booking-pattern/` (delete directory)
- `web/src/app/app.routes.ts` (remove route)

**Changes:**
- Delete `ManagerBookingPatternComponent` and its template/spec
- Remove route for `/manager/children/:id/booking-pattern`

**Test scenarios:**
- `npx ng lint` passes
- `ng build` passes

#### IU-3.2: Remove booking pattern service methods

**Files:**
- `web/src/app/features/staff/data/staff-api.service.ts`

**Changes:**
- Remove `listChildBookingPatterns`, `getCurrentChildBookingPattern`, `getChildBookingPattern`, `createChildBookingPattern`, `updateChildBookingPattern`

**Test scenarios:**
- `npx ng lint` passes
- No broken imports

#### IU-3.3: Remove booking pattern models

**Files:**
- `web/src/app/features/staff/models/booking-pattern.models.ts` (delete)

**Changes:**
- Delete `BookedSession`, `SessionTypeRef`, `BookingPattern`, `BookingPatternInput`, `BookingPatternEntryInput` interfaces

**Test scenarios:**
- `ng build` passes
- No broken imports

#### IU-3.4: Add `term_time_only` to create-recurring booking page

**Files:**
- `web/src/app/features/staff/pages/create-recurring-booking/create-recurring-booking.component.ts`
- `web/src/app/features/staff/pages/create-recurring-booking/create-recurring-booking.component.html`
- `web/src/app/features/staff/pages/create-recurring-booking/booking-summary-sidebar/booking-summary-sidebar.component.ts`

**Changes:**
- Add `termTimeOnly = false` boolean field
- Add checkbox/toggle in the form template
- Pass `term_time_only` in the `CreateRecurringBookingRequest`
- Show in summary sidebar

**Test scenarios:**
- Toggle term_time_only → included in API request
- Default is false

#### IU-3.5: Update booking models and API service

**Files:**
- `web/src/app/features/staff/models/booking.models.ts`
- `web/src/app/features/staff/data/bookings-api.service.ts`

**Changes:**
- Add `term_time_only` to `CreateRecurringBookingRequest`
- Add `term_time_only` to `UnifiedBooking` if needed for display
- Remove `days_of_week` from request/response models if present
- Remove `session_template_id` from models if present

**Test scenarios:**
- `ng build` passes

#### IU-3.6: Add "Current Booking" section to child profile

**Files:**
- `web/src/app/features/staff/pages/` (child profile component — find the existing one)
- New component or inline section

**Changes:**
- Call `bookingsApi` to fetch active recurring booking for the child
- Display: session grid (read-only), effective dates, term_time_only status
- "Manage Bookings" link to `/manager/bookings?childId={id}`

**Test scenarios:**
- Child with active booking → shows summary
- Child with no booking → shows "No active booking" message
- Link navigates to filtered bookings list

#### IU-3.7: Remove unused booking wizard drawer (optional cleanup)

**Files:**
- `web/src/app/shared/components/booking/booking-wizard-drawer/` (delete if unused)

**Changes:**
- Check if `BookingWizardDrawerComponent` is referenced anywhere. If not, delete.

**Test scenarios:**
- `ng build` passes

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Billing calculation produces different results after switching data source | Medium | High | Write integration test that creates a booking with entries, generates invoice, and compares calculation to expected values. Run before and after Phase 1.5. |
| `session_entries` JSONB parsing fails for edge cases (empty array, missing session type) | Low | Medium | Unit test the `BookingEntriesLookup` implementation with: empty entries, archived session types, missing session types. |
| Term module breaks when `BookingPatternLookup` is removed | Medium | Medium | Phase 1.6 removes the interface dependency. Verify term creation still works. If terms stored `booking_pattern_id`, check for NULL references. |
| Frontend build breaks from removed imports | Low | Low | `ng build` after each frontend change. Angular's compiler catches missing imports immediately. |
| Orphaned references to pattern tables in code not caught by build | Low | Medium | `grep -r "booking_pattern" api/` after Phase 2 to catch any remaining references. |

---

## Sources & Research

- **Existing data model:** `bookings` table already has `session_entries` JSONB with per-day session types. `SessionGridComponent` already renders the day × session type grid. `CreateRecurringBookingRequest` already accepts `session_entries`.
- **Billing pipeline:** `ComputeInvoicePrefill` → `ListBookingPatternEntries` → `child_booking_pattern_entries`. This is the primary integration point to rewire.
- **Term module:** `BookingPatternLookup` interface in `api/internal/modules/term/application/`. Used by `CreateTermUseCase` and `RequestScheduleChangeUseCase`.
- **Pattern CRUD:** Full CRUD in `api/internal/modules/children/application/booking_patterns.go`. Repository in `api/internal/modules/children/infrastructure/postgres/repository.go`.
- **Frontend pattern page:** `web/src/app/features/staff/pages/manager-booking-pattern/`. Service methods in `staff-api.service.ts`.
