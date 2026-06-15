# Room Management Implementation Plan

## Goal

Enable nursery operators to define and manage rooms within each nursery site. Owners manage rooms across all sites within their tenant; managers manage rooms within their assigned site. Staff can view rooms read-only; parents have no room access. Rooms are site-scoped operational resources that support future child assignment, attendance tracking, ratio safety, and reporting.

## Non-Goals

- Child-to-room assignment (deferred until room/session planning)
- Room-move tracking in attendance (deferred per "Attendance Capture Scope")
- Ratio safety calculations (deferred per Post-MVP sequence)
- Room-specific billing or funding rules
- Bulk room import/export
- Room usage reporting or analytics
- Parent visibility of rooms

## Context Alignment

Five new terms added to `CONTEXT.md`: Nursery Room, Room Age Group, Room Capacity, Room Archive, Room Reactivation.

Reconciled terms:
- **Site vs Branch**: API uses `site_id` (product-facing language). Database uses `branch_id` (existing engineering convention — see `branches` table in `000001_init_core_tables.up.sql`). These are the same location boundary per the glossary's "Branch Scope" definition.
- **Archive vs Deactivate**: Room uses "archive" for setting `is_active = false` to distinguish from user/guardian deactivation. Reactivation is the inverse action. This follows the "No Hard-Delete Core Records" policy.
- **Age group values**: baby, toddler, preschool, mixed — constrained enum, not free text.

## Current State

### Database
- `branches` table exists with `id`, `tenant_id`, `name`, `is_active`, `core_hourly_rate_minor` (migrations `000001` and `000022`)
- `branches` has `UNIQUE (tenant_id, name)` constraint
- Last migration: `000028_split_child_names`
- No `rooms` table exists yet

### Backend module pattern (14 existing modules, e.g. `children`)
```
api/internal/modules/rooms/
  domain/          → entities.go, repository.go (zero framework imports)
  application/     → one use case per file
  infrastructure/postgres/ → repository.go, repository_test.go
  interfaces/http/ → handler.go, dto.go
```

### Key contracts to follow
- `tenant.ActorFromGinContext(c)` → `ActorContext` with `UserID`, `MembershipID`, `TenantID`, `BranchID`
- `tenant.OwnerActorFromGinContext(c)` → `OwnerActorContext` (no `BranchID` — owner is tenant-wide)
- `transaction.Manager.ExecTx(ctx, fn)` for transactional writes
- `audit.Writer` for audit persistence
- Role middleware: `requireRoles("manager", "owner")` for shared access
- `pgtype.UUID` for sqlc ↔ domain conversion (see `uuidToPgtype`/`pgtypeUUIDToUUID` in owner repository)
- sqlc queries in `api/db/query/rooms.sql`, generated to `api/internal/platform/db/sqlc/rooms.sql.go`
- `pgx.ErrNoRows` for not-found checks (`isNoRows` helper)

### Frontend pattern (owner feature)
```
web/src/app/features/owner/
  models/owner.models.ts       → TypeScript interfaces
  data/owner-api.service.ts    → HTTP service, snake_case API mapping
  pages/owner-overview/        → standalone component + template + spec
  pages/owner-manager-access/  → standalone component + template + spec
  utils/owner-formatters.ts    → formatting helpers + specs
```

### Routing
- Frontend routes in `web/src/app/app.routes.ts` with `roleGuard` and `data: { roles: [...] }`
- Navigation in `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts`
- Route constants in `web/src/app/core/constants/roles.ts`
- All routes under `AppLayoutComponent` with `authGuard`

## Decisions

### Product/domain decisions (from interview)
1. Archive is blocked if active children are assigned to the room — error: "Room has X active children assigned — reassign them before archiving."
2. Archived rooms can be reactivated (restore `is_active = true`).
3. "Mixed" age group means a room intentionally serves multiple age bands — it's a valid operational choice, not a catch-all.
4. Age group is a constrained enum (baby, toddler, preschool, mixed) with no free-text escape hatch.

### Implementation decisions (agent ownership)
1. **Module placement**: New `api/internal/modules/rooms/` following the 4-layer pattern (domain/application/infrastructure/interfaces).
2. **API routing**: Register on a shared route group with `requireRoles("manager", "owner")`. Route prefix: `POST/GET /sites/:site_id/rooms` etc. Each handler checks the actor — owner validates site belongs to tenant; manager validates site_id matches session `branch_id`. This supports both roles on the same endpoint set.
3. **Archive action**: `POST /sites/:site_id/rooms/:room_id/actions/archive` (returns 409 if children assigned). Reactivation: `POST /sites/:site_id/rooms/:room_id/actions/activate`.
4. **No DELETE endpoint**: Policy is archive, not hard delete. The spec's DELETE suggestion is rejected in favor of archive.
5. **Database**: `branch_id` column (not `site_id`) for consistency with existing `branches` table. Migration `000029`.
6. **Duplicate name enforcement**: Application-layer check before create/update. Unique constraint `UNIQUE (branch_id, name) WHERE is_active = true` (partial index) if PostgreSQL supports it; otherwise enforced in service layer with a query check.
7. **Auditing**: Room create, update, archive, and reactivate are audit-significant. Follow existing `audit.Writer` pattern.
8. **Frontend placement**: Add room management under the owner feature as `web/src/app/features/owner/pages/owner-rooms/`. Manager also sees the page linked from their sidebar but scoped to their single site.
9. **Migration number**: `000029` (next after `000028_split_child_names`).
10. **sqlc**: Queries in `api/db/query/rooms.sql`, run `make sqlc-generate` after writing.

## Acceptance Criteria

1. **Create room**: POST to `/api/v1/sites/:site_id/rooms` with name, age_group, capacity, description creates a room and returns it with 201.
2. **List rooms**: GET `/api/v1/sites/:site_id/rooms` returns all active rooms for the site (owner: any site; manager: only their site). Supports `?include_archived=true` query param.
3. **Get room detail**: GET `/api/v1/sites/:site_id/rooms/:room_id` returns the room.
4. **Update room**: PATCH `/api/v1/sites/:site_id/rooms/:room_id` updates fields. Changing name to an existing active room name in the same site returns 409.
5. **Archive room**: POST to archive endpoint sets `is_active = false`. Returns 409 if active children assigned.
6. **Reactivate room**: POST to activate endpoint restores `is_active = true`.
7. **Tenant isolation**: Room A in tenant 1, site X is invisible to tenant 2 even if tenant 2 has a site with same name.
8. **Cross-site name uniqueness**: "Baby Room" in Site A and "Baby Room" in Site B of the same tenant coexist without conflict.
9. **Duplicate active name blocked**: Creating/updating to a name that already exists as an active room in the same site returns 409.
10. **Permission enforcement**: Parent receives 403; staff receives 403 for write operations, 200 for read; manager can only access their assigned site; owner can access any site.
11. **Frontend**: Owner sees room management page with site picker; manager sees room management page for their site; staff sees read-only room list; parent sees nothing.
12. **Validation**: name required, site_id required, capacity required (>0), age_group required (must be one of baby/toddler/preschool/mixed).

## Implementation Tasks

### Task 1: Database migration

- Objective: Create the `rooms` table with proper indexes and constraints.
- Depends on: None
- Target files/symbols:
  - `api/db/migrations/000029_add_rooms.up.sql`
  - `api/db/migrations/000029_add_rooms.down.sql`
- Required changes:
  - Create `rooms` table: `id UUID PRIMARY KEY`, `tenant_id UUID NOT NULL REFERENCES tenants(id)`, `branch_id UUID NOT NULL REFERENCES branches(id)`, `name TEXT NOT NULL`, `description TEXT`, `age_group TEXT NOT NULL`, `capacity INT NOT NULL`, `is_active BOOLEAN NOT NULL DEFAULT true`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
  - Add indexes: `idx_rooms_tenant_id` (tenant_id), `idx_rooms_branch_id` (branch_id), `idx_rooms_active` (branch_id, is_active), `idx_rooms_branch_active_name` (branch_id, name) WHERE is_active = true
  - Add unique partial constraint: `CREATE UNIQUE INDEX idx_rooms_active_name_per_branch ON rooms (branch_id, name) WHERE is_active = true;`
  - Down migration drops the table
- Tests/verification: `make migrate-up` and `make migrate-down` succeed; `make migrate-verify` passes if VERIFY_DATABASE_URL configured.
- Expected outcome: Table exists with all indexes and the partial unique index for active room name uniqueness per branch.

### Task 2: SQL queries (sqlc)

- Objective: Write named SQL queries for room CRUD operations.
- Depends on: Task 1
- Target files/symbols:
  - `api/db/query/rooms.sql` (new)
  - `api/internal/platform/db/sqlc/rooms.sql.go` (generated)
  - `api/internal/platform/db/sqlc/models.go` (auto-updated with Room struct)
- Required changes:
  - `RoomsListByBranch :many` — list rooms for branch, filterable by is_active (default active only, `include_archived` flag for all)
  - `RoomsGetByID :one` — get single room by id, scoped to tenant and branch
  - `RoomsCreate :exec` — insert room
  - `RoomsUpdate :execrows` — update room fields with conditional SET (follow `ChildrenUpdate` pattern)
  - `RoomsArchive :exec` — set `is_active = false`, `updated_at = now()`
  - `RoomsReactivate :exec` — set `is_active = true`, `updated_at = now()`
  - `RoomsCheckActiveNameExists :one` — check if an active room with given name exists in the branch (used for duplicate check, exclude current room ID on update)
  - `RoomsCountActiveChildren :one` — placeholder returning 0 (child-room assignment not yet implemented; returns 0 to enable archive gate without blocking on a future feature)
- Tests/verification: Run `make sqlc-generate`. Confirm `rooms.sql.go` is generated with expected function signatures.
- Expected outcome: Generated Go code with typed query functions.

### Task 3: Domain layer

- Objective: Define domain entities, repository interface, age group constants, and domain errors.
- Depends on: None (can run in parallel with Tasks 1-2)
- Target files/symbols:
  - `api/internal/modules/rooms/domain/entities.go` (new)
  - `api/internal/modules/rooms/domain/repository.go` (new)
- Required changes:
  - `Room` struct: ID, TenantID, BranchID, Name, Description, AgeGroup, Capacity, IsActive, CreatedAt, UpdatedAt (all Go stdlib + uuid types, zero framework imports)
  - `AgeGroup` type and constants: `AgeGroupBaby`, `AgeGroupToddler`, `AgeGroupPreschool`, `AgeGroupMixed`
  - `ValidAgeGroups` map and `IsValidAgeGroup(s string) bool` function
  - `Repository` interface:
    ```go
    type Repository interface {
        ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]Room, error)
        GetByID(ctx context.Context, tenantID, branchID, roomID uuid.UUID) (Room, error)
        Create(ctx context.Context, room Room) error
        Update(ctx context.Context, tenantID, branchID, roomID uuid.UUID, fields map[string]any) (int64, error)
        Archive(ctx context.Context, tx pgx.Tx, tenantID, branchID, roomID uuid.UUID) error
        Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, roomID uuid.UUID) error
        ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeRoomID *uuid.UUID) (bool, error)
        CountActiveChildren(ctx context.Context, tx pgx.Tx, tenantID, branchID, roomID uuid.UUID) (int, error)
        Exists(ctx context.Context, tx pgx.Tx, tenantID, branchID, roomID uuid.UUID) (bool, error)
        GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, roomID uuid.UUID) (Room, error)
    }
    ```
  - Domain errors: `ErrRoomNotFound`, `ErrRoomNameDuplicate`, `ErrRoomHasChildren`, `ErrInvalidAgeGroup`, `ErrSiteNotFound`
  - No `pgx` imports in domain; use `pgx.Tx` alias as `Tx` (follow children module pattern)
- Tests/verification: `AgeGroup` validation unit test (not requiring a test file — the type itself should be testable). Write `api/internal/modules/rooms/domain/entities_test.go` testing `IsValidAgeGroup`.
- Expected outcome: Domain package compiles with zero framework imports.

### Task 4: Postgres repository

- Objective: Implement the domain Repository interface using sqlc-generated queries.
- Depends on: Tasks 2, 3
- Target files/symbols:
  - `api/internal/modules/rooms/infrastructure/postgres/repository.go` (new)
  - `api/internal/modules/rooms/infrastructure/postgres/repository_test.go` (new)
- Required changes:
  - `RoomRepository` struct with `*pgxpool.Pool`
  - `NewRepository(pool *pgxpool.Pool) *RoomRepository`
  - Implement all 10 interface methods using `sqlc.New(r.pool)` per query
  - Follow the `uuidToPgtype`/`pgtypeUUIDToUUID`/`isNoRows` helper patterns from owner repository (`api/internal/modules/owner/infrastructure/postgres/repository.go:211-223`)
  - `Update` method: Follow `ChildrenUpdate` sqlc pattern with `@set_` boolean flags for partial updates
  - `CountActiveChildren`: Return 0 (placeholder — child-room assignment not yet built). Must not return an error that blocks archive.
- Tests/verification:
  - Write integration tests in `repository_test.go`
  - Test cases: Create room → retrieve by ID; list by branch (active only, with archived); update name → duplicate detection; archive → verify is_active = false and blocked with children (placeholder); reactivate; tenant isolation (room from tenant A not visible in tenant B queries)
  - Requires `TEST_DATABASE_URL` env var pointing to a disposable test DB
  - Run with `cd api && go test ./internal/modules/rooms/infrastructure/postgres/ -v`
- Expected outcome: All repository methods work correctly with real PostgreSQL.

### Task 5: Application layer (use cases)

- Objective: Implement use case classes — one file per use case.
- Depends on: Task 3 (domain), Task 4 (repository — for types, but use cases compile against interface)
- Target files/symbols:
  - `api/internal/modules/rooms/application/create_room.go` (new)
  - `api/internal/modules/rooms/application/update_room.go` (new)
  - `api/internal/modules/rooms/application/list_rooms.go` (new)
  - `api/internal/modules/rooms/application/get_room.go` (new)
  - `api/internal/modules/rooms/application/archive_room.go` (new)
  - `api/internal/modules/rooms/application/reactivate_room.go` (new)
  - `api/internal/modules/rooms/application/application_test.go` (new)
- Required changes:
  - `CreateRoom` use case: Validates name, age_group, capacity. Checks `ActiveNameExists`. Creates room with UUIDv7. Returns created Room.
  - `UpdateRoom` use case: Validates changed fields. If name changes, checks `ActiveNameExists` with `excludeRoomID`. Returns updated Room.
  - `ListRooms` use case: Returns rooms for a branch. Accepts `includeArchived` flag.
  - `GetRoom` use case: Returns single room by ID. Returns `ErrRoomNotFound` if missing.
  - `ArchiveRoom` use case: Transactional. Calls `CountActiveChildren` — if > 0, returns `ErrRoomHasChildren`. Otherwise calls `Archive`.
  - `ReactivateRoom` use case: Transactional. Calls `GetByIDForUpdate` to verify existence. Calls `Reactivate`.
  - Each use case accepts actor (role-checking for authorization) and validates:
    - Owner actor: validates `siteID` belongs to tenant (via a `SiteExistsChecker` interface)
    - Manager actor: validates `siteID` matches session `branchID`
  - `SiteExistsChecker` interface: `SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error)` — defined in application package
- Tests/verification:
  - Write mock-based unit tests in `application_test.go` for all 6 use cases
  - Test: create with duplicate name → error; archive with children → error; reactivate archived room; manager with wrong site_id → authorization error; owner with nonexistent site → error; invalid age_group → validation error
  - Run with `cd api && go test ./internal/modules/rooms/application/ -v`
- Expected outcome: All use cases compile and pass tests with mock repositories.

### Task 6: HTTP handler

- Objective: Implement Gin handler with route registration, request parsing, and response mapping.
- Depends on: Task 5
- Target files/symbols:
  - `api/internal/modules/rooms/interfaces/http/handler.go` (new)
  - `api/internal/modules/rooms/interfaces/http/dto.go` (new)
- Required changes:
  - `Handler` struct with use case fields
  - `NewHandler(...)` constructor
  - `WithObservability(logger *slog.Logger) *Handler` (follow children/attendance pattern)
  - `RegisterRoutes(protected *gin.RouterGroup)`:
    ```go
    rooms := protected.Group("")
    rooms.Use(requireRoles("manager", "owner"))
    rooms.GET("/sites/:site_id/rooms", h.listRooms)
    rooms.POST("/sites/:site_id/rooms", h.createRoom)
    rooms.GET("/sites/:site_id/rooms/:room_id", h.getRoom)
    rooms.PATCH("/sites/:site_id/rooms/:room_id", h.updateRoom)
    rooms.POST("/sites/:site_id/rooms/:room_id/actions/archive", h.archiveRoom)
    rooms.POST("/sites/:site_id/rooms/:room_id/actions/activate", h.reactivateRoom)
    ```
  - Each handler:
    1. Parse `site_id` and `room_id` from URL params as UUID
    2. Determine actor: try `OwnerActorFromGinContext`, fallback to `ActorFromGinContext` (for manager). If neither, return 403.
    3. For owner: create `OwnerActor` domain type. For manager: create `ManagerActor` domain type with session branch_id.
    4. Bind JSON body (for POST/PATCH)
    5. Call use case, map domain errors via `httpserver.MapDomainError`, return JSON
  - `requireRoles` function (copy from attendance handler or use existing shared helper in `httpserver` — check if `httpserver.RequireRolesWithObservability` exists)
  - DTO types: request structs (create/update) and response structs with snake_case JSON tags
  - Error mapping: `ErrRoomNotFound` → 404, `ErrRoomNameDuplicate` → 409, `ErrRoomHasChildren` → 409, `ErrInvalidAgeGroup` → 400, validation → 400
- Tests/verification: Write handler tests using `httptest` + gin context (follow existing patterns in `api/internal/modules/attendance/interfaces/http/` if handler tests exist there; otherwise defer to integration testing via bootstrap).
- Expected outcome: All 6 endpoints return correct status codes and JSON.

### Task 7: Wire into bootstrap

- Objective: Register the rooms module in the application bootstrap.
- Depends on: Tasks 4, 5, 6
- Target files/symbols:
  - `api/internal/app/bootstrap/bootstrap.go` (modify)
  - `api/internal/app/bootstrap/adapters.go` (possibly modify — add `SiteExistsChecker` adapter if needed)
- Required changes:
  - Add imports for rooms module packages
  - Instantiate `roomsRepo := roomspostgres.NewRepository(pool)`
  - Create a `SiteExistsChecker` adapter that wraps owner repo's `GetActiveSite` method (the owner repo already has this; create adapter if interface shape differs)
  - Instantiate use cases with repo + site checker
  - Instantiate handler and call `RegisterRoutes(protected)` on the shared route group (NOT on the manager-only or owner-only group — use the shared `protected` group)
  - Note: `requireRoles` is defined in the handler file (follow attendance handler pattern), NOT using the shared `httpserver.RequireRolesWithObservability` (which requires logger/recorder params). Use the inline version for simplicity per existing convention.
- Tests/verification: Compile the API: `cd api && go build ./...`. Run `make run-api` and curl test endpoints.
- Expected outcome: API compiles and starts; room endpoints respond.

### Task 8: Frontend — API service and models

- Objective: Add TypeScript models and API service methods for rooms.
- Depends on: Task 7 (API available)
- Target files/symbols:
  - `web/src/app/features/owner/models/owner.models.ts` (modify — add room interfaces)
  - `web/src/app/features/owner/data/owner-api.service.ts` (modify — add room methods)
- Required changes:
  - Add `Room` interface: id, siteId, name, description, ageGroup, capacity, isActive, createdAt, updatedAt
  - Add API response interfaces with snake_case keys: `ApiRoom`, `ApiRoomListResponse`, `ApiCreateRoomRequest`, `ApiUpdateRoomRequest`
  - Add `OwnerApiService` methods:
    - `listRooms(siteId: string, includeArchived?: boolean): Observable<Room[]>`
    - `getRoom(siteId: string, roomId: string): Observable<Room>`
    - `createRoom(siteId: string, data: { name: string; age_group: string; capacity: number; description?: string }): Observable<Room>`
    - `updateRoom(siteId: string, roomId: string, data: Partial<ApiUpdateRoomRequest>): Observable<Room>`
    - `archiveRoom(siteId: string, roomId: string): Observable<void>`
    - `reactivateRoom(siteId: string, roomId: string): Observable<Room>`
  - Map snake_case API responses to camelCase models (follow existing `mapSite` pattern)
- Tests/verification: Update `owner-api.service.spec.ts` with room method tests.
- Expected outcome: API service compiles, tests pass.

### Task 9: Frontend — Room management page

- Objective: Build the room management page for owner and manager users.
- Depends on: Task 8
- Target files/symbols:
  - `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.ts` (new)
  - `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.html` (new)
  - `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.spec.ts` (new)
- Required changes:
  - Standalone component with CommonModule, FormsModule, RouterModule, PageHeaderComponent, SelectComponent, LoadingStateComponent, EmptyStateComponent, AlertComponent
  - **Owner view**: Site picker dropdown to select a site, then room list with create/edit/archive/reactivate actions
  - **Manager view**: No site picker — rooms shown for their assigned site automatically (read site from auth service's current branch)
  - Room list table: columns for name, age group, capacity, status (active/archived), actions
  - Empty state: "No rooms have been added for this site yet."
  - Form: Inline or modal form for create/edit — fields: name (required), age_group (select: baby/toddler/preschool/mixed), capacity (number, min 1), description (textarea, optional)
  - Archive: Confirmation dialog "Archive this room? Children must be reassigned first." → calls archive endpoint. Show error if children assigned.
  - Reactivate: Button on archived rooms → calls reactivate endpoint
  - Validation: Client-side matches server rules (name required, capacity > 0, age_group required)
  - Toggle to show/hide archived rooms
- Tests/verification:
  - Component spec with mocked API service
  - Test: owner sees site picker, manager doesn't; create room; archive room; validation errors display; empty state shown
  - Run with `cd web && npm test`
- Expected outcome: Room management page functional for both owner and manager roles.

### Task 10: Frontend — Routing and navigation

- Objective: Add room management route and sidebar navigation item.
- Depends on: Task 9
- Target files/symbols:
  - `web/src/app/app.routes.ts` (modify)
  - `web/src/app/core/constants/roles.ts` (modify)
  - `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts` (modify)
- Required changes:
  - Add route: `path: 'owner/rooms'` → `OwnerRoomsComponent`, roles: `['owner']`
  - Add route: `path: 'staff/manager/rooms'` → `OwnerRoomsComponent`, roles: `['manager']` (same component, different route — component detects role to show/hide site picker)
  - Add `ROLE_ROUTES.ownerRooms = '/owner/rooms'` and `ROLE_ROUTES.managerRooms = '/staff/manager/rooms'`
  - Add sidebar icon: use `heroBuildingOffice2` or `heroHomeModern` from `@ng-icons/heroicons/outline` for a room/building icon
  - Update `SidebarIcon` type with `'rooms'`
  - Add nav item: Owner sidebar gets "Rooms" under a new "Setup" group (or alongside "Manager access"); Manager sidebar gets "Rooms" under a new "Setup" group
  - Import the icon and register it in providers
- Tests/verification: Navigate to `/owner/rooms` as owner — page renders. Navigate to `/staff/manager/rooms` as manager — page renders with site auto-detected. Navigate as parent — redirected to parent home.
- Expected outcome: Navigation works for authorized roles; blocked for parent.

### Task 11: Staff read-only room view

- Objective: Allow practitioners to view rooms (read-only) for their assigned site.
- Depends on: Task 6 (handler), Task 10 (routing)
- Target files/symbols:
  - `api/internal/modules/rooms/interfaces/http/handler.go` (modify — add practitioner to allowed roles for GET endpoints)
  - `api/internal/modules/rooms/application/list_rooms.go` (modify — allow practitioner actor for list/get)
  - `web/src/app/app.routes.ts` (modify — add practitioner route)
  - `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts` (modify — add to practitioner nav)
- Required changes:
  - GET routes: Allow `requireRoles("manager", "owner", "practitioner")` for list and get endpoints
  - POST/PATCH/archive/activate routes: Keep `requireRoles("manager", "owner")` — practitioners cannot write
  - Application layer: Practice actor passes through for read-only use cases
  - Frontend: Add `path: 'staff/practitioner/rooms'` → `OwnerRoomsComponent` with `roles: ['manager', 'owner', 'practitioner']` (component detects practitioner role → read-only mode, no edit/create buttons)
  - Sidebar: Add "Rooms" to practitioner nav group
- Tests/verification: Login as practitioner → see room list, no create/edit/archive buttons. Attempt direct API call to POST rooms as practitioner → 403.
- Expected outcome: Staff can view rooms; write operations blocked.

## Contracts

### API Contract

**List Rooms**
```
GET /api/v1/sites/:site_id/rooms?include_archived=true|false
Response 200: { rooms: [{ id, name, description, age_group, capacity, is_active, created_at, updated_at }] }
```

**Create Room**
```
POST /api/v1/sites/:site_id/rooms
Body: { name (required), age_group (required, enum), capacity (required, int>0), description? }
Response 201: { id, name, description, age_group, capacity, is_active, created_at, updated_at }
Errors: 400 (validation), 409 (duplicate name), 403 (forbidden)
```

**Get Room**
```
GET /api/v1/sites/:site_id/rooms/:room_id
Response 200: { id, name, description, age_group, capacity, is_active, created_at, updated_at }
Errors: 404, 403
```

**Update Room**
```
PATCH /api/v1/sites/:site_id/rooms/:room_id
Body: { name?, age_group?, capacity?, description? }
Response 200: { id, name, description, age_group, capacity, is_active, created_at, updated_at }
Errors: 400, 404, 409 (duplicate name), 403
```

**Archive Room**
```
POST /api/v1/sites/:site_id/rooms/:room_id/actions/archive
Response 200: {}
Errors: 404, 409 (has active children), 403
```

**Reactivate Room**
```
POST /api/v1/sites/:site_id/rooms/:room_id/actions/activate
Response 200: { id, name, description, age_group, capacity, is_active: true, created_at, updated_at }
Errors: 404, 403
```

### Authorization Contract

| Role        | List | Get | Create | Update | Archive | Reactivate |
|-------------|------|-----|--------|--------|---------|------------|
| Owner       | Any site in tenant | Any site | Any site | Any site | Any site | Any site |
| Manager     | Own site only | Own site only | Own site only | Own site only | Own site only | Own site only |
| Practitioner| Own site only | Own site only | Forbidden | Forbidden | Forbidden | Forbidden |
| Parent      | Forbidden | Forbidden | Forbidden | Forbidden | Forbidden | Forbidden |

### Validation Contract
- `name`: required, non-empty, max 255 chars
- `age_group`: required, must be one of `baby`, `toddler`, `preschool`, `mixed`
- `capacity`: required, integer > 0
- `description`: optional, max 1000 chars

### Error Contract (stable codes)
- `room_not_found` — 404
- `room_name_duplicate` — 409 (active room with same name exists in site)
- `room_has_children` — 409 (cannot archive room with active child assignments)
- `invalid_age_group` — 400
- `site_not_found` — 404
- `forbidden_role` — 403
- `forbidden_site_scope` — 403 (manager accessing wrong site)

## Files to Change

### New files
1. `api/db/migrations/000029_add_rooms.up.sql`
2. `api/db/migrations/000029_add_rooms.down.sql`
3. `api/db/query/rooms.sql`
4. `api/internal/modules/rooms/domain/entities.go`
5. `api/internal/modules/rooms/domain/entities_test.go`
6. `api/internal/modules/rooms/domain/repository.go`
7. `api/internal/modules/rooms/infrastructure/postgres/repository.go`
8. `api/internal/modules/rooms/infrastructure/postgres/repository_test.go`
9. `api/internal/modules/rooms/application/create_room.go`
10. `api/internal/modules/rooms/application/update_room.go`
11. `api/internal/modules/rooms/application/list_rooms.go`
12. `api/internal/modules/rooms/application/get_room.go`
13. `api/internal/modules/rooms/application/archive_room.go`
14. `api/internal/modules/rooms/application/reactivate_room.go`
15. `api/internal/modules/rooms/application/application_test.go`
16. `api/internal/modules/rooms/interfaces/http/handler.go`
17. `api/internal/modules/rooms/interfaces/http/dto.go`
18. `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.ts`
19. `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.html`
20. `web/src/app/features/owner/pages/owner-rooms/owner-rooms.component.spec.ts`

### Modified files
21. `api/internal/app/bootstrap/bootstrap.go` — wire rooms module
22. `api/internal/platform/db/sqlc/models.go` — auto-generated Room struct
23. `api/internal/platform/db/sqlc/rooms.sql.go` — auto-generated query functions
24. `CONTEXT.md` — already updated with Room terms
25. `web/src/app/app.routes.ts` — add room routes
26. `web/src/app/core/constants/roles.ts` — add room route constants
27. `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts` — add nav items + icon
28. `web/src/app/features/owner/models/owner.models.ts` — add room interfaces
29. `web/src/app/features/owner/data/owner-api.service.ts` — add room API methods
30. `web/src/app/features/owner/data/owner-api.service.spec.ts` — add room tests

## Verification

### Backend
```bash
# Migration
make migrate-up
make migrate-verify  # if VERIFY_DATABASE_URL set

# Generate sqlc
make sqlc-generate

# Compile
cd api && go build ./...

# Domain tests
cd api && go test ./internal/modules/rooms/domain/ -v

# Application tests (mock repos)
cd api && go test ./internal/modules/rooms/application/ -v

# Repository integration tests (needs TEST_DATABASE_URL)
cd api && go test ./internal/modules/rooms/infrastructure/postgres/ -v

# Manual curl testing (after make run-api)
curl -X POST http://localhost:8080/api/v1/sites/{site_id}/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Baby Room","age_group":"baby","capacity":12}'

curl http://localhost:8080/api/v1/sites/{site_id}/rooms \
  -H "Authorization: Bearer $TOKEN"

curl -X PATCH http://localhost:8080/api/v1/sites/{site_id}/rooms/{room_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"capacity":15}'

curl -X POST http://localhost:8080/api/v1/sites/{site_id}/rooms/{room_id}/actions/archive \
  -H "Authorization: Bearer $TOKEN"

curl -X POST http://localhost:8080/api/v1/sites/{site_id}/rooms/{room_id}/actions/activate \
  -H "Authorization: Bearer $TOKEN"
```

### Frontend
```bash
cd web && npm test  # Karma unit tests
cd web && npm start # Manual testing at :4200
```

### Manual validation scenarios
1. Login as owner → navigate to /owner/rooms → site picker shows all sites → select site → see room list → create room → edit room → archive room → reactivate room
2. Login as manager → navigate to /staff/manager/rooms → sees rooms for assigned site (no site picker) → create/edit/archive/reactivate
3. Login as practitioner → navigate to /staff/practitioner/rooms → sees read-only room list → no create/edit/archive buttons
4. Login as parent → navigate to /owner/rooms → redirected to parent home
5. Create "Baby Room" in Site A → create "Baby Room" in Site B → both succeed
6. Create "Baby Room" again in Site A → 409 error
7. Login as manager for Site A → cannot see rooms of Site B
8. Archive a room → verify is_active = false → verify room hidden from default list
9. Toggle "show archived" → verify archived room appears

## Assumptions

1. **Child-room assignment not yet implemented**: `CountActiveChildren` returns 0 and does not block archiving. The field/relationship will be added later. The archive gate logic is in place but won't trigger until children are assignable to rooms.
2. **Site existence check**: The owner repository already has `GetActiveSite(ctx, tenantID, siteID)` which can be reused via a small adapter. No new site query is needed.
3. **`requireRoles` helper**: Copied inline in the handler (follow attendance handler pattern at `api/internal/modules/attendance/interfaces/http/handler.go:224-249`) rather than extracted to a shared location.
4. **UUID generation**: Use `uuid.Must(uuid.NewV7())` for room IDs (follows project preference for UUIDv7 per "Entity Identifier Strategy").
5. **API prefix**: Routes register on the `protected` gin group which is already at `/api/v1` (via `cfg.APIBasePath` in bootstrap).
6. **Description field**: Optional free-text with no length constraint beyond DB defaults (use TEXT column).
7. **Manager room routes**: Manager uses same component as owner but the component detects role from `AuthService.currentRole()` to show/hide site picker and scope data appropriately. The manager's site_id is read from the auth session's branch_id.
8. **No practitioner navigation addition**: Per the spec, "Staff sees read-only room list if needed." "If needed" means we add the route and handler access, but sidebar navigation for practitioner is included for completeness.
9. **Owner repository reuse**: The `SiteExistsChecker` adapter wraps `ownerpostgres.NewRepository(pool).GetActiveSite()` — this creates a second repository instance but both share the same pool, which is acceptable per existing bootstrap patterns.

## Risks and Fallbacks

| Risk | Fallback |
|------|----------|
| Partial unique index `WHERE is_active = true` not supported by PostgreSQL version | Enforce uniqueness entirely in application layer with `ActiveNameExists` query check before create/update |
| `make migrate-verify` fails due to missing `VERIFY_DATABASE_URL` | Skip verification; manual visual inspection of migration SQL is sufficient |
| Frontend `OwnerRoomsComponent` used by multiple roles causes role-detection complexity | If role-based branching becomes unwieldy, split into separate `OwnerRoomsComponent` and `ManagerRoomsComponent` |
| `requireRoles` inline function conflicts with `httpserver.RequireRolesWithObservability` | Use the inline version exclusively (follows existing pattern in attendance handler) |
| sqlc generation fails due to missing tool directive | Run `go tool sqlc generate` directly instead of `make sqlc-generate` |
