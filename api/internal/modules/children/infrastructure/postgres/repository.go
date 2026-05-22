package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/children/domain"
)

type ChildRepository struct {
	pool *pgxpool.Pool
}

func NewChildRepository(pool *pgxpool.Pool) *ChildRepository {
	return &ChildRepository{pool: pool}
}

func (r *ChildRepository) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Child, error) {
	statusClause := "AND c.is_active = true"
	switch filter {
	case domain.StatusInactive:
		statusClause = "AND c.is_active = false"
	case domain.StatusAll:
		statusClause = ""
	}

	q := fmt.Sprintf(`
	SELECT c.id,
	       c.full_name,
	       c.date_of_birth,
	       c.start_date,
	       c.end_date,
	       c.core_hourly_rate_minor,
	       c.notes,
	       c.is_active,
	       c.left_at,
	       c.left_reason_code::text,
	       c.left_reason_note,
	       EXISTS (
	           SELECT 1
	           FROM guardian_child_links gcl
	           WHERE gcl.tenant_id = c.tenant_id
	             AND gcl.branch_id = c.branch_id
	             AND gcl.child_id = c.id
	             AND gcl.ended_at IS NULL
	       ) AS has_guardian_link,
	       c.created_at,
	       c.updated_at
	FROM children c
	WHERE c.tenant_id = $1
	  AND c.branch_id = $2
	  %s
	ORDER BY c.updated_at DESC
	LIMIT $3 OFFSET $4`, statusClause)

	rows, err := r.pool.Query(ctx, q, tenantID, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query children: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Child, 0)
	for rows.Next() {
		var child domain.Child
		if err := rows.Scan(
			&child.ID,
			&child.FullName,
			&child.DateOfBirth,
			&child.StartDate,
			&child.EndDate,
			&child.CoreHourlyRateMinor,
			&child.Notes,
			&child.IsActive,
			&child.LeftAt,
			&child.LeftReasonCode,
			&child.LeftReasonNote,
			&child.HasGuardianLink,
			&child.CreatedAt,
			&child.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan child row: %w", err)
		}
		out = append(out, child)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate child rows: %w", err)
	}

	return out, nil
}

func (r *ChildRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	const q = `
	SELECT c.id,
	       c.full_name,
	       c.date_of_birth,
	       c.start_date,
	       c.end_date,
	       c.core_hourly_rate_minor,
	       c.notes,
	       c.is_active,
	       c.left_at,
	       c.left_reason_code::text,
	       c.left_reason_note,
	       EXISTS (
	           SELECT 1
	           FROM guardian_child_links gcl
	           WHERE gcl.tenant_id = c.tenant_id
	             AND gcl.branch_id = c.branch_id
	             AND gcl.child_id = c.id
	             AND gcl.ended_at IS NULL
	       ) AS has_guardian_link,
	       c.created_at,
	       c.updated_at
	FROM children c
	WHERE c.tenant_id = $1
	  AND c.branch_id = $2
	  AND c.id = $3`

	var child domain.Child
	err := r.pool.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&child.ID,
		&child.FullName,
		&child.DateOfBirth,
		&child.StartDate,
		&child.EndDate,
		&child.CoreHourlyRateMinor,
		&child.Notes,
		&child.IsActive,
		&child.LeftAt,
		&child.LeftReasonCode,
		&child.LeftReasonNote,
		&child.HasGuardianLink,
		&child.CreatedAt,
		&child.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child by id: %w", err)
	}

	return child, true, nil
}

func (r *ChildRepository) Create(ctx context.Context, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	const q = `
	INSERT INTO children (
	    id, tenant_id, branch_id, full_name, date_of_birth, start_date, end_date,
	    core_hourly_rate_minor, notes, is_active
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), true)`

	_, err := r.pool.Exec(ctx, q,
		child.ID,
		tenantID,
		branchID,
		child.FullName,
		child.DateOfBirth,
		child.StartDate,
		child.EndDate,
		child.CoreHourlyRateMinor,
		notes,
	)
	if err != nil {
		return fmt.Errorf("insert child: %w", err)
	}

	return nil
}

func (r *ChildRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	setParts := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+3)
	args = append(args, tenantID, branchID, id)
	argPos := 4

	for _, col := range orderedColumns(fields) {
		val := fields[col]
		switch col {
		case "notes":
			setParts = append(setParts, fmt.Sprintf("notes = NULLIF($%d, '')", argPos))
		default:
			setParts = append(setParts, fmt.Sprintf("%s = $%d", col, argPos))
		}
		args = append(args, val)
		argPos++
	}

	q := fmt.Sprintf(`
	UPDATE children
	SET %s, updated_at = now()
	WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`, strings.Join(setParts, ", "))

	ct, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("update child: %w", err)
	}

	return ct.RowsAffected(), nil
}

func (r *ChildRepository) MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	const q = `
	UPDATE children
	SET is_active = false,
	    left_at = now(),
	    left_reason_code = $1,
	    left_reason_note = NULLIF($2, ''),
	    updated_at = now()
	WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`

	_, err := tx.Exec(ctx, q, reasonCode, reasonNote, tenantID, branchID, id)
	if err != nil {
		return fmt.Errorf("mark child inactive: %w", err)
	}

	return nil
}

func (r *ChildRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	const q = `
	SELECT c.id,
	       c.full_name,
	       c.date_of_birth,
	       c.start_date,
	       c.end_date,
	       c.core_hourly_rate_minor,
	       c.notes,
	       c.is_active,
	       c.left_at,
	       c.left_reason_code::text,
	       c.left_reason_note,
	       EXISTS (
	           SELECT 1
	           FROM guardian_child_links gcl
	           WHERE gcl.tenant_id = c.tenant_id
	             AND gcl.branch_id = c.branch_id
	             AND gcl.child_id = c.id
	             AND gcl.ended_at IS NULL
	       ) AS has_guardian_link,
	       c.created_at,
	       c.updated_at
	FROM children c
	WHERE c.tenant_id = $1
	  AND c.branch_id = $2
	  AND c.id = $3
	FOR UPDATE`

	var child domain.Child
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(
		&child.ID,
		&child.FullName,
		&child.DateOfBirth,
		&child.StartDate,
		&child.EndDate,
		&child.CoreHourlyRateMinor,
		&child.Notes,
		&child.IsActive,
		&child.LeftAt,
		&child.LeftReasonCode,
		&child.LeftReasonNote,
		&child.HasGuardianLink,
		&child.CreatedAt,
		&child.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Child{}, false, nil
	}
	if err != nil {
		return domain.Child{}, false, fmt.Errorf("query child for update: %w", err)
	}

	return child, true, nil
}

func (r *ChildRepository) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	const q = `
	SELECT EXISTS (
	  SELECT 1 FROM children WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
	)`
	var exists bool
	err := tx.QueryRow(ctx, q, tenantID, branchID, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check child exists in scope: %w", err)
	}
	return exists, nil
}

func (r *ChildRepository) ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID) ([]domain.AttendanceChild, error) {
	const q = `
	SELECT c.id,
	       c.full_name,
	       EXISTS (
	           SELECT 1
	           FROM guardian_child_links gcl
	           WHERE gcl.tenant_id = c.tenant_id
	             AND gcl.branch_id = c.branch_id
	             AND gcl.child_id = c.id
	             AND gcl.ended_at IS NULL
	       ) AS has_guardian_link
	FROM children c
	WHERE c.tenant_id = $1
	  AND c.branch_id = $2
	  AND c.is_active = true
	ORDER BY c.updated_at DESC`

	rows, err := r.pool.Query(ctx, q, tenantID, branchID)
	if err != nil {
		return nil, fmt.Errorf("query attendance children: %w", err)
	}
	defer rows.Close()

	out := make([]domain.AttendanceChild, 0)
	for rows.Next() {
		var child domain.AttendanceChild
		if err := rows.Scan(&child.ID, &child.FullName, &child.EnrollmentComplete); err != nil {
			return nil, fmt.Errorf("scan attendance child row: %w", err)
		}
		out = append(out, child)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendance rows: %w", err)
	}

	return out, nil
}

// orderedColumns returns column names in a deterministic order for UPDATE statements.
func orderedColumns(fields map[string]any) []string {
	order := []string{"full_name", "date_of_birth", "start_date", "end_date", "core_hourly_rate_minor", "notes"}
	result := make([]string, 0, len(fields))
	for _, col := range order {
		if _, ok := fields[col]; ok {
			result = append(result, col)
		}
	}
	return result
}
