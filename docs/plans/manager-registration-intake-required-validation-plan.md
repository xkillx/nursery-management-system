# Manager Registration Intake Required Validation Implementation Plan

## Goal

Make the manager-assisted registration intake stepper enforce required validation before managers can continue to the next step or jump ahead in the progress nav. When a manager clicks a continue or completion action with missing or invalid required information, the workflow must keep them on or route them to the relevant step, show an error toast, and focus the first invalid input or control.

The actor is a nursery manager transcribing a physical registration form. The business outcome is that guided intake creates reviewed/complete registrations only after required child, health, contacts, funding, consent, and evidence answers have been explicitly captured, with no manager-facing `Unknown` choice accepted as a completed required answer.

## Non-Goals

- Do not change backend API contracts, database schema, or stored enum support for existing `unknown` values.
- Do not remove `unknown` fallback parsing for legacy records, restored drafts, or API payload tolerance.
- Do not change parent self-service flows; this plan only covers `web/src/app/features/staff/pages/manager-registration-intake`.
- Do not change the existing child detail registration editor unless a shared type or helper requires a compile fix.
- Do not introduce file upload, evidence scanning, or digital storage for paper evidence.
- Do not change permission boundaries, routing guards, or manager/practitioner/parent access rules.

## Context Alignment

`CONTEXT.md` defines Manager-Assisted Registration Intake as staff transcription from a physical form and Guided Registration Intake as a stepper aligned to that transcription workflow. This plan keeps that workflow.

`CONTEXT.md` previously defined blank or missing health/social-care/development/dietary/medication information as unknown or incomplete. The interview resolved that `Unknown` should not remain a manager-facing completed answer in the guided intake path. `CONTEXT.md` now includes Guided Registration Required Answer: managers must record definite answers before moving past required sections, while `unknown` remains a legacy or incomplete-data meaning.

`CONTEXT.md` also defines Registration Reviewed/Complete Requirements and Paper-Form Consent Completeness. This plan applies those requirements earlier at step continuation, not only at final submit.

## Current State

- The target page is `web/src/app/features/staff/pages/manager-registration-intake/manager-registration-intake.component.ts` with template `manager-registration-intake.component.html` and tests `manager-registration-intake.component.spec.ts`.
- The component uses template-driven forms via `FormsModule`, not Angular reactive forms.
- The stepper has four `StepperStep` values: `child-basics`, `medical-health`, `contacts-collection`, and `consents-evidence`.
- `canOpenStep(step)` currently returns `true` for all steps in a new registration, so managers can jump ahead before prior required fields are complete.
- `saveChildBasics()` sets `step1Submitted = true` but does not stop when `step1FieldError()` reports missing required first name, surname, date of birth, or start date.
- `focusFirstStep1Error()` already focuses first invalid step-1 native date/text input, but only covers the four `Step1RequiredField` values.
- `saveMedicalHealth()` and `saveContactsCollection()` advance new registrations without checking the existing completion issue collector.
- `saveConsentsEvidence()` checks signer name/date and safeguarding acknowledgement only, while final completion uses broader validation.
- `collectFinalCompletionIssues()` already centralizes most required completion rules and returns `FinalCompletionIssue` objects with `stepKey`, `field`, and `message`.
- `routeToFirstIssue()` currently changes step and focuses the step heading, not the offending field.
- The template exposes `Unknown` options through `yesNoUnknownOptions`, `disabilityStatusOptions`, `noneDetailsUnknownOptions`, `evidenceStatusOptions`, and `waitingListOptions`.
- Office evidence defaults several statuses to `unknown`, and an existing spec says unknown office statuses are allowed for follow-up. This must change for guided intake completion/continuation.
- A shared toast service exists at `web/src/app/shared/services/toast.service.ts` with `error(message, options?)`, and other staff pages inject it from `../../../../shared/services/toast.service`.
- `web/package.json` runs Angular tests through `npm test` / `ng test` from the `web` directory.

## Decisions

Product/domain decisions from the interview:

- Remove manager-facing `Unknown` state from guided intake required choices.
- Managers cannot continue to the next step until required validation for the current step is complete.
- If validation fails after a continue or completion click, the system must show a toast error notification.
- If validation fails after a continue or completion click, the system must auto-focus the first invalid input/control.
- Existing `unknown` data may still be tolerated internally for legacy records, restored drafts, or API compatibility, but must be treated as incomplete in the guided intake UI.

Implementation decisions made by the agent:

- Reuse `collectFinalCompletionIssues()` as the canonical validation source, filtered by step for step continuation.
- Add a field-focus mapping for all required fields that can block each step; use element IDs in the template and a component helper to focus and scroll.
- Use `ToastService.error()` for validation failures while keeping `errorMessage` for page-level/API errors unless local conventions require both.
- Keep disabled state off the final Complete button or make the click path still callable so users receive toast and focus feedback instead of a silent disabled button.
- Preserve parsing/build payload fallback methods that convert missing values to `unknown` where backend contracts currently expect that fallback.

Existing architectural/documentation decisions to honor:

- Registration child creation for new guided intake is frontend-local until final atomic submission.
- Add Child and the section-by-section registration editor are unaffected.
- Registration contact entries do not automatically create guardians or parent portal access.
- Office-use checklist is separate from registration profile data but may be completed in the same workflow.

## Acceptance Criteria

- The progress nav does not allow opening a later step until all required validation for every prior step is complete.
- Clicking a continue button on a step with missing or invalid required fields keeps the manager on that step, shows an error toast, and focuses the first invalid field/control.
- Step 1 blocks on missing first name, surname, date of birth, proposed start date, home address, primary home language, missing disability/SEND answer, and missing disability/access details when disability/SEND is yes.
- Step 2 blocks on blank or legacy `unknown` allergy, medication, dietary, medical history, and social-services answers; it also blocks on missing required details when the answer is yes/details.
- Step 3 blocks on missing primary parent/carer name, relationship, phone, parental responsibility answer, missing valid emergency contact, missing collection password when a non-parent authorised collector is present, and missing/invalid funding answer/details.
- Step 4 blocks on missing signer name, confirmation date, mandatory acknowledgements, unreviewed optional consent decisions, and every required office evidence status not being explicitly `complete`, `missing`, or `not_applicable`.
- Manager-facing required option groups no longer display `Unknown` in the guided intake UI.
- Legacy/restored `unknown` values do not crash the page; they render as incomplete, force a new explicit selection, and cannot be submitted as a completed required answer.
- Automated tests cover blocked step continuation, toast error calls, field focus behavior, removed unknown choices, and changed office evidence validation.

## Implementation Tasks

### Task 1: Add Toast Service and Step Validation Helpers

- Objective: Provide one validation path for continue, nav jumps, and final submit failures.
- Depends on: None.
- Target files/symbols: `manager-registration-intake.component.ts`, `ToastService`, `collectFinalCompletionIssues()`, `FinalCompletionIssue`.
- Required changes:
  - Import and inject `ToastService` from `../../../../shared/services/toast.service`.
  - Add a helper like `issuesForStep(step: StepperStep): FinalCompletionIssue[]` that filters `collectFinalCompletionIssues()` by step.
  - Add a helper like `firstBlockingIssueForStep(step: StepperStep): FinalCompletionIssue | null`.
  - Add a helper like `handleValidationFailure(issue: FinalCompletionIssue, options?)` that sets `finalCompletionIssues`, sets `currentStep` to `issue.stepKey` when needed, calls `toast.error(issue.message, { title: 'Check required details' })`, and focuses the invalid control after the DOM updates.
  - Keep existing API error handling with `errorMessage`; local validation failures should use toast and focus.
- Tests/verification:
  - Unit test that a validation failure calls `ToastService.error()` with the first issue message.
- Expected outcome:
  - Local validation has reusable behavior and no save/continue path silently advances past invalid required fields.

### Task 2: Remove Manager-Facing Unknown Options

- Objective: Remove `Unknown` as a selectable completed state in guided intake required controls.
- Depends on: Task 1.
- Target files/symbols: `yesNoUnknownOptions`, `disabilityStatusOptions`, `noneDetailsUnknownOptions`, `evidenceStatusOptions`, `waitingListOptions`, `manager-registration-intake.component.html`.
- Required changes:
  - Replace `yesNoUnknownOptions` usage for required yes/no questions with `yesNoOptions`, or change the option list to yes/no only where used by required guided intake controls.
  - Change `disabilityStatusOptions` to yes/no only.
  - Change `noneDetailsUnknownOptions` to `none` and `details` only.
  - Change `evidenceStatusOptions` to `complete`, `missing`, and `not_applicable` only.
  - Remove `unknown` from `waitingListOptions` for newly added professional referrals; use `not_applicable` as the default for optional referral status unless a better existing non-unknown default is already present.
  - Leave TypeScript coercion/parsing methods that accept `unknown` for loaded legacy data, but ensure a legacy `unknown` value leaves no visible option selected or is visibly invalid until changed.
- Tests/verification:
  - Unit test option arrays do not include `unknown` for required guided intake controls.
  - Existing specs that expect unknown to block should remain valid or be tightened to assert hidden/legacy unknown blocks.
- Expected outcome:
  - Managers cannot choose `Unknown` in the guided intake UI, while legacy data remains safely handled.

### Task 3: Enforce Step 1 Required Validation Before Continuing

- Objective: Make child basics continuation require all step-1 required completion rules.
- Depends on: Task 1.
- Target files/symbols: `saveChildBasics()`, `step1FieldError()`, `collectFinalCompletionIssues()`, `focusFirstStep1Error()`, `manager-registration-intake.component.html`.
- Required changes:
  - Before setting `isSaving = true`, check `issuesForStep('child-basics')`.
  - If issues exist, mark step 1 submitted/touched as appropriate, call validation failure handling, and return without saving or advancing.
  - Expand focus support beyond the four existing `Step1RequiredField` entries to cover home address, first language select, disability status radio group, disability notes, and access requirements.
  - Add stable IDs and `aria-describedby`/error hooks to the corresponding template controls where missing.
- Tests/verification:
  - Unit test that missing first name/home address/disability status prevents `currentStep` from changing.
  - Unit test that the first invalid step-1 field receives focus after Continue.
- Expected outcome:
  - Step 1 cannot advance until required child profile details are complete.

### Task 4: Enforce Step 2 Required Validation Before Continuing

- Objective: Block medical and health continuation until all required health and safeguarding answers are explicit and complete.
- Depends on: Tasks 1 and 2.
- Target files/symbols: `saveMedicalHealth()`, `collectMedicalSafetyIssues()`, `collectNoneDetailsIssue()`, medical-health template fieldsets and inputs.
- Required changes:
  - At the start of `saveMedicalHealth()`, check `issuesForStep('medical-health')`.
  - Treat blank and legacy `unknown` statuses as blocking issues.
  - Preserve detail requirements for allergy yes, medication yes, dietary details, medical history details, and social-services yes.
  - Add stable IDs/focus targets for required radio groups and detail inputs/areas.
- Tests/verification:
  - Unit tests for blank and legacy unknown statuses blocking `saveMedicalHealth()`.
  - Unit test that a missing allergy detail focuses the allergy detail textarea/input.
- Expected outcome:
  - Step 2 cannot advance until health/safety-required fields are complete.

### Task 5: Enforce Step 3 Required Validation Before Continuing

- Objective: Block contacts and collection continuation until required parent/carer, emergency, collection, and funding answers are complete.
- Depends on: Task 1.
- Target files/symbols: `saveContactsCollection()`, `collectContactsIssues()`, `collectFundingIssues()`, funding helpers, contacts template.
- Required changes:
  - At the start of `saveContactsCollection()`, check `issuesForStep('contacts-collection')`.
  - Ensure funding validation uses the same required rules as final completion.
  - Add stable IDs/focus targets for primary parent/carer fields, parental responsibility radios, emergency contact group, collection password, funding yes/no, funding options, and other benefits.
  - If a group-level issue has no single input, focus the first required control in that group.
- Tests/verification:
  - Unit tests for missing parent phone, parental responsibility, emergency contact, collection password, and funding answer blocking continuation.
  - Unit test that the first invalid contacts field receives focus.
- Expected outcome:
  - Step 3 cannot advance until contacts, collection, and funding required data is complete.

### Task 6: Enforce Step 4 Validation and Final Submit Feedback

- Objective: Make completion click provide toast/focus feedback and block all incomplete consent/evidence requirements.
- Depends on: Tasks 1 and 2.
- Target files/symbols: `saveConsentsEvidence()`, `submitRegistration()`, `collectConsentsIssues()`, `collectOfficeEvidenceIssues()`, `canSubmitLocally()`, consents/evidence template.
- Required changes:
  - At the start of `saveConsentsEvidence()`, check `issuesForStep('consents-evidence')` for existing registrations when saving evidence/consents.
  - In `submitRegistration()`, replace heading-only routing with validation failure handling that focuses the first invalid field after switching steps.
  - Remove or revise `[disabled]="!canSubmitLocally() || isSaving"` on the new-registration Complete button so clicking it can trigger toast and focus. Keep `isSaving` disabled.
  - Update office evidence validation so blank and legacy `unknown` statuses block completion; accepted statuses are `complete`, `missing`, and `not_applicable`.
  - Add stable IDs/focus targets for signer name/date, mandatory acknowledgements, optional consent decision cards, and every required office evidence status select.
- Tests/verification:
  - Update the spec that currently allows unknown office statuses so it expects blocking behavior.
  - Unit test `submitRegistration()` shows a toast and focuses the first failing field.
  - Unit test Complete button remains clickable for invalid forms and blocked by validation rather than only disabled.
- Expected outcome:
  - Final completion gives actionable feedback and no required evidence/consent state can be skipped.

### Task 7: Lock Step Navigation by Prior-Step Completion

- Objective: Prevent progress-nav jumps past incomplete required steps.
- Depends on: Tasks 3, 4, 5, and 6.
- Target files/symbols: `canOpenStep()`, `goToStep()`, `stepIsComplete()`, stepper template.
- Required changes:
  - For new registration, allow opening the current step, previous steps, and the next/later step only if all prior steps have zero validation issues.
  - For existing registration, preserve child-id requirements but still apply prior-step required validation when moving forward.
  - When a locked future step is clicked, route/focus to the first incomplete prior step and show a toast error.
  - Update `stepIsComplete(step)` to reflect validation completion for that specific step, not merely index position, so the visual check mark does not lie.
  - Keep Back navigation unrestricted to prior steps.
- Tests/verification:
  - Unit tests for `canOpenStep()` with incomplete previous steps.
  - Unit test that clicking/jumping ahead calls toast and does not change to the requested future step.
- Expected outcome:
  - The stepper’s visual state and navigation behavior match required validation.

### Task 8: Add Accessible Error and Focus Wiring

- Objective: Make validation feedback usable for keyboard and assistive technology users.
- Depends on: Tasks 3 through 7.
- Target files/symbols: `manager-registration-intake.component.html`, focus helper in component TS.
- Required changes:
  - Add stable `id` attributes to every required input/select/radio group target used by the focus map.
  - For group controls, focus the first radio/selectable control and scroll the containing fieldset into view.
  - Add `aria-invalid` and `aria-describedby` for controls where visible error state is added.
  - Avoid layout shifts from error text by following existing `app-form-field` patterns where possible.
- Tests/verification:
  - Unit tests can spy on `document.querySelector`/focus for representative input, select, and radio group targets.
  - Manual validation with keyboard only.
- Expected outcome:
  - Toast gives global feedback, focus gives exact next action, and screen-reader semantics are not regressed.

### Task 9: Update and Run Tests

- Objective: Prove the changed validation behavior.
- Depends on: Tasks 1 through 8.
- Target files/symbols: `manager-registration-intake.component.spec.ts`.
- Required changes:
  - Provide or spy `ToastService` in the component spec.
  - Add tests for each step continuation method blocking invalid data.
  - Update unknown-related tests so legacy `unknown` remains blocked but no longer appears in option arrays.
  - Update office evidence unknown test from allowed to blocked.
  - Add focus tests for first invalid field routing.
- Tests/verification:
  - Run from repo root: `cd web && npm test -- --watch=false --browsers=ChromeHeadless`.
  - If the local environment lacks ChromeHeadless, run the same test command with the browser configured by the project or document the local browser failure and still run `cd web && npm run build`.
- Expected outcome:
  - Registration intake specs cover the new behavior and Angular build/test passes locally.

## Contracts

- UI validation contract: Required guided intake answers must be explicit before forward navigation. Blank and legacy `unknown` count as incomplete.
- Toast contract: Local validation failures call `ToastService.error(message, { title: 'Check required details' })` or equivalent existing toast option shape.
- Focus contract: A failed continue/nav/complete action focuses and scrolls the first invalid control for the first blocking issue.
- API contract: Existing outgoing payload structure from `buildCompleteRegistrationPayload()` is preserved. Fallbacks to `unknown` may remain only for compatibility when values are missing internally, but the UI must block those missing values before normal submission.
- Data contract: Stored/restored drafts containing `unknown` must not crash. They must be treated as incomplete until the manager chooses an accepted answer.
- Permission contract: No changes to manager, practitioner, parent, or owner permissions.

## Files to Change

- `CONTEXT.md`: already updated with Guided Registration Required Answer.
- `docs/plans/manager-registration-intake-required-validation-plan.md`: this implementation plan.
- `web/src/app/features/staff/pages/manager-registration-intake/manager-registration-intake.component.ts`: inject toast, add validation/focus helpers, remove user-facing unknown option values, enforce step validation, update navigation gating.
- `web/src/app/features/staff/pages/manager-registration-intake/manager-registration-intake.component.html`: add IDs/ARIA/error display hooks, remove unknown option usage from required controls, keep final completion action clickable for validation feedback.
- `web/src/app/features/staff/pages/manager-registration-intake/manager-registration-intake.component.spec.ts`: update and add tests.

## Verification

Run from the repository root:

```sh
cd web && npm test -- --watch=false --browsers=ChromeHeadless
```

If the test runner cannot launch ChromeHeadless in the local environment, run:

```sh
cd web && npm run build
```

Manual validation scenarios:

- Start a new registration, leave Step 1 empty, click Continue. Expect an error toast and focus on first name.
- Fill only Step 1’s original four required fields but leave home address, language, and disability/SEND answer blank. Expect Step 1 still blocks and focuses the first missing required control.
- Try to click Step 3 in the stepper while Step 2 is incomplete. Expect the app to keep/route to Step 2, show a toast, and focus the first incomplete medical/safety answer.
- On Step 2, verify allergy, medication, medical history, dietary, social-services, disability, and office evidence required controls do not offer `Unknown`.
- On Step 3, leave funding answer blank and click Next. Expect toast and focus on funding yes/no.
- On Step 4, leave office evidence statuses at blank or legacy `unknown`. Expect Complete Registration to show toast, focus the first evidence status, and not submit.
- Restore a draft or load existing data with `unknown`; verify the page renders, the field is treated as incomplete, and selecting an accepted value allows progress.

## Assumptions

- The implementation should keep template-driven forms because the existing component is built around `FormsModule`.
- Toasts are globally displayed because `app-toast-container` is already mounted in `web/src/app/shared/layout/app-layout/app-layout.component.html`.
- Existing backend support for `unknown` values is still needed for backward compatibility and should not be removed.
- For optional professional referrals, `not_applicable` is an acceptable replacement default for previously unknown waiting-list status.
- A group-level validation issue can focus the first meaningful control in that group when there is no single exact field.

## Risks and Fallbacks

- Risk: Some custom controls such as `app-select` or `app-date-picker` may not expose a directly focusable element by the wrapper ID.
  - Fallback: Focus the first native `input`, `select`, `button`, or `[tabindex]` inside the host element and scroll the host element into view.
- Risk: Removing `Unknown` from option arrays may make legacy loaded values invisible.
  - Fallback: Show no selected option and rely on validation error text/toast to require a new explicit answer, rather than reintroducing `Unknown` as a selectable option.
- Risk: The existing Complete button disabled state may prevent user-triggered feedback.
  - Fallback: Disable only while saving, keep validation inside `submitRegistration()`, and style incomplete state through visible errors rather than disabling the action.
- Risk: Centralizing validation around `collectFinalCompletionIssues()` may make `canOpenStep()` expensive because it recomputes all issues.
  - Fallback: Keep the computation local and simple first; if performance becomes visible, add step-specific issue collectors that reuse the same rule functions.
- Risk: Existing specs rely on office evidence `unknown` being allowed.
  - Fallback: Update those specs to the new product decision and keep separate tests proving legacy unknown is tolerated but incomplete.
