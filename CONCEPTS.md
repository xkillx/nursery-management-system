# Concepts

Shared domain vocabulary for this project — entities, named processes, and status concepts with project-specific meaning. Seeded with core domain vocabulary, then accretes as ce-compound and ce-compound-refresh process learnings; direct edits are fine. Glossary only, not a spec or catch-all.

## Children

### Booking Pattern
A child's planned weekly attendance schedule, defining which session types the child attends on which days of the week. A child has at most one current (open) booking pattern at a time; when a new pattern is created, the previous one is closed. Patterns cannot be backdated or edited once their effective date has passed.

### Session Type
A predefined time slot offered by a nursery (e.g., "Morning 08:00–13:00" or "Afternoon 13:00–18:00"). Session types are configured at the branch level and shared across all children.

### Funding Record
A child's funding entitlement record, storing the funding type (e.g., 15 Hours, 30 Hours), funding model (term-time or stretched), eligibility status, and benefit information. Separate from billing — funding records capture entitlement, not invoices.

### Benefits Status
Tracks whether a child receives benefits that contribute towards nursery fees. Values: `yes`, `no`, `unknown`. Separate from the specific benefit types or amounts.

### Room
A physical room in the nursery where a child is assigned. Each child has at most one current room assignment (`child_room_assignments` with `is_current = true`) at a time. Rooms have a name (e.g., "Toddler Room") and an `age_group`. Used on invoices and registration forms as a child group label.

### Parent Carer
A child's parent or legal guardian recorded in `child_contacts` with `contact_type = 'parent_carer'`. A child may have multiple parent carers. The primary parent carer (lowest `sort_order`) is used as the billing contact on invoices, providing full name and address.

## Billing

### Money
A value object representing a monetary amount in GBP. Wraps a non-negative integer of pence with immutable arithmetic (`Add`, `Multiply`), safe construction via `GBP()` constructor that validates non-negative, and backward-compatible JSON serialization as a raw integer. Replaces raw `int` fields across the billing module to enforce type safety at domain boundaries.

## Registration

### Stepper (Intake Wizard)
A multi-step form wizard that guides staff through child registration or editing. Steps: Child Details, Medical & Health, Contacts & Security, Permissions & Consents, Session Pattern, Funding & Benefits. In registration mode, steps are sequential; in edit mode, steps are independently accessible.

## Settings

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

### Booking Pattern
A child's planned weekly attendance schedule. The system counts day-of-week occurrences in the calendar month × session duration = `booked_core_minutes`. This is the billing basis under advance-pay.

### Funded Allowance Minutes
Monthly funding entitlement expressed in minutes. Derived from `funded_hours_per_week × 52 × 60 / 12` (annualised hours spread evenly across 12 months). Stored per (child, billing_month) in `funding_profiles`.

### Billable Minutes
`max(0, booked_core_minutes - funded_allowance_minutes)`. The minutes that the parent actually pays for after funding deduction.

### Pro-Ration
When a child starts mid-term, billable days are pro-rated: `booked_minutes × (eligible_days / total_days_in_month)`. The term always starts on the 1st; the child's `start_date` determines actual eligibility.
