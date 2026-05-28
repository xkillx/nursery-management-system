package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListPreflightChildren(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]domain.PreflightChildRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.PreflightListChildren(ctx, sqlc.PreflightListChildrenParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		BillingMonth: timeToPgtypeDate(billingMonth),
		StartDate:    timeToPgtypeDate(nextBillingMonth),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.PreflightChildRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PreflightChildRow{
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			FullName:               row.FullName,
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    int(row.CoreHourlyRateMinor),
			HasGuardianLink:        row.HasGuardianLink,
			FundingProfileID:       pgtypeUUIDToUUIDPtr(row.FundingProfileID),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			ExistingInvoiceID:      pgtypeUUIDToUUIDPtr(row.ExistingInvoiceID),
			ExistingInvoiceStatus:  pgtypeTextToStrPtr(row.ExistingInvoiceStatus),
		})
	}
	return result, nil
}

func (r *Repository) ListPreflightAttendanceSessions(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]domain.PreflightAttendanceSessionRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.PreflightListAttendanceSessions(ctx, sqlc.PreflightListAttendanceSessionsParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		CheckInLocalDate: timeToPgtypeDate(periodStartLocalDate),
		CheckInLocalDate_2: timeToPgtypeDate(periodEndExclusiveLocalDate),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.PreflightAttendanceSessionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PreflightAttendanceSessionRow{
			SessionID:        pgtypeUUIDToUUID(row.ID),
			ChildID:          pgtypeUUIDToUUID(row.ChildID),
			Status:           domain.AttendanceSessionStatus(row.Status),
			CheckInAt:        pgtypeTimestamptzToTime(row.CheckInAt),
			CheckOutAt:       pgtypeTimestamptzToTimePtr(row.CheckOutAt),
			CheckInLocalDate: pgtypeDateToTime(row.CheckInLocalDate),
			CheckOutLocalDate: pgtypeDateToTimePtr(row.CheckOutLocalDate),
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
	id := uuid.UUID(u.Bytes)
	return &id
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

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgtypeInt4ToIntPtr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func pgtypeTextToStrPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
