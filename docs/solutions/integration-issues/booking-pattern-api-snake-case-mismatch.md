---
title: Booking Pattern API Snake Case Mismatch
date: 2026-06-27
category: integration-issues
module: children
problem_type: integration_issue
component: service_object
symptoms:
  - Booking pattern API calls returned 400 Bad Request with "Invalid request payload."
  - Booking pattern data never loaded in edit mode (pattern always null)
  - effective_from date in past caused validation errors on save
root_cause: wrong_api
resolution_type: code_fix
severity: high
related_components:
  - frontend_stimulus
tags:
  - api-contract
  - snake-case
  - booking-pattern
  - serialization
  - angular
  - go
---

# Booking Pattern API Snake Case Mismatch

## Problem

The booking pattern API integration had a snake_case/camelCase mismatch between the Go backend and Angular frontend. Every booking pattern API call — both creating (POST) and editing (PATCH) — returned a generic `400 Bad Request` with no useful error, and loaded pattern data was always `undefined` at runtime.

## Symptoms

- **HTTP 400 on every booking pattern save** — both new patterns (POST) and edits (PATCH) returned `{"code":"validation_error","message":"Invalid request payload."}` with no field-level detail.
- **Booking pattern properties always `undefined`** — template bindings like `pattern.isCurrent` and `pattern.effectiveFrom` rendered blank. The API response contained `is_current` and `effective_from` (snake_case), but the TypeScript `BookingPattern` interface declared camelCase properties.
- **Past-date patterns rejected with confusing errors** — when a child had an existing pattern with `effective_from` in the past, editing failed with `booking_pattern_not_editable`, and creating a new pattern with a past date failed with `booking_pattern_backdated`. The form defaulted to the old pattern's past date with no guidance.
- **`benefit_notes` never rendered** — the field was loaded from the funding API and included in the save payload, but the template had no textarea for it.

## What Didn't Work

- **Focusing on backend validation** — the 400 appeared to be a validation failure, so initial debugging targeted Gin's `binding` tags and error messages. Adding `binding:"required"` made the error more frequent because the mismatched fields were always empty.
- **Fixing only the response interface** — converting only the `BookingPattern` response interface to snake_case fixed the display of loaded patterns but left saves broken, which was confusing because the form appeared correct while submission still failed.
- **Global JSON transform** — adding an Angular `HttpInterceptor` to convert all request bodies to snake_case was considered but rejected. It would be a global change affecting all API calls, potentially breaking endpoints that already used consistent casing.

## Solution

### 1. Align Angular model interfaces to snake_case

Changed `BookingPatternInput`, `BookingPattern`, `BookedSession`, and `SessionTypeRef` interfaces in `web/src/app/features/staff/models/booking-pattern.models.ts` to use snake_case, matching the Go DTO's `json` tags:

```typescript
export interface BookingPatternInput {
  effective_from: string;
  entries: BookingPatternEntryInput[];
}

export interface BookingPatternEntryInput {
  day_of_week: number;
  session_type_id: string;
}

export interface BookingPattern {
  id: string;
  child_id: string;
  effective_from: string;
  effective_to: string | null;
  is_current: boolean;
  created_at: string;
  entries: BookedSession[];
}
```

Angular's `HttpClient` serializes JavaScript property names as-is into JSON keys. By matching the Go DTO's `json:"snake_case"` tags, `ShouldBindJSON` now correctly populates every field.

### 2. Default `patternEffectiveFrom` to today when pattern is in the past

When `listChildBookingPatterns` returns a pattern whose `effective_from` is before today, `loadBookingPattern` now defaults `patternEffectiveFrom` to `this.todayIso` instead of using the API value:

```typescript
if (current.effective_from >= this.todayIso) {
  this.patternEffectiveFrom = current.effective_from;
} else {
  this.patternEffectiveFrom = this.todayIso;
}
```

This ensures the form always presents a valid future date. The user can still change it before saving.

### 3. Added `benefit_notes` textarea

Added a `benefit_notes` textarea to the Benefits section of the funding-benefits step template. The field was already loaded from the API response and included in the save payload — it was simply never rendered.

### 4. Updated all consumers

The same snake_case fix was applied to:
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.ts` — payload construction, response property access, date comparison
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.html` — benefit notes textarea, inline field errors
- `web/src/app/features/staff/pages/manager-booking-pattern/manager-booking-pattern.component.ts` — standalone booking pattern component
- `web/src/app/features/staff/pages/manager-child-edit/manager-child-edit-stepper.component.spec.ts` — test mock data

### 5. Added client-side date validation

Added a check in `saveSessionPattern` that rejects `effective_from` dates before today with a clear inline error message, preventing the round-trip to the backend before the user sees the `booking_pattern_backdated` rejection.

## Why This Works

- **Angular serialization is transparent** — `HttpClient` does not transform key names. The JSON body sent over the wire exactly matches the property names of the TypeScript objects. By aligning those names to the Go DTO's `json` tags, `ShouldBindJSON` now correctly maps every field.
- **`binding:"required"` now has data to check** — the Go DTO's `effective_from` field was never empty after the fix, so the required validation passes.
- **Past-date defaulting avoids a domain rejection** — by defaulting to today's date when the API returns a past effective_from, the form always presents a date that satisfies backend domain rules for new patterns.
- **Rendering is an explicit step** — `benefit_notes` existed in the data flow (API → component → save) but the template was simply missing the textarea.

## Prevention

1. **API contract-first approach** — define the JSON contract (field names, types) in a shared document or OpenAPI spec before writing the Go DTO and Angular interface. Verify casing consistency during PR review.
2. **Existing precedent** — `d471e2a` fixed the same class of bug for session type API fields. When a second instance appears, check all API model files for the same pattern rather than fixing one endpoint at a time.
3. **Integration smoke test** — add a single E2E test that creates a booking pattern via the API and verifies the response has expected keys. This would catch serialization mismatches at CI time.
4. **Angular serializer behavior** — document that `HttpClient` sends JS object keys as-is. Developers coming from frameworks with automatic case conversion (Jackson, ASP.NET conventions) may assume transformation happens.

## Related Issues

- Commit `d471e2a` — earlier instance of the same bug class (session type API fields).
- Commit `c53b67e` — the main snake_case fix for booking pattern models.
- `docs/adr/0010-atomic-child-with-pattern-wizard-submit.md` — booking pattern embedded in child creation payload (same contract applies).
- `docs/adr/0011-registration-funding-guidance-only.md` — funding record data model that includes `benefit_notes`.
