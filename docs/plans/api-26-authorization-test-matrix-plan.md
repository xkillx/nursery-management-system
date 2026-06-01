# API-26 Authorization Test Matrix Implementation Plan

## Goal

Add a route-by-route authorization test matrix for all registered MVP API endpoints. The tests must prove that protected routes are default-deny, role-gated, tenant/branch scoped, and parent relationship scoped where applicable, while public routes remain intentionally public and protected by their route-specific controls.

## Non-Goals

- Do not redesign authorization, role inheritance, tenant scoping, or parent access semantics.
- Do not add new product endpoints.
- Do not fix unrelated known contract gaps unless a route denial code required by this matrix is unstable.
- Do not force public auth/session routes to require bearer authentication.
- Do not duplicate unsupported/forged-role coverage route-by-route; keep that at platform middleware level.

## Context Alignment

- `Membership` is the session-bound authorization unit. Each authenticated request acts through exactly one membership with one tenant, branch, and role.
- `Tenant` and branch scope are enforced from the active session membership on every protected business route.
- Parent invoice/payment access uses the documented two-hop model: active parent membership -> active parent membership guardian mapping -> active guardian-child link -> child invoice.
- No `CONTEXT.md` changes are required for this task because the interview resolved test strategy and route classification, not new domain glossary terms.

## Current State

- Server bootstrap is in `api/internal/app/bootstrap/bootstrap.go`.
- Shared auth middleware is in `api/internal/platform/http/authz_middleware.go`.
- Existing platform guard coverage is in `api/internal/platform/http/authz_middleware_test.go`.
- Existing route tests are split across:
  - `api/internal/app/bootstrap/people_routes_test.go`
  - `api/internal/app/bootstrap/attendance_routes_test.go`
  - `api/internal/app/bootstrap/funding_routes_test.go`
  - `api/internal/app/bootstrap/billing_routes_test.go`
  - `api/internal/app/bootstrap/billing_issue_test.go`
  - `api/internal/app/bootstrap/billing_parent_routes_test.go`
  - `api/internal/app/bootstrap/payments_routes_test.go`
  - `api/internal/app/bootstrap/manager_payment_diagnostics_test.go`
  - `api/internal/app/bootstrap/webhook_integration_test.go`
- Integration route tests use `dbtest.RequirePostgres(t)`, which skips when `TEST_DATABASE_URL` is not set and requires the `migrate` CLI when it runs.
- Test helper functions already exist in package `bootstrap`, including `testConfig`, `mustAccessToken`, `doRequest`, `requireStatus`, `requireErrorCode`, and `decodeJSON`.
- `BootstrapOptions` currently supports checkout and webhook test injection but not email sender injection. Current bootstrap always constructs `email.NewSMTPSender`, which makes invite/resend/password-reset success-path route tests depend on SMTP unless a fake sender option is added.
- `docs/API-CONTRACT-MVP.md` documents `/api/v1/authz/probe/*` as debug/integration-test helper routes, not frontend product routes.
- `docs/MVP-30D-API-BACKEND-BACKLOG.md` has stale expected-route entries for manager payment diagnostics at the bottom: `/api/v1/payments/events` and `/api/v1/invoices/:invoice_id/payments`. Current implemented and documented routes are `/api/v1/invoices/:invoice_id/payment-status` and `/api/v1/invoices/:invoice_id/payment-events`.

Current registered MVP route inventory to classify:

- Public/bootstrap: `GET /health`, `GET /api/v1/health`
- Public auth/session/account recovery: `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, `POST /api/v1/auth/switch-membership`, `POST /api/v1/auth/password-reset-requests`, `POST /api/v1/auth/password-resets`
- Public invite acceptance: `POST /api/v1/invites/accept`
- Public Stripe webhook: `POST /api/v1/stripe/webhooks`
- Protected diagnostic: `GET /api/v1/me`, `GET /api/v1/authz/probe/manager`, `GET /api/v1/authz/probe/practitioner`, `GET /api/v1/authz/probe/parent`, `GET /api/v1/authz/probe/scope/:tenant_id/:branch_id`, `GET /api/v1/authz/probe/parent-link/:child_id`
- Manager people: `GET /api/v1/children`, `GET /api/v1/children/:child_id`, `POST /api/v1/children`, `PATCH /api/v1/children/:child_id`, `POST /api/v1/children/:child_id/actions/mark-inactive`, `GET /api/v1/guardians`, `GET /api/v1/guardians/:guardian_id`, `POST /api/v1/guardians`, `PATCH /api/v1/guardians/:guardian_id`, `POST /api/v1/guardians/:guardian_id/actions/deactivate`, `POST /api/v1/guardians/:guardian_id/actions/reactivate`, `POST /api/v1/guardian-child-links`, `POST /api/v1/guardian-child-links/:link_id/actions/end`, `POST /api/v1/parent-membership-guardian-mappings`, `POST /api/v1/parent-membership-guardian-mappings/:mapping_id/actions/end`
- Manager/practitioner attendance: `GET /api/v1/children/attendance`, `POST /api/v1/attendance/check-ins`, `POST /api/v1/attendance/check-outs`, `POST /api/v1/attendance/absence-markers`, `POST /api/v1/attendance/absence-markers/:absence_marker_id/clear`
- Manager attendance correction: `POST /api/v1/attendance/corrections`
- Manager invites: `POST /api/v1/invites`, `GET /api/v1/invites`, `POST /api/v1/invites/:invite_id/resend`, `POST /api/v1/invites/:invite_id/revoke`
- Manager funding: `GET /api/v1/funding/children/:child_id`, `PUT /api/v1/funding/children/:child_id`
- Manager billing/invoices: `GET /api/v1/invoices/drafts/preflight`, `POST /api/v1/invoice-runs/drafts`, `GET /api/v1/invoices`, `GET /api/v1/invoices/:invoice_id`, `POST /api/v1/invoices/:invoice_id/issue`, `POST /api/v1/invoices/bulk-issue`
- Parent billing/payment: `GET /api/v1/parent/invoices`, `GET /api/v1/parent/invoices/:invoice_id`, `POST /api/v1/parent/invoices/:invoice_id/checkout-sessions`
- Manager payment diagnostics: `GET /api/v1/invoices/:invoice_id/payment-status`, `GET /api/v1/invoices/:invoice_id/payment-events`

## Decisions

- Cover all registered MVP routes, but classify public routes separately from protected routes.
- Public routes must prove intentional public access or their route-specific controls; they must not be expected to return `401` only because no bearer token is present.
- Protected business routes must cover unauthenticated, wrong valid MVP role, wrong tenant/branch where applicable, parent relationship failure where applicable, and allowed access.
- Wrong tenant/branch resource access must preserve current concealment behavior: `404 *_not_found` for scoped object access and empty/excluded results for lists. Use `403 forbidden_scope` only for explicit scope probe or scope-selection flows.
- Route-by-route wrong-role cases use only valid MVP roles: `manager`, `practitioner`, `parent`.
- Unsupported role values, such as `director`, remain covered once in `api/internal/platform/http/authz_middleware_test.go` with `403 forbidden_role_unknown`.
- Add a completeness guard that fails when any registered real MVP `/api/v1` route is not classified in the matrix.
- Include `/api/v1/authz/probe/*` as protected diagnostic routes in the classification and matrix.
- No ADR is needed: this is test coverage for existing ADR-backed authorization decisions, not a new hard-to-reverse architecture decision.

## Acceptance Criteria

- A route classification test fails if a real registered `/api/v1` route is missing from the matrix.
- Protected routes without bearer tokens return HTTP `401` with `code: "unauthorized"`.
- Protected routes called by a valid but disallowed MVP role return HTTP `403` with `code: "forbidden_role"`.
- Allowed role access reaches the route handler and returns the expected success status for each route where a deterministic success setup is practical.
- For manager-scoped object routes, other-tenant/branch resource IDs return the route's stable `*_not_found` code.
- For manager-scoped list routes, other-tenant/branch records are excluded from results.
- For protected create routes that accept foreign scoped IDs, cross-scope body IDs return stable not-found denial codes and do not create records.
- Parent invoice list/detail/checkout routes enforce the two-hop parent relationship at request time.
- Public routes are included in the route inventory as intentional-public routes and have tests proving they are not accidentally mounted under bearer auth.
- Existing platform middleware unsupported-role coverage remains green.
- `go test ./...` passes from `api/` without `TEST_DATABASE_URL` by skipping integration tests as today.
- With a valid `TEST_DATABASE_URL` and `migrate` CLI, the bootstrap route matrix tests run and pass.

## Implementation Tasks

Each task must include:

- Objective
- Depends on
- Target files/symbols
- Required changes
- Tests/verification
- Expected outcome

### Task 1: Add Fake Email Injection To Bootstrap Options

- Objective: Make public and manager invite/password-reset success-path route tests deterministic without SMTP.
- Depends on: Current `BootstrapOptions`.
- Target files/symbols:
  - `api/internal/app/bootstrap/bootstrap.go`
  - `BootstrapOptions`
  - `BootstrapWithOptions`
  - `email.NewSMTPSender`
- Required changes:
  - Add an optional `EmailSender email.Sender` field to `BootstrapOptions`.
  - In `BootstrapWithOptions`, choose `opts.EmailSender` when non-nil; otherwise construct the existing SMTP sender exactly as today.
  - Pass the selected sender into `resetapp.NewEmailAdapter` and `inviteapp.NewInviteEmailAdapter`.
  - Do not change production behavior when `Bootstrap` or `BootstrapWithOptions` is called without `EmailSender`.
- Tests/verification:
  - Existing compile checks.
  - New route matrix harness should pass `email.NewFakeSender()` through `BootstrapWithOptions`.
- Expected outcome: Tests can exercise invite create/resend and known-user password reset without network SMTP.

### Task 2: Create Authorization Matrix Test Harness

- Objective: Centralize route classification and reusable request builders for API-26.
- Depends on: Task 1.
- Target files/symbols:
  - Create `api/internal/app/bootstrap/authorization_matrix_test.go`
  - Existing helpers in `people_routes_test.go`
  - `dbtest.RequirePostgres`
  - `BootstrapWithOptions`
- Required changes:
  - Define an `authorizationMatrixHarness` with:
    - `router *gin.Engine`
    - `pool *pgxpool.Pool`
    - token manager
    - tenant/branch A and B
    - manager, practitioner, and parent tokens for scope A
    - manager and parent tokens for scope B where needed
    - seeded child, guardian, link, mapping, invoice, invite, funding, attendance, and absence IDs needed by route builders
  - Use `dbtest.RequirePostgres(t)` and `dbtest.Reset(t, pool)`.
  - Build the router with `BootstrapWithOptions(testConfig(), logger, pool, BootstrapOptions{EmailSender: email.NewFakeSender(), CheckoutProvider: fakeCheckoutProvider, WebhookVerifier: fakeWebhookVerifier if needed})`.
  - Prefer simple direct DB seed helpers over full workflow calls when a route needs existing state.
  - Keep route tests independent: either use fresh IDs per case or create state in each case's setup closure.
- Tests/verification:
  - `go test ./internal/app/bootstrap -run TestAuthorization -count=1` with `TEST_DATABASE_URL`.
- Expected outcome: A focused API-26 harness exists without disturbing existing domain-route tests.

### Task 3: Add Route Classification Completeness Guard

- Objective: Fail when new real MVP routes are registered without matrix classification.
- Depends on: Task 2.
- Target files/symbols:
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
  - `(*gin.Engine).Routes()`
- Required changes:
  - Define route classification entries with method, path, classification, allowed roles, and scenario applicability.
  - Include classifications:
    - `public`
    - `protected_diagnostic`
    - `protected_business`
  - Include root `GET /health` in public route assertions, but make the completeness guard focus on real `/api/v1` routes.
  - Compare `router.Routes()` against the classification map.
  - Fail on unclassified `/api/v1` routes.
  - Fail on stale matrix entries whose route is no longer registered.
  - Do not include package-local test-only routes from `api/internal/platform/http/*_test.go`; they are not registered by bootstrap.
- Tests/verification:
  - Add `TestAuthorizationMatrixRouteClassificationIsComplete`.
- Expected outcome: Future route additions cannot silently bypass API-26 classification.

### Task 4: Cover Public Routes As Intentional Public Routes

- Objective: Prove public routes are not protected by bearer auth and still return stable route-specific responses.
- Depends on: Task 2.
- Target files/symbols:
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
  - Public route handlers listed in Current State.
- Required changes:
  - Add public-route table tests that call each route without bearer auth.
  - Use malformed payloads where useful to prove the handler is reached before authn, for example:
    - `POST /api/v1/auth/login` with malformed or incomplete JSON expects `400 validation_error`.
    - `POST /api/v1/auth/password-reset-requests` with valid unknown email expects `202`.
    - `POST /api/v1/auth/password-resets` with invalid token and valid password expects `400 password_reset_token_invalid`.
    - `POST /api/v1/invites/accept` with invalid token and valid password expects `400 invite_token_invalid`.
    - `POST /api/v1/auth/logout` without refresh cookie expects `204`.
    - `POST /api/v1/auth/refresh` without refresh cookie expects `401 unauthorized` as session-cookie auth failure, not bearer auth.
    - `POST /api/v1/auth/switch-membership` with valid JSON but no refresh cookie expects `401 unauthorized`.
    - `POST /api/v1/stripe/webhooks` without bearer auth expects the existing webhook-specific response: `503 payment_provider_unconfigured` when no verifier is configured, or `400 payment_webhook_invalid_signature` when verifier/signature setup is used.
    - `GET /api/v1/health` expects `200`.
  - For cases returning `401 unauthorized`, assert the path is classified public so the test name makes clear this is route-specific cookie/session behavior.
- Tests/verification:
  - `TestAuthorizationMatrixPublicRoutesAreIntentional`.
- Expected outcome: Public routes are explicitly covered without weakening protected route checks.

### Task 5: Cover Protected Route Authentication And Role Matrix

- Objective: Prove all protected routes reject unauthenticated requests and wrong valid MVP roles.
- Depends on: Task 3.
- Target files/symbols:
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
  - All protected diagnostic and business routes listed in Current State.
- Required changes:
  - For each protected route classification entry, define:
    - method
    - path builder
    - body builder
    - allowed roles
    - expected allowed status or status set
    - disallowed valid role tokens
  - Add `TestAuthorizationMatrixProtectedRoutesRequireAuthentication`.
  - Add `TestAuthorizationMatrixProtectedRoutesRejectWrongRoles`.
  - Expected role rules:
    - `GET /api/v1/me`: manager, practitioner, parent allowed.
    - Authz probes: only their named role except scope probe allows all valid roles; parent-link probe requires parent.
    - Manager routes: manager allowed; practitioner and parent forbidden.
    - Manager/practitioner attendance and absence routes: manager and practitioner allowed; parent forbidden.
    - Parent routes: parent allowed; manager and practitioner forbidden.
  - For allowed status checks, avoid testing business-rule depth beyond proving the handler is reached and the route can succeed with seeded data.
- Tests/verification:
  - The wrong-role assertions must expect `403 forbidden_role`.
  - The unauthenticated assertions must expect `401 unauthorized`.
- Expected outcome: Every protected route has default-deny and wrong-role coverage in one central matrix.

### Task 6: Cover Tenant/Branch Scope Matrix

- Objective: Prove protected resource routes use session tenant/branch scope and avoid cross-scope existence leaks.
- Depends on: Task 5.
- Target files/symbols:
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
  - Existing repository-backed route handlers.
- Required changes:
  - Add `TestAuthorizationMatrixTenantBranchScope`.
  - For object routes with scoped path IDs, call the route with an allowed role in scope A and a scope B resource ID, expecting stable not-found codes:
    - child routes: `child_not_found`
    - guardian routes: `guardian_not_found`
    - guardian-child link end: `guardian_child_link_not_found`
    - parent mapping end: `parent_mapping_not_found`
    - funding child routes: `child_not_found`
    - invoice detail/issue/payment diagnostics: `invoice_not_found`
    - parent invoice detail/checkout for other-scope invoice: `invoice_not_found`
    - absence marker clear for other-scope marker: `absence_marker_not_found`
    - invite resend/revoke for other-scope invite: `invite_not_found`
  - For create routes that accept scoped IDs, submit scope B IDs in the body while authenticated in scope A and assert the relevant stable not-found code:
    - guardian-child link create with other-scope guardian or child
    - parent membership guardian mapping create with other-scope membership or guardian
    - attendance check-in/check-out/correction with other-scope child or session where practical
    - absence marker create with other-scope child
    - draft generation selected child IDs from another scope should not create scope A records and should report an exception or validation/not-found behavior matching current implementation.
  - For list routes, seed both scope A and scope B records and assert scope B records are excluded:
    - children
    - attendance child list
    - guardians
    - invites
    - invoice list
    - parent invoice list
    - payment events where applicable
  - Mark routes with no caller-supplied resource or foreign ID as not applicable for wrong-scope assertions, because they derive tenant and branch only from the session and create new scope A records.
  - Keep `403 forbidden_scope` expectations only for `GET /api/v1/authz/probe/scope/:tenant_id/:branch_id`.
- Tests/verification:
  - Assert both HTTP status and error `code` for object/body denial cases.
  - For list cases, decode JSON and assert excluded IDs are absent.
- Expected outcome: Cross-scope access is denied with current concealment semantics.

### Task 7: Cover Parent Relationship Matrix

- Objective: Prove parent-facing invoice/payment routes enforce the two-hop parent relationship.
- Depends on: Task 6.
- Target files/symbols:
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
  - `api/internal/app/bootstrap/billing_parent_routes_test.go` seed helpers can be reused or copied if they remain package-private.
  - `GET /api/v1/parent/invoices`
  - `GET /api/v1/parent/invoices/:invoice_id`
  - `POST /api/v1/parent/invoices/:invoice_id/checkout-sessions`
  - `GET /api/v1/authz/probe/parent-link/:child_id`
- Required changes:
  - Add `TestAuthorizationMatrixParentRelationship`.
  - Seed:
    - linked child with issued invoice for allowed parent access
    - unlinked child with issued invoice
    - linked child whose guardian-child link is ended
    - linked child whose parent-membership guardian mapping is ended
  - Assert:
    - parent invoice list includes linked issued invoices only.
    - parent invoice list excludes unlinked, ended-link, ended-mapping, draft, and wrong-scope invoices.
    - parent invoice detail for unlinked or relationship-ended invoice returns `404 invoice_not_found`.
    - checkout-session creation for unlinked or relationship-ended invoice returns `404 invoice_not_found`.
    - checkout-session creation for linked payable invoice reaches payment provider behavior; with fake checkout provider it should return `201`.
    - parent-link probe returns `403 forbidden_parent_child_link` for mismatched `linked_child_id` and `200` for matching query/path.
- Tests/verification:
  - Assert stable status/code pairs.
- Expected outcome: Parent relationship failure is covered explicitly for all parent-only business routes.

### Task 8: Update Existing Tests To Avoid Duplication Where Practical

- Objective: Keep the suite maintainable after adding central matrix coverage.
- Depends on: Tasks 4 to 7.
- Target files/symbols:
  - Existing route test files in `api/internal/app/bootstrap`.
- Required changes:
  - Do not remove valuable domain behavior tests.
  - If existing tests duplicate only the new matrix's unauthenticated/wrong-role assertion, either leave them for now or simplify only when doing so reduces obvious noise without losing domain context.
  - Keep existing business behavior tests, validation tests, idempotency tests, billing/payment state tests, and webhook processing tests.
- Tests/verification:
  - Full bootstrap package tests remain green with `TEST_DATABASE_URL`.
- Expected outcome: Central matrix exists without risky cleanup churn.

### Task 9: Align Backlog Route Documentation

- Objective: Remove stale route names discovered during API-26 exploration.
- Depends on: Task 3.
- Target files/symbols:
  - `docs/MVP-30D-API-BACKEND-BACKLOG.md`
- Required changes:
  - In the bottom "Expected API Routes" list, replace stale manager payment routes:
    - remove `GET /api/v1/payments/events`
    - remove `GET /api/v1/invoices/:invoice_id/payments`
    - add `GET /api/v1/invoices/:invoice_id/payment-status`
    - add `GET /api/v1/invoices/:invoice_id/payment-events`
  - Do not alter completed task history except the stale expected-route list.
- Tests/verification:
  - Documentation-only check.
- Expected outcome: Backlog route inventory matches current registered routes and API contract.

## Contracts

- Protected bearer-token auth contract:
  - Missing/invalid bearer token: HTTP `401`, code `unauthorized`.
  - Disallowed valid MVP role: HTTP `403`, code `forbidden_role`.
  - Unsupported role claim: covered once in platform middleware as HTTP `403`, code `forbidden_role_unknown`.
- Scope contract:
  - Resource routes conceal other tenant/branch records as not found or list exclusion.
  - Explicit scope probe returns HTTP `403`, code `forbidden_scope` for tenant/branch mismatch.
- Parent relationship contract:
  - Parent access requires active parent membership guardian mapping and active guardian-child link in the same tenant/branch.
  - Parent detail/checkout failures for unlinked invoices return HTTP `404`, code `invoice_not_found`.
  - Parent lists exclude inaccessible invoices.
- Public-route contract:
  - Public routes are classified and tested as intentional public routes.
  - Cookie-backed auth/session routes may still return `401 unauthorized` for missing refresh token; that does not make them bearer-protected routes.
  - Stripe webhook remains public at bearer-auth layer and protected by webhook signature/provider configuration.
- Error response shape remains `{ code, message, request_id, details? }`.

## Files to Change

- Create `api/internal/app/bootstrap/authorization_matrix_test.go`.
- Change `api/internal/app/bootstrap/bootstrap.go` to add optional fake email sender injection through `BootstrapOptions`.
- Optionally adjust existing bootstrap route tests only to reduce clear duplication.
- Update `docs/MVP-30D-API-BACKEND-BACKLOG.md` stale expected-route entries for manager payment diagnostics.
- Do not edit `CONTEXT.md` for API-26 unless a new domain term is introduced during implementation.
- Do not create an ADR for API-26.

## Verification

From the repository root:

```sh
cd api && go test ./...
```

Expected without `TEST_DATABASE_URL`:

- Unit tests run.
- Repository/bootstrap integration tests using `dbtest.RequirePostgres` skip as they do today.
- No compile errors from `BootstrapOptions` changes.

With a disposable PostgreSQL test database whose database name contains `test` or `repository`, and with `migrate` installed:

```sh
export TEST_DATABASE_URL='postgres://.../nursery_test?...'
cd api && go test ./internal/app/bootstrap -run 'TestAuthorizationMatrix|TestPeople|TestAttendance|TestFunding|TestBilling|TestInvoice|TestParent|TestPayments|TestManagerPayment|TestWebhook' -count=1
cd api && go test ./...
```

Expected with `TEST_DATABASE_URL`:

- Authorization matrix route classification is complete.
- All protected route authn/wrong-role tests pass.
- Scope and parent relationship matrix tests pass.
- Existing bootstrap integration tests remain green.

If `migrate` is missing, the failure should remain the existing `dbtest.RequirePostgres` failure message: `migrate binary not found in PATH; install golang-migrate CLI`.

## Assumptions

- API-26 is test hardening; it should preserve current runtime authorization behavior.
- The route matrix should follow current registered routes rather than stale route names in older backlog sections.
- `GET /health` remains public, but the strict completeness guard only needs to fail on unclassified `/api/v1` routes.
- Success-path tests may use direct database seeding instead of exercising entire workflows, because API-26 verifies authorization boundaries rather than domain calculations.
- Routes that create new resources without caller-supplied scoped IDs have wrong-scope marked not applicable, because tenant and branch come only from the session.
- Public route tests may assert validation or route-specific auth errors to prove handler reachability without bearer auth.

## Risks and Fallbacks

- Risk: Some allowed access cases mutate shared seeded state and make later matrix cases flaky.
  - Fallback: Use per-case setup closures and fresh deterministic UUIDs, or split destructive allowed cases into independent subtests with new harness setup.
- Risk: Invite/password-reset success tests fail due to SMTP.
  - Fallback: Implement Task 1 before public success tests and pass `email.NewFakeSender()` through `BootstrapOptions`.
- Risk: Some route-specific wrong-scope behavior is currently a business exception rather than a clean not-found.
  - Fallback: Preserve the current stable code if it is documented by existing tests; only fix behavior if it violates the decisions above or leaks cross-scope existence.
- Risk: The route completeness guard catches deferred/not-implemented contract routes.
  - Fallback: Only compare against `router.Routes()` from the actual bootstrapped API, not against proposed/deferred docs.
- Risk: Full `go test ./...` is slow or skipped without a database.
  - Fallback: Report both no-database compile/unit result and database-backed bootstrap result separately; do not remove `dbtest.RequirePostgres` skipping behavior.
