package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/parentmappings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ParentMappingRepository struct {
	pool *pgxpool.Pool
}

func NewParentMappingRepository(pool *pgxpool.Pool) *ParentMappingRepository {
	return &ParentMappingRepository{pool: pool}
}

func (r *ParentMappingRepository) FindActiveByMembership(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.ParentMapping, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ParentMappingsFindActiveByMembership(ctx, sqlc.ParentMappingsFindActiveByMembershipParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		MembershipID: uuidToPgtype(membershipID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParentMapping{}, false, nil
		}
		return domain.ParentMapping{}, false, err
	}
	return mapRow(row), true, nil
}

func (r *ParentMappingRepository) Create(ctx context.Context, tx pgx.Tx, mapping domain.ParentMapping) error {
	q := sqlc.New(tx)
	return q.ParentMappingsCreate(ctx, sqlc.ParentMappingsCreateParams{
		ID:           uuidToPgtype(mapping.ID),
		TenantID:     uuidToPgtype(mapping.TenantID),
		BranchID:     uuidToPgtype(mapping.BranchID),
		MembershipID: uuidToPgtype(mapping.MembershipID),
		GuardianID:   uuidToPgtype(mapping.GuardianID),
	})
}

func (r *ParentMappingRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.ParentMapping, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ParentMappingsGetByIDForUpdate(ctx, sqlc.ParentMappingsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParentMapping{}, false, nil
		}
		return domain.ParentMapping{}, false, err
	}
	return mapRowFromForUpdate(row), true, nil
}

func (r *ParentMappingRepository) End(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx)
	return q.ParentMappingsEnd(ctx, sqlc.ParentMappingsEndParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		Column2:         reasonNote,
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		ID:              uuidToPgtype(id),
	})
}

func (r *ParentMappingRepository) GetMembershipForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.MembershipInfo, bool, error) {
	q := sqlc.New(tx)
	row, err := q.ParentMappingsGetMembershipForScope(ctx, sqlc.ParentMappingsGetMembershipForScopeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(membershipID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MembershipInfo{}, false, nil
		}
		return domain.MembershipInfo{}, false, err
	}
	return domain.MembershipInfo{
		ID:       pgtypeUUIDToUUID(row.ID),
		Role:     row.Role,
		IsActive: row.IsActive,
	}, true, nil
}

func mapRow(row sqlc.ParentMappingsFindActiveByMembershipRow) domain.ParentMapping {
	return domain.ParentMapping{
		ID:              pgtypeUUIDToUUID(row.ID),
		MembershipID:    pgtypeUUIDToUUID(row.MembershipID),
		GuardianID:      pgtypeUUIDToUUID(row.GuardianID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapRowFromForUpdate(row sqlc.ParentMappingsGetByIDForUpdateRow) domain.ParentMapping {
	return domain.ParentMapping{
		ID:              pgtypeUUIDToUUID(row.ID),
		MembershipID:    pgtypeUUIDToUUID(row.MembershipID),
		GuardianID:      pgtypeUUIDToUUID(row.GuardianID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
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
