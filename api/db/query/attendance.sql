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
          (check_out_at IS NOT NULL AND check_in_at < $4 AND check_out_at > $5)
          OR
          (check_out_at IS NULL AND check_in_at < $4)
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

-- name: AttendanceListIncompleteSessionsForPeriod :many
SELECT s.child_id,
       c.first_name AS child_first_name,
       c.middle_name AS child_middle_name,
       c.last_name AS child_last_name,
       s.id AS session_id,
       s.check_in_at,
       s.check_in_local_date
FROM attendance_sessions s
JOIN children c
  ON c.tenant_id = s.tenant_id
 AND c.branch_id = s.branch_id
 AND c.id = s.child_id
WHERE s.tenant_id = $1
  AND s.branch_id = $2
  AND s.status = 'open'
  AND s.check_in_local_date >= $3
  AND s.check_in_local_date < $4
ORDER BY s.check_in_local_date ASC, c.first_name ASC, c.middle_name ASC NULLS FIRST, c.last_name ASC NULLS FIRST, c.id ASC, s.check_in_at ASC;

-- name: AttendanceListSessionsForCorrection :many
SELECT s.id, s.child_id, s.status, s.check_in_at, s.check_out_at,
       s.check_in_local_date, s.check_out_local_date,
       EXTRACT(EPOCH FROM (s.check_out_at - s.check_in_at))::bigint / 60 AS duration_minutes,
       s.created_at, s.updated_at
FROM attendance_sessions s
WHERE s.tenant_id = $1
  AND s.branch_id = $2
  AND s.child_id = $3
  AND s.check_in_local_date = $4
ORDER BY s.check_in_at ASC;

-- name: AttendanceListSessionEventsForHistory :many
SELECT e.id, e.session_id, e.event_type, e.occurred_at, e.local_date,
       e.recorded_by_user_id, e.recorded_by_membership_id,
       u.email AS recorded_by_label,
       e.reason_code, e.reason_note, e.details,
       e.created_at
FROM attendance_events e
LEFT JOIN users u ON u.id = e.recorded_by_user_id
WHERE e.tenant_id = $1
  AND e.branch_id = $2
  AND e.session_id = $3
ORDER BY e.occurred_at ASC, e.created_at ASC;

-- name: AttendanceGetIssuedInvoiceWarningForMonth :one
SELECT i.id, i.invoice_number, i.billing_month, i.status
FROM invoices i
WHERE i.tenant_id = $1
  AND i.branch_id = $2
  AND i.child_id = $3
  AND i.billing_month = $4
  AND i.invoice_kind = 'monthly'
  AND i.status IN ('issued', 'payment_failed', 'paid', 'overdue');
