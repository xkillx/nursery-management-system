# API Contract — MVP (Week 1)

Frontend integration contract for implemented API routes. All routes are live unless marked **Deferred / not implemented**.

## Common Conventions

**Base path:** `/api/v1`

**Authentication:** Protected routes require `Authorization: Bearer <access_token>`.

**Date formats:**
- Dates: `YYYY-MM-DD`
- Timestamps: RFC 3339 (`2026-05-26T10:00:00Z`)

**Roles:** `manager`, `practitioner`, `parent`

**Success responses:** Plain JSON resources, no global envelope.

**Error response shape:**

```json
{
  "code": "error_code",
  "message": "Human-readable message",
  "request_id": "uuid",
  "details": { "field": "field_name" }
}
```

The `details` object is optional and only present when a specific field is identified.

**Cookie / CSRF mechanics:**

- `refresh_token` is set as an `HttpOnly` cookie.
- `csrf_token` is set as a readable cookie.
- Session actions (`POST /auth/refresh`, `POST /auth/logout`, `POST /auth/switch-membership`) require:
  - `X-CSRF-Token` header matching the `csrf_token` cookie value.
  - Trusted `Origin` or `Referer` header matching the request host.
  - `credentials: "include"` (browser clients) so cookies are sent.

**Request correlation headers** (all responses):

| Header | Direction | Description |
|--------|-----------|-------------|
| `X-Request-ID` | Request/Response | Accepted on request (preserved when safe, otherwise generated). Always present on response. |
| `X-Correlation-ID` | Request/Response | Accepted on request (preserved when safe, otherwise generated). Always present on response. |
| `traceparent` | Request only | W3C traceparent parsed for trace ID; only the trace ID appears in logs, never returned. |

**Lifecycle reason codes** (used by mark-inactive, deactivate, end link, end mapping):

| Code | Label |
|------|-------|
| `duplicate_record` | Duplicate record |
| `entered_in_error` | Entered in error |
| `left_nursery` | Left nursery |
| `safeguarding_direction` | Safeguarding direction |
| `contact_update` | Contact update |
| `access_revoked` | Access revoked |
| `other` | Other (requires `reason_note`) |

---

## Auth / Session

### POST /api/v1/auth/login

Public.

**Request:**

```json
{
  "email": "user@example.test",
  "password": "Pass1234",
  "membership_id": "uuid"
}
```

`membership_id` is optional. Required when user has multiple memberships.

**Response 200:**

```json
{
  "access_token": "jwt",
  "token_type": "Bearer",
  "expires_in_seconds": 900,
  "user": { "id": "uuid", "email": "user@example.test" },
  "active_membership": {
    "membership_id": "uuid",
    "tenant_id": "uuid",
    "branch_id": "uuid",
    "role": "manager"
  },
  "available_memberships": [
    { "membership_id": "uuid", "tenant_id": "uuid", "branch_id": "uuid", "role": "manager" }
  ]
}
```

Sets `refresh_token` (HttpOnly) and `csrf_token` cookies.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Malformed payload or multi-membership login without `membership_id` |
| 401 | `unauthorized` | Invalid credentials or session |
| 403 | `forbidden_scope_selection` | Selected membership does not belong to user |

### POST /api/v1/auth/refresh

Public. Cookie-backed, CSRF-protected. No request body.

**Response 200:** Same shape as login response.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 401 | `unauthorized` | Invalid or missing refresh token |
| 403 | `forbidden_scope_selection` | Invalid CSRF token or origin |

### POST /api/v1/auth/logout

Public. Cookie-backed. CSRF-protected when refresh cookie is present.

**Response 204:** Clears `refresh_token` and `csrf_token` cookies. Idempotent.

### POST /api/v1/auth/switch-membership

Public. Cookie-backed, CSRF-protected.

**Request:**

```json
{ "membership_id": "uuid" }
```

**Response 200:** Same shape as login response.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Missing or invalid `membership_id` |
| 401 | `unauthorized` | Invalid refresh token |
| 403 | `forbidden_scope_selection` | Target membership not available to user |

### GET /api/v1/me

Roles: manager, practitioner, parent.

**Response 200:**

```json
{ "auth": { "user_id": "uuid", "tenant_id": "uuid", "branch_id": "uuid", "role": "manager", "membership_id": "uuid" } }
```

Note: current shape returns the raw authorization context for debugging. See Known Contract Gaps.

---

## Password Reset

### POST /api/v1/auth/password-reset-requests

Public.

**Request:**

```json
{ "email": "user@example.test" }
```

**Response 202:**

```json
{ "status": "accepted" }
```

Same response for known and unknown emails.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload |
| 429 | `rate_limited` | Too many requests |

### POST /api/v1/auth/password-resets

Public.

**Request:**

```json
{ "token": "reset-token", "new_password": "NewPass123" }
```

Password minimum 8 characters.

**Response 204:** Empty.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload or password too short |
| 400 | `password_reset_token_invalid` | Token does not exist |
| 400 | `password_reset_token_expired` | Token has expired |
| 400 | `password_reset_token_used` | Token has already been used |

---

## Invites

### POST /api/v1/invites

Roles: manager.

**Request:**

```json
{ "email": "new@example.test", "role": "practitioner" }
```

`role` accepts `practitioner` or `parent`.

**Response 201** (new invite) or **200** (existing pending invite refreshed):

```json
{
  "id": "uuid",
  "email": "new@example.test",
  "role": "practitioner",
  "status": "pending",
  "expires_at": "2026-06-02T10:00:00Z",
  "accepted_at": null,
  "revoked_at": null,
  "created_at": "2026-05-26T10:00:00Z",
  "updated_at": "2026-05-26T10:00:00Z"
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload |
| 400 | `invite_role_not_allowed` | Role `manager` is not invitable |
| 409 | `invite_email_already_registered` | Email already has a user account |
| 409 | `invite_scope_conflict` | Pending invite exists for same email with different role |

### GET /api/v1/invites

Roles: manager.

**Query params:** `status=pending|accepted|revoked|expired|all` (defaults to `pending`)

**Response 200:**

```json
{ "items": [ { ...inviteResponse } ] }
```

### POST /api/v1/invites/:invite_id/resend

Roles: manager.

**Response 200:** Updated invite response.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid `invite_id` |
| 404 | `invite_not_found` | Invite does not exist |
| 409 | `invite_not_pending` | Invite is no longer pending |

### POST /api/v1/invites/:invite_id/revoke

Roles: manager.

**Response 200:** Updated invite response.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid `invite_id` |
| 404 | `invite_not_found` | Invite does not exist |
| 409 | `invite_not_pending` | Invite is no longer pending |
| 409 | `invite_already_accepted` | Invite was already accepted |

### POST /api/v1/invites/accept

Public. Rate limited.

**Request:**

```json
{ "token": "invite-token", "new_password": "NewPass123" }
```

Password minimum 8 characters. Does not start a session.

**Response 204:** Empty.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload or password too short |
| 400 | `invite_token_invalid` | Token does not match any invite |
| 400 | `invite_token_expired` | Invite has expired |
| 400 | `invite_token_revoked` | Invite has been revoked |
| 400 | `invite_token_accepted` | Invite has already been accepted |
| 429 | `rate_limited` | Too many requests from this IP |

---

## Children

### GET /api/v1/children

Roles: manager.

**Query params:**
- `status=active|inactive|all` (defaults to `active`)
- `limit=1-200` (defaults to `50`)
- `offset=0+` (defaults to `0`)

**Response 200:**

```json
{
  "items": [
    {
      "id": "uuid",
      "full_name": "Alex Child",
      "date_of_birth": "2022-03-15",
      "start_date": "2025-01-06",
      "end_date": null,
      "core_hourly_rate_minor": 500,
      "notes": null,
      "is_active": true,
      "left_at": null,
      "left_reason_code": null,
      "left_reason_note": null,
      "enrollment_complete": true,
      "missing_requirements": [],
      "created_at": "2026-05-26T10:00:00Z",
      "updated_at": "2026-05-26T10:00:00Z"
    }
  ]
}
```

### GET /api/v1/children/:child_id

Roles: manager.

**Response 200:** Single child object (no `items` wrapper).

**Errors:**

| Status | Code | When |
|--------|------|------|
| 404 | `child_not_found` | Child does not exist |

### POST /api/v1/children

Roles: manager.

**Request:**

```json
{
  "full_name": "Alex Child",
  "date_of_birth": "2022-03-15",
  "start_date": "2025-01-06",
  "end_date": "2026-07-31",
  "core_hourly_rate_minor": 500,
  "notes": "Allergies: none"
}
```

`end_date` and `notes` are optional.

**Response 201:** Single child object.

### PATCH /api/v1/children/:child_id

Roles: manager.

**Request:** Partial update. Same fields as create. Empty/null fields are ignored (see Known Contract Gaps).

**Response 200:** Updated child object.

### POST /api/v1/children/:child_id/actions/mark-inactive

Roles: manager.

**Request:**

```json
{ "reason_code": "left_nursery", "reason_note": "Family relocating" }
```

`reason_note` is required when `reason_code` is `other`.

**Response 200:** Updated child object.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 404 | `child_not_found` | Child does not exist |
| 400 | `child_lifecycle_reason_required` | Missing `reason_code` |
| 400 | `lifecycle_reason_invalid` | Unknown reason code |
| 400 | `reason_note_required_for_other` | `reason_code` is `other` without note |

### GET /api/v1/children/attendance

Roles: manager, practitioner.

Returns children with current attendance state for the current `Europe/London` local day.

**Response 200:**

```json
{
  "items": [
    {
      "id": "uuid",
      "full_name": "Alex Child",
      "enrollment_complete": true,
      "attendance_state": "checked_in",
      "open_session_id": "uuid",
      "checked_in_at": "2026-05-26T08:00:00Z",
      "has_incomplete_session": true,
      "absence_marker_id": null,
      "absence_marked_at": null
    }
  ]
}
```

`attendance_state` values: `not_checked_in`, `checked_in`, `absent`.

`absence_marker_id` and `absence_marked_at` are nullable. When present, the child has an active (non-cleared) absence marker for the current `Europe/London` local day and `attendance_state` is `absent`.

---

## Guardians

### GET /api/v1/guardians

Roles: manager.

**Query params:**
- `status=active|inactive|all` (defaults to `active`)
- `limit=1-200` (defaults to `50`)
- `offset=0+` (defaults to `0`)

**Response 200:**

```json
{
  "items": [
    {
      "id": "uuid",
      "full_name": "Avery Parent",
      "email": "avery@example.test",
      "phone": "+441234567890",
      "notes": null,
      "is_active": true,
      "deactivated_at": null,
      "deactivation_reason_code": null,
      "deactivation_reason_note": null,
      "created_at": "2026-05-26T10:00:00Z",
      "updated_at": "2026-05-26T10:00:00Z"
    }
  ]
}
```

### GET /api/v1/guardians/:guardian_id

Roles: manager.

**Response 200:** Single guardian object.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 404 | `guardian_not_found` | Guardian does not exist |

### POST /api/v1/guardians

Roles: manager.

**Request:**

```json
{
  "full_name": "Avery Parent",
  "email": "avery@example.test",
  "phone": "+441234567890",
  "notes": "Key contact"
}
```

`email`, `phone`, `notes` are optional.

**Response 201:** Single guardian object.

### PATCH /api/v1/guardians/:guardian_id

Roles: manager.

**Request:** Partial update. Same fields as create. Empty/null fields are ignored (see Known Contract Gaps).

**Response 200:** Updated guardian object.

### POST /api/v1/guardians/:guardian_id/actions/deactivate

Roles: manager.

**Request:**

```json
{ "reason_code": "access_revoked", "reason_note": "Per safeguarding direction" }
```

`reason_note` is required when `reason_code` is `other`.

**Response 200:** Updated guardian object.

### POST /api/v1/guardians/:guardian_id/actions/reactivate

Roles: manager. No request body.

**Response 200:** Updated guardian object.

---

## Guardian-Child Links

### POST /api/v1/guardian-child-links

Roles: manager.

**Request:**

```json
{ "guardian_id": "uuid", "child_id": "uuid" }
```

Idempotent: creating an already-active pair returns the existing link with **200**.

**Response 201** (new) or **200** (existing):

```json
{
  "id": "uuid",
  "guardian_id": "uuid",
  "child_id": "uuid",
  "ended_at": null,
  "ended_reason_code": null,
  "ended_reason_note": null,
  "created_at": "2026-05-26T10:00:00Z",
  "updated_at": "2026-05-26T10:00:00Z"
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid UUIDs |
| 404 | `guardian_not_found` | Guardian does not exist |
| 400 | `guardian_not_active` | Guardian is deactivated |
| 404 | `child_not_found` | Child does not exist |

### POST /api/v1/guardian-child-links/:link_id/actions/end

Roles: manager.

**Request:**

```json
{ "reason_code": "left_nursery", "reason_note": null }
```

**Response 200:** Updated link object.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 404 | `guardian_child_link_not_found` | Link does not exist |
| 400 | `relationship_reason_required` | Missing `reason_code` |
| 400 | `lifecycle_reason_invalid` | Unknown reason code |
| 400 | `reason_note_required_for_other` | `reason_code` is `other` without note |

---

## Parent Membership-Guardian Mappings

### POST /api/v1/parent-membership-guardian-mappings

Roles: manager.

**Request:**

```json
{ "membership_id": "uuid", "guardian_id": "uuid" }
```

Idempotent: creating the same active membership/guardian pair returns existing mapping with **200**.

**Response 201** (new) or **200** (existing):

```json
{
  "id": "uuid",
  "membership_id": "uuid",
  "guardian_id": "uuid",
  "ended_at": null,
  "ended_reason_code": null,
  "ended_reason_note": null,
  "created_at": "2026-05-26T10:00:00Z",
  "updated_at": "2026-05-26T10:00:00Z"
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid UUIDs |
| 404 | `membership_not_found` | Membership does not exist |
| 400 | `membership_not_parent` | Membership role is not `parent` |
| 400 | `membership_not_active` | Membership is not active |
| 404 | `guardian_not_found` | Guardian does not exist |
| 400 | `guardian_not_active` | Guardian is deactivated |
| 409 | `parent_mapping_active_conflict` | Membership already mapped to a different active guardian |

### POST /api/v1/parent-membership-guardian-mappings/:mapping_id/actions/end

Roles: manager.

**Request:**

```json
{ "reason_code": "access_revoked", "reason_note": null }
```

**Response 200:** Updated mapping object.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 404 | `parent_mapping_not_found` | Mapping does not exist |
| 400 | `relationship_reason_required` | Missing `reason_code` |
| 400 | `lifecycle_reason_invalid` | Unknown reason code |
| 400 | `reason_note_required_for_other` | `reason_code` is `other` without note |

---

## Attendance

### POST /api/v1/attendance/check-ins

Roles: manager, practitioner.

**Request:**

```json
{ "child_id": "uuid" }
```

**Response 201:**

```json
{
  "id": "uuid",
  "child_id": "uuid",
  "status": "open",
  "check_in_at": "2026-05-26T08:00:00Z",
  "check_out_at": null,
  "check_in_local_date": "2026-05-26",
  "check_out_local_date": null,
  "duration_minutes": null,
  "created_at": "2026-05-26T08:00:00Z",
  "updated_at": "2026-05-26T08:00:00Z"
}
```

`status` values: `open`, `complete`, `corrected`.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload |
| 404 | `child_not_found` | Child does not exist |
| 409 | `attendance_session_already_open` | Child already checked in |
| 409 | `absence_marker_exists` | Child has an active absence marker for today |
| 409 | `child_enrollment_incomplete` | Child enrollment is not complete |

### POST /api/v1/attendance/check-outs

Roles: manager, practitioner.

**Request:**

```json
{ "child_id": "uuid" }
```

**Response 200:** Updated session object with `check_out_at` and `duration_minutes` populated.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload |
| 409 | `attendance_session_not_open` | No open session for child |
| 409 | `attendance_invalid_time_order` | Check-out time is not after check-in |

### POST /api/v1/attendance/corrections

Roles: manager.

**Request — existing session:**

```json
{
  "session_id": "uuid",
  "check_in_at": "2026-05-26T08:00:00Z",
  "check_out_at": "2026-05-26T16:00:00Z",
  "reason_code": "incorrect_time",
  "reason_note": null
}
```

**Request — missed session:**

```json
{
  "child_id": "uuid",
  "check_in_at": "2026-05-26T08:00:00Z",
  "check_out_at": "2026-05-26T16:00:00Z",
  "reason_code": "missed_check_in",
  "reason_note": null
}
```

Exactly one of `session_id` or `child_id` is required. Supplying both or neither is a validation error.

Correction reason codes: `missed_check_in`, `missed_check_out`, `incorrect_time`, `duplicate_entry`, `other`.

`reason_note` is required when `reason_code` is `other`.

The correction event's `occurred_at` and `local_date` reflect the manager action instant/day. The corrected interval dates (`check_in_local_date`, `check_out_local_date`) are represented by the returned session and event details.

**Response 200** (corrected existing session) or **201** (created missed session).

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload, timestamps, or UUIDs |
| 400 | `attendance_correction_reason_required` | Missing `reason_code` |
| 400 | `attendance_correction_reason_invalid` | Unknown reason code |
| 400 | `reason_note_required_for_other` | `reason_code` is `other` without note |
| 409 | `attendance_correction_future_time` | Corrected times are in the future |
| 409 | `attendance_session_overlap` | Corrected interval overlaps another session |
| 409 | `attendance_outside_enrollment_window` | Corrected dates are outside child start/end dates |

### POST /api/v1/attendance/absence-markers

Roles: manager, practitioner.

Mark a child as absent for the current `Europe/London` local day. Absence markers are non-billing and never change invoice calculations. Idempotent: if an active (non-cleared) marker already exists for the same child and local date, returns the existing marker with **200**.

**Request:**

```json
{ "child_id": "uuid" }
```

**Response 201** (new marker):

```json
{
  "id": "uuid",
  "child_id": "uuid",
  "local_date": "2026-05-26",
  "marked_at": "2026-05-26T08:30:00Z",
  "cleared_at": null,
  "created_at": "2026-05-26T08:30:00Z",
  "updated_at": "2026-05-26T08:30:00Z"
}
```

**Response 200** (existing active marker): same shape.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid payload or child_id |
| 404 | `child_not_found` | Child does not exist |
| 409 | `absence_attendance_exists` | Child already has an open attendance session for today |

### POST /api/v1/attendance/absence-markers/:absence_marker_id/clear

Roles: manager, practitioner.

Clear an active absence marker. Sets `cleared_at` to the current server time. The marker is no longer considered active.

**Request:** No body required.

**Response 200:**

```json
{
  "id": "uuid",
  "child_id": "uuid",
  "local_date": "2026-05-26",
  "marked_at": "2026-05-26T08:30:00Z",
  "cleared_at": "2026-05-26T09:00:00Z",
  "created_at": "2026-05-26T08:30:00Z",
  "updated_at": "2026-05-26T09:00:00Z"
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid `absence_marker_id` |
| 404 | `absence_marker_not_found` | Marker does not exist |

---

## Deferred / Not Implemented

These contracts are proposed for future implementation. Frontend should use typed mock adapters until backend support lands.

### GET /api/v1/children/:child_id/guardian-child-links

Roles: manager. Query: `status=active|ended|all`.

```json
{
  "items": [
    {
      "id": "link-uuid",
      "guardian_id": "guardian-uuid",
      "child_id": "child-uuid",
      "guardian": {
        "id": "guardian-uuid",
        "full_name": "Avery Parent",
        "email": "avery@example.test",
        "phone": "+441234567890",
        "is_active": true
      },
      "ended_at": null,
      "ended_reason_code": null,
      "ended_reason_note": null,
      "created_at": "2026-05-25T10:00:00Z",
      "updated_at": "2026-05-25T10:00:00Z"
    }
  ]
}
```

### GET /api/v1/guardians/:guardian_id/parent-membership-guardian-mappings

Roles: manager. Query: `status=active|ended|all`.

```json
{
  "items": [
    {
      "id": "mapping-uuid",
      "membership_id": "membership-uuid",
      "guardian_id": "guardian-uuid",
      "parent_user": {
        "id": "user-uuid",
        "email": "parent-login@example.test"
      },
      "ended_at": null,
      "ended_reason_code": null,
      "ended_reason_note": null,
      "created_at": "2026-05-25T10:00:00Z",
      "updated_at": "2026-05-25T10:00:00Z"
    }
  ]
}
```

### GET /api/v1/attendance/sessions/:session_id/events

Roles: manager. Query: `event_type=check_in|check_out|correction`.

```json
{
  "items": [
    {
      "id": "event-uuid",
      "session_id": "session-uuid",
      "child_id": "child-uuid",
      "event_type": "correction",
      "occurred_at": "2026-05-25T10:30:00Z",
      "local_date": "2026-05-25",
      "recorded_by_user_id": "user-uuid",
      "recorded_by_membership_id": "membership-uuid",
      "reason_code": "incorrect_time",
      "reason_note": null,
      "details": {
        "corrected_check_in": "2026-05-25T08:00:00Z",
        "corrected_check_out": "2026-05-25T16:00:00Z"
      }
    }
  ]
}
```

### POST /api/v1/invoices/:invoice_id/adjustments

**Deferred / not implemented.** This route must not be added during month 1 unless UAT or pilot operations produce a real post-issue correction that must be handled inside the product before the next invoice/payment cycle.

Creates a manager-only follow-up adjustment invoice linked to the original monthly invoice. The original invoice is never edited, regenerated, deleted, refunded, or directly offset by this endpoint.

**Role:** `manager` only.

**Original invoice requirements:**

- The `:invoice_id` invoice is in the manager's current tenant and branch scope.
- `invoice_kind = monthly`.
- `status` is one of `issued`, `payment_failed`, `overdue`, or `paid`.
- `draft` invoices are rejected because existing draft regeneration remains the correction path.
- Adjustment invoices are rejected as adjustment targets; adjustment chains are not supported.

**Request:**

```json
{
  "confirm": true,
  "reason_code": "attendance_correction",
  "reason_note": "Late checkout correction for 2026-05-15 after invoice issue.",
  "lines": [
    {
      "description": "Additional childcare after corrected checkout",
      "line_amount_minor": 750
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `confirm` | boolean | yes | Must be `true` to create and issue the adjustment invoice |
| `reason_code` | string | yes | One of `attendance_correction`, `funding_correction`, `extra_correction`, `billing_error`, `other` |
| `reason_note` | string | yes | Non-empty manager explanation for the adjustment |
| `lines` | array | yes | One or more adjustment lines |
| `lines[].description` | string | yes | Parent/manager readable line description |
| `lines[].line_amount_minor` | integer | yes | Signed line amount in GBP minor units |

The first deferred version allows signed adjustment lines, but the net adjustment invoice total must be non-negative to match the current schema. Negative net credits, refunds, credit carry-forward, partial offsets, account balances, and Stripe payment/refund behavior are separate post-MVP settlement design work.

**Creation behavior:**

- Creates a separate `invoice_kind = adjustment` invoice.
- Sets `adjusts_invoice_id` to the original monthly invoice.
- Persists both `adjustment_reason_code` and non-empty `adjustment_reason_note`.
- Creates the invoice already issued in one confirmed action.
- Assigns an invoice number using the adjustment invoice billing month.
- Sets `issued_at`, `locked_at`, and `due_at` to the creation instant.
- Records an `invoice_runs` row with `run_type = issue` and details including `mode: "adjustment_issue"`.
- Writes an audit event for the adjustment invoice creation/issue action.
- Inserts only `line_kind = adjustment` invoice lines for the adjustment invoice.

**Response 201:**

```json
{
  "invoice_id": "adjustment-invoice-uuid",
  "invoice_kind": "adjustment",
  "invoice_number": "INV-202605-0007",
  "status": "issued",
  "adjusts_invoice_id": "original-invoice-uuid",
  "adjustment_reason_code": "attendance_correction",
  "adjustment_reason_note": "Late checkout correction for 2026-05-15 after invoice issue.",
  "issued_at": "2026-06-01T10:00:00Z",
  "locked_at": "2026-06-01T10:00:00Z",
  "due_at": "2026-06-01T10:00:00Z",
  "issued_run_id": "uuid",
  "currency_code": "GBP",
  "total_due_minor": 750,
  "lines": [
    {
      "line_kind": "adjustment",
      "description": "Additional childcare after corrected checkout",
      "sort_order": 1,
      "line_amount_minor": 750
    }
  ]
}
```

**Payment and parent disclosure boundaries:**

- Parent checkout remains limited to `invoice_kind = monthly`; adjustment invoices are not payable through Stripe.
- Parent invoice list/detail remains monthly-invoice only unless a later parent disclosure design explicitly changes that boundary.
- Manager invoice review may show adjustment invoices and their link to the original invoice.

**Errors:**

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed invoice ID, missing `confirm`, empty lines, invalid amount, or invalid field shape |
| 400 | `adjustment_reason_required` | Missing `reason_code` or `reason_note` |
| 400 | `adjustment_reason_invalid` | Unknown `reason_code` |
| 400 | `adjustment_negative_total_not_supported` | Signed lines produce a negative net total |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `invoice_not_found` | Original invoice absent from tenant/branch scope |
| 409 | `invoice_not_monthly` | Original invoice is not a monthly invoice |
| 409 | `invoice_not_issued` | Original invoice is still draft |
| 409 | `adjustment_chain_not_supported` | Original invoice is itself an adjustment invoice |

---

## Invoice Draft Preflight

Manager-only, read-only endpoint that returns child-month readiness for draft invoice generation. No side effects: does not create invoice runs, invoices, invoice lines, or audit logs.

**Route is manager-only.** Unauthenticated requests receive `401 unauthorized`. Practitioner and parent requests receive `403 forbidden_role`.

### GET /api/v1/invoices/drafts/preflight?billing_month=YYYY-MM

Returns eligible children, blocked children with stable blocker codes, and summary totals for eligible children only.

**Request:**

| Parameter | Location | Required | Description |
|-----------|----------|----------|-------------|
| `billing_month` | query | yes | Billing month as `YYYY-MM` |

**Response 200** (always 200, even when children are blocked):

```json
{
  "billing_month": "2026-05",
  "currency_code": "GBP",
  "period": {
    "start_date": "2026-05-01",
    "end_date": "2026-05-31",
    "end_exclusive_date": "2026-06-01"
  },
  "summary": {
    "total_children_count": 3,
    "eligible_children_count": 1,
    "blocked_children_count": 2,
    "included_session_count": 2,
    "raw_attended_minutes": 615,
    "rounded_attended_minutes": 630,
    "funded_allowance_minutes": 300,
    "funded_deduction_minutes": 300,
    "core_billable_minutes": 330,
    "subtotal_minor": 5250,
    "funded_deduction_minor": 2500,
    "total_due_minor": 2750,
    "blocker_counts": [
      { "code": "missing_funding_profile", "children_count": 1 },
      { "code": "incomplete_attendance", "children_count": 1 }
    ]
  },
  "eligible_children": [
    {
      "child_id": "uuid",
      "child_name": "Alex Child",
      "core_hourly_rate_minor": 500,
      "funding_profile_id": "uuid",
      "funded_allowance_minutes": 300,
      "raw_attended_minutes": 615,
      "rounded_attended_minutes": 630,
      "included_session_count": 2,
      "funded_deduction_minutes": 300,
      "core_billable_minutes": 330,
      "subtotal_minor": 5250,
      "funded_deduction_minor": 2500,
      "total_due_minor": 2750,
      "existing_invoice": null
    }
  ],
  "blocked_children": [
    {
      "child_id": "uuid",
      "child_name": "Bailey Child",
      "blockers": [
        {
          "code": "missing_funding_profile",
          "message": "Funding profile is missing for this billing month.",
          "field": "funding_profile"
        }
      ]
    }
  ]
}
```

**Stable blocker codes** (listed in priority order):

| Code | Meaning |
|------|---------|
| `missing_child_name` | Child full name is blank |
| `missing_child_date_of_birth` | Child date of birth is zero |
| `missing_child_start_date` | Child start date is zero |
| `missing_guardian_link` | No active guardian-child link |
| `missing_billing_rate` | Core hourly rate is negative |
| `missing_funding_profile` | No funding profile for this child/month |
| `incomplete_attendance` | Attendance session missing check-out |
| `invoice_already_issued` | Monthly invoice in issued/paid/overdue status |

A child can have multiple blockers. Blocked children contribute to counts and blocker breakdowns only; they do not affect money or minute totals.

**Existing invoice handling:**
- Existing monthly draft: child is eligible, `existing_invoice` field populated.
- Existing monthly invoice in `issued`, `payment_failed`, `paid`, or `overdue`: blocked with `invoice_already_issued`.

**Summary money totals** include eligible children only. Money calculation uses deterministic integer ceiling: `amount_minor = ceil(minutes * hourly_rate_minor / 60)`.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Missing or malformed `billing_month` |
| 401 | `unauthorized` | No valid token |
| 403 | `forbidden_role` | Non-manager role |

---

## Draft Invoice Generation

Manager-only endpoint that generates or regenerates draft monthly invoices for eligible child-months. Creates invoice runs, invoices, invoice lines, and audit logs inside a single transaction.

**Route is manager-only.** Unauthenticated requests receive `401 unauthorized`. Practitioner and parent requests receive `403 forbidden_role`.

### POST /api/v1/invoice-runs/drafts

**Request body:**

```json
{
  "billing_month": "2026-05",
  "child_ids": ["uuid", "uuid"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `billing_month` | string | yes | Billing month as `YYYY-MM` |
| `child_ids` | string[] | no | Optional array of child UUIDs. Omit for full-month generation. Empty array = no-op. Duplicates are deduplicated. |

**Response 200:**

```json
{
  "run_id": "uuid",
  "billing_month": "2026-05",
  "status": "completed",
  "summary": {
    "eligible_count": 3,
    "success_count": 2,
    "blocked_count": 1,
    "total_due_minor": 3000
  },
  "generated": [
    {
      "child_id": "uuid",
      "child_name": "Alex Child",
      "action": "created",
      "invoice_id": "uuid",
      "subtotal_minor": 4000,
      "funded_deduction_minor": 2500,
      "total_due_minor": 1500
    }
  ],
  "blocked": [
    {
      "child_id": "uuid",
      "child_name": "Bailey Child",
      "blockers": [
        { "code": "missing_funding_profile", "message": "Funding profile is missing for this billing month." }
      ]
    }
  ]
}
```

**`action` values:** `created` (new draft), `updated` (regenerated existing draft).

**`status` values:** `completed` (all eligible children generated), `completed_with_exceptions` (one or more children blocked).

**Blocker codes** include the same codes as preflight, plus:

| Code | Meaning |
|------|---------|
| `child_not_found` | Selected child ID not found in tenant/branch scope |
| `child_not_in_billing_month` | Selected child not active during the billing month |

**Idempotency:** Regenerating a draft for the same child/month updates the existing draft invoice in place. The invoice ID stays stable. System-calculated lines (`core_childcare`, `funded_deduction`) are replaced. Manual `extra` lines are preserved.

**Transaction semantics:** The entire operation runs in one database transaction. Unexpected errors roll back all changes. No partial state is left behind.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Missing/malformed `billing_month`, malformed `child_ids` |
| 401 | `unauthorized` | No valid token |
| 403 | `forbidden_role` | Non-manager role |

---

## Manager Invoice Review

Manager-only endpoints for reviewing generated invoices across all statuses, including drafts. Parents see only issued-or-later invoices; see the Parent Invoice View section below.

**Routes are manager-only.** Practitioner and parent requests receive `403 forbidden_role`.

### `GET /api/v1/invoices`

List invoices for the active tenant/branch scope.

**Query parameters (all optional):**

| Param | Type | Default | Notes |
|-------|------|---------|-------|
| `billing_month` | `YYYY-MM` | — | Filter to one month |
| `status` | `draft`, `issued`, `payment_failed`, `paid`, `overdue` | — | Filter by invoice status |
| `child_id` | `uuid` | — | Filter to one child |
| `limit` | `int` | `50` | `1..200` |
| `offset` | `int` | `0` | `0+` |

**Ordering:** `billing_month DESC, child_name ASC, created_at DESC, id ASC`

**Response `200`:**

```json
{
  "items": [
    {
      "invoice_id": "uuid",
      "invoice_kind": "monthly",
      "invoice_number": null,
      "invoice_number_display": "Draft",
      "child_id": "uuid",
      "child_name": "Alex Child",
      "billing_month": "2026-05",
      "period": { "start_date": "2026-05-01", "end_date": "2026-05-31" },
      "status": "draft",
      "due_status": "not_due",
      "currency_code": "GBP",
      "subtotal_minor": 4000,
      "funded_deduction_minor": 2500,
      "total_due_minor": 1500,
      "amount_paid_minor": 0,
      "due_at": null,
      "issued_at": null,
      "paid_at": null,
      "payment_failed_at": null,
      "payment_status_updated_at": null,
      "generated_run_id": "uuid",
      "generated_run_status": "completed_with_exceptions",
      "generated_run_started_at": "2026-05-29T10:00:00Z",
      "generated_run_completed_at": "2026-05-29T10:00:03Z",
      "generated_run_exception_count": 2,
      "created_at": "2026-05-29T10:00:03Z",
      "updated_at": "2026-05-29T10:00:03Z"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

**Draft display rules:**
- `invoice_number`: persisted `null`
- `invoice_number_display`: `"Draft"`
- `due_at`: persisted `null`
- `due_status`: `"not_due"`

**`due_status` values:**

| Invoice status | `due_status` |
|----------------|-------------|
| `draft` | `not_due` |
| `issued`, `payment_failed` | `due` |
| `overdue` | `overdue` |
| `paid` | `paid` |

### `GET /api/v1/invoices/:invoice_id`

Full detail for one invoice including lines, calculation, and run exceptions.

**Response `200`:**

```json
{
  "invoice_id": "uuid",
  "invoice_kind": "monthly",
  "invoice_number": null,
  "invoice_number_display": "Draft",
  "child_id": "uuid",
  "child_name": "Alex Child",
  "billing_month": "2026-05",
  "period": { "start_date": "2026-05-01", "end_date": "2026-05-31" },
  "status": "draft",
  "due_status": "not_due",
  "currency_code": "GBP",
  "subtotal_minor": 4000,
  "funded_deduction_minor": 2500,
  "total_due_minor": 1500,
  "amount_paid_minor": 0,
  "issued_at": null,
  "locked_at": null,
  "due_at": null,
  "paid_at": null,
  "payment_failed_at": null,
  "payment_status_updated_at": null,
  "adjusts_invoice_id": null,
  "adjustment_reason_code": null,
  "adjustment_reason_note": null,
  "generated_run_id": "uuid",
  "generated_run_status": "completed_with_exceptions",
  "generated_run_started_at": "2026-05-29T10:00:00Z",
  "generated_run_completed_at": "2026-05-29T10:00:03Z",
  "generated_run_exception_count": 2,
  "generated_run_exceptions": [
    {
      "child_id": "uuid",
      "child_name": "Blocked Child",
      "blocker_codes": ["missing_funding_profile"]
    }
  ],
  "calculation": {
    "core_hourly_rate_minor": 500,
    "raw_attended_minutes": 480,
    "rounded_attended_minutes": 480,
    "funded_allowance_minutes": 300,
    "funded_deduction_minutes": 300,
    "core_billable_minutes": 180,
    "included_session_count": 1,
    "core_subtotal_minor": 4000,
    "extras_total_minor": 0,
    "source_sessions": [
      {
        "session_id": "uuid",
        "status": "complete",
        "check_in_at": "2026-05-15T08:00:00Z",
        "check_out_at": "2026-05-15T16:00:00Z",
        "raw_elapsed_minutes": 480,
        "rounded_billable_minutes": 480
      }
    ]
  },
  "lines": [
    {
      "line_id": "uuid",
      "line_kind": "core_childcare",
      "description": "Core childcare",
      "sort_order": 1,
      "quantity_minutes": 480,
      "unit_amount_minor": 500,
      "line_amount_minor": 4000,
      "raw_attended_minutes": 480,
      "rounded_attended_minutes": 480,
      "funded_allowance_minutes": null,
      "funded_deduction_minutes": null,
      "core_billable_minutes": null,
      "session_count": 1
    },
    {
      "line_id": "uuid",
      "line_kind": "funded_deduction",
      "description": "Funded hours deduction",
      "sort_order": 2,
      "quantity_minutes": null,
      "unit_amount_minor": null,
      "line_amount_minor": -2500,
      "funded_allowance_minutes": 300,
      "funded_deduction_minutes": 300,
      "core_billable_minutes": 180,
      "session_count": null
    }
  ],
  "created_at": "2026-05-29T10:00:03Z",
  "updated_at": "2026-05-29T10:00:03Z"
}
```

**Error responses:**

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed query param or invoice_id |
| 401 | `unauthorized` | Missing/invalid token |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `invoice_not_found` | Invoice absent from tenant/branch scope |

---

### Parent Invoice View

Parent-facing invoice endpoints. Parents see only issued-or-later invoices for children reachable through the active parent membership → guardian mapping → active guardian-child link.

**Role:** `parent` only. Managers and practitioners receive `403 forbidden_role`.

#### `GET /api/v1/parent/invoices`

List invoices visible to the authenticated parent.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `billing_month` | `string` | No | Filter by billing month (`YYYY-MM`) |
| `status` | `string` | No | One of `issued`, `payment_failed`, `paid`, `overdue`. `draft` is invalid. |
| `child_id` | `string` (UUID) | No | Filter to a specific child. Unlinked child returns empty list. |
| `limit` | `int` | No | Page size, default 50, valid 1–200 |
| `offset` | `int` | No | Page offset, default 0, valid ≥ 0 |

Ordering: `overdue` → `payment_failed` → `issued` → `paid`, then `due_at ASC NULLS LAST`, then `billing_month DESC`, then `child_name ASC`, then `invoice_id ASC`.

Response:

```json
{
  "items": [
    {
      "invoice_id": "uuid",
      "invoice_kind": "monthly",
      "invoice_number": "INV-202605-0001",
      "invoice_number_display": "INV-202605-0001",
      "child_id": "uuid",
      "child_name": "Alex Child",
      "billing_month": "2026-05",
      "period": { "start_date": "2026-05-01", "end_date": "2026-05-31" },
      "status": "issued",
      "due_status": "due",
      "currency_code": "GBP",
      "subtotal_minor": 4000,
      "funded_deduction_minor": 2500,
      "total_due_minor": 1500,
      "amount_paid_minor": 0,
      "issued_at": "2026-05-29T10:00:00Z",
      "due_at": "2026-05-29T10:00:00Z",
      "paid_at": null,
      "payment_failed_at": null,
      "payment_status_updated_at": null
    }
  ],
  "limit": 50,
  "offset": 0
}
```

#### `GET /api/v1/parent/invoices/:invoice_id`

Detail for a single parent-visible invoice.

Response includes `calculation` (without `source_sessions`) and `lines` (without line IDs or line-level calculation snapshots).

```json
{
  "invoice_id": "uuid",
  "invoice_kind": "monthly",
  "invoice_number": "INV-202605-0001",
  "invoice_number_display": "INV-202605-0001",
  "child_id": "uuid",
  "child_name": "Alex Child",
  "billing_month": "2026-05",
  "period": { "start_date": "2026-05-01", "end_date": "2026-05-31" },
  "status": "issued",
  "due_status": "due",
  "currency_code": "GBP",
  "subtotal_minor": 4000,
  "funded_deduction_minor": 2500,
  "total_due_minor": 1500,
  "amount_paid_minor": 0,
  "issued_at": "2026-05-29T10:00:00Z",
  "due_at": "2026-05-29T10:00:00Z",
  "paid_at": null,
  "payment_failed_at": null,
  "payment_status_updated_at": null,
  "calculation": {
    "core_hourly_rate_minor": 500,
    "raw_attended_minutes": 480,
    "rounded_attended_minutes": 480,
    "funded_allowance_minutes": 300,
    "funded_deduction_minutes": 300,
    "core_billable_minutes": 180,
    "included_session_count": 1,
    "core_subtotal_minor": 4000,
    "extras_total_minor": 0
  },
  "lines": [
    {
      "line_kind": "core_childcare",
      "description": "Core childcare",
      "sort_order": 1,
      "quantity_minutes": 480,
      "unit_amount_minor": 500,
      "line_amount_minor": 4000
    },
    {
      "line_kind": "funded_deduction",
      "description": "Funded hours deduction",
      "sort_order": 2,
      "quantity_minutes": null,
      "unit_amount_minor": null,
      "line_amount_minor": -2500
    }
  ]
}
```

**Omitted fields:** Parent responses do not include `generated_run_id`, `generated_run_status`, `generated_run_started_at`, `generated_run_completed_at`, `generated_run_exception_count`, `generated_run_exceptions`, `locked_at`, `adjusts_invoice_id`, `adjustment_reason_code`, `adjustment_reason_note`, `calculation.source_sessions`, line IDs, line-level calculation snapshots, `created_at`, or `updated_at`.

#### Error responses

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed `billing_month`, `child_id`, `invoice_id`; invalid parent status filter (including `draft`); invalid limit/offset |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden_role` | Manager or practitioner accessing parent routes |
| 404 | `invoice_not_found` | Invoice absent from parent-visible set (wrong scope, draft, unlinked child, wrong tenant/branch, nonexistent) |

---

## Parent Invoice Payment

Parent-only endpoint for initiating Stripe Checkout payment for a payable monthly invoice.

**Route is parent-only.** Unauthenticated requests receive `401 unauthorized`. Manager and practitioner requests receive `403 forbidden_role`.

### POST /api/v1/parent/invoices/:invoice_id/checkout-sessions

Creates a Stripe Checkout Session and returns the checkout URL. Empty request body.

**Payable conditions:** `invoice_kind = monthly`, status in `issued`, `payment_failed`, `overdue`, `currency_code = GBP`, `total_due_minor > 0`, `amount_paid_minor = 0`.

**Response 201:**

```json
{
  "checkout_session_id": "cs_test_...",
  "checkout_url": "https://checkout.stripe.com/c/pay/cs_test_...",
  "payment_attempt_id": "uuid"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `checkout_session_id` | string | Stripe Checkout Session ID |
| `checkout_url` | string | URL to redirect parent to Stripe hosted checkout |
| `payment_attempt_id` | string | Local payment attempt UUID for correlation |

#### Error responses

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed `invoice_id` |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden_role` | Manager or practitioner accessing parent route |
| 404 | `invoice_not_found` | Invoice not visible to parent |
| 409 | `invoice_not_payable` | Paid, zero-total, partial-paid, draft, non-monthly, non-GBP, or otherwise not payable |
| 502 | `payment_provider_error` | Stripe failed to create Checkout Session |
| 503 | `payment_provider_unconfigured` | Stripe secret key missing in local/dev runtime |

### POST /api/v1/stripe/webhooks

Public route. No bearer auth, no cookies, no CSRF required.

Receives Stripe webhook events and reconciles Checkout Session outcomes to local payment attempts and invoices.

**Request:**
- Headers: `Stripe-Signature` (required)
- Body: Raw Stripe event payload (max 64 KB)

**Response 200:**
```json
{ "status": "processed" }
```

Status can be: `processed`, `ignored`, `rejected`, `duplicate`.

**Errors:**
- `400 validation_error` — unreadable/oversized body
- `400 payment_webhook_invalid_signature` — invalid Stripe signature
- `503 payment_provider_unconfigured` — webhook secret not configured (local env only)
- `500 internal_error` — transient processing failure (Stripe may retry)

**Event types handled:**
- `checkout.session.completed` (when `payment_status == paid`)
- `checkout.session.async_payment_succeeded`
- `checkout.session.async_payment_failed`
- `checkout.session.expired`

Other event types (including `payment_intent.*`) are stored but ignored.

**Idempotency:** Duplicate Stripe event IDs return `200` with status `duplicate` and do not reprocess.

**Retry safety:** Transient DB errors roll back all changes and return `500` so Stripe retries the same event.

---

## Manager Payment Diagnostics

Manager-only, invoice-scoped endpoints for inspecting payment state, latest payment attempt diagnostics, and payment event timeline. Payment events are backed by `payment_reconciliation_records` joined to `stripe_webhook_events` for local webhook processing status. Raw Stripe webhook payloads and `stripe_checkout_url` are never returned.

**Routes are manager-only.** Unauthenticated requests receive `401 unauthorized`. Practitioner and parent requests receive `403 forbidden_role`.

### `GET /api/v1/invoices/:invoice_id/payment-status`

Returns invoice payment state summary with latest payment attempt diagnostics and latest payment event.

**Payment events are reconciliation records, not raw Stripe payloads.** `stripe_checkout_url` is not returned.

**Response 200:**

```json
{
  "invoice_id": "uuid",
  "invoice_kind": "monthly",
  "invoice_number": "INV-202605-0001",
  "invoice_number_display": "INV-202605-0001",
  "child_id": "uuid",
  "child_name": "Alex Child",
  "billing_month": "2026-05",
  "status": "payment_failed",
  "due_status": "due",
  "currency_code": "GBP",
  "total_due_minor": 1500,
  "amount_paid_minor": 0,
  "issued_at": "2026-05-29T10:00:00Z",
  "due_at": "2026-05-29T10:00:00Z",
  "paid_at": null,
  "payment_failed_at": "2026-05-31T12:00:00Z",
  "payment_status_updated_at": "2026-05-31T12:00:00Z",
  "checkout_retry_available": true,
  "checkout_retry_reason_code": "available",
  "latest_payment_attempt": {
    "payment_attempt_id": "uuid",
    "status": "payment_failed",
    "amount_minor": 1500,
    "currency_code": "GBP",
    "stripe_checkout_session_id": "cs_test_123",
    "stripe_payment_intent_id": "pi_test_123",
    "stripe_expires_at": "2026-05-31T13:00:00Z",
    "failure_reason": null,
    "provider_error_code": null,
    "provider_error_message": null,
    "created_at": "2026-05-31T11:55:00Z",
    "updated_at": "2026-05-31T12:00:00Z"
  },
  "latest_payment_event": {
    "payment_event_id": "uuid",
    "payment_attempt_id": "uuid",
    "stripe_event_id": "evt_123",
    "stripe_event_type": "checkout.session.async_payment_failed",
    "stripe_checkout_session_id": "cs_test_123",
    "stripe_payment_intent_id": "pi_test_123",
    "outcome": "payment_failed",
    "reason_code": "payment_failed",
    "previous_invoice_status": "issued",
    "new_invoice_status": "payment_failed",
    "attempt_previous_status": "checkout_created",
    "attempt_new_status": "payment_failed",
    "amount_minor": 1500,
    "currency_code": "GBP",
    "webhook_processing_status": "processed",
    "webhook_processing_reason": "payment_failed",
    "webhook_received_at": "2026-05-31T12:00:00Z",
    "webhook_processed_at": "2026-05-31T12:00:00Z",
    "created_at": "2026-05-31T12:00:00Z"
  }
}
```

**Nullability:**

- `latest_payment_attempt` is `null` when no attempt exists.
- `latest_payment_event` is `null` when no reconciliation record exists.
- Provider IDs and timestamps are nullable when absent.

**Retry reason codes:**

| Code | Condition |
|------|-----------|
| `available` | Monthly, issued/payment_failed/overdue, GBP, positive balance, fully unpaid |
| `invoice_not_issued` | Status not in issued/payment_failed/overdue (e.g. draft) |
| `invoice_already_paid` | Status is paid |
| `invoice_zero_total` | `total_due_minor <= 0` |
| `invoice_partial_paid` | `amount_paid_minor > 0` |
| `invoice_not_monthly` | `invoice_kind` is not monthly |
| `currency_not_supported` | Currency is not GBP |

**Retry availability:** `checkout_retry_available` is `true` only when `invoice_kind = monthly`, `status IN ('issued', 'payment_failed', 'overdue')`, `currency_code = GBP`, `total_due_minor > 0`, and `amount_paid_minor = 0`. Existing open or previous attempts do not block retry. `payment_failed` and `overdue` are retryable.

**Errors:**

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed `invoice_id` |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `invoice_not_found` | Invoice absent from tenant/branch scope |

### `GET /api/v1/invoices/:invoice_id/payment-events`

Returns a paginated list of payment events (reconciliation records) for an invoice, newest first.

| Parameter | Type | Default | Valid |
|-----------|------|---------|-------|
| `limit` | `int` | `50` | `1..200` |
| `offset` | `int` | `0` | `0+` |

**Ordering:** `created_at DESC, id DESC`.

**Response 200:**

```json
{
  "items": [
    {
      "payment_event_id": "uuid",
      "payment_attempt_id": "uuid",
      "stripe_event_id": "evt_123",
      "stripe_event_type": "checkout.session.completed",
      "stripe_checkout_session_id": "cs_test_123",
      "stripe_payment_intent_id": "pi_test_123",
      "outcome": "paid",
      "reason_code": "paid",
      "previous_invoice_status": "issued",
      "new_invoice_status": "paid",
      "attempt_previous_status": "checkout_created",
      "attempt_new_status": "paid",
      "amount_minor": 1500,
      "currency_code": "GBP",
      "webhook_processing_status": "processed",
      "webhook_processing_reason": "paid",
      "webhook_received_at": "2026-05-31T12:00:00Z",
      "webhook_processed_at": "2026-05-31T12:00:00Z",
      "created_at": "2026-05-31T12:00:00Z"
    }
  ],
  "limit": 50,
  "offset": 0
}
```

**Errors:**

| Status | Code | Condition |
|--------|------|-----------|
| 400 | `validation_error` | Malformed `invoice_id`, invalid `limit`, invalid `offset` |
| 401 | `unauthorized` | Missing or invalid token |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `invoice_not_found` | Invoice absent from tenant/branch scope |

---

## Invoice Issue

Manager-only endpoints for issuing individual invoices or bulk-issuing all draft invoices for a billing month. Issuing locks invoice contents, assigns an invoice number, and sets due date. Issue does not recalculate invoice contents; full details remain available through `GET /api/v1/invoices/:invoice_id`.

**Routes are manager-only.** Unauthenticated requests receive `401 unauthorized`. Practitioner and parent requests receive `403 forbidden_role`.

### POST /api/v1/invoices/:invoice_id/issue

Issue a single draft invoice. Sets status to `issued`, assigns `invoice_number`, populates `issued_at`, `locked_at`, and `due_at`.

**Request:**

```json
{ "confirm": true }
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `confirm` | boolean | yes | Must be `true` to proceed |

**Response 200:**

```json
{
  "invoice_id": "uuid",
  "invoice_number": "INV-2026-00001",
  "status": "issued",
  "issued_at": "2026-05-29T14:00:00Z",
  "locked_at": "2026-05-29T14:00:00Z",
  "due_at": "2026-05-29T14:00:00Z",
  "issued_run_id": "uuid",
  "total_due_minor": 1500
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid `invoice_id` or `confirm` not `true` |
| 401 | `unauthorized` | No valid token |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `invoice_not_found` | Invoice absent from tenant/branch scope |
| 409 | `invoice_not_draft` | Invoice status is not `draft` |
| 409 | `invoice_not_monthly` | Invoice kind is not `monthly` |

### POST /api/v1/invoices/bulk-issue

Bulk-issue all draft invoices for a billing month. Each eligible draft invoice is issued atomically. Invoices with blockers are skipped and reported in the `blocked` array. The entire operation runs in a single transaction; unexpected errors roll back all changes.

**Request:**

```json
{
  "billing_month": "2026-05",
  "invoice_ids": ["uuid", "uuid"],
  "confirm": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `billing_month` | string | yes | Billing month as `YYYY-MM` |
| `invoice_ids` | string[] | no | Optional array of invoice UUIDs. Omit to issue all drafts for the billing month. Empty array = no-op. Duplicates are deduplicated. |
| `confirm` | boolean | yes | Must be `true` to proceed |

**Response 200:**

```json
{
  "run_id": "uuid",
  "billing_month": "2026-05",
  "status": "completed_with_exceptions",
  "summary": {
    "total_count": 5,
    "issued_count": 3,
    "blocked_count": 2,
    "total_due_minor": 8500
  },
  "issued": [
    {
      "invoice_id": "uuid",
      "invoice_number": "INV-2026-00001",
      "child_id": "uuid",
      "child_name": "Alex Child",
      "status": "issued",
      "issued_at": "2026-05-29T14:00:00Z",
      "locked_at": "2026-05-29T14:00:00Z",
      "due_at": "2026-05-29T14:00:00Z",
      "total_due_minor": 1500
    }
  ],
  "blocked": [
    {
      "invoice_id": "uuid",
      "child_id": "uuid",
      "child_name": "Bailey Child",
      "blockers": [
        { "code": "invoice_not_draft", "message": "Invoice status is not draft." }
      ]
    }
  ]
}
```

**`status` values:** `completed` (all invoices issued), `completed_with_exceptions` (one or more invoices blocked).

**Blocker codes:**

| Code | Meaning |
|------|---------|
| `invoice_not_found` | Invoice ID not found in tenant/branch scope |
| `invoice_not_in_billing_month` | Invoice does not belong to the requested billing month |
| `invoice_not_draft` | Invoice status is not `draft` |
| `invoice_not_monthly` | Invoice kind is not `monthly` |

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Missing/malformed `billing_month`, `confirm` not `true`, or malformed `invoice_ids` |
| 401 | `unauthorized` | No valid token |
| 403 | `forbidden_role` | Non-manager role |

---

## Funding v1

Manager-only endpoints for maintaining a child's funded-hours allowance per billing month. Parents see funding effects through issued invoices, not through these routes.

**Routes are manager-only.** Practitioner and parent requests receive `403 forbidden_role`.

**`billing_month`** is required on all funding requests, formatted as `YYYY-MM`.

**Zero allowance is explicit** and distinct from a missing funding profile. A missing profile means no funding has been configured for that child/month.

### GET /funding/children/:child_id?billing_month=YYYY-MM

Retrieve a child's funding profile for a specific billing month.

**Request:**

| Parameter | Location | Required | Description |
|-----------|----------|----------|-------------|
| `child_id` | path | yes | UUID of the child |
| `billing_month` | query | yes | Billing month as `YYYY-MM` |

**Response 200:**

```json
{
  "id": "uuid",
  "child_id": "uuid",
  "billing_month": "2026-05",
  "funded_allowance_minutes": 570,
  "created_at": "2026-05-26T10:00:00Z",
  "updated_at": "2026-05-26T10:00:00Z"
}
```

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid child_id, missing billing_month, or invalid month format |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `funding_profile_not_found` | No profile for this child/month |

### PUT /funding/children/:child_id

Create or update a child's funding profile for a billing month. Upsert semantics: returns `201` on create, `200` on update or unchanged save. An unchanged save does not update `updated_at` or write an audit event.

**Request body:**

```json
{
  "billing_month": "2026-05",
  "funded_allowance_minutes": 570
}
```

| Field | Type | Required | Bounds |
|-------|------|----------|--------|
| `billing_month` | string | yes | `YYYY-MM` format |
| `funded_allowance_minutes` | integer | yes | 0–44640 |

**Response 201** (created):

```json
{
  "id": "uuid",
  "child_id": "uuid",
  "billing_month": "2026-05",
  "funded_allowance_minutes": 570,
  "created_at": "2026-05-26T10:00:00Z",
  "updated_at": "2026-05-26T10:00:00Z"
}
```

**Response 200** (updated or unchanged): same shape as above.

**Errors:**

| Status | Code | When |
|--------|------|------|
| 400 | `validation_error` | Invalid child_id, invalid month, or allowance outside 0–44640 |
| 403 | `forbidden_role` | Non-manager role |
| 404 | `child_not_found` | Child does not exist in tenant/branch scope |
| 409 | `funding_month_outside_enrollment_window` | Billing month is fully before start_date or fully after end_date |

---

## Known Contract Gaps

1. **Relationship read endpoints not implemented.** Child detail linked guardian and parent access status must use mock data until `GET /children/:child_id/guardian-child-links` and `GET /guardians/:guardian_id/parent-membership-guardian-mappings` are built.

2. **Attendance event history not implemented.** Correction history UI must use mock data until `GET /attendance/sessions/:session_id/events` is built.

3. **`/me` response is a debug shape.** Current response returns raw authorization context. Not a rich user profile.

4. **Guardian and child PATCH cannot clear optional fields.** Sending an empty string for `notes`, `email`, or `phone` is ignored because empty fields are dropped before update. To clear, send `null` if the DTO supports it — otherwise the field retains its previous value.

5. **Invite error HTTP status codes.** The following invite error codes are produced by application use cases but the error mapper does not route them to their intended HTTP status. They surface as HTTP 500 with the correct code in the response body:
   - `invite_role_not_allowed` (intended 400/403)
   - `invite_not_pending` (intended 409)
   - `invite_already_accepted` (intended 409)
   - `invite_email_already_registered` (intended 409)
   - `invite_scope_conflict` (intended 409)

6. **`parent_mapping_active_conflict` returns 500.** The error mapper's `_conflict` suffix check has a length mismatch bug, so this code surfaces as HTTP 500 with the correct code in the response body. Intended status is 409.

7. **Attendance correction malformed timestamps map to generic `validation_error`.** Invalid RFC 3339 strings or malformed UUIDs are caught by binding validation and return `validation_error` without distinguishing the field.

8. **Authz probe routes exist but are not documented above.** `/api/v1/authz/probe/manager`, `/practitioner`, `/parent`, `/scope/:tenant_id/:branch_id`, and `/parent-link/:child_id` return `{ "status": "ok" }` for authorized requests. These are debug/integration-test helpers, not production frontend routes.

9. **Health check routes.** `GET /health` and `GET /api/v1/health` return `{ "status": "ok", "timestamp": "...", "request_id": "..." }`. Returns `503 db_unavailable` if the database is unreachable.
