package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type ParentRepository struct {
	pool *pgxpool.Pool
}

func NewParentRepository(pool *pgxpool.Pool) *ParentRepository {
	return &ParentRepository{pool: pool}
}

func (r *ParentRepository) List(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, limit, offset int) ([]domain.Parent, error) {
	q := r.query(tx)
	rows, err := q.ParentsList(ctx, sqlc.ParentsListParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	return mapParentRows(rows), nil
}

func (r *ParentRepository) ListFiltered(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, isActive *bool, search *string, limit, offset int) ([]domain.Parent, int, error) {
	q := r.query(tx)

	var isActiveParam bool
	var searchParam string
	if isActive != nil {
		isActiveParam = *isActive
	}
	if search != nil {
		searchParam = *search
	}

	rows, err := q.ParentsListFiltered(ctx, sqlc.ParentsListFilteredParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  isActiveParam,
		Column4:  searchParam,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := q.ParentsCount(ctx, sqlc.ParentsCountParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  isActiveParam,
		Column4:  searchParam,
	})
	if err != nil {
		return nil, 0, err
	}

	return mapParentRows(rows), int(count), nil
}

func (r *ParentRepository) GetByID(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.Parent, bool, error) {
	q := r.query(tx)
	row, err := q.ParentsGetByID(ctx, sqlc.ParentsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Parent{}, false, nil
		}
		return domain.Parent{}, false, err
	}
	return mapParentRow(row), true, nil
}

func (r *ParentRepository) GetByUserID(ctx context.Context, tx domain.Tx, tenantID uuid.UUID, userID uuid.UUID) (domain.Parent, bool, error) {
	q := r.query(tx)
	row, err := q.ParentsGetByUserID(ctx, sqlc.ParentsGetByUserIDParams{
		TenantID: uuidToPgtype(tenantID),
		UserID:   uuidToPgtype(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Parent{}, false, nil
		}
		return domain.Parent{}, false, err
	}
	return mapParentRow(row), true, nil
}

func (r *ParentRepository) Create(ctx context.Context, tx domain.Tx, parent domain.Parent) error {
	q := r.query(tx)
	return q.ParentsCreate(ctx, sqlc.ParentsCreateParams{
		ID:                        uuidToPgtype(parent.ID),
		TenantID:                  uuidToPgtype(parent.TenantID),
		BranchID:                  uuidToPgtype(parent.BranchID),
		FirstName:                 parent.FirstName,
		LastName:                  stringPtrToPgtypeText(parent.LastName),
		Email:                     stringPtrToPgtypeText(parent.Email),
		Phone:                     stringPtrToPgtypeText(parent.Phone),
		AddressLine1:              stringPtrToPgtypeText(parent.AddressLine1),
		AddressLine2:              stringPtrToPgtypeText(parent.AddressLine2),
		AddressCity:               stringPtrToPgtypeText(parent.AddressCity),
		AddressPostcode:           stringPtrToPgtypeText(parent.AddressPostcode),
		RelationshipToChild:       stringPtrToPgtypeText(parent.RelationshipToChild),
		HasParentalResponsibility: parent.HasParentalResponsibility,
		CanPickUp:                 parent.CanPickUp,
		IsEmergencyContact:        parent.IsEmergencyContact,
		Notes:                     stringPtrToPgtypeText(parent.Notes),
		UserID:                    uuidPtrToPgtypeUUID(parent.UserID),
		IsActive:                  parent.IsActive,
	})
}

func (r *ParentRepository) Update(ctx context.Context, tx domain.Tx, parent domain.Parent) error {
	q := r.query(tx)
	return q.ParentsUpdate(ctx, sqlc.ParentsUpdateParams{
		TenantID:                  uuidToPgtype(parent.TenantID),
		BranchID:                  uuidToPgtype(parent.BranchID),
		ID:                        uuidToPgtype(parent.ID),
		FirstName:                 parent.FirstName,
		LastName:                  stringPtrToPgtypeText(parent.LastName),
		Email:                     stringPtrToPgtypeText(parent.Email),
		Phone:                     stringPtrToPgtypeText(parent.Phone),
		AddressLine1:              stringPtrToPgtypeText(parent.AddressLine1),
		AddressLine2:              stringPtrToPgtypeText(parent.AddressLine2),
		AddressCity:               stringPtrToPgtypeText(parent.AddressCity),
		AddressPostcode:           stringPtrToPgtypeText(parent.AddressPostcode),
		RelationshipToChild:       stringPtrToPgtypeText(parent.RelationshipToChild),
		HasParentalResponsibility: parent.HasParentalResponsibility,
		CanPickUp:                 parent.CanPickUp,
		IsEmergencyContact:        parent.IsEmergencyContact,
		Notes:                     stringPtrToPgtypeText(parent.Notes),
		UserID:                    uuidPtrToPgtypeUUID(parent.UserID),
		IsActive:                  parent.IsActive,
	})
}

func (r *ParentRepository) SoftDelete(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := r.query(tx)
	return q.ParentsSoftDelete(ctx, sqlc.ParentsSoftDeleteParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *ParentRepository) SetUserID(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, userID *uuid.UUID) error {
	q := r.query(tx)
	return q.ParentsSetUserID(ctx, sqlc.ParentsSetUserIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
		UserID:   uuidPtrToPgtypeUUID(userID),
	})
}

func (r *ParentRepository) ListChildrenByParent(ctx context.Context, tx domain.Tx, tenantID, branchID, parentID uuid.UUID) ([]domain.ParentChild, error) {
	q := r.query(tx)
	rows, err := q.ParentChildrenListByParent(ctx, sqlc.ParentChildrenListByParentParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ParentID: uuidToPgtype(parentID),
	})
	if err != nil {
		return nil, err
	}
	return mapParentChildRows(rows), nil
}

func (r *ParentRepository) ListParentsByChild(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) ([]domain.ParentChild, error) {
	q := r.query(tx)
	rows, err := q.ParentChildrenListByChild(ctx, sqlc.ParentChildrenListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, err
	}
	return mapParentChildRowsFromChild(rows), nil
}

func (r *ParentRepository) FindActiveByPair(ctx context.Context, tx domain.Tx, tenantID, branchID, parentID, childID uuid.UUID) (domain.ParentChild, bool, error) {
	q := r.query(tx)
	row, err := q.ParentChildrenFindActiveByPair(ctx, sqlc.ParentChildrenFindActiveByPairParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ParentID: uuidToPgtype(parentID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParentChild{}, false, nil
		}
		return domain.ParentChild{}, false, err
	}
	return mapParentChildRow(row), true, nil
}

func (r *ParentRepository) CreateLink(ctx context.Context, tx domain.Tx, link domain.ParentChild) error {
	q := r.query(tx)
	return q.ParentChildrenCreate(ctx, sqlc.ParentChildrenCreateParams{
		ID:       uuidToPgtype(link.ID),
		TenantID: uuidToPgtype(link.TenantID),
		BranchID: uuidToPgtype(link.BranchID),
		ParentID: uuidToPgtype(link.ParentID),
		ChildID:  uuidToPgtype(link.ChildID),
	})
}

func (r *ParentRepository) EndLink(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	q := r.query(tx)
	return q.ParentChildrenEnd(ctx, sqlc.ParentChildrenEndParams{
		EndedReasonCode: nullLifecycleReasonCode(reasonCode),
		Column2:         reasonNote,
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		ID:              uuidToPgtype(id),
	})
}

func (r *ParentRepository) HasActiveParentForChild(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	q := r.query(tx)
	return q.ParentChildrenHasActiveForChild(ctx, sqlc.ParentChildrenHasActiveForChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
}

func (r *ParentRepository) query(tx domain.Tx) *sqlc.Queries {
	if tx != nil {
		return sqlc.New(tx.(pgx.Tx))
	}
	return sqlc.New(r.pool)
}

// --- Mapping helpers ---

func mapParentRow(row sqlc.Parent) domain.Parent {
	return domain.Parent{
		ID:                      pgtypeUUIDToUUID(row.ID),
		TenantID:                pgtypeUUIDToUUID(row.TenantID),
		BranchID:                pgtypeUUIDToUUID(row.BranchID),
		FirstName:               row.FirstName,
		LastName:                pgtypeTextToStringPtr(row.LastName),
		Email:                   pgtypeTextToStringPtr(row.Email),
		Phone:                   pgtypeTextToStringPtr(row.Phone),
		AddressLine1:            pgtypeTextToStringPtr(row.AddressLine1),
		AddressLine2:            pgtypeTextToStringPtr(row.AddressLine2),
		AddressCity:             pgtypeTextToStringPtr(row.AddressCity),
		AddressPostcode:         pgtypeTextToStringPtr(row.AddressPostcode),
		RelationshipToChild:     pgtypeTextToStringPtr(row.RelationshipToChild),
		HasParentalResponsibility: row.HasParentalResponsibility,
		CanPickUp:               row.CanPickUp,
		IsEmergencyContact:      row.IsEmergencyContact,
		Notes:                   pgtypeTextToStringPtr(row.Notes),
		UserID:                  pgtypeUUIDToUUIDPtr(row.UserID),
		IsActive:                row.IsActive,
		CreatedAt:               pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:               pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapParentRows(rows []sqlc.Parent) []domain.Parent {
	result := make([]domain.Parent, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapParentRow(row))
	}
	return result
}

func mapParentChildRow(row sqlc.ParentChildrenFindActiveByPairRow) domain.ParentChild {
	return domain.ParentChild{
		ID:              pgtypeUUIDToUUID(row.ID),
		TenantID:        pgtypeUUIDToUUID(row.TenantID),
		BranchID:        pgtypeUUIDToUUID(row.BranchID),
		ParentID:        pgtypeUUIDToUUID(row.ParentID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
		EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
		EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapParentChildRows(rows []sqlc.ParentChildrenListByParentRow) []domain.ParentChild {
	result := make([]domain.ParentChild, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ParentChild{
			ID:              pgtypeUUIDToUUID(row.ID),
			TenantID:        pgtypeUUIDToUUID(row.TenantID),
			BranchID:        pgtypeUUIDToUUID(row.BranchID),
			ParentID:        pgtypeUUIDToUUID(row.ParentID),
			ChildID:         pgtypeUUIDToUUID(row.ChildID),
			EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
			EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
			EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
			CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
			UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
		})
	}
	return result
}

func mapParentChildRowsFromChild(rows []sqlc.ParentChildrenListByChildRow) []domain.ParentChild {
	result := make([]domain.ParentChild, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ParentChild{
			ID:              pgtypeUUIDToUUID(row.ID),
			TenantID:        pgtypeUUIDToUUID(row.TenantID),
			BranchID:        pgtypeUUIDToUUID(row.BranchID),
			ParentID:        pgtypeUUIDToUUID(row.ParentID),
			ChildID:         pgtypeUUIDToUUID(row.ChildID),
			EndedAt:         pgtypeTimestamptzToTimePtr(row.EndedAt),
			EndedReasonCode: ifaceToStringPtr(row.EndedReasonCode),
			EndedReasonNote: pgtypeTextToStringPtr(row.EndedReasonNote),
			CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
			UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
		})
	}
	return result
}

// --- pgtype helpers ---

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func pgtypeUUIDToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes)
	return &v
}

func uuidPtrToPgtypeUUID(p *uuid.UUID) pgtype.UUID {
	if p == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: [16]byte(*p), Valid: true}
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

func stringPtrToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
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
