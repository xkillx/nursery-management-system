package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/parentchildmappings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ParentChildMappingRepository struct {
	pool *pgxpool.Pool
}

func NewParentChildMappingRepository(pool *pgxpool.Pool) *ParentChildMappingRepository {
	return &ParentChildMappingRepository{pool: pool}
}

func (r *ParentChildMappingRepository) FindActiveByPair(ctx context.Context, tx domain.Tx, tenantID, branchID, membershipID, childID uuid.UUID) (domain.ParentChildMapping, bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.ParentChildMappingsFindActiveByPair(ctx, sqlc.ParentChildMappingsFindActiveByPairParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		MembershipID: uuidToPgtype(membershipID),
		ChildID:      uuidToPgtype(childID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParentChildMapping{}, false, nil
		}
		return domain.ParentChildMapping{}, false, err
	}
	return mapFindActiveByPairRow(row), true, nil
}

func (r *ParentChildMappingRepository) ListActiveByMembership(ctx context.Context, tx domain.Tx, tenantID, branchID, membershipID uuid.UUID) ([]domain.ParentChildMapping, error) {
	q := sqlc.New(tx.(pgx.Tx))
	rows, err := q.ParentChildMappingsListActiveByMembership(ctx, sqlc.ParentChildMappingsListActiveByMembershipParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		MembershipID: uuidToPgtype(membershipID),
	})
	if err != nil {
		return nil, err
	}
	mappings := make([]domain.ParentChildMapping, 0, len(rows))
	for _, row := range rows {
		mappings = append(mappings, mapListRow(row))
	}
	return mappings, nil
}

func (r *ParentChildMappingRepository) Create(ctx context.Context, tx domain.Tx, mapping domain.ParentChildMapping) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.ParentChildMappingsCreate(ctx, sqlc.ParentChildMappingsCreateParams{
		ID:           uuidToPgtype(mapping.ID),
		TenantID:     uuidToPgtype(mapping.TenantID),
		BranchID:     uuidToPgtype(mapping.BranchID),
		MembershipID: uuidToPgtype(mapping.MembershipID),
		ChildID:      uuidToPgtype(mapping.ChildID),
	})
}

func (r *ParentChildMappingRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.ParentChildMapping, bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.ParentChildMappingsGetByIDForUpdate(ctx, sqlc.ParentChildMappingsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParentChildMapping{}, false, nil
		}
		return domain.ParentChildMapping{}, false, err
	}
	return mapGetByIDForUpdateRow(row), true, nil
}

func (r *ParentChildMappingRepository) End(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.ParentChildMappingsEnd(ctx, sqlc.ParentChildMappingsEndParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		Column2:         reasonNote,
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		ID:              uuidToPgtype(id),
	})
}

func (r *ParentChildMappingRepository) GetMembershipForScope(ctx context.Context, tx domain.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.MembershipInfo, bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.ParentChildMappingsGetMembershipForScope(ctx, sqlc.ParentChildMappingsGetMembershipForScopeParams{
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

func mapFindActiveByPairRow(row sqlc.ParentChildMappingsFindActiveByPairRow) domain.ParentChildMapping {
	return domain.ParentChildMapping{
		ID:              pgtypeUUIDToUUID(row.ID),
		MembershipID:    pgtypeUUIDToUUID(row.MembershipID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapListRow(row sqlc.ParentChildMappingsListActiveByMembershipRow) domain.ParentChildMapping {
	return domain.ParentChildMapping{
		ID:              pgtypeUUIDToUUID(row.ID),
		MembershipID:    pgtypeUUIDToUUID(row.MembershipID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapGetByIDForUpdateRow(row sqlc.ParentChildMappingsGetByIDForUpdateRow) domain.ParentChildMapping {
	return domain.ParentChildMapping{
		ID:              pgtypeUUIDToUUID(row.ID),
		MembershipID:    pgtypeUUIDToUUID(row.MembershipID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
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
