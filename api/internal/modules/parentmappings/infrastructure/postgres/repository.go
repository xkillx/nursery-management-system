package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/parentmappings/domain"
)

type ParentMappingRepository struct {
	pool *pgxpool.Pool
}

func NewParentMappingRepository(pool *pgxpool.Pool) *ParentMappingRepository {
	return &ParentMappingRepository{pool: pool}
}

func (r *ParentMappingRepository) FindActiveByMembership(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.ParentMapping, bool, error) {
	const q = `
SELECT id, membership_id, guardian_id, ended_at, ended_reason_code::text, ended_reason_note, created_at, updated_at
FROM parent_membership_guardians
WHERE tenant_id = $1 AND branch_id = $2 AND membership_id = $3 AND ended_at IS NULL
LIMIT 1`

	var row domain.ParentMapping
	err := tx.QueryRow(ctx, q, tenantID, branchID, membershipID).Scan(
		&row.ID, &row.MembershipID, &row.GuardianID, &row.EndedAt, &row.EndedReasonCode, &row.EndedReasonNote, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ParentMapping{}, false, nil
	}
	if err != nil {
		return domain.ParentMapping{}, false, err
	}
	return row, true, nil
}

func (r *ParentMappingRepository) Create(ctx context.Context, tx pgx.Tx, mapping domain.ParentMapping) error {
	const q = `INSERT INTO parent_membership_guardians (id, tenant_id, branch_id, membership_id, guardian_id) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, q, mapping.ID, mapping.TenantID, mapping.BranchID, mapping.MembershipID, mapping.GuardianID)
	return err
}

func (r *ParentMappingRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.ParentMapping, bool, error) {
	const q = `
SELECT id, membership_id, guardian_id, ended_at, ended_reason_code::text, ended_reason_note, created_at, updated_at
FROM parent_membership_guardians
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
FOR UPDATE`

	var row domain.ParentMapping
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&row.ID, &row.MembershipID, &row.GuardianID, &row.EndedAt, &row.EndedReasonCode, &row.EndedReasonNote, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ParentMapping{}, false, nil
	}
	if err != nil {
		return domain.ParentMapping{}, false, err
	}
	return row, true, nil
}

func (r *ParentMappingRepository) End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
UPDATE parent_membership_guardians
SET ended_at = now(), ended_reason_code = $1, ended_reason_note = NULLIF($2, ''), updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, id)
	return err
}

func (r *ParentMappingRepository) GetMembershipForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.MembershipInfo, bool, error) {
	const q = `SELECT id, role, is_active FROM memberships WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`

	var info domain.MembershipInfo
	err := tx.QueryRow(ctx, q, tenantID, branchID, membershipID).Scan(&info.ID, &info.Role, &info.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MembershipInfo{}, false, nil
	}
	if err != nil {
		return domain.MembershipInfo{}, false, err
	}
	return info, true, nil
}
