package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/guardianlinks/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type GuardianChildLinkRepository struct {
	pool *pgxpool.Pool
}

func NewGuardianChildLinkRepository(pool *pgxpool.Pool) *GuardianChildLinkRepository {
	return &GuardianChildLinkRepository{pool: pool}
}

func (r *GuardianChildLinkRepository) FindActiveByPair(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID, childID uuid.UUID) (domain.GuardianChildLink, bool, error) {
	q := sqlc.New(tx)
	row, err := q.GuardianLinksFindActiveByPair(ctx, sqlc.GuardianLinksFindActiveByPairParams{
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		GuardianID: uuidToPgtype(guardianID),
		ChildID:    uuidToPgtype(childID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GuardianChildLink{}, false, nil
		}
		return domain.GuardianChildLink{}, false, err
	}
	return mapLinkRow(row), true, nil
}

func (r *GuardianChildLinkRepository) ListActiveByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) ([]domain.LinkedGuardianChildLink, error) {
	q := sqlc.New(tx)
	rows, err := q.GuardianLinksListActiveByChild(ctx, sqlc.GuardianLinksListActiveByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.LinkedGuardianChildLink, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapLinkedGuardianRow(row))
	}
	return result, nil
}

func (r *GuardianChildLinkRepository) Create(ctx context.Context, tx pgx.Tx, link domain.GuardianChildLink) error {
	q := sqlc.New(tx)
	return q.GuardianLinksCreate(ctx, sqlc.GuardianLinksCreateParams{
		ID:         uuidToPgtype(link.ID),
		TenantID:   uuidToPgtype(link.TenantID),
		BranchID:   uuidToPgtype(link.BranchID),
		GuardianID: uuidToPgtype(link.GuardianID),
		ChildID:    uuidToPgtype(link.ChildID),
	})
}

func (r *GuardianChildLinkRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.GuardianChildLink, bool, error) {
	q := sqlc.New(tx)
	row, err := q.GuardianLinksGetByIDForUpdate(ctx, sqlc.GuardianLinksGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GuardianChildLink{}, false, nil
		}
		return domain.GuardianChildLink{}, false, err
	}
	return mapLinkRowFromForUpdate(row), true, nil
}

func (r *GuardianChildLinkRepository) End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.GuardianLinksEnd(ctx, sqlc.GuardianLinksEndParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		Column2:         reasonNote,
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		ID:              uuidToPgtype(id),
	})
}

func mapLinkRow(row sqlc.GuardianLinksFindActiveByPairRow) domain.GuardianChildLink {
	return domain.GuardianChildLink{
		ID:              pgtypeUUIDToUUID(row.ID),
		GuardianID:      pgtypeUUIDToUUID(row.GuardianID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapLinkRowFromForUpdate(row sqlc.GuardianLinksGetByIDForUpdateRow) domain.GuardianChildLink {
	return domain.GuardianChildLink{
		ID:              pgtypeUUIDToUUID(row.ID),
		GuardianID:      pgtypeUUIDToUUID(row.GuardianID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapLinkedGuardianRow(row sqlc.GuardianLinksListActiveByChildRow) domain.LinkedGuardianChildLink {
	return domain.LinkedGuardianChildLink{
		ID:         pgtypeUUIDToUUID(row.ID),
		GuardianID: pgtypeUUIDToUUID(row.GuardianID),
		ChildID:    pgtypeUUIDToUUID(row.ChildID),
		Guardian: domain.LinkedGuardianSummary{
			ID:       pgtypeUUIDToUUID(row.GuardianTableID),
			FullName: row.GuardianFullName,
			Email:    pgtypeTextToStringPtr(row.GuardianEmail),
			Phone:    pgtypeTextToStringPtr(row.GuardianPhone),
			IsActive: row.GuardianIsActive,
		},
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgtypeTimestamptzToTime(row.UpdatedAt),
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

func nullLifecycleReasonCode(s string) sqlc.NullLifecycleReasonCode {
	if s == "" {
		return sqlc.NullLifecycleReasonCode{}
	}
	return sqlc.NullLifecycleReasonCode{
		LifecycleReasonCode: sqlc.LifecycleReasonCode(s),
		Valid:               true,
	}
}
