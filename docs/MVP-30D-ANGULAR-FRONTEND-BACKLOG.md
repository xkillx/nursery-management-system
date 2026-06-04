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
- `Manager`, `Practitioner`, and `Parent` are the only month-1 roles.
- Practitioners must not see guardian contact details or billing information.
- Child and guardian management is manager-only.
- Issued invoices are immutable; direct edit UI must not exist.
- Public signup is not part of month 1; users are manager-invited except the first seeded manager.
- Funding v1 is simple per-child monthly funded-hours allowance.
- Invoices are per child, not combined at family level.

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
| FE-03 | Build shared MVP UI foundation: page header, action button styles, status badges, form field wrapper, inline alert, empty state, loading state, data table pattern, drawer/modal, confirmation dialog, toast/notification pattern. | FE-01 | New screens use consistent primitives instead of one-off TailAdmin fragments. |
| FE-04 | Replace generic dashboard with manager operations dashboard. Show today's attendance summary, incomplete attendance, invoice run status, unpaid/overdue invoices, and quick actions using mock data first. | FE-02, FE-03 | Manager lands on operational dashboard with no ecommerce/template copy. |
| FE-05 | Update sign-in UX for session scope model. Auto-enter when one membership is returned; show a minimal membership picker only when multiple memberships require explicit selection. | Auth API contract | Login works for one-scope users and supports explicit `membership_id` retry for multi-scope users. |
| FE-06 | Remove/block public signup. Replace `/signup` with a no-public-signup message or remove route entirely. | FE-01 | No public account creation form is visible or linked. |
| FE-07 | Add forgot-password and reset-password screens with invalid/expired link states. | Password reset API contract | User can request reset, submit new password from token route, and return to sign-in. |
| FE-08 | Add invite acceptance and set-password screens for manager-invited practitioner/parent users. | Invite API contract | Invite token route supports valid, expired, already-used, and invalid token states. |
| FE-09 | Add manager user-invite screen for practitioner and parent roles. Include send invite, invite status list, resend/revoke actions when API supports them. | FE-03, invite API contract | Manager can invite non-manager users; manager role is unavailable in the invite UI. |
| FE-10 | Improve manager child and guardian list UX. Keep separate screens, add cross-links/placeholders for linked records, better status filters, GBP rate formatting, enrollment badges, and no hard-delete actions. | Existing child/guardian screens | Existing screens become pilot-usable and remove raw minor-unit display. |
| FE-11 | Add child detail/enrollment surface. Include child basics, linked guardians, missing enrollment requirements, current core hourly rate, and monthly funded-hours editor placeholder. | FE-10 | Manager can inspect enrollment readiness from one child-focused surface. |

## Week 2 - Attendance, Corrections, Funding

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-12 | Redesign practitioner attendance list mobile-first. Use large check-in/check-out actions, quick search/filter, checked-in/not-in badges, refresh, row errors, and no guardian contact/billing fields. | Existing attendance screen, FE-03 | Mobile viewport is the primary usable experience; desktop remains clean. |
| FE-13 | Integrate attendance list with `/api/v1/children/attendance`, `/api/v1/attendance/check-ins`, and `/api/v1/attendance/check-outs`. | Attendance API | Enrollment-incomplete children cannot be checked in; duplicate/open-session errors map to row messages. |
| FE-14 | Add optional light polling/manual refresh strategy for attendance when multiple devices are used. | FE-13 | Polling can be enabled without websockets/SSE and does not block manual refresh. |
| FE-15 | Build manager attendance correction workflow. Use child/session selector, corrected check-in/check-out interval, required reason code, optional note, validation feedback, and correction history. | Correction API contract | Only managers can access correction UI; reason is required before submit. |
| FE-16 | Map attendance correction errors to specific UI states: overlap, outside enrollment window, missing reason, incomplete session, authorization. | FE-15 | Expected errors explain what the manager must change. |
| FE-17 | Add low-priority absence marker UI only after core attendance works and the API exists. | Absence API | Absence is a simple marker and does not expose billing-rule configuration. |
| FE-18 | Add funded-hours editing on child detail/enrollment. Show monthly allowance, save state, validation, and last-updated status. | FE-11, funding API | Manager can edit per-child monthly funded-hours allowance. |
| FE-19 | Add optional funding overview if time remains. Show children missing allowance or with unusual values. | FE-18 | Overview supports triage but is not required for the invoice flow. |

## Week 3 - Invoicing, Parent Portal, Stripe Redirect

| ID | Task | Dependencies | Done check |
|---|---|---|---|
| FE-20 | Build guided manager invoice run workflow with mock data: month selector, preflight summary, exception list, draft generation, draft review, bulk issue confirmation, one-by-one issue fallback. | FE-03 | Workflow is understandable without spreadsheets and makes bulk issue the default. |
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
- The pilot uses one tenant and one default branch, but the frontend must respect selected membership scope.
- Production deployment details are handled outside this frontend backlog.
- Playwright is acceptable to add during week 4 for smoke and screenshot QA.
- Mock data is allowed only to unblock UI development; final acceptance requires real API integration for critical flows.
