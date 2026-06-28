---
title: Deepen Domain Model in Billing and Children Modules
date: 2026-06-28
category: architecture-patterns
module: billing, children
problem_type: architecture_pattern
component: documentation
severity: medium
applies_when:
  - "Domain structs are anemic (many fields, few methods)"
  - "Monetary or other primitive values appear frequently with the same validation rules"
  - "Repository interfaces are fragmented into per-concept sub-interfaces"
  - "Entity validation logic is scattered across application services"
  - "DB types leak into the domain layer"
tags:
  - value-object
  - money
  - entity-behavior
  - repository-pattern
  - clean-architecture
---

# Deepen Domain Model in Billing and Children Modules

## Context

Two Go modules in a Clean Architecture nursery management system — billing and children — had anemic domain models. Monetary values throughout the billing module were raw `int` (pence) with no type safety, making it impossible to distinguish a price from a quantity at the type level without reading field names. DB row DTOs (`PreflightChildRow`, `PreflightAttendanceSessionRow`) were defined in the billing domain package, leaking infrastructure concerns upward. The `Child` entity carried 24 fields but only 2 methods; validation for name changes, deactivation dates, and attendance eligibility lived in application services instead of the entity. The children module's repository was fragmented into 11 embedded sub-interfaces, making the interface surface area large and hard to navigate. This friction surfaced during code review as repeated "where does this validation belong?" conversations and during onboarding as confusion over what a `Child` could do vs. what services did to it.

## Guidance

### 1. Introduce Value Objects for Primitive Types

Wrap domain primitives (money, quantities, identifiers) in typed value objects with unexported fields. Construct via named constructors that validate invariants and return errors. Arithmetic operations return new instances — values are immutable.

```go
type Money struct {
    minor int
}

func GBP(minor int) (Money, error) {
    if minor < 0 {
        return Money{}, fmt.Errorf("GBP amount must not be negative")
    }
    return Money{minor: minor}, nil
}

func (m Money) Add(other Money) Money {
    return Money{minor: m.minor + other.minor}
}
```

Include `MarshalJSON`/`UnmarshalJSON` for backward-compatible wire format so the value object is invisible to API consumers:

```go
func (m Money) MarshalJSON() ([]byte, error) {
    return fmt.Appendf(nil, "%d", m.minor), nil
}
```

### 2. Keep DB Row Types Out of Domain

Database row DTOs belong in the infrastructure layer, not the domain package. If repository interface methods return concrete DTOs, either remove dead-code methods or convert to domain-type returns at the boundary. In the billing module, `PreflightChildRow` and `PreflightAttendanceSessionRow` moved from `billing/domain/preflight.go` to `billing/infrastructure/postgres/dto.go`, and five dead methods were removed from the `BillingRepository` interface.

### 3. Propagate Type Changes Layer by Layer

Introduce a value object in the domain first, then propagate outward:

1. **Domain** — define the type and update domain structs
2. **Infrastructure** — update repository mappings and DTO conversions
3. **Application** — update use case structs and logic
4. **HTTP** — update handler DTOs (JSON tags preserved for backward compatibility)

```go
// Before — raw int
type DraftInvoiceCreateParams struct {
    SubtotalMinor        int
    FundedDeductionMinor int
    TotalDueMinor        int
}

// After — typed Money
type DraftInvoiceCreateParams struct {
    Subtotal        Money
    FundedDeduction Money
    TotalDue        Money
}
```

Because `Money` implements `MarshalJSON` to emit the same integer wire format, no API contract breaks.

### 4. Merge Fragmented Repository Interfaces

Collapse embedded sub-interfaces into one flat `Repository` interface, grouping methods by concern area with comments rather than by interface embedding. This pattern already existed in the billing module and served as a model. Pure interface merge — no implementation changes needed.

```go
// Before: 11 embedded sub-interfaces
type Repository interface {
    ChildIdentityRepository
    ChildProfileRepository
    // ... 9 more
}

// After: one flat interface with concern-grouped methods
type Repository interface {
    // Identity
    List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int) ([]Child, error)
    // Profile
    GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
    // ... remaining concern groups inline
}
```

### 5. Add Behavioral Methods to Entities

Extract validation from application services into entity methods. An entity should enforce its own invariants. Methods should validate inputs, detect no-ops, and return clear errors.

```go
func (c *Child) Activate(startDate time.Time, hourlyRateMinor int) error {
    if c.IsActive {
        return fmt.Errorf("child is already active")
    }
    if startDate.Before(time.Now().Truncate(24 * time.Hour)) {
        return fmt.Errorf("start date must not be in the past")
    }
    c.StartDate = startDate
    c.SiteCoreHourlyRateMinor = &hourlyRateMinor
    c.IsActive = true
    c.EndDate = nil
    return nil
}

func (c Child) IsEligibleForAttendance(localDate time.Time) bool {
    if !c.IsActive {
        return false
    }
    if localDate.Before(c.StartDate) {
        return false
    }
    if c.EndDate != nil && !localDate.Before(*c.EndDate) {
        return false
    }
    return true
}
```

## Why This Matters

- **Type safety at domain boundaries.** A `Money` parameter cannot be confused with a quantity or an ID — the compiler enforces the distinction.
- **Validation consistency.** Extraction of date checks and reason-code validation into entity methods eliminates scattered copies of the same logic across services.
- **Discoverable behavior.** New team members see what a `Child` can do by reading the entity's methods, not by hunting through 11 application services.
- **Simplified interfaces.** The flat `Repository` reduces cognitive load: one file to scan, not 11 embedded sub-interfaces.
- **Easier testing.** Entity methods have no external dependencies and are trivially unit-testable without mocks.
- **Backward-compatible wire format.** JSON marshal methods on value objects mean API consumers are unaffected by the domain change.

## When to Apply

- Domain structs carry many fields but few or no methods
- The same primitive type (money, percentage, rate) appears across multiple structs with repeated validation
- Repository interfaces use embedding for per-concept sub-interfaces
- Entity invariants (activation rules, name requirements, date ranges) are enforced in application-layer code
- DB-specific types or concrete DTOs appear in domain package files
- Code reviewers frequently ask "why is this validation here?"

## Examples

### Money value object

```go
// Define in billing/domain/money.go
type Money struct {
    minor int
}

func GBP(minor int) (Money, error) {
    if minor < 0 {
        return Money{}, fmt.Errorf("GBP amount must not be negative")
    }
    return Money{minor: minor}, nil
}

func MustGBP(minor int) Money {
    m, err := GBP(minor)
    if err != nil {
        panic(err)
    }
    return m
}

func (m Money) Add(other Money) Money {
    return Money{minor: m.minor + other.minor}
}

func (m Money) Multiply(factor int) Money {
    return Money{minor: m.minor * factor}
}

func (m Money) String() string {
    return fmt.Sprintf("GBP %d.%02d", m.minor/100, m.minor%100)
}

func (m Money) MarshalJSON() ([]byte, error) {
    return fmt.Appendf(nil, "%d", m.minor), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
    var v int
    _, err := fmt.Sscanf(string(data), "%d", &v)
    if err != nil {
        return fmt.Errorf("invalid money value: %w", err)
    }
    m.minor = v
    return nil
}
```

### Entity behavior for Child

```go
func (c *Child) Activate(startDate time.Time, hourlyRateMinor int) error {
    if c.IsActive {
        return fmt.Errorf("child is already active")
    }
    now := time.Now().UTC()
    if startDate.Before(now.Truncate(24 * time.Hour)) {
        return fmt.Errorf("start date must not be in the past")
    }
    c.StartDate = startDate
    c.SiteCoreHourlyRateMinor = &hourlyRateMinor
    c.IsActive = true
    c.EndDate = nil
    return nil
}

func (c *Child) Deactivate(reasonCode ReasonCode, deactivatedAt time.Time) error {
    if !c.IsActive {
        return fmt.Errorf("child is already inactive")
    }
    if _, ok := ValidReasonCodes[reasonCode]; !ok {
        return fmt.Errorf("invalid reason code: %s", reasonCode)
    }
    c.EndDate = &deactivatedAt
    c.IsActive = false
    return nil
}

func (c *Child) ChangeName(firstName string, lastName *string) error {
    if firstName == "" {
        return fmt.Errorf("first name must not be empty")
    }
    if c.FirstName == firstName && stringPtrEqual(c.LastName, lastName) {
        return fmt.Errorf("no change in name")
    }
    c.FirstName = firstName
    c.LastName = lastName
    return nil
}

func (c Child) IsEligibleForAttendance(localDate time.Time) bool {
    if !c.IsActive {
        return false
    }
    if localDate.Before(c.StartDate) {
        return false
    }
    if c.EndDate != nil && !localDate.Before(*c.EndDate) {
        return false
    }
    return true
}
```

### Before/after: repository interface flattening

```go
// Before — 11 embedded sub-interfaces
type Repository interface {
    ChildIdentityRepository
    ChildProfileRepository
    ChildContactRepository
    ChildHealthProfileRepository
    ChildSafeguardingProfileRepository
    ChildConsentRepository
    ChildFundingRepository
    ChildCollectionSettingsRepository
    ChildRoomAssignmentsRepository
    ChildBillingProfileRepository
    ChildLeavingRepository
    ChildBookingPatternsRepository
}

// After — one flat interface with group headers
type Repository interface {
    // Identity
    List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int) ([]Child, error)
    // Profile
    GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
    // ... remaining concern groups inline
}
```

## Related

- [Implementation plan](docs/plans/2026-06-28-002-refactor-domain-model-deepening-plan.md)
- Money value object test patterns: `api/internal/modules/billing/domain/money_test.go`
- Existing billing single-interface pattern: `api/internal/modules/billing/domain/repository.go`
- Full-stack DTO flow: `docs/solutions/integration-issues/benefit-checklist-persistence.md`
- Clean Architecture dependency rule: `docs/solutions/architecture-patterns/clean-architecture-dependency-rule.md`
