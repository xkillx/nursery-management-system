package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type AdHocBookingRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *AdHocBookingRepository {
	return &AdHocBookingRepository{pool: pool}
}

func (r *AdHocBookingRepository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, childID *uuid.UUID, from, to *time.Time, activeOnly bool) ([]domain.AdHocBooking, error) {
	q := sqlc.New(r.pool)
	params := sqlc.AdHocBookingsListByBranchParams{
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
	rows, err := q.AdHocBookingsListByBranch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query ad-hoc bookings list: %w", err)
	}
	out := make([]domain.AdHocBooking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAdHocBooking(row))
	}
	return out, nil
}

func (r *AdHocBookingRepository) ListActiveByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]domain.AdHocBooking, error) {
	q := sqlc.New(r.pool)
	rows, err := q.AdHocBookingsListByChildAndDateRange(ctx, sqlc.AdHocBookingsListByChildAndDateRangeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
		FromDate: timeToPgtypeDate(from),
		ToDate:   timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query ad-hoc bookings by child date range: %w", err)
	}
	out := make([]domain.AdHocBooking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAdHocBooking(row))
	}
	return out, nil
}

func (r *AdHocBookingRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.AdHocBooking, error) {
	q := sqlc.New(r.pool)
	row, err := q.AdHocBookingsGetByID(ctx, sqlc.AdHocBookingsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.AdHocBooking{}, domainerrors.NotFound("ad_hoc_booking", "Ad-hoc booking not found.")
	}
	if err != nil {
		return domain.AdHocBooking{}, fmt.Errorf("query ad-hoc booking by id: %w", err)
	}
	return mapAdHocBooking(row), nil
}

func (r *AdHocBookingRepository) Create(ctx context.Context, booking domain.AdHocBooking) error {
	q := sqlc.New(r.pool)
	return q.AdHocBookingsCreate(ctx, sqlc.AdHocBookingsCreateParams{
		ID:                   uuidToPgtype(booking.ID),
		TenantID:             uuidToPgtype(booking.TenantID),
		BranchID:             uuidToPgtype(booking.BranchID),
		ChildID:              uuidToPgtype(booking.ChildID),
		CalendarDate:         timeToPgtypeDate(booking.CalendarDate),
		SessionTypeID:        uuidToPgtype(booking.SessionTypeID),
		BookedByMembershipID: uuidToPgtype(booking.BookedByMembershipID),
	})
}

func (r *AdHocBookingRepository) Cancel(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.AdHocBookingsCancel(ctx, sqlc.AdHocBookingsCancelParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *AdHocBookingRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.AdHocBooking, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.AdHocBookingsGetByIDForUpdate(ctx, sqlc.AdHocBookingsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.AdHocBooking{}, domainerrors.NotFound("ad_hoc_booking", "Ad-hoc booking not found.")
	}
	if err != nil {
		return domain.AdHocBooking{}, fmt.Errorf("query ad-hoc booking for update: %w", err)
	}
	return mapAdHocBooking(row), nil
}
