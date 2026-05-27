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

	"nursery-management-system/api/internal/modules/absence/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type AbsenceRepository struct {
	pool *pgxpool.Pool
}

func NewAbsenceRepository(pool *pgxpool.Pool) *AbsenceRepository {
	return &AbsenceRepository{pool: pool}
}

func (r *AbsenceRepository) Create(ctx context.Context, tx pgx.Tx, marker domain.AbsenceMarker) (domain.AbsenceMarker, error) {
	q := sqlc.New(tx)
	row, err := q.AbsenceMarkersCreate(ctx, sqlc.AbsenceMarkersCreateParams{
		ID:                   uuidToPgtype(marker.ID),
		TenantID:             uuidToPgtype(marker.TenantID),
		BranchID:             uuidToPgtype(marker.BranchID),
		ChildID:              uuidToPgtype(marker.ChildID),
		LocalDate:            timeToPgtypeDate(marker.LocalDate),
		MarkedAt:             timeToPgtypeTimestamptz(marker.MarkedAt),
		MarkedByUserID:       uuidToPgtype(marker.MarkedByUserID),
		MarkedByMembershipID: uuidToPgtype(marker.MarkedByMembershipID),
	})
	if err != nil {
		return domain.AbsenceMarker{}, fmt.Errorf("create absence marker: %w", err)
	}
	return mapAbsenceMarker(row), nil
}

func (r *AbsenceRepository) FindActiveByChildDate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (domain.AbsenceMarker, bool, error) {
	q := sqlc.New(tx)
	row, err := q.AbsenceMarkersFindActiveByChildDate(ctx, sqlc.AbsenceMarkersFindActiveByChildDateParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ChildID:   uuidToPgtype(childID),
		LocalDate: timeToPgtypeDate(localDate),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AbsenceMarker{}, false, nil
	}
	if err != nil {
		return domain.AbsenceMarker{}, false, fmt.Errorf("find active absence marker: %w", err)
	}
	return mapAbsenceMarker(row), true, nil
}

func (r *AbsenceRepository) GetByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.AbsenceMarker, bool, error) {
	q := sqlc.New(tx)
	row, err := q.AbsenceMarkersGetByID(ctx, sqlc.AbsenceMarkersGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AbsenceMarker{}, false, nil
	}
	if err != nil {
		return domain.AbsenceMarker{}, false, fmt.Errorf("get absence marker by id: %w", err)
	}
	return mapAbsenceMarker(row), true, nil
}

func (r *AbsenceRepository) Clear(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, clearedAt time.Time, clearedByUserID, clearedByMembershipID uuid.UUID) (domain.AbsenceMarker, bool, error) {
	q := sqlc.New(tx)
	row, err := q.AbsenceMarkersClear(ctx, sqlc.AbsenceMarkersClearParams{
		TenantID:              uuidToPgtype(tenantID),
		BranchID:              uuidToPgtype(branchID),
		ID:                    uuidToPgtype(id),
		ClearedAt:             timeToPgtypeTimestamptz(clearedAt),
		ClearedByUserID:       uuidToPgtype(clearedByUserID),
		ClearedByMembershipID: uuidToPgtype(clearedByMembershipID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AbsenceMarker{}, false, nil
	}
	if err != nil {
		return domain.AbsenceMarker{}, false, fmt.Errorf("clear absence marker: %w", err)
	}
	return mapAbsenceMarker(row), true, nil
}

func (r *AbsenceRepository) HasAttendanceForChildDate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (bool, error) {
	q := sqlc.New(tx)
	return q.AbsenceMarkersHasAttendanceForChildDate(ctx, sqlc.AbsenceMarkersHasAttendanceForChildDateParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		ChildID:          uuidToPgtype(childID),
		CheckInLocalDate: timeToPgtypeDate(localDate),
	})
}

func mapAbsenceMarker(row sqlc.AbsenceMarker) domain.AbsenceMarker {
	return domain.AbsenceMarker{
		ID:                    pgtypeUUIDToUUID(row.ID),
		TenantID:              pgtypeUUIDToUUID(row.TenantID),
		BranchID:              pgtypeUUIDToUUID(row.BranchID),
		ChildID:               pgtypeUUIDToUUID(row.ChildID),
		LocalDate:             pgtypeDateToTime(row.LocalDate),
		MarkedAt:              pgtypeTimestamptzToTime(row.MarkedAt),
		MarkedByUserID:        pgtypeUUIDToUUID(row.MarkedByUserID),
		MarkedByMembershipID:  pgtypeUUIDToUUID(row.MarkedByMembershipID),
		ClearedAt:             pgtypeTimestamptzToTimePtr(row.ClearedAt),
		ClearedByUserID:       pgtypeUUIDToUUIDPtr(row.ClearedByUserID),
		ClearedByMembershipID: pgtypeUUIDToUUIDPtr(row.ClearedByMembershipID),
		CreatedAt:             pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:             pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

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
	id := uuid.UUID(u.Bytes)
	return &id
}

func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
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

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}
