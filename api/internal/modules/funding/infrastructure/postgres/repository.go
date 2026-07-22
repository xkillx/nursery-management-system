package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListOverview(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]domain.OverviewRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.FundingOverviewList(ctx, sqlc.FundingOverviewListParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return nil, fmt.Errorf("list funding overview: %w", err)
	}

	result := make([]domain.OverviewRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapOverviewRow(row))
	}
	return result, nil
}

func (r *Repository) ListOverviewPaginated(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time, limit, offset int) ([]domain.OverviewRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.FundingOverviewListPaginated(ctx, sqlc.FundingOverviewListPaginatedParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  timeToPgtypeDate(billingMonth),
		Limit:    pgtype.Int4{Int32: int32(limit), Valid: true},
		Offset:   pgtype.Int4{Int32: int32(offset), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list funding overview paginated: %w", err)
	}

	result := make([]domain.OverviewRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapOverviewRowFromPaginatedRow(row))
	}
	return result, nil
}

func (r *Repository) CountOverview(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.FundingOverviewCount(ctx, sqlc.FundingOverviewCountParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return 0, fmt.Errorf("count funding overview: %w", err)
	}
	return int(count), nil
}

func mapOverviewRowFromPaginatedRow(row sqlc.FundingOverviewListPaginatedRow) domain.OverviewRow {
	return domain.OverviewRow{
		ChildID:            pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:     row.ChildFirstName,
		ChildMiddleName:    pgtypeTextToStringPtr(row.ChildMiddleName),
		ChildLastName:      pgtypeTextToStringPtr(row.ChildLastName),
		IsActive:           row.IsActive,
		StartDate:          pgtypeDateToTime(row.StartDate),
		EndDate:            pgtypeDateToTimePtr(row.EndDate),
		FundingRecordID:    pgtypeUUIDToUUIDPtr(row.FundingRecordID),
		FundingEnabled:     row.FundingEnabled.Bool,
		FundingType:        pgtypeTextToStringPtr(row.FundingType),
		FundingModel:       pgtypeTextToStringPtr(row.FundingModel),
		FundedHoursPerWeek: numericToFloat64Ptr(row.FundedHoursPerWeek),
		FundingStartDate:   pgtypeDateToTimePtr(row.FundingStartDate),
		FundingEndDate:     pgtypeDateToTimePtr(row.FundingEndDate),
		FundingUpdatedAt:   pgtypeTimestamptzToTimePtr(row.FundingUpdatedAt),
		ChildPhotoPath:     pgtypeTextToStringPtr(row.ProfilePhotoPath),
	}
}

func mapOverviewRow(row sqlc.FundingOverviewListRow) domain.OverviewRow {
	return domain.OverviewRow{
		ChildID:            pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:     row.ChildFirstName,
		ChildMiddleName:    pgtypeTextToStringPtr(row.ChildMiddleName),
		ChildLastName:      pgtypeTextToStringPtr(row.ChildLastName),
		IsActive:           row.IsActive,
		StartDate:          row.StartDate.Time,
		EndDate:            pgtypeDateToTimePtr(row.EndDate),
		FundingRecordID:    pgtypeUUIDToUUIDPtr(row.FundingRecordID),
		FundingEnabled:     row.FundingEnabled.Bool,
		FundingType:        pgtypeTextToStringPtr(row.FundingType),
		FundingModel:       pgtypeTextToStringPtr(row.FundingModel),
		FundedHoursPerWeek: numericToFloat64Ptr(row.FundedHoursPerWeek),
		FundingStartDate:   pgtypeDateToTimePtr(row.FundingStartDate),
		FundingEndDate:     pgtypeDateToTimePtr(row.FundingEndDate),
		FundingUpdatedAt:   pgtypeTimestamptzToTimePtr(row.FundingUpdatedAt),
		ChildPhotoPath:     pgtypeTextToStringPtr(row.ProfilePhotoPath),
	}
}

type HistoryRepository struct {
	pool *pgxpool.Pool
}

func NewHistoryRepository(pool *pgxpool.Pool) *HistoryRepository {
	return &HistoryRepository{pool: pool}
}

func (r *HistoryRepository) Create(ctx context.Context, history domain.FundingHistory) error {
	q := sqlc.New(r.pool)
	return q.ChildFundingHistoryInsert(ctx, sqlc.ChildFundingHistoryInsertParams{
		ID:                 uuidToPgtype(history.ID),
		TenantID:           uuidToPgtype(history.TenantID),
		BranchID:           uuidToPgtype(history.BranchID),
		ChildID:            uuidToPgtype(history.ChildID),
		FundingType:        pgtype.Text{String: ptrStr(history.FundingType), Valid: history.FundingType != nil},
		FundingModel:       pgtype.Text{String: ptrStr(history.FundingModel), Valid: history.FundingModel != nil},
		FundedHoursPerWeek: ptrToNumeric(history.FundedHoursPerWeek),
		FundingStartDate:   ptrToPgtypeDate(history.FundingStartDate),
		FundingEndDate:     ptrToPgtypeDate(history.FundingEndDate),
		ChangedAt:          pgtype.Timestamptz{Time: history.ChangedAt, Valid: true},
		ChangedByUserID:    uuidToPgtype(history.ChangedByUserID),
	})
}

func (r *HistoryRepository) ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.FundingHistory, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ChildFundingHistoryListByChild(ctx, sqlc.ChildFundingHistoryListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list funding history: %w", err)
	}

	result := make([]domain.FundingHistory, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapHistoryRow(row))
	}
	return result, nil
}

func mapHistoryRow(row sqlc.ChildFundingHistory) domain.FundingHistory {
	var fundingType *string
	if row.FundingType.Valid {
		fundingType = &row.FundingType.String
	}

	var fundingModel *string
	if row.FundingModel.Valid {
		fundingModel = &row.FundingModel.String
	}

	var fundedHoursPerWeek *float64
	if row.FundedHoursPerWeek.Valid {
		f, _ := row.FundedHoursPerWeek.Float64Value()
		fundedHoursPerWeek = &f.Float64
	}

	var fundingStartDate *time.Time
	if row.FundingStartDate.Valid {
		t := row.FundingStartDate.Time
		fundingStartDate = &t
	}

	var fundingEndDate *time.Time
	if row.FundingEndDate.Valid {
		t := row.FundingEndDate.Time
		fundingEndDate = &t
	}

	return domain.FundingHistory{
		ID:                 pgtypeUUIDToUUID(row.ID),
		TenantID:           pgtypeUUIDToUUID(row.TenantID),
		BranchID:           pgtypeUUIDToUUID(row.BranchID),
		ChildID:            pgtypeUUIDToUUID(row.ChildID),
		FundingType:        fundingType,
		FundingModel:       fundingModel,
		FundedHoursPerWeek: fundedHoursPerWeek,
		FundingStartDate:   fundingStartDate,
		FundingEndDate:     fundingEndDate,
		ChangedAt:          pgtypeTimestamptzToTime(row.ChangedAt),
		ChangedByUserID:    pgtypeUUIDToUUID(row.ChangedByUserID),
	}
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{}
	}
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%g", *f))
	return n
}

func ptrToPgtypeDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func (r *Repository) ListExpiringSoon(ctx context.Context, tenantID, branchID uuid.UUID, withinDays int) ([]domain.ExpiringFundingRecord, error) {
	q := sqlc.New(r.pool)
	rows, err := q.FundingExpiringSoon(ctx, sqlc.FundingExpiringSoonParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  int32(withinDays),
	})
	if err != nil {
		return nil, fmt.Errorf("list expiring funding: %w", err)
	}

	result := make([]domain.ExpiringFundingRecord, 0, len(rows))
	for _, row := range rows {
		var fundingType *string
		if row.FundingType != "" {
			fundingType = &row.FundingType
		}
		var fundedHours *float64
		if row.FundedHoursPerWeek.Valid {
			f := numericToFloat64(row.FundedHoursPerWeek)
			fundedHours = &f
		}
		result = append(result, domain.ExpiringFundingRecord{
			FundingRecordID:    pgtypeUUIDToUUID(row.FundingRecordID),
			ChildID:            pgtypeUUIDToUUID(row.ChildID),
			ChildFirstName:     row.ChildFirstName,
			ChildMiddleName:    pgtypeTextToStringPtr(row.ChildMiddleName),
			ChildLastName:      pgtypeTextToStringPtr(row.ChildLastName),
			FundingType:        fundingType,
			FundedHoursPerWeek: fundedHours,
			FundingEndDate:     pgtypeDateToTime(row.FundingEndDate),
		})
	}
	return result, nil
}

func (r *Repository) GetFundedChildrenCount(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) (domain.EnhancedOverviewMetrics, error) {
	q := sqlc.New(r.pool)
	row, err := q.FundingFundedChildrenCount(ctx, sqlc.FundingFundedChildrenCountParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return domain.EnhancedOverviewMetrics{}, fmt.Errorf("count funded children: %w", err)
	}
	return domain.EnhancedOverviewMetrics{
		TotalFundedChildren: int(row.TotalFundedCount),
		FifteenHourCount:    int(row.FifteenHourCount),
		ThirtyHourCount:     int(row.ThirtyHourCount),
	}, nil
}

func (r *Repository) GetBookedHoursThisWeek(ctx context.Context, tenantID, branchID uuid.UUID) (float64, error) {
	q := sqlc.New(r.pool)
	val, err := q.FundingBookedHoursThisWeek(ctx, sqlc.FundingBookedHoursThisWeekParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return 0, fmt.Errorf("get booked hours this week: %w", err)
	}
	return numericToFloat64(val), nil
}

func (r *Repository) GetExpiringSoonCount(ctx context.Context, tenantID, branchID uuid.UUID, withinDays int) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.FundingExpiringSoonCount(ctx, sqlc.FundingExpiringSoonCountParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  int32(withinDays),
	})
	if err != nil {
		return 0, fmt.Errorf("count expiring soon: %w", err)
	}
	return int(count), nil
}

func (r *Repository) GetChildAllocation(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonthStart, billingMonthEnd time.Time) ([]domain.AllocationEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.FundingChildAllocation(ctx, sqlc.FundingChildAllocationParams{
		TenantID:          uuidToPgtype(tenantID),
		BranchID:          uuidToPgtype(branchID),
		ChildID:           uuidToPgtype(childID),
		BillingMonthEnd:   timeToPgtypeDate(billingMonthEnd),
		BillingMonthStart: timeToPgtypeDate(billingMonthStart),
	})
	if err != nil {
		return nil, fmt.Errorf("get child allocation: %w", err)
	}

	result := make([]domain.AllocationEntry, 0, len(rows))
	for _, row := range rows {
		var endDate *time.Time
		if row.EffectiveEndDate.Valid {
			t := row.EffectiveEndDate.Time
			endDate = &t
		}
		result = append(result, domain.AllocationEntry{
			BookingID:              pgtypeUUIDToUUID(row.BookingID),
			EffectiveStartDate:     pgtypeDateToTime(row.EffectiveStartDate),
			EffectiveEndDate:       endDate,
			DayOfWeek:              row.DayOfWeek,
			SessionTypeName:        row.SessionTypeName,
			SessionDurationMinutes: int(row.SessionDurationMinutes),
		})
	}
	return result, nil
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
	v := uuid.UUID(u.Bytes)
	return &v
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func pgtypeDateToTimePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func pgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
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

func numericToFloat64(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

func numericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	v := f.Float64
	return &v
}
