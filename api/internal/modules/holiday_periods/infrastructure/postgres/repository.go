package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/holiday_periods/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, hp domain.HolidayPeriod) error {
	q := sqlc.New(r.pool)
	return q.HolidayPeriodsCreate(ctx, sqlc.HolidayPeriodsCreateParams{
		ID:        uuidToPgtype(hp.ID),
		TenantID:  uuidToPgtype(hp.TenantID),
		BranchID:  uuidToPgtype(hp.BranchID),
		Name:      hp.Name,
		StartDate: timeToPgtypeDate(hp.StartDate),
		EndDate:   timeToPgtypeDate(hp.EndDate),
		Type:      string(hp.Type),
	})
}

func (r *Repository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.HolidayPeriod, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.HolidayPeriodsGetByID(ctx, sqlc.HolidayPeriodsGetByIDParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		if isNoRows(err) {
			return domain.HolidayPeriod{}, false, nil
		}
		return domain.HolidayPeriod{}, false, fmt.Errorf("get holiday period by id: %w", err)
	}
	return mapHolidayPeriod(row), true, nil
}

func (r *Repository) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID) ([]domain.HolidayPeriod, error) {
	q := sqlc.New(r.pool)
	rows, err := q.HolidayPeriodsListByBranch(ctx, sqlc.HolidayPeriodsListByBranchParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return nil, fmt.Errorf("list holiday periods: %w", err)
	}
	out := make([]domain.HolidayPeriod, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapHolidayPeriod(row))
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	q := sqlc.New(r.pool)
	params := sqlc.HolidayPeriodsUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	}
	if v, ok := fields["name"].(string); ok {
		params.SetName = true
		params.Name = v
	}
	if v, ok := fields["start_date"].(time.Time); ok {
		params.SetStartDate = true
		params.StartDate = timeToPgtypeDate(v)
	}
	if v, ok := fields["end_date"].(time.Time); ok {
		params.SetEndDate = true
		params.EndDate = timeToPgtypeDate(v)
	}
	if v, ok := fields["type"].(string); ok {
		params.SetType = true
		params.Type = v
	}
	rowsAffected, err := q.HolidayPeriodsUpdate(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("update holiday period: %w", err)
	}
	if rowsAffected == 0 {
		return 0, domainerrors.NotFound("holiday_period", "Holiday period not found.")
	}
	return rowsAffected, nil
}

func (r *Repository) Delete(ctx context.Context, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(r.pool)
	rowsAffected, err := q.HolidayPeriodsDelete(ctx, sqlc.HolidayPeriodsDeleteParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		return fmt.Errorf("delete holiday period: %w", err)
	}
	if rowsAffected == 0 {
		return domainerrors.NotFound("holiday_period", "Holiday period not found.")
	}
	return nil
}

func (r *Repository) ListForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, monthStart, monthEnd time.Time) ([]domain.HolidayPeriod, error) {
	q := sqlc.New(r.pool)
	rows, err := q.HolidayPeriodsListForBranchAndMonth(ctx, sqlc.HolidayPeriodsListForBranchAndMonthParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		EndDate:   timeToPgtypeDate(monthStart), // $3: end_date >= monthStart
		StartDate: timeToPgtypeDate(monthEnd),   // $4: start_date <= monthEnd
	})
	if err != nil {
		return nil, fmt.Errorf("list holiday periods for month: %w", err)
	}
	out := make([]domain.HolidayPeriod, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapHolidayPeriod(row))
	}
	return out, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func isNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func mapHolidayPeriod(row sqlc.HolidayPeriod) domain.HolidayPeriod {
	return domain.HolidayPeriod{
		ID:        pgtypeUUIDToUUID(row.ID),
		TenantID:  pgtypeUUIDToUUID(row.TenantID),
		BranchID:  pgtypeUUIDToUUID(row.BranchID),
		Name:      row.Name,
		StartDate: pgtypeDateToTime(row.StartDate),
		EndDate:   pgtypeDateToTime(row.EndDate),
		Type:      domain.HolidayPeriodType(row.Type),
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}
