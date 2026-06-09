# MVP 30-Day Angular Frontend Backlog

Source: `docs/PRD-MVP-1M.md`  
Related baseline: `docs/DECISION-BASELINE.md`, `CONTEXT.md`

## Goal and Scope

Build the Angular web frontend for the month-1 nursery MVP with professional operational UI/UX for:

- Manager administration, attendance correction, funding, invoice generation, invoice status, and user invites.
- Practitioner mobile/tablet-first attendance check-in and check-out.
- Parent portal invoice viewing and Stripe Checkout payment redirect.

This backlog is frontend-only. It includes Angular UI, client-side state, routing, API integration, mock/API-contract work, frontend tests, and visual QA. Backend implementation, database work, Stripe webhook processing, deployment, and production operations are out of scope except where listed as API dependencies.

## Context Alignment Notes

- Use `check-in` and `check-out` wording, not sign-in/sign-out, for child attendance.
- `Parent Portal (MVP)` means the parent-side surface for viewing issued invoices for linked children and initiating full invoice payment.
- Current implemented month-1 role surface is `Manager`, `Practitioner`, and `Parent`; the product-owner update below adds `Owner` as a required user type that needs an explicit access model before implementation.
- Practitioners must not see guardian contact details or billing information.
- Child and guardian management is manager-only.
- Issued invoices are immutable; direct edit UI must not exist.
- Public signup is not part of month 1; users are manager-invited except the first seeded manager.
- Funding v1 is simple per-child monthly funded-hours allowance.
- Invoices are per child, not combined at family level.

## Product-Owner Update: Users and Access

- Intended system users now include the owner, nursery managers for the current four sites, staff in each site, and parents.
- Nursery managers need site/branch-scoped staff administration and operational workflows.
- Staff need site/branch-scoped workflows for the children they support, with no billing or guardian-contact leakage unless later role decisions explicitly allow it.
- The owner likely needs cross-site visibility and administration across all four sites; this must be planned as a distinct owner experience, not hidden inside the branch manager UI.
- Parents only access the `app/` parent portal, and their navigation and data must be limited to records that concern them and their linked child or children.
- Future tickets must confirm whether owner and four-site UX is required for the live pilot or scheduled as post-MVP before changing route guards, navigation, or completed MVP screens.

## Decisions to Honor

- Create this as a separate Angular-only backlog; do not replace `docs/MVP-30D-BACKLOG.md`.
- Remove or block TailAdmin demo routes/pages from the production navigation.
- Professional UI/UX means dense, work-focused operational screens, not marketing pages or decorative dashboards.
- Use API contract/mock-state tasks before each real `/api/v1` integration so frontend work can proceed ahead of backend completion.
- Keep one Angular app and auth/session infrastructure, with role-specific staff and parent layouts.
- Practitioner attendance is mobile/tablet-first.
- Manager workflows are desktop-dense with responsive mobile support.
- Add a pragmatic shared frontend foundation only where MVP screens need it.
- Login supports membership/scope selection only when a user has multiple available memberships.
- Manager invite UX supports practitioner and parent invites only; manager invitation remains out of scope.
- Child and guardian lists remain separate, with relationship actions and cross-links.
- Attendance corrections are a manager workflow separate from practitioner attendance.
- Absence marking is lower priority than check-in/check-out and corrections.
- Funding editing lives on child detail/enrollment; a funding overview is optional.
- Invoice generation is a guided monthly run workflow.
- Stripe payment uses hosted Checkout redirect only; no card-entry UI in Angular.
- Parent invoices are grouped by child, with unpaid, failed, and overdue invoices surfaced first.
- Accessibility, responsive behavior, and useful loading/empty/error states are part of every ticket's done criteria.
- Add Playwright after critical screens exist for frontend smoke and screenshot QA.
- Expected API errors should map to domain-specific UI recovery paths.

## Global Frontend Definition of Done

- Screen is reachable only by the correct role through Angular routes and guards.
- UI handles loading, empty, validation, authorization, and unknown error states.
- Expected API errors map to useful UI actions or explanations.
- Unknown API errors show a generic message with request id when present.
- Keyboard navigation, labels, focus states, and ARIA names are present for custom controls.
- Mobile and desktop layouts have no text overlap, clipped controls, or hidden required actions.
- Touch targets are large enough for practitioner attendance and parent payment actions.
- Component/service tests cover route guards, API mapping, form validation, and state transitions where risk is meaningful.
- `npm run build` passes in `web/`.

## Week 1 - Frontend Foundation, Auth, and Core Staff UX

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-01 ~~done~~ | ~~Audit current Angular app and remove TailAdmin demo navigation from production shell: ecommerce, charts, forms, UI elements, calendar, generic invoice demo, public signup links.~~ | Existing `web/` app | ~~Sidebar/header only show role-relevant MVP routes; hidden demo routes cannot be reached from app navigation.~~ |
| FE-02 ~~done~~ | ~~Create role-specific route map and default routes: manager dashboard, practitioner attendance, parent invoices.~~ | FE-01 | ~~`defaultRouteForRole(parent)` points to parent invoices; wrong-role routes redirect to the user's default route.~~ |
| FE-03 ~~done~~ | ~~Build shared MVP UI foundation: page header, action button styles, status badges, form field wrapper, inline alert, empty state, loading state, data table pattern, drawer/modal, confirmation dialog, toast/notification pattern.~~ | FE-01 | ~~New screens use consistent primitives instead of one-off TailAdmin fragments.~~ |
| FE-04 ~~done~~ | ~~Replace generic dashboard with manager operations dashboard. Show today's attendance summary, incomplete attendance, invoice run status, unpaid/overdue invoices, and quick actions using mock data first.~~ | ~~FE-02, FE-03~~ | ~~Manager lands on operational dashboard with no ecommerce/template copy.~~ |
| FE-05 ~~done 2026-06-05~~ | ~~Update sign-in UX for session scope model. Auto-enter when one membership is returned; show a minimal membership picker only when multiple memberships require explicit selection.~~ | ~~Auth API contract~~ | ~~Login works for one-scope users and supports explicit `membership_id` retry for multi-scope users. Verified: 67 Go tests, 114 Angular tests pass.~~ |
| FE-06 ~~done 2026-06-06~~ | ~~Remove/block public signup. Replace `/signup` with a no-public-signup message or remove route entirely.~~ | ~~FE-01~~ | ~~No public account creation form is visible or linked. `/signup` shows invitation-only message. Unused `SignupFormComponent` deleted. 120 tests pass.~~ |
| FE-07 ~~done 2026-06-06~~ | ~~Add forgot-password and reset-password screens with invalid/expired link states.~~ | ~~Password reset API contract~~ | ~~User can request reset, submit new password from token route, and return to sign-in. Verified: forgot-password and reset-password routes, account-enumeration-safe confirmation, terminal link states, 145 Angular tests pass.~~ |
| FE-08 ~~done 2026-06-06~~ | ~~Add invite acceptance and set-password screens for manager-invited practitioner/parent users.~~ | ~~Invite API contract~~ | ~~Public `/invite-accept` route renders set-password form for valid tokens; terminal states for invalid, expired, already-accepted, revoked tokens; rate-limited as nonterminal form error. 164 Angular tests pass.~~ |
| FE-09 ~~done 2026-06-06~~ | ~~Add manager user-invite screen for practitioner and parent roles. Include send invite, invite status list, resend/revoke actions when API supports them.~~ | ~~FE-03, invite API contract~~ | ~~Manager can invite non-manager users; manager role is unavailable in the invite UI. Verified: invite form, status list, resend/revoke actions, confirmation dialog, role guard, sidebar link. 192 Angular tests pass.~~ |
| FE-10 ~~done 2026-06-07~~ | ~~Improve manager child and guardian list UX. Keep separate screens, add cross-links/placeholders for linked records, better status filters, GBP rate formatting, enrollment badges, and no hard-delete actions.~~ | ~~Existing child/guardian screens~~ | ~~Screens pilot-usable with GBP rates, enrollment badges, missing-requirement labels, cross-links, linked-record placeholders, no hard-delete actions. 237 Angular tests pass.~~ |
| FE-11 ~~done 2026-06-07~~ | ~~Add child detail/enrollment surface. Include child basics, linked guardians, missing enrollment requirements, current core hourly rate, and monthly funded-hours editor placeholder.~~ | ~~FE-10~~ | ~~Manager can inspect enrollment readiness from one child-focused surface. Backend: GET /children/:child_id/guardian-child-links with linked guardian summary, child existence check, manager-only authz. Frontend: /staff/manager/children/:childId route, child detail component with linked guardians, link-guardian action, edit child basics, funded-hours placeholder. 252 Angular tests, 405 Go tests pass.~~ |

## Week 2 - Attendance, Corrections, Funding

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-12 ~~done 2026-06-07~~ | ~~Redesign practitioner attendance list mobile-first. Use large check-in/check-out actions, quick search/filter, checked-in/not-in badges, refresh, row errors, and no guardian contact/billing fields.~~ | ~~Existing attendance screen, FE-03~~ | ~~Mobile-first card layout with full-width touch actions, inline pending state, search/filter pills, row errors per child, privacy regression tests. 260 Angular tests pass.~~ |
| FE-13 ~~done 2026-06-08~~ | ~~Integrate attendance list with `/api/v1/children/attendance`, `/api/v1/attendance/check-ins`, and `/api/v1/attendance/check-outs`.~~ | ~~Attendance API~~ | ~~Enrollment-incomplete children cannot be checked in; duplicate/open-session errors map to row messages; absent state rendered with no check-in action; API conflict errors shown as row-level messages with request IDs. 270 Angular tests pass.~~ |
| FE-14 ~~done 2026-06-08~~ | ~~Add optional light polling/manual refresh strategy for attendance when multiple devices are used.~~ | ~~FE-13~~ | ~~30s polling default-on with toggle, manual refresh always available, last-updated timestamp, background refresh preserves list/actions, no WebSocket/SSE. 281 Angular tests pass.~~ |
| FE-15 ~~done 2026-06-08~~ | ~~Build manager attendance correction workflow. Use child/session selector, corrected check-in/check-out interval, required reason code, optional note, validation feedback, and correction history.~~ | ~~Correction API contract~~ | ~~Manager-only route at /staff/manager/attendance-corrections with sidebar link. Child/date selector with active-first ordering. Session list with open/complete/corrected status. Missed-session mode for dates with no sessions. Reason code required, note required for 'other'. Correction history panel with chronological events. Issued-invoice warning from API. Dashboard deep links with query params. Backend adds GET /attendance/sessions and GET /attendance/sessions/:id/history manager-only endpoints. 409+ tests pass.~~ |
| FE-16 ~~done 2026-06-08~~ | ~~Map attendance correction errors to specific UI states: overlap, outside enrollment window, missing reason, incomplete session, authorization.~~ | ~~FE-15~~ | ~~Correction submit errors map to actionable UI: overlap shows time guidance, enrollment-window shows child range, missing reason/note shows field guidance, incomplete sessions show info alert, forbidden shows no-access warning, unknown errors show fallback with request ID. Local validation hints for missing reason, note, and time order. Field errors clear on relevant input changes. 309 Angular tests pass.~~ |
| FE-17 ~~done 2026-06-08~~ | ~~Add low-priority absence marker UI only after core attendance works and the API exists.~~ | ~~Absence API~~ | ~~Absence is a simple marker and does not expose billing-rule configuration. Staff can mark absent and clear absence from today's attendance list. Eligible not-in children show Check in + Mark absent. Absent children show Clear absence. Row-level pending/error/reload behavior matches check-in/check-out. 331 Angular tests pass.~~ |
| FE-18 ~~done 2026-06-08~~ | ~~Add funded-hours editing on child detail/enrollment. Show monthly allowance, save state, validation, and last-updated status.~~ | ~~FE-11, funding API~~ | ~~Manager can edit per-child monthly funded-hours allowance with billing month selector, hours/minutes inputs, client validation, API error mapping, and last-updated timestamp. Explicit zero distinct from missing profile. 358 Angular tests pass.~~ |
| FE-19 ~~done 2026-06-09~~ | ~~Add optional funding overview if time remains. Show children missing allowance or with unusual values.~~ | ~~FE-18~~ | ~~Manager-only funding overview at /staff/manager/funding with billing month selector. Flags: missing profile, explicit zero, under one hour, above 160 hours. Summary counts. Review link deep-links to child detail with billing_month query param. Backend GET /api/v1/funding/overview endpoint. 370 Angular tests, 20 Go application tests pass.~~ |

## Week 3 - Invoicing, Parent Portal, Stripe Redirect

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-20 | Build guided manager invoice run workflow with mock data: month selector, preflight summary, exception list, draft generation, draft review, bulk issue confirmation, one-by-one issue fallback. | FE-03 | Workflow is understandable without spreadsheets and makes bulk issue the default. ✅ Done 2026-06-09 — 97 tests pass, build clean. |
| FE-21 | Integrate invoice run with `/api/v1` invoice draft generation and issue endpoints. | Invoice API | Incomplete attendance blocks only affected children and appears in exceptions. |
| FE-22 | Build manager invoice list/detail. Include status filters, per-child invoice identity, line items, funded deduction summary, net due, due/overdue status, and immutable issued state. | FE-21 | Issued invoices have no direct edit controls. |
| FE-23 | Add lean manager payment/reconciliation UX. Show payment status, payment event detail, retry checkout action where applicable, and basic webhook/payment history if API exists. | Payments API | Manager can understand paid/unpaid/failed status and retry payment flow. |
| FE-24 | Add minimal issued-invoice adjustment UI after core invoicing works, if API exists. Require reason and show linked adjustment invoice relationship. | Adjustment API | No adjustment entry point appears before invoice is issued. |
| FE-25 | Create parent portal layout. Use same auth/session infrastructure, but only parent-relevant navigation and no staff sidebar/admin template pages. | FE-02 | Parent users land in parent portal, not staff dashboard. |
| FE-26 | Build parent invoices page grouped by child, with urgent unpaid, failed, and overdue invoices surfaced first. | Parent invoice API contract | Parent sees only issued-or-later invoices for linked children. |
| FE-27 | Add parent invoice detail and Stripe Checkout redirect action. Call API to create a fresh checkout session, then redirect browser to hosted Stripe Checkout. | Stripe checkout API | No card fields or custom payment UI exist in Angular. |
| FE-28 | Add payment return/status UX. On return from Stripe, refresh invoice status and optionally poll briefly until paid/failed state is visible. | FE-27 | Parent sees clear paid, failed, canceled, or still-processing states. |

## Week 4 - Hardening, QA, Pilot Polish

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-29 | Add domain-specific API error mapping across auth, attendance, child/guardian, invoice, and payment flows. | Core flows | Known errors produce action-oriented UI; unknown errors remain generic with request id. |
| FE-30 | Add Angular unit/component tests for auth service, guards, role default routes, API mappers, attendance state, invoice run state, and parent payment redirect behavior. | Core flows | High-risk state and routing behavior is covered. |
| FE-31 | Add Playwright setup for frontend smoke tests and screenshots. | Core screens exist | Playwright can run against local Angular dev server. |
| FE-32 | Add Playwright smoke tests: login role routing, practitioner attendance mobile, manager invoice run desktop, parent invoices mobile, and Stripe redirect handoff mock. | FE-31 | Smoke suite verifies the MVP paths without relying on real Stripe. |
| FE-33 | Perform visual QA at mobile, tablet, and desktop widths for login/scope selection, practitioner attendance, manager invoice run, manager dashboard, parent invoices, and role navigation. | FE-31 | Screenshots show no overlap, clipped text, missing actions, or template content. |
| FE-34 | Accessibility pass for all MVP screens. Check keyboard flow, focus visibility, form labels/errors, dialog focus handling, aria names, and color contrast. | Core screens | Critical flows are usable without a mouse. |
| FE-35 | Pilot copy and terminology pass. Replace internal/template wording with nursery domain terms from `CONTEXT.md`. | Core screens | UI consistently uses Manager, Practitioner, Parent, Guardian, Child, check-in, check-out, invoice, funded-hours allowance. |
| FE-36 | Final frontend production readiness pass. Remove dead demo imports/routes where safe, confirm environment API base URL, run build/tests, and document remaining API blockers. | FE-30 to FE-35 | `npm run build` passes and remaining blockers are explicit. |

## Post-MVP Frontend Feature Backlog

These items are product features intentionally outside the month-1 pilot critical path unless UAT or pilot operation makes them necessary. Source forms live in `docs/forms/`.

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-PM-01 | Build manager child registration/enrolment editor based on `docs/forms/child-application-form.md`. Use sectioned forms for child demographics, medical information, doctor/health visitor, social-services contact, developmental concerns/referrals, dietary requirements, parent/carer details, parental responsibility, emergency contacts, authorised collectors, benefits/funding notes, routines/free-text notes, and office-use fields. | FE-11; API-PM-02 | Manager can save partial sections, see completion state, and review missing registration requirements without losing existing child/guardian CRUD behavior. |
| FE-PM-02 | Build parent/guardian digital registration and consent journey based on `docs/forms/child-application-form.md` and `docs/forms/parental-consent-form.md`. Support draft save, clear review before submission, signer details, date capture, and confirmation of GDPR/safeguarding acknowledgements. | Parent portal shell; API-PM-02; API-PM-03 | Parent/guardian can complete required registration and consent fields on mobile and desktop; validation explains missing required fields; submitted consent state is read-only except through an explicit supersede/update flow. |
| FE-PM-03 | Add consent review/history UI for managers. Show current consent decisions and historical superseded records for urgent medical treatment, plasters, SENCO, health visitor, transition documents, outings, face painting, sun cream, nappy cream, photographs, website/promotional use, coursework, and social media. | FE-PM-01; API-PM-03 | Manager can quickly inspect current consent before operational decisions and can trace who changed a decision and when. |
| FE-PM-04 | Add registration office-use checklist UI for deposit, application/start/date-left, sessions/days requested, term-time-only status, contract/handbook handoff, Red Book check, birth certificate/passport check, and proof-of-address check. | FE-PM-01; API-PM-04 | Manager can mark document/checklist completion from child enrollment; checklist status feeds the child detail readiness summary. |
| FE-PM-05 | Build room/session planning and booking UI. Include room setup, session templates, child bookings, extra-session requests, capacity warnings, eligibility messages, and links from child detail to booked sessions. | FE-PM-01; API-PM-05 | Manager can plan and adjust booked attendance from a calendar/list workflow; capacity and child-eligibility failures are visible before save; practitioners see only the operational session list needed for attendance. |
| FE-PM-06 | Build ratio safety dashboard and live room checks. Show room-level ratio state, staff assignment, child age-band counts, at-risk sessions, and manager override flow where the API allows it. | FE-PM-05; API-PM-06 | Manager can see unsafe or near-limit sessions before and during the day; practitioner attendance surfaces clear ratio warnings without exposing unnecessary staff or child-sensitive data; override actions require reason and confirmation. |
| FE-PM-07 | Build safeguarding and incident record screens. Support practitioner incident submission, manager restricted review, status/follow-up actions, confidential notes, and clear separation from normal child profile data. | API-PM-07 | Practitioners can submit permitted incidents; managers can triage and close records with history; restricted safeguarding data is not visible in parent, invoice, or routine attendance surfaces. |
| FE-PM-08 | Build learning journey and EYFS observation screens. Support practitioner observation drafts, EYFS tagging, next steps, manager review, and parent-visible approved entries. | API-PM-08 | Practitioners can capture observations; managers can approve entries; parents see only approved learning journey content for linked children; draft/rejected content stays staff-only. |
| FE-PM-09 | Build owner and four-site navigation/access experience. Add owner landing view, site switcher or cross-site filters, branch-scoped manager/staff views, and parent `app/` route isolation. | API-PM-09; product decision on owner MVP scope | Owner can move between or summarize the current four sites without using a branch manager-only shell; nursery managers and staff see only their site scope; parents land only in the `app/` parent portal and cannot see staff or owner navigation. |

## Files to Create or Change

Expected Angular files and folders:

- `web/src/app/app.routes.ts`
- `web/src/app/core/constants/roles.ts`
- `web/src/app/core/services/auth.service.ts`
- `web/src/app/core/guards/auth.guard.ts`
- `web/src/app/core/guards/role.guard.ts`
- `web/src/app/core/errors/api-error.mapper.ts`
- `web/src/app/shared/layout/app-layout/*`
- `web/src/app/shared/layout/app-sidebar/*`
- `web/src/app/shared/layout/app-header/*`
- `web/src/app/shared/layout/parent-portal-layout/*`
- `web/src/app/shared/components/ui/*`
- `web/src/app/pages/auth-pages/sign-in/*`
- `web/src/app/pages/auth-pages/invite-accept/*`
- `web/src/app/pages/auth-pages/forgot-password/*`
- `web/src/app/pages/auth-pages/reset-password/*`
- `web/src/app/features/staff/**`
- `web/src/app/features/attendance/**`
- `web/src/app/features/funding/**`
- `web/src/app/features/invoicing/**`
- `web/src/app/features/payments/**`
- `web/src/app/features/parent-portal/**`
- `web/src/app/features/registration/**` for post-MVP registration and consent workflows.
- `web/src/app/features/planning/**` for post-MVP room, session, and booking workflows.
- `web/src/app/features/ratios/**` for post-MVP staff-to-child ratio safety workflows.
- `web/src/app/features/safeguarding/**` for post-MVP incident and safeguarding workflows.
- `web/src/app/features/learning/**` for post-MVP learning journey and EYFS workflows.
- `web/src/environments/environment.ts`
- `web/src/environments/environment.development.ts`
- `web/package.json` for Playwright scripts when FE-31 starts.
- `web/playwright.config.ts` when FE-31 starts.
- `web/e2e/**` or `web/tests/e2e/**` for Playwright tests.

Existing TailAdmin demo pages can remain temporarily if needed, but production navigation and role routes must not expose them. Remove them when safe during FE-36.

## API Dependencies to Track

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- invite create/list/resend/revoke/accept endpoints
- password forgot/reset endpoints
- child list/create/update/detail endpoints
- guardian list/create/update/detail endpoints
- guardian-child link/end/relink endpoints
- parent membership to guardian mapping endpoints
- attendance daily list, check-in, check-out, correction, correction history endpoints
- optional absence marker endpoints
- funding profile read/update endpoints
- invoice draft generation, preflight, draft review, issue, list, detail, adjustment endpoints
- parent invoice list/detail endpoints
- checkout session creation endpoint
- payment/reconciliation event endpoints
- post-MVP registration profile, consent ledger, office-use checklist, planning, ratio, safeguarding, and learning endpoints from API-PM-02 to API-PM-08

When an API is unavailable, build the UI with a typed local mock adapter and keep the real service method shape aligned to the expected `/api/v1` contract.

## Verification Steps

- Run from `web/`: `npm run build`.
- Run Angular tests from `web/`: `npm test` or the project-approved non-watch equivalent once configured.
- Run Playwright smoke tests after FE-31: `npx playwright test`.
- Manually validate role navigation:
  - Manager cannot see parent-only portal navigation.
  - Practitioner lands on attendance and cannot access billing or guardian contact details.
  - Parent lands on invoices and cannot access staff routes.
- Manually validate core responsive screens at mobile, tablet, and desktop widths.
- Capture screenshot QA for:
  - login and membership selection
  - practitioner attendance mobile
  - manager operations dashboard desktop
  - manager invoice run desktop
  - parent invoices mobile
- Confirm no public signup flow is available.
- Confirm no hard-delete UI exists for child, guardian, attendance, or invoice records.
- Confirm issued invoices have no direct edit controls.
- Confirm parent payment uses hosted Stripe redirect only.

## Explicit Assumptions

- The existing Angular app in `web/` remains the frontend application for the MVP.
- Angular 21 and Tailwind CSS remain in use.
- The backend API will use `/api/v1` and plain JSON resources with the documented error shape.
- The first manager is seeded outside the web UI.
- The older one-default-branch pilot assumption is superseded for planning by the product-owner update identifying one owner and four nursery sites.
- The frontend must respect selected membership scope for managers and staff, relationship scope for parents, and an explicit owner cross-site access model once prioritized.
- Production deployment details are handled outside this frontend backlog.
- Playwright is acceptable to add during week 4 for smoke and screenshot QA.
- Mock data is allowed only to unblock UI development; final acceptance requires real API integration for critical flows.
