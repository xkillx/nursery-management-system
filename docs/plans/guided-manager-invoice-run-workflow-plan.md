# Guided Manager Invoice Run Workflow Implementation Plan

## Goal

Build FE-20: a manager-only guided monthly invoice run workflow using mock data. The workflow must let a Manager choose a billing month, review draft preflight readiness, proceed with eligible child-months even when other children are blocked, generate draft invoices, review draft calculations without spreadsheets, bulk issue ready drafts by default, and fall back to one-by-one issue when needed.

The first implementation is frontend-only and mock-data-backed. It should shape its models and state transitions to match the existing `/api/v1` invoice contract so FE-21 can replace mock operations with real API calls without changing the user workflow.

## Non-Goals

- Do not call real invoice APIs in FE-20; real integration is FE-21.
- Do not create or modify backend handlers, database schema, migrations, generated API clients, or OpenAPI contracts.
- Do not implement manager invoice list/detail beyond the draft review surface needed inside the run workflow; FE-22 owns the full manager invoice list/detail.
- Do not add editable invoice lines, manual extras entry, adjustment invoices, payment collection, payment follow-up, Stripe checkout, or parent invoice behavior.
- Do not fix attendance, enrollment, billing-rate, or funding exceptions inline inside the invoice run workflow. Link to the appropriate existing manager workflow instead.
- Do not expose invoice run UI to Practitioner or Parent roles.
- Do not resurrect TailAdmin ecommerce/demo invoice pages or reuse their ecommerce/customer language.

## Context Alignment

Existing `CONTEXT.md` terms to honor:

- **Billing Month (MVP)**: calendar month in `YYYY-MM` used by funding and invoice workflows.
- **Invoice Draft Preflight (MVP)**: readiness preview for one billing month before draft invoice generation; it is not itself an invoice run.
- **Guided Invoice Run Default Path (MVP)**: eligible child-months remain actionable when other child-months have exceptions.
- **Draft Invoice Generation Exception Handling (MVP)** and **Invoice Issue Exception Handling (MVP)**: expected child/invoice blockers are returned as exceptions while eligible work proceeds.
- **Manager Invoice Review (MVP)**: managers can inspect headers, line items, calculation quantities, status, and due/payment metadata without reconstructing totals elsewhere; review itself does not edit invoice lines.
- **Invoice Run Exception Resolution (MVP)**: exceptions show blocked child-month, reason, and next workflow, without inline editing.
- **Bulk Invoice Issue Default Selection (MVP)**: all ready draft invoices are selected by default for bulk issue; managers can remove individual drafts.
- **Invoice Issue Confirmation (MVP)** and **Invoice Issue Result Summary (MVP)**: both bulk and one-by-one issue need explicit confirmation, and the result identifies issued invoices, assigned numbers, and skipped/failed drafts.
- **Invoice Run Default Billing Month (MVP)**: default to the most recent completed billing month, not the current month.

No ADR is needed. This is a reversible frontend workflow and mock-data implementation that follows already-settled invoice lifecycle decisions.

## Current State

- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md` defines FE-20 as a mock-data guided manager invoice run workflow with month selector, preflight summary, exception list, draft generation, draft review, bulk issue confirmation, and one-by-one issue fallback.
- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md` marks FE-03 done and requires new screens to use shared MVP UI primitives rather than one-off TailAdmin fragments.
- `web/package.json` uses Angular `^21.2.x`, standalone components, Jasmine/Karma tests, and scripts `npm run build` and `npm test`.
- `web/src/app/app.routes.ts` currently registers manager routes for dashboard, children, child detail, guardians, invites, attendance corrections, funding, practitioner attendance, and parent invoices. There is no manager invoice-run route.
- `web/src/app/core/constants/roles.ts` defines `ROLE_ROUTES` for existing manager, practitioner, parent, and auth routes. There is no `managerInvoiceRun` constant.
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts` builds role-specific nav links from `AuthService.currentRole()`. Manager links currently include Dashboard, Children, Guardians, Invites, Attendance, Attendance corrections, and Funding.
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.spec.ts` asserts manager-visible sidebar links and should be updated to include the new invoice-run link and keep it hidden from Practitioner/Parent roles.
- `web/src/app/app.routes.spec.ts` asserts registered MVP paths and manager-only route metadata for existing manager pages. It should be updated for `/staff/manager/invoice-run`.
- `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.component.ts` has a mock Manager Operations Dashboard and a disabled quick action labelled `Start invoice run`.
- `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.models.ts` has the disabled `Start invoice run` quick action. FE-20 should make it route to the new invoice run workflow.
- `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.component.spec.ts` currently expects `Start invoice run` to be disabled. Update it to expect a live link after the route exists.
- `web/src/app/features/staff/pages/manager-funding-overview/manager-funding-overview.component.ts` is a useful local pattern for a manager page with a billing month selector, `PageHeaderComponent`, `AlertComponent`, `TableShellComponent`, `EmptyStateComponent`, and `LoadingStateComponent`.
- `web/src/app/features/staff/data/staff-api.service.ts` already centralizes staff data access and maps API snake_case to frontend camelCase. FE-20 should not add real invoice calls there yet, but new invoice-run view models should be compatible with the API contract.
- Shared primitives available for FE-20 include `PageHeaderComponent`, `AlertComponent`, `TableShellComponent`, `EmptyStateComponent`, `LoadingStateComponent`, `StatusBadgeComponent`, `ButtonComponent`, `ConfirmationDialogComponent`, and `ToastService`.
- `docs/API-CONTRACT-MVP.openapi.yaml` defines manager invoice APIs for FE-21: `GET /api/v1/invoices/drafts/preflight`, `POST /api/v1/invoice-runs/drafts`, `GET /api/v1/invoices`, `GET /api/v1/invoices/{invoice_id}`, `POST /api/v1/invoices/{invoice_id}/issue`, and `POST /api/v1/invoices/bulk-issue`.
- `docs/API-CONTRACT-MVP.openapi.yaml` defines `PreflightResponse` with `summary`, `eligible_children`, and `blocked_children`; `DraftGenerationResponse` with generated and blocked child results; `InvoiceDetail` with calculation and line details; `IssueInvoiceResponse`; and `BulkIssueResponse`.

## Decisions

Product/domain decisions from the interview:

- Preflight and generation are partial-success workflows: eligible child-months proceed while blocked children remain in an exception list.
- Draft review must expose enough per-child calculation detail for manager approval without spreadsheet reconstruction.
- Draft review must not introduce editable invoice lines.
- Bulk issue starts with all ready draft invoices selected; managers can deselect drafts before confirmation.
- One-by-one issue is a fallback from draft rows, not the primary path.
- The invoice run page defaults to the most recent completed billing month.
- Exceptions are a triage queue with child, reason, and next action; underlying data fixes happen in existing attendance, child detail, and funding workflows.
- Issue completion shows a result summary with issued invoices, invoice numbers, and skipped or failed drafts.

Implementation decisions made by the agent:

- Add a new manager route `/staff/manager/invoice-run` and `ROLE_ROUTES.managerInvoiceRun`.
- Add the manager sidebar label `Invoice run` with test id `staff-link-manager-invoice-run`.
- Enable the Manager Operations Dashboard `Start invoice run` quick action and route it to `/staff/manager/invoice-run`.
- Implement FE-20 as one standalone page component under `web/src/app/features/staff/pages/manager-invoice-run/`.
- Add invoice-run models in `web/src/app/features/staff/models/invoice-run.models.ts`.
- Add mock data/state operations in `web/src/app/features/staff/data/invoice-run-mock.service.ts` or an equivalently named staff feature data service. Keep it injectable so component tests can stub behavior and FE-21 can replace it with API-backed methods.
- Keep mock state deterministic. Do not use random issue failures, random invoice numbers, or local storage.
- Use `Europe/London` when computing the default billing month from the current instant.
- Use existing Tailwind utility style conventions and shared primitives. Avoid nested cards and ecommerce/demo copy.

Existing decisions to honor:

- Manager-only billing workflows.
- Draft invoices are manager-only and parents see only issued-or-later invoices.
- Issued invoices are immutable.
- Bulk issue assigns invoice numbers in deterministic child-name order, using invoice identity only as a tie-breaker.
- Invoice numbers follow `INV-YYYYMM-####`.
- Invoice currency is `GBP`.
- Billing quantities are integer minutes internally and displayed as hours/minutes.

## Acceptance Criteria

- Manager users can navigate to `/staff/manager/invoice-run` from the sidebar and from the dashboard quick action.
- Practitioner and Parent users do not see the invoice-run sidebar link and cannot access the route through role guards.
- Opening the page defaults the month selector to the most recent completed billing month. For a current instant in June 2026, the selected month is `2026-05`.
- Changing the month reloads mock preflight state for that month and resets generated drafts, selection, confirmation, and issue result state for the new month.
- The preflight summary shows total children, eligible children, blocked children, included sessions, attended/funded/billable quantity, funded deduction, and total due in manager-readable formats.
- If some children are blocked, the page still presents draft generation for eligible children as the default next action.
- The exception list shows each blocked child, one or more stable blocker reasons, and a next-action link to attendance corrections, child detail, or funding review.
- Clicking `Generate draft invoices` creates or updates mock drafts only for eligible child-months and keeps blocked child-months visible as exceptions.
- Draft review shows child, draft status, attended quantity, funded-hours deduction, extras placeholder amount, subtotal, net due, and an expandable line-level detail area.
- Draft review has no invoice-line edit controls.
- Bulk issue is presented as the primary action after drafts exist. All ready draft invoices are selected by default.
- Managers can deselect individual ready drafts before opening bulk issue confirmation.
- Bulk issue confirmation states the selected invoice count, selected total, billing month, and that issue locks invoices.
- Confirming bulk issue assigns mock invoice numbers and shows an issue result summary.
- Each ready draft row exposes one-by-one issue fallback with explicit confirmation.
- One-by-one issue updates only that draft and adds it to the issue result summary.
- If no children are eligible, draft generation is disabled or produces a clear empty outcome and no issue action is available.
- Loading, empty, error, and confirmation states are accessible by keyboard and do not overlap or clip on common desktop and mobile widths.
- Component and route tests cover the high-risk state transitions and role navigation behavior.

## Implementation Tasks

### Task 1: Add Invoice-Run Models and Formatting Utilities

- Objective: Define frontend invoice-run state in one place and keep it close to the FE-21 API contract.
- Depends on: Existing OpenAPI invoice schemas and existing staff feature model style.
- Target files/symbols:
  - Create `web/src/app/features/staff/models/invoice-run.models.ts`.
  - Optionally create `web/src/app/features/staff/utils/invoice-run-formatters.ts` if helpers would make tests clearer.
- Required changes:
  - Define types for `InvoiceRunStep`, `InvoiceRunBlockerCode`, `InvoiceRunBlocker`, `InvoiceRunException`, `InvoiceRunPreflightSummary`, `InvoiceRunEligibleChild`, `InvoiceRunPreflight`, `InvoiceDraftLine`, `InvoiceDraftReviewItem`, `DraftGenerationResult`, `IssueSelection`, `IssueResultSummary`, `IssuedInvoiceResult`, and `IssueException`.
  - Include enough fields to mirror API concepts: `billingMonth`, `currencyCode`, period dates, summary counts/minutes/minor-unit totals, eligible children, blocked children, invoice id, child id/name, status, calculation quantities, lines, and issue result data.
  - Define `formatGbp(minorUnits: number)`, `formatMinutes(minutes: number)`, and `formatBillingMonthLabel(month: string)` or reuse existing `formatGbp` from `manager-dashboard.models.ts` by moving it to a shared staff utility if that reduces duplication without broad churn.
  - Define `defaultCompletedBillingMonth(now = new Date()): string` using `Europe/London` current year/month and returning the previous calendar month in `YYYY-MM`.
  - Define a blocker-to-next-action mapping that returns label plus route/query params for:
    - incomplete attendance or missing check-out -> `/staff/manager/attendance-corrections`
    - missing funding profile -> child detail route with `billing_month` query param, or funding overview route when no child id is present
    - missing enrollment or missing core hourly rate -> child detail route
    - existing issued invoice -> no inline fix; show a non-editable explanatory action label
- Tests/verification:
  - Add focused unit tests for `defaultCompletedBillingMonth`, including June to May and January to previous December.
  - Add tests for GBP/minutes/month formatting and blocker next-action mapping.
- Expected outcome: The component can consume strongly typed mock state, and FE-21 can map API responses into the same view models.

### Task 2: Add Mock Invoice-Run Data Service

- Objective: Provide deterministic mock preflight, draft generation, bulk issue, and single issue operations for FE-20.
- Depends on: Task 1.
- Target files/symbols:
  - Create `web/src/app/features/staff/data/invoice-run-mock.service.ts`.
  - Optionally create `web/src/app/features/staff/data/invoice-run-mock.fixtures.ts` if fixture data is large enough to keep separate.
  - Add `web/src/app/features/staff/data/invoice-run-mock.service.spec.ts`.
- Required changes:
  - Implement injectable service methods:
    - `loadPreflight(billingMonth: string): Observable<InvoiceRunPreflight>`
    - `generateDrafts(billingMonth: string, childIds?: string[]): Observable<DraftGenerationResult>`
    - `listDrafts(billingMonth: string): Observable<InvoiceDraftReviewItem[]>`
    - `bulkIssue(billingMonth: string, invoiceIds: string[]): Observable<IssueResultSummary>`
    - `issueOne(invoiceId: string): Observable<IssueResultSummary>`
    - `resetMonth(billingMonth: string): void` if needed by the component when changing months.
  - Seed at least three month scenarios:
    - Default recent completed month with eligible children, blocked children, and no issued result initially.
    - A month with all children eligible.
    - A month with no eligible children and at least one blocker.
  - In the default scenario, include at least:
    - 4 eligible child-months.
    - 2 blocked child-months, one attendance blocker and one funding/enrollment blocker.
    - A zero-total eligible draft to prove zero-total invoices remain issuable.
    - One generated draft with `action = updated` to demonstrate idempotent regeneration.
  - Generate draft rows from eligible preflight data with lines:
    - core childcare line
    - funded-hours deduction line, including a zero-value deduction line when no deduction applies
    - zero-value manual extras placeholder line
  - Assign invoice numbers deterministically in child-name order on issue using `INV-YYYYMM-####`. Use stable sequence values in fixtures so tests can assert exact numbers.
  - Treat deselected drafts as skipped by omission from `bulkIssue`, not as failures.
  - Do not mutate issued drafts back to draft when generating again in the same mock session.
- Tests/verification:
  - Service tests assert partial preflight data, draft generation for eligible children only, deterministic line totals, deterministic issue numbering, zero-total issue support, and one-by-one issue behavior.
- Expected outcome: The component can exercise the full user workflow without any backend dependency.

### Task 3: Create Manager Invoice Run Page Component

- Objective: Build the guided workflow surface and state transitions.
- Depends on: Tasks 1 and 2.
- Target files/symbols:
  - Create `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.ts`.
  - Create `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.html` if the template is too large for inline readability; follow nearby page conventions.
  - Create `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.spec.ts`.
- Required changes:
  - Import and use existing shared components:
    - `PageHeaderComponent`
    - `AlertComponent`
    - `TableShellComponent`
    - `EmptyStateComponent`
    - `LoadingStateComponent`
    - `StatusBadgeComponent`
    - `ButtonComponent`
    - `ConfirmationDialogComponent`
  - Render a page header titled `Invoice run` with concise description using nursery billing terms.
  - Render an input of `type="month"` labelled `Billing month`.
  - Initialize the selected month with `defaultCompletedBillingMonth()`.
  - On month change, reload preflight and clear draft review, selected invoice ids, open confirmations, and issue result summary.
  - Render a guided progress indicator using compact step labels:
    - Preflight
    - Drafts
    - Review
    - Issue
    - Result
  - Render preflight summary metrics with stable labels and no spreadsheet language:
    - Children in month
    - Ready for draft
    - Exceptions
    - Sessions included
    - Attended time
    - Funded deduction
    - Estimated total
  - Render exceptions in a table/list with child, issue, detail, and next action. Keep the list visible after draft generation and issue.
  - Render `Generate draft invoices` as the primary action when `eligibleChildren.length > 0` and drafts have not yet been generated for the selected month.
  - Disable draft generation while loading or when no eligible children exist.
  - After generation, show generation result counts and a toast or inline success alert.
  - Render draft review rows with:
    - selection checkbox for ready drafts
    - child name
    - status
    - attended quantity
    - funded deduction amount
    - extras placeholder amount
    - net due
    - expand/collapse calculation detail
    - one-by-one issue action
  - Use stable dimensions for checkboxes, action buttons, summary tiles, and row controls so selection/expansion does not shift the layout.
  - Expanded detail must show line descriptions, quantities, unit amount where present, line amount, and calculation quantities. It must not render editable inputs.
  - Select all ready drafts by default immediately after draft generation.
  - Allow manager to select/deselect individual ready drafts and select/deselect all ready drafts.
  - Render bulk issue as the visually primary issue path. Disable it when no ready draft is selected.
  - Use `ConfirmationDialogComponent` for bulk and single issue confirmation.
  - Bulk confirmation content must show selected invoice count, selected total, billing month label, and immutable issue warning.
  - Single issue confirmation content must show the child name, draft total, and immutable issue warning.
  - After issue, render result summary with:
    - issued count
    - selected total issued
    - billing month
    - issued invoice list with child, invoice number, issue time, and total
    - skipped/failed list when present
  - Keep one-by-one issue available for drafts not yet issued after a bulk run.
  - Use router links for exception next actions. Do not implement inline attendance/funding/enrollment editors.
  - Include loading and error states even though the mock service is local.
- Tests/verification:
  - Component tests should cover:
    - default billing month is most recent completed month by stubbing utility or service clock
    - month change reloads and resets draft/issue state
    - preflight summary and exception next actions render
    - draft generation proceeds when blockers exist
    - draft review shows calculation details and no edit controls
    - all ready drafts are selected by default
    - deselecting a draft changes selected count and total
    - bulk confirmation opens with correct count/total
    - confirming bulk issue shows assigned invoice numbers
    - one-by-one issue opens confirmation and issues only that draft
    - no eligible children disables generation and hides issue actions
- Expected outcome: Managers can complete the whole FE-20 workflow with deterministic mock data and clear state transitions.

### Task 4: Register Route, Sidebar Link, and Dashboard Quick Action

- Objective: Make the invoice run workflow a first-class manager surface.
- Depends on: Task 3.
- Target files/symbols:
  - `web/src/app/core/constants/roles.ts`
  - `web/src/app/app.routes.ts`
  - `web/src/app/app.routes.spec.ts`
  - `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts`
  - `web/src/app/shared/layout/app-sidebar/app-sidebar.component.spec.ts`
  - `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.models.ts`
  - `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.component.spec.ts`
- Required changes:
  - Add `managerInvoiceRun: '/staff/manager/invoice-run'` to `ROLE_ROUTES`.
  - Import `ManagerInvoiceRunComponent` into `app.routes.ts`.
  - Add route:
    - path `staff/manager/invoice-run`
    - component `ManagerInvoiceRunComponent`
    - `canActivate: [authGuard, roleGuard]`
    - `data: { roles: ['manager'] }`
    - title `Invoice Run | Nursery Management`
  - Add sidebar manager nav item:
    - label `Invoice run`
    - path `ROLE_ROUTES.managerInvoiceRun`
    - test id `staff-link-manager-invoice-run`
  - Keep the link hidden for Practitioner and Parent roles.
  - Update dashboard mock quick action `Start invoice run` to route to `ROLE_ROUTES.managerInvoiceRun` or the literal path if avoiding an import cycle.
  - Update dashboard tests so `Start invoice run` is no longer counted as disabled and appears as an enabled router link.
- Tests/verification:
  - Route spec asserts the new path exists and has manager-only roles.
  - Sidebar spec asserts manager sees invoice-run link, and practitioner/parent do not.
  - Dashboard spec asserts quick action links to `/staff/manager/invoice-run`.
- Expected outcome: The new page is reachable by Manager users through normal navigation and blocked for other roles.

### Task 5: Align Mock UI With FE-21 API Integration Boundaries

- Objective: Reduce rework when replacing mock data with real invoice endpoints.
- Depends on: Tasks 1 through 4.
- Target files/symbols:
  - `web/src/app/features/staff/models/invoice-run.models.ts`
  - `web/src/app/features/staff/data/invoice-run-mock.service.ts`
  - `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.ts`
- Required changes:
  - Keep method names and return shapes close to future API operations:
    - preflight
    - generate drafts
    - list draft invoices
    - issue single invoice
    - issue bulk invoices
  - Keep mock blocker codes stable and compatible with known backend concepts:
    - `incomplete_attendance`
    - `missing_funding_profile`
    - `missing_core_hourly_rate`
    - `missing_guardian_link`
    - `existing_issued_invoice`
  - Keep status values compatible with invoice lifecycle: `draft`, `issued`, `payment_failed`, `paid`, `overdue`.
  - Preserve field names in view models that correspond directly to API fields, using frontend camelCase.
  - Add comments only where they explain the mock/API boundary, for example a short note above the mock service stating FE-21 replaces it with real `/api/v1` calls.
  - Do not add HTTP calls to `StaffApiService` during FE-20. Leave that for FE-21 so this ticket remains mock-data scoped.
- Tests/verification:
  - Existing mock-service and component tests should demonstrate that API-like responses can drive the UI without backend calls.
- Expected outcome: FE-21 can add API methods and mapping without redesigning FE-20 state or templates.

### Task 6: Accessibility, Responsive Layout, and Copy Pass

- Objective: Ensure the workflow is understandable and usable without spreadsheet context or visual breakage.
- Depends on: Tasks 3 and 4.
- Target files/symbols:
  - `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.ts`
  - Template file if separate.
- Required changes:
  - Use semantic sections with `aria-labelledby` for preflight, exceptions, draft review, issue selection, and results.
  - Ensure the month input has a visible label.
  - Ensure expand/collapse buttons have `aria-expanded` and descriptive accessible names.
  - Ensure checkboxes have accessible names that include child names.
  - Ensure confirmation dialogs have clear titles and no ambiguous destructive language.
  - Use `table` for dense draft/exception review where comparison matters, with mobile-safe overflow wrappers.
  - Keep copy in project terms: Manager, Child, Billing month, Invoice run, Draft invoice, Funded-hours deduction, Exception, Issue.
  - Do not use Customer, Product, Order, Sales, Revenue, plan names, or other TailAdmin/ecommerce copy.
  - Do not use visible instructional paragraphs that explain UI mechanics. Use labels, headings, table headers, button text, and concise state messages instead.
- Tests/verification:
  - Component tests assert banned ecommerce/template terms do not appear.
  - Manual validation at desktop and mobile widths checks no clipped text, overlapping controls, or hidden required actions.
- Expected outcome: The screen is operational, dense, and clear for a nursery manager.

### Task 7: Run Verification and Update Backlog Status Only if Implementation Is Completed

- Objective: Verify the FE-20 implementation and keep docs honest.
- Depends on: Tasks 1 through 6.
- Target files/symbols:
  - `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`
  - Any test files created or modified above.
- Required changes:
  - Run the exact commands in the Verification section.
  - If implementation and tests pass, update FE-20 in `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md` with done marker, date, and concise verified test/build summary matching prior backlog style.
  - If implementation is incomplete or tests fail for an unrelated existing reason, do not mark FE-20 done. Document the blocker in the final implementation response.
- Tests/verification:
  - Backlog update is done only after passing verification or after explicitly documenting any local command limitation.
- Expected outcome: FE-20 status reflects actual implementation state.

## Contracts

UI route and permissions:

- New route: `/staff/manager/invoice-run`.
- Allowed role: `manager`.
- Sidebar test id: `staff-link-manager-invoice-run`.
- Dashboard quick action label remains `Start invoice run` and links to the new route.

Mock data and state:

- Billing month format: `YYYY-MM`.
- Currency: `GBP`.
- Money values: integer minor units in state; display as `£x.xx`.
- Quantity values: integer minutes in state; display as hours/minutes.
- Default month: most recent completed billing month using `Europe/London`.
- Draft line kinds to support: `core_childcare`, `funded_deduction`, `extra`.
- Invoice statuses to support in FE-20 mock UI: `draft`, `issued`.
- Blocker codes must be stable strings and mapped to manager-readable labels and next actions.

Workflow states:

- Preflight loaded.
- Draft generation pending/succeeded/failed.
- Draft review available.
- Bulk confirmation open/closed.
- Single confirmation open/closed.
- Issue result available.
- Month change resets draft generation and issue state.

Error handling:

- Mock-service failures should be displayed through `AlertComponent` with a clear message.
- Unknown error text should remain generic and non-technical.
- No request id is expected for FE-20 mock failures.

API integration boundary for FE-21:

- FE-20 must not call real endpoints, but its data shapes must map cleanly to:
  - `GET /api/v1/invoices/drafts/preflight`
  - `POST /api/v1/invoice-runs/drafts`
  - `GET /api/v1/invoices?billing_month=YYYY-MM&status=draft`
  - `GET /api/v1/invoices/{invoice_id}`
  - `POST /api/v1/invoices/{invoice_id}/issue`
  - `POST /api/v1/invoices/bulk-issue`

## Files to Change

Create:

- `web/src/app/features/staff/models/invoice-run.models.ts`
- `web/src/app/features/staff/data/invoice-run-mock.service.ts`
- `web/src/app/features/staff/data/invoice-run-mock.service.spec.ts`
- `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.ts`
- `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.spec.ts`
- Optional if keeping the component smaller:
  - `web/src/app/features/staff/pages/manager-invoice-run/manager-invoice-run.component.html`
  - `web/src/app/features/staff/data/invoice-run-mock.fixtures.ts`
  - `web/src/app/features/staff/utils/invoice-run-formatters.ts`
  - `web/src/app/features/staff/utils/invoice-run-formatters.spec.ts`

Modify:

- `web/src/app/core/constants/roles.ts`
- `web/src/app/app.routes.ts`
- `web/src/app/app.routes.spec.ts`
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts`
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.spec.ts`
- `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.models.ts`
- `web/src/app/features/staff/pages/manager-dashboard/manager-dashboard.component.spec.ts`
- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md` only after implementation and verification pass.

Already modified during planning:

- `CONTEXT.md` was updated with the FE-20 domain decisions listed in Context Alignment.

## Verification

Run from the repository root:

```sh
cd web
npm test -- --watch=false --browsers=ChromeHeadless
npm run build
```

Optional static checks from the repository root:

```sh
git diff --check
rg -n "Ecommerce|Sales|Orders|Customers|Products|Starter Plan|Pro Plan|Enterprise Plan" web/src/app/features/staff/pages/manager-invoice-run web/src/app/features/staff/data/invoice-run-mock.service.ts
```

Expected automated coverage:

- Route registration and manager-only role metadata for `/staff/manager/invoice-run`.
- Sidebar visibility for Manager and hidden state for Practitioner/Parent.
- Dashboard quick action enabled for invoice run.
- Default completed billing month utility.
- Preflight summary and exception rendering.
- Partial-success draft generation with blockers still visible.
- Draft review calculation details and no line-edit controls.
- Bulk issue default selection and deselection.
- Bulk issue confirmation and result summary.
- One-by-one issue fallback.
- Empty/no-eligible state.

Manual validation scenarios:

- As a Manager, open `/staff/manager/invoice-run`, confirm the default selected month is the previous completed month, generate drafts, review details, bulk issue, and see invoice numbers.
- Change to a month with no eligible children and confirm issue controls are unavailable.
- Deselect one draft before bulk issue and confirm only selected drafts are issued.
- Use one-by-one issue on a remaining draft.
- Follow exception next-action links and confirm they navigate to existing manager workflows.
- Confirm Practitioner and Parent roles do not show the invoice-run link.
- Check the page at desktop, tablet, and mobile widths for overlapping text, clipped buttons, and inaccessible required actions.

## Assumptions

- FE-20 remains mock-data-only even though backend invoice endpoints already exist.
- Existing shared UI primitives from FE-03 are sufficient; no new global design system primitives are required.
- Direct child detail with `billing_month` query param is an acceptable next action for a child-specific missing funding profile because FE-18/FE-19 already use child detail for funded-hours editing.
- The mock service can use deterministic synchronous `of(...)` observables; adding artificial delay is unnecessary unless the component needs to exercise loading state in tests.
- The initial implementation can keep all invoice-run UI in one page component if tests remain readable. Extract child components only if the template becomes difficult to maintain.
- Full manager invoice detail navigation is deferred to FE-22, so draft review expansion inside FE-20 is enough for calculation inspection.

## Risks and Fallbacks

- Risk: The component becomes too large because it covers preflight, draft review, selection, confirmation, and results.
  - Fallback: Extract local standalone child components under `manager-invoice-run/` for summary tiles, exception list, draft table, and result summary. Keep state orchestration in the page component.
- Risk: Current shared `StatusBadgeComponent` does not recognize every invoice-run display status.
  - Fallback: Use existing recognized invoice statuses where possible (`draft`, `issued`) and render run-specific labels as plain text badges local to the component.
- Risk: Browser month input has inconsistent display formatting across platforms.
  - Fallback: Keep `type="month"` for native selection and display a separate formatted billing-month label beside summaries.
- Risk: `Europe/London` default-month calculation is awkward around month boundaries in local browser time.
  - Fallback: Use `Intl.DateTimeFormat` with `timeZone: 'Europe/London'` to derive the current London year/month before subtracting one month, and cover January rollover in tests.
- Risk: Existing dashboard tests assume exactly two disabled future actions.
  - Fallback: Update the assertion to check `Review payment follow-up` remains disabled and `Start invoice run` is a live link.
- Risk: Existing test environment lacks ChromeHeadless locally.
  - Fallback: Run `npm run build` and the focused TypeScript/Jasmine tests available in the environment, then report the missing browser dependency explicitly in the implementation final response.
