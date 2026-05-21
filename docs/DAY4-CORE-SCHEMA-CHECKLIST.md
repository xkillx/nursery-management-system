# Day 4 Core Schema Checklist

Purpose: decision-to-schema checklist for Day 4 implementation and review.

## Scope of Day 4

- Core schema includes tenant/scope/auth dependencies plus child/guardian relationship model.
- Use forward-only migrations (next sequential version after `000003`).
- Done gate includes migration integrity (`up -> down -> up`) and invariant checks, not just command success.

## Core Tables and Lifecycle

- `users`: global unique `email_normalized`, account-level `is_active`.
- `memberships`: one row per `(tenant_id, branch_id, user_id)`, role constrained to `manager|practitioner|parent`, add `is_active` and optional `ended_at`.
- `children`: scoped by tenant/branch, enrollment `start_date` + optional `end_date`, lifecycle active/left semantics, include current core billing rate.
- `guardians`: scoped by tenant/branch, contact entity separate from login identity, email optional.
- `guardian_child_links`: many-to-many, one active link per pair, historical ended rows retained.
- `parent_membership_guardians`: explicit parent portal-access mapping, one active mapping per parent membership, historical ended rows retained.
- `refresh_tokens`: remains the session persistence model (no parallel sessions table).
- `audit_logs`: append-only mutation log with scope fields, actor nullable for system actions, explicit `request_id` correlation.

## Hard Invariants

- Membership branch must belong to membership tenant (DB-enforced consistency).
- Parent mapping rows may reference parent-role memberships only (app validation + DB trigger guard).
- Enrollment date invariant: `start_date <= end_date` when `end_date` present.
- Active-state consistency checks where both `is_active` and `ended_at` exist.
- Core FKs default to no cascade delete.

## Authorization Relationship Model

- Parent invoice visibility requires both:
  - active `parent_membership_guardians` mapping, and
  - active `guardian_child_links` relationship.
- Authorization uses live relationship state at request time (no long-lived entitlement cache).
- Unlink removes parent visibility immediately; relink restores it immediately.

## Data and Audit Semantics

- All timestamps stored as UTC `TIMESTAMPTZ`; business-time interpretation uses `Europe/London`.
- Monetary values stored in integer minor units (`GBP` pence).
- `audit_logs.details` is supplemental; query-critical dimensions are top-level columns.
- Authorization denials remain telemetry (logs/metrics), not persisted audit mutation events.

## Enrollment and Operations Rules

- Child requires minimum enrollment data before attendance/invoicing workflows.
- Attendance eligibility depends on active enrollment state, not mere row existence.
- Post-enrollment manager corrections for historical attendance remain allowed.

## Migration/Test Checklist

- Additive forward migrations only; do not edit previously applied migrations.
- Backfill-first pattern for nullable->non-null or new constraints.
- Regenerate/verify `sqlc` compatibility as part of Day 4 done criteria.
- Validate canonical active-row predicates in auth and relationship queries.
- Include realistic seed/invariant scenarios:
  - manager/practitioner/parent memberships,
  - child with two guardians,
  - one mapped parent membership,
  - active/inactive and link end/relink transitions.

## Local Verification Run

- Ran `migrate up -> down -> up` against `postgres://nursery_app:nursery_app@localhost:5432/nursery_management?sslmode=disable`.
- Loaded `api/cmd/seed/scenarios.sql` to validate core relationship constraints and relink behavior.
