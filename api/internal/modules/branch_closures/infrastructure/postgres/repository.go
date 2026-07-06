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

	"nursery-management-system/api/internal/modules/branch_closures/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, closure domain.BranchClosureDay) error {
	q := sqlc.New(r.pool)
	return q.BranchClosureDaysCreate(ctx, sqlc.BranchClosureDaysCreateParams{
		ID:       uuidToPgtype(closure.ID),
		TenantID: uuidToPgtype(closure.TenantID),
		BranchID: uuidToPgtype(closure.BranchID),
		Date:     timeToPgtypeDate(closure.Date),
		Reason:   pgtypeTextFromPtr(closure.Reason),
	})
}

func (r *Repository) ListByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]domain.BranchClosureDay, error) {
	q := sqlc.New(r.pool)
	rows, err := q.BranchClosureDaysListByBranchAndDateRange(ctx, sqlc.BranchClosureDaysListByBranchAndDateRangeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Date:     timeToPgtypeDate(from),
		Date_2:   timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query closure days list: %w", err)
	}
	out := make([]domain.BranchClosureDay, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapClosureDay(row))
	}
	return out, nil
}

func (r *Repository) ListByBranchAndDateRangePaginated(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time, limit, offset int) ([]domain.BranchClosureDay, error) {
	q := sqlc.New(r.pool)
	rows, err := q.BranchClosureDaysListByBranchAndDateRangePaginated(ctx, sqlc.BranchClosureDaysListByBranchAndDateRangePaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Date:     timeToPgtypeDate(from),
		Date_2:   timeToPgtypeDate(to),
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("query closure days list paginated: %w", err)
	}
	out := make([]domain.BranchClosureDay, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapClosureDay(row))
	}
	return out, nil
}

func (r *Repository) CountByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.BranchClosureDaysCountByBranchAndDateRange(ctx, sqlc.BranchClosureDaysCountByBranchAndDateRangeParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Date:     timeToPgtypeDate(from),
		Date_2:   timeToPgtypeDate(to),
	})
	if err != nil {
		return 0, fmt.Errorf("query closure days count: %w", err)
	}
	return int(count), nil
}

func (r *Repository) Delete(ctx context.Context, tenantID, branchID, id uuid.UUID) error {
	q := sqlc.New(r.pool)
	rowsAffected, err := q.BranchClosureDaysDelete(ctx, sqlc.BranchClosureDaysDeleteParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(id),
	})
	if err != nil {
		return fmt.Errorf("delete closure day: %w", err)
	}
	if rowsAffected == 0 {
		return domainerrors.NotFound("closure_day", "Closure day not found.")
	}
	return nil
}

func (r *Repository) DateExists(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time) (bool, error) {
	q := sqlc.New(r.pool)
	exists, err := q.BranchClosureDaysDateExists(ctx, sqlc.BranchClosureDaysDateExistsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Date:     timeToPgtypeDate(date),
	})
	if err != nil {
		return false, fmt.Errorf("check date exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) ListClosureDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]time.Time, error) {
	q := sqlc.New(r.pool)
	from := month
	to := month.AddDate(0, 1, 0).AddDate(0, 0, -1)
	rows, err := q.BranchClosureDaysListClosureDatesForMonth(ctx, sqlc.BranchClosureDaysListClosureDatesForMonthParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Date:     timeToPgtypeDate(from),
		Date_2:   timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("query closure dates for month: %w", err)
	}
	out := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		out = append(out, pgtypeDateToTime(row))
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

func pgtypeTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtypeTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func mapClosureDay(row sqlc.BranchClosureDay) domain.BranchClosureDay {
	return domain.BranchClosureDay{
		ID:        pgtypeUUIDToUUID(row.ID),
		TenantID:  pgtypeUUIDToUUID(row.TenantID),
		BranchID:  pgtypeUUIDToUUID(row.BranchID),
		Date:      pgtypeDateToTime(row.Date),
		Reason:    pgtypeTextToPtr(row.Reason),
		CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
	}
}
