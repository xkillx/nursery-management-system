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
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
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

func (r *Repository) ListActiveTermsForGeneration(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]domain.AdvancePayTermRow, error) {
	monthEnd := billingMonth.AddDate(0, 1, -1)
	rows, err := r.queriesTx(tx).BillingListActiveTermsForGeneration(ctx, sqlc.BillingListActiveTermsForGenerationParams{
		TenantID:      uuidToPgtype(tenantID),
		BranchID:      uuidToPgtype(branchID),
		BillingMonth:  timeToPgtypeDate(billingMonth),
		TermStartDate: timeToPgtypeDate(monthEnd),
	})
	if err != nil {
		return nil, fmt.Errorf("list active terms for generation: %w", err)
	}
	return mapAdvancePayTermRows(rows), nil
}

func (r *Repository) ListActiveTerms(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]domain.AdvancePayTermRow, error) {
	monthEnd := billingMonth.AddDate(0, 1, -1)
	rows, err := sqlc.New(r.pool).BillingListActiveTermsForGeneration(ctx, sqlc.BillingListActiveTermsForGenerationParams{
		TenantID:      uuidToPgtype(tenantID),
		BranchID:      uuidToPgtype(branchID),
		BillingMonth:  timeToPgtypeDate(billingMonth),
		TermStartDate: timeToPgtypeDate(monthEnd),
	})
	if err != nil {
		return nil, fmt.Errorf("list active terms (no tx): %w", err)
	}
	return mapAdvancePayTermRows(rows), nil
}

func mapAdvancePayTermRows(rows []sqlc.BillingListActiveTermsForGenerationRow) []domain.AdvancePayTermRow {
	out := make([]domain.AdvancePayTermRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.AdvancePayTermRow{
			TermID:                 pgtypeUUIDToUUID(row.TermID),
			TenantID:               pgtypeUUIDToUUID(row.TenantID),
			BranchID:               pgtypeUUIDToUUID(row.BranchID),
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			TermStartDate:          pgtypeDateToTime(row.TermStartDate),
			TermEndDate:            pgtypeDateToTime(row.TermEndDate),
			BookingPatternID:       pgtypeUUIDToUUID(row.BookingPatternID),
			SiteHourlyRateMinor:    int(row.SiteHourlyRateMinor),
			Status:                 row.Status,
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			HasGuardianLink:        row.HasGuardianLink,
			FundingProfileID:       pgtypeUUIDToUUIDPtr(row.FundingProfileID),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
		})
	}
	return out
}

func (r *Repository) ListBookingPatternEntries(ctx context.Context, tx domain.Tx, tenantID, branchID, patternID uuid.UUID) ([]domain.BookingPatternEntryRow, error) {
	// We deliberately query the child_booking_pattern_entries + session_types tables directly
	// rather than going through the children module's repository, to keep the billing
	// module independent of the children module's application layer.
	rows, err := r.queriesTx(tx).ChildBookingPatternEntriesListByPattern(ctx, sqlc.ChildBookingPatternEntriesListByPatternParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		PatternID: uuidToPgtype(patternID),
	})
	if err != nil {
		return nil, fmt.Errorf("list booking pattern entries: %w", err)
	}
	out := make([]domain.BookingPatternEntryRow, 0, len(rows))
	for _, row := range rows {
		startMin := timeOfDayToMinutes(row.SessionTypeStartTime)
		endMin := timeOfDayToMinutes(row.SessionTypeEndTime)
		if endMin <= startMin {
			continue
		}
		out = append(out, domain.BookingPatternEntryRow{
			DayOfWeek:       int(row.DayOfWeek),
			SessionTypeID:   pgtypeUUIDToUUID(row.SessionTypeID),
			SessionTypeName: row.SessionTypeName,
			StartMinutes:    startMin,
			EndMinutes:      endMin,
		})
	}
	return out, nil
}

// timeOfDayToMinutes converts a pgtype.Time (microseconds since midnight) to
// minutes since midnight, rounded down.
func timeOfDayToMinutes(t pgtype.Time) int {
	return int(t.Microseconds / (60 * 1_000_000))
}

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

// --- Manager Invoice Review (API-18) read-only methods ---

func (r *Repository) ListInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceListForManagerReview(ctx, sqlc.InvoiceListForManagerReviewParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		BillingMonth: timeToPgtypeDatePtr(filters.BillingMonth),
		Status:       strToPgtypeTextPtr(filters.Status),
		ChildID:      uuidToPgtypePtr(filters.ChildID),
		Limit:        pgtype.Int4{Int32: int32(filters.Limit), Valid: true},
		Offset:       pgtype.Int4{Int32: int32(filters.Offset), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.InvoiceReviewRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapInvoiceReviewRow(row))
	}
	return result, nil
}

func (r *Repository) GetInvoiceForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.InvoiceGetForManagerReview(ctx, sqlc.InvoiceGetForManagerReviewParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(invoiceID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.InvoiceReviewRow{}, false, nil
		}
		return domain.InvoiceReviewRow{}, false, err
	}
	return mapInvoiceReviewRowFromGet(row), true, nil
}

func (r *Repository) ListInvoiceLinesForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceLinesForManagerReview(ctx, sqlc.InvoiceLinesForManagerReviewParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		InvoiceID: uuidToPgtype(invoiceID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.InvoiceReviewLineRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceReviewLineRow{
			ID:                     pgtypeUUIDToUUID(row.ID),
			LineKind:               row.LineKind,
			Description:            row.Description,
			SortOrder:              int(row.SortOrder),
			QuantityMinutes:        pgtypeInt4ToIntPtr(row.QuantityMinutes),
			UnitAmountMinor:        pgtypeInt4ToIntPtr(row.UnitAmountMinor),
			LineAmountMinor:        int(row.LineAmountMinor),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			FundedDeductionMinutes: pgtypeInt4ToIntPtr(row.FundedDeductionMinutes),
			CoreBillableMinutes:    pgtypeInt4ToIntPtr(row.CoreBillableMinutes),
			SessionCount:           pgtypeInt4ToIntPtr(row.SessionCount),
		})
	}
	return result, nil
}

func mapInvoiceReviewRow(row sqlc.InvoiceListForManagerReviewRow) domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:                      pgtypeUUIDToUUID(row.ID),
		InvoiceKind:             row.InvoiceKind,
		InvoiceNumber:           pgtypeTextToStrPtr(row.InvoiceNumber),
		Status:                  row.Status,
		ChildID:                 pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:          row.ChildFirstName,
		ChildMiddleName:         pgtypeTextToStrPtr(row.ChildMiddleName),
		ChildLastName:           pgtypeTextToStrPtr(row.ChildLastName),
		BillingMonth:            pgtypeDateToTime(row.BillingMonth),
		PeriodStartDate:         pgtypeDateToTime(row.PeriodStartDate),
		PeriodEndDate:           pgtypeDateToTime(row.PeriodEndDate),
		CurrencyCode:            row.CurrencyCode,
		SubtotalMinor:           int(row.SubtotalMinor),
		FundedDeductionMinor:    int(row.FundedDeductionMinor),
		TotalDueMinor:           int(row.TotalDueMinor),
		AmountPaidMinor:         int(row.AmountPaidMinor),
		DueAt:                   pgtypeTimestamptzToTimePtr(row.DueAt),
		IssuedAt:                pgtypeTimestamptzToTimePtr(row.IssuedAt),
		LockedAt:                pgtypeTimestamptzToTimePtr(row.LockedAt),
		PaidAt:                  pgtypeTimestamptzToTimePtr(row.PaidAt),
		PaymentFailedAt:         pgtypeTimestamptzToTimePtr(row.PaymentFailedAt),
		PaymentStatusUpdatedAt:  pgtypeTimestamptzToTimePtr(row.PaymentStatusUpdatedAt),
		AdjustsInvoiceID:        pgtypeUUIDToUUIDPtr(row.AdjustsInvoiceID),
		AdjustmentReasonCode:    pgtypeTextToStrPtr(row.AdjustmentReasonCode),
		AdjustmentReasonNote:    pgtypeTextToStrPtr(row.AdjustmentReasonNote),
		GeneratedRunID:          pgtypeUUIDToUUIDPtr(row.GeneratedRunID),
		GeneratedRunStatus:      pgtypeTextToStrPtr(row.GeneratedRunStatus),
		GeneratedRunStartedAt:   pgtypeTimestamptzToTimePtr(row.GeneratedRunStartedAt),
		GeneratedRunCompletedAt: pgtypeTimestamptzToTimePtr(row.GeneratedRunCompletedAt),
		GeneratedRunDetails:     json.RawMessage(row.GeneratedRunDetails),
		CalculationDetails:      json.RawMessage(row.CalculationDetails),
		CreatedAt:               pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:               pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

func mapInvoiceReviewRowFromGet(row sqlc.InvoiceGetForManagerReviewRow) domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:                      pgtypeUUIDToUUID(row.ID),
		InvoiceKind:             row.InvoiceKind,
		InvoiceNumber:           pgtypeTextToStrPtr(row.InvoiceNumber),
		Status:                  row.Status,
		ChildID:                 pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:          row.ChildFirstName,
		ChildMiddleName:         pgtypeTextToStrPtr(row.ChildMiddleName),
		ChildLastName:           pgtypeTextToStrPtr(row.ChildLastName),
		BillingMonth:            pgtypeDateToTime(row.BillingMonth),
		PeriodStartDate:         pgtypeDateToTime(row.PeriodStartDate),
		PeriodEndDate:           pgtypeDateToTime(row.PeriodEndDate),
		CurrencyCode:            row.CurrencyCode,
		SubtotalMinor:           int(row.SubtotalMinor),
		FundedDeductionMinor:    int(row.FundedDeductionMinor),
		TotalDueMinor:           int(row.TotalDueMinor),
		AmountPaidMinor:         int(row.AmountPaidMinor),
		DueAt:                   pgtypeTimestamptzToTimePtr(row.DueAt),
		IssuedAt:                pgtypeTimestamptzToTimePtr(row.IssuedAt),
		LockedAt:                pgtypeTimestamptzToTimePtr(row.LockedAt),
		PaidAt:                  pgtypeTimestamptzToTimePtr(row.PaidAt),
		PaymentFailedAt:         pgtypeTimestamptzToTimePtr(row.PaymentFailedAt),
		PaymentStatusUpdatedAt:  pgtypeTimestamptzToTimePtr(row.PaymentStatusUpdatedAt),
		AdjustsInvoiceID:        pgtypeUUIDToUUIDPtr(row.AdjustsInvoiceID),
		AdjustmentReasonCode:    pgtypeTextToStrPtr(row.AdjustmentReasonCode),
		AdjustmentReasonNote:    pgtypeTextToStrPtr(row.AdjustmentReasonNote),
		GeneratedRunID:          pgtypeUUIDToUUIDPtr(row.GeneratedRunID),
		GeneratedRunStatus:      pgtypeTextToStrPtr(row.GeneratedRunStatus),
		GeneratedRunStartedAt:   pgtypeTimestamptzToTimePtr(row.GeneratedRunStartedAt),
		GeneratedRunCompletedAt: pgtypeTimestamptzToTimePtr(row.GeneratedRunCompletedAt),
		GeneratedRunDetails:     json.RawMessage(row.GeneratedRunDetails),
		CalculationDetails:      json.RawMessage(row.CalculationDetails),
		CreatedAt:               pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:               pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Parent Invoice View (API-21) read-only methods ---

func (r *Repository) ListInvoicesForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID, filters domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceListForParent(ctx, sqlc.InvoiceListForParentParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		ID:           uuidToPgtype(membershipID),
		BillingMonth: timeToPgtypeDatePtr(filters.BillingMonth),
		Status:       strToPgtypeTextPtr(filters.Status),
		ChildID:      uuidToPgtypePtr(filters.ChildID),
		Limit:        pgtype.Int4{Int32: int32(filters.Limit), Valid: true},
		Offset:       pgtype.Int4{Int32: int32(filters.Offset), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.ParentInvoiceRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapParentInvoiceRow(row.ID, row.InvoiceKind, row.InvoiceNumber, row.Status, row.ChildID, row.ChildFirstName, row.ChildMiddleName, row.ChildLastName, row.BillingMonth, row.PeriodStartDate, row.PeriodEndDate, row.CurrencyCode, row.SubtotalMinor, row.FundedDeductionMinor, row.TotalDueMinor, row.AmountPaidMinor, row.DueAt, row.IssuedAt, row.PaidAt, row.PaymentFailedAt, row.PaymentStatusUpdatedAt, row.CalculationDetails))
	}
	return result, nil
}

func (r *Repository) GetInvoiceForParent(ctx context.Context, tenantID, branchID, membershipID, invoiceID uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.InvoiceGetForParent(ctx, sqlc.InvoiceGetForParentParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(membershipID),
		ID_2:     uuidToPgtype(invoiceID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ParentInvoiceRow{}, false, nil
		}
		return domain.ParentInvoiceRow{}, false, err
	}
	return mapParentInvoiceRow(row.ID, row.InvoiceKind, row.InvoiceNumber, row.Status, row.ChildID, row.ChildFirstName, row.ChildMiddleName, row.ChildLastName, row.BillingMonth, row.PeriodStartDate, row.PeriodEndDate, row.CurrencyCode, row.SubtotalMinor, row.FundedDeductionMinor, row.TotalDueMinor, row.AmountPaidMinor, row.DueAt, row.IssuedAt, row.PaidAt, row.PaymentFailedAt, row.PaymentStatusUpdatedAt, row.CalculationDetails), true, nil
}

func (r *Repository) ListInvoiceLinesForParent(ctx context.Context, tenantID, branchID, membershipID, invoiceID uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceLinesForParent(ctx, sqlc.InvoiceLinesForParentParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ID:        uuidToPgtype(membershipID),
		InvoiceID: uuidToPgtype(invoiceID),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.ParentInvoiceLineRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ParentInvoiceLineRow{
			LineKind:        row.LineKind,
			Description:     row.Description,
			SortOrder:       int(row.SortOrder),
			QuantityMinutes: pgtypeInt4ToIntPtr(row.QuantityMinutes),
			UnitAmountMinor: pgtypeInt4ToIntPtr(row.UnitAmountMinor),
			LineAmountMinor: int(row.LineAmountMinor),
		})
	}
	return result, nil
}

func mapParentInvoiceRow(id pgtype.UUID, invoiceKind string, invoiceNumber pgtype.Text, status string, childID pgtype.UUID, childFirstName string, childMiddleName pgtype.Text, childLastName pgtype.Text, billingMonth pgtype.Date, periodStartDate pgtype.Date, periodEndDate pgtype.Date, currencyCode string, subtotalMinor int32, fundedDeductionMinor int32, totalDueMinor int32, amountPaidMinor int32, dueAt pgtype.Timestamptz, issuedAt pgtype.Timestamptz, paidAt pgtype.Timestamptz, paymentFailedAt pgtype.Timestamptz, paymentStatusUpdatedAt pgtype.Timestamptz, calculationDetails []byte) domain.ParentInvoiceRow {
	return domain.ParentInvoiceRow{
		ID:                     pgtypeUUIDToUUID(id),
		InvoiceKind:            invoiceKind,
		InvoiceNumber:          pgtypeTextToStrPtr(invoiceNumber),
		Status:                 status,
		ChildID:                pgtypeUUIDToUUID(childID),
		ChildFirstName:         childFirstName,
		ChildMiddleName:        pgtypeTextToStrPtr(childMiddleName),
		ChildLastName:          pgtypeTextToStrPtr(childLastName),
		BillingMonth:           pgtypeDateToTime(billingMonth),
		PeriodStartDate:        pgtypeDateToTime(periodStartDate),
		PeriodEndDate:          pgtypeDateToTime(periodEndDate),
		CurrencyCode:           currencyCode,
		SubtotalMinor:          int(subtotalMinor),
		FundedDeductionMinor:   int(fundedDeductionMinor),
		TotalDueMinor:          int(totalDueMinor),
		AmountPaidMinor:        int(amountPaidMinor),
		DueAt:                  pgtypeTimestamptzToTimePtr(dueAt),
		IssuedAt:               pgtypeTimestamptzToTimePtr(issuedAt),
		PaidAt:                 pgtypeTimestamptzToTimePtr(paidAt),
		PaymentFailedAt:        pgtypeTimestamptzToTimePtr(paymentFailedAt),
		PaymentStatusUpdatedAt: pgtypeTimestamptzToTimePtr(paymentStatusUpdatedAt),
		CalculationDetails:     json.RawMessage(calculationDetails),
	}
}

// --- Invoice Issue (API-19) transactional methods ---

func (r *Repository) GetInvoiceForIssueForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	row, err := r.queriesTx(tx).GetInvoiceForIssueForUpdate(ctx, sqlc.GetInvoiceForIssueForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(invoiceID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.InvoiceIssueCandidateRow{}, false, nil
		}
		return domain.InvoiceIssueCandidateRow{}, false, err
	}
	return mapIssueCandidateRow(row), true, nil
}

func (r *Repository) ListDraftInvoicesForIssueForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	rows, err := r.queriesTx(tx).ListDraftInvoicesForIssueForUpdate(ctx, sqlc.ListDraftInvoicesForIssueForUpdateParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return nil, err
	}
	return mapIssueCandidateRows(rows), nil
}

func (r *Repository) ListSelectedInvoicesForIssueForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, invoiceIDs []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	pgIDs := make([]pgtype.UUID, len(invoiceIDs))
	for i, id := range invoiceIDs {
		pgIDs[i] = uuidToPgtype(id)
	}
	rows, err := r.queriesTx(tx).ListSelectedInvoicesForIssueForUpdate(ctx, sqlc.ListSelectedInvoicesForIssueForUpdateParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		Column3:  pgIDs,
	})
	if err != nil {
		return nil, err
	}
	return mapSelectedIssueCandidateRows(rows), nil
}

func (r *Repository) AllocateInvoiceNumberSequence(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, year, month int) (int, error) {
	seq, err := r.queriesTx(tx).AllocateInvoiceNumberSequence(ctx, sqlc.AllocateInvoiceNumberSequenceParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		BillingYear:  int32(year),
		BillingMonth: int32(month),
	})
	if err != nil {
		return 0, err
	}
	return int(seq), nil
}

func (r *Repository) MarkInvoiceIssued(ctx context.Context, tx domain.Tx, params domain.IssueInvoiceUpdateParams) (int64, error) {
	n, err := r.queriesTx(tx).MarkInvoiceIssued(ctx, sqlc.MarkInvoiceIssuedParams{
		ID:                   uuidToPgtype(params.ID),
		TenantID:             uuidToPgtype(params.TenantID),
		BranchID:             uuidToPgtype(params.BranchID),
		InvoiceNumber:        pgtype.Text{String: params.InvoiceNumber, Valid: true},
		IssuedSequence:       pgtype.Int4{Int32: int32(params.IssuedSequence), Valid: true},
		IssuedRunID:          uuidToPgtype(params.IssuedRunID),
		IssuedAt:             pgtype.Timestamptz{Time: params.IssuedAt, Valid: true},
		IssuedByUserID:       uuidToPgtype(params.IssuedByUserID),
		IssuedByMembershipID: uuidToPgtype(params.IssuedByMembershipID),
		DueAt:                pgtype.Timestamptz{Time: params.DueAt, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

// --- Overdue Transition (API-20) transactional methods ---

func (r *Repository) TryAcquireOverdueTransitionJobLock(ctx context.Context, tx domain.Tx) (bool, error) {
	return r.queriesTx(tx).TryAcquireOverdueTransitionJobLock(ctx)
}

func (r *Repository) MarkIssuedInvoicesOverdue(ctx context.Context, tx domain.Tx, cutoffUTC time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	rows, err := r.queriesTx(tx).MarkIssuedInvoicesOverdue(ctx, pgtype.Timestamptz{Time: cutoffUTC, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]domain.OverdueTransitionedInvoice, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.OverdueTransitionedInvoice{
			ID:       pgtypeUUIDToUUID(row.ID),
			TenantID: pgtypeUUIDToUUID(row.TenantID),
			BranchID: pgtypeUUIDToUUID(row.BranchID),
		})
	}
	return result, nil
}

func mapIssueCandidateRow(row sqlc.GetInvoiceForIssueForUpdateRow) domain.InvoiceIssueCandidateRow {
	return domain.InvoiceIssueCandidateRow{
		ID:              pgtypeUUIDToUUID(row.ID),
		ChildID:         pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:  row.ChildFirstName,
		ChildMiddleName: pgtypeTextToStrPtr(row.ChildMiddleName),
		ChildLastName:   pgtypeTextToStrPtr(row.ChildLastName),
		BillingMonth:    pgtypeDateToTime(row.BillingMonth),
		InvoiceKind:     row.InvoiceKind,
		Status:          row.Status,
		TotalDueMinor:   int(row.TotalDueMinor),
	}
}

func mapIssueCandidateRows(rows []sqlc.ListDraftInvoicesForIssueForUpdateRow) []domain.InvoiceIssueCandidateRow {
	result := make([]domain.InvoiceIssueCandidateRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceIssueCandidateRow{
			ID:              pgtypeUUIDToUUID(row.ID),
			ChildID:         pgtypeUUIDToUUID(row.ChildID),
			ChildFirstName:  row.ChildFirstName,
			ChildMiddleName: pgtypeTextToStrPtr(row.ChildMiddleName),
			ChildLastName:   pgtypeTextToStrPtr(row.ChildLastName),
			BillingMonth:    pgtypeDateToTime(row.BillingMonth),
			InvoiceKind:     row.InvoiceKind,
			Status:          row.Status,
			TotalDueMinor:   int(row.TotalDueMinor),
		})
	}
	return result
}

func mapSelectedIssueCandidateRows(rows []sqlc.ListSelectedInvoicesForIssueForUpdateRow) []domain.InvoiceIssueCandidateRow {
	result := make([]domain.InvoiceIssueCandidateRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceIssueCandidateRow{
			ID:              pgtypeUUIDToUUID(row.ID),
			ChildID:         pgtypeUUIDToUUID(row.ChildID),
			ChildFirstName:  row.ChildFirstName,
			ChildMiddleName: pgtypeTextToStrPtr(row.ChildMiddleName),
			ChildLastName:   pgtypeTextToStrPtr(row.ChildLastName),
			BillingMonth:    pgtypeDateToTime(row.BillingMonth),
			InvoiceKind:     row.InvoiceKind,
			Status:          row.Status,
			TotalDueMinor:   int(row.TotalDueMinor),
		})
	}
	return result
}

// --- Helpers ---

func mapCandidateRows(rows []sqlc.ListCandidateChildrenForUpdateRow) []domain.PreflightChildRow {
	result := make([]domain.PreflightChildRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PreflightChildRow{
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
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
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
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

func timeToPgtypeDatePtr(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func strToPgtypeTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func uuidToPgtypePtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(*u), Valid: true}
}
