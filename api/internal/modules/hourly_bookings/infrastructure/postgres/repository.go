package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/hourly_bookings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type HourlyBookingRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *HourlyBookingRepository {
	return &HourlyBookingRepository{pool: pool}
}

func (r *HourlyBookingRepository) Create(ctx context.Context, booking domain.HourlyBooking) error {
	q := sqlc.New(r.pool)
	return q.HourlyBookingsCreate(ctx, sqlc.HourlyBookingsCreateParams{
		ID:                   uuidToPgtype(booking.ID),
		TenantID:             uuidToPgtype(booking.TenantID),
		BranchID:             uuidToPgtype(booking.BranchID),
		ChildID:              uuidToPgtype(booking.ChildID),
		CalendarDate:         timeToPgtypeDate(booking.CalendarDate),
		StartTimeMinutes:     int32(booking.StartTimeMinutes),
		DurationMinutes:      int32(booking.DurationMinutes),
		SessionTypeID:        ptrToPgtypeUUID(booking.SessionTypeID),
		BookedByMembershipID: uuidToPgtype(booking.BookedByMembershipID),
	})
}

func (r *HourlyBookingRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]domain.HourlyBooking, error) {
	q := sqlc.New(r.pool)
	params := sqlc.HourlyBookingsListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  pgtype.Date{Valid: false},
		Column5:  pgtype.Date{Valid: false},
		Column6:  activeOnly,
	}
	if childID != nil {
		params.Column3 = uuidToPgtype(*childID)
	}
	if from != nil {
		params.Column4 = timeToPgtypeDate(*from)
	}
	if to != nil {
		params.Column5 = timeToPgtypeDate(*to)
	}
	rows, err := q.HourlyBookingsListByBranch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query hourly bookings list: %w", err)
	}
	out := make([]domain.HourlyBooking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapHourlyBooking(row))
	}
	return out, nil
}

func (r *HourlyBookingRepository) ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool, limit, offset int) ([]domain.HourlyBooking, error) {
	q := sqlc.New(r.pool)
	params := sqlc.HourlyBookingsListByBranchPaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  pgtype.Date{Valid: false},
		Column5:  pgtype.Date{Valid: false},
		Column6:  activeOnly,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	}
	if childID != nil {
		params.Column3 = uuidToPgtype(*childID)
	}
	if from != nil {
		params.Column4 = timeToPgtypeDate(*from)
	}
	if to != nil {
		params.Column5 = timeToPgtypeDate(*to)
	}
	rows, err := q.HourlyBookingsListByBranchPaginated(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query hourly bookings list paginated: %w", err)
	}
	out := make([]domain.HourlyBooking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapHourlyBooking(row))
	}
	return out, nil
}

func (r *HourlyBookingRepository) CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) (int, error) {
	q := sqlc.New(r.pool)
	params := sqlc.HourlyBookingsCountByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  pgtype.Date{Valid: false},
		Column5:  pgtype.Date{Valid: false},
		Column6:  activeOnly,
	}
	if childID != nil {
		params.Column3 = uuidToPgtype(*childID)
	}
	if from != nil {
		params.Column4 = timeToPgtypeDate(*from)
	}
	if to != nil {
		params.Column5 = timeToPgtypeDate(*to)
	}
	count, err := q.HourlyBookingsCountByBranch(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("query hourly bookings count: %w", err)
	}
	return int(count), nil
}

func (r *HourlyBookingRepository) ListActiveByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]domain.HourlyBooking, error) {
	q := sqlc.New(r.pool)
	rows, err := q.HourlyBookingsListByChildAndDateRange(ctx, sqlc.HourlyBookingsListByChildAndDateRangeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
		FromDate: timeToPgtypeDate(from),
		ToDate:   timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query hourly bookings by child date range: %w", err)
	}
	out := make([]domain.HourlyBooking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapHourlyBooking(row))
	}
	return out, nil
}

func (r *HourlyBookingRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.HourlyBooking, error) {
	q := sqlc.New(r.pool)
	row, err := q.HourlyBookingsGetByID(ctx, sqlc.HourlyBookingsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.HourlyBooking{}, domainerrors.NotFound("hourly_booking", "Hourly booking not found.")
	}
	if err != nil {
		return domain.HourlyBooking{}, fmt.Errorf("query hourly booking by id: %w", err)
	}
	return mapHourlyBooking(row), nil
}

func (r *HourlyBookingRepository) Cancel(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.HourlyBookingsCancel(ctx, sqlc.HourlyBookingsCancelParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}
