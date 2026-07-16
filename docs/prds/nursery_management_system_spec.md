# UK Nursery Management System (NMS) Product Specification

This document defines a production-ready web application for a UK Nursery Management System (NMS) designed for nursery managers, staff, owners, finance teams, and parents.[cite:1][cite:9] The product scope combines attendance, occupancy, ratios, parent communication, child learning journeys, and finance operations, aligned with the feature direction visible in leading nursery platforms such as Famly, Blossom Educational, and Connect Childcare.[cite:1][cite:6][cite:9]

## Product principles

The application should prioritize simplicity, minimal clicks, high-density workflows, and excellent mobile support because nursery operators frequently spend long working days inside the product managing attendance, staffing, parent communication, bookings, and billing.[cite:1][cite:8][cite:10] The UX should follow proven SaaS patterns from modern products such as Linear, Stripe Dashboard, GitHub, and Notion: keyboard-friendly navigation, compact tables, predictable page structure, strong search, and contextual side panels.[cite:10]

## Site map

### Staff-facing application

- Dashboard
- Children
  - Child List
  - Child Profile
  - Enrolment & Plans
  - Medical & SEN
  - Emergency Contacts
  - Authorized Collectors
  - Documents & Consents
  - Learning Journey
  - Observations
  - Assessments & EYFS
  - Progress Reports
  - Daily Diary
- Parents
  - Parent List
  - Parent Profile
  - Linked Children
  - Contact Preferences
  - Communication History
- Rooms & Attendance
  - Room List
  - Room Overview
  - Live Room Register
  - Headcounts
  - Ratios & Alerts
  - Daily Attendance Overview
  - Future Attendance Forecast
- Bookings & Schedules
  - Booking Calendar
  - Weekly Planner
  - Daily Planner
  - Availability & Occupancy Planner
  - Waitlist
  - Enquiries
  - Booking Requests
  - Session Templates & Pricing
- Staff & HR
  - Staff List
  - Staff Profile
  - Qualifications & Training
  - Contracts & Hour Types
  - Staff Rotas / Shift Planner
  - Leave & Absence
  - Payroll Export
- Finance & Billing
  - Finance Dashboard
  - Invoices List
  - Invoice Detail
  - Batch Invoicing
  - Funding & Grants
  - Payments & Reconciliation
  - Debt Management
  - Revenue & Margin Reports
  - Revenue Forecasting
  - Pricing & Discounts
  - Finance Exports
- Learning & Curriculum
  - Curriculum Frameworks
  - Learning Journals
  - Observations Feed
  - Cohort Tracking
  - Assessments & Next Steps
  - Reports for Ofsted / Inspections
- Communication & Messaging
  - Newsfeed
  - Direct Messaging
  - Announcements & Events
  - Templates & AI Assistant
  - Notification Settings
- Compliance & Safeguarding
  - Accident & Incident Reports
  - Medication Records
  - Safeguarding Logs
  - Policies & Procedures
  - Audit Logs
- Documents & Files
  - Document Library
  - Per-child Documents
  - Per-staff Documents
  - Policy Packs
  - Signed Consents
- Reports & Insights
  - Occupancy Reports
  - Ratio Compliance Reports
  - Attendance Reports
  - Funding Reports
  - Financial Summaries
  - Parent Engagement
  - Group-level Insights
- Enquiries & Admissions
  - Enquiry List
  - Enquiry Detail
  - Pipeline Status
  - Registration Forms
  - Conversion Reports
- Settings & Administration
  - Organisation Settings
  - Site / Nursery Settings
  - Rooms Configuration
  - Session Types & Pricing
  - Funding Rules
  - Role & Permissions
  - Audit Trail
  - Integrations
- Help & Support
  - In-app Help Center
  - Onboarding Checklists
  - Tutorials & Webinars

### Parent portal

- Parent Dashboard
- My Children
  - Daily Diary
  - Learning Journey
  - Attendance & Sessions
  - Invoices & Payments
  - Consents & Documents
- Messaging
  - Staff Messages
  - Announcements & Events
- Bookings
  - Booking Requests
  - Extra Sessions / Ad-hoc
- Account & Settings
  - Contact Details
  - Notification Preferences
  - Language Preferences

This structure reflects core functional areas found in leading childcare products, especially attendance, ratios, learning journals, occupancy, messaging, and invoicing.[cite:1][cite:8][cite:10]

## User roles

| Role | Primary dashboard | Core permissions | Daily workflow | Most-used pages | Navigation emphasis |
|------|-------------------|------------------|----------------|-----------------|---------------------|
| Super Admin | Group overview | Full cross-tenant administration | Manage organisations, permissions, billing, integrations | Group Insights, Roles, Audit Logs | Group-first navigation |
| Nursery Owner | Commercial dashboard | Business-wide read/write, strong finance control | Monitor revenue, occupancy, debt, staffing efficiency | Finance Dashboard, Occupancy, Reports | Finance and performance first |
| Nursery Manager | Operations dashboard | Full nursery operations | Attendance, staffing, bookings, incidents, parent comms | Room Register, Rotas, Calendar, Messaging | Operations-first |
| Deputy Manager | Operations dashboard | Similar to manager with scoped authority | Cover rota gaps, monitor rooms, review incidents | Registers, Rotas, Daily Attendance | Operations-first |
| Room Leader | Room dashboard | Assigned-room operational control | Manage room register, observations, diaries | Live Room Register, Daily Diary, Observations | Room-first |
| Practitioner | Task dashboard | Child-level write within assignment | Log care events, learning observations, messages | Daily Diary, Observation Editor, Messaging | Mobile-first slim nav |
| Finance/Admin | Finance dashboard | Billing, funding, debt management | Invoicing, reconciliation, funding claims | Invoices, Payments, Funding | Finance-first |
| Parent | Parent home | Access only own family data | Review updates, pay invoices, request bookings | Parent Dashboard, Messages, Invoices | Consumer-style simple nav |

The role model should be strict and auditable because child data, safeguarding logs, attendance records, and payment data require clear visibility boundaries and traceability.[cite:6][cite:10]

## Major pages

### Dashboard

**Purpose:** central operational view for each role.[cite:1][cite:4]

**Primary users:** all roles with role-specific variants.[cite:1][cite:4]

**Main components:** KPI cards, alerts, timeline widgets, activity feeds, occupancy or finance charts, quick actions, notification center.[cite:4][cite:10][cite:14]

**Actions:** open room register, send announcement, approve booking, run invoicing, review debt, add observation.[cite:8][cite:10]

**Filters:** nursery, date range, room, status, staff group.[cite:14]

**Empty state:** role-based guidance such as “No urgent issues right now” with next-step actions.

**Loading state:** dashboard skeletons for cards, charts, and activity feed.

**Error state:** persistent inline banner with retry and partial fallback rendering.

### Child List

**Purpose:** searchable high-density register of all enrolled, waiting, and leaving children.

**Primary users:** manager, deputy, admin, room leader.[cite:1][cite:8]

**Components:** search, token filters, bulk actions bar, sortable data table, saved views, quick-create button.

**Actions:** create child, archive child, assign room, export CSV, bulk message parents.

**Filters:** status, room, age band, funding eligibility, SEN flag, attendance pattern.

**Table columns:** child, DOB, room, status, funded hours, attendance pattern, balance state, key alerts.

**Dialogs:** create child wizard, archive flow, bulk assignment, export options.

### Child Profile

**Purpose:** source of truth for personal, legal, medical, attendance, learning, and communication information.[cite:1][cite:8][cite:11]

**Primary users:** manager, room leader, practitioner, parent (partial).

**Main components:** profile header, tabbed sections, activity timeline, document panel, quick actions drawer.

**Tabs:** overview, contacts, sessions, medical, collectors, learning, diary, documents, finance summary, communication history.

**Actions:** edit details, add observation, log medication, upload consent, move room, mark leaving.

**Permissions:** parents get read access to family-safe sections only; staff access depends on assignment and role.[cite:11]

### Live Room Register

**Purpose:** real-time attendance and ratio management with rapid tablet interaction.[cite:2][cite:14]

**Primary users:** room leader, practitioner, manager.

**Main components:** room summary bar, present/expected counts, ratio badge, child status list, quick toggles, headcount confirmation panel.[cite:14]

**Actions:** sign in, sign out, move room, mark absent, mark sleep or outdoor status, open child quick drawer.

**Filters:** present, expected, absent, session block, key worker.

**Widgets:** ratio indicator, occupancy meter, alert badges.

### Booking Calendar

**Purpose:** visual scheduling of sessions, occupancy, availability, and booking requests.[cite:14]

**Primary users:** manager, deputy, owner, admin.

**Main components:** week/month calendar, room lanes, occupancy heatmap, waitlist side panel, booking drawer.

**Actions:** create booking, drag sessions, approve request, convert enquiry, move child between sessions.

**Dialogs:** conflict resolution, booking wizard, pricing summary.

### Staff Rotas

**Purpose:** align staffing shifts to child attendance and ratios.[cite:8][cite:14]

**Primary users:** manager, deputy, owner.

**Main components:** weekly rota grid, leave overlays, staffing gap warnings, publish controls.

**Actions:** assign shifts, copy schedule, approve leave, publish rota, rebalance rooms.

### Finance Dashboard

**Purpose:** manage revenue, debt, payment flows, and funding operations.[cite:10]

**Primary users:** finance/admin, owner, manager (read-limited).

**Main components:** revenue trend chart, unpaid invoices table, debt ageing card, payment status widgets, export status card.[cite:10]

**Actions:** run batch invoicing, resend invoices, export accounting data, apply funding adjustments.

### Invoices

**Purpose:** create, review, send, and reconcile invoices for childcare sessions and extras.[cite:10]

**Primary users:** finance/admin, owner.

**Main components:** invoice table, detail panel, payment history timeline, reminder workflow, credit note tools.

**Filters:** paid, unpaid, overdue, funding type, payment method, date range.

### Learning Journals and Observations

**Purpose:** record rich child development observations and share appropriate updates with parents.[cite:1][cite:8][cite:11]

**Primary users:** practitioner, room leader, manager, parent (read).

**Main components:** observation feed, photo cards, EYFS tags, draft autosave, approval/share controls.

**Actions:** add observation, attach media, tag learning area, publish to parent, export report.

### Messaging and Newsfeed

**Purpose:** centralise parent communication and internal updates inside one audited system.[cite:1][cite:11]

**Primary users:** staff and parents.

**Main components:** conversation list, message thread, announcement composer, event cards, read receipts.

**Actions:** send direct message, publish room update, post event, attach photo or file.

## Detailed screen layout

### Base layout

- **Header:** logo, nursery switcher, breadcrumb, global search / command palette, notifications, avatar menu.
- **Left sidebar:** grouped navigation for Home, Operations, Children, Staff, Finance, Learning, Reports, Settings.
- **Toolbar:** page title, contextual actions, search, view toggles, filter chips, saved views.
- **Content area:** filter rail, primary table or canvas, optional right-side detail drawer.
- **Table region:** sticky header, row selection, sorting, column settings, inline statuses.
- **Pagination:** page size, current range, next/previous controls.
- **Drawer / modal:** side sheet for details; modal for focused create/edit flows.
- **Footer:** support, system status, version.

This shell should remain consistent across modules because predictable layout reduces cognitive load for managers who use the system continuously throughout the day.[cite:1][cite:4][cite:10]

## CRUD patterns

Every major module should include these screens and flows:

### List page

- Dense, sortable table.
- Search plus advanced filters.
- Saved views.
- Bulk selection and actions.
- Export control.

### Create page

- Wizard for complex records such as child enrolment, staff onboarding, and booking creation.
- Short forms for incidents, quick messages, or simple configuration entities.

### Edit page

- Full detail form or section-based inline editing.
- Autosave for long content.
- Activity history and change log.

### Detail page

- Summary header.
- Tabs for related information.
- Timeline or audit feed.
- Related actions in contextual menu.

### Archive/delete flow

- Prefer archive over hard delete for operational records.
- Show consequences before confirmation.
- Offer short-term undo where safe.

### Bulk actions

- Assign room.
- Send communication.
- Export selected rows.
- Change status.
- Archive records.

### Import/export

- CSV import with downloadable template and field mapping.
- CSV/XLS export for operational lists.
- PDF invoice export and accounting integrations for finance workflows.[cite:10]

## Dashboard specifications

### Owner dashboard

- **KPIs:** occupancy, revenue MTD, debt, staffing cost ratio, enquiry conversion.[cite:4][cite:10]
- **Charts:** revenue trend, occupancy by room, aged debt, pipeline funnel.
- **Widgets:** top-performing rooms, low-occupancy warnings, unpaid invoice snapshot.
- **Quick actions:** run finance report, review forecasts, inspect debt.

### Manager dashboard

- **KPIs:** children present, rooms below ratio, absent staff, open incidents, pending booking requests.[cite:14]
- **Charts:** attendance trend, ratio incidents, staffing coverage.
- **Widgets:** live room cards, task list, activity feed, message alerts.
- **Quick actions:** open room register, approve requests, send announcement.

### Room Leader dashboard

- **KPIs:** room headcount, ratio state, observations completed, incidents open.
- **Widgets:** child roster, quick diary buttons, room notices.

### Finance dashboard

- **KPIs:** outstanding balance, paid today, overdue count, failed payments.[cite:10]
- **Charts:** invoice status distribution, debt ageing, payments over time.
- **Widgets:** batch invoicing, failed direct debit alerts, export shortcuts.

### Parent dashboard

- **KPIs:** outstanding balance, upcoming sessions, unread updates.[cite:1][cite:11]
- **Widgets:** latest diary entries, invoices snapshot, event reminders, quick booking request.

## Forms

### Child enrolment form

**Sections:**
- Child details
- Parent/legal contacts
- Emergency contacts
- Medical and dietary information
- Sessions and funding
- Authorized collectors
- Consents and documents
- Notes

**Validation:** required legal identity fields, emergency contact minimums, date validation, file type restrictions, duplicate record checks.

**Conditional logic:** SEN section appears when flagged; medication details appear when medication is required; funding fields appear when funded hours apply.

**Behavior:** autosave draft, save-and-exit, progress steps, validation summary.

### Staff form

**Sections:**
- Personal details
- Role and room assignment
- Qualifications
- Contract and pay settings
- Training and compliance documents
- Emergency contact

### Booking form

**Sections:**
- Child selection
- Session type
- Schedule
- Pricing and discounts
- Funding allocation
- Notes

### Observation form

**Sections:**
- Child
- Observation text
- Media
- EYFS tags
- Next steps
- Parent visibility

**Behavior:** autosave, draft recovery, mobile camera support.[cite:1][cite:11]

## Mobile experience

The following pages should be specifically optimized for tablets and phones because they support the most time-sensitive nursery-floor workflows:[cite:1][cite:2][cite:14]

- Attendance
- Check-in / sign-in screen
- Live room register
- Daily diary
- Pickup workflow
- Messaging
- Observation capture
- Incident logging
- Medication records

Mobile patterns should include large touch targets, swipe-friendly interactions, offline resilience for room use, sticky action bars, and simplified navigation.[cite:2][cite:14]

## UX patterns

| Pattern area | Recommended approach |
|--------------|----------------------|
| Large tables | Sticky headers, row selection, saved filters, column visibility, keyboard navigation |
| Scheduling | Multi-room week view with drag-drop sessions and occupancy overlays |
| Calendar | Google Calendar-style navigation with room/resource lanes |
| Booking | Step-based wizard with real-time availability and price summary |
| Invoice management | Stripe-like statuses, ageing tables, clear action buttons, batch processing |
| Timeline | GitHub-style chronological feed with icons and collapsible details |
| Activity feed | Notion-like compact cards with rich media and comments |
| Notifications | Grouped bell panel with unread states and direct deep links |
| Search | Global command palette plus scoped local search |
| Filtering | Token-based filters and saved views |
| Bulk actions | Contextual action bar with confirmation and undo |
| Undo | Snackbar-based undo for low-risk operations |
| Autosave | Draft preservation with “Saved just now” feedback |

These patterns support fast task switching and lower cognitive load in high-frequency SaaS workflows.[cite:10]

## Design system

### Visual language

- Calm enterprise aesthetic with approachable childcare warmth.
- Neutral backgrounds with blue or teal primary accents.
- Strong semantic colors for success, warning, danger, and info.
- Clear accessible contrast targets aligned to WCAG 2.2 AA.

### Foundations

- **Typography:** modern sans-serif with compact but readable scale.
- **Spacing:** 4px baseline system.
- **Grid:** 12-column desktop, adaptive tablet, single-column mobile.
- **Icons:** simple outlined icon family.

### Components

- Buttons: primary, secondary, tertiary, destructive.
- Inputs: text, select, multi-select, masked input, date, time, currency, tag input.
- Cards: KPI card, entity summary card, mobile record card.
- Tables: dense, sticky-headed, configurable.
- Badges: enrolment state, invoice state, ratio state, incident state.
- Tabs and accordions: for detail screens and dense information groups.
- Timelines: vertical event logs for audits and child history.
- Calendar widgets: booking and rota scheduling.
- Charts: line, bar, area, donut for finance, occupancy, and engagement.
- Empty states: short explanation with one primary CTA.
- Skeletons: cards, charts, rows, and profile placeholders.
- Dark mode: optional, with maintained accessibility.

## Page inventory

| Page Name | Module | User Roles | Priority | Complexity | Estimated Story Points |
|-----------|--------|------------|----------|------------|------------------------|
| Global Dashboard | Core | Owner, Manager, Deputy | High | High | 13 |
| Child List | Children | Manager, Deputy, Room Leader, Admin | High | Medium | 8 |
| Child Profile | Children | Staff, Parent (partial) | High | High | 13 |
| Parent List | Parents | Manager, Admin | Medium | Medium | 8 |
| Parent Profile | Parents | Manager, Admin | Medium | Medium | 8 |
| Live Room Register | Attendance | Room Leader, Practitioner, Manager | High | High | 13 |
| Daily Attendance Overview | Attendance | Manager, Deputy | High | Medium | 8 |
| Booking Calendar | Bookings | Manager, Deputy, Owner | High | High | 13 |
| Waitlist | Bookings | Manager, Admin | Medium | Medium | 8 |
| Enquiries | Admissions | Manager, Admin | Medium | Medium | 8 |
| Staff List | Staff & HR | Manager, Admin, Owner | Medium | Medium | 8 |
| Staff Profile | Staff & HR | Manager, Admin | Medium | Medium | 8 |
| Staff Rotas | Staff & HR | Manager, Deputy, Owner | High | High | 13 |
| Finance Dashboard | Finance | Owner, Finance/Admin | High | High | 13 |
| Invoices List | Finance | Finance/Admin, Owner | High | Medium | 8 |
| Invoice Detail | Finance | Finance/Admin, Owner | High | Medium | 8 |
| Batch Invoicing | Finance | Finance/Admin | High | High | 13 |
| Learning Journals | Learning | Practitioner, Room Leader, Manager | High | Medium | 8 |
| Observation Editor | Learning | Practitioner, Room Leader | High | Medium | 8 |
| Messaging & Newsfeed | Communication | Staff, Parent | High | High | 13 |
| Incident Log | Compliance | Manager, Deputy, Room Leader | Medium | Medium | 8 |
| Documents Library | Documents | Manager, Admin, Staff | Medium | Medium | 8 |
| Reports Overview | Reports | Owner, Manager, Finance/Admin | Medium | Medium | 8 |
| Settings: Nursery | Settings | Manager, Admin | Medium | Medium | 8 |
| Settings: Roles & Permissions | Settings | Super Admin, Owner | High | High | 13 |
| Parent Dashboard | Parent Portal | Parent | High | Medium | 8 |
| Parent Invoices | Parent Portal | Parent | High | Medium | 8 |
| Parent Messages | Parent Portal | Parent | High | Medium | 8 |
| Sign-in Screen | Attendance | Parent, Staff | High | Medium | 8 |

## Development order

1. Foundation: authentication, tenancy, roles, app shell, design system.
2. Children and parents core records.
3. Rooms and attendance with live register and ratio logic.[cite:14]
4. Booking calendar and occupancy planning.[cite:14]
5. Staff and rota planning.
6. Finance, invoices, funding, and exports.[cite:10]
7. Learning journals and observations.[cite:8]
8. Messaging and newsfeed.[cite:1][cite:11]
9. Compliance and safeguarding.
10. Reporting and analytics.
11. Parent portal.
12. Accessibility, performance, QA hardening.

This order reduces delivery risk by building dependencies in sequence while shipping operational value early.[cite:10][cite:14]

## UI inspiration

| Module | Primary inspiration |
|--------|---------------------|
| Global dashboard | Linear, Stripe, Vercel |
| Child and parent records | Notion, Google Drive |
| Attendance and room register | Famly, Google Calendar |
| Bookings and occupancy | Google Calendar, ClickUp, Monday.com |
| Staff rotas | Monday.com, ClickUp, Linear |
| Finance and invoices | Stripe Dashboard |
| Learning journals | Notion, Figma, Blossom Educational |
| Messaging and newsfeed | Famly-style feed plus Slack-like messaging |
| Reports and analytics | Stripe, Vercel |
| Parent portal | Consumer mobile SaaS patterns |

## Component inventory

- App shell
- Sidebar navigation
- Breadcrumbs
- Global command palette
- KPI cards
- Data table
- Filter bar
- Token filter input
- Entity drawer
- Modal dialog
- Multi-step wizard
- Calendar
- Scheduler grid
- Timeline
- Activity feed
- Status badge
- Avatar group
- Date range picker
- Uploader
- Rich text editor
- Notification center
- Chart components
- Empty state component
- Skeleton loader
- Sign-in kiosk panel
- Mobile bottom navigation
- Ratio alert widget
- Consent capture component
- Audit trail component

## Implementation notes

The product should feel operationally fast rather than visually decorative. The strongest benchmark signals from the nursery software market are real-time attendance, ratio awareness, occupancy management, parent engagement, and finance automation; these areas deserve the most interaction design attention and engineering investment.[cite:1][cite:8][cite:10][cite:14]
