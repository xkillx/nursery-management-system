# PRD — Booking, Sessions & Funding Management

## Module Overview

The Booking, Sessions & Funding module enables nurseries to manage contracted attendance, ad-hoc bookings, government-funded childcare, daily attendance, and monthly billing.

The module is designed around reusable session templates and recurring booking patterns to minimize administrative work while supporting the majority of UK nursery operational scenarios.

---

# Goals

## Business Goals

- Simplify child booking management
- Reduce manual scheduling effort
- Support UK childcare funding models
- Enable accurate monthly invoicing
- Provide parents with self-service booking capabilities

## User Goals

### Nursery Staff

Nursery staff should be able to:

- Create recurring bookings in minutes
- Record attendance quickly
- Apply funded hours automatically
- Generate invoices with minimal manual adjustments

### Parents

Parents should be able to:

- View contracted sessions
- Request additional sessions
- View funded hours
- Pay invoices online

---

# Scope

## Included in MVP

- Session Templates
- Booking Management
- Funding Allocation
- Attendance Register
- Billing Integration
- Parent Booking Portal

## Out of Scope

- Holiday Clubs
- Flexible Hourly Booking
- Government Funding Claim Submission
- AI Scheduling
- Occupancy Forecasting

---

# Functional Requirements

# 1. Session Management

## Purpose

Nurseries define reusable session templates that can be assigned to bookings.

## Functional Requirements

The system shall allow administrators to:

- Create session templates
- Edit session templates
- Archive session templates
- Configure start and end times
- Define session capacity
- Configure whether meals are included
- Configure whether funded hours are eligible
- Configure applicable age groups or rooms

## Default Session Templates

| Session | Typical Time | Description |
|----------|--------------|-------------|
| Full Day | 08:00–18:00 | Standard nursery day |
| Morning (AM) | 08:00–13:00 | Half-day morning session |
| Afternoon (PM) | 13:00–18:00 | Half-day afternoon session |
| Breakfast Club | 07:30–09:00 | Before-school care |
| After School Club | 15:00–18:00 | After-school care |

---

# 2. Booking Management

## Purpose

Manage recurring and one-off attendance schedules for children.

## Supported Booking Types

| Booking Type | Description |
|--------------|-------------|
| Regular Recurring | Weekly contracted schedule |
| Full Day | Whole-day attendance |
| Half Day | Morning or afternoon attendance |
| Ad-hoc | One-time booking |
| Funded Booking | Booking using funded childcare hours |
| Wraparound Care | Before and after-school care |

## Functional Requirements

The system shall allow staff to:

- Create bookings
- Edit bookings
- Cancel bookings
- Clone bookings
- View booking calendar
- Search bookings
- Filter bookings
- View child timetable

---

# 3. Recurring Booking Rules

The system shall support:

- Every Monday
- Every Tuesday
- Monday–Friday
- Monday, Wednesday, Friday
- Every two weeks (optional)

Each recurring booking shall include:

- Effective start date
- Optional end date
- Assigned room
- Session template
- Funding allocation
- Booking status

---

# 4. Session Capacity

The system shall prevent overbooking.

Each session shall define:

- Maximum children
- Current occupancy
- Remaining places
- Waiting list indicator (future enhancement)

The system shall display warnings before capacity is exceeded.

---

# 5. Funding Management

## Supported Funding Types

| Funding Type | Description |
|--------------|-------------|
| Private | Parent pays full fees |
| 15 Hours Funding | Government-funded childcare |
| 30 Hours Funding | Extended government funding |
| Mixed Funding | Combination of funded and private payment |

Funding information shall include:

- Funding type
- Weekly funded hours
- Funding start date
- Funding end date
- Local authority reference (optional)
- Eligibility period

---

# 6. Mixed Funding

The system shall support partial funding.

## Example

Weekly Attendance

- 30 hours

Funding

- 15 funded hours

Parent Pays

- Remaining 15 hours
- Meals
- Snacks
- Activities

Funding deductions shall be calculated automatically before invoice generation.

---

# 7. Attendance Management

Attendance records are automatically generated from confirmed bookings.

## Attendance Statuses

| Status | Description |
|---------|-------------|
| Booked | Session scheduled |
| Present | Child attended |
| Absent | Child did not attend |
| Late Arrival | Arrived after session start |
| Early Pickup | Collected before session end |
| No Show | Did not attend without notice |

Staff shall be able to:

- Mark attendance
- Record arrival time
- Record departure time
- Add attendance notes
- Override attendance status

---

# 8. Billing Integration

Invoices shall be generated from confirmed bookings and attendance.

## Supported Charge Types

| Charge Type | Description |
|-------------|-------------|
| Session Fee | Cost of booked session |
| Funded Deduction | Government funding deduction |
| Meals | Breakfast, lunch, tea |
| Snacks | Additional snacks |
| Extra Hours | Attendance beyond booked session |
| Late Pickup Fee | Late collection charge |
| Activities | Optional classes or trips |
| Registration Fee | One-time enrolment fee |

Each charge shall appear as an individual invoice line item.

---

# 9. Parent Portal

Parents shall be able to:

- View upcoming bookings
- View recurring schedules
- Request additional sessions
- Cancel eligible bookings
- View funded hours
- View attendance history
- View invoices
- Pay invoices online

Cancellation permissions shall follow nursery-defined policies.

---

# Business Rules

## Booking

- A child must belong to an active nursery before bookings can be created.
- A booking must reference a valid session template.
- Bookings cannot exceed session capacity unless overridden by an authorised user.
- Cancelled bookings shall not generate attendance or invoice charges.

## Funding

- Funded hours shall always be applied before private charges.
- Funding cannot exceed the child's eligible entitlement.
- Mixed funding shall automatically split funded and chargeable hours.
- Funding periods shall respect effective dates.

## Attendance

- Attendance records shall be automatically created from confirmed bookings.
- Attendance may only be modified by authorised staff.
- Attendance changes after invoicing shall trigger invoice recalculation or adjustment according to nursery policy.

## Billing

- Only confirmed bookings shall generate invoice items.
- Funding deductions shall appear as separate invoice lines.
- Additional charges shall be itemised individually.
- Invoice totals shall equal:

```
Total Charges
− Funding Deductions
= Amount Payable
```

---

# User Stories

## Nursery Administrator

- As an administrator, I want to configure reusable session templates so bookings remain consistent.
- As an administrator, I want to define funding rules so invoices are calculated correctly.

## Nursery Staff

- As a practitioner, I want to create recurring bookings so I do not have to enter schedules every week.
- As a practitioner, I want to record attendance quickly using the daily register.

## Parent

- As a parent, I want to request additional sessions.
- As a parent, I want to see my funded hours.
- As a parent, I want to pay invoices online.

---

# Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Performance | Booking calendar loads within 2 seconds for a typical nursery. |
| Availability | Booking and attendance features are available during nursery operating hours with high reliability. |
| Security | Role-Based Access Control (RBAC) protects booking, funding and billing operations. |
| Audit | All changes to bookings, attendance, funding and invoices are audit logged. |
| Scalability | Supports multiple rooms and multiple nursery branches without architectural redesign. |

---

# Acceptance Criteria

- Staff can create recurring bookings.
- Staff can create ad-hoc bookings.
- Session templates are configurable.
- Capacity limits prevent overbooking.
- Attendance records are automatically generated from bookings.
- Funding deductions are calculated automatically.
- Mixed funding invoices are calculated correctly.
- Monthly invoices include itemised charges and funding deductions.
- Parents can:
  - View bookings
  - View attendance
  - View funded hours
  - View invoices
  - Request additional bookings
  - Pay invoices online

---

# Future Enhancements

## Booking

- Hourly bookings
- Flexible bookings
- Emergency bookings
- Holiday clubs
- Waiting lists

## Operations

- Automatic room transfers
- Staff ratio validation
- Occupancy forecasting
- AI scheduling

## Funding

- Automatic funding validation
- Local authority funding claim exports
- Childcare Voucher support
- Tax-Free Childcare integration

## Payments

- Direct Debit collection
- Additional payment gateways

## Parent Experience

- Mobile application
- Parent messaging
- Digital check-in/check-out

## Child Records

- EYFS learning journal
- Accident records
- Medication administration

---

# Summary

## Booking Types

- Regular Recurring Booking
- Full-Day Booking
- Half-Day Booking
- Ad-hoc Booking
- Funded Booking
- Wraparound Care

## Session Templates

- Full Day
- Morning (AM)
- Afternoon (PM)
- Breakfast Club
- After School Club

## Funding Types

- Private
- 15 Hours Funding
- 30 Hours Funding
- Mixed Funding

## Attendance Statuses

- Booked
- Present
- Absent
- Late Arrival
- Early Pickup
- No Show

## Charge Types

- Session Fee
- Funding Deduction
- Meals
- Snacks
- Extra Hours
- Late Pickup Fee
- Activities
- Registration Fee

---

# Architecture Principle

> **Booking is the source of truth.**

All downstream operational processes derive from bookings:

```
Session Template
        │
        ▼
Recurring / Ad-hoc Booking
        │
        ├──────────────► Attendance Register
        │
        ├──────────────► Room Occupancy
        │
        ├──────────────► Staff Ratio Validation
        │
        ├──────────────► Funding Allocation
        │
        └──────────────► Billing & Invoice Generation
```

This architecture minimizes duplicated business logic, ensures consistency across modules, and provides a scalable foundation for future enhancements.
