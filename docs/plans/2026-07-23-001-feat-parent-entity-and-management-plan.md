---
title: Parent Entity and Management - Plan
type: feat
date: 2026-07-23
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-plan-bootstrap
execution: code
---

## Goal Capsule

Introduce a dedicated `Parent` entity to replace free-text parent/carer contacts in the children module. Parents become first-class entities linkable to multiple children, with optional portal access via user account. Staff can manage parents independently, select them during child registration, and link/unlink them from children bidirectionally.

**Authority hierarchy:** Product decisions from user interview (2026-07-23). Architecture follows existing Clean Architecture patterns in this codebase.

**Stop conditions:** All implementation units complete, `go build ./...` passes, `ng build` passes, `ng lint` passes, parent CRUD endpoints functional, child registration shows parent selector, billing reads from parents table.

**Execution profile:** Sequential units with dependency ordering. Each unit is one atomic commit.

**Tail ownership:** After ship, monitor that parent portal login still works and invoice generation succeeds with parent-linked children.

---

## Product Contract

### Summary

A new `parents` module provides a dedicated entity for parent/guardian data, replacing the free-text `parent_carer` contact type in `child_contacts`. Parents link to children via a `parent_children` table and optionally to user accounts for portal access. Staff manage parents through a standalone list page and from within child detail views. Child registration offers a parent selector with inline creation. Billing reads parent contact data from the parents table.

### Problem Frame

Currently, parent/guardian information is stored as free-text `child_contacts` rows with `contact_type = 'parent_carer'`. This data is disconnected from user accounts, cannot be reused across siblings, and provides no structured management surface. Staff type the same parent details repeatedly for siblings. Billing reads parent contact from these free-text records. There is no way to search for parents, view all children linked to a parent, or manage parent data independently.

### Requirements

**Parent Entity**

- R1. A `parents` table stores parent/guardian data: name, email, phone, address, relationship to child, parental responsibility flag, pickup authorization, emergency contact flag, notes, optional user_id link, active flag, and timestamps.
- R2. Parents are scoped by `tenant_id` and `branch_id` (multi-tenant isolation).
- R3. A `parent_children` table links parents to children with soft-delete lifecycle (`ended_at`, `ended_reason_code`, `ended_reason_note`).

**Parent CRUD**

- R4. Staff can create, view, edit, and soft-delete parents.
- R5. Staff can search parents by name, email, or phone. Filter by active/inactive status.
- R6. Parent detail page shows all fields and a list of linked children.
- R7. Staff can link a parent to a child from either the parent detail page or the child detail page.
- R8. Staff can unlink a parent from a child from either side.

**Child Registration**

- R9. Child registration's Contacts & Security step replaces the free-text parent/carer section with a parent selector dropdown plus an "Add new parent" inline form.
- R10. Emergency contacts and authorised collectors remain as free-text `child_contacts` entries.

**Billing Integration**

- R11. Invoices pull parent contact info (name, address, email, phone) from the linked parent record via the `parents` table.
- R12. If no parent is linked to a child, invoice generation returns a validation error.

**Portal Access**

- R13. Parent record has an optional `user_id` linking to a user account with `role=parent`.
- R14. Staff can invite a parent to portal access from the parent detail page, which creates a user account and sends an invite email.
- R15. Staff can revoke portal access by deactivating the user account (keeps `user_id` linked for re-activation).
- R16. Parent portal reads children from `parent_children` via the parent's `user_id`.

**Data Cleanup**

- R17. Remove `parent_carer` from `child_contacts` contact type. Emergency contacts and authorised collectors remain.

### Scope Boundaries

**Deferred to Follow-Up Work**

- Domain events for parent lifecycle (e.g., `ParentDeactivated`). Not needed for initial scope; can be added later if audit trail or event-driven behavior is required.
- Parent portal "My Children" dashboard view. Portal stays invoice-only for now.
- Room-based parent filtering (filter parents by which room their children attend).
- Bulk parent import/export.
- Parent merge (duplicate detection and merge workflow).

**Outside this product's identity**

- Parent self-registration (parents are always created by staff, not by parents themselves).

---

## Planning Contract

### Key Technical Decisions

**KTD1. No domain events for parents (initially).**
The children module emits `ChildDeactivated` events. The parents module will not emit events in the initial implementation. Parent lifecycle is simple CRUD; events add scope without immediate value. Rationale: keeps scope small; the event system is extensible if needed later.

**KTD2. Soft-delete on parent_children links.**
Following the `parent_membership_children` pattern, `parent_children` uses `ended_at` + `ended_reason_code` + `ended_reason_note` for lifecycle management. This preserves link history and allows re-activation. Rationale: consistent with existing patterns; supports audit requirements.

**KTD3. Remove parent_carer from child_contacts entirely.**
Clean cut. The `parents` table is the sole source of truth for parent contact data. No backward compatibility layer. Rationale: no live data to migrate; database reset makes this safe.

**KTD4. Billing adapter reads from parents table.**
Replace the existing `parentContactLookupAdapter` (which queries `child_contacts` for `parent_carer`) with a new adapter that queries `parents` via `parent_children`. Rationale: single source of truth for parent contact data.

**KTD5. Portal access via parent_children, not parent_membership_children.**
Parent portal reads children by finding the parent record via `user_id`, then listing children from `parent_children`. The existing `parent_membership_children` table remains for backward compatibility but is not the primary portal access path. Rationale: parent_children is the canonical parent-child relationship.

### Assumptions

- Database will be reset (no live data migration needed).
- Existing `parent_membership_children` table and module remain functional during the transition but are not the primary parent-child relationship going forward.
- The `parentchildmappings` module continues to exist for membership-based portal access but is supplemented by the new parents module for data relationships.
- Email sending for portal invites uses the existing `platform/email` package.

### Sequencing

Units are ordered by dependency: schema first, then backend module, then integrations, then frontend. Each unit is independently committable.

---

## Implementation Units

### U1. Database Schema: parents and parent_children tables

**Goal:** Create the database tables and sqlc queries for the parents module.

**Requirements:** R1, R2, R3

**Dependencies:** None (first unit)

**Files:**
- `api/db/migrations/000NNN_add_parents_tables.up.sql` (create)
- `api/db/migrations/000NNN_add_parents_tables.down.sql` (create)
- `api/db/query/parents.sql` (create)
- `api/db/query/parent_children.sql` (create)
- `api/internal/platform/db/sqlc/` (regenerated)

**Approach:**
Create migration with two tables:

`parents` table: `id` (uuid PK), `tenant_id` (uuid FK), `branch_id` (uuid FK), `first_name` (text NOT NULL), `last_name` (text), `email` (text), `phone` (text), `address_line1` (text), `address_line2` (text), `address_city` (text), `address_postcode` (text), `relationship_to_child` (text), `has_parental_responsibility` (bool DEFAULT false), `can_pick_up` (bool DEFAULT false), `is_emergency_contact` (bool DEFAULT false), `notes` (text), `user_id` (uuid FK to users, nullable), `is_active` (bool DEFAULT true), `created_at` (timestamptz DEFAULT now()), `updated_at` (timestamptz DEFAULT now()).

`parent_children` table: `id` (uuid PK), `tenant_id` (uuid FK), `branch_id` (uuid FK), `parent_id` (uuid FK to parents), `child_id` (uuid FK to children), `ended_at` (timestamptz), `ended_reason_code` (lifecycle_reason_code), `ended_reason_note` (text), `created_at` (timestamptz DEFAULT now()), `updated_at` (timestamptz DEFAULT now()). Unique constraint on `(tenant_id, branch_id, parent_id, child_id) WHERE ended_at IS NULL`.

Add DB triggers: enforce parent scope (parent and child in same tenant+branch), enforce user scope (user_id if set belongs to same tenant).

Write sqlc queries following the `parent_membership_children.sql` pattern: `ParentsList`, `ParentsListFiltered`, `ParentsCount`, `ParentsGetByID`, `ParentsCreate`, `ParentsUpdate`, `ParentsSoftDelete`, `ParentChildrenListByParent`, `ParentChildrenListByChild`, `ParentChildrenFindActiveByPair`, `ParentChildrenCreate`, `ParentChildrenEnd`.

Run `make sqlc-generate` to produce generated code.

**Patterns to follow:**
- `api/db/migrations/000001_baseline.up.sql` (table conventions, trigger patterns)
- `api/db/query/parent_membership_children.sql` (query naming, scoping, COALESCE patterns)

**Test scenarios:**
- Migration up/down applies cleanly
- sqlc generates without errors
- Queries compile against the schema

**Verification:** `make migrate-up` succeeds, `make sqlc-generate` succeeds, `go build ./...` in `api/` passes.

---

### U2. Backend Domain and Application Layer

**Goal:** Create the parents module domain entities, repository interface, and use cases.

**Requirements:** R1, R4, R5, R6

**Dependencies:** U1

**Files:**
- `api/internal/modules/parents/domain/parent.go` (create)
- `api/internal/modules/parents/domain/parent_child.go` (create)
- `api/internal/modules/parents/domain/repository.go` (create)
- `api/internal/modules/parents/domain/errors.go` (create)
- `api/internal/modules/parents/application/create_parent.go` (create)
- `api/internal/modules/parents/application/update_parent.go` (create)
- `api/internal/modules/parents/application/get_parent.go` (create)
- `api/internal/modules/parents/application/list_parents.go` (create)
- `api/internal/modules/parents/application/soft_delete_parent.go` (create)
- `api/internal/modules/parents/application/link_child.go` (create)
- `api/internal/modules/parents/application/unlink_child.go` (create)
- `api/internal/modules/parents/application/invite_to_portal.go` (create)
- `api/internal/modules/parents/application/revoke_portal_access.go` (create)
- `api/internal/modules/parents/application/interfaces.go` (create)

**Approach:**
Domain layer: `Parent` struct with all fields from the schema. `ParentChild` struct for the linking entity. `Repository` interface with `Tx = any` following the `parentchildmappings` pattern. Domain errors for not-found, already-exists, validation.

Application layer: One use case per file. Each use case wraps operations in `txMgr.ExecTx` for mutations. Use cases: `CreateParent` (validate + insert), `UpdateParent` (partial update via fields map), `GetParent` (single lookup with children), `ListParents` (search + filter + pagination), `SoftDeleteParent` (set is_active=false), `LinkChild` (idempotent create), `UnlinkChild` (end with reason), `InviteToPortal` (create user account + send email), `RevokePortalAccess` (deactivate user).

Cross-module interfaces in `interfaces.go`: `UserCreator` (for portal invite), `EmailSender` (for invite email), `ChildExistenceChecker` (for link validation).

**Patterns to follow:**
- `api/internal/modules/parentchildmappings/application/create_mapping.go` (transaction pattern, idempotent create)
- `api/internal/modules/children/application/list_children.go` (use case structure)
- `api/internal/modules/children/domain/child.go` (entity struct, domain methods)

**Test scenarios:**
- Happy path: create parent, get parent, list parents, update parent, soft delete
- Link child: create link, idempotent re-link returns existing, unlink with reason
- Edge cases: create with minimal fields (only first_name), search with no results, link to non-existent child returns validation error
- Error paths: get non-existent parent returns not-found, update inactive parent returns error

**Verification:** `go build ./...` passes, unit tests pass.

---

### U3. Backend Infrastructure and HTTP Layer

**Goal:** Implement the Postgres repository and HTTP handler for the parents module.

**Requirements:** R1, R2, R4, R5, R6, R7, R8

**Dependencies:** U1, U2

**Files:**
- `api/internal/modules/parents/infrastructure/postgres/repository.go` (create)
- `api/internal/modules/parents/infrastructure/postgres/repository_test.go` (create)
- `api/internal/modules/parents/interfaces/http/handler.go` (create)
- `api/internal/modules/parents/interfaces/http/dto.go` (create)

**Approach:**
Repository: Implements `domain.Repository` using sqlc-generated queries. Constructor takes `*pgxpool.Pool`. Methods use `sqlc.New(tx.(pgx.Tx))` for transactional operations and `sqlc.New(r.pool)` for read-only queries. Includes pgtype conversion helpers (`uuidToPgtype`, `pgtypeUUIDToUUID`, etc.).

Handler: Groups use cases, registers routes under `/parents` with manager role guard. Endpoints: `GET /parents` (list), `POST /parents` (create), `GET /parents/:parent_id` (get), `PUT /parents/:parent_id` (update), `DELETE /parents/:parent_id` (soft delete), `POST /parents/:parent_id/link-child` (link), `DELETE /parents/:parent_id/link-child/:child_id` (unlink), `POST /parents/:parent_id/invite` (portal invite), `POST /parents/:parent_id/revoke-access` (revoke).

DTOs: Request/response structs with JSON tags. Mapping functions between domain entities and DTOs.

**Patterns to follow:**
- `api/internal/modules/parentchildmappings/infrastructure/postgres/repository.go` (sqlc integration, pgtype helpers)
- `api/internal/modules/children/interfaces/http/handler.go` (route registration, error handling)
- `api/internal/modules/parentchildmappings/interfaces/http/handler.go` (simpler handler pattern)

**Test scenarios:**
- Repository integration tests with `dbtest.RequirePostgres(t)` + `dbtest.Reset(t, pool)`
- Create parent and verify retrieval
- Link parent to child and verify listing
- Search parents by name
- Soft delete and verify inactive status

**Verification:** `go build ./...` passes, `go vet ./...` passes, integration tests pass.

---

### U4. Bootstrap Wiring and Cross-Module Adapters

**Goal:** Wire the parents module into the application and create adapters for cross-module integration.

**Requirements:** R11, R16

**Dependencies:** U2, U3

**Files:**
- `api/internal/app/bootstrap/adapters.go` (modify)
- `api/internal/app/bootstrap/wire.go` (modify)
- `api/internal/app/bootstrap/wire_gen.go` (regenerated)
- `api/internal/app/bootstrap/providers.go` (modify)

**Approach:**
Add `parentsSet` in `wire.go` with repository, use cases, handler, and bindings. Add `ParentsHandler` to `appComponents` struct in `providers.go`. Register routes in `buildGinEngine()`.

Create adapters in `adapters.go`:
- `parentContactLookupAdapter` — implements billing's `ParentContactLookup` interface. Queries `parents` via `parent_children` to find the primary parent for a child. Replaces the existing `parentContactLookupAdapter` that queries `child_contacts`.
- `parentChildLookupForPortalAdapter` — implements a `ParentChildLookup` interface for the parent portal. Finds children by parent's `user_id`.

Run `make wire-generate` to regenerate Wire code.

**Patterns to follow:**
- `api/internal/app/bootstrap/adapters.go` lines 165-215 (existing parentContactLookupAdapter)
- `api/internal/app/bootstrap/wire.go` (wire.NewSet patterns)

**Test scenarios:**
- Billing can look up parent contact via the new adapter
- Portal can list children via parent's user_id
- Wire generates without errors

**Verification:** `make wire-generate` succeeds, `go build ./...` passes.

---

### U5. Update Children Module: Remove parent_carer Contacts

**Goal:** Remove `parent_carer` from child contacts handling. Emergency contacts and authorised collectors remain.

**Requirements:** R17

**Dependencies:** U2, U3, U4

**Files:**
- `api/internal/modules/children/domain/child_contact.go` (modify)
- `api/internal/modules/children/application/helpers.go` (modify)
- `api/internal/modules/children/interfaces/http/create_child_dto.go` (modify)
- `api/internal/modules/children/interfaces/http/dto_helpers.go` (modify)

**Approach:**
Remove `ContactTypeParentCarer` from the `ContactType` constants. Update `buildChildContactEntries` to no longer accept parent_carer contacts. Update DTO mapping to remove parent_carer from the contacts payload. Keep `ContactTypeEmergencyContact` and `ContactTypeAuthorisedCollector` unchanged.

The `HasParentCarerContact` field on `Child` struct will be updated to check for a linked parent in the `parent_children` table instead of checking `child_contacts`. This requires a new cross-module interface in the children application layer.

**Patterns to follow:**
- `api/internal/modules/children/domain/child_contact.go` (contact type constants)
- `api/internal/modules/children/application/helpers.go` (buildChildContactEntries)

**Test scenarios:**
- Child registration no longer accepts parent_carer contacts
- Emergency contacts and authorised collectors still work
- `HasParentCarerContact` reflects whether a parent is linked via parent_children

**Verification:** `go build ./...` passes, existing children tests pass.

---

### U6. Frontend: Parent List Page

**Goal:** Create the standalone parent list page in the staff portal.

**Requirements:** R4, R5

**Dependencies:** U3

**Files:**
- `web/src/app/features/staff/data/parents-api.service.ts` (create)
- `web/src/app/features/staff/models/parents.models.ts` (create)
- `web/src/app/features/staff/pages/manager-parents/manager-parents.component.ts` (create)
- `web/src/app/features/staff/pages/manager-parents/manager-parents.component.html` (create)
- `web/src/app/features/staff/pages/manager-parents/manager-parents.component.spec.ts` (create)
- `web/src/app/app.routes.ts` (modify)

**Approach:**
API service: `ParentsApiService` with methods for list, get, create, update, delete, link, unlink, invite, revoke. Follows the `StaffApiService` pattern.

Models: `ParentRecord` interface with all parent fields. `ParentListResponse` for paginated results.

Page component: Table with search input, status filter (active/inactive/all), sortable columns (name, email, phone, created_at), pagination. Responsive card layout on mobile. Empty state when no parents. Each row links to parent detail page.

Routes: Add `manager/parents` (list), `manager/parents/new` (create), `manager/parents/:parentId` (detail) under the manager role guard.

Sidebar: Add "Parents" item under the "People" group in the sidebar navigation.

**Patterns to follow:**
- `web/src/app/features/staff/pages/manager-children/` (list page pattern)
- `web/src/app/features/staff/data/staff-api.service.ts` (API service pattern)

**Test scenarios:**
- Component renders with empty state
- Search filters results by name
- Status filter toggles between active/inactive/all
- Pagination navigates through results
- Click on row navigates to parent detail

**Verification:** `npx ng lint` passes, `ng build` passes, component renders in browser.

---

### U7. Frontend: Parent Detail Page

**Goal:** Create the parent detail page with view/edit capabilities and linked children list.

**Requirements:** R6, R7, R8, R14, R15

**Dependencies:** U6

**Files:**
- `web/src/app/features/staff/pages/manager-parent-detail/manager-parent-detail.component.ts` (create)
- `web/src/app/features/staff/pages/manager-parent-detail/manager-parent-detail.component.html` (create)
- `web/src/app/features/staff/pages/manager-parent-detail/manager-parent-detail.component.spec.ts` (create)
- `web/src/app/app.routes.ts` (modify - add detail route)

**Approach:**
Page shows parent fields in a card layout. Edit button opens inline edit form (or navigates to edit route). "Linked Children" section shows a table of linked children with unlink action. "Link Child" button opens a child selector modal. "Invite to Portal" button (visible when user_id is null) triggers portal invite flow. "Revoke Access" button (visible when user_id is set) deactivates the user account.

Parent form component (shared): Reusable form with fields for first_name, last_name, email, phone, address fields, relationship_to_child, checkboxes for has_parental_responsibility, can_pick_up, is_emergency_contact, and notes textarea.

**Patterns to follow:**
- `web/src/app/features/staff/pages/manager-child-detail/` (detail page pattern)
- `web/src/app/features/staff/pages/manager-children/manager-children.component.ts` (table with actions)

**Test scenarios:**
- Component renders parent data
- Edit form saves changes
- Link child modal shows available children
- Unlink child removes the link
- Invite to portal sends request and shows success
- Revoke access deactivates user

**Verification:** `npx ng lint` passes, `ng build` passes, component renders in browser.

---

### U8. Frontend: Update Child Registration with Parent Selector

**Goal:** Update the child registration stepper's Contacts & Security step to use a parent selector.

**Requirements:** R9, R10

**Dependencies:** U6

**Files:**
- (Modify existing registration stepper contacts step component)
- (Modify existing registration draft storage if needed)

**Approach:**
Replace the free-text parent/carer contact section with a parent dropdown selector. The dropdown lists active parents with name and email. "Add new parent" button opens an inline form that creates a parent record and selects it. Emergency contacts and authorised collectors sections remain unchanged.

The selected parent_id is stored in the registration draft and used to create the parent_children link when the child is created.

**Patterns to follow:**
- Existing stepper wizard pattern in child registration
- `web/src/app/features/staff/pages/manager-child-edit/` (edit form pattern)

**Test scenarios:**
- Dropdown shows existing parents
- "Add new parent" creates record and selects it
- Selected parent is linked to child on registration submit
- Emergency contacts still work as free-text

**Verification:** `npx ng lint` passes, `ng build` passes, registration flow works end-to-end.

---

### U9. Frontend: Update Child Detail with Parents Section

**Goal:** Add a "Parents" section to the child detail page showing linked parents.

**Requirements:** R7, R8

**Dependencies:** U7

**Files:**
- (Modify existing child detail page component)

**Approach:**
Add a "Parents" tab or section to the child detail page. Shows linked parent cards with name, relationship, email, phone. Each card has an "Unlink" action. "Link Parent" button opens a parent selector modal. This mirrors the parent detail page's "Linked Children" section but from the child's perspective.

**Patterns to follow:**
- Existing tab navigation in child detail (overview, attendance, funding, health, contacts)

**Test scenarios:**
- Parents section shows linked parents
- Link parent adds a new link
- Unlink parent removes the link
- Changes reflect immediately without page refresh

**Verification:** `npx ng lint` passes, `ng build` passes.

---

## Verification Contract

**Backend:**
- `go fmt ./...` — no formatting issues
- `go vet ./...` — no vet warnings
- `go build ./...` — compiles cleanly
- `make sqlc-generate` — generates without errors
- `make wire-generate` — generates without errors
- `make migrate-up` — migration applies cleanly
- Integration tests pass with `TEST_DATABASE_URL`

**Frontend:**
- `npx ng lint` — no lint errors
- `ng build` — production build with zero errors and warnings
- `npm test` — unit tests pass

**Functional:**
- Parent CRUD endpoints return correct responses
- Parent list page loads and displays parents
- Parent detail page shows linked children
- Child registration shows parent selector
- Billing reads parent contact from parents table

---

## Definition of Done

- All 9 implementation units complete
- `go build ./...` passes in `api/`
- `ng build` passes in `web/` with zero warnings
- `npx ng lint` passes in `web/`
- Parent CRUD API endpoints functional
- Parent list page renders in browser
- Parent detail page renders with linked children
- Child registration parent selector works
- Billing invoice generation uses parent contact from parents table
- No `parent_carer` references remain in child_contacts handling
- All new files follow existing naming conventions
