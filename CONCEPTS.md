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

## Registration

### Stepper (Intake Wizard)
A multi-step form wizard that guides staff through child registration or editing. Steps: Child Details, Medical & Health, Contacts & Security, Permissions & Consents, Session Pattern, Funding & Benefits. In registration mode, steps are sequential; in edit mode, steps are independently accessible.
