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
      "has_incomplete_session": true
    }
  ]
}
```

`attendance_state` values: `not_checked_in`, `checked_in`.

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
