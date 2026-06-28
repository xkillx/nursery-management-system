---
title: Fix P0 Dependency Rule Violations - Plan
type: refactor
date: 2026-06-28
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
---

## Goal Capsule

Remove two infrastructure types from the domain and application layers where they violate the Clean Architecture dependency rule:
- `pgx.Tx` type alias in domain repository interfaces (8 modules)
- `*pgxpool.Pool` in application use case constructors (3 use cases)

No behavior changes. No schema changes. The dependency rule is the only objective — infrastructure types must not appear outside infrastructure.

---

## Product Contract

### Summary

Refactor to seal two P0 dependency-rule leaks across the Go backend: pull `pgx.Tx` out of all domain repository interfaces (replacing with `any` so the domain imports nothing from `pgx`), and remove `*pgxpool.Pool` from use case constructors (dead code in rooms, push into a transaction wrapper in children).

### Problem Frame

The architecture review found two violations of the inward-only dependency rule. Both types are PostgreSQL-specific (`github.com/jackc/pgx/v5`) leaking into layers that should be framework-agnostic.

The `type Tx = pgx.Tx` alias in 8 module domain packages means the domain layer — which should define its own abstractions — is directly coupled to the pgx library. Every repository interface that uses `tx Tx` transitively pulls pgx into the domain import graph.

The `*pgxpool.Pool` in 3 use case constructors means application logic depends on the pool type from pgx. In ArchiveRoom and ReactivateRoom the pool field is completely unused (dead code). In UpdateChild it's used only to satisfy the audit writer's querier parameter.

### Requirements

R1. No `pgx` import in any module's `domain/` package
R2. No `*pgxpool.Pool` in any module's `application/` package
R3. All existing tests pass without change. U4 introduces a transactional boundary making update+audit atomic — existing observable contract preserved apart from the new atomicity guarantee (no behavioral regression, but the audit write is no longer fire-and-forget on a separate connection)
R4. Compile-time correctness: `go build ./...` passes without error

### Scope Boundaries

**In scope:**
- 8 domain repository interface files: `type Tx = pgx.Tx` → `type Tx any`
- 3 use case files: remove `*pgxpool.Pool` from struct and constructor
- Infrastructure repository implementations: add `tx.(pgx.Tx)` assertions
- `bootstrap.go`: remove pool params from room use cases, add txMgr to children UpdateChild
- Test files: update constructor calls where pool param is removed
- Adapter files: remove unused imports

**Deferred to Follow-Up Work:**
- P1 domain model deepening (Money value object, Child entity behavior, etc.) — covered in docs/solutions/IMPROVEMENTS.md, not this refactor
- P1 repository consolidation from 11 mini-repos to 1 aggregate repo — separate work
- Introducing a proper domain-native Tx interface with Go stdlib types — potential future improvement over the minimal `any` approach

**Out of scope:**
- Database schema or migration changes
- Frontend changes
- Behavioral changes or feature additions
- Test improvements beyond making tests compile

### Dependencies

None. All changed code is self-contained within the Go `api/` package tree.

---

## Planning Contract

### Key Technical Decisions

**KTD1 — Domain Tx type: `any` over typed interface**
- `type Tx = pgx.Tx` → `type Tx any`
- Rationale: A real typed Tx interface (`Exec`, `Query`, `QueryRow`) would still need pgx types in the method signatures (`pgconn.CommandTag`, `pgx.Rows`, `pgx.Row`), defeating the purpose. Using `any` is the minimal change that severs the import dependency. Infrastructure repos cast to `pgx.Tx` internally via `tx.(pgx.Tx)`.
- Alternative considered: Define a full domain Querier interface with stdlib-compatible return types. Rejected as significantly more invasive — requires wrapping pgx return types, affects all sqlc callsites, and provides no behavioral benefit for a refactor whose goal is dependency isolation.

**KTD2 — Pool removal in UpdateChild: wrap in transaction**
- Replace the `pool` field with existing `txMgr` by wrapping the update+audit path in `txMgr.ExecTx`.
- Rationale: The audit writer's `Write` method accepts any `dbExecQuerier` (a private interface requiring only `Exec`). By using the tx from `ExecTx`, we eliminate the pool without adding a new abstraction. This also makes the update+audit atomic (currently they run in separate connections).
- Alternative considered: Threading a querier interface through the repo to the use case. Rejected because it adds a new abstraction for one use case and the `txMgr.ExecTx` pattern is already established across the codebase.

**KTD3 — Pool removal in ArchiveRoom/ReactivateRoom: dead code only**
- The `pool` field is stored in the struct but never read in `Execute()`. Both use `audit.WriteSystemWithTx` (which takes `pgx.Tx`) inside the `ExecTx` callback. Remove the field, parameter, and import.
- Confirmed by tests: `NewArchiveRoom(repo, txMgr, nil, nil)` — both `nil` values include the pool.

### Implementation Constraints

- Run `go fmt ./... && go vet ./... && go build ./...` after each unit to catch stale imports and type errors
- The `type Tx any` change is cross-module — better to apply it in one pass (U1) rather than module-by-module, to avoid intermediate non-compilable states
- Room use case tests already pass `nil` for pool — removing the param produces a compile error that pins the fix locations

---

## Implementation Units

### U1. Replace `pgx.Tx` with `any` in domain repository interfaces

**Goal:** Remove `import "github.com/jackc/pgx/v5"` from all 8 domain repository files by replacing `type Tx = pgx.Tx` with `type Tx any`.

**Requirements:** R1

**Files:**
- `api/internal/modules/children/domain/repository.go` — change `type Tx = pgx.Tx` → `type Tx any`, remove pgx import
- `api/internal/modules/billing/domain/repository.go` — same
- `api/internal/modules/funding/domain/repository.go` — same
- `api/internal/modules/rooms/domain/repository.go` — same
- `api/internal/modules/sessiontypes/domain/repository.go` — same
- `api/internal/modules/sessiontemplates/domain/repository.go` — same
- `api/internal/modules/term/domain/repository.go` — same
- `api/internal/modules/payments/domain/repository.go` — same

**Approach:**
- In each file: `type Tx = pgx.Tx` → `type Tx any`, then delete the `"github.com/jackc/pgx/v5"` import line
- Verify no other code in the domain package uses pgx types
- After all 8 files, run `go build ./...` — will produce type-assertion errors in infrastructure repos (fixed in U2)

**Test scenarios:**
- Compilation: `go build ./api/internal/modules/children/domain/...` succeeds without pgx imports
- Compilation: same for all 7 other modules

**Verification:** `go build ./api/internal/modules/...` succeeds; grep for `import.*pgx` in domain/ directories returns nothing

---

### U2. Update infrastructure repository implementations

**Goal:** Fix type errors from U1 by adding `tx.(pgx.Tx)` assertions in postgres repository methods where the interface now passes `any` instead of `pgx.Tx`.

**Requirements:** R1, R4

**Files (representative — all postgres repos that implement domain interfaces with `tx` params):**
- `api/internal/modules/children/infrastructure/postgres/repository.go` — add `tx.(pgx.Tx)` at each method boundary
- `api/internal/modules/billing/infrastructure/postgres/repository.go` — same
- `api/internal/modules/funding/infrastructure/postgres/repository.go` — same
- `api/internal/modules/rooms/infrastructure/postgres/repository.go` — same
- `api/internal/modules/sessiontypes/infrastructure/postgres/repository.go` — same
- `api/internal/modules/sessiontemplates/infrastructure/postgres/repository.go` — same
- `api/internal/modules/term/infrastructure/postgres/term_repository.go` — same
- `api/internal/modules/term/infrastructure/postgres/term_schedule_change_repository.go` — same
- `api/internal/modules/payments/infrastructure/postgres/repository.go` — same

**Approach:**
- In each postgres repository method that receives `tx` as a parameter, add `pgxTx := tx.(pgx.Tx)` at the top (or cast inline where passed to sqlc)
- Most methods pass `tx` to `sqlc.New(tx)` — change to `sqlc.New(tx.(pgx.Tx))`
- For methods that pass tx to helper functions, do the assertion at the method boundary

**Pattern to follow:**
```go
// Before
func (r *Repo) Create(ctx context.Context, tx domain.Tx, ...) error {
    q := sqlc.New(tx)
}
// After
func (r *Repo) Create(ctx context.Context, tx domain.Tx, ...) error {
    q := sqlc.New(tx.(pgx.Tx))
}
```

**Test scenarios:**
- Compilation: `go build ./...` passes for entire `api/` tree
- No runtime type-assertion panics: the real callers always pass `pgx.Tx` (from `txMgr.ExecTx`), so the assertion always succeeds against the concrete type

**Verification:** `go test ./api/internal/modules/...` passes

---

### U3. Remove `*pgxpool.Pool` from rooms ArchiveRoom and ReactivateRoom

**Goal:** Remove dead `pool` field from `ArchiveRoom` and `ReactivateRoom` use cases.

**Requirements:** R2, R3

**Files:**
- `api/internal/modules/rooms/application/archive_room.go` — remove `pool  *pgxpool.Pool` field, remove from constructor signature, remove pgxpool import
- `api/internal/modules/rooms/application/reactivate_room.go` — same
- `api/internal/modules/rooms/application/application_test.go` — update `NewArchiveRoom` and `NewReactivateRoom` calls to drop the `nil` pool param, remove pgxpool import
- `api/internal/app/bootstrap/bootstrap.go` — remove `pool` arg from `NewArchiveRoom` and `NewReactivateRoom` calls

**Approach:**
- Remove `pool  *pgxpool.Pool` from struct, remove parameter from `NewArchiveRoom`/`NewReactivateRoom`
- Remove `"github.com/jackc/pgx/v5/pgxpool"` import — ensure no other code in those files uses it
- In bootstrap.go: `roomsapp.NewArchiveRoom(roomsRepo, txManager, auditWriter, pool)` → `roomsapp.NewArchiveRoom(roomsRepo, txManager, auditWriter)`
- In tests: `application.NewArchiveRoom(repo, txMgr, nil, nil)` → `application.NewArchiveRoom(repo, txMgr, nil)`

**Test scenarios:**
- Compilation: `go build ./api/internal/modules/rooms/...` succeeds
- Existing behavior: `TestArchiveRoom_Success`, `TestArchiveRoom_HasChildren`, `TestArchiveRoom_AlreadyArchived`, `TestReactivateRoom_Success`, `TestReactivateRoom_AlreadyActive` all pass
- No regression in bootstrap wiring: `TestBootstrapWithOptions` passes

**Verification:** `go test ./api/internal/modules/rooms/... ./api/internal/app/bootstrap/...` passes

---

### U4. Remove `*pgxpool.Pool` from children UpdateChild

**Goal:** Replace `pool` usage in `UpdateChild.Execute` with a transaction wrapper via `txMgr`.

**Requirements:** R2, R3

**Files:**
- `api/internal/modules/children/domain/repository.go` — add `UpdateWithTx(ctx, tx Tx, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)` to `ChildIdentityRepository` interface
- `api/internal/modules/children/infrastructure/postgres/repository.go` — implement `UpdateWithTx` (identical to `Update` but using tx instead of pool for sqlc queries)
- `api/internal/modules/children/application/update_child.go` — add `txMgr *transaction.Manager` field, remove `pool *pgxpool.Pool`, wrap `Execute` body in `txMgr.ExecTx`, use `repo.UpdateWithTx` and `audit.WriteWithTx` inside the callback
- `api/internal/app/bootstrap/bootstrap.go` — change `childapp.NewUpdateChild(childRepo, auditWriter, pool)` → `childapp.NewUpdateChild(childRepo, auditWriter, txManager)`
- `api/internal/modules/children/infrastructure/postgres/repository_test.go` — if tests construct `UpdateChild` with pool directly, update to use txMgr

**Approach:**

UpdateChild.Execute becomes:
```go
func (uc *UpdateChild) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params UpdateChildParams) (domain.Child, error) {
    // ... validation unchanged ...
    err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
        rowsAffected, err := uc.repo.UpdateWithTx(ctx, tx, actor.TenantID, actor.BranchID, id, fields)
        if err != nil { return err }
        if rowsAffected == 0 { return domainerrors.NotFound(...) }
        return uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
            ActionType: "child_updated", EntityType: "child", EntityID: id, Details: map[string]any{},
        })
    })
    if err != nil {
        return domain.Child{}, err
    }
    // Read fresh data AFTER transaction commits — the pool can see committed changes.
    updated, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
    if err != nil || !found { return domain.Child{}, domainerrors.Internal(...) }
    return updated, nil
}
```

Critical: `GetByID` must run **after** the transaction commits, not inside the `ExecTx` callback. A pool-based read inside the uncommitted transaction would return stale pre-update data.

**Test scenarios:**
- Compilation: `go build ./api/internal/modules/children/...` succeeds
- Update+audit happens in the same transaction: if update succeeds but audit write fails, the update is rolled back (improvement over current behavior)
- Existing behavior: all children module use case and handler tests pass
- No regression: `TestBootstrapWithOptions` passes

**Verification:** `go test ./api/internal/modules/children/... ./api/internal/app/bootstrap/...` passes

---

### U5. Clean up bootstrap.go and final verification

**Goal:** Ensure bootstrap.go compiles cleanly after all unit changes, verify no stale pgxpool imports remain in application packages.

**Requirements:** R2, R4

**Files:**
- `api/internal/app/bootstrap/bootstrap.go` — verify all constructor calls match updated signatures; remove any stale pgxpool imports no longer needed

**Approach:**
- After U3 and U4, verify: `NewArchiveRoom`, `NewReactivateRoom` no longer pass `pool`; `NewUpdateChild` passes `txManager` instead of `pool`
- Run `go fmt ./... && go vet ./... && go build ./...` from `api/` directory
- Use `rg 'pgxpool' api/internal/modules/ -g '*.go' | grep '/application/'` to confirm zero matches

**Test scenarios:**
- Full compilation: `go build ./...` in `api/` succeeds
- Lint clean: `go vet ./...` produces zero warnings
- No pool imports remain in any application package
- Verify all module tests pass

**Verification:** `go test ./api/...` passes

---

## Verification Contract

```sh
# After each unit — type-safety check
cd api && go build ./...

# Final — full verification
cd api && go fmt ./... && go vet ./... && go build ./... && go test ./...
```

Key grep assertions after completion:
- `grep -r 'import.*pgx' internal/modules/*/domain/` — zero matches
- `grep -r 'pgxpool' internal/modules/*/application/` — zero matches

---

## Definition of Done

- `go build ./...` succeeds from `api/`
- `go vet ./...` produces zero warnings from `api/`
- `go test ./api/...` all pass
- No `github.com/jackc/pgx/v5` import in any `domain/` package
- No `github.com/jackc/pgx/v5/pgxpool` import in any `application/` package
- All 8 `type Tx = pgx.Tx` replaced with `type Tx any`
- `*pgxpool.Pool` removed from all 3 use case constructors
