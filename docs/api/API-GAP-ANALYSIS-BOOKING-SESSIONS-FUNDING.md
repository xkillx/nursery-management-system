# API Gap Analysis — Booking, Sessions & Funding UX Flow

> Based on UX spec: `docs/ux/BOOKING-SESSIONS-FUNDING-FLOW.md`

---

## Existing Endpoints (already built)

| Domain | Endpoints | Module |
|--------|-----------|--------|
| **Session Types** | `GET/POST/PATCH/Archive/Reactivate /sites/:site_id/session-types[/:id]` | `sessiontypes` |
| **Session Templates** | `GET/POST/PATCH/Archive/Reactivate /sites/:site_id/session-templates[/:id]` | `sessiontemplates` |
| **Funding** | `GET /funding/overview`, `GET/PUT /funding/children/:child_id` | `funding` |
| **Attendance** | `POST /attendance/check-ins`, `POST /attendance/check-outs`, `POST /attendance/corrections`, `GET /attendance/sessions` | `attendance` |
| **Billing** | Full invoice lifecycle (22 manager + 3 parent endpoints) including line items, PDF, export | `billing` |
| **Ad-hoc Bookings** | `GET/POST /ad-hoc-bookings`, `POST .../cancel` | `ad_hoc_bookings` |
| **Hourly Bookings** | `GET/POST /hourly-bookings`, `POST .../cancel` | `hourly_bookings` |
| **Children** | `GET /children`, `GET /children/:id`, booking patterns CRUD, funding sub-resource | `children` |
| **Rooms** | CRUD exists | `rooms` |

---

## GAPS — New APIs Needed

### 1. Unified Bookings Module (NEW — biggest gap)

The UX needs a **single unified booking concept** that covers recurring, ad-hoc, and hourly bookings. Today these are fragmented across 3 modules with no shared list/detail view.

**Manager endpoints:**

| Method | Path | Purpose | UX Page |
|--------|------|---------|---------|
| `GET` | `/sites/:site_id/bookings` | Unified list with filters (date range, room, session_type, status, funding_type, search by child name). Returns recurring + ad-hoc + hourly. Supports `view=calendar` (grouped by date) and `view=list` (flat table). | Bookings Calendar, Booking List |
| `GET` | `/sites/:site_id/bookings/:booking_id` | Single booking detail (child, session, schedule, funding, capacity info) | Booking Detail Panel |
| `POST` | `/sites/:site_id/bookings` | Create recurring booking (child_id, session_template_id, days[], effective_start, effective_end, room_id, funding_type, funding_hours, la_reference) | New Booking Wizard (Step 4) |
| `PATCH` | `/sites/:site_id/bookings/:booking_id` | Edit booking (schedule, room, funding, end date) | Edit Booking |
| `POST` | `/sites/:site_id/bookings/:booking_id/clone` | Clone booking for same child or different child | Booking List actions |
| `POST` | `/sites/:site_id/bookings/:booking_id/cancel` | Cancel booking | Booking List / Detail actions |
| `POST` | `/sites/:site_id/bookings/:booking_id/pause` | Pause booking | Booking Detail actions |
| `GET` | `/sites/:site_id/bookings/capacity` | Room capacity snapshot for date range (used for warnings in wizard) | New Booking Wizard (Step 4 warning) |

### 2. Attendance Register Enhancement

| Method | Path | Purpose | UX Page |
|--------|------|---------|---------|
| `GET` | `/sites/:site_id/register?date=YYYY-MM-DD` | Daily register — returns children expected today (from bookings) merged with attendance status. Includes room capacity bar data. | Attendance Register |
| `GET` | `/sites/:site_id/register/summary?date=YYYY-MM-DD` | Per-room booking count per day (for date picker badges) | Attendance Register date picker |

### 3. Funding Module Enhancement

| Method | Path | Purpose | UX Page |
|--------|------|---------|---------|
| `GET` | `/funding/overview` (enhance) | Must return: total funded children, 15h count, 30h count, hours utilization this week, expiring soon count | Funding Overview metric cards |
| `GET` | `/funding/children/:child_id` (enhance) | Must return: allocation table (date, session, funded_hours, charged_hours) + history of previous funding periods | Child Funding Detail |
| `GET` | `/funding/expiring?within=30` | Children with funding expiring within N days | Funding Overview filter |

### 4. Billing Enhancement

| Method | Path | Purpose | UX Page |
|--------|------|---------|---------|
| Invoice lines (existing) | `POST/PUT/DELETE /invoices/:id/lines` | Verify "Funded Deduction" line type is supported | Invoices funding section |

### 5. Parent Portal Endpoints (ALL NEW)

| Method | Path | Purpose | UX Page |
|--------|------|---------|---------|
| `GET` | `/parent/bookings` | My upcoming bookings (next 7 days) + recurring pattern | Parent: My Bookings |
| `GET` | `/parent/bookings/recurring` | Recurring booking patterns for my children | Parent: My Bookings |
| `POST` | `/parent/bookings/requests` | Request a new session (creates pending booking for manager approval) | Parent: My Bookings |
| `POST` | `/parent/bookings/:booking_id/cancel` | Cancel eligible booking | Parent: My Bookings |
| `GET` | `/parent/funding` | My children's funding entitlement + usage this week | Parent: My Funding |
| `GET` | `/parent/funding/:child_id/breakdown` | Detailed breakdown: date, session, funded hours, charged hours | Parent: My Funding |
| `GET` | `/parent/attendance` | My children's attendance records | Parent: My Attendance |

---

## Summary

| Category | New Endpoints | Enhanced Endpoints |
|----------|--------------|-------------------|
| **Unified Bookings** (new module) | 8 | — |
| **Attendance Register** | 2 | — |
| **Funding** | 1 | 2 (enhance existing) |
| **Billing** | — | 1 (verify line types) |
| **Parent Portal** | 7 | — |
| **Total** | **18 new** | **3 enhanced** |

---

## Recommended Implementation Order

| Phase | Work | Rationale |
|-------|------|-----------|
| **P1** | Unified bookings module (CRUD + list/calendar) | Foundation — every other page depends on bookings existing |
| **P2** | Attendance register endpoint + capacity endpoint | Operational core — needs booking data to pre-populate |
| **P3** | Funding enhancements + billing line type verification | Billing integration — needs bookings to calculate utilization |
| **P4** | Parent portal endpoints (7 new) | Self-service — reads from same booking/funding/attendance data |
