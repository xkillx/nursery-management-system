---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
created: 2026-07-17
---

# fix: Repair 3 broken API unit tests

## Problem Frame

Three API tests are failing after production code changed without updating the corresponding tests:

- `siteprofile/application`: 2 tests fail because commit `355193d` made description a required field but didn't update the tests
- `term/application`: 1 test fails because it uses a hardcoded "future" date (2026-07-01) that is now in the past (today is 2026-07-17)

## Requirements

- All `go test ./...` passes with zero failures
- Tests accurately reflect current validation and domain logic
- No production code changes required — only test files need updating

## Key Technical Decisions

**KTD 1: Fix tests, not production code.** The siteprofile validation making description required is intentional (`355193d` commit message: "make description required with user-friendly error messages"). The term `initialStatus()` function correctly derives status from dates. Both are correct; the tests are stale.

**KTD 2: Use a safely-future date for term test.** Replace `2026-07-01` with `2027-07-01` to avoid future date-rot. The domain's own tests (`term_test.go:90`) already use `2027-01-01` as a future date.

## Implementation Units

### U1. Fix siteprofile AllEmptyFields test

**Goal:** Update `TestUpdateSiteProfile_AllEmptyFields` to expect 8 field errors (description is now required).

**Requirements:** R1 — all tests pass

**Files:**
- `api/internal/modules/siteprofile/application/application_test.go`

**Approach:**
- Change line 148 assertion from `len(fields) != 7` to `len(fields) != 8`
- Update comment on line 148 from "all required except description" to "all required fields"

**Test scenarios:**
- N/A — this unit IS a test fix

**Verification:** `go test -run TestUpdateSiteProfile_AllEmptyFields ./internal/modules/siteprofile/application/` passes

### U2. Fix siteprofile NameMaxLength test

**Goal:** Update `TestUpdateSiteProfile_NameMaxLength` to include a non-empty description so it passes validation.

**Requirements:** R1 — all tests pass

**Files:**
- `api/internal/modules/siteprofile/application/application_test.go`

**Approach:**
- Add `Description: "A valid description"` to the `UpdateSiteProfileInput` struct literal in the test (around line 204)

**Test scenarios:**
- N/A — this unit IS a test fix

**Verification:** `go test -run TestUpdateSiteProfile_NameMaxLength ./internal/modules/siteprofile/application/` passes

### U3. Fix term NoManagerRequired test date

**Goal:** Update `TestCreateTermDomainValidation_NoManagerRequired` to use a future start date so `initialStatus()` returns `pre_term`.

**Requirements:** R1 — all tests pass

**Files:**
- `api/internal/modules/term/application/create_term_test.go`

**Approach:**
- Change `time.Date(2026, 7, 1, ...)` to `time.Date(2027, 7, 1, ...)` on line 188

**Test scenarios:**
- N/A — this unit IS a test fix

**Verification:** `go test -run TestCreateTermDomainValidation_NoManagerRequired ./internal/modules/term/application/` passes

## Verification Contract

Run from `api/`:
```
go test ./internal/modules/siteprofile/application/ ./internal/modules/term/application/
```
All 3 previously-failing tests must pass. Then confirm full suite:
```
go test ./...
```
Zero failures expected.

## Definition of Done

- [ ] `TestUpdateSiteProfile_AllEmptyFields` passes
- [ ] `TestUpdateSiteProfile_NameMaxLength` passes
- [ ] `TestCreateTermDomainValidation_NoManagerRequired` passes
- [ ] `go test ./...` reports zero failures
