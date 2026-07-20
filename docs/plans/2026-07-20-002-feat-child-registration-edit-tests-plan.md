---
title: Child Registration Edit-Mode Tests - Plan
type: feat
date: 2026-07-20
topic: child-registration-edit-tests
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-brainstorm
execution: code
---

## Goal Capsule

- **Objective:** Add Playwright e2e tests that verify the child registration edit flow — all fields load correctly when editing a previously registered child, and changes persist after save.
- **Product authority:** User request — fill coverage gap in existing e2e test suite.
- **Open blockers:** None.

## Product Contract

### Summary

Create a dedicated Playwright spec (`child-registration-edit.spec.ts`) that registers a child through the full 4-step wizard, navigates to the edit page, verifies every field loads with the correct value, modifies fields, saves, and verifies the changes persist on reload.

Product Contract unchanged — requirements carried forward from brainstorm.

### Requirements

**R1. New spec file:** `web/e2e/child-registration-edit.spec.ts` — covers the edit round-trip for a registered child.

**R2. Test data setup:** Each test registers a child via the full UI wizard (reusing existing page object helpers), then navigates to the edit page for that child. No API seeding — tests validate the full UI round-trip.

**R3. Field coverage — Step 1 (Child Basics):** After navigating to edit, verify these fields load with the values entered during registration:
- `first_name`, `last_name`, `date_of_birth`, `start_date`
- `home_address`, `first_language`, `sex`
- `primary_room_id` (selected room)
- `disability_status` (radio)

**R4. Field coverage — Step 2 (Medical & Health):** Verify:
- `allergy_status` radio, `allergy_details` textarea (when yes)
- `medication_status` radio, `medication_name`/`medication_dosage` inputs (when yes)
- `immunisation_status` radio
- `dietary_status` radio, `special_dietary_requirements` (when details)
- `medical_history_status` radio
- `social_services_status` radio
- `developmental_concerns` gate + individual concern checkboxes (when yes)

**R5. Field coverage — Step 3 (Contacts & Collection):** Verify:
- Primary parent/carer: `fullName`, `relationship`, `telephone`, `email`, `hasResponsibility`, address fields
- Emergency contact: `fullName`, `relationship`, `telephone`
- `collection_password`

**R6. Field coverage — Step 4 (Consents):** Verify:
- All 6 required consents load as checked
- `signer_name` loads correctly
- Optional consents that were toggled load correctly

**R7. Edit-and-save round-trip:** After verifying initial load, modify a subset of fields (e.g., change first name, update allergy details, change parent phone), save via "Save changes" button, reload the edit page, and verify the modified values persist.

**R8. Page object updates:** Add to `ChildRegistrationPage`:
- `navigateToEditChild(childId: string)` — navigates to `/manager/children/:childId/edit`
- `expectFieldValue(fieldId: string, expectedValue: string)` — asserts an input/select has the expected value
- `expectRadioSelected(name: string, value: string)` — asserts a radio button is selected
- `expectCheckboxChecked(fieldId: string)` — asserts a checkbox is checked

**R9. Helper updates:** Add to `web/e2e/helpers/auth.ts`:
- `registerChildAndGetId(page: Page)` — runs the full registration wizard and returns the child ID from the resulting URL

### Scope Boundaries

- No changes to the Angular application code — tests only.
- No changes to existing test files — new spec + page object additions only.
- No API-level assertions — this is purely UI/UX integration testing.
- Photo upload testing is out of scope (separate concern).
- Draft restore/persist testing is out of scope (separate concern).

### Out of Scope (Deferred)

- Testing the "Mark Reviewed" / "Complete" workflow transitions.
- Testing concurrent edits or conflict resolution.
- Testing the booking pattern or funding sections of the edit page.

### Acceptance Criteria

1. `npx playwright test child-registration-edit` passes with all tests green.
2. Each test registers a child via UI, navigates to edit, and verifies field values.
3. The edit-and-save test modifies fields, saves, reloads, and confirms changes persisted.
4. No existing tests break — `npx playwright test` still passes.

## Planning Contract

### Key Technical Decisions

**KTD1: Extend existing page object, don't create a new one.** The `ChildRegistrationPage` at `web/e2e/page-objects/child-registration.page.ts` already has all the form interaction methods (fill, select, click). Edit-mode tests need the same interactions plus assertion helpers. Adding `navigateToEditChild()`, `expectFieldValue()`, `expectRadioSelected()`, and `expectCheckboxChecked()` to the existing page object keeps one place for form selectors.

**KTD2: Capture child ID from URL after registration.** After the registration wizard completes, the app navigates to `/manager/children/:childId`. The `registerChildAndGetId` helper will extract the child ID from the URL using `page.url()` and regex. This avoids API calls and keeps the test purely UI-driven.

**KTD3: Two test groups — field verification and edit round-trip.** The spec organizes tests into two `test.describe` blocks: one that verifies all fields load correctly on edit (read-only assertions), and one that modifies fields, saves, reloads, and verifies persistence. This separation makes failures easy to diagnose.

**KTD4: Use flatpickr API for date field assertions.** Date picker fields (date of birth, start date) use flatpickr. Assert values via the flatpickr instance (`input._flatpickr.selectedDates`) rather than `input.value` to avoid format mismatches.

### Assumptions

- The edit page at `/manager/children/:childId/edit` uses the same stepper component as registration (`ManagerChildEditStepperComponent`), so the same page object selectors apply.
- The "Save changes" button triggers an API save and shows a success toast — tests will wait for the toast to confirm save completed.
- After save, the edit page remains open (no redirect) — tests can reload the same URL to verify persistence.

### Sequencing

1. U1 (page object helpers) — no dependencies, enables all other units
2. U2 (registerChildAndGetId helper) — no dependencies, enables test setup
3. U3 (field verification tests) — depends on U1 and U2
4. U4 (edit round-trip test) — depends on U1 and U2

## Implementation Units

### U1. Add edit helper methods to ChildRegistrationPage

**Goal:** Extend the page object with methods needed for edit-mode testing.

**Requirements:** R8

**Files:** `web/e2e/page-objects/child-registration.page.ts`

**Approach:** Add four methods to `ChildRegistrationPage`:
- `navigateToEditChild(childId)` — calls `page.goto(\`/manager/children/${childId}/edit\`)` and waits for the stepper to be visible
- `expectFieldValue(id, expectedValue)` — locates `input#${id}` or `select#${id}` and asserts its value
- `expectRadioSelected(name, value)` — asserts `label[for="${name}-${value}"]` has the `aria-checked` attribute or the radio input is checked
- `expectCheckboxChecked(id)` — asserts the checkbox input is checked

**Test Scenarios:**
- Methods exist and are callable (smoke — actual assertions tested in U3/U4)

**Verification:** TypeScript compiles without errors (`npx tsc --noEmit` in `web/`).

### U2. Add registerChildAndGetId helper

**Goal:** Provide a reusable helper that registers a child via UI and returns the child ID.

**Requirements:** R9

**Files:** `web/e2e/helpers/auth.ts`

**Approach:** Add `registerChildAndGetId(page: Page): Promise<string>` that:
1. Calls `navigateToNewRegistration(page)` (existing helper)
2. Creates a `ChildRegistrationPage` instance
3. Fills all 4 steps using existing page object methods (same flow as `child-registration-happy-path.spec.ts`)
4. Submits the registration
5. Waits for navigation away from `/manager/children/new`
6. Extracts child ID from the URL: `page.url().match(/\/manager\/children\/([^/]+)/)?.[1]`
7. Returns the child ID

**Test Scenarios:**
- Helper returns a valid UUID-format string after registration

**Verification:** TypeScript compiles; helper is importable from test files.

### U3. Create field verification tests

**Goal:** Verify that all fields load with correct values when editing a previously registered child.

**Requirements:** R1, R2, R3, R4, R5, R6

**Files:** `web/e2e/child-registration-edit.spec.ts`

**Approach:** Create a `test.describe('Child Registration - Edit Field Verification')` block with:
- `beforeEach`: register a child via UI, get child ID, navigate to edit page
- **Step 1 fields test:** Assert first name, last name, date of birth, start date, home address, language, sex, room, and disability status all load with the registered values
- **Step 2 fields test:** Navigate to step 2 in edit mode, assert allergy status, medication status, immunisation status, dietary status, social services status load correctly
- **Step 3 fields test:** Navigate to step 3, assert parent/carer name, relationship, phone, email, responsibility, address, emergency contact, and collection password load correctly
- **Step 4 fields test:** Navigate to step 4, assert all 6 required consents are checked and signer name is populated

**Test Scenarios:**
- Step 1 fields match registration input
- Step 2 fields match registration input
- Step 3 fields match registration input
- Step 4 consents and signer name match registration input

**Verification:** `npx playwright test child-registration-edit --grep "Field Verification"` passes.

### U4. Create edit-and-save round-trip test

**Goal:** Verify that changes made in edit mode persist after save and reload.

**Requirements:** R7

**Files:** `web/e2e/child-registration-edit.spec.ts`

**Approach:** Create a `test.describe('Child Registration - Edit and Save')` block with:
- `beforeEach`: register a child via UI, get child ID, navigate to edit page
- **Modify and save test:** Change first name to a new value, change parent phone number, click "Save changes", wait for success toast, reload the page, assert the modified values persist
- **Modify allergy details test:** Navigate to step 2, change allergy status to "yes" and add details, save, reload, assert allergy details persist

**Test Scenarios:**
- Modified first name persists after save + reload
- Modified parent phone persists after save + reload
- Allergy details persist after save + reload

**Verification:** `npx playwright test child-registration-edit --grep "Edit and Save"` passes.

## Verification Contract

**Test commands:**
- `npx playwright test child-registration-edit` — runs all new edit tests
- `npx playwright test` — runs full suite to verify no regressions

**Quality gates:**
- All new tests pass
- No existing tests break
- TypeScript compiles: `npx tsc --noEmit` in `web/`

## Definition of Done

- [ ] `web/e2e/child-registration-edit.spec.ts` created with field verification and edit round-trip tests
- [ ] `web/e2e/page-objects/child-registration.page.ts` extended with edit helper methods
- [ ] `web/e2e/helpers/auth.ts` extended with `registerChildAndGetId`
- [ ] `npx playwright test child-registration-edit` passes
- [ ] `npx playwright test` full suite passes
- [ ] No Angular application code changed
