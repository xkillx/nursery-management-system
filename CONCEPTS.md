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
