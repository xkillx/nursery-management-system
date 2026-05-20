# UK Nursery Management System (NMS) MVP PRD

## 1. Executive Summary

### Product vision and mission

The platform is intended to become the operational system of record for UK nurseries by unifying childcare operations, parent engagement, funding and invoicing, staffing and ratios, learning journeys, safeguarding, SEND, and inspection evidence into a single SaaS ERP. It is designed as a compliance-first, AI-ready, multi-tenant platform for independent nurseries, nursery groups, franchises, and enterprise childcare operators.

The mission is to reduce administrative and compliance burden so leaders and practitioners can focus more time on children while maintaining funding accuracy, regulatory confidence, and a strong parent experience. The product must optimise for fast MVP delivery without compromising tenant isolation, security controls, auditability, and architecture quality because those foundations directly affect compliance trust and enterprise readiness.

### Problem statement

UK nurseries operate under a difficult mix of paper workflows, fragmented software, spreadsheet-based funding calculations, inconsistent attendance evidence, and weak safeguarding records. This creates operational drag, revenue leakage, higher inspection stress, and avoidable compliance exposure.

Childcare funding complexity has increased as funded entitlements expanded beyond the historic 3-to-4-year-old cohorts to younger children in staged rollouts from 2024 onward. Providers now need to manage eligibility codes, reconfirmation cycles, stretched offers, session structures, consumables, local authority variations, and invoice transparency at a level many current systems do not handle well.

### Why current NMS products fail

Many incumbent nursery systems succeed at journals, parent updates, or basic registers, but fail to solve the economically and operationally critical problems of funding accuracy, ratio evidence, safeguarding structure, and group-level visibility. Common weaknesses include shallow funding logic, poor room-staff UX, weak offline support, fragmented safeguarding notes, and limited auditability for inspection readiness.

A second failure mode is architectural: products often evolve feature-by-feature without strong domain boundaries, leaving finance, attendance, parent communication, and compliance workflows tightly coupled and difficult to scale or audit. This is especially problematic for multi-branch operators that need standardised policies, reporting, and delegated governance.

### Competitive opportunity and strategic differentiation

The competitive opportunity is to build a UK-specific nursery ERP that goes deeper than incumbents on high-pain, high-risk workflows: funding, invoicing, attendance, ratios, safeguarding, SEND, and inspection readiness. Strategic differentiation should centre on explainable funding calculations, real-time attendance-driven ratio monitoring, structured safeguarding records, mobile-first operational UX, and enterprise-capable tenancy and reporting.

The strongest product position is not “another nursery app,” but a compliance-grade operational backbone for childcare providers. The market moat comes from operational intelligence built from clean, standardised childcare data over time, especially where attendance, finance, staffing, safeguarding, and parent engagement intersect.

### MVP philosophy, AI strategy, and multi-tenant strategy

The MVP should focus on the workflows where nurseries lose the most time or money: attendance, ratios, funding allocation, invoice generation, child and parent records, parent communication, and baseline compliance evidence. Each MVP slice should be vertically complete, meaning each domain release includes data model, APIs, web UI, mobile UI, events, audit logging, authorization rules, reporting hooks, and QA coverage.

AI should be introduced as an assistive layer rather than an autonomous decision-maker, especially in compliance-sensitive domains. Early AI should support drafting, tagging, summarisation, anomaly hints, and forecasting, while safeguarding and finance decisions remain human-owned and fully auditable.

The recommended multi-tenant model is a shared SaaS control plane with row-level tenant isolation using tenant and branch IDs across records, enforced with PostgreSQL row-level security and application-level authorization. Enterprise customers should later be able to move to dedicated database or cluster isolation without requiring a different logical product model.

## 2. Industry and Domain Analysis

### UK nursery operations in practice

A typical nursery operates through room-based care structures organised by age and developmental stage, often with separate baby, toddler, and preschool rooms and flexible movement throughout the day. Daily operations include opening checks, staff attendance, child sign-in, room balancing, meal and sleep tracking, nappy/toileting records, incident logging, parent handovers, and end-of-day reconciliation.

Managers and deputies spend significant time on exceptions rather than ideal flows: late pickups, ad hoc sessions, absences, room moves, staffing cover, funding questions, parent concerns, medication issues, and document chasing. A useful NMS must therefore handle interruptions, partial information, and last-minute operational adjustments rather than assuming perfect schedules.

### Room structures, staffing hierarchy, and occupancy management

Room structures exist not just for pedagogy but for compliance, staffing, and economics. The number of children present by age band, the number and qualification level of staff available, and the booked session patterns all affect whether a setting can safely and profitably operate each room.

Typical hierarchy includes owner or franchise operator, area manager for groups, nursery manager, deputy manager, room leader, practitioners, bank staff, finance/admin staff, HR/admin staff, and designated safeguarding lead roles. In smaller settings, several of these responsibilities collapse into one or two people, which increases the importance of task prioritisation, delegation visibility, and role-appropriate interfaces.

Occupancy management is the central commercial lever of the nursery business. Nurseries make or lose money based on how effectively they fill sessions, balance age-band demand, align funded and paid hours, avoid underutilised rooms, and prevent staffing inefficiencies caused by poor attendance visibility.

### Funding mechanisms and where money is made or lost

Government-funded childcare is a major operational and financial driver for English nurseries, including 15 hours for all 3- and 4-year-olds and expanded entitlements for eligible working families and younger children in phased rollouts. Providers must model entitlement use accurately against booked sessions, actual attendance, stretched delivery patterns, term structures, and local authority administration rules.

Money is commonly lost through under-claiming funded hours, over-generous manual adjustments, incorrect invoice line construction, failure to reconcile bookings and attendance, and poor handling of consumables, extras, and sibling discounts. Money is also lost through administrative waste: managers and finance staff spending hours each month reconstructing what happened rather than relying on a trustworthy system of record.

### Parent expectations and operational leverage points

Parents expect mobile-first access to communication, diaries, invoices, payment status, and practical updates about their child. They also expect transparency on fees and funded hours, prompt responses, secure document handling, and confidence that the nursery is safe, professional, and well organised.

Software creates leverage when it reduces double-entry, explains funding clearly, makes attendance and room operations visible in real time, and reduces inspection panic by preserving structured evidence continuously rather than collecting it retrospectively. Competitors are weakest where daily operational complexity meets finance and compliance, which is exactly where the MVP should concentrate.

### Compliance pain points, safeguarding, and multi-site complexity

Safeguarding is both a daily responsibility and a documentation burden, requiring timely concern recording, chronology, access control, escalation, and evidence of effective action. Where systems are weak, staff often resort to paper or separate files because they do not trust the product to handle confidentiality or nuance properly.

Inspection preparation is another major pain point because evidence is often scattered across registers, policies, training records, incident books, staff files, and ad hoc documents. Software advantage comes from producing inspection-ready views from operational data that has already been captured as part of daily work.

Multi-site operators face additional problems around standardisation, branch governance, benchmarking, role hierarchy, and cross-site reporting. A product that solves only single-site nursery administration will struggle to become the platform of choice for groups, franchises, or childcare organizations pursuing central oversight.

## 3. Regulatory and Compliance Deep Dive

### EYFS statutory framework

The EYFS statutory framework sets legal standards for learning, development, safeguarding, welfare, staffing, and operational practice for early years providers in England. Its operational impact is direct: room organisation, ratios, qualifications, supervision, key-person practice, training, and welfare processes all need to align to the framework.

Required system features include configurable room and age-band policies, ratio calculation rules, staff qualification records, welfare workflows, daily care capture, and evidence generation for training and supervision compliance. Data implications include storing child age bands, room occupancy, staff roles, qualifications, attendance events, and policy acknowledgements in a queryable and auditable manner.

Security implications include restricting access to sensitive child information while keeping operational data sufficiently available to authorised staff for safe care delivery. UX implications include the need for fast room-level interfaces because compliance depends on accurate capture during busy operational periods rather than perfect end-of-day reconciliation.

Edge cases include mixed-age rooms, temporarily reassigned staff, children moving between rooms during the day, and situations where settings use permitted flexibilities in ratios or room practice. Risk of non-compliance includes enforcement action, safeguarding concerns, weaker Ofsted outcomes, and increased liability exposure.

### Ofsted expectations

Ofsted inspection expectations extend beyond static policy possession and focus heavily on leadership, safeguarding culture, quality of education, behaviour, personal development, and evidence of safe and effective practice. Operationally this means the NMS must support inspection readiness continuously rather than only through ad hoc export features.

Required features include inspection dashboards, evidence packs, staff suitability and training views, incident and safeguarding chronology, attendance and ratio evidence, complaints logs, and exportable reports aligned with real inspection conversations. Audit requirements are high because inspection confidence depends on whether the system can demonstrate who recorded what, when, and under what authority.

Edge cases include surprise inspections, disputed records, retrospective corrections, and evidence requests spanning multiple branches or time periods. Risk of non-compliance includes poor inspection grades, reputational damage, occupancy decline, and potential escalation to regulators if serious weaknesses are uncovered.

### Safeguarding obligations and serious incident handling

Safeguarding requirements require providers to maintain effective child protection arrangements, safer recruitment, staff training, concern escalation, and appropriate referral pathways. Operationally, safeguarding data must not behave like ordinary operational notes because access, chronology, and evidence quality are materially more important.

Required system features include concern capture, case creation, restricted timelines, action tracking, escalation workflows, related-party linking, attachment control, and case outcome tracking. Data implications include highly sensitive notes, external agency references, linked incidents, staff allegations, and retention-controlled chronology records.

Security implications are severe: safeguarding data requires additional access restrictions, stronger review controls, exhaustive access logging, and careful export logic. UX implications include secure but fast concern submission, DSL triage interfaces, and chronology views that reduce cognitive load during stressful review or inspection scenarios.

Edge cases include siblings across settings, allegations involving staff, external referrals, requests for redaction, and legal subject-access complexities. Risk of non-compliance is critical because failures can result in harm to children, adverse inspection findings, legal exposure, and severe reputational loss.

### GDPR, retention, and audit logging

Nurseries process large quantities of personal and special category data, including health, SEND, and safeguarding records, which creates significant GDPR and UK data protection obligations. Operational impact includes lawful data collection, access control, retention handling, deletion governance, subject access support, and secure document and media management.

Required features include soft delete, retention policy enforcement, hard-delete workflows where legally permitted, export tooling for data access requests, access review support, and immutable audit records for critical changes. Data implications include retention markers, lawful basis metadata where appropriate, deletion states, and provenance logs.

Security implications include encryption in transit and at rest, secrets management, secure backups, restricted admin access, and documented incident response. UX implications include clear administrative controls for archive, restore, export, and restricted deletion, while avoiding accidental destructive actions.

Risk of non-compliance includes data breaches, enforcement action, legal disputes, and loss of trust among parents and enterprise customers. Audit logging is especially important because regulated customers will expect detailed evidence trails for high-risk domains such as finance, safeguarding, staff suitability, and permissions changes.

### Ratio requirements, DBS, SEND, and funding compliance

Ratio requirements affect scheduling, room occupancy, staff deployment, and escalation workflows in real time, not just as a reporting concern. The system therefore needs live room status, staff presence data, age-aware occupancy logic, and historical snapshots for later review.

DBS and staff suitability obligations require clear records of checks, qualifications, training, and review status. SEND obligations require structured documentation, support plans, reviews, and secure collaboration records where settings support children with additional needs.

Funding compliance requires the product to support entitlement logic, invoice clarity, booking-to-claim alignment, and rule explainability because both parents and local authorities may challenge outcomes. The product should treat funding as a governed domain with explainable decisions rather than a collection of configurable invoice discounts.

### Compliance matrix

| Regulation area | Operational impact | Required system capability | Data implications | Audit and risk implications |
|---|---|---|---|---|
| EYFS welfare and staffing | Room staffing, supervision, daily care operations | Attendance, room config, ratio engine, welfare logs | Child ages, staff presence, room assignments | Breach evidence, daily snapshots, corrective action logs. |
| Ofsted inspection expectations | Continuous readiness, evidence availability | Inspection dashboards, exports, suitability and incident views | Structured operational and compliance records | Weak audit trails undermine inspection confidence. |
| Safeguarding | Concern logging, DSL review, escalation | Restricted case management and chronology | Sensitive notes, referrals, linked persons | Critical access logging and non-repudiation needed. |
| GDPR and retention | Data lifecycle, access, deletion, export | Retention engine, SAR tools, archive and purge | Retention metadata, soft-delete states | Breach and over-retention risk if unmanaged. |
| Funding compliance | Claims, invoices, parent transparency | Rule engine, explainable allocations, funding reports | Eligibility codes, booking patterns, term logic | Margin loss and clawback risk if calculations are wrong. |
| Staff suitability / DBS | Recruitment and ongoing compliance | Staff records, alerts, checklist views | DBS and training status | Missed expiry or incomplete checks create inspection risk. |

### Compliance risk table

| Risk | Likelihood | Impact | Primary mitigation |
|---|---|---|---|
| Incorrect ratio calculations | Medium | High | Deterministic rules, event-driven snapshots, override logging, strong test fixtures. |
| Funding rule defects | High | High | Versioned rule engine, gold-standard scenarios, controlled rollout by feature flag. |
| Safeguarding data leakage | Low | Critical | Restricted schemas, ABAC, MFA, encryption, exhaustive audit logs. |
| Weak inspection evidence | Medium | High | Evidence-first reporting and continuous compliance dashboards. |
| Retention misconfiguration | Medium | Medium-High | Policy engine, legal review, scheduled archival and purge jobs. |

## 4. Competitor Reverse Engineering

### Market structure and competitor clustering

The UK nursery software market contains products that tend to cluster into journal-first, admin-first, or broader nursery operations suites, with varying strengths around parent engagement, learning documentation, and finance. Public market signals and internal roadmap findings suggest the most consistent weaknesses are in funding complexity, mobile operational UX, and compliance depth rather than in basic messaging or diaries.

### Competitor patterns by segment

| Competitor pattern | Typical strengths | Typical weaknesses | Implication for product strategy |
|---|---|---|---|
| Journal-first products | Observations, EYFS tagging, parent-facing diaries | Shallow funding and workforce depth | Differentiate on finance, operations, and compliance. |
| Admin-first legacy suites | Registers, child records, some billing | Dated UX, limited mobile speed, fragmented modules | Win on usability and architectural coherence. |
| Emerging all-in-one apps | Parent app polish, fast onboarding | Limited enterprise controls, compliance depth, reporting maturity | Win on trust, auditability, and group readiness. |

### Reverse-engineered competitor themes

Across named competitors such as Famly, eyworks, Tapestry, Ovivio, Connect Childcare, Parenta, Kinderly, and Cheqdin, the most important reverse-engineered distinction is not feature checklist breadth but which workflows appear to be treated as primary products versus adjacent add-ons. Vendors with strong parent and journal experiences often appear weaker in rule-heavy funding and workforce operations, while more admin-oriented systems often carry UX debt and fragmented module experiences.

The likely data and architecture pattern in many products is historical module accumulation rather than clean domain separation, which tends to create uneven UX, brittle cross-module workflows, and limited explainability in financial calculations. By contrast, the recommended product should treat identity, child/family, attendance, finance, safeguarding, and reporting as bounded contexts from the start.

### Market gap analysis and opportunities to disrupt

The clearest market gap is a platform that combines parent-friendly experiences with true back-office and compliance strength. Most products offer some subset of registers, journals, invoices, or messaging, but fewer appear to solve the end-to-end chain from child booking to attendance to ratio evidence to funding calculation to parent invoice to management reporting.

The best disruption opportunities are:

- Explainable funding engine with local authority-aware rule packs.
- Live ratio compliance with room-aware dashboards and historical evidence.
- Structured safeguarding and SEND records with access controls trusted by DSLs and managers.
- Multi-branch analytics and delegated governance usable by groups and franchises.
- AI assistance embedded into operational workflows rather than bolted on as a generic chat feature.

### Feature moat analysis

Funding, compliance evidence, and operational intelligence form the strongest moat because they are hard to retrofit into products designed first for journals or basic administration. Over time, cross-tenant benchmarking, anomaly detection, and forecasting can become additional defensible layers once enough clean and permissioned data exists.

## 5. Personas and User Journey Mapping

### Persona summaries

| Persona | Goals | Frustrations | Key KPIs | Primary devices |
|---|---|---|---|---|
| Owner | Margin, occupancy, compliance confidence, growth | Poor branch visibility, invoice leakage, inspection risk | Occupancy, revenue, debt, inspection outcomes | Laptop, tablet |
| Area Manager | Cross-site control and consistency | Inconsistent branch processes, fragmented reporting | Branch performance, staffing health, incident trends | Laptop, mobile |
| Nursery Manager | Smooth daily operation and parent trust | Firefighting, paperwork, slow systems | Ratios met, incidents, parent satisfaction | Desktop, tablet |
| Deputy Manager | Operational continuity | Last-minute staffing changes, missing data | Register accuracy, rota stability, room readiness | Tablet, mobile |
| Room Leader | Safe room operation, staff coordination | Slow capture flows, poor visibility | Attendance accuracy, observations, incidents | Tablet |
| Practitioner | Deliver care and record quickly | Too many clicks, duplicate entry | Timely records, observation quality | Mobile, tablet |
| Parent | Trust, clarity, convenience | Unclear fees, delayed updates, clunky apps | Payment timeliness, engagement, response time | Mobile |
| Finance Admin | Accurate invoicing and collections | Manual fixes, spreadsheet dependency | Invoice error rate, debt, reconciliation speed | Desktop |
| HR Admin | Staff suitability and records | Expiry chasing, fragmented staff data | DBS compliance, training completion | Desktop |
| DSL | Safeguarding vigilance and chronology | Mixed confidentiality, weak case tools | Review timeliness, action completion | Secure laptop |
| Franchise Operator | Consistency, governance, scalability | Branch variation, poor oversight | Standardisation, branch margin, inspection spread | Laptop |

### Persona detail and permissions

Owners and area managers need cross-branch reporting, branch benchmarking, policy and finance visibility, and limited ability to intervene without drowning in local operational detail. Nursery managers and deputies need control panels for day-of operations, staffing, parent issues, incidents, and funding exceptions.

Room leaders and practitioners require fast mobile or tablet flows, often under poor connectivity, with minimal navigation and careful role restrictions around child data, medicine, incidents, and room attendance. Finance admins need strong review and override workflows in billing domains, while DSLs require sharply restricted, chronology-first views over safeguarding concerns and follow-up actions.

### End-to-end user journeys

#### Enrollment journey

1. Parent submits enquiry.
2. Nursery records prospect and books visit.
3. Parent completes digital registration with child, guardian, consent, medical, and funding information.
4. System creates child and guardian records, sends onboarding communications, and generates a provisional funding estimate.
5. Manager confirms room placement and key person assignment.
6. Finance/admin verifies entitlement details and booking pattern.

Pain points include duplicate data entry across forms, funding ambiguity before enrollment, delayed document collection, and lack of visibility into incomplete registrations. Automation opportunities include guided digital forms, pre-validation, document reminders, and instant estimate generation.

#### Attendance and ratio journey

1. Staff reviews expected children and room headcount.
2. Guardians sign children in; staff record arrivals and room movements.
3. Staff presence is updated, including breaks and cover.
4. Ratio engine recalculates room safety status continuously.
5. Alerts prompt intervention before or during breaches.
6. Emergency mode provides live on-site roll call.

Pain points include room transfers, late arrivals, ad hoc bookings, and offline connectivity. Automation opportunities include projected ratio warnings, discrepancy alerts, and end-of-day reconciliation summaries.

#### Invoicing journey

1. Funding engine calculates entitlement usage and paid hours.
2. Invoice engine creates draft invoices with line-item explainability.
3. Finance admin reviews exceptions, credits, and adjustments.
4. Invoices are approved and issued to guardians.
5. Parent sees invoice and payment options in the app.
6. Payment events reconcile invoices and debt status.

Pain points include mid-cycle schedule changes, temporary eligibility changes, invoice disputes, and credit note handling. Automation opportunities include anomaly detection, exception queues, reminder schedules, and reusable correction workflows.

#### Safeguarding, inspection, staffing, and SEND journeys

Safeguarding journeys begin with rapid concern capture and continue into restricted chronology, actions, review, referral, and evidence production. Inspection journeys centre on assembling confidence from already-captured operational data rather than performing emergency manual preparation.

Staffing journeys in MVP should focus on the data needed for ratios and compliance, such as staff presence, role, qualification, DBS status, and basic availability. SEND journeys in MVP should support identification, basic documentation, and linkage to child and learning records, while deeper graduated response workflows can follow later.

## 6. MVP Scope Definition

### Exact MVP boundaries

The MVP should include the following core modules and capabilities:

- IAM and tenant management.
- Child and guardian records with consent and medical basics.
- Parent portal basics, including diary, messages, invoices, and absence reporting.
- Attendance capture with room views, corrections, emergency roll call, and auditability.
- Ratio engine with live status and alerts.
- Funding engine version 1 with explainable entitlement calculations.
- Invoice generation and basic payment integration.
- Messaging and notification infrastructure.
- Minimal learning journal capability.
- Basic incident and safeguarding logging.
- Baseline reporting and inspection pack support.

### Explicitly out of scope for MVP

The MVP should not include deep rota optimisation, payroll, advanced HR, full SEND casework, advanced safeguarding case management, sophisticated enterprise controls, or broad custom workflow builders. These areas are important but should follow stable core data models, event streams, and pilot validation of the operational heart of the product.

### MoSCoW prioritisation

| Priority | Included capabilities |
|---|---|
| Must have | IAM, tenant model, child/guardian domain, attendance, ratio engine, funding v1, invoicing, payments basics, parent portal basics, messaging, minimal incidents and safeguarding, baseline reporting. |
| Should have | Emergency roll call, offline attendance support, inspection pack exports, basic learning timeline, core audit dashboards. |
| Could have | Light AI drafting, limited occupancy forecasting, simple rota visibility, SMS support. |
| Won't have in MVP | Payroll, advanced HR, deep SEND, advanced safeguarding case management, franchise governance suite, broad analytics warehouse, advanced optimization engines. |

### RICE scoring and dependency mapping

A representative prioritisation shows attendance/ratios and funding/invoicing as the highest impact MVP capabilities because they jointly affect compliance, revenue, and daily operational trust. Child and parent domain must land before billing and communication, while IAM and tenancy underpin every other module.

Dependency sequence should be:

1. Identity and tenant core.
2. Child and parent core.
3. Attendance core.
4. Ratio engine.
5. Funding engine.
6. Invoice engine and payment integration.
7. Parent portal and messaging.
8. Incident and safeguarding basics.
9. Reporting and inspection pack.

### MVP success criteria

Success should be measured by live pilot adoption, billing trust, attendance/ratio reliability, and operational replacement of paper or spreadsheets in core flows. A meaningful MVP milestone is when pilot nurseries can run attendance, ratio monitoring, funding-aware invoicing, and parent communication without parallel shadow systems after a controlled transition period.

## 7. Detailed Module Breakdown

### 1. IAM and Multi-Tenant

**Business purpose:** establish secure identity, invitations, role assignment, tenant and branch scoping, session control, and MFA.

**Functional requirements:** login, logout, token refresh, invite user, accept invite, reset password, assign role, enable MFA, review sessions, suspend user, and manage tenant/branch membership.

**Non-functional requirements:** strong password handling, rate limiting, resilient session store, exhaustive auth audits, support for multiple-tenancy membership, and future SSO readiness.

**Data model:** users, roles, permissions, memberships, sessions, MFA methods, invitation tokens, password reset tokens, auth events.

**Events:** `UserInvited`, `UserActivated`, `RoleChanged`, `SessionRevoked`, `TenantCreated`.

**Edge cases:** one user across multiple tenants, temporary branch-level access, dormant invites, forced logout after privilege change.

### 2. Child Management

**Business purpose:** maintain the source of truth for child identity, relationships, consents, medical information, dietary needs, and enrollment status.

**Functional requirements:** create child, edit child, link guardians, manage relationships and consents, upload documents, assign key person, manage enrollment states.

**Permissions:** branch-scoped access for general staff; tighter access for sensitive notes and medical details.

**Edge cases:** siblings, shared guardianship, foster placements, multiple emergency contacts, legal name changes.

### 3. Parent Portal

**Business purpose:** provide a single secure interface for parents to view daily updates, invoices, messages, forms, and absences.

**Functional requirements:** message view, invoice view, payment initiation, absence notification, forms and signatures, diary timeline, media access.

**UX considerations:** fast mobile performance, low-friction navigation, signed media URLs, and clear billing explanations.

### 4. Attendance

**Business purpose:** capture accurate child and staff attendance, room movements, and emergency occupancy state.

**Functional requirements:** sign in/out, room move, corrections, absence reason handling, emergency roll call, discrepancy review.

**Validations:** no duplicate active sign-in, room must belong to branch, correction requires reason and actor, time ordering rules for replayed events.

**Events:** `AttendanceRecorded`, `AttendanceCorrected`, `RoomMoveRecorded`.

**Offline considerations:** append local events on device, sync later, preserve original local timestamp and device ID for audit.

### 5. Ratio Engine

**Business purpose:** convert attendance and staffing data into real-time compliance and operational decision support.

**Functional requirements:** room-level ratio status, projected breaches, historical snapshots, alerts, configurable rules by room and age band.

**Scalability concerns:** event bursts at session boundaries, repeated recalculation, need for low latency during operational peaks.

**Compliance concerns:** outputs must be explainable and preserved historically because ratio evidence is not useful if it only reflects current state.

### 6. Funding Engine

**Business purpose:** calculate entitlement usage and fee outcomes accurately and explainably.

**Functional requirements:** support entitlements, stretched offers, funded versus paid hours, consumables and extras, sibling discounts, term structures, and local authority variation over time.

**Data model:** eligibility records, rule sets, booking plans, attendance summaries, funding allocations, calculation versions.

**Edge cases:** age transitions, reconfirmation failures, mid-period changes, split-family billing, local authority rule updates.

**Testing:** gold-standard real-world style scenarios with deterministic expected outputs.

### 7. Invoicing and Payments

**Business purpose:** turn bookings and funding outcomes into trustworthy invoices and payment records.

**Functional requirements:** draft invoices, preview, approval, issue, adjustments, credits, allocations, webhook reconciliation, debt tracking.

**Data model:** invoices, invoice lines, credit notes, payment attempts, payment allocations, debt actions.

**Compliance and security:** invoice immutability after issue, explicit adjustment flows, tokenised payment provider usage, cryptographic webhook validation.

### 8. Messaging and Notifications

**Business purpose:** deliver direct messages, announcements, and event-triggered notifications across channels.

**Functional requirements:** direct message threads, group messages, newsletters, delivery preferences, push/email notifications, retries, templates.

**Scalability concerns:** fanout, delivery retries, SMS or push cost control, tenant-level branding and throttling.

### 9. Learning Journeys

**Business purpose:** provide lightweight observation and journal functionality sufficient for MVP parent value and early learning evidence.

**Functional requirements:** record observation, upload media, EYFS tag, review, and optionally share to parent timeline.

**Future extensibility:** deeper curriculum planning and advanced assessment should remain post-MVP.

### 10. Safeguarding

**Business purpose:** support restricted concern capture and basic case progression with DSL oversight.

**Functional requirements:** log concern, create case, append chronology entry, assign follow-up action, attach restricted files, record referral and outcome.

**Audit requirements:** exhaustive access logs, append-only chronology, no silent edits, redaction only through explicit supervised workflow.

### 11. Incident and Medication

**Business purpose:** record accidents, body map information, medicine administration, and parental acknowledgement.

**Edge cases:** retrospective entries, refusal to sign, medicines administered by different staff than planned, linked incident and safeguarding escalation.

### 12. SEND Management

**Business purpose:** provide a lightweight foundation for SEND identification and support planning in MVP.

**Functional requirements:** SEND profile, support note, meeting note, outcome tracking, and links to child records and observations.

### 13. Staff and HR

**Business purpose:** hold the minimum staff data needed for ratios, suitability, and compliance workflows in MVP.

**Functional requirements:** staff record, role, qualification, DBS status, training status, basic availability and attendance.

### 14. Reporting and Analytics

**Business purpose:** convert operational data into dashboards, reports, and exports for managers, owners, and inspectors.

**Functional requirements:** attendance reports, ratio reports, invoice summaries, debt views, safeguarding counts, inspection pack exports.

**Architecture implications:** use read models and replicas before moving to a full warehouse.

### 15. AI Assistant

**Business purpose:** provide safe, assistive productivity improvements in writing, tagging, forecasting, and summarisation.

**Governance:** prompt versioning, human approval, output logging, AI feature flags, provenance-aware grounding.

### 16. Admin and Configuration

**Business purpose:** allow tenants to configure branch structures, fee plans, age bands, term calendars, base funding settings, branding, and feature flags.

**UX considerations:** guided setup wizard with templates for common nursery structures.

## 8. Domain-Driven Design (DDD)

### Bounded contexts

The recommended bounded contexts are Identity and Access, Organization, Child and Family, Attendance and Ratios, Staff and HR, Learning and EYFS, SEND, Safeguarding, Incidents and Health, Finance, Communication, Reporting and Analytics, and AI.

This separation aligns to real ownership, data sensitivity, transactional cohesion, and future service extraction boundaries. Finance should own invoices and funding allocations, Safeguarding should own restricted chronology and cases, and Attendance should own append-only registers and room state rather than leaking these concerns across modules.

### Aggregates, entities, and value objects

Representative aggregates include Child, Invoice, Attendance Session or Event Stream, Safeguarding Case, Observation, and Staff Member. Representative entities include guardian, consent, medical profile, invoice line, payment allocation, case event, room, and qualification.

Value objects should include money, entitlement period, attendance interval, room capacity policy, notification preference, legal relationship, and audit actor context. These help constrain logic and reduce leakage of weakly modelled primitives across domains.

### Domain events and repositories

Core events should include `ChildRegistered`, `AttendanceRecorded`, `RatioBreached`, `FundingCalculated`, `InvoiceIssued`, `ObservationPublished`, `SafeguardingConcernRaised`, and `DBSExpiringSoon`. Repositories should remain context-owned and should not be bypassed by cross-domain table coupling.

### Context map, anti-corruption layers, and extraction opportunities

A context map should define upstream/downstream relationships explicitly: Child is upstream to Attendance, Finance, Learning, SEND, and Safeguarding; Attendance is upstream to Reporting and partially to Finance; Finance is upstream to Parent Portal billing views; Safeguarding is intentionally isolated but may publish limited redacted events.

Anti-corruption layers are especially important around auth providers, payment providers, AI models, and future external nursery integrations. Later extraction opportunities are strongest for Notifications, Reporting, Finance, AI orchestration, and possibly Safeguarding if regulatory and operational isolation needs increase.

## 9. System Architecture

### Modular monolith architecture

The recommended architecture is a modular monolith first, using strong domain boundaries within a single deployable backend, a shared relational database, and event-driven internal workflows. This approach balances speed of iteration, transactional simplicity, lower operational overhead, and future extraction potential.

A pure microservice architecture is not recommended for the MVP because it would add deployment complexity, distributed transaction challenges, higher testing cost, and debugging overhead before product-market fit is established. A plain undisciplined monolith is equally risky because it would create long-term coupling across finance, safeguarding, and attendance domains.

### Event-driven internals and service evolution

The product should use a transactional outbox pattern from day one so that domain state and domain events are committed atomically before being published to an internal broker such as NATS. This supports reliable event propagation to notifications, reporting, AI indexing, and future extracted services without lost-event problems.

Planned evolution path is:

- 0–18 months: modular monolith.
- 18–36 months: modular services where justified.
- Later: hybrid services for enterprise scale, extracting high-load or high-risk domains selectively.

### C4-oriented architecture view

At a system level, the platform serves parents, practitioners, room leaders, managers, owners, finance admins, HR admins, and DSLs through web and mobile interfaces. At the container level, the stack includes Angular web frontend, Flutter mobile apps, Go backend services, PostgreSQL, Redis, NATS, object storage, search, observability stack, and AI orchestration layer.

At the component level, each bounded context should expose APIs, command handlers, query views, event publishers, and domain services with explicit ownership boundaries. This container and component clarity will make later service extraction more intentional and less disruptive.

### Multi-tenant isolation, authorization, and realtime

The recommended multi-tenant strategy is shared infrastructure with row-level data isolation plus branch scoping and application-level checks. Authorization should combine RBAC for baseline role grants with ABAC for branch scope, child assignment, employment context, case sensitivity, and other dynamic constraints.

Realtime requirements include live room attendance, ratio alerts, message updates, and notification badges. WebSockets or SSE can support dashboards and messaging, while Redis or broker-backed fanout can support low-latency distribution.

### Mobile sync, caching, jobs, and media architecture

Staff mobile workflows should support offline buffering for attendance and observation capture, then replay with conflict-aware reconciliation when connectivity returns. Parent apps should favour fast load times, reliable notification delivery, and secure media access through short-lived signed URLs.

Redis should be used for short-lived caches, room state acceleration, and token/session support where appropriate, while PostgreSQL remains the source of truth. Background jobs should handle invoice generation, notification delivery, attendance reconciliation, retention, indexing, and reporting workloads without competing with the main API plane.

### Observability, scaling bottlenecks, and cost implications

Observability should include metrics, logs, traces, queue backlog indicators, invoice batch health, auth anomalies, and business-health alerts rather than only infrastructure metrics. Likely bottlenecks include reporting queries, ratio recalculation spikes, media uploads, invoice generation, and AI latency if AI is allowed onto hot user paths.

Cost will remain manageable if the product avoids premature platform complexity, but messaging, file storage, observability, CI, and AI calls can significantly affect per-tenant economics as usage scales. SMS and media-heavy parent usage are especially important unit-economics watchpoints.

## 10. Database and Data Architecture

### PostgreSQL schema and tenancy strategy

PostgreSQL should serve as the primary OLTP store, with tenant-aware shared schemas or context-owned schemas and strong isolation through tenant and branch identifiers on records. Every tenant-owned table should include at least `tenant_id`, and branch-scoped tables should also include `branch_id`, along with lifecycle fields such as `created_at`, `updated_at`, and `deleted_at`.

### Partitioning, indexing, and reporting read models

High-volume tables such as attendance, notifications, observations, and audit logs should use partitioning strategies appropriate to access patterns, most likely monthly partitions with tenant-aware indexing. Reporting should rely on read models, replicas, or derived tables so operational queries do not degrade primary OLTP performance.

Indexing should prioritize compound keys around tenant and branch scope, time-series access, parent/child lookup patterns, and financial batch generation. Search-heavy areas can later add dedicated search infrastructure, but PostgreSQL plus careful indexing is sufficient for MVP and early growth.

### Audit architecture, retention, soft delete, and archival

A central audit architecture should capture actor, action, timestamp, target, before/after metadata, branch/tenant context, and request correlation identifiers for sensitive operations. Soft delete is appropriate for user-facing removal flows, while hard delete should be delayed and policy-driven to respect retention obligations and legal exceptions.

Archival strategy should distinguish between operational access and long-term retention, especially for safeguarding and finance. Immutable records are especially important for invoices, attendance evidence, safeguarding chronology, and certain staff suitability records.

### GDPR, safeguarding data, and analytics implications

GDPR implications include lawful retention, access rights handling, deletion boundaries, and strong minimisation where possible. Safeguarding data handling requires more restrictive access and longer retention than many general operational records, which should influence schema boundaries and query patterns.

Analytics should avoid reading directly from heavily normalized OLTP tables for large cross-branch or trend analysis. Instead, denormalized reporting views and later warehouse pipelines should support dashboards, forecasting, and benchmarking without harming daily operational performance.

## 11. Security Architecture

### RBAC, ABAC, and safeguarding isolation

Security should combine RBAC with ABAC because role-only models are too coarse for real nursery workflows, especially around safeguarding, HR, branch-specific access, and child-linked permissions. Managers may have broad operational rights but should not automatically see all safeguarding or HR records without explicit policy allowance.

Safeguarding isolation should include separate logical boundaries, restricted queries, stronger review procedures, and highly visible access logging. This is a trust issue as much as a security issue because DSLs and operators will only adopt in-system safeguarding if they trust confidentiality controls.

### MFA, encryption, secrets, and secure SDLC

MFA should be mandatory for managers, finance staff, DSLs, admins, and other privileged roles. Encryption should include TLS 1.3 in transit and AES-256 or equivalent at rest with KMS-managed keys.

Secrets management should rely on managed secret stores or Vault-style tooling rather than environment sprawl or repository storage. Secure SDLC controls should include SAST, DAST, dependency scanning, secret scanning, and release gates that block production deploys on critical findings.

### OWASP, API, mobile, and upload security

The target should be roughly aligned to OWASP ASVS Level 2 for the product’s sensitivity and B2B profile. API security should include strict authorization checks, rate limiting, request validation, IDOR prevention, and strong auditability of privileged actions.

Mobile security should include secure token storage, session expiry handling, and careful offline data storage boundaries. Upload security should include file type restrictions, malware scanning, signed download URLs, and controlled document/media visibility.

### Tenancy isolation and incident response

Tenant isolation depends on overlapping controls: application-layer scoping, database RLS, strong auth claims, tests for cross-tenant leakage, and safe query patterns. Incident response should include prepared runbooks for auth compromise, payment webhook abuse, suspicious file uploads, data leaks, and messaging abuse.

## 12. AI Strategy

### AI principles and architecture

AI should be introduced as an efficiency and insight layer, not a substitute for professional judgement in safeguarding, finance, or compliance-critical decisions. All model access should flow through a dedicated AI orchestration layer that manages providers, prompts, safety policies, grounding sources, observability, and cost control.

The recommended foundation is retrieval-augmented workflows using indexed guidance such as EYFS references, internal policy documents, SEND guidance, help content, and tenant-specific documents stored with provenance metadata. Prompt templates should be versioned and narrow in purpose rather than relying on a single generic assistant prompt.

### AI use cases

| AI feature | Business value | Risks | Human-in-loop strategy | Data and architecture implications |
|---|---|---|---|---|
| EYFS observation drafting | Faster practitioner documentation | Hallucinated or low-quality descriptions | Staff review before save or share | Needs child context, prior observations, prompt logging. |
| Communication drafting | Faster parent messaging and newsletters | Tone errors or factual mistakes | User must approve before sending | Needs templates, tenant voice settings, moderation logs. |
| Occupancy forecasting | Better staffing and commercial planning | Over-trust in predictions | Forecasts shown as advisory | Needs attendance history, booking patterns, reporting pipeline. |
| Staffing optimisation | Reduced rota inefficiency | Unrealistic operational suggestions | Manager approval required | Needs future rota/staffing domain maturity. |
| Anomaly detection | Flags unusual billing, attendance, or incidents | False positives and alert fatigue | Review queue only, not auto-action | Needs event streams, explainability, threshold tuning. |
| Safeguarding assistance | Faster chronology summarisation | Severe sensitivity and hallucination risk | Restricted, reviewed, non-autonomous use only | Needs strongest governance, redaction, and audit control. |
| Inspection readiness scoring | Better manager preparation | False confidence | Scores must show evidence and caveats | Needs reporting models and explainable inputs. |

### Governance, risk, and UX implications

AI governance should assign ownership, success metrics, prompt/version history, cost budgets, and explicit release controls for each AI capability. Hallucination risk is acceptable only where the human user is clearly in charge and source context can be inspected.

UX should favour embedded, workflow-specific “assist” actions over an open-ended generic chatbot because constrained workflows are easier to govern, explain, and trust. AI output should be clearly labelled as draft or suggestion, never disguised as authoritative compliance advice.

## 13. Reporting and Analytics

### Operational dashboards

Operational dashboards should show daily attendance, expected versus actual occupancy, live room ratios, incidents, staff presence, and pending operational exceptions. The primary goal is to support day-of decision-making rather than historical reporting alone.

### Finance and occupancy dashboards

Finance dashboards should include invoicing volume, funding versus fee income, aged debt, payment completion, credit note usage, and invoice exception queues. Occupancy dashboards should show room utilization, age-band demand, booked versus attended sessions, and future capacity pressure.

### Safeguarding and inspection readiness dashboards

Safeguarding dashboards should surface concern volumes, review timeliness, open cases, serious incidents, and required follow-up actions in a restricted role-aware way. Inspection readiness views should aggregate suitability status, policy acknowledgements, training completion, incidents, ratios, and exportable evidence packs.

### KPI catalog and executive dashboards

Executive reporting should include occupancy, margin proxies, debt, branch comparison, staffing health, safeguarding trends, and inspection readiness indicators. Branch comparison dashboards are especially valuable for groups and franchises because they expose process inconsistency and underperformance across sites.

## 14. UX and Product Design Strategy

### Mobile-first and tablet-first strategy

The product should be mobile-first overall and tablet-first in core room workflows because the highest-frequency staff actions happen while moving, supervising children, and managing rooms. Attendance, incidents, observations, and room dashboards must therefore be fast, touch-friendly, and tolerant of intermittent connectivity.

### Parent UX and operational UX principles

Parent UX should prioritize reassurance, clarity, secure communication, and low-friction billing interactions. Operational UX should prioritize speed, exception handling, and cognitive simplicity under pressure rather than visual flourish.

### Accessibility, offline-first, and high-speed data entry

Accessibility should align to strong baseline standards including contrast, readable typography, predictable navigation, and screen-reader support where relevant. Offline-first behavior is especially important for in-room staff experiences, where unreliable connectivity should not cause data loss or force fallback to paper.

High-speed data entry patterns should include sticky filters, defaults, bulk actions, scan-friendly layouts, and minimal taps for repetitive tasks. Notification fatigue should be mitigated through preference controls, batching, severity-based routing, and role-aware defaults.

### Navigation model and workflow optimization

Staff navigation should be organised by operational job rather than by system module labels alone, such as Today, Children, Attendance, Incidents, Messages, Reports, and Settings. Parents should see child feed, messages, invoices, forms, and account actions in a simple, trust-focused hierarchy.

## 15. DevOps and Platform Engineering

### CI/CD, environments, and Kubernetes strategy

Recommended environments are local, dev, staging, and production, with clear deployment contracts and reproducible infrastructure definitions. CI/CD should use automated build, test, scan, and deployment workflows with feature flags and safe rollout strategies such as blue-green or canary deployment where justified.

Kubernetes is a valid long-term target on AWS, but managed container platforms are acceptable initially if the deployment model, observability contract, image standards, and secrets workflows are aligned with the intended long-term platform. Infrastructure as code should be mandatory to reduce drift and improve repeatability in a regulated B2B SaaS environment.

### Observability, backup, DR, and release engineering

Monitoring must cover both system health and business health, with alerts for queue backlog spikes, invoice batch failures, ratio event delays, and auth anomalies. Backup strategy should include point-in-time database recovery, object storage durability, tested restore procedures, and documented recovery objectives.

Release engineering should use feature flags, rollback plans, smoke tests, and environment promotion controls. Platform team responsibilities should include CI/CD, runtime health, secrets, environments, developer experience, observability, and resilience patterns.

### Operational maturity roadmap

A realistic maturity path is basic secure SDLC and observability at MVP, stronger MFA and audit controls in Phase 1, formal penetration testing and broader compliance readiness in later enterprise stages, and eventually more mature SRE/platform operations once scale and SLAs demand it.

### Testing and QA Strategy

### Test pyramid and test layers

A balanced test pyramid should emphasise domain-heavy unit tests, robust integration tests, a small number of high-value end-to-end tests, and targeted performance, security, accessibility, and compliance suites. Finance logic, ratio logic, and authorization logic require especially strong regression coverage because failures there have disproportionate business impact.

Required test layers include unit tests for domain services and validation, integration tests for repositories and event publication, E2E tests for critical user journeys, contract tests for APIs and future extracted services, load tests for spikes, security tests for privilege abuse, and compliance tests for retention and auditability.

### Quality gates, release criteria, and UAT

No production release should proceed if critical finance regressions, privilege boundary failures, or audit pipeline failures are detected. Release criteria should therefore include passing regression suites, core journey E2E stability, observability health, migration validation, and manual review for high-risk modules.

UAT should use real pilot nursery workflows and realistic edge cases, especially around attendance corrections, invoice disputes, entitlement changes, and safeguarding access boundaries. Mobile testing must include poor network conditions, interrupted sync, device fragmentation, and notification behavior.

## 17. Delivery Roadmap

### Recommended implementation roadmap

The recommended development order is identity and tenant core, child and parent domain, attendance core, ratio engine, funding engine, invoice engine, payments, parent app and messaging, learning MVP, incident management, safeguarding core, reporting and inspection pack, staff management and DBS tracking, rotas, SEND, multi-branch management, analytics, and AI features. This order is optimal because it secures tenancy first, then builds the operational and financial heart of the product before layering broader operational completeness and intelligence.

### Suggested sprint roadmap

A representative MVP sprint grouping is:

- Sprints 1–2: platform foundations, IAM, tenant model, audit baseline.
- Sprints 3–4: child, guardian, consent, parent onboarding.
- Sprints 5–6: attendance capture, room views, staffing hooks.
- Sprints 7–8: ratio engine, alerts, dashboard views.
- Sprints 9–10: funding calculation core.
- Sprints 11–12: invoices, payment workflows, parent billing UX.
- Sprints 13–14: messaging, notifications, diary timeline.
- Sprints 15–16: learning MVP, incidents, reports, hardening.

### Staffing and hiring roadmap

A serious MVP team should include a founding CTO or principal architect, 2–3 productive backend engineers, 2 frontend/mobile engineers, 1 product manager, 1 QA automation engineer, and part-time DevOps/platform support. As pilots begin, stronger design support, dedicated DevOps ownership, and at least one engineer focused deeply on funding/finance rules and another on mobile staff workflow UX become high-leverage additions.

After MVP validation, hiring should expand into data/reporting, platform engineering, deeper finance and safeguarding capability, customer implementation support, and eventually AI engineering once there is enough production data to justify it.

### Risks and why the sequence is optimal

Major technical risks include underestimating funding complexity, weak room-staff UX, poor offline behavior, and delayed audit/security foundations. Product risks include scope creep into payroll or broad workflow tooling before the core operational engine is validated.

Operational and compliance risks include inadequate safeguarding isolation, poor support processes during pilot rollout, and insufficient regression coverage for finance and authorization logic. The recommended sequence is optimal because it first creates a secure and coherent platform base, then proves the operational and financial core, and only after that broadens into enterprise governance and AI-assisted capabilities.

## 18. Final Strategic Recommendations

### What not to build early

Do not build payroll, advanced rota optimization, broad custom workflow builders, or deep enterprise governance early. These are tempting expansion vectors but they introduce complexity before the product has fully solved the higher-value problems of funding, attendance, compliance evidence, and parent trust.

### Dangerous architectural mistakes and hidden risks

The most dangerous architectural mistake is prematurely adopting microservices before product-market fit and before clear isolation boundaries are proven in production. Another major risk is building a monolith without real bounded-context discipline, which would make finance, safeguarding, attendance, and reporting harder to evolve safely.

Hidden risks include fragile funding rules, poor staff UX causing shadow processes, and safeguarding workflows that users do not trust enough to use consistently. A product can look feature-complete on paper while still failing commercially if practitioners and managers revert to WhatsApp, spreadsheets, and paper for high-stress workflows.

### Highest ROI investments and moat opportunities

The highest ROI investments are a deterministic funding engine, a trustworthy ratio engine, mobile-first room workflows, audit and evidence architecture, and inspection-ready reporting. These capabilities directly influence revenue, operational efficiency, compliance confidence, and customer retention.

Strategic moat opportunities include cross-tenant operational intelligence, explainable finance automation, safeguarding trust, and AI assistance built on normalized childcare operational data. Over time, this combination can turn the platform from a system of record into a system of guidance and operational advantage.

### What would make the platform dominate, and what would cause failure

The platform can dominate if it becomes the most trusted operational backbone for UK nurseries by combining funding accuracy, ratio confidence, safeguarding trust, parent-friendly UX, mobile usability, and enterprise-grade architecture. It must win not by having the longest feature list, but by being the most reliable and operationally credible system in the most painful workflows.

The most likely failure mode is losing focus and shipping a shallow all-in-one app that underdelivers in the regulated, high-consequence workflows where nurseries actually need help. A second likely failure mode is neglecting adoption realities, especially staff speed, offline tolerance, and trust in sensitive workflows.

