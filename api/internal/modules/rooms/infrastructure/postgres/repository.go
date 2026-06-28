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

	"nursery-management-system/api/internal/modules/rooms/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.Room, error) {
	q := sqlc.New(r.pool)
	rows, err := q.RoomsListByBranch(ctx, sqlc.RoomsListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  !includeArchived,
	})
	if err != nil {
		return nil, fmt.Errorf("query rooms list: %w", err)
	}
	out := make([]domain.Room, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRoom(row))
	}
	return out, nil
}

func (r *RoomRepository) GetByID(ctx context.Context, tenantID, branchID, roomID uuid.UUID) (domain.Room, error) {
	q := sqlc.New(r.pool)
	row, err := q.RoomsGetByID(ctx, sqlc.RoomsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	})
	if isNoRows(err) {
		return domain.Room{}, domainerrors.NotFound("room", "Room not found.")
	}
	if err != nil {
		return domain.Room{}, fmt.Errorf("query room by id: %w", err)
	}
	return mapRoom(row), nil
}

func (r *RoomRepository) Create(ctx context.Context, room domain.Room) error {
	q := sqlc.New(r.pool)
	return q.RoomsCreate(ctx, sqlc.RoomsCreateParams{
		ID:       uuidToPgtype(room.ID),
		TenantID: uuidToPgtype(room.TenantID),
		BranchID: uuidToPgtype(room.BranchID),
		Name:     room.Name,
		Column5:  pgtypeTextToInterface(room.Description),
		AgeGroup: room.AgeGroup,
		Capacity: int32(room.Capacity),
	})
}

func (r *RoomRepository) Update(ctx context.Context, tenantID, branchID, roomID uuid.UUID, fields map[string]any) (int64, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	params := sqlc.RoomsUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	}

	if v, ok := fields["name"]; ok {
		params.SetName = int32(1)
		params.Name = v.(string)
	}
	if v, ok := fields["age_group"]; ok {
		params.SetAgeGroup = int32(1)
		params.AgeGroup = v.(string)
	}
	if v, ok := fields["capacity"]; ok {
		params.SetCapacity = int32(1)
		params.Capacity = int32(v.(int))
	}
	if v, ok := fields["description"]; ok {
		params.SetDescription = int32(1)
		params.Description = v.(string)
	}

	q := sqlc.New(r.pool)
	return q.RoomsUpdate(ctx, params)
}

func (r *RoomRepository) Archive(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.RoomsArchive(ctx, sqlc.RoomsArchiveParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	})
}

func (r *RoomRepository) Reactivate(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.RoomsReactivate(ctx, sqlc.RoomsReactivateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	})
}

func (r *RoomRepository) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeRoomID *uuid.UUID) (bool, error) {
	q := sqlc.New(r.pool)
	params := sqlc.RoomsCheckActiveNameExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Name:     name,
		Column4:  pgtype.UUID{Valid: false},
	}
	if excludeRoomID != nil {
		params.Column4 = uuidToPgtype(*excludeRoomID)
	}
	return q.RoomsCheckActiveNameExists(ctx, params)
}

func (r *RoomRepository) CountActiveChildren(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (int, error) {
	q := sqlc.New(tx.(pgx.Tx))
	count, err := q.RoomsCountActiveChildren(ctx, sqlc.RoomsCountActiveChildrenParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		RoomID:   uuidToPgtype(roomID),
	})
	if err != nil {
		return 0, fmt.Errorf("count active children: %w", err)
	}
	return int(count), nil
}

func (r *RoomRepository) Exists(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (bool, error) {
	q := sqlc.New(tx.(pgx.Tx))
	return q.RoomsExists(ctx, sqlc.RoomsExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	})
}

func (r *RoomRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (domain.Room, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.RoomsGetByIDForUpdate(ctx, sqlc.RoomsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(roomID),
	})
	if isNoRows(err) {
		return domain.Room{}, domainerrors.NotFound("room", "Room not found.")
	}
	if err != nil {
		return domain.Room{}, fmt.Errorf("query room for update: %w", err)
	}
	return mapRoom(row), nil
}

func (r *RoomRepository) CountAssignedChildrenByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (map[uuid.UUID]int, error) {
	q := sqlc.New(r.pool)
	rows, err := q.RoomsCountAssignedChildrenByBranch(ctx, sqlc.RoomsCountAssignedChildrenByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return nil, fmt.Errorf("count assigned children: %w", err)
	}
	out := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		out[pgtypeUUIDToUUID(row.RoomID)] = int(row.AssignedCount)
	}
	return out, nil
}

func mapRoom(row sqlc.Room) domain.Room {
	return domain.Room{
		ID:          pgtypeUUIDToUUID(row.ID),
		TenantID:    pgtypeUUIDToUUID(row.TenantID),
		BranchID:    pgtypeUUIDToUUID(row.BranchID),
		Name:        row.Name,
		Description: pgtypeTextToStringPtr(row.Description),
		AgeGroup:    row.AgeGroup,
		Capacity:    int(row.Capacity),
		IsActive:    row.IsActive,
		CreatedAt:   pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:   pgtypeTimestamptzToTime(row.UpdatedAt),
	}
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

func pgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgtypeTextToInterface(s *string) interface{} {
	if s == nil {
		return ""
	}
	return *s
}
