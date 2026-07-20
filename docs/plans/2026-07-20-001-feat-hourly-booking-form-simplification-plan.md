---
title: Hourly Booking Form Simplification - Plan
type: feat
date: 2026-07-20
topic: hourly-booking-form-simplification
artifact_contract: ce-unified-plan/v1
artifact_readiness: requirements-only
product_contract_source: ce-brainstorm
execution: code
---

## Goal Capsule

- **Objective:** Simplify the hourly booking creation form by removing unused fields and replacing the duration input with an end time picker.
- **Product authority:** User request — direct simplification of an existing workflow.
- **Open blockers:** None.

## Product Contract

### Summary

Remove the session type dropdown and room capacity sidebar from the hourly booking form. Replace the duration (minutes) number input with an end time picker so managers enter Start Time and End Time directly.

### Requirements

- R1. Remove the session type `<select>` dropdown from the booking form. The API field `session_type_id` becomes optional/omitted.
- R2. Replace the duration `<input type="number">` with an `<input type="time">` end time field. The form sends `start_time_minutes` and `duration_minutes` to the API — the component computes `duration_minutes = end_time - start_time`.
- R3. Remove the "Room Capacity" sidebar card entirely (display-only, not a form input).
- R4. Validation requires end time to be after start time. Display a field-level error when end time is at or before start time.
- R5. The financial impact card updates to use the computed duration from start/end times.

### Scope Boundaries

- No changes to the API, database schema, or backend modules — `start_time_minutes` and `duration_minutes` remain the API contract.
- No changes to the booking list, detail, or edit views — only the creation form.
- The "Child's Recent Bookings" sidebar card remains.
