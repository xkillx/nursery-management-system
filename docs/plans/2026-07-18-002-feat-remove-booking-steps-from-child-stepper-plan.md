# Remove Booking Steps from Child Registration/Edit Stepper - Plan

**Date:** 2026-07-18
**Artifact contract:** ce-unified-plan/v1
**Artifact readiness:** requirements-only
**Product contract source:** ce-brainstorm

## Goal Capsule

**Objective:** Remove step 5 (Session Pattern) and step 6 (Funding & Benefits) from the child registration and edit stepper, reducing it from 6 steps to 4. Booking and funding data will be managed exclusively via the dedicated booking page after the child exists.

**Product authority:** User decision — the dedicated booking pattern page (`manager/children/:childId/booking-pattern`) already exists and covers step 5 functionality. Funding can be managed post-creation.

**Open blockers:** None.

## Product Contract

### What we're building

Remove the session-pattern and funding-benefits steps from the `ManagerChildEditStepperComponent`. The stepper will contain only:

1. Child Details (`child-basics`)
2. Medical & Health (`medical-health`)
3. Contacts & Security (`contacts-collection`)
4. Permissions & Consents (`consents-evidence`)

### Behavior changes

| Scenario | Current | After |
|----------|---------|-------|
| New registration | 6-step stepper, child + booking + funding created atomically | 4-step stepper, child created with steps 1-4 only |
| Edit child | 6-step stepper, all data editable | 4-step stepper, booking/funding editable via dedicated pages |
| Step navigation | Steps 5-6 in progress bar | Steps 5-6 removed from progress bar |
| Draft auto-save | Saves all 6 steps to localStorage | Saves steps 1-4 only |
| Final save (new) | `createChildFromSessionPatternStep()` includes booking_pattern + funding payload | Save at step 4, no booking/funding in payload |

### Scope boundaries

**In scope:**
- Remove step 5 and 6 templates from `.html`
- Remove step 5 and 6 state, validation, and save methods from `.ts`
- Update `StepperStep` type to remove `'session-pattern'` and `'funding-benefits'`
- Update step metadata array (labels, descriptions, icons)
- Update progress bar to show 4 steps
- Update draft storage to exclude steps 5-6
- Update final save payload for new registration (remove `booking_pattern` and funding fields)
- Update step navigation (next/previous/save logic)
- Remove session types API loading from stepper init

**Out of scope:**
- Changes to the dedicated booking pattern page
- Changes to funding API endpoints
- Backend changes
- Removing booking/funding models or API services (used elsewhere)

### Files affected

- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.html`
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.spec.ts`
- `web/src/app/features/staff/data/registration-draft.storage.ts` (if it encodes step-specific data)

### Outstanding questions

- Should the child detail/overview page show a link or prompt to set up booking pattern if none exists yet?
