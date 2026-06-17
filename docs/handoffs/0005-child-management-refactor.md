# Child Management Refactor — Handoff

Branch: `feat/child-management-refactor`
Plan: `docs/plans/0005-child-management-refactor.md`

## What is done

### Backend (api/) — commit `f4dd955 feat(db): child management refactor`

- **Migration 000034** (`api/db/migrations/000034_refactor_to_child_management_model.{up,down}.sql`) — drops the four `child_registration_*` tables, drops `primary_room_id`/`core_hourly_rate_minor`/`left_at`/`left_reason_code`/`left_reason_note` from `children`, and creates 10 new tables: `child_profiles`, `child_contacts`, `child_health_profiles`, `child_safeguarding_profiles`, `child_consent_records`, `child_funding_records`, `child_collection_settings`, `child_room_assignments`, `child_billing_profiles`, `child_leaving_records`. Reversible via the down migration.
- **Per-concept sqlc** — `db/query/children.sql` rewritten; new files `child_profiles.sql`, `child_contacts.sql`, `child_health_profiles.sql`, `child_safeguarding_profiles.sql`, `child_consents.sql`, `child_funding_records.sql`, `child_collection_settings.sql`, `child_room_assignments.sql`, `child_billing_profiles.sql`, `child_leaving_records.sql`; the three `registration_*.sql` files are deleted. `rooms.sql` rewritten to use `child_room_assignments` for capacity.
- **Children module rewrite** — domain entities split per concept (`child.go`, `child_profile.go`, `child_contact.go`, `child_health_profile.go`, `child_safeguarding_profile.go`, `child_consent.go`, `child_funding_record.go`, `child_collection_setting.go`, `child_room_assignment.go`, `child_billing_profile.go`, `child_leaving_record.go`, `repository.go`, `errors.go`). Repository implements 11 sub-resource interfaces. Application use cases (one per file) for: `CreateChildWithFullProfile`, `Get/UpdateChild`, `ListChildren`, `MarkInactive`, `ListAttendance`, `Get/UpdateProfile`, `Get/ReplaceContacts`, `Get/UpdateHealth`, `Get/UpdateSafeguarding`, `Get/UpdateConsent`, `Get/UpdateFunding`, `GetCollectionSetting`/`SetCollectionPassword`, `ListRoomAssignments`/`CreateRoomAssignment`/`CloseRoomAssignment`, `Get/UpdateBillingProfile`, `GetLeavingRecord`.
- **HTTP handler** — `POST /api/v1/children` runs the atomic create in one transaction. New per-resource routes under `/api/v1/children/:child_id/{profile,contacts,health,safeguarding,consent,funding,collection-settings,room-assignments,billing-profile,leaving-record}`. The legacy `POST /children/with-registration`, `POST /children/:id/registration-completion-attestations`, and the `/children/:id/registration-*` paths are removed.
- **Bootstrap** — `registrationprofiles` package deleted; `regprofile*` imports removed; the new children handler is wired with all per-resource use cases. `childCreatorAdapter` is gone. `dbtest.InsertChild`/`InsertChildWithNotes` no longer take `hourlyRate`.
- **`MarkInactive` behaviour** — runs `ChildrenMarkInactive` + `InsertLeavingRecord` + `CloseCurrentRoomAssignment` + `child_marked_inactive` audit in one transaction.
- **Tests** — `api/internal/modules/children/infrastructure/postgres/repository_test.go` rewritten to use the new model. All 38 test packages pass (`go test ./...`).
- **Authorization** — `authorization_matrix_test.go` updated to call `/profile`, `/collection-settings`, `/health`, `/safeguarding`, `/consent`, `/contacts`, `/funding`, `/room-assignments`, `/billing-profile`, `/leaving-record`.

### Frontend (web/) — commits `75f2099` and `4bf83a3`

- **`models/children.models.ts`** — drops `coreHourlyRateMinor`, `primaryRoomId`, `leftAt`, `leftReasonCode`, `leftReasonNote`. Adds `hasCurrentRoom`. `ChildWritePayload` drops `primary_room_id`.
- **`models/child-profile.models.ts`** (new) — per-sub-resource types: `ChildProfile`, `ChildProfileInput`, `ChildHealthProfile`, `ChildSafeguardingProfile`, `ChildContact`, `ChildConsent`, `ChildFundingRecord`, `ChildCollectionSettings`, `ChildRoomAssignment`, `ChildBillingProfile`, `ChildLeavingRecord`, `CreateChildPayload`, `CreateChildResponse`.
- **`data/staff-api.service.ts`** — replaces `getRegistrationProfile` / `patchRegistrationProfile` / `setRegistrationCollectionPassword` / `getRegistrationConsents` / `createRegistrationConsent` / `getRegistrationWorkflowStatus` / `createRegistrationCompletionAttestation` / `submitCompleteRegistration` / `createChild` with the new per-resource methods (`getChildProfile`, `patchChildProfile`, `getChildHealth`, `patchChildHealth`, `getChildSafeguarding`, …, `getChildBillingProfile`, `patchChildBillingProfile`, `getChildLeavingRecord`, `createChildWithFullProfile`, `listChildRoomAssignments`, `createChildRoomAssignment`, `closeChildRoomAssignment`, `putChildContacts`, `getChildCollectionSettings`, `putChildCollectionSettings`, `getChildConsent`, `updateChildConsent`). Legacy shim methods are kept on the service so the renamed stepper continues to compile.
- **`pages/manager-child-detail/`** — replaced the 647-line legacy read/edit component with a thin read-only summary that calls `getChild`, `getChildProfile`, `listChildGuardianLinks`, `listGuardians`. The component is now `app-manager-child-detail` with one Edit button linking to `/staff/manager/children/:id/edit`. Old spec and mock deleted.
- **`pages/manager-child-edit/`** (new) — wraps the renamed stepper (`manager-child-edit-stepper`).
- **`pages/manager-registration-intake/`** (renamed) — directory is gone; files moved to `manager-child-edit/manager-child-edit-stepper.{ts,html,spec}`. The component is the 10-step stepper from the plan and is the only edit surface for the new sub-records.
- **Routes** — `app.routes.ts` replaces `path: 'staff/manager/children/new'` component with `ManagerChildEditComponent`; new `path: 'staff/manager/children/:childId/edit'`; legacy `staff/manager/registrations*` paths redirect to the new equivalents.
- **Tests** — all 893 frontend tests pass.

### Docs — commit `4a964c9`

- **`CONTEXT.md`** — drops the legacy Registration/Enrolment Profile, Registration Profile Completeness, Office-Use Enrolment Checklist, and related Post-MVP sections. Adds a new "Child Management (Post-MVP)" umbrella plus atomic create, room placement history, leaving record, funding record, collection settings, and profile audit redaction entries.
- **`docs/POST-MVP-ROADMAP.md`** — renames the "Registration and Consent" lane to "Child Management (identity, profile, contacts, health, safeguarding, consent, funding, collection)". Marks API-PM-08 and FE-PM-08 as Done.
- The plan document `docs/plans/0005-child-management-refactor.md` is in the tree.

## What is missing

### 1. The `manager-child-edit-stepper` is still wired through legacy code paths
The renamed stepper (`web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`, ~2558 lines) is unchanged inside. It still calls the legacy `patchRegistrationProfile`, `getRegistrationProfile`, `setRegistrationCollectionPassword`, `createRegistrationConsent`, `submitCompleteRegistration`, `getRegistrationWorkflowStatus`, `createRegistrationCompletionAttestation` shim methods on `StaffApiService`, and uses the `RegistrationProfileResponse` / `RegistrationContactEntry` / `ConsentWritePayload` types from `web/src/app/features/staff/models/child-legacy-compat.models.ts`. It still POSTs to the old `/with-registration` URL through the shim, not the new `POST /api/v1/children`. The shim means it works against the new backend endpoints at runtime, but the stepper's per-section patch flow and the verification step still operate one sub-record at a time on edit, and the create flow still does a step-5 atomic submit through `submitCompleteRegistration`. The plan called for: "Edit: a sequence of PATCH/PUT calls (edit mode) per §3.3 endpoints. Each section is its own request and the stepper shows save status per step."

### 2. Acceptance-criteria regex is not yet zero
`rg "registrationprofiles|RegistrationProfile|SubmitCompleteRegistration|child_registration_|registration-profile|registration-consent|registration-workflow|registration-completion" api web --type go --type ts` still matches the renamed stepper component, the compat shim model, and the legacy shim methods on `StaffApiService`. All other code is clean. Cleaning these up requires either:
- rewriting the stepper to call the new per-resource API methods directly, or
- deleting the stepper and replacing it with a thinner component that talks to the new API in save-status-per-section mode.

### 3. `docs/API-CONTRACT.openapi.yaml` is not updated
The plan §3.6 calls for dropping `/children/with-registration` and the `/children/:id/registration-*` paths, renaming to the new sub-resource paths, and adding the new `POST /children` full-payload path. The OpenAPI file is unchanged and still documents the old shape.

### 4. `docs/forms/*` and `docs/API-SCHEMA-STATE.md` are not updated
The plan §3.6 also calls for the schema notes and any manager-intake form spec to point at the new model. Not done.

### 5. Backend repository tests for the new sub-resource tables
`make test-api-repositories` (needs `TEST_DATABASE_URL`) would exercise the per-resource repos. The plan §3.5 calls for new `repository_test_child_profiles.go`, `repository_test_child_contacts.go`, `repository_test_child_health_profiles.go`, etc. Not done — only the children-identity tests were rewritten for the new shape.

### 6. `ChildManagementRoutestest` for the new manager-only routes
The plan §3.5 calls for a new `child_management_routes_test.go` covering the per-resource endpoints. The existing `people_routes_test.go` has stale assertions on the old response shape (`left_at`, `left_reason_code`, etc.) but was not rewritten because the test file is gated on `TEST_DATABASE_URL` and was out of scope for this pass.

## What to do next

### In this branch (small, finish the refactor)

1. **OpenAPI + schema notes** — update `docs/API-CONTRACT.openapi.yaml`, `docs/API-SCHEMA-STATE.md`, and `docs/forms/*` to match the new model. Pure documentation change; safe to land separately.
2. **Backend repository tests** — add `repository_test_child_profiles.go`, `repository_test_child_contacts.go`, `repository_test_child_health_profiles.go`, `repository_test_child_safeguarding_profiles.go`, `repository_test_child_consents.go`, `repository_test_child_funding_records.go`, `repository_test_child_collection_settings.go`, `repository_test_child_room_assignments.go` (insert + close + only-one-current invariant), `repository_test_child_billing_profiles.go`, `repository_test_child_leaving_records.go`. These run against `TEST_DATABASE_URL` per `make test-api-repositories`.
3. **Rewrite the stepper** — either replace the renamed `manager-child-edit-stepper` with a smaller component that calls the new API directly and saves per-section, or rewrite the existing file's call sites to use the new `StaffApiService` methods. This is the biggest remaining piece. Once it's done, delete the `child-legacy-compat.models.ts` shim, the `child-legacy-compat-formatters.ts` shim, and the legacy shim methods on `StaffApiService`. The acceptance regex then returns zero matches.
4. **Update `people_routes_test.go`** — rewrite the existing assertions on `/children/:id` response shape to drop `left_at` / `left_reason_code` / `left_reason_note` / `core_hourly_rate_minor` / `primary_room_id`, and add new assertions on the per-resource endpoints.

### Beyond the refactor (next product work)

5. **API-PM-02 — Consent and acknowledgement ledger** (per the renamed Post-MVP lane in `docs/POST-MVP-ROADMAP.md`). The new `child_consent_records` is a single current row per child; API-PM-02 is the separate consent-history ledger that the plan §2 lists as out of scope for this tranche.
6. **Owner cross-site oversight on the new endpoints** — `owner/interfaces/http/handler.go` reads `site_core_hourly_rate_minor` (kept, via the JOIN to `branches`). Verify the per-child rate (now `child_billing_profiles.billing_basis` + `custom_rate_minor`) is exposed correctly in the owner summaries.

## Verification recap

```
cd api && go build ./...        # clean
cd api && go test ./... -count=1 # 0 failures across 38 packages
cd api && go tool sqlc generate  # idempotent
cd web && npm test -- --watch=false --browsers=ChromeHeadless  # 893/893 SUCCESS
cd web && npx tsc --noEmit -p tsconfig.json  # 0 new errors (25 pre-existing TS2307 module-resolution errors unrelated to this work)
```

The dev database has the new schema applied via `make migrate-up`. The down migration in `000034_refactor_to_child_management_model.down.sql` reverses the change in full.
