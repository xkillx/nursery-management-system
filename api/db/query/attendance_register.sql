-- name: AttendanceRegisterSummary :many
SELECT
    'recurring' AS booking_type,
    b.room_id,
    COALESCE(r.name, '') AS room_name,
    d.dt::date AS register_date,
    COUNT(*) AS booking_count
FROM bookings b
JOIN rooms r ON r.id = b.room_id AND r.tenant_id = b.tenant_id AND r.branch_id = b.branch_id
CROSS JOIN generate_series(@from_date::date, @to_date::date, '1 day') AS d(dt)
WHERE b.tenant_id = $1
  AND b.branch_id = $2
  AND b.status = 'active'
  AND b.effective_start_date <= d.dt
  AND (b.effective_end_date IS NULL OR b.effective_end_date >= d.dt)
  AND EXTRACT(DOW FROM d.dt)::int = ANY(b.days_of_week)
GROUP BY b.room_id, r.name, d.dt::date

UNION ALL

SELECT
    'ad_hoc' AS booking_type,
    NULL::uuid AS room_id,
    '' AS room_name,
    ah.calendar_date AS register_date,
    COUNT(*) AS booking_count
FROM ad_hoc_bookings ah
WHERE ah.tenant_id = $1
  AND ah.branch_id = $2
  AND ah.calendar_date >= @from_date
  AND ah.calendar_date <= @to_date
  AND ah.status = 'active'
GROUP BY ah.calendar_date

UNION ALL

SELECT
    'hourly' AS booking_type,
    NULL::uuid AS room_id,
    '' AS room_name,
    h.calendar_date AS register_date,
    COUNT(*) AS booking_count
FROM hourly_bookings h
WHERE h.tenant_id = $1
  AND h.branch_id = $2
  AND h.calendar_date >= @from_date
  AND h.calendar_date <= @to_date
  AND h.status = 'active'
GROUP BY h.calendar_date

ORDER BY register_date, room_name;
