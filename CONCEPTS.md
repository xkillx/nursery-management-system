# Concepts

Shared domain vocabulary for this project — entities, named processes, and status concepts with project-specific meaning. Seeded with core domain vocabulary, then accretes as ce-compound and ce-compound-refresh process learnings; direct edits are fine. Glossary only, not a spec or catch-all.

## Children

### Booking Pattern
A child's planned weekly attendance schedule, defining which session types the child attends on which days of the week. A child has at most one current (open) booking pattern at a time; when a new pattern is created, the previous one is closed. Patterns cannot be backdated or edited once their effective date has passed. A pattern may carry an optional `term_time_only` flag; when true, billing calculations exclude dates outside the branch's academic term calendar.

### Session Entries
Per-day session type and room selections stored on a recurring booking as a JSONB array of `{day_of_week, session_type_id, room_id}` objects. Each entry specifies which session type and which room the child attends on that day. Replaces the session template abstraction for new bookings — managers pick session types directly on a day × session-type grid instead of selecting a pre-configured template. Stored in the `bookings.session_entries` column. Existing bookings that reference a `session_template_id` remain valid.

### Ad-Hoc Booking
A one-off session booking for a specific child on a specific calendar date, independent from the recurring booking pattern. Used for backup childcare, inset-day extras, or casual sessions. Billed at the ad-hoc rate (branch hourly rate × ad-hoc multiplier). Status: `active` or `cancelled`. Does not affect funding allowance calculations.

### Hourly Booking
A flexible-duration session booking for a specific child on a specific calendar date, with explicit start time and duration in minutes. Unlike ad-hoc bookings (which use predefined session types), hourly bookings allow arbitrary time blocks. Used for irregular schedules, extra hours, or custom arrangements. Status: `active` or `cancelled`.

### Session Type
A predefined time slot offered by a nursery (e.g., "Morning 08:00–13:00" or "Afternoon 13:00–18:00"). Session types are configured at the branch level and shared across all children. Each session type has a `kind` classification (`standard`, `wraparound_before`, `wraparound_after`, `core`, `extended`).

### Funding Record
A child's funding entitlement record, storing the funding type (e.g., 15 Hours, 30 Hours), funding model (term-time or stretched), eligibility status, and benefit information. Separate from billing — funding records capture entitlement, not invoices.

### Benefits Status
Tracks whether a child receives benefits that contribute towards nursery fees. Values: `yes`, `no`, `unknown`. Separate from the specific benefit types or amounts.

### Room
A physical room in the nursery. Rooms have a name (e.g., "Toddler Room") and an `age_group`. Room assignment lives on session entries (`session_entries[*].room_id`), not on the child or booking. Each session entry specifies its own room, allowing a child to attend different rooms on different days. An optional `home_room_id` on the children table provides an administrative label for registration and child list filtering.

### Home Room
An optional administrative room label on the children table (`children.home_room_id`). Assigned during registration (can be skipped). Used by child list filtering and as a display label on the child profile. Not used by operational features — capacity, attendance, billing, and registers all read room from session entries.

### Parent Carer
A child's parent or legal guardian recorded in `child_contacts` with `contact_type = 'parent_carer'`. A child may have multiple parent carers. The primary parent carer (lowest `sort_order`) is used as the billing contact on invoices, providing full name and address.

## Billing

### Money
A value object representing a monetary amount in GBP. Wraps a non-negative integer of pence with immutable arithmetic (`Add`, `Multiply`), safe construction via `GBP()` constructor that validates non-negative, and backward-compatible JSON serialization as a raw integer. Replaces raw `int` fields across the billing module to enforce type safety at domain boundaries.

## Registration

### Stepper (Intake Wizard)
A multi-step form wizard that guides staff through child registration or editing. Steps: Child Details, Medical & Health, Contacts & Security, Permissions & Consents, Session Pattern, Funding & Benefits. In registration mode, steps are sequential; in edit mode, steps are independently accessible.

## Settings

### Academic Term
A branch-level operational period representing a school term (e.g., "Autumn 2026", "Spring 2027"). Each academic term has a name, kind (`autumn`, `spring`, `summer`), start date, and end date. Configured per branch to reflect local school calendars. Used to determine which weeks are term-time for term-time-only booking patterns and funding calculations.

### Site Profile
A branch's brand identity (name, description, phone, email, website, address) used on parent-facing surfaces. One Site Profile per branch, separate from branch ops (archive, hourly rate). Currently rendered on the parent invoice header.

## Invoicing

### Term
A 12-month (adjustable) commercial commitment between a nursery and a parent for one child. Links a child's booking pattern + hourly rate at a snapshot. Lifecycle: `pre_term → active → pending_renewal → ended | terminated`. Only active terms are billable. Start date must be the 1st of a calendar month.

### Advance-Pay (Prepaid Monthly Invoicing)
Invoices are generated for the next calendar month before service is delivered. The billing month IS the service month. Generation typically runs on the 25th of the preceding month.

### Invoice Status
- `draft` — editable, not yet issued to parent
- `issued` — locked, invoice number assigned, parent can view/pay
- `paid` — payment received in full
- `payment_failed` — payment attempt failed
- `overdue` — due date passed without payment

### Invoice Lines
- `core_childcare` — the booking-driven childcare cost (booked minutes × hourly rate)
- `funded_deduction` — the funding entitlement deduction (funded minutes × hourly rate, subtracted)
- `extra` — ad-hoc charges added by manager
- `ad_hoc` — one-off booking charges for individual ad-hoc sessions

### Booking Pattern
A child's planned weekly attendance schedule. The system counts day-of-week occurrences in the calendar month × session duration = `booked_core_minutes`. This is the billing basis under advance-pay.

### Funded Allowance Minutes
Monthly funding entitlement expressed in minutes. Two computation paths:
- **Stretched model:** `funded_hours_per_week × 52 × 60 / 12` (annualised hours spread evenly across 12 months).
- **Term-time-only model:** `funded_hours_per_week × 60 × term_weekdays_in_month / 5`. Term months only; zero during holidays. Derived from the branch's academic term calendar.

Stored per (child, billing_month) in `funding_profiles`.

### Billable Minutes
`max(0, booked_core_minutes - funded_allowance_minutes)`. The minutes that the parent actually pays for after funding deduction.

### Pro-Ration
When a child starts mid-term, billable days are pro-rated: `booked_minutes × (eligible_days / total_days_in_month)`. The term always starts on the 1st; the child's `start_date` determines actual eligibility.
