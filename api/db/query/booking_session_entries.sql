-- name: BookingSessionEntriesCreate :exec
INSERT INTO booking_session_entries (id, tenant_id, branch_id, booking_id, day_of_week, session_type_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: BookingSessionEntriesCreateBatch :batchexec
INSERT INTO booking_session_entries (id, tenant_id, branch_id, booking_id, day_of_week, session_type_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: BookingSessionEntriesDeleteByBooking :exec
DELETE FROM booking_session_entries
WHERE tenant_id = $1
  AND branch_id = $2
  AND booking_id = $3;

-- name: BookingSessionEntriesListByBooking :many
SELECT id, tenant_id, branch_id, booking_id, day_of_week, session_type_id, created_at, updated_at
FROM booking_session_entries
WHERE tenant_id = $1
  AND branch_id = $2
  AND booking_id = $3
ORDER BY day_of_week;

-- name: BookingSessionEntriesForChildInMonth :many
SELECT
    bse.day_of_week,
    st.id AS session_type_id,
    st.name AS session_type_name,
    st.start_time AS session_type_start_time,
    st.end_time AS session_type_end_time
FROM bookings b
JOIN booking_session_entries bse ON bse.tenant_id = b.tenant_id
    AND bse.branch_id = b.branch_id
    AND bse.booking_id = b.id
JOIN session_types st ON st.id = bse.session_type_id
    AND st.tenant_id = b.tenant_id
    AND st.branch_id = b.branch_id
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.child_id = $3
  AND b.status = 'active'
  AND b.effective_start_date <= @month_end
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= @month_start)
ORDER BY bse.day_of_week;
