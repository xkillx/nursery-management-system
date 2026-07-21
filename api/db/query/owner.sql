-- Owner module queries: site summaries and manager-access administration.
-- All aggregate queries return counts only; no named records or operational identifiers.

-- ── Site lookups ──────────────────────────────────────────────────────────────

-- name: OwnerGetActiveSites :many
SELECT id, name, core_hourly_rate_minor, funded_hourly_rate_minor
FROM branches
WHERE tenant_id = $1 AND is_active = true
ORDER BY name;

-- name: OwnerGetActiveSite :one
SELECT id, name, core_hourly_rate_minor, funded_hourly_rate_minor
FROM branches
WHERE tenant_id = $1 AND id = $2 AND is_active = true;

-- ── Manager access status ────────────────────────────────────────────────────

-- name: OwnerCountActiveManagersByBranches :many
SELECT branch_id, COUNT(*)::int AS count
FROM memberships
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND role = 'manager'
  AND is_active = true
  AND ended_at IS NULL
GROUP BY branch_id;

-- name: OwnerCountPendingManagerInvitesByBranches :many
SELECT branch_id, COUNT(*)::int AS count
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND role = 'manager'
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
GROUP BY branch_id;

-- ── Children ──────────────────────────────────────────────────────────────────

-- name: OwnerCountActiveChildrenByBranches :many
SELECT branch_id, COUNT(*)::int AS count
FROM children
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND is_active = true
GROUP BY branch_id;

-- ── Attendance today ──────────────────────────────────────────────────────────

-- name: OwnerCountAttendanceTodayByBranches :many
SELECT branch_id,
       COUNT(DISTINCT child_id)::int AS checked_in_count
FROM attendance_sessions
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND status = 'open'
  AND check_in_local_date = $3
GROUP BY branch_id;

-- name: OwnerCountIncompleteAttendanceByBranches :many
SELECT branch_id, COUNT(*)::int AS count
FROM attendance_sessions
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND status = 'open'
  AND check_in_local_date >= $3
  AND check_in_local_date < $4
GROUP BY branch_id;

-- ── Funding readiness ─────────────────────────────────────────────────────────

-- name: OwnerGetFundingReadinessByBranches :many
SELECT c.branch_id,
       COUNT(*)::int AS included_child_count,
       COUNT(*) FILTER (WHERE fp.id IS NULL)::int AS missing_profile_count,
       COUNT(*) FILTER (WHERE fp.funded_allowance_minutes = 0)::int AS explicit_zero_count,
       COUNT(*) FILTER (WHERE fp.funded_allowance_minutes > 0 AND fp.funded_allowance_minutes < 60)::int AS under_one_hour_count,
       COUNT(*) FILTER (WHERE fp.funded_allowance_minutes > 9600)::int AS above_160_hours_count
FROM children c
LEFT JOIN funding_profiles fp
  ON fp.tenant_id = c.tenant_id
  AND fp.branch_id = c.branch_id
  AND fp.child_id = c.id
  AND fp.billing_month = $3
WHERE c.tenant_id = $1
  AND c.branch_id = ANY($2::uuid[])
  AND c.is_active = true
  AND c.start_date < ($3 + INTERVAL '1 month')::date
  AND (c.end_date IS NULL OR c.end_date >= $3)
GROUP BY c.branch_id;

-- ── Invoice / payment health ──────────────────────────────────────────────────

-- name: OwnerGetInvoicePaymentHealthByBranches :many
SELECT branch_id,
       currency_code,
       COUNT(*) FILTER (WHERE status = 'draft')::int AS draft_count,
       COUNT(*) FILTER (WHERE status = 'issued')::int AS issued_count,
       COUNT(*) FILTER (WHERE status = 'overdue')::int AS overdue_count,
       COUNT(*) FILTER (WHERE status = 'payment_failed')::int AS payment_failed_count,
       COUNT(*) FILTER (WHERE status = 'paid')::int AS paid_count,
       COALESCE(SUM(total_due_minor) FILTER (WHERE status IN ('issued', 'overdue')), 0)::bigint AS total_issued_minor,
       COALESCE(SUM(amount_paid_minor) FILTER (WHERE status = 'paid'), 0)::bigint AS total_paid_minor,
       COALESCE(SUM(total_due_minor - amount_paid_minor) FILTER (WHERE status IN ('issued', 'overdue')), 0)::bigint AS outstanding_minor,
       COALESCE(SUM(total_due_minor - amount_paid_minor) FILTER (WHERE status = 'overdue'), 0)::bigint AS overdue_outstanding_minor
FROM invoices
WHERE tenant_id = $1
  AND branch_id = ANY($2::uuid[])
  AND billing_month = $3
GROUP BY branch_id, currency_code;

-- ── User lookup ───────────────────────────────────────────────────────────────

-- name: OwnerFindActiveUserByEmail :one
SELECT id, email, is_active
FROM users
WHERE email_normalized = $1 AND is_active = true;

-- ── Manager membership queries ────────────────────────────────────────────────

-- name: OwnerFindManagerMembershipForUser :one
SELECT id, tenant_id, branch_id, user_id, role, is_active, ended_at
FROM memberships
WHERE tenant_id = $1
  AND branch_id = $2
  AND user_id = $3
  AND role = 'manager';

-- name: OwnerCreateManagerMembership :exec
INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active)
VALUES ($1, $2, $3, $4, 'manager', true);

-- name: OwnerReactivateManagerMembership :exec
UPDATE memberships
SET is_active = true,
    ended_at = NULL,
    updated_at = now()
WHERE id = $1 AND tenant_id = $2 AND role = 'manager';

-- name: OwnerDeactivateManagerMembership :execrows
UPDATE memberships
SET is_active = false,
    ended_at = now(),
    updated_at = now()
WHERE id = $1 AND tenant_id = $2 AND role = 'manager' AND is_active = true;

-- ── Manager access listing ────────────────────────────────────────────────────

-- name: OwnerListManagerAccess :many
SELECT m.id AS membership_id,
       m.user_id,
       u.email,
       m.is_active,
       m.ended_at
FROM memberships m
JOIN users u ON u.id = m.user_id
WHERE m.tenant_id = $1
  AND m.branch_id = $2
  AND m.role = 'manager'
  AND (
      sqlc.arg('status_filter') = 'all'
      OR (sqlc.arg('status_filter') = 'active' AND m.is_active = true AND m.ended_at IS NULL)
      OR (sqlc.arg('status_filter') = 'inactive' AND m.is_active = false)
  )
ORDER BY u.email;

-- ── Manager invite queries ────────────────────────────────────────────────────

-- name: OwnerFindPendingManagerInvite :one
SELECT id, email, email_normalized, expires_at, send_count, created_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND email_normalized = $3
  AND role = 'manager'
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: OwnerCreateManagerInvite :exec
INSERT INTO manager_invites (
    id, tenant_id, branch_id, email, email_normalized, role,
    token_hash, expires_at,
    created_by_user_id, created_by_membership_id
) VALUES ($1, $2, $3, $4, $5, 'manager', $6, $7, $8, $9);

-- name: OwnerRefreshManagerInvite :exec
UPDATE manager_invites
SET token_hash = $2,
    expires_at = $3,
    resent_at = now(),
    resent_by_user_id = $4,
    resent_by_membership_id = $5,
    send_count = send_count + 1,
    updated_at = now()
WHERE id = $1;

-- ── Refresh token revocation ──────────────────────────────────────────────────

-- name: OwnerRevokeRefreshTokensByMembership :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE membership_id = $1 AND revoked_at IS NULL;

-- name: OwnerListManagerAccessPaginated :many
SELECT m.id AS membership_id,
       m.user_id,
       u.email,
       m.is_active,
       m.ended_at
FROM memberships m
JOIN users u ON u.id = m.user_id
WHERE m.tenant_id = $1
  AND m.branch_id = $2
  AND m.role = 'manager'
  AND (
      sqlc.arg('status_filter') = 'all'
      OR (sqlc.arg('status_filter') = 'active' AND m.is_active = true AND m.ended_at IS NULL)
      OR (sqlc.arg('status_filter') = 'inactive' AND m.is_active = false)
  )
ORDER BY u.email
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: OwnerCountManagerAccess :one
SELECT COUNT(*)
FROM memberships m
WHERE m.tenant_id = $1
  AND m.branch_id = $2
  AND m.role = 'manager'
  AND (
      sqlc.arg('status_filter') = 'all'
      OR (sqlc.arg('status_filter') = 'active' AND m.is_active = true AND m.ended_at IS NULL)
      OR (sqlc.arg('status_filter') = 'inactive' AND m.is_active = false)
  );
