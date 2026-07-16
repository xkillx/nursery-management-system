# Funding Management (MVP Scope)

## Objective

Provide a flexible funding management module that supports the majority of English nursery government funding scenarios while keeping the initial implementation simple and extensible.

This module follows the **Pareto Principle (80/20 Rule)**:

> Deliver the 20% of funding functionality that solves approximately 80% of everyday nursery operations.

The primary goal is to ensure nurseries can:

* Register funded children
* Allocate funded hours
* Generate correct invoices
* Manage funding periods
* Track funding history

---

# Business Goals

The Funding Management module shall enable nurseries to:

* Reduce manual funding calculations.
* Automatically deduct funded hours from invoices.
* Support both Term-Time and Stretched funding models.
* Maintain funding history for auditing.
* Prepare for future Local Authority claim integrations.

---

# MVP Scope (Phase 1)

## Supported Funding Schemes

### Universal Entitlement

* Age: 3–4 years
* 15 funded hours/week
* 570 funded hours/year

---

### Working Parent Entitlement

* Eligible children from 9 months to school age
* Up to 30 funded hours/week
* 1,140 funded hours/year

---

## Funding Models

Support:

* Term-Time
* Stretched

The system shall automatically calculate monthly funded hours based on the selected funding model.

---

## Funding Profile

Each child may have one active funding profile.

A funding profile shall include:

* Funding type
* Funding model
* Weekly funded hours
* Annual funded hours
* Effective date
* Expiry date
* Status

---

## Registration Integration

The Child Registration Wizard shall include a dedicated step:

### Step 4 — Funding & Benefits

Fields:

* Funding Type
* Funding Eligibility Code (optional)
* Funding Model
* Effective Date
* Expected Weekly Hours

---

## Booking Integration

When creating recurring bookings:

The system shall:

* Read the child's funding profile.
* Allocate funded hours automatically.
* Calculate remaining billable hours.

---

## Billing Integration

Invoice calculation:

```text
Booked Hours
− Funded Hours
= Billable Hours
```

Example

```text
Booked Hours:      120

Funded Hours:       95

Billable Hours:     25

Rate:              £8/hr

Invoice:         £200
```

---

## Funding History

The system shall maintain historical funding records.

Changes to funding shall never overwrite previous records.

Examples:

* Funding starts
* Funding ends
* Child changes eligibility
* Funding hours change

---

# Out of Scope (Phase 1)

The following features are intentionally excluded from the MVP:

* 2-Year-Old Disadvantaged Funding
* Early Years Pupil Premium (EYPP)
* Disability Access Fund (DAF)
* SEND Inclusion Funding
* Local Authority claim submission
* Government API integrations
* Funding forecasting
* Funding analytics
* Multi-authority funding rules

These features will be delivered in later phases.

---

# Phase 2

The following funding schemes will be added after MVP:

* 2-Year-Old Disadvantaged Funding
* Early Years Pupil Premium (EYPP)
* Disability Access Fund (DAF)
* SEND Inclusion Funding

Additional capabilities:

* Funding validation
* Funding claim management
* Funding reporting
* Funding audits

---

# Functional Requirements

## FR-001

The system shall allow a funding profile to be assigned to a child.

---

## FR-002

The system shall support Universal and Working Parent funding.

---

## FR-003

The system shall support both Term-Time and Stretched funding models.

---

## FR-004

The system shall automatically calculate funded hours available for each billing period.

---

## FR-005

The system shall automatically deduct funded hours from booked hours before invoice generation.

---

## FR-006

The system shall preserve funding history and effective dates.

---

## FR-007

The system shall prevent overlapping active funding profiles for the same child.

---

## FR-008

The system shall display remaining funded hours for the current billing period.

---

## Non-Functional Requirements

* Funding calculations shall be deterministic and reproducible.
* Funding rules shall be configurable without database schema changes.
* Historical funding records shall be immutable.
* Funding calculations shall complete within invoice generation without noticeable delay.
* The design shall allow additional government funding schemes to be added with minimal code changes.

---

# Future Architecture

The funding module should evolve into a reusable **Funding Engine**, where government funding schemes are represented as configurable rules rather than hard-coded logic.

```text
Funding Engine
├── Eligibility Rules
├── Funding Profiles
├── Funding Allocation
├── Booking Consumption
├── Invoice Deduction
├── Claim Management
└── Reporting
```

This architecture minimizes duplication, simplifies maintenance, and allows new government funding schemes to be introduced through configuration instead of significant code changes.
