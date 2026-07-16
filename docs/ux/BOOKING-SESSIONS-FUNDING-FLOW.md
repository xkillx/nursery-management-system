# Booking, Sessions & Funding — UX Page Flow

> Based on PRD: `docs/prds/PRD-Booking-Sessions-Funding.md`

---

## Information Architecture

```
Manager Sidebar
├── Bookings (NEW)
│   ├── Calendar View (default)
│   ├── Booking List
│   └── New Booking (stepper wizard)
├── Attendance (existing)
│   └── Register (enhance)
├── Funding (NEW)
│   ├── Overview
│   └── Child Funding Detail
├── Billing (existing)
│   └── Invoices (enhance)
└── Setup
    ├── Session Types (existing)
    └── Session Templates (existing)

Parent Portal
├── My Bookings (NEW)
├── My Attendance (NEW)
├── My Funding (NEW)
└── My Invoices (existing)
```

---

## Page-by-Page Specs

### 1. Bookings Calendar — `/manager/bookings`

**Purpose:** Visual overview of all bookings across the week/room.

| Element | Detail |
|---|---|
| Layout | FullCalendar component (already integrated) |
| Views | Week (default), Month, Day, List |
| Filters | Room dropdown, Session type chips, Status pills |
| Events | Color-coded by session type, stacked per child |
| Empty state | "No bookings this week" with quick-add CTA |
| Primary action | `+ New Booking` button (top-right) |
| Secondary | Toggle view, Print, Export |

**Interactions:**

- Click event → side panel (booking detail)
- Click empty slot → pre-filled new booking modal
- Drag event → reschedule (future enhancement)

---

### 2. Booking List — `/manager/bookings/list`

**Purpose:** Searchable, filterable table of all bookings.

| Column | Data |
|---|---|
| Child | Name + avatar |
| Session | Template name + time |
| Days | Visual day indicators (M T W T F pills) |
| Effective | Start → End dates |
| Funding | Badge (Private / 15h / 30h / Mixed) |
| Status | Active / Paused / Cancelled badge |
| Actions | Edit, Clone, Cancel, View child |

**Filters:** Search, Room, Session type, Funding type, Status, Date range

---

### 3. New Booking Wizard — `/manager/bookings/new`

**Purpose:** 4-step stepper wizard for creating a booking.

```
Step 1: Child & Session
├── Child selector (searchable dropdown)
├── Session type selector (card grid)
└── Booking type (Recurring / Ad-hoc toggle)

Step 2: Schedule
├── Day picker (M T W T F S S toggles)
├── Effective start date
├── Optional end date
└── Room assignment

Step 3: Funding
├── Funding type selector (Private / 15h / 30h / Mixed)
├── Weekly funded hours (auto-calculated)
├── Funding start/end dates
└── Local authority reference (optional)

Step 4: Review & Confirm
├── Summary card (child, sessions, schedule, funding)
├── Capacity warning (if applicable)
└── Confirm button
```

**Validation:** Each step validates before "Continue". Back returns without clearing.

---

### 4. Booking Detail Panel (side panel)

**Purpose:** Quick view/edit without leaving calendar.

| Section | Content |
|---|---|
| Header | Child name + session type + status badge |
| Schedule | Day pills + effective dates |
| Funding | Type badge + hours breakdown |
| Capacity | Room occupancy indicator |
| Actions | Edit, Clone, Cancel, View child |

---

### 5. Funding Overview — `/manager/funding`

**Purpose:** Dashboard of all funded children and hours utilization.

| Metric Cards | Detail |
|---|---|
| Total Funded Children | Count with trend |
| 15h Children | Count + percentage |
| 30h Children | Count + percentage |
| Hours Utilization | This week / total available |

**Table columns:**

| Column | Data |
|---|---|
| Child | Name |
| Funding Type | Badge |
| Weekly Entitlement | Hours |
| Hours Used | This week / entitlement |
| Period | Start → End dates |
| Status | Active / Expiring Soon / Expired |

**Filters:** Funding type, Status, Expiring within (30 / 60 / 90 days)

---

### 6. Child Funding Detail — `/manager/funding/:childId`

**Purpose:** Full funding history and allocation for a child.

| Section | Content |
|---|---|
| Summary | Type, hours, period, LA reference |
| Allocation Table | Date, Session, Funded Hours, Charged Hours |
| History | Previous funding periods |
| Actions | Edit funding, End funding |

---

### 7. Attendance Register — `/manager/attendance` (enhance existing)

**Purpose:** Daily check-in/out with booking-aware context.

| Enhancement | Detail |
|---|---|
| Pre-populated | Children auto-loaded from confirmed bookings |
| Status column | Booked → Present / Absent / Late / Early Pickup / No Show |
| Quick actions | Check-in, Check-out, Mark absent buttons |
| Capacity bar | Per-room occupancy indicator |
| Date picker | Navigate days, shows booking count per day |

---

### 8. Invoices — `/manager/invoices` (enhance existing)

**Purpose:** Add funding deduction line items.

| Enhancement | Detail |
|---|---|
| Line items | Session Fee, Funded Deduction, Meals, Snacks, Extra Hours, Late Pickup |
| Funding section | Separate section showing funded hours deduction |
| Total formula | Total Charges − Funding Deductions = Amount Payable |
| Status badges | Draft, Sent, Paid, Overdue, Partial |

---

### 9. Parent Portal: My Bookings — `/parent/bookings`

**Purpose:** View upcoming and recurring bookings.

| Section | Content |
|---|---|
| Upcoming | Next 7 days as cards (date, session, time, status) |
| Recurring | Weekly pattern display (day pills + session) |
| Funding | Badge showing funded hours used this week |
| Actions | Request session, Cancel eligible booking |

**Design:** Mobile-first cards, not tables. Large touch targets.

---

### 10. Parent Portal: My Funding — `/parent/funding`

**Purpose:** View funding entitlement and usage.

| Section | Content |
|---|---|
| Entitlement | Type badge + weekly hours |
| This Week | Hours used / entitlement (progress bar) |
| Period | Start → End dates |
| Breakdown | Table: Date, Session, Funded Hours, Charged Hours |

---

## Navigation Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    MANAGER FLOW                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Dashboard ──► Bookings Calendar ──► Booking Detail Panel    │
│       │              │                      │                │
│       │              ▼                      ▼                │
│       │         New Booking           Edit Booking           │
│       │         (4-step wizard)       (same wizard)          │
│       │                                                      │
│       ├──► Attendance Register ──► Mark Attendance           │
│       │                                                      │
│       ├──► Funding Overview ──► Child Funding Detail         │
│       │                                                      │
│       └──► Invoices ──► Invoice Detail ──► Generate          │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                    PARENT FLOW                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Parent Home ──► My Bookings ──► Request Session             │
│       │                                                      │
│       ├──► My Attendance                                     │
│       │                                                      │
│       ├──► My Funding                                        │
│       │                                                      │
│       └──► My Invoices ──► Pay Online                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Implementation Priority

| Phase | Pages | Rationale |
|---|---|---|
| **P1** | Session Types (exists), Session Templates (exists), Booking Wizard, Bookings List | Foundation for everything |
| **P2** | Attendance Register (enhance), Bookings Calendar | Operational core |
| **P3** | Funding Overview, Child Funding Detail, Invoice enhancement | Billing integration |
| **P4** | Parent Portal (all pages) | Self-service |

---

## Key UX Decisions

1. **Stepper wizard** for booking creation (matches child registration pattern already in codebase)
2. **Side panel** for booking detail (keeps calendar context visible)
3. **Dual rendering** on list pages (table desktop, cards mobile — existing pattern)
4. **Progress bars** for funding utilization (visual > numbers)
5. **Day pills** (M T W T F) for recurring schedule display (compact, recognizable)
6. **Color-coded session types** across all views for consistency
7. **Mobile-first** parent portal with card layouts and large touch targets
