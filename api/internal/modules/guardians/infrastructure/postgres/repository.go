package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardians/domain"
)

type GuardianRepository struct {
	pool *pgxpool.Pool
}

func NewGuardianRepository(pool *pgxpool.Pool) *GuardianRepository {
	return &GuardianRepository{pool: pool}
}

func (r *GuardianRepository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *GuardianRepository) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Guardian, error) {
	statusClause := "AND g.is_active = true"
	if filter == domain.StatusInactive {
		statusClause = "AND g.is_active = false"
	}
	if filter == domain.StatusAll {
		statusClause = ""
	}

	q := fmt.Sprintf(`
	SELECT g.id,
	       g.full_name,
	       g.email,
	       g.phone,
	       g.notes,
	       g.is_active,
	       g.deactivated_at,
	       g.deactivation_reason_code::text,
	       g.deactivation_reason_note,
	       g.created_at,
	       g.updated_at
	FROM guardians g
	WHERE g.tenant_id = $1
	  AND g.branch_id = $2
	  %s
	ORDER BY g.updated_at DESC
	LIMIT $3 OFFSET $4`, statusClause)

	rows, err := r.pool.Query(ctx, q, tenantID, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list guardians: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Guardian, 0)
	for rows.Next() {
		var g domain.Guardian
		if err := rows.Scan(
			&g.ID,
			&g.FullName,
			&g.Email,
			&g.Phone,
			&g.Notes,
			&g.IsActive,
			&g.DeactivatedAt,
			&g.DeactivationReasonCode,
			&g.DeactivationReasonNote,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan guardian row: %w", err)
		}
		out = append(out, g)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guardian rows: %w", err)
	}

	return out, nil
}

func (r *GuardianRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Guardian, error) {
	const q = `
	SELECT g.id,
	       g.full_name,
	       g.email,
	       g.phone,
	       g.notes,
	       g.is_active,
	       g.deactivated_at,
	       g.deactivation_reason_code::text,
	       g.deactivation_reason_note,
	       g.created_at,
	       g.updated_at
	FROM guardians g
	WHERE g.tenant_id = $1
	  AND g.branch_id = $2
	  AND g.id = $3`

	var g domain.Guardian
	err := r.pool.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&g.ID,
		&g.FullName,
		&g.Email,
		&g.Phone,
		&g.Notes,
		&g.IsActive,
		&g.DeactivatedAt,
		&g.DeactivationReasonCode,
		&g.DeactivationReasonNote,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Guardian{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("get guardian by id: %w", err)
	}

	return g, nil
}

func (r *GuardianRepository) Create(ctx context.Context, guardian domain.Guardian) error {
	const q = `
	INSERT INTO guardians (id, tenant_id, branch_id, full_name, email, phone, notes, is_active)
	VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), true)`

	_, err := r.pool.Exec(ctx, q,
		guardian.ID,
		guardian.TenantID,
		guardian.BranchID,
		guardian.FullName,
		derefStr(guardian.Email),
		derefStr(guardian.Phone),
		derefStr(guardian.Notes),
	)
	if err != nil {
		return fmt.Errorf("insert guardian: %w", err)
	}

	return nil
}

func (r *GuardianRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+3)
	args = append(args, tenantID, branchID, id)
	argPos := 4

	for name, value := range fields {
		if name == "email" || name == "phone" || name == "notes" {
			setClauses = append(setClauses, fmt.Sprintf("%s = NULLIF($%d, '')", name, argPos))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", name, argPos))
		}
		args = append(args, value)
		argPos++
	}

	q := fmt.Sprintf(`
	UPDATE guardians
	SET %s, updated_at = now()
	WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`, strings.Join(setClauses, ", "))

	ct, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("update guardian: %w", err)
	}

	return ct.RowsAffected(), nil
}

func (r *GuardianRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Guardian, error) {
	const q = `
	SELECT g.id,
	       g.full_name,
	       g.email,
	       g.phone,
	       g.notes,
	       g.is_active,
	       g.deactivated_at,
	       g.deactivation_reason_code::text,
	       g.deactivation_reason_note,
	       g.created_at,
	       g.updated_at
	FROM guardians g
	WHERE g.tenant_id = $1
	  AND g.branch_id = $2
	  AND g.id = $3
	FOR UPDATE`

	var g domain.Guardian
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&g.ID,
		&g.FullName,
		&g.Email,
		&g.Phone,
		&g.Notes,
		&g.IsActive,
		&g.DeactivatedAt,
		&g.DeactivationReasonCode,
		&g.DeactivationReasonNote,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Guardian{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("get guardian for update: %w", err)
	}

	return g, nil
}

func (r *GuardianRepository) GetActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, bool, error) {
	const q = `
	SELECT is_active
	FROM guardians
	WHERE tenant_id = $1
	  AND branch_id = $2
	  AND id = $3`

	var isActive bool
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(&isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("get guardian active: %w", err)
	}

	return isActive, true, nil
}

func (r *GuardianRepository) Deactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
	UPDATE guardians
	SET is_active = false,
	    deactivated_at = now(),
	    deactivation_reason_code = $1,
	    deactivation_reason_note = NULLIF($2, ''),
	    updated_at = now()
	WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`

	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, id)
	if err != nil {
		return fmt.Errorf("deactivate guardian: %w", err)
	}

	return nil
}

func (r *GuardianRepository) CascadeLinks(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
	UPDATE guardian_child_links
	SET ended_at = now(),
	    ended_reason_code = $1,
	    ended_reason_note = $2,
	    updated_at = now()
	WHERE tenant_id = $3
	  AND branch_id = $4
	  AND guardian_id = $5
	  AND ended_at IS NULL`

	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, guardianID)
	if err != nil {
		return fmt.Errorf("cascade guardian child links: %w", err)
	}

	return nil
}

func (r *GuardianRepository) CascadeMappings(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
	UPDATE parent_membership_guardians
	SET ended_at = now(),
	    ended_reason_code = $1,
	    ended_reason_note = $2,
	    updated_at = now()
	WHERE tenant_id = $3
	  AND branch_id = $4
	  AND guardian_id = $5
	  AND ended_at IS NULL`

	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, guardianID)
	if err != nil {
		return fmt.Errorf("cascade parent membership guardians: %w", err)
	}

	return nil
}

func (r *GuardianRepository) Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	const q = `
	UPDATE guardians
	SET is_active = true,
	    deactivated_at = NULL,
	    deactivation_reason_code = NULL,
	    deactivation_reason_note = NULL,
	    updated_at = now()
	WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`

	_, err := tx.Exec(ctx, q, tenantID, branchID, id)
	if err != nil {
		return fmt.Errorf("reactivate guardian: %w", err)
	}

	return nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
