package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/bookings/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (r *BookingRepository) Create(ctx context.Context, booking domain.Booking) error {
	q := sqlc.New(r.pool)

	var sessionTemplateID pgtype.UUID
	if booking.SessionTemplateID != nil {
		sessionTemplateID = uuidToPgtype(*booking.SessionTemplateID)
	}

	var sessionEntriesJSON []byte
	if len(booking.SessionEntries) > 0 {
		sessionEntriesJSON, _ = json.Marshal(booking.SessionEntries)
	}

	return q.BookingsCreate(ctx, sqlc.BookingsCreateParams{
		ID:                   uuidToPgtype(booking.ID),
		TenantID:             uuidToPgtype(booking.TenantID),
		BranchID:             uuidToPgtype(booking.BranchID),
		ChildID:              uuidToPgtype(booking.ChildID),
		SessionTemplateID:    sessionTemplateID,
		DaysOfWeek:           booking.DaysOfWeek,
		EffectiveStartDate:   timeToPgtypeDate(booking.EffectiveStartDate),
		EffectiveEndDate:     timeToPgtypeDatePtr(booking.EffectiveEndDate),
		FundingType:          stringToPgtypeText(booking.FundingType),
		FundingHoursPerWeek:  float64ToPgtypeNumeric(booking.FundingHoursPerWeek),
		LaReference:          stringToPgtypeText(booking.LaReference),
		SessionEntries:       sessionEntriesJSON,
		BookedByMembershipID: uuidToPgtype(booking.BookedByMembershipID),
	})
}

func (r *BookingRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Booking, error) {
	q := sqlc.New(r.pool)
	row, err := q.BookingsGetByID(ctx, sqlc.BookingsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("query booking by id: %w", err)
	}
	return mapBooking(bookingsGetByIDRowToBookingRow(row)), nil
}

func (r *BookingRepository) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.Booking, error) {
	q := sqlc.New(tx.(pgx.Tx))
	row, err := q.BookingsGetByIDForUpdate(ctx, sqlc.BookingsGetByIDForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if isNoRows(err) {
		return domain.Booking{}, domain.ErrBookingNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("query booking for update: %w", err)
	}
	return mapBooking(bookingsGetByIDForUpdateRowToBookingRow(row)), nil
}

func (r *BookingRepository) ListByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.ListFilters, limit, offset int) ([]domain.Booking, error) {
	q := sqlc.New(r.pool)
	params := sqlc.BookingsListByBranchPaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  "",
		Column5:  "",
		Column6:  pgtype.Date{Valid: false},
		Column7:  pgtype.Date{Valid: false},
		Column8:  filters.ActiveOnly,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	}
	if filters.ChildID != nil {
		params.Column3 = uuidToPgtype(*filters.ChildID)
	}
	if filters.Status != nil {
		params.Column4 = *filters.Status
	}
	if filters.FundingType != nil {
		params.Column5 = *filters.FundingType
	}
	if filters.From != nil {
		params.Column6 = timeToPgtypeDate(*filters.From)
	}
	if filters.To != nil {
		params.Column7 = timeToPgtypeDate(*filters.To)
	}
	rows, err := q.BookingsListByBranchPaginated(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query bookings list paginated: %w", err)
	}
	out := make([]domain.Booking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBooking(bookingsListByBranchPaginatedRowToBookingRow(row)))
	}
	return out, nil
}

func (r *BookingRepository) CountByBranch(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.ListFilters) (int, error) {
	q := sqlc.New(r.pool)
	params := sqlc.BookingsCountByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  "",
		Column5:  "",
		Column6:  pgtype.Date{Valid: false},
		Column7:  pgtype.Date{Valid: false},
		Column8:  filters.ActiveOnly,
	}
	if filters.ChildID != nil {
		params.Column3 = uuidToPgtype(*filters.ChildID)
	}
	if filters.Status != nil {
		params.Column4 = *filters.Status
	}
	if filters.FundingType != nil {
		params.Column5 = *filters.FundingType
	}
	if filters.From != nil {
		params.Column6 = timeToPgtypeDate(*filters.From)
	}
	if filters.To != nil {
		params.Column7 = timeToPgtypeDate(*filters.To)
	}
	count, err := q.BookingsCountByBranch(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("query bookings count: %w", err)
	}
	return int(count), nil
}

func (r *BookingRepository) Update(ctx context.Context, tx domain.Tx, booking domain.Booking) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.BookingsUpdate(ctx, sqlc.BookingsUpdateParams{
		TenantID:            uuidToPgtype(booking.TenantID),
		BranchID:            uuidToPgtype(booking.BranchID),
		ID:                  uuidToPgtype(booking.ID),
		DaysOfWeek:          booking.DaysOfWeek,
		EffectiveStartDate:  timeToPgtypeDate(booking.EffectiveStartDate),
		EffectiveEndDate:    timeToPgtypeDatePtr(booking.EffectiveEndDate),
		FundingType:         stringToPgtypeText(booking.FundingType),
		FundingHoursPerWeek: float64ToPgtypeNumeric(booking.FundingHoursPerWeek),
		LaReference:         stringToPgtypeText(booking.LaReference),
	})
}

func (r *BookingRepository) Cancel(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.BookingsCancel(ctx, sqlc.BookingsCancelParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *BookingRepository) Pause(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(tx.(pgx.Tx))
	return q.BookingsPause(ctx, sqlc.BookingsPauseParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
}

func (r *BookingRepository) ListByChildAndDateRange(ctx context.Context, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]domain.Booking, error) {
	q := sqlc.New(r.pool)
	rows, err := q.BookingsListByChildAndDateRange(ctx, sqlc.BookingsListByChildAndDateRangeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
		FromDate: timeToPgtypeDate(from),
		ToDate:   timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query bookings by child date range: %w", err)
	}
	out := make([]domain.Booking, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBooking(bookingsListByChildAndDateRangeRowToBookingRow(row)))
	}
	return out, nil
}

func (r *BookingRepository) ListUnifiedByBranchPaginated(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.ListFilters, limit, offset int) ([]domain.UnifiedBookingRow, error) {
	q := sqlc.New(r.pool)
	params := sqlc.BookingsUnifiedListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgtype.UUID{Valid: false},
		Column4:  "",
		Column5:  "",
		Column6:  pgtype.Date{Valid: false},
		Column7:  pgtype.Date{Valid: false},
		Column8:  filters.ActiveOnly,
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	}
	if filters.ChildID != nil {
		params.Column3 = uuidToPgtype(*filters.ChildID)
	}
	if filters.Status != nil {
		params.Column4 = *filters.Status
	}
	if filters.FundingType != nil {
		params.Column5 = *filters.FundingType
	}
	if filters.From != nil {
		params.Column6 = timeToPgtypeDate(*filters.From)
	}
	if filters.To != nil {
		params.Column7 = timeToPgtypeDate(*filters.To)
	}
	rows, err := q.BookingsUnifiedListByBranch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query unified bookings list: %w", err)
	}
	out := make([]domain.UnifiedBookingRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapUnifiedBookingRow(row))
	}
	return out, nil
}
