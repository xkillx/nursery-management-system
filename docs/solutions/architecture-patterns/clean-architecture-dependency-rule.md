---
title: "Clean Architecture Dependency Rule: Removing Infrastructure Types from Domain and Application Layers"
date: 2026-06-28
module: api
problem_type: architecture_pattern
component: tooling
severity: high
symptoms:
  - pgx.Tx type alias leaked into 8 domain repository interface packages
  - "*pgxpool.Pool appeared in 3 application use case constructors"
  - Domain layer imported github.com/jackc/pgx/v5 — violates inward-only dependency rule
resolution_type: code_fix
tags:
  - dependency-rule
  - clean-architecture
  - domain-layer
  - application-layer
  - refactor
  - type-alias
---

# Clean Architecture Dependency Rule: Removing Infrastructure Types from Domain and Application Layers

## Context

An architecture review flagged **P0 Clean Architecture dependency-rule violations**: infrastructure types (`pgx.Tx`, `*pgxpool.Pool`) leaked into domain and application layers across 12 domain packages and 3 use cases. The Go compiler didn't catch the violations because the project used `type Tx = pgx.Tx` (type alias), which makes `Tx` and `pgx.Tx` the same type — but the import of `github.com/jackc/pgx/v5` still sat in domain-layer packages, breaking the rule that domain must depend on nothing.

Three distinct patterns of violation were found:

1. **U1 (12 packages):** Domain repository interfaces defined `type Tx = pgx.Tx`, pulling the pgx dependency into domain packages that should be free of infrastructure concerns.
2. **U2 (3 use cases):** `*pgxpool.Pool` fields in `NewArchiveRoom`, `NewReactivateRoom`, and `NewUpdateChild` constructors — use cases should only depend on domain interfaces and pure types.
3. **U3 (5+ domain files):** Method signatures in domain repository interfaces used `pgx.Tx` directly (via the type alias), coupling domain contracts to a specific database driver.

## Guidance

### Strategy 1: Replace `type Tx = pgx.Tx` with `type Tx = any` (U1/U3)

Replace the infrastructure-aware type alias with a fully generic one:

```go
// Before — imports github.com/jackc/pgx/v5 into domain
type Tx = pgx.Tx

// After — zero infrastructure imports in domain
type Tx = any
```

**Why `= any` (alias) not `any` (defined type):** Multiple domain packages each define their own `Tx`. Adapter methods in `bootstrap/adapters.go` sometimes implement interfaces from **multiple** domain packages simultaneously. With `type Tx any` (defined type), each package's `Tx` is a distinct Go type — cross-package adapters fail interface satisfaction. With `type Tx = any` (alias), all `Tx` resolve to the same `any`, preserving cross-package compatibility.

After removing the alias, add type assertions in every infrastructure repository method that receives the `tx` parameter:

```go
// Before (infra):
func (r *Repo) Create(ctx context.Context, tx domain.Tx, ...) error {
    q := sqlc.New(tx)  // tx is pgx.Tx via alias
}

// After (infra):
func (r *Repo) Create(ctx context.Context, tx domain.Tx, ...) error {
    q := sqlc.New(tx.(pgx.Tx))  // explicit assertion at infrastructure boundary
}
```

The assertion is safe because infrastructure is the only layer that ever creates real transactions — domain never constructs `Tx`, only receives it.

### Strategy 2: Remove dead infrastructure fields (U2 — ArchiveRoom, ReactivateRoom)

If a use case stores a `*pgxpool.Pool` but never reads it in `Execute()`, it's dead code:

```go
// Before:
func NewArchiveRoom(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer, pool *pgxpool.Pool) *ArchiveRoom

// After:
func NewArchiveRoom(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ArchiveRoom
```

Confirm by checking tests — if they pass `nil` for the pool parameter, it's unused.

### Strategy 3: Replace pool with transaction wrapper (U2 — UpdateChild)

When a use case uses `pool` only to satisfy an audit writer's argument that expects a `pgx.Tx` or `pgxpool.Pool`, wrap the entire update+audit path in `txMgr.ExecTx`:

```go
// Before (separate connections, no atomicity):
func (uc *UpdateChild) Execute(...) (domain.Child, error) {
    rowsAffected, err := uc.repo.Update(ctx, tenantID, branchID, id, fields)
    // ... check rowsAffected ...
    uc.audit.Write(ctx, uc.pool, actor, params) // separate connection
    updated, _, _ = uc.repo.GetByID(ctx, tenantID, branchID, id)
    return updated, nil
}

// After (atomic transaction):
func (uc *UpdateChild) Execute(...) (domain.Child, error) {
    err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
        rowsAffected, err := uc.repo.UpdateWithTx(ctx, tx, tenantID, branchID, id, fields)
        if err != nil { return err }
        if rowsAffected == 0 { return domain.NewNotFoundError(...) }
        return uc.audit.Write(ctx, tx, actor, params) // same transaction
    })
    if err != nil { return domain.Child{}, err }
    // GetByID runs AFTER commit — pool read sees committed data
    updated, _, _ := uc.repo.GetByID(ctx, tenantID, branchID, id)
    return updated, nil
}
```

Add `UpdateWithTx` to the domain repository interface and implement it in the postgres adapter (identical to `Update` but accepts the `tx` parameter instead of using the pool for `sqlc.New`).

## Why This Matters

- **Testability:** Domain use cases can be unit-tested with mock transactions instead of a real database pool.
- **Module independence:** Domain packages no longer import `github.com/jackc/pgx/v5`, so swapping the database driver doesn't touch domain code.
- **Framework flexibility:** The same pattern allows migrating from pgx to `database/sql`, DynamoDB, or any other backend without changing application/domain interfaces.
- **Compiler enforcement:** Once infrastructure imports are removed from domain, new violations show up at `go vet`/`go build` time — the compiler now enforces the dependency rule.

## When to Apply

- When reviewing PRs that add `pgx`, `pgxpool`, or `pgx/v5` imports to `internal/domain/` or `internal/application/` packages.
- When adding a new domain repository interface — define `type Tx = any` in the package, not `type Tx = pgx.Tx`.
- When a use case constructor accepts an infrastructure type (`*pgxpool.Pool`, `*sql.DB`, `*gorm.DB`, etc.) — check whether it's actually used in `Execute()` (dead code removal) or only used to pass to an audit writer / side-effecting call (transaction wrapping).
- When creating adapter methods in `bootstrap/adapters.go` that span multiple domain packages — prefer `any` over concrete infrastructure types.

## Related

This doc covers the same dependency-rule enforcement approach documented in the broader architecture conventions at `docs/agents/ARCHITECTURE.md` (Cross-module imports section) and `CONCEPTS.md` (Transaction Management pattern). See also the Clean Architecture dependency-rule checklists in `docs/agents/PROJECT-CONTEXT.md` for ongoing audit guidance.
