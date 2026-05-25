-- name: AttendanceInsertOpenSession :exec
INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_in_local_date)
VALUES ($1, $2, $3, $4, 'open', $5, $6);

-- name: AttendanceInsertCheckInEvent :exec
INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id)
VALUES ($1, $2, $3, $4, $5, 'check_in', $6, $7, $8, $9, NULLIF($10, ''));

-- name: AttendanceAttachCheckInEvent :exec
UPDATE attendance_sessions SET check_in_event_id = $1
WHERE tenant_id = $2 AND branch_id = $3 AND id = $4;

-- name: AttendanceGetOpenSessionForUpdate :one
SELECT id, child_id, status, check_in_at, check_in_local_date, created_at, updated_at
FROM attendance_sessions
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3 AND status = 'open'
FOR UPDATE;

-- name: AttendanceInsertCheckOutEvent :exec
INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id)
VALUES ($1, $2, $3, $4, $5, 'check_out', $6, $7, $8, $9, NULLIF($10, ''));

-- name: AttendanceCompleteSession :exec
UPDATE attendance_sessions
SET status = 'complete',
    check_out_at = $1,
    check_out_local_date = $2,
    check_out_event_id = $3,
    updated_at = $4
WHERE tenant_id = $5 AND branch_id = $6 AND id = $7;

-- name: AttendanceGetSessionForCorrection :one
SELECT id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date, created_at, updated_at
FROM attendance_sessions
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
FOR UPDATE;

-- name: AttendanceHasOverlappingSession :one
SELECT EXISTS (
    SELECT 1 FROM attendance_sessions
    WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
      AND status IN ('open', 'complete', 'corrected')
      AND (
          (check_out_at IS NOT NULL AND check_in_at < $5 AND check_out_at > $4)
          OR
          (check_out_at IS NULL AND check_in_at < $5)
      )
      AND id != ALL (SELECT unnest(CASE WHEN $6::uuid IS NOT NULL THEN ARRAY[$6::uuid] ELSE ARRAY[]::uuid[] END))
);

-- name: AttendanceInsertCorrectionEvent :exec
INSERT INTO attendance_events (id, tenant_id, branch_id, child_id, session_id, event_type, occurred_at, local_date, recorded_by_user_id, recorded_by_membership_id, request_id, reason_code, reason_note, details)
VALUES ($1, $2, $3, $4, $5, 'correction', $6, $7, $8, $9, NULLIF($10, ''), $11, NULLIF($12, ''), $13::jsonb);

-- name: AttendanceCorrectSession :exec
UPDATE attendance_sessions
SET status = 'corrected',
    check_in_at = $1,
    check_out_at = $2,
    check_in_local_date = $3,
    check_out_local_date = $4,
    corrected_by_event_id = $5,
    updated_at = $6
WHERE tenant_id = $7 AND branch_id = $8 AND id = $9;

-- name: AttendanceInsertCorrectedSession :exec
INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date)
VALUES ($1, $2, $3, $4, 'corrected', $5, $6, $7, $8);

-- name: AttendanceAttachCorrectedEvent :exec
UPDATE attendance_sessions SET corrected_by_event_id = $1
WHERE tenant_id = $2 AND branch_id = $3 AND id = $4;
