package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardians/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
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
	q := sqlc.New(r.pool)
	rows, err := q.GuardiansList(ctx, sqlc.GuardiansListParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		Limit:        int32(limit),
		Offset:       int32(offset),
		StatusFilter: string(filter),
	})
	if err != nil {
		return nil, fmt.Errorf("list guardians: %w", err)
	}
	out := make([]domain.Guardian, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGuardianRow(row.ID, row.FullName, row.Email, row.Phone, row.Notes,
			row.IsActive, row.DeactivatedAt, row.DeactivationReasonCode, row.DeactivationReasonNote,
			row.CreatedAt, row.UpdatedAt))
	}
	return out, nil
}

func (r *GuardianRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Guardian, error) {
	q := sqlc.New(r.pool)
	row, err := q.GuardiansGetByID(ctx, sqlc.GuardiansGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Guardian{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("get guardian by id: %w", err)
	}
	return mapGuardianRow(row.ID, row.FullName, row.Email, row.Phone, row.Notes,
		row.IsActive, row.DeactivatedAt, row.DeactivationReasonCode, row.DeactivationReasonNote,
		row.CreatedAt, row.UpdatedAt), nil
}

func (r *GuardianRepository) Create(ctx context.Context, guardian domain.Guardian) error {
	q := sqlc.New(r.pool)
	return q.GuardiansCreate(ctx, sqlc.GuardiansCreateParams{
		ID:       uuidToPgtype(guardian.ID),
		TenantID: uuidToPgtype(guardian.TenantID),
		BranchID: uuidToPgtype(guardian.BranchID),
		FullName: guardian.FullName,
		Column5:  derefStr(guardian.Email),
		Column6:  derefStr(guardian.Phone),
		Column7:  derefStr(guardian.Notes),
	})
}

func (r *GuardianRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.GuardiansUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["full_name"]; ok {
		params.SetFullName = int32(1)
		params.FullName = v.(string)
	}
	if v, ok := fields["email"]; ok {
		params.SetEmail = int32(1)
		s, _ := v.(string)
		params.Email = s
	}
	if v, ok := fields["phone"]; ok {
		params.SetPhone = int32(1)
		s, _ := v.(string)
		params.Phone = s
	}
	if v, ok := fields["notes"]; ok {
		params.SetNotes = int32(1)
		s, _ := v.(string)
		params.Notes = s
	}

	q := sqlc.New(r.pool)
	ct, err := q.GuardiansUpdate(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("update guardian: %w", err)
	}
	return ct, nil
}

func (r *GuardianRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Guardian, error) {
	q := sqlc.New(tx)
	row, err := q.GuardiansGetByIDForUpdate(ctx, sqlc.GuardiansGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Guardian{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Guardian{}, fmt.Errorf("get guardian for update: %w", err)
	}
	return mapGuardianRow(row.ID, row.FullName, row.Email, row.Phone, row.Notes,
		row.IsActive, row.DeactivatedAt, row.DeactivationReasonCode, row.DeactivationReasonNote,
		row.CreatedAt, row.UpdatedAt), nil
}

func (r *GuardianRepository) GetActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, bool, error) {
	q := sqlc.New(tx)
	isActive, err := q.GuardiansGetActive(ctx, sqlc.GuardiansGetActiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("get guardian active: %w", err)
	}
	return isActive, true, nil
}

func (r *GuardianRepository) Deactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.GuardiansDeactivate(ctx, sqlc.GuardiansDeactivateParams{
		DeactivationReasonCode: nullLifecycleReasonCode(reasonCode),
		Column2:                reasonNote,
		TenantID:               uuidToPgtype(tenantID),
		BranchID:               uuidToPgtype(branchID),
		ID:                     uuidToPgtype(id),
	})
}

func (r *GuardianRepository) CascadeLinks(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.GuardiansCascadeLinks(ctx, sqlc.GuardiansCascadeLinksParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		EndedReasonNote: stringToPgtypeText(reasonNote),
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		GuardianID:      uuidToPgtype(guardianID),
	})
}

func (r *GuardianRepository) CascadeMappings(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.GuardiansCascadeMappings(ctx, sqlc.GuardiansCascadeMappingsParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		EndedReasonNote: stringToPgtypeText(reasonNote),
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		GuardianID:      uuidToPgtype(guardianID),
	})
}

func (r *GuardianRepository) Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx)
	return q.GuardiansReactivate(ctx, sqlc.GuardiansReactivateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func mapGuardianRow(
	id pgtype.UUID, fullName string, email, phone, notes pgtype.Text,
	isActive bool, deactivatedAt pgtype.Timestamptz,
	deactivationReasonCode interface{}, deactivationReasonNote pgtype.Text,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.Guardian {
	return domain.Guardian{
		ID:                     pgtypeUUIDToUUID(id),
		FullName:               fullName,
		Email:                  pgtypeTextToStringPtr(email),
		Phone:                  pgtypeTextToStringPtr(phone),
		Notes:                  pgtypeTextToStringPtr(notes),
		IsActive:               isActive,
		DeactivatedAt:          pgtypeTimestamptzToTimePtr(deactivatedAt),
		DeactivationReasonCode: ifaceToStringPtr(deactivationReasonCode),
		DeactivationReasonNote: pgtypeTextToStringPtr(deactivationReasonNote),
		CreatedAt:              pgtypeTimestamptzToTime(createdAt),
		UpdatedAt:              pgtypeTimestamptzToTime(updatedAt),
	}
}

func ifaceToStringPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

func nullLifecycleReasonCode(s string) sqlc.NullLifecycleReasonCode {
	if s == "" {
		return sqlc.NullLifecycleReasonCode{}
	}
	return sqlc.NullLifecycleReasonCode{
		LifecycleReasonCode: sqlc.LifecycleReasonCode(s),
		Valid:               true,
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func stringToPgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
