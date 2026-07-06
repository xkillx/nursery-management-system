# Architecture Gap Report: Clean Architecture + DDD + Hexagonal

> Analyzed: 2026-07-06
> Codebase: Nursery Management System (Go 1.24 + Angular 20 + PostgreSQL)
> Scope: `api/internal/` (Go backend)

---

## Executive Summary

The codebase follows a **well-structured modular architecture** with clear separation into `domain/`, `application/`, `infrastructure/`, and `interfaces/http/` per module. The dependency rule is *mostly* respected, cross-module imports are avoided via a composition root (`bootstrap/`), and the transaction/event patterns are solid. However, several DDD and Clean Architecture patterns are missing or incomplete.

**Overall Maturity: 6.5/10** — Strong structural foundation, weak on domain richness.

---

## What's Working Well

| Aspect | Evidence |
|--------|----------|
| **Module isolation** | 20 modules, zero cross-module imports outside `bootstrap/` |
| **Composition root** | `bootstrap/wire.go` + `adapters.go` wires all dependencies |
| **Use case pattern** | Each operation is a dedicated struct with `Execute()` method |
| **Transaction management** | `transaction.Manager.ExecTx()` used consistently |
| **Domain events** | `events.Dispatcher` with typed handlers, in-tx dispatch |
| **Audit trail** | `audit.Writer` integrated into use cases |
| **Error mapping** | `platform/errors.DomainError` → HTTP via `MapDomainError()` |
| **Auth extraction** | `tenant.ActorFromGinContext()` — handlers don't parse JWT |
| **Adapter pattern** | Cross-module ports defined in consumer, adapters in `bootstrap/` |
| **Domain behavior** | `Child.Activate()`, `Child.Deactivate()`, `Term.NewTerm()`, `ValidateCorrectionReason()` |

---

## Gaps

### 1. Domain Layer — Anemic Models (HIGH)

**Problem:** Most domain "entities" are data bags with no behavior. Business logic lives in application services instead of entities.

| Module | Entity | Has Behavior? | Evidence |
|--------|--------|---------------|----------|
| `billing` | `Invoice` | No | `invoice.go` is 370 lines of structs, constants, and DTOs — zero methods |
| `attendance` | `Session` | No | Pure data struct; all logic in `application/check_in_child.go` |
| `funding` | `FundingProfile` | No | Data struct only |
| `rooms` | `Room` | No | Data struct only |
| `sessiontypes` | `SessionType` | No | Data struct only |
| `children` | `Child` | **Yes** | `Activate()`, `Deactivate()`, `ChangeName()`, `IsEligibleForAttendance()` |
| `term` | `Term` | **Yes** | `NewTerm()`, `DeriveStatus()`, `ShouldBePendingRenewal()` |
| `billing` | `Money` | **Yes** | Value object with `Add()`, `Multiply()`, `String()` |

**Fix:** Move validation, state transitions, and business rules into entities. Example: `Invoice.Issue()` should enforce draft-only issuance, not the use case.

---

### 2. Domain Layer — Missing Value Objects (MEDIUM)

**Problem:** Primitive obsession. IDs, names, rates, and addresses are raw `string`/`int`/`uuid.UUID` instead of value objects.

| What Should Be a VO | Current Type | Why It Matters |
|---------------------|--------------|----------------|
| `EmailAddress` | `string` | No validation at domain boundary |
| `HourlyRate` | `int` (minor units) | `Money` exists but isn't used for rates |
| `ChildName` | `string` fields | No trim/validate invariants |
| `Postcode` | `string` | No format validation |
| `ReasonCode` | `string` constants | Exists as constants but not a VO with validation |
| `BookingPattern` | Entity with no behavior | Could enforce "entries must be within term dates" |

**Exception:** `Money` value object in `billing/domain/` is well-implemented (immutable, structural equality, JSON marshaling). Use it as the template.

---

### 3. Domain Layer — Repository Interfaces Leak Infrastructure (MEDIUM)

**Problem:** `billing/domain/repository.go:10` imports `github.com/jackc/pgx/v5` for the `SiteRateRepository` interface parameter `tx pgx.Tx`.

```go
// billing/domain/repository.go:29
UpdateCoreHourlyRate(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, rateMinor int) error
```

**Workaround used elsewhere:** `type Tx = any` (16 modules) — a type alias that erases the concrete pgx type while preserving compile-time safety at the adapter level.

**Fix:** Change `SiteRateRepository.UpdateCoreHourlyRate` to use `tx Tx` (the `any` alias) instead of `tx pgx.Tx`, and remove the `pgx` import from `billing/domain/`.

---

### 4. Domain Layer — No Domain Services (MEDIUM)

**Problem:** Zero domain services found across 20 modules. Stateless domain logic that doesn't belong in an entity is placed directly in application services.

**Examples of logic that should be domain services:**
- Billing calculation (`billing/application/compute_invoice_prefill.go`) — hourly rate × minutes, funding deductions
- Attendance overlap detection — pure domain rule, currently in use case
- Term date derivation (`TermEndDateFor`, `ValidateTermStartDate`) — already in domain but could be a service if more complex

**Fix:** Extract stateless business rules into `domain/services/` when they span multiple entities or don't fit a single entity.

---

### 5. Domain Layer — Weak Aggregate Boundaries (MEDIUM)

**Problem:** The `Child` repository (`children/domain/repository.go`) manages 12+ sub-resources (profile, contacts, health, safeguarding, consent, funding, collection settings, room assignments, billing profile, leaving records, booking patterns) through a single mega-repository. This suggests either:

1. `Child` is a god aggregate (too many child entities), or
2. Some sub-resources should be separate aggregates with their own repositories

**DDD Rule:** One repository per aggregate. If sub-resources are always accessed via `Child`, they're part of the aggregate. If they're accessed independently (e.g., `ListContactsByChild`), they may warrant separate aggregates.

**Fix:** Evaluate whether contacts, health profiles, and booking patterns are truly part of the `Child` aggregate or should be separate bounded contexts.

---

### 6. Application Layer — Missing Command/Query Separation (LOW)

**Problem:** Use cases use ad-hoc input structs instead of formal Commands/Queries. No CQRS distinction.

**Current pattern:**
```go
type CreateDraftInvoiceInput struct { ... }
func (uc *CreateDraftInvoice) Execute(ctx, actor, input) (result, error)
```

**This is acceptable** for the current complexity level. CQRS is only needed when read and write models diverge significantly. The billing module is approaching that threshold (preflight reads vs. generation writes).

---

### 7. Application Layer — `Tx = any` Pattern (LOW)

**Problem:** 16 modules define `type Tx = any` in their domain repositories to avoid importing `pgx`. This erases type safety — any value can be passed as `tx`.

**Current mitigation:** Compile-time interface checks in `bootstrap/adapters.go` (e.g., `var _ childdomain.Repository = (*postgreschild.ChildRepository)(nil)`).

**Alternative considered:** A shared `domain/ports.Tx` interface in `platform/`. Rejected because it would create a domain dependency on platform. The `any` alias with compile-time adapter checks is a pragmatic compromise.

---

### 8. Infrastructure — Handler Imports Infrastructure (MEDIUM)

**Problem:** `invites/interfaces/http/handler.go:14` imports `invites/infrastructure/tokens` directly, bypassing the application layer.

```go
import "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"

type Handler struct {
    tokenMgr *tokens.Manager  // ← infrastructure leak
}
```

**Fix:** Define a port interface in `invites/application/` (e.g., `TokenValidator`) and inject the `tokens.Manager` through the composition root.

---

### 9. Infrastructure — Platform Dependencies in Domain (LOW)

**Problem:** Domain entities import `platform/errors` for error definitions.

```go
// term/domain/term.go:8
import domainerrors "nursery-management-system/api/internal/platform/errors"
```

**Impact:** Domain depends on a platform package. If `platform/errors` ever changed, domain code would need updating.

**Fix:** Define domain-specific error types within each module's `domain/` package, or accept this as a pragmatic trade-off since `platform/errors` is a stable, framework-agnostic utility.

---

### 10. Testing — Inconsistent Coverage (MEDIUM)

| Module | Domain Tests | Application Tests | Infrastructure Tests | Handler Tests |
|--------|-------------|-------------------|---------------------|---------------|
| `billing` | 5 | 2 | 1 | 0 |
| `children` | 1 | 4 | 3 | 0 |
| `attendance` | 1 | 5 | 1 | 0 |
| `authentication` | 0 | 6 | 1 | 1 |
| `term` | 1 | 1 | 1 | 0 |
| `rooms` | 1 | 1 | 0 | 0 |
| `funding` | 0 | 2 | 1 | 0 |
| `absence` | 0 | 1 | 1 | 0 |
| `payments` | 0 | 3 | 1 | 1 |

**Gaps:**
- No handler-level tests for most modules (only `authentication`, `payments`, `passwordreset`)
- Domain logic tests are sparse given the anemic models
- No architecture tests (e.g., "domain must not import infrastructure")

---

### 11. Events — Defined but Underused (LOW)

**Problem:** Domain events exist (`InvoiceIssued`, `InvoiceMarkedOverdue`, `ChildDeactivated`) but the event dispatcher is only used in `children/mark_inactive.go`. Other state transitions (invoice issued, attendance corrected) use audit logging instead of events.

**Fix:** Use events for cross-module side effects (e.g., "when invoice is issued, notify parent"). Audit and events serve different purposes — audit is for compliance, events for reactive workflows.

---

### 12. `invoicerun` Module — Structural Anomaly (LOW)

**Problem:** `invoicerun/` doesn't follow the standard `domain/`, `application/`, `infrastructure/`, `interfaces/` structure. It's a flat package with `runners.go` and `scheduler.go` that imports `term/application` and `platform/db/sqlc` directly.

**Fix:** Either formalize it as a cross-cutting orchestrator (document why it's different) or refactor into the standard module layout.

---

## Priority Matrix

| # | Gap | Severity | Effort | Priority |
|---|-----|----------|--------|----------|
| 1 | Anemic domain models | High | High | P1 — start with `billing` and `attendance` |
| 3 | pgx import in billing domain | Medium | Low | P1 — quick fix |
| 8 | Handler imports infrastructure | Medium | Low | P1 — quick fix |
| 2 | Missing value objects | Medium | Medium | P2 — incremental |
| 4 | No domain services | Medium | Medium | P2 — extract as complexity grows |
| 5 | Weak aggregate boundaries | Medium | High | P2 — requires domain analysis |
| 10 | Inconsistent test coverage | Medium | Medium | P2 — add handler tests first |
| 11 | Events underused | Low | Medium | P3 — when reactive workflows needed |
| 6 | No CQRS | Low | N/A | P3 — only if read/write models diverge |
| 7 | `Tx = any` pattern | Low | Low | P3 — pragmatic compromise |
| 9 | Platform deps in domain | Low | Low | P3 — stable utility |
| 12 | `invoicerun` structure | Low | Low | P3 — document or refactor |

---

## Recommended Next Steps

1. **Quick wins (P1):** Fix `billing/domain/repository.go` pgx import; move `invites` handler's `tokens.Manager` behind a port interface
2. **Domain enrichment (P1):** Add behavior to `Invoice` entity (`.Issue()`, `.MarkOverdue()`, `.AddLine()`) — the billing module has the most complex domain logic and will benefit most
3. **Value objects (P2):** Start with `EmailAddress`, `HourlyRate`, `ReasonCode` — small, high-value refactorings
4. **Architecture tests (P2):** Add a test that asserts domain packages never import infrastructure
5. **Domain expert review (P2):** Run an Event Storming session to validate aggregate boundaries, especially around `Child`

---

## References

- [Architecture & Conventions](docs/agents/ARCHITECTURE.md)
- [Design System](DESIGN.md)
- [Project Context](docs/agents/PROJECT-CONTEXT.md)
