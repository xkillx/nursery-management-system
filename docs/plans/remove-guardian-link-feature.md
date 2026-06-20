# Remove Guardian Link Feature Implementation Plan

## Goal

Remove the entire Guardian domain from the system: `guardians`, `guardian_child_links`, and `parent_membership_guardians` tables and their supporting code. Replace the two-hop parent-access model (`parent_membership → guardian → guardian_child_link → child`) with a one-hop model (`parent_membership → child`). Replace the enrollment-completeness `has_guardian_link` check with a `has_parent_carer_contact` check that reads the child's existing `child_contacts` row of type `parent_carer`. The actor in the staff surface keeps the same manager workflows they have today — only the data shape and a handful of routes change.

The end state lets a manager:

- Create a child and record at least one `parent_carer` contact, and the child is enrollment-complete.
- Invite a parent and (separately) attach the parent's membership to one or more children.
- Revoke a parent's access to one child by ending that mapping (with reason), while leaving the parent's other mappings intact.

The end state lets a parent:

- Sign in and view invoices only for children currently mapped to their parent membership.

## Non-Goals

- No new contact-relationship modeling beyond what `child_contacts` already supports. `parent_carer` / `emergency_contact` / `authorised_collector` stay as the only contact types.
- No preservation of historical rows from the dropped tables. The three tables are dropped; their data is not migrated (the pilot is not live).
- No change to the parent invitation flow itself. Parent invites are still created without a child reference; the child attachment is a separate step.
- No change to manager / practitioner / owner role permissions.
- No change to invoice calculation, payment processing, attendance, or funding logic.
- No GDPR Art. 15 export changes beyond dropping the now-removed endpoints.
- No change to `collection_password` (stays on the child, not the guardian).

## Context Alignment

The implementation must rewrite parts of `CONTEXT.md` to reflect the new model. The following existing glossary entries are **removed or rewritten**:

- "Guardian-Child Link" — entry removed. The concept no longer exists.
- "Parent Membership Guardian Mapping" — replaced with "Parent Membership Child Mapping".
- "Parent Mapping Change Flow" — rewritten to talk about child mappings, not guardian mappings.
- "Parent Membership End Cascade" — rewritten. Ending a parent membership ends all active child mappings, not the (removed) guardian mapping.
- "Parent Membership End Cascade Reason Attribution" — rewritten to apply to child mappings.
- "Parent Mapping End Visibility Rule" — rewritten for child mappings.
- "Parent Mapping Idempotent Create" — preserved in spirit, retargeted to child mappings.
- "Parent Mapping Active-Entity Requirement" — rewritten: requires the parent membership to be active and the child to exist in the same tenant+branch.
- "Parent-Guardian Email Independence" — removed (no guardian identity left).
- "Guardian Email Auto-Link Policy" — removed.
- "Parent Invite Mapping Separation" — removed.
- "Guardian Identity Separation" — removed.
- "Guardian Creation Minimum Data" — removed.
- "Contact Detail Scope" — preserved; clarifies that the child record now carries the full parent/contact set directly.
- "Child and Guardian Management Lifecycle" — renamed to "Child and Contact Management Lifecycle".
- "Guardian Link Lifecycle" — removed.
- "Guardian Deactivation Cascade" — removed.
- "Guardian Deactivation Idempotency" — removed.
- "Guardian Lifecycle Timestamp Semantics" — removed.
- "Guardian Deactivation Reason Requirement" — removed.
- "Deactivation Cascade Reason Attribution" — removed (last consumer was the guardian cascade).
- "Guardian Reactivation Policy" — removed.
- "Relationship End Reason Requirement" — preserved, retargeted to child mappings.
- "Relationship End Reason Shape" — preserved.
- "Lifecycle Reason Vocabulary" — preserved. The shared reason codes stay.
- "Lifecycle Other-Reason Note Requirement" — preserved.
- "Relationship End Terminology" — preserved; "end" still applies to parent-membership-child mappings.
- "Child Guardian Link Requirement Enforcement" — renamed to "Child Parent Carer Contact Requirement Enforcement". The blocker is now "no parent_carer contact" instead of "no active guardian link".
- "Guardian Link Reactivation" — removed.
- "Guardian Link Idempotent Create" — removed.
- "Guardian-Child Link Cardinality" — removed.
- "Parent Relationship Check Freshness" — rewritten to "Parent access to child-linked resources is authorized against current parent-membership-to-child mappings at request time."
- "Guardian Link End Visibility Rule" — rewritten to "When a parent-membership-to-child mapping is ended, that parent immediately loses access to that child's invoices, including historical invoices."
- "Guardian Relink Visibility Rule" — rewritten to "When a parent-membership-to-child mapping is recreated for a pair that previously had one, parent visibility for that child's invoices is restored based on the current active-mapping check."
- "Guardian-child links are valid only when guardian and child belong to the same tenant and branch scope." — removed.

A new entry is added:

- **Parent Membership Child Mapping**: A relationship showing that a parent-role membership in a given tenant+branch is attached to a specific child. A parent membership may have multiple active child mappings, and a child may have multiple parent memberships attached. Active mappings grant parent portal access for that child; ended mappings do not.

Two existing entries are kept verbatim:

- "Parent Account Provisioning" — parent user accounts are still created by manager invitation only.
- "Parent Role" — still describes a guardian-side role that views invoices and completes payments, but "guardian" here means the parent's relationship to a child through a parent-membership-child mapping, not a separate `guardians` row.

## Current State

Authoritative code locations (verified by exploration):

- `api/internal/modules/guardianlinks/` — the entire module: domain (`entities.go`, `repository.go`), application (`create_link.go`, `end_link.go`, `list_child_links.go`, `actor.go`), infrastructure (`postgres/repository.go`, `postgres/repository_test.go`), and interfaces (`http/handler.go`, `http/dto.go`). Routes registered: `POST /api/v1/guardian-child-links`, `POST /api/v1/guardian-child-links/:link_id/actions/end`, `GET /api/v1/children/:child_id/guardian-child-links`. (`api/internal/modules/guardianlinks/interfaces/http/handler.go:38-42`)
- `api/internal/modules/guardians/` — the entire module. Domain `Guardian` entity, deactivate/reactivate use cases, GET/LIST/CREATE/UPDATE. Routes: `GET/POST/PATCH/DELETE /api/v1/guardians[/:id]` and `POST /api/v1/guardians/:id/actions/deactivate|reactivate`. The `GuardiansCascadeLinks` and `GuardiansCascadeMappings` queries (`api/db/query/guardians.sql:89-109`) end `guardian_child_links` and `parent_membership_guardians` rows when a guardian is deactivated.
- `api/internal/modules/parentmappings/` — the entire module: domain (`entities.go`, `repository.go`), application (`create_mapping.go`, `end_mapping.go`, `actor.go`), infrastructure (postgres repo + test), interfaces (http handler + dto). Routes: `POST /api/v1/parent-membership-guardians`, `POST /api/v1/parent-membership-guardians/:id/actions/end`.
- `api/db/migrations/000001_baseline.up.sql:558-570` — `CREATE TABLE guardian_child_links`. `api/db/migrations/000001_baseline.up.sql:572-587` — `CREATE TABLE guardians`. `api/db/migrations/000001_baseline.up.sql:762-779` — `CREATE TABLE parent_membership_guardians`. Plus triggers `enforce_parent_membership_guardian_active`, `enforce_parent_membership_guardian_role`, `prevent_non_parent_with_active_guardian_mapping` at `000001_baseline.up.sql:78-145`.
- `api/db/query/guardian_child_links.sql` — all five queries consumed by the guardianlinks module.
- `api/db/query/children.sql` — every children list/get/attendance query uses `EXISTS (SELECT 1 FROM guardian_child_links gcl WHERE … AND gcl.ended_at IS NULL) AS has_guardian_link` at lines 20-27, 61-68, 121-128, 151-158. (`api/db/query/children.sql:20-27`)
- `api/db/query/guardians.sql:80-98` — `GuardiansCascadeLinks` updates `guardian_child_links` rows when a guardian is deactivated. `api/db/query/guardians.sql:100-109` — `GuardiansCascadeMappings` updates `parent_membership_guardians` rows.
- `api/db/query/invoices.sql` — joins `parent_membership_guardians pmg` then `guardian_child_links gcl` for parent invoice visibility at lines 499-509, 556-564, 585-593. Also lines 1389-1396, 1608-1615, 1923-1930 (parent-invoice queries).
- `api/db/query/payment_attempts.sql:14-24` — same two-hop join.
- `api/internal/platform/db/sqlc/guardian_child_links.sql.go` — generated by sqlc from `api/db/query/guardian_child_links.sql`.
- `api/internal/platform/db/sqlc/children.sql.go` — generated; defines `ChildrenList`, `ChildrenGetByID`, `ChildrenGetByIDForUpdate`, `ChildrenListAttendance` returning `has_guardian_link`. (`api/internal/platform/db/sqlc/children.sql.go:92`, `:173`, `:282`, `:379`)
- `api/internal/platform/db/sqlc/invoices.sql.go` — generated; all parent-invoice queries join `pmg` then `gcl`. (`api/internal/platform/db/sqlc/invoices.sql.go:65`, `:664`, `:886`, `:1117`, `:1389`, `:1608`, `:1923`)
- `api/internal/platform/db/sqlc/payment_attempts.sql.go:102` — same join.
- `api/internal/platform/db/sqlc/guardians.sql.go:15` — `UpdateGuardianChildLinksEnded` (cascade).
- `api/internal/platform/dbtest/dbtest.go:224` — truncates `guardian_child_links` between tests. `api/internal/platform/dbtest/dbtest.go:400` — inserts a `guardian_child_links` row in the test fixture.
- `api/cmd/seed/scenarios.sql:33-55` — seeds two guardians, two guardian-child links, one parent-membership-guardian mapping, plus a link-end-and-relink scenario.
- `api/internal/app/bootstrap/bootstrap.go:221-237` — wires the `guardianlinks` and `parentmappings` modules. `bootstrap.go:268-269` — registers their routes under the `manager` group.
- `api/internal/app/bootstrap/people_routes_test.go`, `billing_parent_routes_test.go`, `authorization_matrix_test.go` — bootstrap-level tests that touch `guardian_child_links` and `parent_membership_guardians` directly.
- `api/internal/modules/guardians/infrastructure/postgres/repository_test.go:314` — counts active links for a guardian.
- `api/internal/modules/guardianlinks/infrastructure/postgres/repository_test.go` — full test file for the dropped module.
- `web/src/app/features/staff/data/staff-api.service.ts:422-432` — `listChildGuardianLinks` and `createGuardianChildLink` methods. `staff-api.service.ts:434-452` — `listGuardians`, `createGuardian`, `updateGuardian` (and the matching `deactivateGuardian` is not shown in the slice but is present below).
- `web/src/app/features/staff/models/guardians.models.ts` — `ChildGuardianLinkRecord`, `GuardianRecord`, etc.
- `web/src/app/features/staff/pages/manager-guardians/` — entire manager-facing guardians page.
- `web/src/app/features/staff/components/guardian-form/` — entire form component.
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts:42` — `linkedGuardians: ChildGuardianLinkRecord[]`. `manager-child-detail.component.html:86-97` — "Linked guardians" section.
- `web/src/app/core/errors/api-error-presenter.ts:14, :81, :288` — `people.guardianLink` context and `guardian_child_link_not_found` code path.
- `docs/API-CONTRACT.openapi.yaml` — `Guardian-Child Links` tag, `GuardianChildLink` schema, three routes.
- `docs/API-SCHEMA-STATE.md` — references `idx_guardian_child_links_active_pair`.
- `docs/API_REVIEW.md` — sections 3.7, 4.6, 6.2, 7.4, 7.6, plus several others, discuss the guardian model. The implementing agent removes or rewrites the sections that recommend changes tied to the dropped model.
- `docs/adr/0002-two-hop-parent-access-authorization.md` — the ADR that codifies the two-hop model. This ADR is **superseded** and must be either deleted or annotated as superseded. The plan instructs the agent to delete it and create a replacement ADR-0008 describing the one-hop model.

## Decisions

### Product / domain (from the interview)

- A child is enrollment-complete when the child record has at least one `child_contacts` row of type `parent_carer` in the same tenant+branch. (Decision: "Require at least one parent_carer contact".)
- Parent memberships map directly to children. The `parent_membership → guardian → child` chain is collapsed to `parent_membership → child`. (Decision: "Map parent membership directly to child".)
- The `guardians` table, the `guardian_child_links` table, and the `parent_membership_guardians` table are all dropped. (Inferred from the chain collapse — the guardians table has no remaining product role once parent access is one-hop.)
- Existing rows in those three tables are not backfilled. The pilot is not live; tables are dropped in the same migration. (Decision: "Hard-delete is acceptable".)
- A single parent membership may map to many children. The mapping is created via a separate manager endpoint, after the child exists. Siblings share one parent account. (Decision: "Many children, create later".)
- Ending a parent membership ends all of its active child mappings in the same transaction, with a system-cascade reason. (Mirror of the existing rule; retargeted to the new table.)
- The end-with-reason lifecycle and shared reason vocabulary apply to the new `parent_membership_children` table the same way they did to `parent_membership_guardians`.
- "End" timestamps and the cascade-reason attribution are preserved verbatim, retargeted to the new table.

### Implementation (made by the agent)

- New table: `parent_membership_children` with columns `(id, tenant_id, branch_id, membership_id, child_id, created_at, updated_at, ended_at, ended_reason_code, ended_reason_note)`. The reason-code check constraint mirrors the existing `parent_membership_guardians_end_reason_check`. A partial unique index on `(tenant_id, branch_id, membership_id, child_id) WHERE ended_at IS NULL` enforces one active mapping per pair.
- New triggers on `parent_membership_children`:
  - `enforce_parent_membership_child_role` — the membership must be role `parent` and active on insert.
  - `enforce_parent_membership_child_scope` — the child must exist in the same tenant+branch on insert.
  - `cascade_parent_membership_child_end` — when `memberships` is ended, the trigger ends all active `parent_membership_children` rows for that membership with `ended_reason_code = 'system_cascade'`.
- New module: `api/internal/modules/parentchildmappings/` mirroring the structure of the dropped `parentmappings` module. Routes: `POST /api/v1/parent-membership-children` (idempotent create) and `POST /api/v1/parent-membership-children/:id/actions/end` (manager-initiated end with reason).
- `api/internal/modules/guardians/` and `api/internal/modules/guardianlinks/` are deleted in their entirety.
- The web `manager-guardians` page and `guardian-form` component are deleted. The "Linked guardians" section in the manager child detail page is replaced with a "Parent carers" section that lists the child's `parent_carer` contacts (which already exist in the child profile).
- The `has_guardian_link` field in `ChildrenList` / `ChildrenGetByID` / `ChildrenGetByIDForUpdate` / `ChildrenListAttendance` is replaced with `has_parent_carer_contact` in both SQL and generated Go. The corresponding `Child.HasGuardianLink` domain field is replaced with `HasParentCarerContact`. The `MissingRequirements()` method drops `guardian_link` and adds `parent_carer_contact`. `EnrollmentComplete()` is unchanged in shape.
- All invoice and payment-attempt queries that previously joined `pmg` then `gcl` are rewritten to join `parent_membership_children pmc` directly against `invoices.child_id`. The membership role/active check stays the same.
- `docs/adr/0002-two-hop-parent-access-authorization.md` is deleted. A new `docs/adr/0008-one-hop-parent-access-authorization.md` is created, mirroring the original ADR's structure but describing the one-hop model. (ADR test: hard to reverse because every parent-facing query depends on it; surprising without context because a future reader will ask why parent access is structured this way; the result of a real trade-off — one-hop trades guardian-identity indirection for simpler queries and direct manager control.)
- `CONTEXT.md` is rewritten per the "Context Alignment" section. The implementation does not need a separate change-log entry; the glossary entry replacements ARE the change log.
- The seed file is updated to drop guardian inserts and use the new `parent_membership_children` table.
- The bootstrap test fixtures that insert `guardian_child_links` rows are updated to insert `parent_membership_children` rows instead.

### Existing decisions honored

- ADR-0001 session-bound authorization guards still apply.
- ADR-0005 / ADR-0006 booking-pattern decisions unaffected.
- ADR-0007 12-month fixed-term contract unaffected.
- AGENTS.md architecture rules (handler → application → domain, no cross-module imports, txMgr.ExecTx, tenant.ActorFromGinContext, MapDomainError) are preserved.

## Acceptance Criteria

1. Running `make migrate-up` against a database that has the baseline schema applies a new migration that drops `guardians`, `guardian_child_links`, and `parent_membership_guardians`, and creates `parent_membership_children` with its indexes and triggers. `make migrate-down` reverses this safely (drops `parent_membership_children`; the three dropped tables are not recreated, but the migration is reversible in the sense that the database is left in a consistent state).
2. The `api/internal/modules/guardians/`, `api/internal/modules/guardianlinks/`, and `api/internal/modules/parentmappings/` directories no longer exist in the repository.
3. `grep -ri "guardian_child_links\|guardians" --include="*.go" --include="*.sql" --include="*.ts"` returns zero results outside of `docs/adr/0008-one-hop-parent-access-authorization.md` and historical references in `docs/API_REVIEW.md` and `docs/API-SCHEMA-STATE.md` that the agent updates as part of the change. (The plan instructs the agent to update those docs to remove the dropped model.)
4. `go build ./...` succeeds in the `api/` directory.
5. `npm run build` (or `ng build`) succeeds in `web/`.
6. `make test-api-repositories` passes. The test fixtures insert `parent_membership_children` rows; the children/invoices/payments tests no longer reference the dropped tables.
7. A child with at least one `child_contacts` row of type `parent_carer` returns `enrollment_complete = true` and is included in the attendance list. A child with zero `parent_carer` contacts returns `enrollment_complete = false` and is blocked from check-in and draft-invoice generation (matching the prior `has_guardian_link = false` behavior).
8. A parent-role user with an active `parent_membership_children` mapping to child X can call `GET /api/v1/parent/invoices` and see child X's issued-or-later invoices. A parent with no active mapping to child X cannot see child X's invoices. Ending a mapping with a reason removes access at the next request. Recreating a mapping restores access. The session-membership scope (tenant+branch) check still applies.
9. `POST /api/v1/parent-membership-children` is idempotent for the same `(membership_id, child_id)` pair and returns success without duplicating the row. Mapping the same parent membership to a different child while one is already active is allowed (no `ErrActiveConflict`); each child is an independent mapping.
10. The manager child detail page shows a "Parent carers" section that lists the child's `parent_carer` contacts, with name and (if present) email/phone. The "Linked guardians" section is removed.
11. The manager guardians page (`/staff/manager/guardians`) and the guardian form component are removed from the route table and from the navigation. Any menu / breadcrumb / link references to them in `web/src/app/app.routes.ts`, sidebar, or other layout files are removed.
12. The web `staffApi.listChildGuardianLinks` and `staffApi.createGuardianChildLink` methods are removed. The `ChildGuardianLinkRecord` model and `people.guardianLink` error context are removed. The `guardian_child_link_not_found` error-code branch in `api-error-presenter.ts` is removed.
13. The `api-error-presenter` and its spec no longer reference the removed error codes.
14. The OpenAPI spec in `docs/API-CONTRACT.openapi.yaml` removes the `Guardian-Child Links` tag, the `GuardianChildLink` schema, and the three routes. It adds the new `Parent-Child Mappings` tag, the `ParentChildMapping` schema, and the two new routes.
15. `docs/API_REVIEW.md` is updated: sections 3.7, 4.6, 6.2, 7.4, 7.6 are removed or rewritten to reflect the dropped model.
16. `docs/API-SCHEMA-STATE.md` drops the `idx_guardian_child_links_active_pair` row and adds the new `idx_parent_membership_children_active_pair` row.
17. `CONTEXT.md` reflects the new domain language per the "Context Alignment" section. The "Guardian-Child Link", "Parent Membership Guardian Mapping", and "Parent Mapping …" entries are gone; new "Parent Membership Child Mapping" and "Child Parent Carer Contact Requirement Enforcement" entries are present.
18. The `docs/adr/0002-two-hop-parent-access-authorization.md` ADR is deleted. The new `docs/adr/0008-one-hop-parent-access-authorization.md` ADR exists and describes the one-hop model.
19. The `audit` writer continues to log manager-initiated child-mapping create/end actions with the same `audit_action_type` vocabulary where possible (`parent_mapping_created` and `parent_mapping_ended` are reused; `entity_type` changes to `parent_membership_child_mapping`). The `actor`, `tenant_id`, `branch_id`, `request_id` propagation is unchanged.
20. The seed command (`api/cmd/seed`) succeeds and creates one parent membership, one child with one `parent_carer` contact, and one active `parent_membership_children` mapping.

## Implementation Tasks

### Task 1: Database migration

- **Objective**: Drop the three guardian-related tables and add the `parent_membership_children` table with its indexes and triggers.
- **Depends on**: none.
- **Target files / symbols**: new file `api/db/migrations/000004_drop_guardian_domain.up.sql` and matching `000004_drop_guardian_domain.down.sql`. The agent must run `make migrate-create name=drop_guardian_domain` to generate the filenames.
- **Required changes**:
  - In the up migration, in this order: drop the three triggers that depend on the old tables (`enforce_parent_membership_guardian_active`, `enforce_parent_membership_guardian_role`, `prevent_non_parent_with_active_guardian_mapping`); drop `parent_membership_guardians`; drop `guardian_child_links`; drop `guardians`. Then `CREATE TABLE parent_membership_children` mirroring the schema described in "Decisions / Implementation".
  - Add the `parent_membership_children_end_reason_check` constraint matching the `lifecycle_reason_code` enum and the `other`-note requirement.
  - Add the partial unique index `idx_parent_membership_children_active_pair ON parent_membership_children (tenant_id, branch_id, membership_id, child_id) WHERE ended_at IS NULL`.
  - Add the three new triggers listed under "Decisions / Implementation".
  - In the down migration, drop the new triggers, drop the new table, and add a `RAISE EXCEPTION` noting that the three dropped tables are not recreated. (Reversibility note: a down migration that does not recreate the dropped tables is acceptable here because the new table is the source of truth and the old data has been intentionally discarded.)
- **Tests / verification**: `make migrate-up` against a local PostgreSQL test database succeeds; `make migrate-down` succeeds; `make migrate-up` again succeeds.
- **Expected outcome**: Schema is in the new state; the three old tables are gone; the new table exists with its indexes and triggers.

### Task 2: Delete the `guardianlinks` module

- **Objective**: Remove the entire `api/internal/modules/guardianlinks/` directory and all references to it.
- **Depends on**: Task 1 (the migration drops the table the module read from).
- **Target files / symbols**:
  - Delete: `api/internal/modules/guardianlinks/` (all subdirectories and files).
  - Edit: `api/internal/app/bootstrap/bootstrap.go` — remove the `linkapp`, `linkpostgres`, `linkhandler` imports (lines 29-31); remove the wiring block at lines 221-229; remove `linksHandler.RegisterRoutes(manager)` at line 268.
  - Edit: `api/internal/platform/db/sqlc/` — delete `guardian_child_links.sql.go`; regenerate via `make sqlc-generate`.
  - Edit: `api/db/query/` — delete `guardian_child_links.sql`.
  - Edit: `api/internal/platform/dbtest/dbtest.go` — remove `guardian_child_links` from the truncate list (line 224) and remove the insert at line 400.
- **Required changes**: Per the file list above.
- **Tests / verification**: `go build ./...` in `api/`. `go test ./internal/app/bootstrap/...` (the people-routes and authorization-matrix tests are updated by later tasks).
- **Expected outcome**: No references to the `guardianlinks` package anywhere in `api/`. `go build` succeeds.

### Task 3: Delete the `guardians` module

- **Objective**: Remove the entire `api/internal/modules/guardians/` directory.
- **Depends on**: Task 1.
- **Target files / symbols**:
  - Delete: `api/internal/modules/guardians/` (all subdirectories and files).
  - Edit: `api/internal/app/bootstrap/bootstrap.go` — remove the `guardianapp`, `guardianpostgres`, `guardianhandler` imports (lines 25-27); remove the wiring block at lines 210-219; remove `guardiansHandler.RegisterRoutes(manager)` at line 267.
  - Edit: `api/db/query/guardians.sql` — delete the entire file (every query in it is tied to the dropped `guardians` table or to the `GuardiansCascadeLinks` / `GuardiansCascadeMappings` cascades that target dropped tables). The `guardian_active` lookup the new module needs is replaced by a `child_contacts` existence check in Task 5.
  - Edit: `api/internal/platform/db/sqlc/` — regenerate (the `guardians.sql.go` file disappears).
  - Edit: `api/internal/platform/db/sqlc/children.sql.go` — the rows using `has_guardian_link` are replaced in Task 5; this task only deletes the guardian-generated file.
- **Required changes**: Per the file list above.
- **Tests / verification**: `go build ./...`. `go vet ./...`.
- **Expected outcome**: No references to the `guardians` package anywhere in `api/`. `go build` succeeds.

### Task 4: Delete the `parentmappings` module and add the `parentchildmappings` module

- **Objective**: Remove `parentmappings` and replace it with `parentchildmappings` that targets the new `parent_membership_children` table.
- **Depends on**: Task 1.
- **Target files / symbols**:
  - Delete: `api/internal/modules/parentmappings/` (all subdirectories and files).
  - Create: `api/internal/modules/parentchildmappings/` with the structure:
    - `domain/entities.go` — `ParentChildMapping` struct (id, tenant_id, branch_id, membership_id, child_id, ended_at, ended_reason_code, ended_reason_note, created_at, updated_at).
    - `domain/repository.go` — `Repository` interface with `FindActiveByPair(ctx, tx, tenant, branch, membership, child) (ParentChildMapping, bool, error)`, `ListActiveByMembership(ctx, tx, tenant, branch, membership) ([]ParentChildMapping, error)`, `Create(ctx, tx, mapping) error`, `GetByIDForUpdate(ctx, tx, tenant, branch, id) (ParentChildMapping, bool, error)`, `End(ctx, tx, tenant, branch, id, reasonCode, reasonNote) error`.
    - `application/actor.go` — `ActorContext` struct (mirror the old one).
    - `application/create_mapping.go` — `CreateMappingUseCase` and `CreateMappingParams { TenantID, BranchID, MembershipID, ChildID }`. Validates membership role=`parent` and `is_active=true`; validates child exists in scope; calls `FindActiveByPair` for idempotency; calls `Create`; writes audit `parent_mapping_created` with `entity_type=parent_membership_child_mapping` and `Details={membership_id, child_id}`. Errors: `ErrMembershipNotFound`, `ErrMembershipNotParent`, `ErrMembershipNotActive`, `ErrChildNotFound`. The `MembershipChecker` interface lives in the new module.
    - `application/end_mapping.go` — `EndMappingUseCase.Execute(ctx, actor, mappingID, reasonCode, reasonNote)`. Mirror the old behavior: end if not already ended, write audit `parent_mapping_ended`. Error: `ErrMappingNotFound`.
    - `application/list_membership_mappings.go` — `ListMembershipMappingsUseCase` returning active mappings for a given membership. Optional, but needed for the manager "this parent has these children" view. If the implementing agent chooses to skip the UI listing for now, this task can stop at create/end. The plan recommends including it.
    - `infrastructure/postgres/repository.go` — sqlc-backed implementation. `infrastructure/postgres/repository_test.go` — covers the same scenarios as the dropped `parentmappings` test.
    - `interfaces/http/dto.go` and `interfaces/http/handler.go` — `RegisterRoutes` exposes `POST /api/v1/parent-membership-children` and `POST /api/v1/parent-membership-children/:id/actions/end` (and optionally `GET /api/v1/parent-membership-children?membership_id=…`).
  - Edit: `api/db/query/parent_membership_children.sql` — new file with the sqlc queries matching the new repo.
  - Edit: `api/internal/app/bootstrap/bootstrap.go` — remove the old `mappingapp`/`mappingpostgres`/`mappinghandler` imports (lines 33-35); remove the wiring block at lines 231-237; remove `mappingsHandler.RegisterRoutes(manager)` at line 269. Add the new module's wiring block (mirror the dropped one with `parentchildapp`/`parentchildpostgres`/`parentchildhandler`) and call `mappingsHandler.RegisterRoutes(manager)` for the new handler.
  - Edit: `api/internal/app/bootstrap/adapters.go` — remove the `membershipCheckerAdapter` (it depended on the old repo's `GetForScope` query, which the new repo must also expose or be replaced). The new `membershipChecker` lives inside the new module. If the old `membershipCheckerAdapter` was used elsewhere, the agent replaces it with a new adapter targeting the new repo.
- **Required changes**: Per the file list above.
- **Tests / verification**:
  - `make sqlc-generate` runs without errors.
  - `go build ./...` succeeds.
  - `go test ./internal/modules/parentchildmappings/...` covers: idempotent create, end with reason, end-of-already-ended no-op, membership-not-parent error, child-not-found error.
- **Expected outcome**: The new module is fully wired; the old one is gone; the `parent_membership_children` SQL queries are exercisable from the API.

### Task 5: Update children module to use `has_parent_carer_contact`

- **Objective**: Replace the `has_guardian_link` SQL flag and Go field with `has_parent_carer_contact`, reading from `child_contacts`.
- **Depends on**: Task 1.
- **Target files / symbols**:
  - `api/db/query/children.sql` — every occurrence of `EXISTS (SELECT 1 FROM guardian_child_links gcl …) AS has_guardian_link` (lines 20-27, 61-68, 121-128, 151-158) becomes `EXISTS (SELECT 1 FROM child_contacts cc WHERE cc.tenant_id = c.tenant_id AND cc.branch_id = c.branch_id AND cc.child_id = c.id AND cc.contact_type = 'parent_carer') AS has_parent_carer_contact`.
  - `api/internal/platform/db/sqlc/children.sql.go` — regenerated; verify the new field name appears.
  - `api/internal/modules/children/domain/child.go` — rename `HasGuardianLink` to `HasParentCarerContact`. Update `MissingRequirements()` to check `HasParentCarerContact` and return the missing-requirement code `parent_carer_contact` (the existing code `guardian_link` is replaced).
  - `api/internal/modules/children/infrastructure/postgres/repository.go` — update the row mapping to use the new field name.
  - `api/internal/modules/children/infrastructure/postgres/repository_test.go` — update assertions to use the new flag name and to seed `child_contacts` rows in test setup where required.
- **Required changes**: Per the file list above.
- **Tests / verification**:
  - `make sqlc-generate` runs cleanly.
  - `go build ./...` succeeds.
  - `make test-api-repositories` passes. Tests for attendance eligibility and invoice preflight confirm that a child without a `parent_carer` contact is blocked from check-in and from draft-invoice generation.
- **Expected outcome**: A child with at least one `parent_carer` contact is enrollment-complete; one without is not. The behavior matches the previous `has_guardian_link` behavior in every observable way.

### Task 6: Rewrite invoice and payment-attempt authorization queries

- **Objective**: Replace the two-hop `pmg` + `gcl` joins in invoice and payment queries with a one-hop `pmc` join.
- **Depends on**: Task 4 (the new table must exist).
- **Target files / symbols**:
  - `api/db/query/invoices.sql` — every query that previously joined `parent_membership_guardians pmg … AND pmg.ended_at IS NULL` and then `guardian_child_links gcl … AND gcl.ended_at IS NULL` (lines 499-509, 556-564, 585-593, and the parent-invoice queries at lines 1389-1396, 1608-1615, 1923-1930) becomes a single join `JOIN parent_membership_children pmc ON pmc.tenant_id = i.tenant_id AND pmc.branch_id = i.branch_id AND pmc.membership_id = m.id AND pmc.child_id = i.child_id AND pmc.ended_at IS NULL`.
  - `api/db/query/payment_attempts.sql` — same rewrite at lines 14-24.
  - `api/internal/platform/db/sqlc/invoices.sql.go` — regenerated; the Go function signatures may change (the `parent_mapping_id` parameter or join result column changes). The agent updates any Go call site accordingly.
  - `api/internal/platform/db/sqlc/payment_attempts.sql.go` — regenerated.
  - `api/internal/modules/billing/infrastructure/postgres/repository.go` and `api/internal/modules/payments/infrastructure/postgres/repository.go` — call sites updated to match the new generated signatures.
- **Required changes**: Per the file list above.
- **Tests / verification**:
  - `make sqlc-generate` runs cleanly.
  - `go build ./...` succeeds.
  - `make test-api-repositories` passes. The `billing_parent_routes_test.go` (lines 415 area) is updated to use `parent_membership_children` rows in the test fixture.
  - Manual scenario: a parent with an active mapping to child X sees child X's invoices; ending the mapping removes access; recreating the mapping restores access.
- **Expected outcome**: Parent invoice access uses the one-hop model. The visible behavior to a parent (sees invoices for mapped children, not for others) is unchanged.

### Task 7: Update bootstrap wiring and tests

- **Objective**: Remove the wiring for the dropped modules and ensure all bootstrap tests pass.
- **Depends on**: Tasks 2, 3, 4, 5, 6.
- **Target files / symbols**:
  - `api/internal/app/bootstrap/bootstrap.go` — final wiring pass: only the new `parentchildmappings` handler is registered under the `manager` group; no `guardiansHandler` or `linksHandler` calls remain.
  - `api/internal/app/bootstrap/people_routes_test.go` — every reference to `guardian_child_links`, `guardians`, `parent_membership_guardians` is replaced with the new equivalents (or removed if the test was specifically about a dropped behavior). The test that checks `guardian-child-links` end-and-relink at line 677 is removed (the behavior is gone). The cascade tests at line 1034 are removed.
  - `api/internal/app/bootstrap/billing_parent_routes_test.go` — the `UPDATE guardian_child_links SET ended_at = now()` at line 415 becomes an `UPDATE parent_membership_children SET ended_at = now()`.
  - `api/internal/app/bootstrap/authorization_matrix_test.go` — the `INSERT INTO guardian_child_links` test fixture at line 1034 is replaced with an `INSERT INTO parent_membership_children` fixture where the authorization matrix now relies on the new table.
  - `api/internal/app/bootstrap/billing_routes_test.go`, `payments_routes_test.go`, `manager_payment_diagnostics_test.go`, `webhook_integration_test.go` — searched for stale references; updated where present.
- **Required changes**: Per the file list above.
- **Tests / verification**:
  - `make test-api-repositories` passes.
  - `go test ./internal/app/bootstrap/...` passes.
  - `go test ./...` passes.
- **Expected outcome**: All Go tests pass; the bootstrap test fixtures no longer reference dropped tables; the authorization matrix tests cover the new one-hop model.

### Task 8: Update the seed data

- **Objective**: Remove the guardian seed rows and use the new `parent_membership_children` table.
- **Depends on**: Task 1.
- **Target files / symbols**: `api/cmd/seed/scenarios.sql`. Remove the `INSERT INTO guardians`, `INSERT INTO guardian_child_links`, and `INSERT INTO parent_membership_guardians` blocks (lines 33-55). Replace the `parent_membership_guardians` insert with an `INSERT INTO parent_membership_children` row linking the parent membership to the seed child.
- **Required changes**: Per above.
- **Tests / verification**: `cd api && set -a && source .env && set +a && go run ./cmd/seed -email o@x.com -password 'X' -local -manager-email m@x.com -staff-email s@x.com -parent-email p@x.com` succeeds. `make test-api-repositories` passes.
- **Expected outcome**: The seed produces a child with one `parent_carer` contact (which the agent adds as part of this task) and a parent membership with one active child mapping. The seed child is enrollment-complete.

### Task 9: Remove web references to the guardian-link feature

- **Objective**: Delete the manager-guardians page, the guardian-form component, the `ChildGuardianLinkRecord` model, and the `listChildGuardianLinks` / `createGuardianChildLink` API methods. Replace the "Linked guardians" section in the manager child detail page with a "Parent carers" section.
- **Depends on**: Tasks 5, 6, 7 (the API surface must be stable).
- **Target files / symbols**:
  - Delete: `web/src/app/features/staff/pages/manager-guardians/` and `web/src/app/features/staff/components/guardian-form/`.
  - Edit: `web/src/app/app.routes.ts` — remove the `staff/manager/guardians` route (lines 126-135). Remove the `ManagerGuardiansComponent` import (line 10).
  - Edit: `web/src/app/features/staff/data/staff-api.service.ts` — remove `listChildGuardianLinks` (lines 422-426) and `createGuardianChildLink` (lines 428-432). Update `getChildFullProfile` to drop the call to `listChildGuardianLinks`.
  - Edit: `web/src/app/features/staff/models/guardians.models.ts` — remove `ChildGuardianLinkRecord`, `ChildGuardianLinkApiModel`, `GuardianChildLinkWritePayload`, and the `linkedGuardians` projection. Keep `GuardianRecord` and the contact-related types if any survive the rewrite; otherwise trim to a smaller file.
  - Edit: `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts` — remove `linkedGuardians` (line 42), the import of `ChildGuardianLinkRecord` (line 15), and the `listChildGuardianLinks` subscription (line 73-80). Add a `parentCarers` array sourced from the existing `getChildContacts` response (filter `contactType === 'parent_carer'`).
  - Edit: `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.html` — replace the "Linked guardians" section (lines 86-97) with a "Parent carers" section that lists `parentCarers` and shows name, email, phone, and `hasParentalResponsibility` if present.
  - Edit: `web/src/app/core/errors/api-error-presenter.ts` — remove the `people.guardianLink` context (line 14), the `guardian_child_link_not_found` code (line 81), and the switch branch (line 288). Update the `people` context union to remove the `guardianLink` member.
  - Edit: `web/src/app/core/errors/api-error-presenter.spec.ts` — remove the test cases at lines 216-222.
  - Edit: any sidebar / nav component that links to `/staff/manager/guardians` (search the web tree).
  - Edit: `web/src/app/features/staff/utils/invoice-run-formatters.ts` and `invoice-run-formatters.spec.ts` — remove any guardian-link reference (the grep at the start found one; the agent inspects and removes it).
  - Edit: `web/src/app/features/staff/utils/manager-list-formatters.ts` and `manager-list-formatters.spec.ts` — same.
  - Edit: `web/src/app/features/staff/models/guardians.models.ts` — keep only the contact-related types that are still used (e.g. `ChildContact`); the rest goes away.
- **Required changes**: Per the file list above.
- **Tests / verification**:
  - `cd web && npm run build` succeeds.
  - `cd web && npm test` passes (Karma) — the `manager-child-detail.component.spec.ts` and `manager-children.component.spec.ts` are updated for the new flag name and the removed API methods.
  - Manual: the manager child detail page renders the "Parent carers" section; the manager guardians route returns 404 (route removed).
- **Expected outcome**: No web code references the dropped model. The UI presents the new "Parent carers" section.

### Task 10: Update docs

- **Objective**: Update `CONTEXT.md`, `docs/API-CONTRACT.openapi.yaml`, `docs/API-SCHEMA-STATE.md`, `docs/API_REVIEW.md`, and the ADRs to reflect the new model.
- **Depends on**: Tasks 1-9.
- **Target files / symbols**:
  - `CONTEXT.md` — apply every change listed in the "Context Alignment" section above. Every "guardian" mention that referred to the dropped model is gone.
  - `docs/API-CONTRACT.openapi.yaml` — remove the `Guardian-Child Links` tag, the `GuardianChildLink` schema, and the three routes. Add the `Parent-Child Mappings` tag, the `ParentChildMapping` schema, and the two new routes (`POST /api/v1/parent-membership-children`, `POST /api/v1/parent-membership-children/{id}/actions/end`).
  - `docs/API-SCHEMA-STATE.md` — drop the `idx_guardian_child_links_active_pair` row; add `idx_parent_membership_children_active_pair` on `parent_membership_children`.
  - `docs/API_REVIEW.md` — remove or rewrite sections 3.7 (phone normalisation on guardians becomes "phone normalisation on child_contacts"), 4.6 (concurrent active guardian-child links — irrelevant now), 6.2 (guardian deactivation cascade — irrelevant now), 7.4 (one-guardian-per-membership — irrelevant now), 7.6 (branches.is_active cascade — kept but verified). The `RegisterParentRoutes` recommendation at line 484 is updated to reflect the one-hop model.
  - `docs/adr/0002-two-hop-parent-access-authorization.md` — delete.
  - `docs/adr/0008-one-hop-parent-access-authorization.md` — create, mirroring the structure of ADR-0002 but describing the one-hop model and the rationale: parent-membership-to-child mappings give direct, immediate, manager-controllable access; identity is now a property of the parent membership, not a separate guardian record.
- **Required changes**: Per above.
- **Tests / verification**: Manual inspection. `grep -ri "guardian_child_link" docs/` returns zero results outside of historical context the agent chose to retain. `grep -ri "two-hop" docs/adr/` returns only the new ADR-0008 (ADR-0002 is gone).
- **Expected outcome**: The docs accurately describe the new domain.

### Task 11: Final verification

- **Objective**: Run the full test suite and a manual end-to-end scenario.
- **Depends on**: All prior tasks.
- **Target files / symbols**: N/A.
- **Required changes**: None.
- **Tests / verification**:
  - `make migrate-up` against a fresh database: clean apply.
  - `make sqlc-generate`: clean.
  - `go build ./...` in `api/`: clean.
  - `make test-api-repositories`: passes.
  - `go test ./...` in `api/`: passes.
  - `cd web && npm run build`: clean.
  - `cd web && npm test`: passes.
  - Manual: start the API and the web app. Sign in as a manager. Create a child with a `parent_carer` contact; confirm `enrollmentComplete = true`. Create a child without one; confirm `enrollmentComplete = false`. Invite a parent. Map the parent membership to a child. Switch to a parent-role browser, sign in, confirm the child appears in the invoice list. End the mapping; refresh; confirm the child disappears. Create a second mapping; confirm the child reappears.
- **Expected outcome**: The system behaves as described in the acceptance criteria.

## Contracts

### Database

- `parent_membership_children` (new): `(id uuid PK, tenant_id uuid, branch_id uuid, membership_id uuid FK→memberships.id, child_id uuid FK→children.id, created_at timestamptz, updated_at timestamptz, ended_at timestamptz NULL, ended_reason_code lifecycle_reason_code NULL, ended_reason_note text NULL)` with a `parent_membership_children_end_reason_check` constraint matching the existing pattern and a partial unique index `(tenant_id, branch_id, membership_id, child_id) WHERE ended_at IS NULL`.
- `parent_membership_guardians` (dropped), `guardian_child_links` (dropped), `guardians` (dropped).

### API

Removed routes:

- `POST /api/v1/guardian-child-links`
- `POST /api/v1/guardian-child-links/:link_id/actions/end`
- `GET /api/v1/children/:child_id/guardian-child-links`
- `GET /api/v1/guardians`, `POST /api/v1/guardians`, `GET /api/v1/guardians/:id`, `PATCH /api/v1/guardians/:id`, `POST /api/v1/guardians/:id/actions/deactivate`, `POST /api/v1/guardians/:id/actions/reactivate`
- `POST /api/v1/parent-membership-guardians`, `POST /api/v1/parent-membership-guardians/:id/actions/end`

New routes (manager-only):

- `POST /api/v1/parent-membership-children` — body `{ membership_id, child_id }`. Returns 201 with the new mapping, or 200 if an active mapping for the same pair already exists (idempotent). Errors: `membership_not_found`, `membership_not_parent`, `membership_not_active`, `child_not_found`.
- `POST /api/v1/parent-membership-children/:id/actions/end` — body `{ reason_code, reason_note? }`. Returns 200 with the updated mapping. Errors: `mapping_not_found`, `lifecycle_reason_invalid`, `reason_note_required_for_other` (only when `reason_code='other'`).
- `GET /api/v1/parent-membership-children?membership_id=…` — returns active child mappings for a parent membership. (Optional but recommended for the manager view.)

Unchanged routes:

- `GET /api/v1/parent/invoices`, `GET /api/v1/parent/invoices/:id` — authorization now uses the new one-hop join. Visible behavior is unchanged for parents.
- All children, attendance, funding, billing, payment, invites, authentication routes.

### Audit

- `parent_mapping_created` (action_type) on `parent_membership_child_mapping` (entity_type) — `Details: { membership_id, child_id }`.
- `parent_mapping_ended` (action_type) on `parent_membership_child_mapping` (entity_type) — `ReasonCode`, `ReasonNote`, `Details: {}`.
- The audit writer's `actor`, `tenant_id`, `branch_id`, `request_id` propagation is unchanged.

### Permissions

- `parent-membership-children` create/end/list: manager-only, branch-scoped.
- `parent/invoices` and `parent/invoices/:id`: parent-only; the handler authorizes against the new one-hop join at request time.

### Error codes

- `membership_not_found` (404)
- `membership_not_parent` (400)
- `membership_not_active` (400)
- `child_not_found` (404)
- `parent_child_mapping_not_found` (404) — replaces `guardian_child_link_not_found`
- `lifecycle_reason_invalid` (400)
- `reason_note_required_for_other` (400)
- `relationship_reason_required` (400) — reused for the end action

## Files to Change

Deleted:

- `api/internal/modules/guardianlinks/` (entire directory)
- `api/internal/modules/guardians/` (entire directory)
- `api/internal/modules/parentmappings/` (entire directory)
- `api/internal/platform/db/sqlc/guardian_child_links.sql.go`
- `api/internal/platform/db/sqlc/guardians.sql.go`
- `api/db/query/guardian_child_links.sql`
- `api/db/query/guardians.sql`
- `web/src/app/features/staff/pages/manager-guardians/`
- `web/src/app/features/staff/components/guardian-form/`
- `docs/adr/0002-two-hop-parent-access-authorization.md`

Created:

- `api/db/migrations/000004_drop_guardian_domain.up.sql` (and matching `.down.sql`) — exact name determined by `make migrate-create name=drop_guardian_domain`
- `api/internal/modules/parentchildmappings/` (new module: `domain/entities.go`, `domain/repository.go`, `application/actor.go`, `application/create_mapping.go`, `application/end_mapping.go`, `application/list_membership_mappings.go`, `infrastructure/postgres/repository.go`, `infrastructure/postgres/repository_test.go`, `interfaces/http/dto.go`, `interfaces/http/handler.go`)
- `api/db/query/parent_membership_children.sql`
- `docs/adr/0008-one-hop-parent-access-authorization.md`

Modified:

- `api/db/migrations/000001_baseline.down.sql` — remove the `DROP TABLE IF EXISTS guardian_child_links` line and the equivalent guardian lines so that a down migration of the new 000004 file leaves the schema in a consistent state. (The agent makes this surgical edit; the up migration in 000001 stays as historical record.)
- `api/db/query/children.sql` — replace `has_guardian_link` with `has_parent_carer_contact`
- `api/db/query/invoices.sql` — replace `pmg` + `gcl` joins with `pmc`
- `api/db/query/payment_attempts.sql` — replace `pmg` + `gcl` join with `pmc`
- `api/cmd/seed/scenarios.sql` — drop guardian inserts; add `parent_membership_children` insert; add a `child_contacts` row of type `parent_carer` for the seed child
- `api/internal/app/bootstrap/bootstrap.go` — remove dropped-module imports and wiring; add the new module's wiring
- `api/internal/app/bootstrap/adapters.go` — remove or replace `membershipCheckerAdapter`, `guardianCheckerAdapter`, `childCheckerAdapter` entries that targeted the dropped repos
- `api/internal/app/bootstrap/people_routes_test.go` — remove or replace dropped-model references
- `api/internal/app/bootstrap/billing_parent_routes_test.go` — replace `UPDATE guardian_child_links` with `UPDATE parent_membership_children`
- `api/internal/app/bootstrap/authorization_matrix_test.go` — replace `INSERT INTO guardian_child_links` fixture
- `api/internal/platform/dbtest/dbtest.go` — remove dropped tables from the truncate list and from the fixture inserts
- `api/internal/modules/children/domain/child.go` — rename field, update `MissingRequirements()`
- `api/internal/modules/children/infrastructure/postgres/repository.go` — update field mapping
- `api/internal/modules/children/infrastructure/postgres/repository_test.go` — update field references
- `api/internal/modules/billing/infrastructure/postgres/repository.go` — call sites for the new sqlc signatures
- `api/internal/modules/payments/infrastructure/postgres/repository.go` — same
- `web/src/app/app.routes.ts` — remove the manager-guardians route
- `web/src/app/features/staff/data/staff-api.service.ts` — remove `listChildGuardianLinks` and `createGuardianChildLink`; update `getChildFullProfile`
- `web/src/app/features/staff/models/guardians.models.ts` — remove `ChildGuardianLinkRecord` and related types
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts` — replace `linkedGuardians` with `parentCarers`
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.html` — replace the "Linked guardians" section with "Parent carers"
- `web/src/app/core/errors/api-error-presenter.ts` — remove `people.guardianLink` and `guardian_child_link_not_found`
- `web/src/app/core/errors/api-error-presenter.spec.ts` — remove the related test cases
- `web/src/app/features/staff/utils/invoice-run-formatters.ts` and `.spec.ts` — remove any guardian-link reference
- `web/src/app/features/staff/utils/manager-list-formatters.ts` and `.spec.ts` — same
- `web/src/app/features/staff/pages/manager-children/manager-children.component.spec.ts` — update flag name in any assertion
- `CONTEXT.md` — apply the rewrites in "Context Alignment"
- `docs/API-CONTRACT.openapi.yaml` — replace the guardian-link routes and schema with the new ones
- `docs/API-SCHEMA-STATE.md` — update the index list
- `docs/API_REVIEW.md` — remove or rewrite the dropped-model sections

Generated (via `make sqlc-generate`):

- `api/internal/platform/db/sqlc/children.sql.go`
- `api/internal/platform/db/sqlc/invoices.sql.go`
- `api/internal/platform/db/sqlc/payment_attempts.sql.go`
- `api/internal/platform/db/sqlc/parent_membership_children.sql.go` (new)
- (Deletion of the old `guardian_child_links.sql.go` and `guardians.sql.go` is part of the sqlc output when their `.sql` files are removed.)

## Verification

Run from the repository root:

```bash
# Database
make migrate-up
make migrate-down
make migrate-up
make sqlc-generate

# API
go build ./...
make test-api-repositories
go test ./...

# Web
cd web && npm install && npm run build && npm test
```

Manual validation scenarios:

1. Create a child with a `parent_carer` contact via the manager UI. Confirm `enrollmentComplete = true`. The child appears in the attendance list.
2. Create a child without a `parent_carer` contact. Confirm `enrollmentComplete = false`. Attempt a check-in — blocked. Attempt draft-invoice generation — the child appears as a preflight blocker with the reason `parent_carer_contact`.
3. As a manager, invite a parent. After acceptance, map the parent's membership to the child from scenario 1. As the parent, view invoices — the child appears. End the mapping with reason `access_revoked`. Refresh — the child disappears. Create a new mapping. Refresh — the child reappears.
4. Open the manager child detail page. Confirm the "Parent carers" section lists the contact. Confirm the "Linked guardians" section is gone.
5. Open `/staff/manager/guardians` — confirm the route is 404.
6. Run the seed command — confirm it succeeds and the seed child has a `parent_carer` contact and an active `parent_membership_children` mapping.

## Assumptions

- The pilot is not live, so dropping the three tables and their data is safe. (Confirmed by the user during the interview.)
- The `child_contacts` table already supports a `parent_carer` contact type with name, email, phone, and `has_parental_responsibility` fields. The plan does not add a new contact type or new columns.
- The `lifecycle_reason_code` enum already supports the reason codes used by the new table's check constraint (`duplicate_record`, `entered_in_error`, `left_nursery`, `safeguarding_direction`, `contact_update`, `access_revoked`, `other`). No new enum values are added.
- The audit writer's `WriteWithTx` API can write the same audit event shapes used by the dropped module. The plan does not change the audit writer.
- The `tenant.ActorFromGinContext` and `httpserver.MapDomainError` helpers stay the same; the new handler reuses them.
- The bootstrap test fixtures can be updated without changing the test framework or runner. The agent uses the same patterns already present in `people_routes_test.go`.
- The web `getChildContacts` endpoint already returns the contact data the new "Parent carers" section needs. The plan does not add a new endpoint.
- The manager child detail page already has the styling primitives (sections, lists, empty-state) that the new "Parent carers" section needs. The plan does not add new UI primitives.
- The `sqlc` tool is already wired via `go tool sqlc` in `go.mod`. `make sqlc-generate` runs cleanly.

## Risks and Fallbacks

- **Risk**: The dropped triggers on `parent_membership_guardians` may have been the only enforcement that a non-parent membership cannot have an active mapping. Without them, the new `parent_membership_children` table needs its own equivalent. **Fallback**: Task 1 adds `enforce_parent_membership_child_role` to enforce the same constraint on the new table. The agent verifies by attempting to insert a row for a `manager` membership — the trigger must reject it.
- **Risk**: A pre-existing reference in the codebase to `parentmappings` or `guardianlinks` is missed during the search. **Fallback**: `grep -ri "parentmappings\|guardianlinks\|guardian_child_links\|guardians\b" --include="*.go" --include="*.sql" --include="*.ts" --include="*.md"` is run as the final step of Task 11. Any remaining match is fixed in place.
- **Risk**: A bootstrap test inserts a `guardian_child_links` row in a test setup helper and the agent updates the test bodies but not the helper. **Fallback**: Task 7 explicitly lists `dbtest.go` (line 224 truncate list, line 400 fixture insert). The agent runs the full bootstrap test suite to catch any leftover fixture rows.
- **Risk**: The new `parent_membership_children` cascade trigger (`cascade_parent_membership_child_end`) does not fire because the existing `memberships` ended-cascade is done in application code (not a trigger). **Fallback**: Task 1 implements the cascade as a trigger and the agent verifies by manually ending a parent membership and observing the active child mappings are ended with `ended_reason_code = 'system_cascade'`. If the trigger does not fire, the implementing agent re-implements the cascade in the `end_membership` use case in the authentication / membership module (whichever owns membership lifecycle in the codebase) and the plan is updated to document that.
- **Risk**: The web "Parent carers" section needs contact data that the current `getChildContacts` endpoint does not return (e.g. `has_parental_responsibility`). **Fallback**: Task 9 verifies the endpoint's response shape against the contact domain type. If a field is missing, the agent adds it to the contact DTO and the corresponding SQL column is already present (`child_contacts.has_parental_responsibility`).
- **Risk**: The seed file currently has no `child_contacts` row for the seed child. Without a contact, the seed child is not enrollment-complete and downstream tests fail. **Fallback**: Task 8 explicitly adds a `parent_carer` contact to the seed child. The agent runs the full test suite to verify.
- **Risk**: `docs/API_REVIEW.md` recommends fixes that no longer apply (e.g. cascade-to-parent-membership on guardian deactivation). The plan instructs the agent to remove those sections, but a future reader may want to know why they were removed. **Fallback**: The agent adds a short note at the top of `docs/API_REVIEW.md` pointing to ADR-0008 and the new module, so historical recommendations can be traced.
- **Risk**: The audit event names change (`entity_type: parent_membership_guardian_mapping` → `parent_membership_child_mapping`). Any consumer of the audit log that hard-codes the old name breaks. **Fallback**: The plan notes that the new `entity_type` is `parent_membership_child_mapping`. If audit-log consumers exist in code, the agent updates them in Task 6 / 7. The agent grep's for the old `entity_type` string and updates all matches.
- **Risk**: A down migration that does not recreate the dropped tables surprises the operator. **Fallback**: The down migration raises a clear error explaining the irreversibility, and the migration file's header comment documents the choice. The plan also documents this in the "Migrations" section of `CONTEXT.md` (a small glossary entry: "Guardian Domain Removal Migration — irreversible; the dropped tables are not recreated on down.").
