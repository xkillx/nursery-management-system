-- name: SessionTemplateEntriesListByTemplate :many
SELECT
    e.id,
    e.tenant_id,
    e.branch_id,
    e.template_id,
    e.day_of_week,
    e.session_type_id,
    e.created_at,
    e.updated_at,
    st.name AS session_type_name,
    st.start_time AS session_type_start_time,
    st.end_time AS session_type_end_time,
    st.is_active AS session_type_is_active
FROM session_template_entries e
JOIN session_types st
  ON st.tenant_id = e.tenant_id
 AND st.branch_id = e.branch_id
 AND st.id = e.session_type_id
WHERE e.tenant_id = $1 AND e.branch_id = $2 AND e.template_id = $3
ORDER BY e.day_of_week ASC, st.start_time ASC;

-- name: SessionTemplateEntriesInsert :exec
INSERT INTO session_template_entries (id, tenant_id, branch_id, template_id, day_of_week, session_type_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: SessionTemplateEntriesDeleteByTemplate :exec
DELETE FROM session_template_entries
WHERE tenant_id = $1 AND branch_id = $2 AND template_id = $3;
