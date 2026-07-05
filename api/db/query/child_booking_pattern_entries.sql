-- name: ChildBookingPatternEntriesListByPattern :many
SELECT
    e.id,
    e.tenant_id,
    e.branch_id,
    e.pattern_id,
    e.day_of_week,
    e.session_type_id,
    e.created_at,
    e.updated_at,
    st.name AS session_type_name,
    st.start_time AS session_type_start_time,
    st.end_time AS session_type_end_time,
    st.is_active AS session_type_is_active,
    st.kind AS session_type_kind,
    st.flat_fee_minor AS session_type_flat_fee_minor
FROM child_booking_pattern_entries e
JOIN session_types st
  ON st.tenant_id = e.tenant_id
 AND st.branch_id = e.branch_id
 AND st.id = e.session_type_id
WHERE e.tenant_id = $1 AND e.branch_id = $2 AND e.pattern_id = $3
ORDER BY e.day_of_week ASC, st.start_time ASC;

-- name: ChildBookingPatternEntriesInsert :exec
INSERT INTO child_booking_pattern_entries (id, tenant_id, branch_id, pattern_id, day_of_week, session_type_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ChildBookingPatternEntriesDeleteByPattern :exec
DELETE FROM child_booking_pattern_entries
WHERE tenant_id = $1 AND branch_id = $2 AND pattern_id = $3;
