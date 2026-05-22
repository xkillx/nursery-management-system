# Staff UI Basics Implementation Plan

## Goal and Scope

Build the Day 6 "Staff UI basics" slice in `web/` with real API wiring and role-aware behavior, covering:

- Sign-in integrated with backend auth
- Manager child list + child create/edit
- Manager guardian list + guardian create/edit
- Practitioner read-only attendance child list

Out of scope for this slice:

- Attendance check-in/check-out writes (Day 10)
- Relationship/lifecycle action buttons (deactivate/end link/reactivate/mark inactive)
- Multi-membership login selector UX
- Parent billing/portal screens

## Context Alignment Notes (reconciled with `CONTEXT.md`)

- **Staff UI basics** means manager/practitioner screens with policy-correct permissions, matching backlog Day 6 (`docs/MVP-30D-BACKLOG.md`).
- **Child and Guardian Write Authority (MVP)**: child/guardian writes are manager-only; practitioner is read-only.
- **Practitioner Contact Visibility (MVP)**: practitioner attendance views must not expose guardian contact details.
- **Session Scope Mode (MVP)** + **Session Membership Binding (MVP)**: the UI must operate against one active membership scope returned by auth.
- **Session CSRF Mechanism (MVP)**: cookie-backed session actions (`refresh/logout`) require double-submit CSRF header.
- **API Error Contract (MVP)**: UI should handle stable `code` values and `request_id` from API errors.
- **Child and Guardian Default Listing Scope (MVP)**: manager list views default to active records and support explicit inactive/all views.

## Decisions That Must Be Honored

1. Deliver exactly four Day 6 screens:
   - Sign-in
   - Manager children (list + create/edit)
   - Manager guardians (list + create/edit)
   - Practitioner attendance child list (read-only)
2. Practitioner Day 6 is strictly read-only; attendance write actions are deferred.
3. Login UX targets single-membership happy path; explicit multi-membership selector is deferred.
4. Access token is in-memory only; no `localStorage`/`sessionStorage` for bearer token persistence.
5. Session bootstrap uses `POST /api/v1/auth/refresh` with cookie + CSRF header.
6. Route-level guards must enforce role access, and UI must also hide unauthorized menus/actions.
7. Default post-login routing:
   - manager -> manager child list
   - practitioner -> practitioner attendance child list
   - parent -> neutral placeholder/non-staff page (no manager/practitioner actions)
8. Manager forms must include API-validated fields now (not name-only stubs).
9. Manager lists must support `status` filter and pagination (`limit`/`offset`) with active default.
10. Error handling must be code-aware for key auth/authorization/validation codes.
11. Staff routes should live under a dedicated `/staff/...` route group; keep template routes intact.
12. Add a minimal automated frontend test baseline (auth service, guards, role-based nav visibility) plus manual smoke checks.

## Step-by-Step Tasks With Dependencies

### 1) Add frontend environment + API configuration (foundation)
**Depends on:** none

Tasks:
- Create Angular environment files with `apiBaseUrl`.
- Wire environment file replacement in `web/angular.json` for production/development.
- Create a small API URL helper so components do not hardcode endpoint strings.

Files to create/change:
- Create `web/src/environments/environment.ts`
- Create `web/src/environments/environment.development.ts`
- Change `web/angular.json`
- Create `web/src/app/core/config/api.config.ts` (or equivalent)

---

### 2) Implement auth/session core (in-memory token + cookie session actions)
**Depends on:** Step 1

Tasks:
- Create strongly typed models for auth response and API error shape.
- Implement `AuthService` with:
  - `login(email, password, membershipId?)`
  - `refresh()`
  - `logout()`
  - in-memory access token state (signal/observable)
  - active membership + role state
- Implement CSRF cookie reader utility and apply `X-CSRF-Token` header for `refresh/logout`.
- Ensure `withCredentials: true` is sent for cookie-based auth calls.
- Add bootstrap session restore on app startup (attempt refresh once).
- Provide `Authorization: Bearer` automatically to protected API calls.

Files to create/change:
- Create `web/src/app/core/models/auth.models.ts`
- Create `web/src/app/core/models/api-error.models.ts`
- Create `web/src/app/core/utils/cookie.util.ts`
- Create `web/src/app/core/services/auth.service.ts`
- Create `web/src/app/core/http/auth.interceptor.ts`
- Change `web/src/app/app.config.ts` (provide `HttpClient`, interceptor chain, app initializer)

---

### 3) Implement route guards and role gate utilities
**Depends on:** Step 2

Tasks:
- Add authentication guard for protected staff routes.
- Add role guard for manager-only/practitioner-only route access.
- Enforce direct URL protection even when navigation links are hidden.

Files to create/change:
- Create `web/src/app/core/guards/auth.guard.ts`
- Create `web/src/app/core/guards/role.guard.ts`
- Create `web/src/app/core/constants/roles.ts`

---

### 4) Replace template sign-in behavior with real auth flow
**Depends on:** Steps 2-3

Tasks:
- Remove social/OAuth placeholder behavior from sign-in form UX for this slice.
- Bind sign-in fields correctly using existing input components (`valueChange` outputs).
- On success: route by role decision table.
- On failure: map API error codes to field/global messages and include request id in support text.

Files to create/change:
- Change `web/src/app/shared/components/auth/signin-form/signin-form.component.ts`
- Change `web/src/app/shared/components/auth/signin-form/signin-form.component.html`
- Change `web/src/app/pages/auth-pages/sign-in/sign-in.component.ts` (if orchestration lives here)

---

### 5) Build staff API clients and domain models
**Depends on:** Steps 1-2

Tasks:
- Implement typed service(s) for:
  - `GET/POST/PATCH /children`
  - `GET/POST/PATCH /guardians`
  - `GET /children/attendance`
- Add shared query builder for `status`, `limit`, `offset`.
- Normalize API payload mapping to UI view models.

Files to create/change:
- Create `web/src/app/features/staff/data/staff-api.service.ts`
- Create `web/src/app/features/staff/models/children.models.ts`
- Create `web/src/app/features/staff/models/guardians.models.ts`
- Create `web/src/app/features/staff/models/attendance-child.models.ts`

---

### 6) Implement manager children screen (list + create/edit)
**Depends on:** Steps 3, 5

Tasks:
- Add manager children page under `/staff/manager/children`.
- Render list with:
  - default active status
  - status filter (`active|inactive|all`)
  - pagination controls (limit/offset, next/prev)
- Add create/edit form with required API fields:
  - `full_name`, `date_of_birth`, `start_date`, `core_hourly_rate_minor`
  - optional `end_date`, `notes`
- Display lifecycle/enrollment state read-only (e.g., `is_active`, `enrollment_complete`, missing requirements).

Files to create/change:
- Create `web/src/app/features/staff/pages/manager-children/manager-children.component.ts`
- Create `web/src/app/features/staff/pages/manager-children/manager-children.component.html`
- Create `web/src/app/features/staff/components/child-form/child-form.component.ts`
- Create `web/src/app/features/staff/components/child-form/child-form.component.html`
- Optionally create `web/src/app/features/staff/components/list-pagination/list-pagination.component.ts`

---

### 7) Implement manager guardians screen (list + create/edit)
**Depends on:** Steps 3, 5

Tasks:
- Add manager guardians page under `/staff/manager/guardians`.
- Render list with same status + pagination behavior as children list.
- Add create/edit form:
  - required: `full_name`
  - optional: `email`, `phone`, `notes`
- Show guardian lifecycle state read-only (`is_active`, `deactivated_at`, reason fields) without action buttons.

Files to create/change:
- Create `web/src/app/features/staff/pages/manager-guardians/manager-guardians.component.ts`
- Create `web/src/app/features/staff/pages/manager-guardians/manager-guardians.component.html`
- Create `web/src/app/features/staff/components/guardian-form/guardian-form.component.ts`
- Create `web/src/app/features/staff/components/guardian-form/guardian-form.component.html`

---

### 8) Implement practitioner attendance child list (read-only)
**Depends on:** Steps 3, 5

Tasks:
- Add practitioner page under `/staff/practitioner/attendance-children`.
- Fetch from `GET /children/attendance`.
- Show only attendance-facing child fields (`id`, `full_name`, enrollment flag).
- Do not show guardian contact data and do not render write actions.

Files to create/change:
- Create `web/src/app/features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component.ts`
- Create `web/src/app/features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component.html`

---

### 9) Add staff route tree + role-aware navigation
**Depends on:** Steps 3, 6, 7, 8

Tasks:
- Define `/staff/...` routes and attach guards:
  - manager pages -> manager role guard
  - practitioner page -> practitioner or manager if desired by policy
- Keep template/demo routes present but non-default.
- Update sidebar/header/user menu to:
  - display role-appropriate staff links
  - hide unauthorized links/actions
  - support sign out wired to real logout
- Ensure parent role sees no staff actions.

Files to create/change:
- Change `web/src/app/app.routes.ts`
- Change `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts`
- Change `web/src/app/shared/layout/app-sidebar/app-sidebar.component.html`
- Change `web/src/app/shared/components/header/user-dropdown/user-dropdown.component.ts`
- Change `web/src/app/shared/components/header/user-dropdown/user-dropdown.component.html`

---

### 10) Standardize frontend API error mapping
**Depends on:** Steps 2, 5

Tasks:
- Create shared error mapper that handles these codes at minimum:
  - `unauthorized`
  - `forbidden_role`
  - `forbidden_role_unknown`
  - `forbidden_scope_selection`
  - `validation_error`
  - `internal_error`
- Implement behavior:
  - unauthorized -> clear auth + redirect sign-in
  - forbidden role -> access denied UI/state
  - validation -> field-level binding when `details.field` exists
  - others -> generic message + request id

Files to create/change:
- Create `web/src/app/core/errors/api-error.mapper.ts`
- Change feature page/components to consume mapped errors

---

### 11) Add tests (targeted baseline)
**Depends on:** Steps 2, 3, 9, 10

Tasks:
- Auth service unit tests:
  - login state set
  - refresh bootstrap success/failure
  - logout clears state
  - CSRF header attachment for session actions
- Guard tests:
  - unauthenticated blocked
  - wrong role blocked
  - correct role allowed
- Navigation/component test:
  - manager sees manager links
  - practitioner does not see manager-only links/actions

Files to create/change:
- Create `web/src/app/core/services/auth.service.spec.ts`
- Create `web/src/app/core/guards/auth.guard.spec.ts`
- Create `web/src/app/core/guards/role.guard.spec.ts`
- Create one feature/nav spec file (e.g., `web/src/app/shared/layout/app-sidebar/app-sidebar.component.spec.ts`)

---

### 12) Execute verification and polish
**Depends on:** Steps 1-11

Tasks:
- Fix failing tests/build issues.
- Run manual role walkthrough with API.
- Confirm no practitioner exposure of guardian contact fields.
- Confirm role guard + backend 403 behavior are both coherent.

## Files to Create or Change (Consolidated)

Likely new files:

- `web/src/environments/environment.ts`
- `web/src/environments/environment.development.ts`
- `web/src/app/core/config/api.config.ts`
- `web/src/app/core/models/auth.models.ts`
- `web/src/app/core/models/api-error.models.ts`
- `web/src/app/core/utils/cookie.util.ts`
- `web/src/app/core/services/auth.service.ts`
- `web/src/app/core/http/auth.interceptor.ts`
- `web/src/app/core/guards/auth.guard.ts`
- `web/src/app/core/guards/role.guard.ts`
- `web/src/app/core/constants/roles.ts`
- `web/src/app/core/errors/api-error.mapper.ts`
- `web/src/app/features/staff/data/staff-api.service.ts`
- `web/src/app/features/staff/models/children.models.ts`
- `web/src/app/features/staff/models/guardians.models.ts`
- `web/src/app/features/staff/models/attendance-child.models.ts`
- `web/src/app/features/staff/pages/manager-children/manager-children.component.ts`
- `web/src/app/features/staff/pages/manager-children/manager-children.component.html`
- `web/src/app/features/staff/pages/manager-guardians/manager-guardians.component.ts`
- `web/src/app/features/staff/pages/manager-guardians/manager-guardians.component.html`
- `web/src/app/features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component.ts`
- `web/src/app/features/staff/pages/practitioner-attendance-children/practitioner-attendance-children.component.html`
- `web/src/app/features/staff/components/child-form/child-form.component.ts`
- `web/src/app/features/staff/components/child-form/child-form.component.html`
- `web/src/app/features/staff/components/guardian-form/guardian-form.component.ts`
- `web/src/app/features/staff/components/guardian-form/guardian-form.component.html`

Likely changed files:

- `web/angular.json`
- `web/src/app/app.config.ts`
- `web/src/app/app.routes.ts`
- `web/src/app/shared/components/auth/signin-form/signin-form.component.ts`
- `web/src/app/shared/components/auth/signin-form/signin-form.component.html`
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.ts`
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.html`
- `web/src/app/shared/components/header/user-dropdown/user-dropdown.component.ts`
- `web/src/app/shared/components/header/user-dropdown/user-dropdown.component.html`

Test files to add:

- `web/src/app/core/services/auth.service.spec.ts`
- `web/src/app/core/guards/auth.guard.spec.ts`
- `web/src/app/core/guards/role.guard.spec.ts`
- `web/src/app/shared/layout/app-sidebar/app-sidebar.component.spec.ts`

## Verification Steps

### Automated checks

From `web/`:

```bash
npm install
npm run build
npm test -- --watch=false
```

### Manual validation matrix

1. **Sign-in (manager)**
   - Valid credentials login succeeds.
   - Lands on manager children page.
2. **Sign-in (practitioner)**
   - Valid credentials login succeeds.
   - Lands on practitioner attendance child list.
3. **Route guard checks**
   - Practitioner blocked from manager URLs.
   - Unauthenticated user redirected from `/staff/*`.
4. **Manager children**
   - List loads with active default.
   - Status filter and pagination work.
   - Create and edit child succeed with API-valid payload.
5. **Manager guardians**
   - List/filter/pagination works.
   - Create and edit guardian succeed.
6. **Practitioner attendance list**
   - Read-only list loads.
   - No guardian email/phone displayed.
   - No write buttons present.
7. **Session handling**
   - App refresh restores session through `/auth/refresh` when cookies are valid.
   - Logout invalidates session and returns to sign-in.
8. **Error contract handling**
   - `validation_error` shows field/global feedback.
   - `forbidden_role` shows access denied behavior.
   - Unknown/internal errors surface generic message with `request_id`.

### API interaction checks (optional but recommended)

- Confirm browser requests include `Authorization: Bearer <access_token>` for protected endpoints.
- Confirm `refresh/logout` include `X-CSRF-Token` and send cookies (`withCredentials`).

## Explicit Assumptions Made to Unblock Execution

1. API and web app are served from same host in local/dev, or CORS is configured elsewhere; this plan does not add backend CORS middleware.
2. Valid seeded users exist for manager and practitioner roles for manual testing.
3. Parent role staff pages are intentionally out of scope in Day 6 and can route to non-staff placeholder.
4. Existing TailAdmin template pages remain present; MVP staff routes are introduced under `/staff/...` without deleting template routes.
5. Angular standalone architecture remains the project standard (no NgModule migration).
6. Route-level access control in frontend is UX/control flow support; backend authorization remains source-of-truth.
