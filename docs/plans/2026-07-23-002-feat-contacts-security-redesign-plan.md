---
title: Contacts & Security Step Redesign - Plan
type: feat
date: 2026-07-23
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
product_contract_source: ce-plan-bootstrap
execution: code
---

## Goal Capsule

Redesign the "Contacts & Security" step in the child registration/edit stepper to support multiple parent/guardian linking with expanded fields, improved UX, and stricter validation. Replace the single-parent selector with a multi-parent card list using a unified search/create combobox. Add address and safeguarding flags to parent cards. Enforce required sections (parent + emergency contact + password) before step completion.

**Authority hierarchy:** Product decisions from user interview (2026-07-23). Architecture follows existing Clean Architecture patterns in this codebase.

**Stop conditions:** All implementation units complete, `go build ./...` passes, `ng build` passes, `ng lint` passes, multi-parent linking functional, validation enforces all three sections, draft auto-save works for parent changes.

**Execution profile:** Sequential units with dependency ordering. Each unit is one atomic commit.

**Tail ownership:** After ship, monitor that parent portal login still works, billing reads parent data correctly, and emergency contacts remain functional.

---

## Product Contract

### Summary

The Contacts & Security step (step 3) in the child registration/edit stepper is redesigned to support linking multiple parents/guardians (up to 4) with expanded fields including address and safeguarding flags. A unified search/create combobox replaces the current dropdown + inline form pattern. Parent cards show read-only summaries by default with inline edit capability. Emergency contacts and collection password sections remain separate. All three sections are required before proceeding.

### Problem Frame

Currently, the Contacts & Security step allows selecting only one parent/guardian via a dropdown. The inline parent creation form captures minimal fields (name, email, phone, relationship). Address and safeguarding flags (parental responsibility, can pick up, emergency contact) exist on the parent model but are not exposed in the registration form. There is no way to link multiple parents during registration. The step does not enforce emergency contact or password requirements.

### Requirements

**Multi-Parent Linking**

- R1. Staff can link up to 4 parents/guardians to a child from the Contacts & Security step.
- R2. Each linked parent displays as a card with: name, relationship, phone, email, address, and three flag badges (parental responsibility, can pick up, emergency contact).
- R3. Cards are read-only by default; clicking "Edit" expands the card for inline editing.
- R4. A unified search/create combobox replaces the dropdown + "Add new parent" button. Searching filters out already-linked parents. "Create new: [name]" appears when no exact match exists.
- R5. Creating a new parent inline saves immediately via API, then expands the card for address/flag entry.
- R6. Removing a parent from an existing child shows a confirmation dialog. For new registrations, removal is instant.
- R7. Parent cards show flag badges (colored chips) for true values only: green=parental responsibility, blue=can pick up, orange=emergency contact.
- R8. Flags are edited via toggle switches in the expanded card view.
- R9. Address fields (line1, line2, city, postcode) appear in a 2x2 grid in the expanded card.
- R10. A "Same as child's home address" checkbox pre-fills address from step 1.
- R11. Default flag values for new parents: all true (has_parental_responsibility, can_pick_up, is_emergency_contact).
- R12. Parent-specific relationship dropdown: Mother, Father, Step-Mother, Step-Father, Grandmother, Grandfather, Guardian, Foster Carer, Legal Guardian, Other.
- R13. "Other" relationship shows a free text input for custom relationship.
- R14. At least one contact method (phone OR email) required per parent.
- R15. Fuzzy matching shows "Similar parents" section in dropdown to prevent duplicates.
- R16. Inactive parents shown with "Inactive" badge and "Reactivate and link?" prompt.
- R17. Portal status shown as subtle badge: "Portal Active" or "No Portal".
- R18. "Add parent" button in section header, disabled with tooltip when cap (4) reached.
- R19. Fresh combobox appears per addition (not persistent).
- R20. New parent card auto-expands in edit mode after creation.

**Emergency Contacts**

- R21. Emergency contacts section remains separate from parents (Ofsted categorisation).
- R22. No new fields added to emergency contacts.
- R23. At least one emergency contact required before proceeding.

**Collection Password**

- R24. Collection password stays in the Contacts & Security step.
- R25. Password validation: minimum 6 characters, not blank/spaces only.
- R26. Password required before proceeding.

**Validation & Save**

- R27. All three sections required: at least 1 parent + 1 emergency contact + password.
- R28. Save all at once with transaction rollback on failure.
- R29. API validation errors map to specific fields. Network errors show top banner.
- R30. Draft auto-save extends to parent changes: new parent drafts saved to localStorage, existing links re-fetched from API on restore.

**UX Polish**

- R31. Section order unchanged: Parent → Emergency Contacts → Password.
- R32. Simple completion indicator: checkmark when all filled.
- R33. Mobile: full-width stacked cards.
- R34. Initials avatar only (no photos).
- R35. No link metadata, notes, photo, work address, employment, language, or communication preference on card.
- R36. Edit/Remove actions only on card (no extra actions).
- R37. Remove button only visible in edit mode.
- R38. Failed save: parent stays in system, retry on next save.

### Scope Boundaries

**Deferred to Follow-Up Work**

- Parent card drag-and-drop reordering.
- Parent merge (duplicate detection and merge workflow).
- Parent photo upload support.
- Multiple phone/email fields per parent.
- Work address field on parent model.
- Job title/employer fields on parent model.
- Language preference field on parent model.
- Communication preference field on parent model.
- Manual step completion indicators (progress dots, issue count).

**Outside this product's identity**

- Parent self-registration (parents are always created by staff).
- Parent card customization (themes, layouts).

---

## Planning Contract

### Key Technical Decisions

**KTD1. Multi-parent via parent_children table.**
The `parent_children` table already supports many-to-many relationships. No schema changes needed for linking. The UI changes from single-select to multi-card list. Rationale: reuses existing infrastructure, no migration required.

**KTD2. Unified search/create combobox replaces dropdown + inline form.**
A single combobox component handles both searching existing parents and creating new ones. The existing `parentsApi.list` supports search. The `parentsApi.create` handles creation. Rationale: faster UX, fewer clicks, consistent with modern CRM patterns.

**KTD3. Parent creation is immediate, linking is deferred to step save.**
New parents are created via `parentsApi.create` immediately when the user completes the inline form. The parent-to-child link is created when the step is saved. Rationale: parent is a valid entity on its own; linking failure doesn't orphan the parent.

**KTD4. Flags default to true for new parents.**
`has_parental_responsibility`, `can_pick_up`, and `is_emergency_contact` default to true. Rationale: matches the common case (parent registering the child has all rights). Manager can uncheck for edge cases.

**KTD5. Transaction-based save with rollback.**
All parent links, emergency contacts, and password save in a single transaction. If any part fails, everything rolls back. Rationale: prevents orphaned data, consistent with existing `saveContactsCollection` flow.

**KTD6. Draft auto-save for new parent drafts only.**
Existing parent links are re-fetched from API on draft restore. Only new parents (created inline but not yet persisted) are saved to localStorage. Rationale: prevents stale data, keeps draft payload small.

**KTD7. Confirmation dialog for unlinking in edit mode.**
For existing children, unlinking a parent shows a confirmation dialog. For new registrations, removal is instant (no link exists yet). Rationale: destructive action needs confirmation; new registration has no link to destroy.

### Dependencies

- `parentsApi.list` — search and filter parents (existing)
- `parentsApi.create` — create new parent (existing)
- `parentsApi.update` — update parent details (existing)
- `parentsApi.linkChild` — link parent to child (existing)
- `parentsApi.unlinkChild` — unlink parent from child (existing)
- `parentsApi.get` — get parent details (existing)
- `staffApi.putChildContacts` — save emergency contacts (existing)
- Collection password API — save password (existing)

### Risks

1. **Combobox complexity.** The unified search/create combobox is more complex than a simple dropdown. Mitigation: follow existing combobox patterns in the codebase, test thoroughly with keyboard navigation.
2. **Draft payload size.** Storing multiple parent drafts in localStorage could hit size limits. Mitigation: only store essential fields, compress where possible.
3. **Transaction failures.** If the parent link API fails, the whole step fails. Mitigation: clear error messages, retry capability, parent stays in system.

---

## Implementation Units

### Unit 1: Multi-Parent Data Model & State Management

**Goal:** Update the component's data model to support multiple parents with expanded fields.

**Files:**
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`

**Changes:**
1. Replace `selectedParentId: string` with `linkedParents: LinkedParentEntry[]` array.
2. Define `LinkedParentEntry` interface: `{ id: string; parentId: string; firstName: string; lastName: string; email: string; phone: string; relationship: string; addressLine1: string; addressLine2: string; addressCity: string; addressPostcode: string; hasParentalResponsibility: boolean; canPickUp: boolean; isEmergencyContact: boolean; portalStatus: 'active' | 'none'; isEditing: boolean; isNew: boolean; }`.
3. Add `parentSearchTerm: string`, `parentSearchResults: ParentRecord[]`, `isSearchingParents: boolean`.
4. Add `showCreateParentForm: boolean`, `newParentDraft: NewParentDraft`.
5. Add `linkedParentAddresses: string[]` (for address display).
6. Update `step3` object to remove parent-specific fields (address, flags) that are now on `LinkedParentEntry`.
7. Add `parentCap = 4` constant.
8. Add computed properties: `isParentCapReached`, `linkedParentIds`.

**Test scenarios:**
- LinkedParents array initializes empty.
- parentCap constant is 4.
- isParentCapReached returns true when 4 parents linked.
- linkedParentIds returns array of parent IDs.

---

### Unit 2: Unified Search/Create Combobox Component

**Goal:** Create a reusable combobox component that searches existing parents and offers inline creation.

**Files:**
- `web/src/app/shared/components/form/parent-combobox/parent-combobox.component.ts`
- `web/src/app/shared/components/form/parent-combobox/parent-combobox.component.html`
- `web/src/app/shared/components/form/parent-combobox/parent-combobox.component.spec.ts`

**Changes:**
1. Create `ParentComboboxComponent` with inputs: `excludeIds: string[]`, `branchId: string`.
2. Outputs: `parentSelected: EventEmitter<ParentRecord>`, `parentCreateRequested: EventEmitter<{ name: string }>`.
3. Implement debounced search (200ms) using `parentsApi.list`.
4. Show search results with differentiating info (email, phone).
5. Show "Similar parents" section for fuzzy matches.
6. Show "Create new: [name]" option when no exact match.
7. Show inactive parents with "Inactive" badge.
8. Handle keyboard navigation (arrow keys, enter, escape).
9. Implement ARIA attributes for accessibility.

**Test scenarios:**
- Search returns matching parents.
- Search filters out excluded IDs.
- "Create new" appears when no exact match.
- Fuzzy matches appear in "Similar parents" section.
- Inactive parents show "Inactive" badge.
- Keyboard navigation works (arrow keys, enter, escape).
- Debounce prevents excessive API calls.

---

### Unit 3: Parent Card Component

**Goal:** Create a reusable parent card component with collapsed/expanded states.

**Files:**
- `web/src/app/shared/components/ui/parent-card/parent-card.component.ts`
- `web/src/app/shared/components/ui/parent-card/parent-card.component.html`
- `web/src/app/shared/components/ui/parent-card/parent-card.component.spec.ts`

**Changes:**
1. Create `ParentCardComponent` with input: `parent: LinkedParentEntry`.
2. Outputs: `editRequested: EventEmitter<void>`, `removeRequested: EventEmitter<void>`, `saveRequested: EventEmitter<LinkedParentEntry>`, `cancelRequested: EventEmitter<void>`.
3. Collapsed view: avatar initials, name, relationship, phone, email, flag badges, edit button.
4. Expanded view: all fields editable (name, relationship, phone, email, address grid, flag toggles).
5. Flag badges: green for parental responsibility, blue for can pick up, orange for emergency contact. Only show for true values.
6. Flag toggles: three toggle switches in expanded view.
7. Address grid: 2x2 layout (line1, line2, city, postcode).
8. "Same as child's home address" checkbox with output: `useChildAddress: EventEmitter<boolean>`.
9. Relationship dropdown with parent-specific options.
10. "Other" relationship shows free text input.
11. Validation: at least one of phone/email required.
12. Portal status badge: "Portal Active" (green) or "No Portal" (grey).
13. Remove button only visible in expanded/edit mode.

**Test scenarios:**
- Collapsed view shows name, relationship, phone, email, flag badges.
- Expanded view shows all editable fields.
- Flag badges display correct colors for true values.
- Flag toggles update parent data.
- Address grid displays correctly.
- "Same as child's home address" pre-fills address.
- Relationship dropdown shows parent-specific options.
- "Other" relationship shows free text input.
- Validation requires phone OR email.
- Portal status badge shows correct status.
- Remove button only visible in edit mode.

---

### Unit 4: Parent Section Template & Logic

**Goal:** Replace the single-parent selector with multi-parent card list and combobox.

**Files:**
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.html`
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`

**Changes:**
1. Replace `selectedParentId` dropdown with `linkedParents` card list.
2. Add "Add parent" button in section header (disabled when cap reached).
3. Show `ParentComboboxComponent` when "Add parent" clicked.
4. On parent selected: add to `linkedParents` array, hide combobox.
5. On parent create requested: show inline create form with pre-filled name.
6. On parent created: add to `linkedParents` array in edit mode, hide combobox.
7. On edit requested: expand card.
8. On save requested: update parent via `parentsApi.update`, collapse card.
9. On remove requested: for new registrations, instant remove. For existing children, confirmation dialog then `parentsApi.unlinkChild`.
10. Update `loadAvailableParents()` to exclude linked parent IDs.
11. Update `loadChildAndStatus()` to load linked parents from `parentsApi.listParentsByChild`.
12. Update draft auto-save to handle `linkedParents` array.
13. Update draft restore to merge API-loaded parents with localStorage drafts.

**Test scenarios:**
- Parent section shows linked parents as cards.
- "Add parent" button shows combobox.
- Combobox filters out linked parents.
- Selecting parent adds card to list.
- Creating parent adds card in edit mode.
- Editing parent expands card.
- Saving parent updates API.
- Removing parent shows confirmation (edit mode) or instant (new registration).
- Cap at 4 parents enforced.
- Draft auto-save includes parent changes.
- Draft restore merges API and localStorage data.

---

### Unit 5: Validation & Save Logic

**Goal:** Update validation and save logic to enforce all three sections and handle multi-parent linking.

**Files:**
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`

**Changes:**
1. Update `step3FieldError` to check: at least 1 parent linked, at least 1 emergency contact, password filled.
2. Update `collectContactsIssues` to validate all three sections.
3. Update `saveContactsCollection` to:
   - Link all parents in `linkedParents` array via `parentsApi.linkChild`.
   - Save emergency contacts via `staffApi.putChildContacts`.
   - Save collection password via API.
   - Use transaction pattern: if any step fails, roll back previous steps.
4. Update `proceedWithContactsSave` to handle multi-parent linking.
5. Update `autoCreateParentAndContinue` to work with new combobox flow.
6. Update error handling: API validation errors map to fields, network errors show top banner.
7. Update `step3Submitted` flag to trigger validation display.

**Test scenarios:**
- Validation requires at least 1 parent.
- Validation requires at least 1 emergency contact.
- Validation requires password (min 6 chars, not blank).
- Save links all parents to child.
- Save creates parent-child links in correct order.
- Save rolls back on failure.
- API validation errors map to specific fields.
- Network errors show top banner.

---

### Unit 6: Draft Auto-Save for Parent Changes

**Goal:** Extend draft auto-save to handle multi-parent data.

**Files:**
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`

**Changes:**
1. Update `saveDraftToStorage` to include `linkedParents` array (new parents only).
2. Update `restoreDraftIfPresent` to:
   - Load existing linked parents from API.
   - Merge with localStorage drafts (new parents only).
   - Pre-select parent IDs from draft.
3. Update `subscribeToDraftAutoSave` to trigger on parent changes.
4. Update `notifyDraftChanged` to include parent data.
5. Add `isNewParent` flag to `LinkedParentEntry` to differentiate API-loaded vs draft parents.

**Test scenarios:**
- Draft saves new parent data to localStorage.
- Draft restores API-loaded parents correctly.
- Draft merges new parent drafts with API data.
- Draft pre-selects parent IDs from localStorage.
- Draft auto-triggers on parent changes.

---

### Unit 7: Mobile Responsiveness & Polish

**Goal:** Ensure parent cards and combobox work correctly on mobile.

**Files:**
- `web/src/app/shared/components/ui/parent-card/parent-card.component.html`
- `web/src/app/shared/components/form/parent-combobox/parent-combobox.component.html`
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.html`

**Changes:**
1. Parent cards: full-width stacked on mobile.
2. Combobox: full-width on mobile, dropdown below input.
3. Flag badges: wrap on small screens.
4. Address grid: stack on mobile (single column).
5. Toggle switches: full-width on mobile.
6. "Add parent" button: full-width on mobile.
7. Edit/Remove buttons: stack on mobile if needed.

**Test scenarios:**
- Parent cards stack vertically on mobile.
- Combobox is full-width on mobile.
- Flag badges wrap correctly.
- Address fields stack on mobile.
- Toggle switches are usable on mobile.
- Buttons are accessible on mobile.

---

## Test Strategy

### Unit Tests

- `ParentComboboxComponent`: search, filtering, keyboard navigation, create option.
- `ParentCardComponent`: collapsed/expanded states, flag badges, validation, edit/remove actions.
- Component-level: validation logic, save logic, draft auto-save.

### Integration Tests

- Multi-parent linking end-to-end: create parents, link to child, save step.
- Validation enforcement: missing parent, missing emergency contact, missing password.
- Transaction rollback: simulate failure, verify rollback.
- Draft auto-save: create parents, navigate away, restore draft.

### Manual Testing

- Create child with 2 parents, verify both linked.
- Edit child, add 3rd parent, verify all 3 shown.
- Remove parent from existing child, verify confirmation dialog.
- Create parent with "Other" relationship, verify free text works.
- Test "Same as child's home address" checkbox.
- Test portal status badge display.
- Test mobile layout on phone viewport.
- Test keyboard navigation in combobox.
- Test draft auto-save/restore with parent changes.

---

## Sources & Research

- Existing parent entity implementation: `docs/plans/2026-07-23-001-feat-parent-entity-and-management-plan.md`
- Current stepper component: `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`
- Current stepper template: `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.html`
- Parent models: `web/src/app/features/staff/models/parents.models.ts`
- Child contact models: `web/src/app/features/staff/models/child-profile.models.ts`
- Parent API service: `web/src/app/features/staff/data/parents-api.service.ts`
- Design system: `DESIGN.md`
- Architecture patterns: `docs/agents/ARCHITECTURE.md`
