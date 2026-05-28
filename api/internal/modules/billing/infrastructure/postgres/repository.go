package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		TenantID:           uuidToPgtype(tenantID),
		BranchID:           uuidToPgtype(branchID),
		CheckInLocalDate:   timeToPgtypeDate(periodStartLocalDate),
		CheckInLocalDate_2: timeToPgtypeDate(periodEndExclusiveLocalDate),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.PreflightAttendanceSessionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PreflightAttendanceSessionRow{
			SessionID:         pgtypeUUIDToUUID(row.ID),
			ChildID:           pgtypeUUIDToUUID(row.ChildID),
			Status:            domain.AttendanceSessionStatus(row.Status),
			CheckInAt:         pgtypeTimestamptzToTime(row.CheckInAt),
			CheckOutAt:        pgtypeTimestamptzToTimePtr(row.CheckOutAt),
			CheckInLocalDate:  pgtypeDateToTime(row.CheckInLocalDate),
			CheckOutLocalDate: pgtypeDateToTimePtr(row.CheckOutLocalDate),
		})
	}
	return result, nil
}

// --- Generation (API-17) transactional methods ---

func (r *Repository) queriesTx(tx pgx.Tx) *sqlc.Queries {
	return sqlc.New(tx)
}

func (r *Repository) ListCandidateChildrenForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]domain.PreflightChildRow, error) {
	rows, err := r.queriesTx(tx).ListCandidateChildrenForUpdate(ctx, sqlc.ListCandidateChildrenForUpdateParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		BillingMonth: timeToPgtypeDate(billingMonth),
		StartDate:    timeToPgtypeDate(nextBillingMonth),
	})
	if err != nil {
		return nil, err
	}
	return mapCandidateRows(rows), nil
}

func (r *Repository) ListSelectedChildrenForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, childIDs []uuid.UUID) ([]domain.PreflightChildRow, error) {
	pgIDs := make([]pgtype.UUID, len(childIDs))
	for i, id := range childIDs {
		pgIDs[i] = uuidToPgtype(id)
	}
	rows, err := r.queriesTx(tx).ListSelectedChildrenForUpdate(ctx, sqlc.ListSelectedChildrenForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgIDs,
	})
	if err != nil {
		return nil, err
	}
	return mapSelectedRows(rows), nil
}

func (r *Repository) ListAttendanceSessions(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, periodStart, periodEndExclusive time.Time) ([]domain.PreflightAttendanceSessionRow, error) {
	rows, err := r.queriesTx(tx).ListAttendanceSessionsForGeneration(ctx, sqlc.ListAttendanceSessionsForGenerationParams{
		TenantID:           uuidToPgtype(tenantID),
		BranchID:           uuidToPgtype(branchID),
		CheckInLocalDate:   timeToPgtypeDate(periodStart),
		CheckInLocalDate_2: timeToPgtypeDate(periodEndExclusive),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.PreflightAttendanceSessionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PreflightAttendanceSessionRow{
			SessionID:         pgtypeUUIDToUUID(row.ID),
			ChildID:           pgtypeUUIDToUUID(row.ChildID),
			Status:            domain.AttendanceSessionStatus(row.Status),
			CheckInAt:         pgtypeTimestamptzToTime(row.CheckInAt),
			CheckOutAt:        pgtypeTimestamptzToTimePtr(row.CheckOutAt),
			CheckInLocalDate:  pgtypeDateToTime(row.CheckInLocalDate),
			CheckOutLocalDate: pgtypeDateToTimePtr(row.CheckOutLocalDate),
		})
	}
	return result, nil
}

func (r *Repository) CreateInvoiceRun(ctx context.Context, tx domain.Tx, params domain.InvoiceRunCreateParams) error {
	return r.queriesTx(tx).CreateInvoiceRun(ctx, sqlc.CreateInvoiceRunParams{
		ID:                      uuidToPgtype(params.ID),
		TenantID:                uuidToPgtype(params.TenantID),
		BranchID:                uuidToPgtype(params.BranchID),
		BillingMonth:            timeToPgtypeDate(params.BillingMonth),
		RunType:                 params.RunType,
		Status:                  params.Status,
		RequestedByUserID:       uuidToPgtype(params.RequestedByUserID),
		RequestedByMembershipID: uuidToPgtype(params.RequestedByMembershipID),
		RequestID:               pgtype.Text{String: params.RequestID, Valid: true},
	})
}

func (r *Repository) CompleteInvoiceRun(ctx context.Context, tx domain.Tx, params domain.InvoiceRunCompleteParams) error {
	return r.queriesTx(tx).CompleteInvoiceRun(ctx, sqlc.CompleteInvoiceRunParams{
		ID:            uuidToPgtype(params.ID),
		TenantID:      uuidToPgtype(params.TenantID),
		BranchID:      uuidToPgtype(params.BranchID),
		Status:        params.Status,
		EligibleCount: int32(params.EligibleCount),
		SuccessCount:  int32(params.SuccessCount),
		BlockedCount:  int32(params.BlockedCount),
		Details:       params.Details,
	})
}

func (r *Repository) GetMonthlyInvoiceForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (domain.InvoiceRow, bool, error) {
	row, err := r.queriesTx(tx).GetMonthlyInvoiceForUpdate(ctx, sqlc.GetMonthlyInvoiceForUpdateParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		ChildID:      uuidToPgtype(childID),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.InvoiceRow{}, false, nil
		}
		return domain.InvoiceRow{}, false, err
	}
	return domain.InvoiceRow{
		ID:                   pgtypeUUIDToUUID(row.ID),
		Status:               row.Status,
		InvoiceKind:          row.InvoiceKind,
		SubtotalMinor:        int(row.SubtotalMinor),
		FundedDeductionMinor: int(row.FundedDeductionMinor),
		TotalDueMinor:        int(row.TotalDueMinor),
		CalculationDetails:   json.RawMessage(row.CalculationDetails),
	}, true, nil
}

func (r *Repository) CreateDraftInvoice(ctx context.Context, tx domain.Tx, params domain.DraftInvoiceCreateParams) error {
	return r.queriesTx(tx).CreateDraftInvoice(ctx, sqlc.CreateDraftInvoiceParams{
		ID:                   uuidToPgtype(params.ID),
		TenantID:             uuidToPgtype(params.TenantID),
		BranchID:             uuidToPgtype(params.BranchID),
		ChildID:              uuidToPgtype(params.ChildID),
		BillingMonth:         timeToPgtypeDate(params.BillingMonth),
		InvoiceKind:          domain.InvoiceKindMonthly,
		Status:               domain.InvoiceStatusDraft,
		CurrencyCode:         params.CurrencyCode,
		GeneratedRunID:       uuidToPgtype(params.GeneratedRunID),
		SubtotalMinor:        int32(params.SubtotalMinor),
		FundedDeductionMinor: int32(params.FundedDeductionMinor),
		TotalDueMinor:        int32(params.TotalDueMinor),
		PeriodStartDate:      timeToPgtypeDate(params.PeriodStartDate),
		PeriodEndDate:        timeToPgtypeDate(params.PeriodEndDate),
		CalculationDetails:   params.CalculationDetails,
	})
}

func (r *Repository) UpdateDraftInvoice(ctx context.Context, tx domain.Tx, params domain.DraftInvoiceUpdateParams) error {
	return r.queriesTx(tx).UpdateDraftInvoice(ctx, sqlc.UpdateDraftInvoiceParams{
		ID:                   uuidToPgtype(params.ID),
		TenantID:             uuidToPgtype(params.TenantID),
		BranchID:             uuidToPgtype(params.BranchID),
		GeneratedRunID:       uuidToPgtype(params.GeneratedRunID),
		SubtotalMinor:        int32(params.SubtotalMinor),
		FundedDeductionMinor: int32(params.FundedDeductionMinor),
		TotalDueMinor:        int32(params.TotalDueMinor),
		CalculationDetails:   params.CalculationDetails,
	})
}

func (r *Repository) DeleteDraftSystemInvoiceLines(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID uuid.UUID) error {
	_, err := r.queriesTx(tx).DeleteDraftSystemInvoiceLines(ctx, sqlc.DeleteDraftSystemInvoiceLinesParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		InvoiceID: uuidToPgtype(invoiceID),
	})
	return err
}

func (r *Repository) ListDraftExtraLines(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID uuid.UUID) ([]domain.ExtraLineRow, error) {
	rows, err := r.queriesTx(tx).ListDraftExtraLines(ctx, sqlc.ListDraftExtraLinesParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		InvoiceID: uuidToPgtype(invoiceID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.ExtraLineRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ExtraLineRow{
			ID:              pgtypeUUIDToUUID(row.ID),
			LineKind:        row.LineKind,
			LineAmountMinor: int(row.LineAmountMinor),
			Details:         json.RawMessage(row.Details),
		})
	}
	return result, nil
}

func (r *Repository) InsertInvoiceLine(ctx context.Context, tx domain.Tx, params domain.InvoiceLineCreateParams) error {
	return r.queriesTx(tx).InsertInvoiceLine(ctx, sqlc.InsertInvoiceLineParams{
		ID:                     uuidToPgtype(params.ID),
		TenantID:               uuidToPgtype(params.TenantID),
		BranchID:               uuidToPgtype(params.BranchID),
		InvoiceID:              uuidToPgtype(params.InvoiceID),
		LineKind:               params.LineKind,
		Description:            params.Description,
		SortOrder:              int32(params.SortOrder),
		QuantityMinutes:        pgtypeInt4OrNil(params.QuantityMinutes),
		UnitAmountMinor:        pgtypeInt4OrNil(params.UnitAmountMinor),
		LineAmountMinor:        int32(params.LineAmountMinor),
		RawAttendedMinutes:     pgtypeInt4OrNil(params.RawAttendedMinutes),
		RoundedAttendedMinutes: pgtypeInt4OrNil(params.RoundedAttendedMinutes),
		FundedAllowanceMinutes: pgtypeInt4OrNil(params.FundedAllowanceMinutes),
		FundedDeductionMinutes: pgtypeInt4OrNil(params.FundedDeductionMinutes),
		CoreBillableMinutes:    pgtypeInt4OrNil(params.CoreBillableMinutes),
		SessionCount:           pgtypeInt4OrNil(params.SessionCount),
		Details:                params.Details,
	})
}

// --- Helpers ---

func mapCandidateRows(rows []sqlc.ListCandidateChildrenForUpdateRow) []domain.PreflightChildRow {
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
	return result
}

func mapSelectedRows(rows []sqlc.ListSelectedChildrenForUpdateRow) []domain.PreflightChildRow {
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
	return result
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

func pgtypeInt4OrNil(v int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(v), Valid: true}
}
