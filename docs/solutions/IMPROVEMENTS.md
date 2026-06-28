# Architecture Improvement Suggestions

Prioritized by impact vs effort. Each suggestion includes a code-level change plan.

---

## P0 — Dependency Rule Violations (Fix Now)

### 1. Remove `pgx.Tx` type alias from domain

**Problem:** `children/domain/repository.go:12` defines `type Tx = pgx.Tx`, pulling a PostgreSQL type into the domain layer.

**Fix:**
```go
// domain/repository.go — define a domain-native transaction interface
type Tx interface {
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Or simpler — use `context.Context` + `transaction.Querier` from platform:

```go
// domain/repository.go
type Tx = any  // or define minimal interface

// infrastructure/postgres/repository.go — implement with pgx.Tx
```

**Files:**
- `api/internal/modules/children/domain/repository.go`
- `api/internal/modules/children/domain/child_room_assignment.go`
- `api/internal/modules/children/domain/child_funding_record.go`
- All repository interfaces across all modules that reference `pgx.Tx`

### 2. Remove `*pgxpool.Pool` from use case constructors

**Problem:** Several use cases accept `*pgxpool.Pool` instead of going through the transaction manager or repository.

**Files with violations:**
- `bootstrap.go:176` — `childapp.NewUpdateChild(childRepo, auditWriter, pool)`
- `bootstrap.go:359-360` — `roomsapp.NewArchiveRoom(..., pool)`, `NewReactivateRoom(..., pool)`
- `bootstrap.go:403` — `childDeactivatorAdapter` (indirect via `MarkInactive`)

**Fix:** Push direct queries into the repository layer. If a use case needs a non-transactional query, the repository should expose both transactional and non-transactional method variants (the repo knows whether it needs a tx or pool internally).

---

## P1 — Domain Model Deepening (Medium Effort, High Value)

### 3. Make `Money` a value object

**Problem:** All monetary values are `int` (minor units). No type-safety prevents unit confusion.

```go
// api/internal/modules/billing/domain/money.go
type Money struct {
    Minor int // pence / cents
}

func GBP(minor int) Money { return Money{Minor: minor} }
func (m Money) Add(other Money) Money { return Money{Minor: m.Minor + other.Minor} }
func (m Money) Multiply(factor int) Money { return Money{Minor: m.Minor * factor} }
```

Propagate to all domain types: `SubtotalMinor`, `TotalDueMinor`, `FundedDeductionMinor` → `Money`.

### 4. Collapse Child sub-repositories into one aggregate repository

**Problem:** 11 separate repository interfaces for the `Child` aggregate. This fragments transactions and makes it unclear what the aggregate boundary is.

**Fix:** Keep one `ChildRepository` interface with all methods. The postgres implementation already does this on one struct — the interface split adds no value and encourages callers to bypass the aggregate root.

```go
// domain/repository.go — single aggregate repository
type ChildRepository interface {
    // identity
    GetByID(...) (Child, bool, error)
    Create(...) error
    Update(...) error
    MarkInactive(...) error

    // profile (always loaded with child)
    GetProfile(...) (*ChildProfile, error)
    UpsertProfile(...) (*ChildProfile, error)

    // contacts
    ReplaceContacts(...) error

    // ... rest — all on one interface
}
```

### 5. Move DTOs out of domain layer

**Problem:** `billing/domain/preflight.go` contains `PreflightChildRow`, `PreflightAttendanceSessionRow` — these are database row projections, not domain concepts.

**Fix:** Move to `billing/infrastructure/postgres/dto.go` or keep them as private to the repository file. Use domain types in the repository interface instead.

### 6. Add domain behavior to `Child` entity

**Problem:** `Child` is anemic — 24 fields, 2 methods.

**Fix:** Move validation and business rules from application services into the entity:
```go
// domain/child.go
func (c *Child) Activate(startDate time.Time, hourlyRateMinor int) error
func (c *Child) Deactivate(reasonCode ReasonCode, deactivatedAt time.Time) error
func (c *Child) ChangeName(firstName string, lastName *string) error
func (c *Child) IsEligibleForAttendance(localDate time.Time) bool
```

---

## P2 — Architecture Patterns (Medium Effort)

### 7. Introduce domain events

**Problem:** No domain events. Cross-aggregate operations are synchronous (e.g., `childDeactivatorAdapter` calls `MarkInactive.Execute()` directly). Adding new side effects requires modifying existing use cases.

**Suggested implementation (lightweight, no event bus):**
```go
// domain/shared/events.go
type DomainEvent interface{ OccurredAt() time.Time }

// domain/child/events.go
type ChildDeactivated struct {
    ChildID    uuid.UUID
    ReasonCode string
    Occurred   time.Time
}

// application/child/event_handlers.go
type ChildDeactivatedHandler interface {
    Handle(ctx context.Context, event ChildDeactivated) error
}
```

Start with the billing domain — `InvoiceIssued`, `InvoiceMarkedOverdue` — and the term module — `TermExpired`, `ChildDeactivated`. Use the transaction manager to ensure handlers run within the same transaction (transactional event dispatcher).

### 8. Migrate `WithObservability` to constructor injection

**Problem:** The `WithObservability()` pattern creates a full struct copy just to add a logger/recorder. It's easy to forget (silent no-op), and creates branching codepaths.

**Fix:** Inject logger and recorder at construction time. Use `slog.Default()` and `metrics.NopRecorder` as defaults for tests:
```go
func NewCheckInChild(
    repo domain.Repository,
    childChecker domain.ChildEnrollmentChecker,
    txMgr *transaction.Manager,
    audit *audit.Writer,
    clock *AttendanceClock,
    logger *slog.Logger,      // required, not optional
    recorder *metrics.Recorder, // nil-safe via no-op defaults
) *CheckInChild
```

### 9. Simplify handler constructors

**Problem:** Children handler constructor takes 30+ parameters. Adding a new sub-resource requires changing the constructor signature and all call sites.

**Option A:** Group use cases into a bag struct:
```go
type ChildrenUseCases struct {
    ListChildren        *ListChildren
    GetChild            *GetChild
    CreateChild         *CreateChildWithFullProfile
    // ...
}
```

**Option B:** Register sub-resource handlers independently (one route registration call per sub-area).

---

## P3 — Code Quality (Low Effort)

### 10. Standardize adapter naming

Some adapters in `bootstrap/adapters.go` follow `XxxAdapter` pattern, others don't. Pick one:
- `membershipCheckerAdapter` ✓
- `childScopeCheckerAdapter` ✓
- `childEnrollmentCheckerAdapter` ✓
- `sessionTypeLookupAdapter` ✓
- `siteRateProviderAdapter` ✓
- `childDeactivatorAdapter` ✓
- `bookingPatternLookupAdapter` ✓

All adhere, but `ownerInviteTokenAdapter` references `invitetokens.Manager` while `ownerEmailSenderAdapter` references `email.Sender` — the underlying "real" type is exposed in the name. Rename to `inviteTokenGeneratorAdapter` / `emailSenderAdapter` for clarity.

### 11. Ensure consistent error naming

Two patterns coexist:
- **Sentinel errors:** `var ErrSessionAlreadyOpen = errors.New("...")` (attendance, funding)
- **DomainError constructors:** `var ErrChildNotFound = func() error { return domainerrors.NotFound(...) }` (children)

Pick one per module. DomainError wrappers are preferred — they carry error codes for the HTTP mapper without string parsing.

### 12. Break up large use case

`billing/application/generate_draft_invoices.go` (562 lines) handles:
- Preflight for each term → invoice generation → run completion → metrics

Extract into a pipeline:
```go
type DraftInvoicePipeline struct {
    preflight    *PreflightTerm
    generator    *GenerateTermInvoices
    finaliser    *CompleteInvoiceRun
    metrics      *InvoiceMetrics
}
```

---

## P4 — Strategic (Plan for Later)

### 13. Evaluate DI container

At 17 modules and growing, manual wiring in `bootstrap.go` (471 lines) is becoming unwieldy. Options:
- **`uber/fx`** — constructor-based DI with lifecycle hooks, minimal boilerplate
- **`google/wire`** — compile-time code generation, no runtime reflection

### 14. Split billing module by process

The billing module covers preflight, draft generation, issue, parent view, and overdue marking — too broad. Consider splitting into:
- `billing-generation` (draft generation)
- `billing-lifecycle` (issue, overdue, parent view)
- `billing-preflight` (separate if logic diverges significantly)

### 15. CQRS for billing reads

The billing module has read endpoints (`ListInvoicesForManagerReview`, `GetInvoiceForParent`) that query complex joins. As the system grows, consider:
- Separate read models (materialized views or denormalized tables)
- Read-only repository interface (`BillingReadRepository`)
- Write commands still go through use cases + events

Don't do this until billing read performance becomes measurable problem.

---

## Priority Matrix

| # | Suggestion | Impact | Effort | Priority |
|---|-----------|--------|--------|----------|
| 1 | Remove `pgx.Tx` from domain | High | Medium | **P0** |
| 2 | Remove `pool` from use cases | High | Low | **P0** |
| 3 | Money value object | Medium | Low | **P1** |
| 4 | Collapse Child repos | Medium | Medium | **P1** |
| 5 | Move DTOs from domain | Medium | Low | **P1** |
| 6 | Deepen Child entity | High | Medium | **P1** |
| 7 | Domain events | High | High | **P2** |
| 8 | Constructor-inject observability | Low | Medium | **P2** |
| 9 | Simplify handler constructors | Medium | Medium | **P2** |
| 10 | Standardize adapter naming | Low | Low | **P3** |
| 11 | Error naming consistency | Low | Low | **P3** |
| 12 | Break up large use case | Low | Medium | **P3** |
| 13 | DI container | Medium | High | **P4** |
| 14 | Split billing module | Medium | High | **P4** |
| 15 | CQRS for billing reads | Low | High | **P4** |
