# UK Nursery Management System (NMS) Roles

## 1. Owner

### Who are they?

The business owner who owns one or multiple nursery branches.

In this case, the owner operates **4 nursery sites**.

### Responsibilities

- Monitor overall business performance
- View revenue across all sites
- Monitor occupancy rates
- Check staff costs
- Review Ofsted readiness
- Review funding claims
- Manage nursery managers
- View reports and analytics

### Typical Questions

- Which nursery is most profitable?
- Which nursery has available capacity?
- How much funding are we receiving?
- Which manager is performing best?
- How many children are enrolled across all sites?

### System Permissions

✅ View all sites

✅ View financial reports

✅ View staff reports

✅ View child statistics

✅ Configure system settings

✅ Manage nursery managers

❌ Usually does not take daily attendance

---

## 2. Nursery Manager

### Who are they?

The person responsible for running a single nursery branch.

Example:

- Manager A → Site 1
- Manager B → Site 2
- Manager C → Site 3
- Manager D → Site 4

### Responsibilities

#### Admissions

- Handle enquiries
- Schedule tours
- Manage waiting list
- Approve registrations

#### Children

- Manage child records
- Verify medical information
- Assign rooms

#### Staff

- Manage schedules
- Approve leave
- Ensure legal staff-child ratios

#### Finance

- Review invoices
- Monitor overdue payments
- Verify funding hours

#### Compliance

- Manage accident reports
- Manage safeguarding concerns
- Prepare for Ofsted inspections

### Typical Questions

- Who is absent today?
- Are staff ratios compliant?
- Which invoices are unpaid?
- Which children receive funded hours?
- Any incidents today?

### System Permissions

✅ Full access to own nursery

✅ Manage children

✅ Manage staff

✅ Manage attendance

✅ Manage invoices

✅ View reports

❌ Cannot access other sites

❌ Cannot change platform settings

---

## 3. Nursery Staff / Practitioners

### Who are they?

Teachers, childcare practitioners, room leaders, and assistants.

They work directly with children every day.

### Responsibilities

#### Daily Care

- Check children in/out
- Record attendance
- Record meals
- Record naps
- Record toileting

#### Learning & Development

- Create observations
- Upload photos
- Track EYFS progress
- Record milestones

#### Health & Safety

- Record accidents
- Record medication
- Record incidents

#### Communication

- Send daily updates to parents

### Typical Questions

- Which children are in my room today?
- Any allergies?
- Did the child eat lunch?
- What observation should I record today?

### System Permissions

✅ View assigned children

✅ Take attendance

✅ Add observations

✅ Add photos

✅ Record incidents

✅ Send parent updates

❌ View finance

❌ Manage staff

❌ Access all nurseries

---

## 4. Parent

### Who are they?

Mother, father, guardian, or emergency contact.

Parents access the system through a mobile application with limited permissions.

### Responsibilities

#### Child Information

- View child's profile
- Update emergency contacts
- View medical information

#### Attendance

- View booked sessions
- Report absence

#### Learning Journey

- View observations
- View photos
- View progress

#### Finance

- View invoices
- Make payments
- View payment history

#### Communication

- Receive messages
- Chat with nursery
- Receive notifications

### Typical Questions

- Did my child eat today?
- Did my child sleep?
- Any accidents?
- How much is next month's invoice?
- What activities did my child do today?

### System Permissions

✅ View own child only

✅ View invoices

✅ View observations

✅ Send messages

✅ Receive notifications

❌ View other children

❌ View staff data

❌ View nursery reports

❌ View finance of nursery

---

# Recommended Hierarchy

```text
Platform
│
└── Owner
     │
     ├── Nursery Site A
     │     ├── Manager
     │     ├── Staff
     │     └── Parents
     │
     ├── Nursery Site B
     │     ├── Manager
     │     ├── Staff
     │     └── Parents
     │
     ├── Nursery Site C
     │     ├── Manager
     │     ├── Staff
     │     └── Parents
     │
     └── Nursery Site D
           ├── Manager
           ├── Staff
           └── Parents
```

---

# Access Matrix

| Context | Owner | Manager | Staff | Parent |
|----------|----------|----------|----------|----------|
| Admissions | View | Manage | View | Submit Registration |
| Child Records | View | Manage | Update | View Own Child |
| Attendance | View | Manage | Record | View |
| EYFS Learning Journey | View Reports | Manage | Create | View |
| Billing & Funding | View | Manage | No Access | View Own |
| Staff Management | View | Manage | Self Only | No Access |
| Messaging | View | Manage | Send | Send |
| Reporting | All Sites | Own Site | Limited | None |

---

# MVP Recommendation

The completed month-1 MVP implements three roles:

1. Nursery Manager
2. Nursery Staff / Practitioner
3. Parent

These roles cover the operational workflows delivered in the MVP baseline: manager administration and invoicing, practitioner attendance, and parent invoice access and payment.

**Owner is the first Post-MVP expansion role.** The first owner release will be oversight-first: cross-site visibility and administration, while branch managers retain routine operational write authority. See `docs/POST-MVP-ROADMAP.md` for the accepted expansion lane.

Additional roles can be introduced later if needed:

- Deputy Manager
- Room Leader
- Finance Administrator
- SENCO (Special Educational Needs Coordinator)
- Super Admin
- Compliance Officer

## Future Role: Super Admin

Super Admin is a platform-level NMS operator, not a nursery owner or branch manager. This role manages the SaaS control plane: tenant lifecycle, tenant status, feature flags, plan/billing metadata, platform health, support-safe account workflows, and cross-tenant audit visibility.

Super Admin access must preserve tenant isolation. Cross-tenant reads and support actions should be explicit, redacted where appropriate, and fully audited. Routine nursery workflows such as attendance, child edits, invoice issue, and safeguarding case work remain tenant roles unless a specific support workflow is later approved.

Backlog references:

- `API-PM-09` — super admin platform management API
- `FE-PM-10` — super admin dashboard for NMS platform management

## Terminology Note

Product-facing documentation prefers **site** for a nursery location. Current engineering, API, and database terms may use **branch** for the same location boundary. This distinction is cosmetic until a later decision separates the concepts.
