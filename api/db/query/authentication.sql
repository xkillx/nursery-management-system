-- name: AuthFindUserByEmail :one
SELECT id, email, password_hash, is_active
FROM users
WHERE email_normalized = $1
LIMIT 1;

-- name: AuthListMembershipsByUserID :many
SELECT m.id, m.tenant_id, t.name AS tenant_name, m.branch_id, b.name AS branch_name, m.role, m.is_active
FROM memberships m
JOIN tenants t ON t.id = m.tenant_id
LEFT JOIN branches b ON b.id = m.branch_id
WHERE m.user_id = $1 AND m.is_active = true AND m.ended_at IS NULL
ORDER BY m.created_at ASC;

-- name: AuthCreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, membership_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: AuthFindActiveRefreshToken :one
SELECT rt.id, rt.user_id, rt.membership_id, rt.token_hash, rt.expires_at, rt.revoked_at,
       u.id AS user_table_id, u.email AS user_email, u.password_hash AS user_password_hash, u.is_active AS user_is_active,
       m.id AS membership_table_id, m.tenant_id AS membership_tenant_id, t.name AS membership_tenant_name, m.branch_id AS membership_branch_id, b.name AS membership_branch_name, m.role AS membership_role, m.is_active AS membership_is_active
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
JOIN memberships m ON m.id = rt.membership_id AND m.user_id = u.id AND m.is_active = true AND m.ended_at IS NULL
JOIN tenants t ON t.id = m.tenant_id
LEFT JOIN branches b ON b.id = m.branch_id
WHERE rt.token_hash = $1
LIMIT 1;

-- name: AuthRevokeRefreshTokenByID :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE id = $1 AND revoked_at IS NULL;

-- name: AuthRevokeRefreshTokenByHash :exec
UPDATE refresh_tokens
SET revoked_at = now(), updated_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: AuthCreateScopeSwitchAuditLog :exec
INSERT INTO audit_logs (id, tenant_id, branch_id, actor_user_id, actor_membership_id, action_type, action_entity_type, action_entity_id, request_id, details)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, jsonb_build_object('from_membership_id', $10::uuid, 'to_membership_id', $11::uuid));
