# Manager Child Registration/Enrolment Editor Implementation Plan

## Goal

Build a manager-only, child-scoped registration/enrolment editor for `FE-PM-01`. Managers should reach it from the existing child detail/enrollment surface, edit the full current registration profile in sectioned saves, edit the office-use enrolment checklist in the same workflow, and see separate completion statuses for the registration profile and office-use checklist.

The editor must support backfilling and ongoing correction of child/family care information, contact entries, collection-password metadata, funding-support notes, routine notes, GDPR declaration metadata, and office-use checklist metadata without changing existing child/guardian CRUD, MVP enrollment gates, attendance, funding, invoicing, parent invoice access, or portal-access relationships.

## Non-Goals

- Do not build the parent/guardian digital registration journey from `FE-PM-02`.
- Do not build the consent and acknowledgement ledger from `API-PM-02` / `FE-PM-03`; capture GDPR declaration metadata only.
- Do not upload, store, preview, or link document files for Red Book, birth certificate/passport, proof of address, contract, or other checklist evidence.
- Do not add a final submit, lock, approve, reopen, or historical version-browsing workflow.
- Do not display stored collection-password material after it is saved.
- Do not automatically create, update, deactivate, or link Guardian records from registration contact entries.
- Do not let office-use start-date or date-left fields update the child's operational enrollment start/end dates.
- Do not expose registration profile or office-use checklist data to practitioner attendance, parent invoice, parent portal, or owner surfaces.
- Do not make registration profile or office-use checklist completion block attendance, funding, invoice generation, or parent invoice access.

## Context Alignment

- `CONTEXT.md:39-41` defines the registration/enrolment profile as child-linked care/family information and separate from consent history and office-use checklist tracking.
- `CONTEXT.md:43-45` gives managers full profile maintenance access within their nursery site and keeps practitioner/parent access out of scope.
- `CONTEXT.md:47-53` defines registration profile completion and the combined registration/enrolment workflow as manager follow-up statuses that do not alter MVP enrollment gates.
- `CONTEXT.md:55-65` says section completion is review-based, supports explicit none/no/not-applicable answers, and treats blank/missing health, social-care, development, dietary, and medication topics as unknown/incomplete.
- `CONTEXT.md:67-73` keeps the profile current-editable and relies on audit records without manager-facing historical browsing.
- `CONTEXT.md:75-81` keeps child full name/date of birth on the child record and treats GDPR declaration metadata as who confirmed it, when confirmed, and the form declaration date.
- `CONTEXT.md:83-89` protects collection password display and keeps registration contacts separate from Guardian records, guardian-child links, parent invitations, parent memberships, and parent portal access.
- `CONTEXT.md:91-109` keeps funding-support notes informational, consent separate, office-use checklist conceptually separate but editable in the same workflow, date-left optional for current children, and office-use dates separate from operational enrollment dates.
- `CONTEXT.md:339-345` keeps child detail as the compact child basics/enrollment/funding summary and makes the manager registration/enrolment editor a dedicated child-scoped workflow reached from child detail.

## Current State

- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md:135` defines `FE-PM-01` as a manager child registration/enrolment editor with sectioned forms, partial section saves, completion state, and missing requirement review.
- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md:138` still lists `FE-PM-04` as a separate office-use checklist UI, but the product decision for this plan is to include office-use checklist editing in `FE-PM-01` while preserving a separate checklist status.
- `docs/MVP-30D-API-BACKEND-BACKLOG.md:198` says `API-PM-01` covers registration/enrolment profile data and includes office-use fields in the owner-provided form source.
- `docs/MVP-30D-API-BACKEND-BACKLOG.md:200` says `API-PM-03` covers office-use checklist metadata without file upload/storage. This plan pulls that metadata subset forward into the `FE-PM-01` implementation because office-use editing is now in scope.
- `docs/forms/README.md` normalizes the owner-provided paper form into child identity/demographics, medical profile, health contacts, social/development, parents/carers, funding/support notes, emergency/collection, routine notes, declarations/consent, and office-use checks.
- `docs/forms/child-application-form.md` is the raw form source for all in-scope registration and office-use fields.
- `docs/API-CONTRACT.openapi.yaml:4464-4564` already documents manager-only `GET`/`PATCH /api/v1/children/{child_id}/registration-profile` and `PUT /api/v1/children/{child_id}/registration-profile/collection-password`.
- `docs/API-CONTRACT.openapi.yaml:1530-1888` already defines registration profile child summary, completeness, contact, profile section, response, patch, and collection-password schemas. It does not define office-use checklist schemas or routes.
- `api/db/migrations/000016_add_child_registration_profiles.up.sql:10-122` creates one current `child_registration_profiles` row per child with profile fields, collection-password hash metadata, section review flags, and scope constraints.
- `api/db/migrations/000016_add_child_registration_profiles.up.sql:132-170` creates repeated `child_registration_contacts` rows for parent/carer, emergency contact, and authorised collector entries.
- `api/internal/modules/registrationprofiles/domain/profile.go:43-124` models the current registration profile and does not include office-use checklist fields.
- `api/internal/modules/registrationprofiles/domain/profile.go:126-163` models registration contact entries and collection-password metadata.
- `api/internal/modules/registrationprofiles/domain/completeness.go:45-87` computes registration profile completion across nine sections.
- `api/internal/modules/registrationprofiles/domain/completeness.go:89-314` defines missing-field logic for demographics/home, medical/dietary, health contacts, social/development, parent responsibility, emergency collection, funding support, routine care, and GDPR declaration.
- `api/internal/modules/registrationprofiles/interfaces/http/handler.go:33-37` registers the existing profile and collection-password routes.
- `api/internal/modules/registrationprofiles/interfaces/http/handler.go:124-216` parses profile patch sections and does not parse office-use checklist data.
- `api/internal/modules/registrationprofiles/application/update_profile.go` creates or updates profile sections, replaces submitted contact arrays by contact type, computes completeness, and writes redacted audit details.
- `api/internal/modules/registrationprofiles/application/patch_merge.go` currently accepts `gdpr_declared_at` from the request; the product decision requires backend-managed confirmation time instead.
- `find api/internal/modules/registrationprofiles -name '*_test.go'` returned no registration-profile module tests, so this feature must add focused backend tests.
- `web/src/app/app.routes.ts:56-62` defines the existing manager child detail route at `/staff/manager/children/:childId`; there is no registration editor route yet.
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts:38-260` loads child detail, linked guardians, and funding allowance. It does not load registration profile or office-use checklist state.
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.html` renders child details, MVP enrollment status, linked guardians, and monthly funded-hours allowance. It has no registration/enrolment summary or edit action.
- `web/src/app/features/staff/data/staff-api.service.ts` maps child, guardian, attendance, invite, funding, and linked-guardian API calls. It has no registration profile, collection password, or office-use checklist client methods.

## Decisions

Product/domain decisions from the interview:

- `FE-PM-01` includes office-use checklist editing in the same manager registration/enrolment workflow.
- Registration profile completion and office-use checklist completion must be separate statuses in the UI and API.
- Office-use checklist completion requires manager review of deposit, application date, start-date check, sessions/days requested, term-time-only space, contract/signature, handbook, Red Book, birth certificate/passport, and proof of address.
- Office-use checklist items may be complete through explicit not-applicable status. Date left is optional for current children and does not make the checklist incomplete.
- Registration profile and office-use checklist completion are manager follow-up statuses only and must not block attendance, funding, invoicing, or parent invoice access.
- Registration parent/carer entries must not automatically create or update Guardian records or guardian-child links.
- Collection password must be set/replace-only. Normal reads show only presence and last-updated metadata.
- Managers access the full editor through a dedicated child-scoped page reached from child detail. Child detail keeps a compact summary and action link.
- Sections save independently and can be complete with explicit no/none/not-applicable answers where nothing applies. Blank means unknown/incomplete.
- Consent decisions remain out of `FE-PM-01`; only GDPR declaration metadata is included.
- Office-use `Start Date` and `Date left` do not update the child record's operational enrollment dates.
- Managers may add, edit, reorder, and remove registration contact entries without changing Guardian records or portal access.
- Branch managers may view/edit the full registration profile and office-use checklist for children in their branch. No narrower permission is introduced.
- No document upload/storage is included.
- No final submit/lock action is included.
- No manager-facing audit/history view is included; show simple last-updated metadata where available.
- GDPR declaration input collects declaring person's name and declaration date from the form. The backend records the manager confirmation timestamp at save time.

Implementation decisions made by the agent:

- Preserve the conceptual boundary by implementing the office-use checklist as a separate child-scoped resource with separate `GET`/`PATCH` routes, while rendering it inside the same Angular editor page.
- Use one current checklist row per child, tenant, and branch, matching the current-profile pattern.
- Add a dedicated `registration_office_check_status` enum for required checklist items: `unknown`, `complete`, `missing`, `not_applicable`.
- Add a dedicated `registration_term_time_only_status` enum for term-time-only space: `unknown`, `yes`, `no`, `not_applicable`.
- Treat checklist completion as derived, not stored. A required checklist item is complete when its status is `complete` or `not_applicable`; term-time-only is complete when status is `yes`, `no`, or `not_applicable`.
- Keep all office-use date fields as metadata only. Required item dates are validated when present but do not by themselves determine completion except `application_date`, which requires a date when its status is `complete`.
- Implement backend-managed `gdpr_declared_at` by setting it server-side when the GDPR declaration section is saved with a declaring name and declaration date. The UI must not expose a manual timestamp input.
- Keep registration profile sections on the existing profile API and office-use checklist on new checklist API methods. The Angular page may save each section through the appropriate endpoint.
- Reuse existing staff page patterns: standalone Angular components, `StaffApiService`, `ApiErrorMapper`, page header, alert/loading/empty states, existing button/badge components, and manager role route guard.

Existing architecture/contracts to honor:

- Manager-only authorization and tenant/branch scope rules from existing registration profile routes and `api/internal/app/bootstrap/authorization_matrix_test.go`.
- Sensitive profile values must not be copied into audit details, per `CONTEXT.md:71-73`.
- Collection-password hash material must never be returned in API responses.
- Existing child/guardian/guardian-child-link/funding methods and route behavior must remain backward compatible.
- Existing MVP `enrollment_complete` and `missing_requirements` fields continue to describe the attendance/invoicing gate only.

## Acceptance Criteria

- A manager can open `/staff/manager/children/:childId`, see compact registration profile and office-use checklist statuses, and navigate to `/staff/manager/children/:childId/registration`.
- A manager can open the registration editor for a child in their branch and see the child name/date of birth from the child record as read-only context.
- The registration editor displays and saves these profile sections independently: demographics/home, medical/dietary, health contacts, social/development, parent/carer responsibility, emergency contacts/authorised collectors/collection, funding-support notes, routine care, and GDPR declaration metadata.
- Saving one profile section preserves unsent sections and returns updated completion state.
- A manager can add, edit, reorder, and remove parent/carer, emergency contact, and authorised collector entries in the registration profile.
- Removing or changing registration contact entries does not deactivate Guardian records, end guardian-child links, revoke parent access, or change existing portal relationships.
- A manager can set or replace a collection password. The UI and API never display the stored password value after save.
- GDPR declaration save collects declaring person name and declaration date from the form and records `gdpr_declared_at` from the backend save time.
- Consent decisions from the form are not shown or saved in this editor.
- A manager can edit office-use checklist metadata for deposit, application date, start-date check, date left, sessions/days requested, term-time-only space, contract/signature, handbook, Red Book, birth certificate/passport, and proof of address.
- Office-use checklist status is separate from registration profile status.
- Office-use date-left is optional for current children and does not make checklist completion fail.
- Office-use start-date and date-left values do not update `children.start_date`, `children.end_date`, attendance windows, funding windows, or billing logic.
- Incomplete profile/checklist sections show actionable missing-field labels to managers.
- Unknown/blank profile answers remain incomplete where required. Explicit no/none/not-applicable answers can complete applicable sections.
- Practitioners, parents, and owners cannot access the registration profile, collection password, or office-use checklist endpoints.
- Existing practitioner attendance APIs, parent invoice APIs, invoice preflight, and funding APIs do not include sensitive registration profile or office-use checklist fields.
- Incomplete registration profile or office-use checklist state does not block attendance check-in, absence, funding allowance save, invoice preflight, invoice generation, or parent invoice visibility.
- Existing child detail child basics, linked guardians, edit child basics, and funded-hours behavior continue to work.

## Implementation Tasks

### Task 1: Define Office-Use Checklist API Contract

- Objective: Add a stable OpenAPI contract for the separate office-use checklist resource and tighten GDPR declaration request semantics.
- Depends on: Interview decisions in this plan.
- Target files/symbols:
  - `docs/API-CONTRACT.openapi.yaml`
  - Existing registration profile schemas near `RegistrationProfileResponse`
  - Existing route section near `/api/v1/children/{child_id}/registration-profile`
- Required changes:
  - Add `RegistrationOfficeCheckStatus` enum: `unknown`, `complete`, `missing`, `not_applicable`.
  - Add `RegistrationTermTimeOnlyStatus` enum: `unknown`, `yes`, `no`, `not_applicable`.
  - Add `RegistrationOfficeUseChecklist` schema with:
    - `deposit_status`, `deposit_paid_date`
    - `application_date_status`, `application_date`
    - `start_date_status`
    - `date_left`
    - `sessions_days_requested_status`, `sessions_days_requested`
    - `term_time_only_space_status`
    - `contract_status`, `contract_date`
    - `handbook_status`, `handbook_date`
    - `red_book_status`, `red_book_checked_date`
    - `birth_certificate_passport_status`, `birth_certificate_passport_checked_date`
    - `proof_of_address_status`, `proof_of_address_checked_date`
    - `notes`
  - Add `RegistrationOfficeUseCompleteness` schema with `is_complete`, `missing_fields`, and `items` where each item has `code`, `status`, and optional `missing_fields`.
  - Add `RegistrationOfficeUseChecklistResponse` schema with child summary, `checklist_exists`, optional checklist metadata, `office_use_checklist`, and `completeness`.
  - Add `RegistrationOfficeUseChecklistPatchRequest` with the same editable fields as the checklist object.
  - Add manager-only routes:
    - `GET /api/v1/children/{child_id}/registration-office-use-checklist`
    - `PATCH /api/v1/children/{child_id}/registration-office-use-checklist`
  - Update GDPR declaration request documentation so managers send `gdpr_declared_by_name` and `gdpr_declaration_date`; `gdpr_declared_at` is response-only/server-managed.
- Tests/verification:
  - Run `npx @redocly/cli@latest lint --extends=minimal docs/API-CONTRACT.openapi.yaml`.
- Expected outcome: API consumers have an explicit manager-only checklist contract that keeps office-use completion separate from registration profile completion.

### Task 2: Add Office-Use Checklist Persistence

- Objective: Persist one current office-use checklist per child with tenant/branch scope and reversible migrations.
- Depends on: Task 1.
- Target files/symbols:
  - `api/db/migrations/000017_add_child_registration_office_checklists.up.sql`
  - `api/db/migrations/000017_add_child_registration_office_checklists.down.sql`
  - `api/db/query/registration_office_checklists.sql`
  - Generated `api/internal/platform/db/sqlc/*`
  - `api/sqlc.yaml`
- Required changes:
  - Create enum type `registration_office_check_status`.
  - Create enum type `registration_term_time_only_status`.
  - Create table `child_registration_office_checklists` with:
    - `id UUID PRIMARY KEY`
    - `tenant_id UUID NOT NULL`
    - `branch_id UUID NOT NULL`
    - `child_id UUID NOT NULL`
    - the fields listed in Task 1
    - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
    - `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
    - foreign key `(tenant_id, branch_id)` to `branches`
    - foreign key `(tenant_id, branch_id, child_id)` to `children`
    - unique `(tenant_id, branch_id, child_id)`
  - Add `CHECK` constraints that status/date fields are structurally valid where needed:
    - `application_date` must be present when `application_date_status = 'complete'`.
    - `sessions_days_requested` must be nonblank when `sessions_days_requested_status = 'complete'`.
    - Date fields may be null when status is `missing`, `unknown`, or `not_applicable`.
  - Add sqlc queries:
    - get child summary including child start/end dates for context
    - get checklist by child
    - get checklist by child for update
    - create checklist
    - update checklist
  - Run `make sqlc-generate` and commit generated sqlc files.
- Tests/verification:
  - Run `make sqlc-generate`.
  - Run `VERIFY_DATABASE_URL=<disposable-postgres-url> make migrate-verify` when a disposable database is available.
  - Run `cd api && go test ./internal/platform/db/...`.
- Expected outcome: Checklist storage exists, is tenant/branch scoped, and is represented in generated sqlc code.

### Task 3: Implement Office-Use Domain, Repository, and Completeness

- Objective: Add backend domain logic for office-use checklist defaults, patch merge, validation, completeness, persistence, and audit.
- Depends on: Task 2.
- Target files/symbols:
  - `api/internal/modules/registrationprofiles/domain/office_checklist.go`
  - `api/internal/modules/registrationprofiles/domain/repository.go`
  - `api/internal/modules/registrationprofiles/domain/office_completeness.go`
  - `api/internal/modules/registrationprofiles/infrastructure/postgres/repository.go`
  - `api/internal/modules/registrationprofiles/application/get_office_checklist.go`
  - `api/internal/modules/registrationprofiles/application/update_office_checklist.go`
  - `api/internal/modules/registrationprofiles/application/office_patch_merge.go`
- Required changes:
  - Add domain types for office check status, term-time status, `OfficeUseChecklist`, `OfficeUseChecklistWithChild`, `OfficeUseCompleteness`, and item-level completeness.
  - Add default checklist values:
    - all required item statuses default to `unknown`
    - nullable dates/text default to nil
    - checklist does not exist until first meaningful save
  - Add repository methods for child summary, get by child, get for update, create, and update.
  - Add patch merge logic that trims text fields, validates enum values, validates dates as `YYYY-MM-DD`, and returns changed field codes.
  - Add completeness logic:
    - required fields: `deposit`, `application_date`, `start_date_check`, `sessions_days_requested`, `term_time_only_space`, `contract`, `handbook`, `red_book`, `birth_certificate_passport`, `proof_of_address`
    - `date_left` never contributes to missing fields
    - required check-status fields are complete when `complete` or `not_applicable`
    - term-time-only is complete when `yes`, `no`, or `not_applicable`
    - `application_date_status = complete` requires `application_date`
    - `sessions_days_requested_status = complete` requires nonblank `sessions_days_requested`
  - Add update use case with transaction, tenant/branch child lookup, create-on-first-change, update-on-existing, recomputed completeness, and audit action:
    - `registration_office_checklist_created`
    - `registration_office_checklist_updated`
  - Audit details must include child ID, checklist ID, changed field codes, checklist created flag, missing fields before, and missing fields after. Do not include free-text notes or dates in audit details.
- Tests/verification:
  - Add table-driven unit tests for `ComputeOfficeUseCompleteness`.
  - Add application tests for create, update, invalid enum/date, no-op patch, and audit details redaction.
  - Run `cd api && go test ./internal/modules/registrationprofiles/...`.
- Expected outcome: Office-use checklist behavior is deterministic, auditable, and independent from the registration profile.

### Task 4: Expose Office-Use Checklist HTTP Routes

- Objective: Add manager-only HTTP handlers for office-use checklist get/update and register them with existing staff route wiring.
- Depends on: Task 3.
- Target files/symbols:
  - `api/internal/modules/registrationprofiles/interfaces/http/handler.go`
  - `api/internal/modules/registrationprofiles/interfaces/http/office_dto.go`
  - `api/internal/app/bootstrap/adapters.go`
  - `api/internal/app/bootstrap/routes.go` or the current route registration file found by existing registration profile registration
  - `api/internal/app/bootstrap/authorization_matrix_test.go`
- Required changes:
  - Extend `Handler` constructor with `GetOfficeUseChecklist` and `UpdateOfficeUseChecklist` use cases.
  - Register:
    - `manager.GET("/children/:child_id/registration-office-use-checklist", h.getOfficeUseChecklistHandler)`
    - `manager.PATCH("/children/:child_id/registration-office-use-checklist", h.updateOfficeUseChecklistHandler)`
  - Add DTO mapping from domain to the response contract in Task 1.
  - Parse raw JSON patch maps so omitted fields preserve current values and explicit `null` clears nullable date/text fields.
  - Use existing domain error mapping for validation, not found, unauthorized, and forbidden responses.
  - Add authorization matrix rows proving unauthenticated users are rejected and practitioner, parent, and owner roles are forbidden.
  - Add tenant/branch scope test proving a manager in scope A receives `child_not_found` for a scope B child.
- Tests/verification:
  - Run `cd api && go test ./internal/app/bootstrap -run 'TestAuthorizationMatrix'`.
  - Run `cd api && go test ./...`.
- Expected outcome: The office-use checklist routes follow existing manager-only authorization and scope behavior.

### Task 5: Tighten GDPR Declaration Confirmation Semantics

- Objective: Make GDPR declaration confirmation time backend-managed.
- Depends on: Existing registration profile update use case.
- Target files/symbols:
  - `api/internal/modules/registrationprofiles/application/patch_merge.go`
  - `api/internal/modules/registrationprofiles/application/update_profile.go`
  - `api/internal/modules/registrationprofiles/interfaces/http/dto.go`
  - `docs/API-CONTRACT.openapi.yaml`
- Required changes:
  - Stop accepting manager-entered `gdpr_declared_at` as an editable UI/request field.
  - In the update use case, when the GDPR declaration section is patched with a nonblank `gdpr_declared_by_name` and valid `gdpr_declaration_date`, set `GDPRDeclaredAt` to the backend current UTC time during the successful transaction.
  - If either declaring name or declaration date is cleared, clear `GDPRDeclaredAt` so completeness correctly reports the declaration incomplete.
  - Keep `gdpr_declared_at` in responses.
  - Preserve existing error semantics for invalid declaration dates.
  - Add tests proving a client-supplied `gdpr_declared_at` is ignored or rejected according to the final OpenAPI request contract, and that the saved response timestamp is generated by the backend.
- Tests/verification:
  - Run `cd api && go test ./internal/modules/registrationprofiles/...`.
  - Run `cd api && go test ./internal/app/bootstrap -run Registration`.
- Expected outcome: GDPR declaration completion reflects a real manager save/confirmation and no manual time entry is possible.

### Task 6: Add Frontend Registration Models and API Methods

- Objective: Give Angular typed access to registration profile, collection password, and office-use checklist contracts.
- Depends on: Tasks 1 and 4.
- Target files/symbols:
  - `web/src/app/features/staff/models/registration-profile.models.ts`
  - `web/src/app/features/staff/data/staff-api.service.ts`
  - `web/src/app/features/staff/data/staff-api.service.spec.ts`
  - `web/src/app/features/staff/utils/registration-profile-formatters.ts`
  - `web/src/app/features/staff/utils/registration-profile-formatters.spec.ts`
- Required changes:
  - Add TypeScript types for:
    - registration profile response and section patch payloads
    - contact entries
    - completion sections and missing-field codes
    - collection-password request
    - office-use checklist response, patch payload, check-status enum, term-time enum, and completeness
  - Add `StaffApiService` methods:
    - `getRegistrationProfile(childId)`
    - `patchRegistrationProfile(childId, patch)`
    - `setRegistrationCollectionPassword(childId, password)`
    - `getRegistrationOfficeUseChecklist(childId)`
    - `patchRegistrationOfficeUseChecklist(childId, patch)`
  - Map snake_case API responses to camelCase frontend records and back to snake_case request payloads.
  - Ensure `gdprDeclaredAt` is read-only in frontend records and not sent in patch payloads.
  - Add formatter functions for registration profile section labels, profile missing-field labels, office-use missing-field labels, check-status labels, and completion badge status.
  - Preserve existing child/funding/guardian API mapping tests.
- Tests/verification:
  - Run `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/data/staff-api.service.spec.ts'`.
  - Run `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/utils/registration-profile-formatters.spec.ts'`.
- Expected outcome: Angular can load and save the new contracts without untyped ad hoc mapping.

### Task 7: Build Dedicated Manager Registration Editor Page

- Objective: Implement the full sectioned manager editor at a child-scoped route.
- Depends on: Task 6.
- Target files/symbols:
  - `web/src/app/app.routes.ts`
  - `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.ts`
  - `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.html`
  - `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.spec.ts`
  - Optional focused child components under `web/src/app/features/staff/components/registration-profile/`
- Required changes:
  - Add manager-only route:
    - `staff/manager/children/:childId/registration`
    - title `Child Registration | Nursery Management`
  - Load registration profile and office-use checklist for the child on init.
  - Render child name and date of birth from the API as read-only context.
  - Render separate status summary for:
    - Registration profile completion
    - Office-use checklist completion
  - Build independent save controls and save/error/success state for these sections:
    - Demographics and home
    - Medical and dietary
    - Doctor and health visitor
    - Social services, development concerns, and professional referrals
    - Parent/carer details and parental responsibility
    - Emergency contacts, authorised collectors, over-18 acknowledgement, and collection password
    - Benefits/funding-support notes
    - Routine/free-text care notes
    - GDPR declaration metadata
    - Office-use checklist
  - Support add/edit/reorder/remove for registration contacts. Reordering must update array order before saving because backend stores sort order from submitted order.
  - In the collection-password section:
    - show `isSet`, last-updated timestamp, and replacement field
    - never show the stored value
    - clear the password input after successful save
  - In the GDPR section:
    - show declaring person name input and declaration date input
    - show backend confirmation timestamp after save when present
    - do not provide a timestamp input
  - In office-use:
    - show existing child start/end dates as context
    - save checklist metadata only
    - label `missing` statuses as still needed rather than complete
  - Use existing UI primitives (`app-page-header`, `app-alert`, `app-loading-state`, `app-status-badge`, `app-button`) and existing Tailwind utility style patterns.
  - Keep the page dense and work-focused; avoid a landing/hero layout.
- Tests/verification:
  - Component tests cover loading, status rendering, section save, section error mapping, contact add/remove/reorder, collection password set/clear, GDPR no timestamp input, office-use save, and manager route configuration.
  - Run `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.spec.ts'`.
- Expected outcome: Managers have a usable full editor without crowding the existing child detail page.

### Task 8: Add Child Detail Registration Summary and Action

- Objective: Keep child detail compact while surfacing registration follow-up status and navigation.
- Depends on: Tasks 6 and 7.
- Target files/symbols:
  - `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts`
  - `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.html`
  - `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.spec.ts`
- Required changes:
  - Load registration profile and office-use checklist completion summaries after the child is loaded.
  - Add a compact `Registration/enrolment follow-up` summary to child detail with:
    - registration profile status badge
    - office-use checklist status badge
    - missing section/item labels when incomplete
    - a `Registration` action link to `/staff/manager/children/:childId/registration`
  - If registration summary load fails, show a non-blocking alert in the registration summary area without breaking child basics, linked guardians, or funding allowance.
  - Preserve existing edit child basics, linked guardian, and funding allowance behavior.
- Tests/verification:
  - Update existing child detail tests to cover registration summary loading, incomplete labels, navigation link, non-blocking load error, and no regression for funding/guardian behaviors.
  - Run `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.spec.ts'`.
- Expected outcome: Managers can see registration follow-up from child detail and enter the full editor deliberately.

### Task 9: Protect Existing Operational Surfaces

- Objective: Prove registration/checklist data does not leak to out-of-scope roles or change operational gates.
- Depends on: Tasks 3, 4, 7, and 8.
- Target files/symbols:
  - `api/internal/modules/children/*`
  - `api/internal/modules/attendance/*`
  - `api/internal/modules/funding/*`
  - `api/internal/modules/billing/*`
  - Existing route tests under `api/internal/app/bootstrap/`
  - Existing Angular tests for practitioner attendance, parent invoices, funding, invoice run, child detail
- Required changes:
  - Add or update backend tests proving:
    - registration/checklist completion does not alter `children.enrollment_complete`
    - attendance check-in continues to use MVP enrollment minimum only
    - invoice preflight blockers remain based on MVP enrollment, funding profile, billing rate, and attendance completeness only
    - practitioner attendance and parent invoice responses do not include registration or office-use fields
  - Add or update frontend tests proving:
    - practitioner attendance screens do not render profile/checklist details
    - parent invoice screens do not render profile/checklist details
    - invoice run/funding screens still route to child detail as before
- Tests/verification:
  - Run `cd api && go test ./internal/modules/attendance/... ./internal/modules/billing/... ./internal/modules/funding/... ./internal/app/bootstrap`.
  - Run relevant Angular specs for practitioner attendance, parent invoice detail/list, funding overview, invoice run, and child detail.
- Expected outcome: `FE-PM-01` is additive manager follow-up functionality and does not change existing operations.

### Task 10: Update Manual Test Cases and Backlog Notes

- Objective: Document the implemented workflow and reconcile the office-use scope movement.
- Depends on: Tasks 1-9.
- Target files/symbols:
  - `docs/FRONTEND-MANUAL-TEST-CASES.md`
  - `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`
  - `docs/MVP-30D-API-BACKEND-BACKLOG.md`
  - `docs/API-SCHEMA-STATE.md`
- Required changes:
  - Add manual test cases for:
    - opening editor from child detail
    - partial profile section save
    - contact add/edit/reorder/remove
    - collection password set/replace without display
    - GDPR declaration confirmation timestamp
    - office-use checklist completion with not-applicable item
    - date-left optional for current child
    - non-blocking status relative to attendance/funding/invoicing
    - no document upload controls
  - Mark or annotate `FE-PM-01` as done after implementation details and test counts are known.
  - Annotate `FE-PM-04` as folded into `FE-PM-01` for checklist metadata UI if the implementation includes the complete checklist UI from this plan. Leave any future document-upload/evidence work out of `FE-PM-04`.
  - Update API backlog notes if `API-PM-03` checklist metadata was implemented by this work.
  - Update `docs/API-SCHEMA-STATE.md` with the new office-use checklist table and enums.
- Tests/verification:
  - Inspect docs for consistency with `CONTEXT.md`.
- Expected outcome: Product docs no longer imply office-use checklist UI is still pending separately after this implementation.

## Contracts

API and data contracts:

- Existing route `GET /api/v1/children/{child_id}/registration-profile` remains manager-only and returns the current registration profile or defaults when no profile exists.
- Existing route `PATCH /api/v1/children/{child_id}/registration-profile` remains manager-only and accepts partial section patches. Submitted contact arrays replace only their submitted contact type.
- Existing route `PUT /api/v1/children/{child_id}/registration-profile/collection-password` remains manager-only and returns profile metadata without password material.
- New route `GET /api/v1/children/{child_id}/registration-office-use-checklist` is manager-only and returns current checklist metadata or defaults when no checklist exists.
- New route `PATCH /api/v1/children/{child_id}/registration-office-use-checklist` is manager-only and creates or updates the current checklist.
- New office-use response contract:
  - `child`: id, full_name, date_of_birth, start_date, end_date
  - `checklist_exists`: boolean
  - `checklist`: id, created_at, updated_at when persisted
  - `office_use_checklist`: editable metadata fields
  - `completeness`: derived checklist completion
- Office-use statuses:
  - check status enum: `unknown`, `complete`, `missing`, `not_applicable`
  - term-time-only enum: `unknown`, `yes`, `no`, `not_applicable`
- Office-use missing field codes:
  - `deposit`
  - `application_date`
  - `start_date_check`
  - `sessions_days_requested`
  - `term_time_only_space`
  - `contract`
  - `handbook`
  - `red_book`
  - `birth_certificate_passport`
  - `proof_of_address`
- GDPR declaration:
  - Request: `gdpr_declared_by_name`, `gdpr_declaration_date`
  - Response: `gdpr_declared_by_name`, `gdpr_declaration_date`, backend-managed `gdpr_declared_at`
- Permission contract:
  - Manager: allowed for children in active tenant/branch session scope
  - Practitioner, parent, owner: forbidden
  - Cross-branch child IDs: not found
- Error contract:
  - Invalid UUID or invalid patch payload: `400 validation_error`
  - Invalid enum/date/status: `400 validation_error` with field name
  - Child outside scope or missing: `404 child_not_found`
  - Unauthorized session: `401 unauthorized`
  - Wrong role: `403 forbidden_role`
- Audit contract:
  - Registration profile updates continue to audit changed section codes and completeness before/after without sensitive values.
  - Office-use checklist updates audit changed field codes and completeness before/after without free-text notes, date values, or document metadata values.

UI contracts:

- Child detail remains the compact child basics/enrollment/funding/guardian page.
- Registration editor route is `/staff/manager/children/:childId/registration`.
- Registration editor shows two top-level statuses: registration profile and office-use checklist.
- Each section has its own save state. A failed section save must not discard unsaved fields in other sections.
- Collection password is write-only after entry and clears from the input after successful save.
- Contact entries edited in registration profile are visibly distinct from linked guardians.
- Office-use start/end dates are shown as administrative context, not operational date edit controls.

## Files to Change

Backend/API:

- `docs/API-CONTRACT.openapi.yaml`
- `docs/API-SCHEMA-STATE.md`
- `api/db/migrations/000017_add_child_registration_office_checklists.up.sql`
- `api/db/migrations/000017_add_child_registration_office_checklists.down.sql`
- `api/db/query/registration_office_checklists.sql`
- Generated files under `api/internal/platform/db/sqlc/`
- `api/internal/modules/registrationprofiles/domain/profile.go`
- `api/internal/modules/registrationprofiles/domain/repository.go`
- `api/internal/modules/registrationprofiles/domain/completeness.go`
- `api/internal/modules/registrationprofiles/domain/office_checklist.go`
- `api/internal/modules/registrationprofiles/domain/office_completeness.go`
- `api/internal/modules/registrationprofiles/application/patch_merge.go`
- `api/internal/modules/registrationprofiles/application/update_profile.go`
- `api/internal/modules/registrationprofiles/application/get_office_checklist.go`
- `api/internal/modules/registrationprofiles/application/update_office_checklist.go`
- `api/internal/modules/registrationprofiles/application/office_patch_merge.go`
- `api/internal/modules/registrationprofiles/infrastructure/postgres/repository.go`
- `api/internal/modules/registrationprofiles/interfaces/http/dto.go`
- `api/internal/modules/registrationprofiles/interfaces/http/office_dto.go`
- `api/internal/modules/registrationprofiles/interfaces/http/handler.go`
- Bootstrap wiring files that currently instantiate/register registration profile routes; locate with `rg -n "NewHandler\\(|registrationprofiles|RegisterRoutes" api/internal/app api/internal/modules`
- `api/internal/app/bootstrap/authorization_matrix_test.go`
- New backend tests under `api/internal/modules/registrationprofiles/domain/`, `application/`, `interfaces/http/`, and existing bootstrap tests.

Frontend:

- `web/src/app/app.routes.ts`
- `web/src/app/features/staff/data/staff-api.service.ts`
- `web/src/app/features/staff/data/staff-api.service.spec.ts`
- `web/src/app/features/staff/models/registration-profile.models.ts`
- `web/src/app/features/staff/utils/registration-profile-formatters.ts`
- `web/src/app/features/staff/utils/registration-profile-formatters.spec.ts`
- `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.ts`
- `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.html`
- `web/src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.spec.ts`
- Optional focused form components under `web/src/app/features/staff/components/registration-profile/`
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts`
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.html`
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.spec.ts`

Docs:

- `docs/FRONTEND-MANUAL-TEST-CASES.md`
- `docs/MVP-30D-ANGULAR-FRONTEND-BACKLOG.md`
- `docs/MVP-30D-API-BACKEND-BACKLOG.md`

## Verification

Run from the repository root unless noted.

- Generate sqlc code:
  - `make sqlc-generate`
- Validate migrations with a disposable database when available:
  - `VERIFY_DATABASE_URL=<disposable-postgres-url> make migrate-verify`
- Run backend tests:
  - `cd api && go test ./internal/modules/registrationprofiles/...`
  - `cd api && go test ./internal/app/bootstrap -run 'TestAuthorizationMatrix|Registration'`
  - `cd api && go test ./...`
- Lint OpenAPI:
  - `npx @redocly/cli@latest lint --extends=minimal docs/API-CONTRACT.openapi.yaml`
- Run focused frontend tests:
  - `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/data/staff-api.service.spec.ts'`
  - `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/pages/manager-child-registration/manager-child-registration.component.spec.ts'`
  - `cd web && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.spec.ts'`
- Run full frontend build/tests:
  - `cd web && npm test -- --watch=false --browsers=ChromeHeadless`
  - `cd web && npm run build`

Manual validation scenarios:

- Manager opens a child detail page, sees registration profile and office-use checklist statuses, and navigates to the registration editor.
- Manager saves demographics/home only, reloads the page, and confirms other profile sections are unchanged.
- Manager saves medical/dietary with explicit no/none answers and reviewed state, then sees the section become complete.
- Manager adds two emergency contacts, reorders them, removes one, reloads, and confirms ordering/removal.
- Manager sets a collection password, reloads, and confirms only presence/last-updated metadata is visible.
- Manager enters GDPR declaring name and declaration date, saves, and sees backend confirmation timestamp.
- Manager marks Red Book not applicable and proof of address missing, then sees checklist incomplete only for proof of address and any other missing items.
- Manager records date left in office-use checklist and confirms the child detail operational end date is unchanged.
- Practitioner and parent sessions cannot access registration editor routes or API endpoints.
- Incomplete registration profile/checklist does not prevent check-in, funded-hours save, invoice preflight, or parent invoice view when MVP enrollment/funding/billing requirements are otherwise satisfied.

## Assumptions

- `FE-PM-01` may include the office-use checklist metadata UI now, and `FE-PM-04` can later be reduced or marked folded into this implementation for metadata-only checklist work.
- Office-use checklist storage belongs in the existing registration profile module because it is child-scoped, manager-only, and adjacent to current registration profile routes; the resource remains separate in API paths, schemas, and completion status.
- A single current office-use checklist per child is sufficient; historical office-use browsing is out of scope and audit records provide accountability.
- Existing registration profile section completeness remains acceptable; implementation should add not-applicable support only where the existing domain model already has a clear status path or where office-use status enums need it.
- A check-status `missing` answer means the manager reviewed the item but the required item is still outstanding, so it remains incomplete.
- Document/check dates are useful metadata but not required for completion except `application_date`, which is itself a required checklist item.
- If a local environment lacks PostgreSQL or migration tooling, migration verification can be skipped locally but must be run in CI or a disposable database before merging.

## Risks and Fallbacks

- Risk: Extending office-use checklist work into `FE-PM-01` creates backlog confusion with `FE-PM-04`.
  - Fallback: Update backlog docs in Task 10 to mark metadata-only office-use checklist UI as folded into `FE-PM-01`, leaving only future evidence/file-upload or expanded checklist work for follow-up.
- Risk: Office-use checklist route naming conflicts with future API naming preferences.
  - Fallback: Keep the resource boundary stable and use the exact path in this plan unless an existing route-registration convention strongly prefers a shorter equivalent. If changed, update OpenAPI, frontend service methods, and manual tests in the same implementation.
- Risk: Backend-managed GDPR timestamp changes existing permissive request behavior.
  - Fallback: Accept but ignore client-supplied `gdpr_declared_at` for one compatibility cycle while documenting it as response-only. Tests must prove persisted value is backend-generated.
- Risk: The editor page becomes too large and hard to test.
  - Fallback: Keep the route-level page as coordinator and extract focused standalone child components for high-complexity sections: contacts, social referrals, collection password, and office-use checklist.
- Risk: Angular tests become slow because the editor has many form controls.
  - Fallback: Put mapping and formatter logic in separately tested utilities/models, and keep component tests scenario-focused on save behavior, state rendering, and permissions-critical UI.
- Risk: Office-use checklist completeness rules are stricter than historic paper records can satisfy.
  - Fallback: Managers can use `not_applicable` where the nursery does not require an item for the child. Do not weaken `missing` to count as complete, because that would hide outstanding administrative work.
- Risk: Generated sqlc files or migrations diverge from the OpenAPI contract.
  - Fallback: Treat OpenAPI and migration/schema tests as authoritative; adjust DTO/domain mapping until API responses match the documented contract.
