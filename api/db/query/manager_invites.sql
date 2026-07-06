-- name: InviteAcquireEmailScopeLock :execrows
SELECT pg_advisory_xact_lock(hashtextextended($1, 0));

-- name: InviteFindUserByEmailNormalized :one
SELECT id
FROM users
WHERE email_normalized = $1
LIMIT 1;

-- name: InviteFindLivePendingByEmailForUpdate :one
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND email_normalized = $3
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
FOR UPDATE;

-- name: InviteCreate :exec
INSERT INTO manager_invites (
    id, tenant_id, branch_id, email, email_normalized, role,
    token_hash, expires_at,
    created_by_user_id, created_by_membership_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: InviteRefreshPending :exec
UPDATE manager_invites
SET token_hash = $2,
    expires_at = $3,
    resent_at = now(),
    resent_by_user_id = $4,
    resent_by_membership_id = $5,
    send_count = send_count + 1,
    updated_at = now()
WHERE id = $1;

-- name: InviteListPending :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC;

-- name: InviteListAccepted :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NOT NULL
ORDER BY created_at DESC;

-- name: InviteListRevoked :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND revoked_at IS NOT NULL
ORDER BY created_at DESC;

-- name: InviteListExpired :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at <= now()
ORDER BY created_at DESC;

-- name: InviteListAll :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
ORDER BY created_at DESC;

-- name: InviteGetByIDForUpdate :one
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE id = $1
  AND tenant_id = $2
  AND branch_id = $3
FOR UPDATE;

-- name: InviteGetByTokenHashForUpdate :one
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE token_hash = $1
FOR UPDATE;

-- name: InviteRevoke :exec
UPDATE manager_invites
SET revoked_at = now(),
    revoked_by_user_id = $2,
    revoked_by_membership_id = $3,
    updated_at = now()
WHERE id = $1
  AND accepted_at IS NULL
  AND revoked_at IS NULL;

-- name: InviteAccept :exec
UPDATE manager_invites
SET accepted_at = now(),
    accepted_user_id = $2,
    accepted_membership_id = $3,
    updated_at = now()
WHERE id = $1
  AND accepted_at IS NULL
  AND revoked_at IS NULL;

-- name: InviteCreateUser :execrows
INSERT INTO users (id, email, email_normalized, password_hash, is_active)
VALUES ($1, $2, $3, $4, true);

-- name: InviteCreateMembership :exec
INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active)
VALUES ($1, $2, $3, $4, $5, true);

-- name: InviteListPendingPaginated :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
ORDER BY created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InviteCountPending :one
SELECT COUNT(*)
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text);

-- name: InviteListPendingPaginatedSortByEmailAsc :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
ORDER BY email ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InviteListPendingPaginatedSortByEmailDesc :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
ORDER BY email DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InviteListPendingPaginatedSortByCreatedAtAsc :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
ORDER BY created_at ASC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: InviteListPendingPaginatedSortByCreatedAtDesc :many
SELECT id, tenant_id, branch_id, email, email_normalized, role, token_hash, expires_at,
       accepted_at, accepted_user_id, accepted_membership_id,
       revoked_at, revoked_by_user_id, revoked_by_membership_id,
       created_by_user_id, created_by_membership_id,
       resent_at, resent_by_user_id, resent_by_membership_id,
       send_count, created_at, updated_at
FROM manager_invites
WHERE tenant_id = $1
  AND branch_id = $2
  AND accepted_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > now()
  AND (sqlc.narg('role')::text IS NULL OR role = sqlc.narg('role')::text)
ORDER BY created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');
