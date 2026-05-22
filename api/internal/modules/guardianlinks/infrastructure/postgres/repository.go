package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
)

type GuardianChildLinkRepository struct {
	pool *pgxpool.Pool
}

func NewGuardianChildLinkRepository(pool *pgxpool.Pool) *GuardianChildLinkRepository {
	return &GuardianChildLinkRepository{pool: pool}
}

func (r *GuardianChildLinkRepository) FindActiveByPair(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID, childID uuid.UUID) (domain.GuardianChildLink, bool, error) {
	const q = `
SELECT id, guardian_id, child_id, ended_at, ended_reason_code::text, ended_reason_note, created_at, updated_at
FROM guardian_child_links
WHERE tenant_id = $1 AND branch_id = $2 AND guardian_id = $3 AND child_id = $4 AND ended_at IS NULL
LIMIT 1`

	var row domain.GuardianChildLink
	err := tx.QueryRow(ctx, q, tenantID, branchID, guardianID, childID).Scan(
		&row.ID, &row.GuardianID, &row.ChildID, &row.EndedAt, &row.EndedReasonCode, &row.EndedReasonNote, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.GuardianChildLink{}, false, nil
	}
	if err != nil {
		return domain.GuardianChildLink{}, false, err
	}
	return row, true, nil
}

func (r *GuardianChildLinkRepository) Create(ctx context.Context, tx pgx.Tx, link domain.GuardianChildLink) error {
	const q = `INSERT INTO guardian_child_links (id, tenant_id, branch_id, guardian_id, child_id) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, q, link.ID, link.TenantID, link.BranchID, link.GuardianID, link.ChildID)
	return err
}

func (r *GuardianChildLinkRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.GuardianChildLink, bool, error) {
	const q = `
SELECT id, guardian_id, child_id, ended_at, ended_reason_code::text, ended_reason_note, created_at, updated_at
FROM guardian_child_links
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
FOR UPDATE`

	var row domain.GuardianChildLink
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&row.ID, &row.GuardianID, &row.ChildID, &row.EndedAt, &row.EndedReasonCode, &row.EndedReasonNote, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.GuardianChildLink{}, false, nil
	}
	if err != nil {
		return domain.GuardianChildLink{}, false, err
	}
	return row, true, nil
}

func (r *GuardianChildLinkRepository) End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
UPDATE guardian_child_links
SET ended_at = now(), ended_reason_code = $1, ended_reason_note = NULLIF($2, ''), updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, id)
	return err
}
