-- name: AttendanceRegisterSummary :many
SELECT
    'recurring' AS booking_type,
    r.id AS room_id,
    COALESCE(r.name, '') AS room_name,
    d.dt::date AS register_date,
    COUNT(*) AS booking_count
FROM bookings b
JOIN children c ON c.id = b.child_id AND c.tenant_id = b.tenant_id AND c.branch_id = b.branch_id
JOIN child_room_assignments cra ON cra.child_id = c.id AND cra.tenant_id = c.tenant_id AND cra.branch_id = c.branch_id
    AND cra.start_date <= d.dt AND (cra.end_date IS NULL OR cra.end_date >= d.dt)
JOIN rooms r ON r.id = cra.room_id AND r.tenant_id = cra.tenant_id AND r.branch_id = cra.branch_id
CROSS JOIN generate_series(@from_date::date, @to_date::date, '1 day') AS d(dt)
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.status = 'active'
  AND b.effective_start_date <= d.dt
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= d.dt)
  AND EXTRACT(DOW FROM d.dt)::int = ANY(b.days_of_week)
GROUP BY r.id, r.name, d.dt::date

UNION ALL

SELECT
    'ad_hoc' AS booking_type,
    r.id AS room_id,
    COALESCE(r.name, '') AS room_name,
    ah.calendar_date AS register_date,
    COUNT(*) AS booking_count
FROM ad_hoc_bookings ah
JOIN children c ON c.id = ah.child_id AND c.tenant_id = ah.tenant_id AND c.branch_id = ah.branch_id
JOIN child_room_assignments cra ON cra.child_id = c.id AND cra.tenant_id = c.tenant_id AND cra.branch_id = c.branch_id
    AND cra.start_date <= ah.calendar_date AND (cra.end_date IS NULL OR cra.end_date >= ah.calendar_date)
JOIN rooms r ON r.id = cra.room_id AND r.tenant_id = cra.tenant_id AND r.branch_id = cra.branch_id
WHERE ah.tenant_id = $1
  AND ah.branch_id = $2
  AND ah.calendar_date >= @from_date
  AND ah.calendar_date <= @to_date
  AND ah.status = 'active'
GROUP BY r.id, r.name, ah.calendar_date

UNION ALL

SELECT
    'hourly' AS booking_type,
    r.id AS room_id,
    COALESCE(r.name, '') AS room_name,
    h.calendar_date AS register_date,
    COUNT(*) AS booking_count
FROM hourly_bookings h
JOIN children c ON c.id = h.child_id AND c.tenant_id = h.tenant_id AND c.branch_id = h.branch_id
JOIN child_room_assignments cra ON cra.child_id = c.id AND cra.tenant_id = c.tenant_id AND cra.branch_id = c.branch_id
    AND cra.start_date <= h.calendar_date AND (cra.end_date IS NULL OR cra.end_date >= h.calendar_date)
JOIN rooms r ON r.id = cra.room_id AND r.tenant_id = cra.tenant_id AND r.branch_id = cra.branch_id
WHERE h.tenant_id = $1
  AND h.branch_id = $2
  AND h.calendar_date >= @from_date
  AND h.calendar_date <= @to_date
  AND h.status = 'active'
GROUP BY r.id, r.name, h.calendar_date

ORDER BY register_date, room_name;
