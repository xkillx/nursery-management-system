-- name: AttestationGetLatestByChild :one
SELECT
    id, tenant_id, branch_id, child_id,
    consent_record_id, profile_updated_at,
    attested_by_user_id, attested_by_membership_id, attested_at,
    request_id, created_at
FROM child_registration_completion_attestations
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3
ORDER BY attested_at DESC
LIMIT 1;

-- name: AttestationCreate :exec
INSERT INTO child_registration_completion_attestations (
    id, tenant_id, branch_id, child_id,
    consent_record_id, profile_updated_at,
    attested_by_user_id, attested_by_membership_id, attested_at,
    request_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8, $9,
    $10
);

-- name: ProfileGetUpdatedAt :one
SELECT updated_at
FROM child_registration_profiles
WHERE tenant_id = $1 AND branch_id = $2 AND child_id = $3;


