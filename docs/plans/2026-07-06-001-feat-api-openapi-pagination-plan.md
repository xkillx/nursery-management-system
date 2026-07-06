---
title: API OpenAPI & Pagination - Plan
type: feat
date: "2026-07-06"
topic: api-openapi-pagination
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-brainstorm
execution: code
---

## Goal Capsule

- **Objective:** Replace the hand-written OpenAPI spec with auto-generated documentation from handler annotations, and add offset-based pagination to all collection endpoints with a standardized response envelope.
- **Product Authority:** Nursery managers and developers integrating with the API.
- **Open Blockers:** None.
- **Execution profile:** Backend-first (Go + sqlc), then frontend (Angular). Each implementation unit is an atomic commit.
- **Tail ownership:** Verify generated OpenAPI spec covers all endpoints; run `go vet`, `go build`, `npm run lint`, `ng build`.

## Product Contract

Product Contract unchanged — all decisions carried forward from brainstorm.

### Summary

Add swaggo/swag annotations to all API handlers so the OpenAPI spec generates from code, eliminating spec drift. Add offset-based pagination to the 14 unpaginated collection endpoints and migrate the 3 existing paginated endpoints, standardizing all collection responses on `{ items, total, page, page_size }`.

### Problem Frame

The existing hand-written OpenAPI spec at `docs/API-CONTRACT.openapi.yaml` (4774 lines) diverges from actual code — it omits pagination metadata (`total`, `limit`, `offset`) that handlers like `children/handler.go:256` and `billing/handler.go:281-288` already return. Without code-generation tooling, this gap widens with every change.

14 of 18 collection endpoints return unbounded arrays with no row limiting at the SQL layer. As data grows, these endpoints will return increasingly large payloads. The three different response wrapping patterns (`{ "rooms": [...] }`, `{ "items": [...] }`, bare `[...]`) make client consumption inconsistent.

### Key Decisions

**Code-gen from swaggo annotations.** The hand-written spec is already stale. Generating from code ensures the spec is always current. Spec-first (oapi-codegen) was considered but rejected — it inverts the relationship and requires higher upfront effort for a team that writes handlers first.

**Offset-based pagination everywhere.** Cursor-based pagination was considered for attendance sessions (real-time data), but most nursery datasets are bounded (rooms, children per branch, session types). Offset-based is simpler, consistent with existing children/billing pagination, and sufficient for expected data volumes.

**Standardize on `{ items, total, page, page_size }`.** Three response patterns exist today. Standardizing breaks some existing consumers but eliminates inconsistency. The alternative (adding pagination metadata to each existing wrapper) preserves inconsistency permanently.

**Default page_size = 50.** Matches the existing children endpoint (`children/handler.go:239`). Clients can override via query parameter. Max page_size = 200 (matches existing billing validation in `billing/application/list_invoices.go`).

### Requirements

**OpenAPI Generation**

- R1. Add `swaggo/swag` as a dependency and configure it to generate `docs/API-CONTRACT.openapi.yaml` from handler annotations.
- R2. Annotate every handler with `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, and `@Success`/`@Failure` swag comments.
- R3. Include pagination parameters (`page`, `page_size`) in annotations for all collection endpoints.
- R4. Include role-based access annotations using `x-roles` custom extension, matching the existing spec pattern.
- R5. Remove the hand-written `docs/API-CONTRACT.openapi.yaml` after the generated spec is verified to cover all endpoints.

**Pagination**

- R6. All collection endpoints accept optional `page` (default 1) and `page_size` (default 50, max 200) query parameters.
- R7. All collection endpoints return `{ "items": [...], "total": N, "page": N, "page_size": N }` where `total` is the full unfiltered count. Endpoints with nested summary structures (`owner/site-summaries`, `funding/overview`) paginate the top-level response only; inner arrays remain unpaginated.
- R8. When `page` and `page_size` are omitted, endpoints return the first 50 records (not all records) — preserving current children endpoint behavior.
- R9. The 14 unpaginated endpoints (`rooms`, `session-types`, `session-templates`, `academic-terms`, `closure-days`, `ad-hoc-bookings`, `invites`, `children/:id/terms`, `terms` (expiring), `owner/site-summaries`, `owner/manager-access`, `funding/overview`, `children/:id/room-assignments`, `children/:id/booking-patterns`) get SQL-level `LIMIT`/`OFFSET` via their sqlc queries. `attendance/sessions` and `attendance/sessions/:id/history` are excluded — results are scoped to a child + date and typically return 1-3 records.
- R10. The `owner/manager-access` endpoint changes from bare array `[...]` to `{ "items": [...], "total": N, "page": N, "page_size": N }`.

**Frontend Alignment**

- R11. Update all Angular services/components that consume collection endpoints to use the new `{ items, total, page, page_size }` envelope.
- R12. Replace client-side `visibleCount` slicing patterns in `manager-rooms.component.ts` and `owner-rooms.component.ts` with server-driven pagination.

### Scope Boundaries

**Deferred for later (P1+):**
- API versioning (`/api/v1/` prefix).
- Global rate limiting expansion.
- Error response standardization.
- Cursor-based pagination for attendance sessions.
- Sorting and filtering on collection endpoints.

**Not in scope:**
- Contract tests for response shape stability.
- OpenAPI UI (Swagger UI) hosting.

### Acceptance Examples

- AE1. `GET /children` without `page` or `page_size` returns `{ "items": [...50 items...], "total": 120, "page": 1, "page_size": 50 }`.
- AE2. `GET /children?page=2&page_size=25` returns items 26-50, with `"page": 2, "page_size": 25, "total": 120`.
- AE3. `GET /children?page=10&page_size=50` when total=30 returns `{ "items": [], "total": 30, "page": 10, "page_size": 50 }`.
- AE4. `GET /owner/manager-access` returns `{ "items": [...], "total": N, "page": 1, "page_size": 50 }` instead of bare `[...]`.
- AE5. `GET /sites/:id/rooms` returns `{ "items": [...], "total": N, "page": 1, "page_size": 50 }` instead of `{ "rooms": [...] }`.

### Dependencies / Assumptions

- The Angular frontend in `web/` is updated in the same PR as the backend changes (same repo, coordinated release).
- sqlc supports LIMIT/OFFSET parameterization — confirmed by existing usage in `children/handler.go` and `billing/handler.go`.
- The `swaggo/swag` tool is compatible with Go 1.26 and Gin v1.12.0.
- `total` reflects the count matching the current query's WHERE clause. Since no filtering exists today, this is equivalent to the full table count. When filtering is added, the count query will include the same WHERE clause.

### Outstanding Questions

- Q3. Max page_size = 200 (resolved — matches existing billing validation).
- Q4. `total` reflects unfiltered count (resolved — filtering not yet implemented; count query uses same WHERE as list query).

### Sources / Research

- Existing OpenAPI spec: `docs/API-CONTRACT.openapi.yaml` (4774 lines, hand-written, version 0.2.0).
- Existing pagination: `children/interfaces/http/handler.go` (limit/offset with total), `billing/interfaces/http/handler.go` (limit/offset without total).
- Pagination helpers: `children/interfaces/http/handler.go` (`parseIntQuery()`), `billing/interfaces/http/handler.go` (`queryParamPtr()`).
- sqlc patterns: `internal/platform/db/sqlc/children.sql` (LIMIT/OFFSET + COUNT), `internal/platform/db/sqlc/invoices.sql` (nullable LIMIT/OFFSET).
- Response wrapping: `rooms/interfaces/http/handler.go` (named key), `invites/interfaces/http/handler.go` (items), `owner/interfaces/http/handler.go` (bare array).
- Auth handler test pattern: `authentication/interfaces/http/handler_test.go` (fake repos, httptest, gin test context).
- Frontend pagination: `web/src/app/features/staff/data/staff-api.service.ts` (children), `web/src/app/features/owner/owner-rooms.component.ts` (visibleCount).

---

## Planning Contract

### Key Technical Decisions

**KTD1: Shared pagination helper module.** Create a shared `internal/platform/http/pagination/` package with `ParsePageParams(c *gin.Context) (page, pageSize int)` and `PaginatedResponse(items interface{}, total, page, pageSize int) gin.H`. Replaces the per-module `parseIntQuery` and `queryParamPtr` helpers. Centralizes validation (default 50, max 200, min 1) and response envelope construction.

**KTD2: sqlc LIMIT/OFFSET via named params.** Follow the `invoices.sql` pattern (`sqlc.narg('limit') sqlc.narg('offset')`) rather than positional params (`$3 $4`). Named params are more readable and less brittle when WHERE clauses change. Each unpaginated query gets a paginated variant plus a companion COUNT query.

**KTD3: Application-layer offset calculation.** Handlers accept `page`/`page_size` from the client but compute `offset = (page - 1) * page_size` before passing to the repository. The repository interface takes `limit`/`offset` (not `page`/`page_size`) — this keeps the domain layer API-agnostic.

**KTD4: Migrate existing 3 endpoints in-place.** Children, manager invoices, and parent invoices already have pagination but return different shapes. Migrate them to the shared helper and new envelope rather than leaving them as exceptions. Children's `{items, total}` is closest; billing's `{items, limit, offset}` needs the most change.

**KTD5: swaggo annotations as a dedicated pass.** Add annotations to all handlers in one unit after pagination is wired, rather than mixing annotation with pagination logic per handler. This keeps the annotation review focused and avoids merge conflicts across modules.

### Assumptions

- The `app-table-pagination` Angular component (mentioned in design system) can be reused for server-driven pagination without modification.
- sqlc's `sqlc.narg` works with `int` type (not just `*int`) — needs verification during implementation.
- The existing `swag init` command can parse all handler modules without special configuration (the `-d ./api --parseInternal --parseDependency` flags should suffice).

### Implementation Constraints

- **Backend-first:** All SQL + handler changes land before frontend changes. Frontend must not be updated until the backend returns the new envelope.
- **Atomic per-module:** Each module's SQL + handler + test changes land as one commit. Do not split a module's changes across commits.
- **No dual-mode:** Endpoints switch directly to the new envelope. No backward-compatible period with both old and new shapes.

---

## Implementation Units

### U1. Shared Pagination Helper

- **Goal:** Create a reusable pagination package that all collection handlers use.
- **Requirements:** R6, R7, R8
- **Dependencies:** None
- **Files:** `api/internal/platform/http/pagination/pagination.go`, `api/internal/platform/http/pagination/pagination_test.go`
- **Approach:** Create `ParsePageParams(c *gin.Context) (page, pageSize int)` that reads `page` and `page_size` query params with defaults (page=1, pageSize=50) and validation (min 1, max 200). Create `PaginatedResponse(items interface{}, total, page, pageSize int) gin.H` that returns the standard envelope `{items, total, page, pageSize}`. Follow existing platform package conventions in `api/internal/platform/`.
- **Patterns to follow:** `children/interfaces/http/handler.go:872` (`parseIntQuery`), `billing/application/list_invoices.go:48-124` (validation logic, max 200).
- **Test scenarios:**
  - Happy path: `ParsePageParams` with `?page=2&page_size=25` returns `(2, 25)`.
  - Defaults: no query params returns `(1, 50)`.
  - Boundary: `page_size=0` clamps to 1, `page_size=300` clamps to 200, `page=0` clamps to 1.
  - Response: `PaginatedResponse(items, 120, 2, 25)` returns `{items, total: 120, page: 2, page_size: 25}`.
- **Verification:** `go test ./api/internal/platform/http/pagination/...` passes.

### U2. Swaggo Tooling Setup

- **Goal:** Add swaggo as a dependency and configure the build pipeline.
- **Requirements:** R1
- **Dependencies:** None
- **Files:** `api/go.mod`, `api/go.sum`, `Makefile`, `api/cmd/server/main.go`, `api/docs/` (generated)
- **Approach:** Add `github.com/swaggo/swag` as a tool dependency. Add `make swagger-generate` target that runs `swag init -g api/cmd/server/main.go -d ./api --parseInternal --parseDependency -o ./api/docs`. Add general API info annotations to `main.go` (`@title`, `@version`, `@host`, `@BasePath`, `@securityDefinitions.apikey`). Add `make swagger-validate` target that runs `swag fmt` and checks for formatting drift. Import generated docs package in main.go.
- **Patterns to follow:** Existing Makefile targets for `sqlc-generate`.
- **Test expectation:** none — tooling setup, verified by `make swagger-generate` producing output without errors.
- **Verification:** `make swagger-generate` produces `api/docs/swagger.json` and `api/docs/swagger.yaml` without errors.

### U3. SQL & Repository Layer for Unpaginated Endpoints

- **Goal:** Add LIMIT/OFFSET and COUNT queries for all 12 unpaginated collection endpoints.
- **Requirements:** R9
- **Dependencies:** None
- **Files:** `api/internal/platform/db/sqlc/` (query files), `api/internal/platform/db/sqlc/` (generated), module domain repositories
- **Approach:** For each of the 12 endpoints, add a paginated variant of the existing `:many` query with `LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset')` and a companion `:one` COUNT query with identical WHERE clause. Run `make sqlc-generate` to regenerate Go code. Update repository interfaces to accept `limit`/`offset` params where needed.

  Modules affected:
  - `rooms.sql` — `RoomsListBySite` + count
  - `session_types.sql` — `SessionTypesListBySite` + count
  - `session_templates.sql` — `SessionTemplatesListBySite` + count
  - `academic_terms.sql` — `AcademicTermsListBySite` + count
  - `branch_closure_days.sql` — `BranchClosureDaysListBySite` + count
  - `ad_hoc_bookings.sql` — `AdHocBookingsListByChild` + count
  - `manager_invites.sql` — `ManagerInvitesList` + count
  - `child_booking_patterns.sql` — `ChildBookingPatternsList` + count
  - `child_room_assignments.sql` — `ChildRoomAssignmentsList` + count
  - `children.sql` — `ChildrenTermsList` + count (for `GET /children/:id/terms`)
  - `term.sql` — `TermsExpiringList` + count (for `GET /terms`)
  - `owner.sql` — `OwnerManagerAccessList` + count, `OwnerSiteSummariesList` + count
  - `funding.sql` — `FundingOverviewList` + count (for `GET /funding/overview`)

- **Patterns to follow:** `children.sql:56` (LIMIT/OFFSET), `children.sql:191` (COUNT query), `invoices.sql:304` (sqlc.narg pattern).
- **Test expectation:** none — SQL layer changes are verified by handler-level tests in U4. If existing integration tests cover these queries, verify LIMIT/OFFSET/COUNT behavior through those. Otherwise, add at least one integration test per module verifying LIMIT limits rows, OFFSET skips, and COUNT returns expected total.
- **Verification:** `make sqlc-generate` succeeds. Generated Go code compiles (`go build ./...`).

### U4. Handler Pagination for Unpaginated Endpoints

- **Goal:** Wire pagination params and response envelope into all 12 unpaginated collection handlers.
- **Requirements:** R6, R7, R8, R10
- **Dependencies:** U1, U3
- **Files:** Module handler files (rooms, session types, session templates, invites, owner, ad hoc bookings, term calendar, branch closures, children)
- **Approach:** For each handler: call `pagination.ParsePageParams(c)` to get page/pageSize, compute `offset = (page - 1) * pageSize`, pass limit/offset to the use case, call the count query, return `pagination.PaginatedResponse(items, total, page, pageSize)`. Remove old response wrapping (`gin.H{"rooms": ...}`, bare arrays, etc.). For `owner/manager-access`, change from bare array to paginated envelope.

  Key handler changes:
  - `rooms/interfaces/http/handler.go:114` — `{"rooms": [...]}` → paginated envelope
  - `sessiontypes/interfaces/http/handler.go:111` — `{"session_types": [...]}` → paginated
  - `sessiontemplates/interfaces/http/handler.go:111` — `{"session_templates": [...]}` → paginated
  - `invites/interfaces/http/handler.go:117` — `{"items": [...]}` → paginated (add total)
  - `owner/interfaces/http/handler.go:207` — bare `[...]` → paginated
  - `owner/interfaces/http/handler.go:71` — bare `[...]` → paginated (top-level only)
  - `ad_hoc_bookings/interfaces/http/handler.go:142` — `{"ad_hoc_bookings": [...]}` → paginated
  - `term_calendar/interfaces/http/handler.go:84` — `{"academic_terms": [...]}` → paginated
  - `branch_closures/interfaces/http/handler.go:115` — `{"closure_days": [...]}` → paginated
  - `children/interfaces/http/handler.go:362` — attendance stays unpaginated (excluded per R9)
  - `children/interfaces/http/handler.go:629` — room assignments → paginated
  - `children/interfaces/http/handler.go:739` — booking patterns → paginated

- **Patterns to follow:** `children/interfaces/http/handler.go:231-256` (existing pagination pattern with count).
- **Test scenarios (per module, representative):**
  - Happy path: GET collection with default params returns `{items, total, page: 1, page_size: 50}`.
  - Custom page: GET with `?page=2&page_size=10` returns correct slice.
  - Empty: GET with `?page=99` returns `{items: [], total: N, page: 99, page_size: 50}`.
  - Boundary: `page_size=0` clamps to 1, `page_size=999` clamps to 200.
- **Verification:** `go test ./api/internal/modules/...` passes for affected modules. `go vet ./...` and `go build ./...` clean.

### U5. Migrate Existing Paginated Endpoints

- **Goal:** Align children, manager invoices, and parent invoices with the shared pagination helper and new envelope.
- **Requirements:** R6, R7, R8
- **Dependencies:** U1
- **Files:** `children/interfaces/http/handler.go`, `billing/interfaces/http/handler.go`, `billing/interfaces/http/dto.go`
- **Approach:**
  - Children: replace `parseIntQuery` calls with `pagination.ParsePageParams`. Add `page`/`page_size` to response (currently returns `{items, total}` only). Keep existing LIMIT/OFFSET + COUNT SQL.
  - Manager invoices: replace `queryParamPtr` calls with `pagination.ParsePageParams`. Add COUNT query. Update `invoiceListResponse` struct from `{items, limit, offset}` to `{items, total, page, page_size}`.
  - Parent invoices: same changes as manager invoices.
- **Patterns to follow:** U1's shared helper.
- **Test scenarios:**
  - Children: `GET /children` returns `{items, total, page: 1, page_size: 50}` (page/page_size are new).
  - Invoices: `GET /invoices` returns `{items, total, page, page_size}` instead of `{items, limit, offset}`.
  - Parent invoices: `GET /parent/invoices` same as above.
- **Verification:** `go test ./api/internal/modules/children/... ./api/internal/modules/billing/...` passes.

### U6. Swaggo Annotations on All Handlers

- **Goal:** Add swaggo comment annotations to every handler method so the generated OpenAPI spec is complete.
- **Requirements:** R2, R3, R4
- **Dependencies:** U2, U4, U5 (annotations reference the new response types and pagination params)
- **Files:** All module handler files (18 modules), `api/docs/` (regenerated)
- **Approach:** Add annotation blocks above each handler method. Use `@Param page query int false "Page number" default(1) minimum(1)` and `@Param page_size query int false "Items per page" default(50) minimum(1) maximum(200)` on collection endpoints. Use `@x-roles ["manager"]` etc. for role annotations. Reference response DTOs in `@Success` annotations. Run `make swagger-generate` after all annotations are added.

  Handler modules to annotate:
  - `authentication`, `children`, `attendance`, `rooms`, `sessiontypes`, `sessiontemplates`, `invites`, `billing`, `payments`, `funding`, `term`, `term_calendar`, `ad_hoc_bookings`, `branch_closures`, `owner`, `siteprofile`, `mappings`, `passwordreset`

- **Patterns to follow:** swaggo annotation format from research (`@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`, `@Security`, `@x-roles`).
- **Test expectation:** none — annotations are documentation, not behavior.
- **Verification:** `make swagger-generate` succeeds. Generated spec at `docs/API-CONTRACT.openapi.yaml` contains all expected paths and schemas. `make swagger-validate` passes.

### U7. Frontend Service Updates

- **Goal:** Update Angular services that consume collection endpoints to handle the new `{ items, total, page, page_size }` envelope.
- **Requirements:** R11
- **Dependencies:** U4, U5 (backend must return new envelope first)
- **Files:** `web/src/app/features/staff/data/staff-api.service.ts`, `web/src/app/features/staff/data/staff-rooms-api.service.ts`, other API services consuming collection endpoints
- **Approach:** Update response type interfaces to include `total`, `page`, `page_size`. Update `map`/`pipe` operators to extract from the new envelope. Services that already use `{items, total}` (children) need `page`/`page_size` added. Services that use named keys (`res.rooms`) or bare arrays need full migration to `res.items`.
- **Patterns to follow:** `staff-api.service.ts:177-185` (existing children pattern with `items`/`total`).
- **Test scenarios:**
  - Children service: response with `page`/`page_size` fields is correctly mapped.
  - Rooms service: response with `{items: [...]}` replaces `{rooms: [...]}`.
  - Manager access service: response with `{items: [...]}` replaces bare array.
- **Verification:** `npm run lint` passes. `ng build` (production) succeeds with zero errors and warnings.

### U8. Frontend visibleCount Migration

- **Goal:** Replace client-side `visibleCount` slicing in rooms components with server-driven pagination.
- **Requirements:** R12
- **Dependencies:** U7
- **Files:** `web/src/app/features/owner/owner-rooms.component.ts`, `web/src/app/features/manager/manager-rooms.component.ts` (or equivalent paths)
- **Approach:** Remove `visibleCount`, `pagedRows`, `showLoadMore`, `loadMore()` properties. Replace with `page`, `pageSize`, `total` state. Wire pagination controls to `page`/`page_size` query params on API calls. Use the existing `app-table-pagination` component if available in the design system.
- **Patterns to follow:** Existing children list component (already server-paginated).
- **Test scenarios:**
  - Initial load: component requests page 1 with default page_size.
  - Next page: clicking next increments page, triggers new API call.
  - Total display: "Showing 1-50 of 120" text updates correctly.
- **Verification:** `npm run lint` passes. `ng build` (production) succeeds with zero errors and warnings. Components render correctly with server-paginated data.

### U9. Remove Old OpenAPI Spec & Final Verification

- **Goal:** Delete the hand-written spec and verify the generated spec is complete.
- **Requirements:** R5
- **Dependencies:** U6
- **Files:** `docs/API-CONTRACT.openapi.yaml` (deleted), `api/docs/` (generated, canonical)
- **Approach:** Delete `docs/API-CONTRACT.openapi.yaml`. Run `make swagger-generate` to produce the final spec. Verify all endpoints are present in the generated spec. Compare path count against registered routes in `buildGinEngine()`. Update any references to the old spec path in docs or config.
- **Test expectation:** none — verification is a manual comparison.
- **Verification:** Generated spec contains all registered routes. `go vet ./...` and `go build ./...` pass. `npm run lint` and `ng build` pass.

---

## Verification Contract

| Gate | Command | Applies to |
|---|---|---|
| Go vet | `go vet ./...` | After every Go file change (run from `api/`) |
| Go build | `go build ./...` | After every Go file change (run from `api/`) |
| Go tests | `go test ./api/internal/...` | After U1, U4, U5 |
| sqlc generate | `make sqlc-generate` | After U3 |
| swagger generate | `make swagger-generate` | After U2, U6, U9 |
| Angular lint | `npm run lint` | After every Angular file change (run from `web/`) |
| Angular build | `ng build` | After U7, U8 (run from `web/`) |

## Definition of Done

- All 16 paginated collection endpoints return `{ items, total, page, page_size }`. The 2 attendance endpoints remain unpaginated per R9.
- All 14 previously unpaginated endpoints have SQL-level LIMIT/OFFSET.
- Generated OpenAPI spec at `docs/API-CONTRACT.openapi.yaml` contains all registered routes.
- Hand-written spec deleted.
- Frontend services use new response envelope.
- Frontend `visibleCount` pattern replaced with server-side pagination.
- `go vet ./...`, `go build ./...`, `go test ./...`, `npm run lint`, `ng build` all pass clean.
- Dead code removed (old `parseIntQuery` helper, old `queryParamPtr` if unused, old response DTOs).
