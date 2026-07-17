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

	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type SessionTypeRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SessionTypeRepository {
	return &SessionTypeRepository{pool: pool}
}

func (r *SessionTypeRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.SessionType, error) {
	q := sqlc.New(r.pool)
	rows, err := q.SessionTypesListByBranch(ctx, sqlc.SessionTypesListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("query session types list: %w", err)
	}
	out := make([]domain.SessionType, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTypeFromListRow(row))
	}
	return out, nil
}

func (r *SessionTypeRepository) ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int) ([]domain.SessionType, error) {
	q := sqlc.New(r.pool)
	rows, err := q.SessionTypesListByBranchPaginated(ctx, sqlc.SessionTypesListByBranchPaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("query session types list paginated: %w", err)
	}
	out := make([]domain.SessionType, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSessionTypeFromListPaginatedRow(row))
	}
	return out, nil
}

func (r *SessionTypeRepository) ListByBranchPaginatedSorted(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool, limit, offset int, sortField, sortDir string) ([]domain.SessionType, error) {
	q := sqlc.New(r.pool)
	pgTenant := uuidToPgtype(tenantID)
	pgBranch := uuidToPgtype(branchID)
	col3 := !includeArchived
	pgLimit := pgtype.Int4{Int32: int32(limit), Valid: true}
	pgOffset := pgtype.Int4{Int32: int32(offset), Valid: true}

	switch sortField + ":" + sortDir {
	case "name:desc":
		rows, err := q.SessionTypesListByBranchPaginatedSortByNameDesc(ctx, sqlc.SessionTypesListByBranchPaginatedSortByNameDescParams{
			TenantID: pgTenant, BranchID: pgBranch, Column3: col3, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, fmt.Errorf("query session types sorted: %w", err)
		}
		out := make([]domain.SessionType, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapSessionTypeSortRow(row))
		}
		return out, nil
	case "created_at:asc":
		rows, err := q.SessionTypesListByBranchPaginatedSortByCreatedAtAsc(ctx, sqlc.SessionTypesListByBranchPaginatedSortByCreatedAtAscParams{
			TenantID: pgTenant, BranchID: pgBranch, Column3: col3, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, fmt.Errorf("query session types sorted: %w", err)
		}
		out := make([]domain.SessionType, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapSessionTypeSortRow(row))
		}
		return out, nil
	case "created_at:desc":
		rows, err := q.SessionTypesListByBranchPaginatedSortByCreatedAtDesc(ctx, sqlc.SessionTypesListByBranchPaginatedSortByCreatedAtDescParams{
			TenantID: pgTenant, BranchID: pgBranch, Column3: col3, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, fmt.Errorf("query session types sorted: %w", err)
		}
		out := make([]domain.SessionType, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapSessionTypeSortRow(row))
		}
		return out, nil
	default:
		return r.ListByBranchPaginated(ctx, tenantID, branchID, includeArchived, limit, offset)
	}
}

func (r *SessionTypeRepository) CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.SessionTypesCountByBranch(ctx, sqlc.SessionTypesCountByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return 0, fmt.Errorf("query session types count: %w", err)
	}
	return int(count), nil
}

func (r *SessionTypeRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.SessionType, error) {
	q := sqlc.New(r.pool)
	row, err := q.SessionTypesGetByID(ctx, sqlc.SessionTypesGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.SessionType{}, domainerrors.NotFound("session_type", "Session type not found.")
	}
	if err != nil {
		return domain.SessionType{}, fmt.Errorf("query session type by id: %w", err)
	}
	return mapSessionTypeFromGetRow(row), nil
}

func (r *SessionTypeRepository) Create(ctx context.Context, st domain.SessionType) error {
	q := sqlc.New(r.pool)
	return q.SessionTypesCreate(ctx, sqlc.SessionTypesCreateParams{
		ID:        uuidToPgtype(st.ID),
		TenantID:  uuidToPgtype(st.TenantID),
		BranchID:  uuidToPgtype(st.BranchID),
		Name:      st.Name,
		StartTime: minutesToPgtypeTime(st.StartMinutes),
		EndTime:   minutesToPgtypeTime(st.EndMinutes),
	})
}

func (r *SessionTypeRepository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.SessionTypesUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}

	if v, ok := fields["name"]; ok {
		params.SetName = int32(1)
		params.Name = v.(string)
	}
	if v, ok := fields["start_time"]; ok {
		params.SetStartTime = int32(1)
		params.StartTime = minutesToPgtypeTime(v.(int))
	}
	if v, ok := fields["end_time"]; ok {
		params.SetEndTime = int32(1)
		params.EndTime = minutesToPgtypeTime(v.(int))
	}

	q := sqlc.New(r.pool)
	return q.SessionTypesUpdate(ctx, params)
}

func (r *SessionTypeRepository) Archive(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTypesArchive(ctx, sqlc.SessionTypesArchiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTypeRepository) Reactivate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTypesReactivate(ctx, sqlc.SessionTypesReactivateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTypeRepository) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	q := sqlc.New(r.pool)
	params := sqlc.SessionTypesCheckActiveNameExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Name:     name,
		Column4:  pgtype.UUID{Valid: false},
	}
	if excludeID != nil {
		params.Column4 = uuidToPgtype(*excludeID)
	}
	return q.SessionTypesCheckActiveNameExists(ctx, params)
}

func (r *SessionTypeRepository) Exists(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	return q.SessionTypesExists(ctx, sqlc.SessionTypesExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *SessionTypeRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.SessionType, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.SessionTypesGetByIDForUpdate(ctx, sqlc.SessionTypesGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.SessionType{}, domainerrors.NotFound("session_type", "Session type not found.")
	}
	if err != nil {
		return domain.SessionType{}, fmt.Errorf("query session type for update: %w", err)
	}
	return mapSessionTypeFromGetForUpdateRow(row), nil
}

func mapSessionTypeFromListRow(row sqlc.SessionType) domain.SessionType {
	st := domain.SessionType{
		ID:           pgtypeUUIDToUUID(row.ID),
		TenantID:     pgtypeUUIDToUUID(row.TenantID),
		BranchID:     pgtypeUUIDToUUID(row.BranchID),
		Name:         row.Name,
		StartMinutes: pgtypeTimeToMinutes(row.StartTime),
		EndMinutes:   pgtypeTimeToMinutes(row.EndTime),
		IsActive:     row.IsActive,
		CreatedAt:    pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	return st
}

func mapSessionTypeFromGetRow(row sqlc.SessionType) domain.SessionType {
	st := domain.SessionType{
		ID:           pgtypeUUIDToUUID(row.ID),
		TenantID:     pgtypeUUIDToUUID(row.TenantID),
		BranchID:     pgtypeUUIDToUUID(row.BranchID),
		Name:         row.Name,
		StartMinutes: pgtypeTimeToMinutes(row.StartTime),
		EndMinutes:   pgtypeTimeToMinutes(row.EndTime),
		IsActive:     row.IsActive,
		CreatedAt:    pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	return st
}

func mapSessionTypeFromGetForUpdateRow(row sqlc.SessionType) domain.SessionType {
	st := domain.SessionType{
		ID:           pgtypeUUIDToUUID(row.ID),
		TenantID:     pgtypeUUIDToUUID(row.TenantID),
		BranchID:     pgtypeUUIDToUUID(row.BranchID),
		Name:         row.Name,
		StartMinutes: pgtypeTimeToMinutes(row.StartTime),
		EndMinutes:   pgtypeTimeToMinutes(row.EndTime),
		IsActive:     row.IsActive,
		CreatedAt:    pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	return st
}

func mapSessionTypeFromListPaginatedRow(row sqlc.SessionType) domain.SessionType {
	st := domain.SessionType{
		ID:           pgtypeUUIDToUUID(row.ID),
		TenantID:     pgtypeUUIDToUUID(row.TenantID),
		BranchID:     pgtypeUUIDToUUID(row.BranchID),
		Name:         row.Name,
		StartMinutes: pgtypeTimeToMinutes(row.StartTime),
		EndMinutes:   pgtypeTimeToMinutes(row.EndTime),
		IsActive:     row.IsActive,
		CreatedAt:    pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	return st
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func minutesToPgtypeTime(m int) pgtype.Time {
	// Microseconds since midnight. pgtype.Time uses Microseconds field.
	us := int64(m) * 60 * 1_000_000
	return pgtype.Time{Microseconds: us, Valid: true}
}

func pgtypeTimeToMinutes(t pgtype.Time) int {
	if !t.Valid {
		return 0
	}
	return int(t.Microseconds / 60 / 1_000_000)
}

func mapSessionTypeSortRow(row sqlc.SessionType) domain.SessionType {
	st := domain.SessionType{
		ID:           pgtypeUUIDToUUID(row.ID),
		TenantID:     pgtypeUUIDToUUID(row.TenantID),
		BranchID:     pgtypeUUIDToUUID(row.BranchID),
		Name:         row.Name,
		StartMinutes: pgtypeTimeToMinutes(row.StartTime),
		EndMinutes:   pgtypeTimeToMinutes(row.EndTime),
		IsActive:     row.IsActive,
		CreatedAt:    pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:    pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	return st
}
