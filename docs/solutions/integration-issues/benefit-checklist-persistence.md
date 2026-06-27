---
title: Persist Benefit Checklist Fields as Dedicated DB Columns
date: 2026-06-27
category: integration-issues
module: children
problem_type: integration_issue
component: database
severity: medium
symptoms:
  - Benefit checkboxes always appear unchecked in edit mode
  - other_benefit_name field always appears empty on edit load
root_cause: missing_workflow_step
resolution_type: migration
tags:
  - benefit-checklist
  - child-funding
  - edit-mode
  - registration
related_components:
  - api
  - frontend
---

# Persist Benefit Checklist Fields as Dedicated DB Columns

## Problem

Benefit checklist checkboxes selected during child registration were not persisted to the database. Individual benefit types (Universal Credit, Income Support, etc.) and the "other benefit name" text field collected in the UI were never stored as structured data, causing them to always appear unchecked on edit-mode load.

## Symptoms

- After saving a child registration with specific benefits selected (e.g., Universal Credit + Child Tax Credit), navigating back to the Funding & Benefits step showed all checkboxes unchecked.
- The `benefit_notes` free-text field contained a concatenated string like `"Benefits: universal_credit, child_tax_credit | Other: Housing Benefit"` but this string was never parsed back into individual checkbox selections on load.
- `benefits_status` was set to `"yes"` but the array of selected benefits was always empty in the API response, making it impossible for the frontend to reconstruct which benefits were chosen.

## What Didn't Work

- **Storing benefits as a concatenated string in `benefit_notes`** — The frontend's `buildFundingPayload` method text-concatenated the checkbox selections into `benefit_notes`:

  ```typescript
  // BEFORE — text concatenation workaround
  const benefitParts = [
    this.step6.benefit_notes,
    this.step6.benefits.length ? `Benefits: ${this.step6.benefits.join(', ')}` : '',
    this.step6.other_benefit_name ? `Other: ${this.step6.other_benefit_name}` : '',
  ].filter(Boolean).join(' | ');
  return {
    benefit_notes: benefitParts || null,
  };
  ```

  The backend `ChildFundingRecord` domain model had no boolean fields for individual benefit types. The `benefits` array and `other_benefit_name` fields existed in the frontend draft state but were never mapped to or from the database.

- **Receiving `benefits: []` on edit load** — `populateDraftsFromView` set `benefits: []` and `other_benefit_name: ''` unconditionally because the API response always returned empty values. The concatenated `benefit_notes` string was loaded but no parsing logic existed to reconstruct individual selections.

## Solution

### 1. Database migration — add typed boolean columns

`api/db/migrations/000002_add_benefit_checklist.up.sql`:

```sql
ALTER TABLE child_funding_records
  ADD COLUMN benefit_universal_credit      boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_income_support        boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_jobseekers_allowance  boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_esa_income_related    boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_child_tax_credit      boolean NOT NULL DEFAULT false,
  ADD COLUMN benefit_other_support         boolean NOT NULL DEFAULT false,
  ADD COLUMN other_benefit_name             text;
```

### 2. Domain model — add boolean fields

`api/internal/modules/children/domain/child_funding_record.go`:

```go
BenefitUniversalCredit     bool
BenefitIncomeSupport       bool
BenefitJobseekersAllowance bool
BenefitESAIncomeRelated    bool
BenefitChildTaxCredit      bool
BenefitOtherSupport        bool
OtherBenefitName           *string
```

### 3. SQL queries — reference new columns

`api/db/query/child_funding_records.sql`: Updated SELECT, INSERT, ON CONFLICT UPDATE, and RETURNING clauses to include all 7 new columns.

### 4. Application input type — add structured fields

`api/internal/modules/children/application/create_child_with_full_profile.go`:

```go
type ChildFundingRecordInput struct {
    // ...
    Benefits         []string
    OtherBenefitName *string
}
```

### 5. Mapper — string array to boolean fields

`api/internal/modules/children/application/helpers.go`:

```go
for _, b := range in.Benefits {
    switch b {
    case "universal_credit":       r.BenefitUniversalCredit = true
    case "income_support":         r.BenefitIncomeSupport = true
    case "jobseekers_allowance":   r.BenefitJobseekersAllowance = true
    case "esa_income_related":     r.BenefitESAIncomeRelated = true
    case "child_tax_credit":       r.BenefitChildTaxCredit = true
    case "other_support":          r.BenefitOtherSupport = true
    }
}
```

Validation added at the same location rejects `benefits_status=yes` with empty `benefits`.

### 6. HTTP DTOs — benefits array in requests and responses

`api/internal/modules/children/interfaces/http/dto_helpers.go`: Request and response structs include `Benefits []string` and `OtherBenefitName *string`. The response builder (`toChildFundingResponse`) iterates the 6 domain booleans and emits strings for true values.

### 7. Repository — pass new columns to sqlc params

`api/internal/modules/children/infrastructure/postgres/repository.go`: `UpsertFunding` passes benefit booleans to `ChildFundingRecordUpsertParams`. Both `mapFundingGetRow` and `mapFundingUpsertRow` map the new columns.

### 8. Frontend TypeScript models — update interfaces

`web/src/app/features/staff/models/child-profile.models.ts`:

```typescript
export interface ChildFundingRecord {
  benefits: string[];
  other_benefit_name: string | null;
}
export interface ChildFundingRecordInput {
  benefits: string[];
  other_benefit_name: string | null;
}
```

### 9. Frontend stepper — restore on load, send on save

`web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts`:

- **Populate from API**: `benefits: f.benefits ?? []`, `other_benefit_name: f.other_benefit_name ?? ''`
- **Save**: sends `benefits` and `other_benefit_name` directly in the payload
- **Registration payload**: no longer text-concatenates into `benefit_notes`

### 10. sqlc regeneration

```bash
make sqlc-generate
```

## Why This Works

The root cause was a data model mismatch: the UI collected structured data (an array of benefit type strings plus a separate "other benefit name" field) but the database schema had no columns for individual benefit types. The frontend worked around this by serializing the array into a free-text string in `benefit_notes`, but no reverse-path parsing ever existed — on edit-mode load, `benefit_notes` was loaded as a plain string and ignored.

The fix eliminates the serialization hack by adding 7 dedicated database columns. The full stack now stores and retrieves benefit selections as structured data: database → sqlc → repository → domain → application mapper → HTTP DTO → frontend model.

## Prevention

- **Schema-first design for structured data**: When the UI collects a set of checkboxes, the database schema should have corresponding boolean columns — never serialize structured data into a free-text field with intent to parse later.
- **Add validation at the schema boundary**: The input validation now rejects `benefits_status=yes` when no benefits are selected, preventing inconsistent states at the API level.
- **Review frontend-backend contract gaps**: A TODO comment in `buildFundingPayload` reading "Backend wiring — new fields may be ignored until the API type is updated" was the signal that frontend and backend were out of sync. Such markers should be blocking issues during code review.

## Related Issues

- Commit `eac2534` — "fix(children): persist benefit checklist fields as dedicated DB columns"
- `api/db/migrations/000002_add_benefit_checklist.up.sql` — migration adding boolean columns
