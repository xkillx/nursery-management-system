# Manager Rooms — Real Occupancy Implementation Plan

## Goal

The manager Rooms page (`/staff/manager/rooms`) becomes a reliable "today's capacity vs demand" view for the manager's own branch. Occupancy is no longer demo data: it is the count of children whose `primary_room_id` points at the room, surfaced alongside capacity, sorted highest-first, with an unmissable over-capacity banner. Children become linkable to a primary room through the existing Children area; archiving a room with children assigned is blocked until the manager reassigns them.

## Non-Goals (explicit)

- Staffing-ratio calculations and statutory adult:child ratio display (per age group). Defer.
- Check-in-now occupancy (live count of currently checked-in children per room). Defer.
- Scheduled / expected occupancy (children booked for a future session). Defer.
- Historical utilisation, trends, or local-authority reporting. Defer.
- Move-child-between-rooms UI on the Rooms page. Rebalancing happens in the Children area.
- Per-child drill-down from the Rooms page. Defer.
- Auto-refresh / polling. Refresh on page load and after create/archive only.
- Owner cross-site dashboard updates. Owner view gets the same backend change but no UI rework.
- Backfill of historical room assignments. Existing children start as `primary_room_id = NULL` (unassigned). Manager assigns them from the Children form.
- Changing the existing CRUD on rooms (create / update / archive / reactivate behaviour remains as in `docs/plans/room-management-plan.md`).

## Context Alignment

Two terms to add to `CONTEXT.md` (after this plan is written, or inline during execution):

- **Primary Room** — the single room a child is operationally attached to for day-to-day care. One child, one primary room at any time. Replaces prior language of "assigned to a room".
- **Unassigned Child** — a child with `primary_room_id IS NULL`. Unassigned children do not contribute to any room's occupancy and do not block archiving.

Existing glossary terms honoured: Site (API) = Branch (DB) per `CONTEXT.md` "Branch Scope" entry. Archive means `is_active = false`; children in an archived room are still counted as assigned (the link persists).

## User-Facing Behaviour (acceptance criteria)

1. On `/staff/manager/rooms`, the page calls the backend and renders the manager's branch rooms with real occupancy. No demo hash.
2. Each room row shows `assigned / capacity` (e.g. `8 / 12`). The assigned count is not shown as a separate number; it's only visible as part of the ratio, per the "Hide assigned count" decision.
3. Rooms are sorted by occupancy % descending; ties broken by name. Active rooms only by default; an existing status filter exposes archived rooms.
4. A red banner at the top of the page lists every over-capacity room by name (e.g. "Sunshine Room is over capacity (14/12)"). Banner is visible from any scroll position and uses the same over-capacity logic (> 100%). Banner is hidden when no room is over capacity.
5. When the manager tries to archive a room that still has children assigned (`primary_room_id = <room>` and `is_active = true`), the API returns a clear error and the UI shows the count of children that need reassignment. Archive succeeds only when zero children are assigned.
6. The Children form gains a `Primary room` field. New children require it; existing children show it as optional (NULL = unassigned) until edited.
7. Manager sees only their own branch's rooms. The existing site-selector on the owner view stays; the manager view does not get a site selector.

## Decisions

### Product / domain (from interview)

1. **Occupancy is "children currently assigned to the room"** — not checked-in, not scheduled, not historical. Count of children where `primary_room_id = room.id` AND child is active.
2. **Staffing ratio is out of scope** for this iteration; deferred to a follow-up.
3. **Sort order is highest-occupancy-first**, ties broken by name. Active rooms only by default.
4. **Over-capacity banner is page-level** and lists each over-capacity room by name with a link to its row. The bar colour is the same red used in row-level over-capacity treatment.
5. **Refresh on page load and after create/archive/edit actions only.** No polling, no live updates.
6. **Empty rooms (`assigned = 0`) show `0%` and remain archivable.** Archive is blocked only when the count is greater than zero.
7. **Rebalancing happens in the Children area**, not on the Rooms page. Rooms page surfaces that rebalance is needed; it does not perform it.
8. **`primary_room_id` is a single current value.** No effective-dated history. Editing a child re-points them. History is not kept.
9. **Archive-with-children is blocked** with a clear error. Manager must move the children out first from the Children area.
10. **Manager authorisation stays branch-scoped.** No cross-branch selector for managers. Owners retain cross-site behaviour unchanged.

### Implementation (agent ownership)

1. **Schema change**: Add `primary_room_id UUID NULL REFERENCES rooms(id) ON DELETE SET NULL` to the `children` table in a new migration `000030_add_child_primary_room.up.sql` / `.down.sql`. Index on `(tenant_id, branch_id, primary_room_id)` for the occupancy query.
2. **Children module update**: Add `PrimaryRoomID *uuid.UUID` to the `Child` domain entity and persist it through the existing create / update use cases. Add a thin use case for setting the primary room only (used by the Children form), or fold it into `UpdateChild` — implementer's call.
3. **Rooms module update**: Extend the existing `ListRooms` use case + handler to accept `?include=occupancy` and return `assigned_count` and `is_over_capacity` per room. Default behaviour (no `include` flag) returns the room list as today. Occupancy is computed in a single SQL query: `SELECT room_id, COUNT(*) FROM children WHERE tenant_id = $1 AND branch_id = $2 AND primary_room_id IS NOT NULL AND is_active = true GROUP BY room_id`.
4. **API contract**: Extend the room list response shape (additive) with `assigned_count: integer` and `is_over_capacity: boolean` when `?include=occupancy=true`. The owner-view existing call continues to work unchanged (no flag = no extra fields). The manager view sends `?include=occupancy=true`.
5. **Archive semantics**: The existing `ArchiveRoom` use case (per `docs/plans/room-management-plan.md`) already returns `room_has_children` when active children exist. The error payload gains a `assigned_count` integer so the UI can show "N children still assigned — reassign them first." Same error path serves both owner and manager views.
6. **Frontend service**: Add `listRooms(siteId, { includeOccupancy: true })` to a new `web/src/app/features/staff/data/staff-rooms-api.service.ts` (per AGENTS.md: staff feature stays self-contained; don't borrow from owner). Existing `OwnerApiService.listRooms` gets an optional second arg for occupancy too, defaulting to `false` so the owner view is unchanged.
7. **Frontend page**: Either reuse `OwnerRoomsComponent` (already branches on `isOwner` and handles the manager-no-site-selector path) and add occupancy rendering, or introduce a thin `ManagerRoomsComponent` that delegates to the same template. The route at `/staff/manager/rooms` must render the manager view with real occupancy on load. Verify the route is wired to a real Angular component and not the TailAdmin placeholder.
8. **Over-capacity banner**: A page-level component (e.g. `OverCapacityBannerComponent` in `web/src/app/shared/components/`) shown above the rooms table when any room has `is_over_capacity === true`. Lists each over-capacity room by name with anchor links to its row. Hidden when empty.
9. **Children form**: Add a `Primary room` `<app-select>` to the existing `ChildFormComponent`, populated from the rooms API (active rooms in the branch). Required for new children. Optional on edit (NULL = unassigned). Wire to the updated create / update use cases.
10. **No auto-refresh**: Component fetches occupancy on `ngOnInit` and after every successful create / archive / reactivate / edit. No `setInterval`, no WebSocket.
11. **Tests**:
    - Domain / application: unit tests for the occupancy query and the updated archive error payload using mock repos.
    - Integration: a `repository_test.go` for the rooms repo's new occupancy SQL against the disposable test DB; a children-repo test confirming `primary_room_id` round-trips and the `ON DELETE SET NULL` behaviour.
    - Handler: a `handler_test.go` for the `?include=occupancy` flag (both with and without) and the archive error code path.
    - Frontend: extend `manager-rooms.component.spec.ts` (or create it if the page is moved into the staff feature) to cover occupancy rendering, sort order, banner visibility, and the empty-state path. Add coverage to `child-form.component.spec.ts` for the new `Primary room` field.
12. **Documentation**:
    - Add `Primary Room` and `Unassigned Child` to `CONTEXT.md` glossary.
    - Note the new query param in `docs/API-CONTRACT.openapi.yaml` (additive change to `GET /sites/{site_id}/rooms`).
    - Note the new column in `docs/API-SCHEMA-STATE.md`.

## Out-of-Scope ADR Candidate

The "occupancy is current primary-room assignment" decision could become an ADR: it is the kind of choice a future reader will wonder about (why not check-ins? why not scheduled?), it is hard to reverse once consumers depend on it, and there were real alternatives (check-in now, scheduled, historical). File as `docs/adr/0006-rooms-occupancy-definition.md` if you confirm the decision; otherwise record it as a glossary entry only.

## Current State (relevant subset)

- `api/internal/modules/rooms/` is implemented (domain, application, infrastructure/postgres, interfaces/http) per `docs/plans/room-management-plan.md`. Endpoints: list / get / create / update / archive / reactivate.
- `api/internal/modules/children/` has no `primary_room_id` and no concept of a child's primary room.
- `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.ts` renders the manager view via `isOwner` branch fallback but currently uses `demoOccupancy()` (line 343) to fake the number.
- Route at `staff/manager/rooms` exists in `web/src/app/app.routes.ts:155-185` and points at `OwnerRoomsComponent` / `OwnerRoomFormComponent` — but the live dev server serves the TailAdmin template, so a route/auth-guard fix is part of "wire the page" below.
- `staff/data/staff-api.service.ts` has no room methods. Manager view currently borrows `OwnerApiService` (cross-feature coupling that AGENTS.md disallows).

## Execution Order

1. **Schema**: write `000030_add_child_primary_room.up.sql` + `.down.sql`; run `make migrate-up` and `make migrate-verify` against disposable DBs.
2. **Children module**: add `PrimaryRoomID *uuid.UUID` to the `Child` entity, plumb through create / update use cases, add the column to the sqlc queries (`api/db/query/children.sql`), regenerate with `make sqlc-generate`, update the repository, add a repo test for round-trip and `ON DELETE SET NULL`.
3. **Rooms module — occupancy query**: add a new sqlc query in `api/db/query/rooms.sql` for `assigned_count per room`; regenerate; add a use-case method `ListRoomsWithOccupancy` (or extend `ListRooms` with a flag — implementer's call); expose `?include=occupancy` on `GET /sites/:site_id/rooms`; add unit + integration tests; update `API-CONTRACT.openapi.yaml` and `API-SCHEMA-STATE.md`.
4. **Rooms module — archive error shape**: extend the existing `room_has_children` error response with `assigned_count: integer`; add a test for the new payload.
5. **Children form (web)**: add `Primary room` to `ChildFormComponent` model, populated from the rooms API (active rooms in branch); required on create, optional on edit; update spec.
6. **Staff rooms API service (web)**: create `web/src/app/features/staff/data/staff-rooms-api.service.ts` with `listRooms` / `archiveRoom` / `reactivateRoom` (latter two already exist on owner service — copy to staff service per AGENTS.md cross-module rule).
7. **Manager rooms page (web)**: ensure `/staff/manager/rooms` renders a real Angular component (not the TailAdmin template). Either reuse `OwnerRoomsComponent` with `?include=occupancy=true` from the manager branch path, or move the page to `web/src/app/features/staff/pages/manager-rooms/` and have the route point at it. Replace `demoOccupancy()` with real `assigned_count`. Add sort-by-occupancy-desc to the rendered list.
8. **Over-capacity banner**: add `web/src/app/shared/components/over-capacity-banner/` component; mount it above the rooms table when any `is_over_capacity === true`; hide otherwise. Unit test.
9. **Manager rooms spec**: write / extend `manager-rooms.component.spec.ts` for occupancy rendering, sort order, banner visibility, empty-state, archive error mapping.
10. **CONTEXT.md**: add `Primary Room` and `Unassigned Child` glossary entries.
11. **Verify end-to-end**: `make run-api`, seed a branch with rooms + children assigned, log in as a manager, confirm banner appears on over-capacity rooms, confirm archive is blocked, confirm sort order, confirm page refresh after archive shows updated list.
12. **Optional ADR**: if you confirm the occupancy-definition decision, write `docs/adr/0006-rooms-occupancy-definition.md`.

## Verification Checklist (executable by the implementing agent)

- [ ] `make migrate-up` and `make migrate-verify` both pass; children table has `primary_room_id UUID NULL` referencing `rooms(id)`; index exists.
- [ ] `make sqlc-generate` runs without error; new occupancy query appears in generated code.
- [ ] `make test-api-repositories` passes; new tests for occupancy and archive error are included.
- [ ] `cd api && go test ./...` passes (no regression in children or rooms modules).
- [ ] `cd web && npm test` passes; new specs for `manager-rooms` and `child-form` pass.
- [ ] `make run-api` + `cd web && npm start`; log in as manager, navigate to `/staff/manager/rooms`, see real occupancy, sort order, banner when applicable.
- [ ] Archive-with-children returns the new error payload with `assigned_count`; UI shows it.
- [ ] Creating / editing a child with a primary room persists; the room's `assigned_count` reflects the change after page refresh.
- [ ] `CONTEXT.md` has the two new glossary entries; `API-CONTRACT.openapi.yaml` documents the new query param; `API-SCHEMA-STATE.md` documents the new column.

## Risk / Open Questions for the Implementer

- **Sqlc ergonomics**: the occupancy query is a `GROUP BY` over a foreign key column. If sqlc's generated code is awkward, an alternative is to keep the count in a hand-written repository method that does not go through sqlc. Decide locally; do not raise to the user.
- **Owner view**: the existing `OwnerRoomsComponent` already renders the manager view via the `isOwner` fallback. Decide whether to (a) extend the shared component with real occupancy and let the owner view pick it up too (additive), or (b) keep owner demo data and only switch the manager view to real data. The first is consistent and small; the second is conservative. Default to (a) unless the owner view's design assumes demo data.
- **Children form validation rule**: spec says "Required for new children". Confirm during implementation that existing children are editable without setting a room, and that bulk-import paths (if any) keep `primary_room_id = NULL` for unassigned.
