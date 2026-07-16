# UK Nursery Management System (NMS) — Pareto (80/20) Feature Prioritization

## Goal

Build the **20% of features that deliver 80% of the value** for most UK nurseries.

---

# Tier 1 — Core MVP (Highest Priority)

These are the features every nursery uses almost every day.

---

## 1. Child Management ⭐⭐⭐⭐⭐

Manage all child information.

### Features

- Child Profile
- Parent/Guardian Information
- Emergency Contacts
- Medical Information
- Allergies
- Dietary Requirements
- Immunisation Records (Optional)
- Authorized Pickup Persons
- Documents
- Child Status (Active/Inactive/Leaving)

---

## 2. Booking Management ⭐⭐⭐⭐⭐

The heart of the nursery system.

### Booking Types

- Regular (Recurring) Booking
- One-off (Ad-hoc) Booking
- Full-Day Booking
- Half-Day Booking
- Morning Session
- Afternoon Session
- Hourly Booking
- Holiday Booking
- Wraparound Care

### Features

- Booking Calendar
- Weekly Schedule
- Booking Availability
- Booking Changes
- Cancellation
- Waitlist

---

## 3. Funding Management ⭐⭐⭐⭐⭐

Essential for UK nurseries.

### Supported Funding

- 15 Hours Funding
- 30 Hours Funding
- Working Family Funding
- Funded + Private Mixed Sessions

### Features

- Funding Allocation
- Funding Period
- Funding Balance
- Funding Validation
- Funding Reports

---

## 4. Attendance ⭐⭐⭐⭐⭐

Used every day by nursery staff.

### Features

- Check In
- Check Out
- Live Register
- Late Arrival
- Early Collection
- Absent
- Sick
- Holiday
- Live Occupancy

---

## 5. Invoicing & Billing ⭐⭐⭐⭐⭐

Automatically generate invoices from bookings.

### Invoice Items

- Session Fees
- Funding Deduction
- Meals
- Snacks
- Registration Fee
- Deposit
- Extra Hours
- Late Collection Fee
- Discounts
- Credits

### Features

- Monthly Invoice
- Manual Invoice
- Credit Notes
- Invoice History
- Outstanding Balance

---

## 6. Payments ⭐⭐⭐⭐⭐

Track all parent payments.

### Features

- Paid
- Unpaid
- Partially Paid
- Refund
- Outstanding Balance
- Payment History

---

## 7. Room & Capacity Management ⭐⭐⭐⭐

### Features

- Baby Room
- Toddler Room
- Preschool Room
- Capacity Management
- Occupancy
- Staff-to-Child Ratio

---

## 8. Staff Management ⭐⭐⭐⭐

### Features

- Staff Profile
- Room Assignment
- Working Shifts
- Key Person Assignment

---

## 9. Parent Portal ⭐⭐⭐⭐

Parents can:

- View Bookings
- View Attendance
- View Invoices
- Pay Invoices
- Request Booking Changes
- Update Child Information
- Download Documents

---

## 10. Reports ⭐⭐⭐⭐

### Operational Reports

- Attendance Report
- Child Register
- Occupancy Report
- Room Utilisation
- Funding Report
- Revenue Report
- Outstanding Invoice Report

---

# Tier 2 — Important Features

These improve nursery operations after the MVP.

---

## Daily Diary

Track children's daily activities.

### Features

- Meals
- Snacks
- Sleep
- Nappy Changes
- Toileting
- Activities
- Photos
- Notes

---

## EYFS Learning Journey

Support child development tracking.

### Features

- Observations
- Assessments
- Milestones
- Next Steps
- Progress Reports

---

## Parent Communication

### Features

- Messages
- Announcements
- Notifications
- Reminders

---

## Admissions

### Features

- Enquiries
- Registration Forms
- Contracts
- Deposits
- Start Date
- Waiting List

---

# Tier 3 — Nice to Have

Useful for larger nurseries or multi-site operations.

## Operations

- Accident Reports
- Incident Reports
- Medication Records
- Visitor Log
- Inventory Management
- Kitchen & Meal Planning
- Staff Training Records
- Audit Logs
- Multi-Branch Support
- Digital Signatures
- API Integrations

---

# Suggested Sidebar Menu

```text
Dashboard

Children
├── Children
├── Parents
├── Admissions
└── Waiting List

Bookings
├── Calendar
├── Regular Bookings
├── One-off Bookings
├── Funding
└── Availability

Attendance
├── Check In / Out
├── Registers
└── Occupancy

Rooms
├── Rooms
├── Capacity
└── Staff Ratio

Billing
├── Invoices
├── Payments
├── Funding
└── Discounts

Daily Care
├── Diary
├── Meals
├── Sleep
├── Nappies
└── Medication

Learning
├── Observations
├── EYFS
└── Reports

Staff
├── Staff
├── Shifts
└── Key Persons

Parents
├── Portal
├── Messages
└── Documents

Reports

Settings
```

---

# Recommended Database Modules

| Module | Priority |
|----------|----------|
| Children | ⭐⭐⭐⭐⭐ |
| Parents | ⭐⭐⭐⭐⭐ |
| Bookings | ⭐⭐⭐⭐⭐ |
| Sessions | ⭐⭐⭐⭐⭐ |
| Attendance | ⭐⭐⭐⭐⭐ |
| Funding | ⭐⭐⭐⭐⭐ |
| Invoices | ⭐⭐⭐⭐⭐ |
| Payments | ⭐⭐⭐⭐⭐ |
| Rooms | ⭐⭐⭐⭐ |
| Staff | ⭐⭐⭐⭐ |
| Parent Portal | ⭐⭐⭐⭐ |
| Reports | ⭐⭐⭐⭐ |
| Daily Diary | ⭐⭐⭐ |
| EYFS | ⭐⭐⭐ |
| Messaging | ⭐⭐⭐ |
| Admissions | ⭐⭐⭐ |
| Waiting List | ⭐⭐⭐ |
| Inventory | ⭐⭐ |
| Audit Logs | ⭐⭐ |
| Multi-Branch | ⭐⭐ |

---

# Recommended MVP Development Order

## Phase 1 — Core Foundation

1. Authentication & Roles
2. Child Management
3. Parent Management
4. Room Management
5. Staff Management

---

## Phase 2 — Daily Operations

6. Booking Management
7. Attendance
8. Funding Management

---

## Phase 3 — Finance

9. Invoice Generation
10. Payment Tracking

---

## Phase 4 — Parent Experience

11. Parent Portal
12. Parent Messaging

---

## Phase 5 — Reporting

13. Attendance Reports
14. Funding Reports
15. Financial Reports

---

## Phase 6 — Child Development

16. Daily Diary
17. EYFS Learning Journey

---

## Phase 7 — Advanced Features

18. Waiting List
19. Admissions
20. Inventory
21. Medication
22. Accident & Incident Logs
23. Multi-Branch Support
24. API Integrations
25. Audit Logs

---

# Pareto Summary

## 80% Value (Build First)

- Child Management
- Parent Management
- Booking Management
- Funding Management
- Attendance
- Invoicing
- Payments
- Rooms & Capacity
- Staff Management
- Parent Portal
- Reports

These modules represent the core functionality used by nearly every UK nursery every day and provide the highest return on development effort.
