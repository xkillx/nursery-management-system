# Typical UK Nursery Management System (NMS) Child Flow

## Core Principle

A UK Nursery Management System revolves around one core concept:

> **A child is registered once, booked many times, attends every day, and is invoiced every billing cycle.**

Most UK nursery software (Famly, Blossom, Nursery Story, eyManage, Connect Childcare, etc.) follows this same business workflow.

---

# Overall Child Lifecycle

```text
Parent Enquiry
      │
      ▼
Registration / Admission
      │
      ▼
Funding Setup
      │
      ▼
Booking / Session Pattern
      │
      ▼
Room Allocation
      │
      ▼
Attendance (Daily Register)
      │
      ▼
Daily Diary / Learning Journey
      │
      ▼
Invoice Generation
      │
      ▼
Payment
      │
      ▼
Reporting
      │
      ▼
Child Leaves Nursery
```

---

# Step 1 — Registration (One Time)

The registration process creates the master record for the child.

## Information Collected

```text
Child
├── Personal Details
├── Parent(s) / Guardians
├── Emergency Contacts
├── Medical Information
├── Allergies
├── GP Details
├── Permissions & Consents
├── Start Date
├── Leaving Date (optional)
└── Documents
```

## Child Status Lifecycle

```text
Enquiry
    │
    ▼
Registered
    │
    ▼
Active
    │
    ▼
Left Nursery
```

### Database Entities

- Child
- Parent
- Emergency Contact
- Medical Record
- Consent
- Document

---

# Step 2 — Funding Setup

Government funding is configured after registration.

## Common Funding Types

- 15 Hours Funding
- 30 Hours Funding
- 2-Year Funding
- EYPP
- Disability Access Fund (DAF)
- No Funding

## Funding Information

```text
Funding Record

effective_from
effective_to

hours_per_week

term_time_only

stretched

local_authority

hourly_rate
```

Funding later reduces the parent's invoice automatically.

---

# Step 3 — Booking / Session Pattern

This is the heart of the nursery system.

A booking defines when the child is expected to attend.

Example:

```text
Monday
08:00 - 18:00

Tuesday
08:00 - 18:00

Wednesday
No Booking

Thursday
08:00 - 13:00

Friday
08:00 - 18:00
```

Booking Pattern stores:

```text
Booking Pattern

Effective From

Effective To

Weekly Schedule

Room

Session Type

Funding Applied

Notes
```

Bookings become the source of truth for:

- Attendance
- Occupancy
- Staff Ratio
- Funding
- Billing
- Reports

---

# Step 4 — Room Allocation

Assign the child to a room.

Example:

```text
Baby Room

Toddler Room

Pre-School
```

Room assignment may be automatic based on:

- Child age
- Booking
- Capacity

---

# Step 5 — Daily Attendance

Every morning the system creates the register from bookings.

Expected Register:

```text
Monday

✓ John
✓ Emma
✓ Oliver
✓ Sophia
```

Staff record:

```text
Present

Absent

Holiday

Sick

Late

Left Early
```

Actual attendance:

```text
Check In
08:04

Check Out
17:46

Collected By
Mother
```

Attendance affects:

- Occupancy
- Staff Ratio
- Safeguarding
- Invoicing (depending on nursery policy)

---

# Step 6 — Daily Diary / Learning Journey

Throughout the day staff record activities.

```text
Meals

Bottle Feed

Sleep

Nappy Changes

Medication

Incidents

Accidents

Observations

Learning Journey

Photos

Messages
```

Parents can usually view these in the Parent App.

---

# Step 7 — Invoice Generation

Usually generated monthly.

Invoice Engine reads:

```text
Booking Pattern
        │
        ▼
Attendance (optional)
        │
        ▼
Government Funding
        │
        ▼
Discounts
        │
        ▼
Extras
        │
        ▼
Invoice
```

Example:

```text
Booked Hours
160

Government Funding
-65

Meals
+20

Late Collection
+15

Sibling Discount
-10

---------------------
Total
£560
```

Additional charges may include:

- Meals
- Trips
- Late Collection
- Extra Sessions
- Consumables

---

# Step 8 — Payment

Parents pay invoices using:

- Bank Transfer
- Debit/Credit Card
- Cash
- Childcare Vouchers
- Tax-Free Childcare

Invoice Status:

```text
Draft
   │
   ▼
Issued
   │
   ▼
Partially Paid
   │
   ▼
Paid
   │
   ▼
Overdue
```

---

# Step 9 — Reporting

Management dashboards typically show:

```text
Occupancy %

Attendance %

Income

Outstanding Debt

Funding Claimed

Room Utilisation

Staff Ratio

Child Count

Available Spaces
```

---

# Complete End-to-End Workflow

```text
Parent
   │
   ▼
Register Child
   │
   ▼
Configure Funding
   │
   ▼
Create Booking Pattern
   │
   ▼
Assign Room
   │
   ▼
Daily Register
   │
   ▼
Child Check-In
   │
   ▼
Daily Diary
   │
   ▼
Child Check-Out
   │
   ▼
End of Month
   │
   ▼
Generate Invoice
   │
   ▼
Receive Payment
   │
   ▼
Management Reports
```

---

# Business Data Dependency

Everything flows from the booking.

```text
Child
   │
   ▼
Funding
   │
   ▼
Booking Pattern
   │
   ├────────────► Occupancy
   │
   ├────────────► Attendance
   │
   ├────────────► Staff Ratio
   │
   ├────────────► Daily Diary
   │
   ▼
Invoice
   │
   ▼
Payment
   │
   ▼
Reports
```

---

# Entity Relationship (Simplified)

```text
Child
 │
 ├── Parents
 ├── Funding Records
 ├── Booking Patterns
 ├── Room Assignment
 ├── Attendance Records
 ├── Daily Diaries
 ├── Learning Journeys
 ├── Invoices
 └── Payments
```

---

# Recommended MVP Implementation Order

To minimize dependencies, implement modules in this order:

1. Child Registration
2. Parent Management
3. Funding Management
4. Booking / Session Pattern
5. Room Allocation
6. Attendance Register
7. Daily Diary
8. Invoice Generation
9. Payment Recording
10. Reporting Dashboard

---

# Key Design Principle

```text
Registration
      │
      ▼
Funding
      │
      ▼
Booking
      │
      ▼
Attendance
      │
      ├── Occupancy
      ├── Staff Ratio
      ├── Daily Diary
      └── Learning Journey
      │
      ▼
Invoice
      │
      ▼
Payment
      │
      ▼
Reporting
```

## Most Important Concept

The **Booking Pattern** is the central object in a UK Nursery Management System because it drives:

- Daily Attendance Registers
- Occupancy Calculations
- Staff-to-Child Ratios
- Government Funding Allocation
- Monthly Invoices
- Capacity Planning
- Revenue Forecasting
- Management Reporting

Everything else in the system is derived from or linked to the booking pattern.
