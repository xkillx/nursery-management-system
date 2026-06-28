# Architecture Review: Go API Backend

**Date:** 2026-06-28
**Scope:** `api/internal/` — Go backend
**Stack:** Go 1.26, Gin, pgx, sqlc
**Pattern:** Clean Architecture + DDD + Hexagonal (opinionated synthesis)

---

## 1. Module Structure (Score: 7/10)

```
api/internal/
├── modules/
│   ├── {absence, attendance, authentication, billing, children,
│   │    funding, invites, invoicerun, owner, parentchildmappings,
│   │    passwordreset, payments, rooms, sessiontemplates,
│   │    sessiontypes, term}
│   │   ├── domain/          # entities, value objects, errors, repository interface
│   │   ├── application/     # use cases (Execute method pattern)
│   │   ├── infrastructure/
│   │   │   └── postgres/    # repository implementations (sqlc)
│   │   └── interfaces/
│   │       └── http/        # handlers, DTOs, route registration
│   └── app/
│       └── bootstrap/       # composition root, adapters
├── platform/                # shared cross-cutting: errors, audit, transaction,
│                            #   tenant context, metrics, http middleware, config
└── cmd/                     # entry points (server, seed)
```

**Consistency:** All 17 modules follow the same 4-layer structure — notable discipline at this scale.

---

## 2. Layer Adherence

### Dependency Rule

| Direction | Status | Evidence |
|-----------|--------|----------|
| Domain → (nothing) | ✅ Clean | No imports of http/sql/gin/framework |
| Application → Domain | ✅ Correct | Imports only domain types + platform interfaces |
| Infrastructure → Domain | ✅ Correct | Implements domain.Repository |
| Handler → Application | ✅ Correct | Injects use case structs, never repos |

### ⚠️ Violation found: `pgx.Tx` type alias in domain

`api/internal/modules/children/domain/repository.go:12`
```go
type Tx = pgx.Tx
```

This introduces a PostgreSQL type into the domain layer. The domain should define its own transaction interface.

### ⚠️ Violation found: `*pgxpool.Pool` in application use cases

Some use case constructors accept `*pgxpool.Pool` directly:
- `childapp.NewUpdateChild` receives `pool` (bootstrap.go:176)
- `roomsapp.NewArchiveRoom` / `NewReactivateRoom` receive `pool` (bootstrap.go:359-360)

The pool type is infrastructure; use cases should only depend on `*transaction.Manager`.

---

## 3. Domain Model (Score: 6.5/10)

### Strong areas
- **Billing domain** is the richest — value objects (`BillingPeriod`, `BillingMonth`, `PreflightResult`, `BookedCoreCalculation`), pure domain functions (`CalculateBookedCoreMinutesInMonth`, `CalculateAttendanceMinutes`, `ComputeFundedDeductionMinor`)
- **Attendance domain** has domain constants (`SessionStatus`, `EventType`) and clear entity boundaries
- **Domain errors** are well-structured: `DomainError` with code, message, field, details
- **Sentinel errors** are properly mapped to `DomainError` in the application layer

### Weak areas
- **Children domain is anemic** — `Child` struct has 24 fields and only 2 methods (`MissingRequirements`, `EnrollmentComplete`). Most logic lives in application services.
- **Money is not a value object** — `api/internal/modules/billing/domain/money.go` is a utility function (`CalculateHourlyAmountMinor`) operating on raw `int`. No type safety for currency amounts. Compare with how `BillingMonth` is wrapped but `Money` is not.
- **DTOs in domain layer** — `api/internal/modules/billing/domain/preflight.go` contains database row projections (`PreflightChildRow`, `PreflightAttendanceSessionRow`) that belong in infrastructure.
- **Repository per sub-resource** — The children module defines **11 separate repository interfaces** (`ChildIdentityRepository`, `ChildProfileRepository`, `ChildContactRepository`, `ChildHealthProfileRepository`, `ChildSafeguardingProfileRepository`, `ChildConsentRepository`, `ChildFundingRepository`, `ChildCollectionSettingsRepository`, `ChildRoomAssignmentsRepository`, `ChildBillingProfileRepository`, `ChildLeavingRepository`, `ChildBookingPatternsRepository`). This fragments the `Child` aggregate boundary — the aggregate root is `Child`, which should have one repository.

---

## 4. Application Layer (Score: 8/10)

### Good patterns
- **Use case per operation** — each use case is a single struct with one `Execute()` method
- **Transaction wrapping** — all mutations go through `txMgr.ExecTx()`
- **Error mapping** — sentinel errors → `DomainError` → HTTP status
- **Observability** — `WithObservability()` builder for optional logger/metrics injection
- **Cross-module adapters** in `bootstrap/adapters.go` are clean — e.g., `childEnrollmentCheckerAdapter` delegates from attendance domain interface to children infrastructure

### Concerns
- **Large use cases** — `GenerateDraftInvoicesUseCase` is 562 lines. Consider splitting into smaller steps or using a pipeline.
- **Boilerplate `WithObservability`** — Creates a full struct copy every time just to add a logger. Alternative: inject logger at construction time or use context-based logging.
- **`pool` injection into use cases** — Several use cases bypass `txManager` and accept `*pgxpool.Pool` for non-transactional queries. The pool type should stay in infrastructure; the repository should expose both tx and non-tx methods.

---

## 5. Ports & Adapters (Score: 7.5/10)

### Adapter wiring
- All cross-module adapters live in `bootstrap/adapters.go` — good centralization
- Compile-time interface satisfaction checks (`var _ domain.Repository = (*postgres.Repo)(nil)`) — excellent practice
- Adapter naming is inconsistent: some use `Adapter` suffix, some don't

### Port placement
- Repository interfaces are in `domain/` (DDD convention) ✓
- Application-layer ports (e.g., `SessionTypeLookup`, `ChildDeactivator`) are in `application/` — correct for Hexagonal
- Some ports would benefit from clearer naming

---

## 6. Domain Events (Score: 2/10)

No domain events are used anywhere. The project relies on:
- **Audit logging** (`audit.Writer`) for recording state changes
- **Direct use case calls** for cross-aggregate operations (e.g., `childDeactivatorAdapter` calls `MarkInactive.Execute()`)

This means cross-aggregate consistency is synchronous (same transaction), which works but limits the ability to:
- React to domain changes asynchronously
- Add new cross-cutting behaviors without modifying existing use cases
- Replay domain events for debugging or projection rebuild

---

## 7. Composition Root (Score: 8/10)

- `bootstrap.go` (471 lines) wires all dependencies — long but linear and readable
- `BootstrapOptions` allows test injection of external providers (Stripe, email)
- Wire order is sensible: logger → router → metrics → auth → repos → txmgr → audit → use cases → handlers → routes

**Room for growth:** DI is manual. At 17 modules and growing, a DI container (e.g., `uber/fx`, `google/wire`) could reduce the 471-line bootstrap.

---

## 8. Cross-Cutting Concerns

| Concern | Implementation | Score |
|---------|---------------|-------|
| Multi-tenancy | `tenant.ActorContext` propagated through all layers | 9/10 |
| Authentication | JWT + Bearer token, `AuthnMiddleware` | 8/10 |
| Authorization | Role middleware (`RequireRoles`) | 7/10 |
| Observability | Structured `slog` + Prometheus via `WithObservability()` | 8/10 |
| Error handling | `DomainError` → `MapDomainError()` → HTTP JSON | 8/10 |
| Testing | Integration tests in bootstrap, repo tests, use case tests | 6/10 |
| Configuration | Env-based `Config` struct with validation | 8/10 |
| Rate limiting | Fixed window per-module | 6/10 |
| Audit logging | `audit.Writer` with per-use-case wiring | 7/10 |

---

## 9. Summary Scores

| Dimension | Score | Key Issue |
|-----------|-------|-----------|
| Layer separation | 8/10 | pgx.Tx in domain |
| Dependency rule | 7/10 | Pool leaking into use cases |
| Domain model richness | 6.5/10 | Anemic children domain |
| Aggregate boundaries | 6/10 | 11 mini-repos for Child |
| Value objects | 5/10 | Raw int for money |
| Domain events | 2/10 | None; audit-only |
| Port/Adapter clarity | 7.5/10 | Good but uneven naming |
| Composition root | 8/10 | Manual, scaling limits |
| **Overall** | **6.5/10** | |

---

## 10. Strengths (TL;DR)

- Strict 4-layer module structure with no layer-skipping
- Use case per operation with clean `Execute()` signature
- Adapter pattern prevents direct cross-module imports
- Transaction manager enforces atomic use cases
- DomainError → HTTP error mapping is clean and consistent
- Multi-tenancy through ActorContext is pervasive and correct
- sqlc-generated type-safe queries keep infra code lean
- Compile-time interface satisfaction checks confirm wiring correctness

---

## 11. Anti-Patterns Present

| Anti-Pattern | Location | Severity |
|-------------|----------|----------|
| Infrastructure leaking into domain | `domain/repository.go:12` (`type Tx = pgx.Tx`) | Medium |
| Leaking pool into use cases | Various `*pgxpool.Pool` params in use case constructors | Medium |
| Repository per sub-entity | 11 repos for `Child` aggregate | Medium |
| DTOs in domain | `billing/domain/preflight.go` DB row types | Medium |
| Missing Money value object | `billing/domain/money.go` raw int | Low |
| Large use case | `billing/application/generate_draft_invoices.go` (562 lines) | Low |
| God handler constructor | `children/interfaces/http/handler.go` (30+ params) | Low |
| No domain events | Entire codebase | Medium |
