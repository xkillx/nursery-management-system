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

func (r *Repository) ListPreflightChildren(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]PreflightChildRow, error) {
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

	result := make([]PreflightChildRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, PreflightChildRow{
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
			HasParentCarerContact:  row.HasParentCarerContact,
			FundingProfileID:       pgtypeUUIDToUUIDPtr(row.FundingProfileID),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			ExistingInvoiceID:      pgtypeUUIDToUUIDPtr(row.ExistingInvoiceID),
			ExistingInvoiceStatus:  pgtypeTextToStrPtr(row.ExistingInvoiceStatus),
		})
	}
	return result, nil
}

func (r *Repository) ListPreflightAttendanceSessions(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]PreflightAttendanceSessionRow, error) {
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

	result := make([]PreflightAttendanceSessionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, PreflightAttendanceSessionRow{
			SessionID:         pgtypeUUIDToUUID(row.ID),
			ChildID:           pgtypeUUIDToUUID(row.ChildID),
			Status:            row.Status,
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
			HasParentCarerContact:  row.HasParentCarerContact,
			FundingProfileID:       pgtypeUUIDToUUIDPtr(row.FundingProfileID),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			TermTimeOnly:           row.TermTimeOnly,
			FundingModel:           row.FundingModel,
			FundedHoursPerWeek:     pgtypeNumericToFloat64Ptr(row.FundedHoursPerWeek),
			AdHocRateMultiplier:    pgtypeNumericToFloat64(row.AdHocRateMultiplier),
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
			SessionTypeKind: row.SessionTypeKind,
		})
	}
	return out, nil
}

func (r *Repository) ListActiveAdHocBookingsForChildInMonth(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]domain.AdHocBookingRow, error) {
	var q *sqlc.Queries
	if tx != nil {
		q = r.queriesTx(tx)
	} else {
		q = sqlc.New(r.pool)
	}
	rows, err := q.BillingListAdHocBookingsForMonth(ctx, sqlc.BillingListAdHocBookingsForMonthParams{
		TenantID:       uuidToPgtype(tenantID),
		BranchID:       uuidToPgtype(branchID),
		ChildID:        uuidToPgtype(childID),
		CalendarDate:   timeToPgtypeDate(from),
		CalendarDate_2: timeToPgtypeDate(to),
	})
	if err != nil {
		return nil, fmt.Errorf("list ad-hoc bookings for month: %w", err)
	}
	out := make([]domain.AdHocBookingRow, 0, len(rows))
	for _, row := range rows {
		startMin := timeOfDayToMinutes(row.SessionTypeStartTime)
		endMin := timeOfDayToMinutes(row.SessionTypeEndTime)
		if endMin <= startMin {
			continue
		}
		out = append(out, domain.AdHocBookingRow{
			ID:              pgtypeUUIDToUUID(row.ID),
			ChildID:         pgtypeUUIDToUUID(row.ChildID),
			CalendarDate:    pgtypeDateToTime(row.CalendarDate),
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

func (r *Repository) queriesTx(tx domain.Tx) *sqlc.Queries {
	return sqlc.New(tx.(pgx.Tx))
}

func (r *Repository) ListCandidateChildrenForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, billingMonth, nextBillingMonth time.Time) ([]PreflightChildRow, error) {
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

func (r *Repository) ListSelectedChildrenForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, childIDs []uuid.UUID) ([]PreflightChildRow, error) {
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

func (r *Repository) ListAttendanceSessions(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, periodStart, periodEndExclusive time.Time) ([]PreflightAttendanceSessionRow, error) {
	rows, err := r.queriesTx(tx).ListAttendanceSessionsForGeneration(ctx, sqlc.ListAttendanceSessionsForGenerationParams{
		TenantID:           uuidToPgtype(tenantID),
		BranchID:           uuidToPgtype(branchID),
		CheckInLocalDate:   timeToPgtypeDate(periodStart),
		CheckInLocalDate_2: timeToPgtypeDate(periodEndExclusive),
	})
	if err != nil {
		return nil, err
	}

	result := make([]PreflightAttendanceSessionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, PreflightAttendanceSessionRow{
			SessionID:         pgtypeUUIDToUUID(row.ID),
			ChildID:           pgtypeUUIDToUUID(row.ChildID),
			Status:            row.Status,
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
		SubtotalMinor:        int32(params.Subtotal.Minor()),
		FundedDeductionMinor: int32(params.FundedDeduction.Minor()),
		TotalDueMinor:        int32(params.TotalDue.Minor()),
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
		SubtotalMinor:        int32(params.Subtotal.Minor()),
		FundedDeductionMinor: int32(params.FundedDeduction.Minor()),
		TotalDueMinor:        int32(params.TotalDue.Minor()),
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
		UnitAmountMinor:        pgtypeInt4OrNil(params.UnitAmount.Minor()),
		LineAmountMinor:        int32(params.LineAmount.Minor()),
		RawAttendedMinutes:     pgtypeInt4OrNil(params.RawAttendedMinutes),
		RoundedAttendedMinutes: pgtypeInt4OrNil(params.RoundedAttendedMinutes),
		FundedAllowanceMinutes: pgtypeInt4OrNil(params.FundedAllowanceMinutes),
		FundedDeductionMinutes: pgtypeInt4OrNil(params.FundedDeductionMinutes),
		CoreBillableMinutes:    pgtypeInt4OrNil(params.CoreBillableMinutes),
		SessionCount:           pgtypeInt4OrNil(params.SessionCount),
		Details:                params.Details,
	})
}

func (r *Repository) GetInvoiceLine(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID, lineID uuid.UUID) (domain.InvoiceLine, bool, error) {
	row, err := r.queriesTx(tx).InvoiceLineGet(ctx, sqlc.InvoiceLineGetParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		InvoiceID: uuidToPgtype(invoiceID),
		ID:        uuidToPgtype(lineID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.InvoiceLine{}, false, nil
		}
		return domain.InvoiceLine{}, false, err
	}
	return domain.InvoiceLine{
		ID:                     pgtypeUUIDToUUID(row.ID),
		LineKind:               row.LineKind,
		Description:            row.Description,
		SortOrder:              int(row.SortOrder),
		QuantityMinutes:        pgtypeInt4ToInt(row.QuantityMinutes),
		UnitAmount:             pgtypeInt4ToMoney(row.UnitAmountMinor),
		LineAmount:             domain.MustGBP(int(row.LineAmountMinor)),
		FundedAllowanceMinutes: pgtypeInt4ToInt(row.FundedAllowanceMinutes),
		FundedDeductionMinutes: pgtypeInt4ToInt(row.FundedDeductionMinutes),
		CoreBillableMinutes:    pgtypeInt4ToInt(row.CoreBillableMinutes),
		SessionCount:           pgtypeInt4ToInt(row.SessionCount),
	}, true, nil
}

func (r *Repository) UpdateInvoiceLine(ctx context.Context, tx domain.Tx, tenantID, branchID, lineID uuid.UUID, description string, quantityMinutes int, unitAmount, lineAmount domain.Money) (int64, error) {
	n, err := r.queriesTx(tx).InvoiceLineUpdate(ctx, sqlc.InvoiceLineUpdateParams{
		ID:              uuidToPgtype(lineID),
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		Description:     description,
		QuantityMinutes: pgtypeInt4OrNil(quantityMinutes),
		UnitAmountMinor: pgtypeInt4OrNil(unitAmount.Minor()),
		LineAmountMinor: int32(lineAmount.Minor()),
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *Repository) DeleteInvoiceLine(ctx context.Context, tx domain.Tx, tenantID, branchID, lineID uuid.UUID) (int64, error) {
	n, err := r.queriesTx(tx).InvoiceLineDelete(ctx, sqlc.InvoiceLineDeleteParams{
		ID:       uuidToPgtype(lineID),
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}

// --- Manager Invoice Review (API-18) read-only methods ---

func (r *Repository) ListInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	q := sqlc.New(r.pool)
	pgTenant := uuidToPgtype(tenantID)
	pgBranch := uuidToPgtype(branchID)
	pgBillingMonth := timeToPgtypeDatePtr(filters.BillingMonth)
	pgBillingMonthFrom := timeToPgtypeDatePtr(filters.BillingMonthFrom)
	pgBillingMonthTo := timeToPgtypeDatePtr(filters.BillingMonthTo)
	pgStatuses := strSliceToPgtypeTextArray(filters.Statuses)
	pgChildID := uuidToPgtypePtr(filters.ChildID)
	pgSearch := strToPgtypeTextPtr(filters.Search)
	pgLimit := pgtype.Int4{Int32: int32(filters.Limit), Valid: true}
	pgOffset := pgtype.Int4{Int32: int32(filters.Offset), Valid: true}

	switch filters.SortField + ":" + filters.SortDir {
	case "billing_month:asc":
		rows, err := q.InvoiceListForManagerReviewSortByBillingMonthAsc(ctx, sqlc.InvoiceListForManagerReviewSortByBillingMonthAscParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]domain.InvoiceReviewRow, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapInvoiceReviewRowSort(row))
		}
		return result, nil
	case "due_at:asc":
		rows, err := q.InvoiceListForManagerReviewSortByDueAtAsc(ctx, sqlc.InvoiceListForManagerReviewSortByDueAtAscParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]domain.InvoiceReviewRow, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapInvoiceReviewRowSort(row))
		}
		return result, nil
	case "due_at:desc":
		rows, err := q.InvoiceListForManagerReviewSortByDueAtDesc(ctx, sqlc.InvoiceListForManagerReviewSortByDueAtDescParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]domain.InvoiceReviewRow, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapInvoiceReviewRowSort(row))
		}
		return result, nil
	case "total_amount:asc":
		rows, err := q.InvoiceListForManagerReviewSortByTotalAmountAsc(ctx, sqlc.InvoiceListForManagerReviewSortByTotalAmountAscParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]domain.InvoiceReviewRow, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapInvoiceReviewRowSort(row))
		}
		return result, nil
	case "total_amount:desc":
		rows, err := q.InvoiceListForManagerReviewSortByTotalAmountDesc(ctx, sqlc.InvoiceListForManagerReviewSortByTotalAmountDescParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
		})
		if err != nil {
			return nil, err
		}
		result := make([]domain.InvoiceReviewRow, 0, len(rows))
		for _, row := range rows {
			result = append(result, mapInvoiceReviewRowSort(row))
		}
		return result, nil
	default:
		rows, err := q.InvoiceListForManagerReview(ctx, sqlc.InvoiceListForManagerReviewParams{
			TenantID: pgTenant, BranchID: pgBranch, BillingMonth: pgBillingMonth, BillingMonthFrom: pgBillingMonthFrom, BillingMonthTo: pgBillingMonthTo, Statuses: pgStatuses, ChildID: pgChildID, Search: pgSearch, Limit: pgLimit, Offset: pgOffset,
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
}

func (r *Repository) CountInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceReviewFilters) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.InvoiceCountForManagerReview(ctx, sqlc.InvoiceCountForManagerReviewParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		BillingMonth:     timeToPgtypeDatePtr(filters.BillingMonth),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
		Statuses:         strSliceToPgtypeTextArray(filters.Statuses),
		ChildID:          uuidToPgtypePtr(filters.ChildID),
		Search:           strToPgtypeTextPtr(filters.Search),
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
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
		line := domain.InvoiceReviewLineRow{
			ID:                     pgtypeUUIDToUUID(row.ID),
			LineKind:               row.LineKind,
			Description:            row.Description,
			SortOrder:              int(row.SortOrder),
			QuantityMinutes:        pgtypeInt4ToIntPtr(row.QuantityMinutes),
			UnitAmount:             pgtypeInt4ToMoneyPtr(row.UnitAmountMinor),
			LineAmount:             domain.MustGBP(int(row.LineAmountMinor)),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			FundedDeductionMinutes: pgtypeInt4ToIntPtr(row.FundedDeductionMinutes),
			CoreBillableMinutes:    pgtypeInt4ToIntPtr(row.CoreBillableMinutes),
			SessionCount:           pgtypeInt4ToIntPtr(row.SessionCount),
		}
		if len(row.Details) > 0 {
			var details struct {
				FundingModel string `json:"funding_model"`
			}
			if json.Unmarshal(row.Details, &details) == nil && details.FundingModel != "" {
				line.FundingModel = &details.FundingModel
			}
		}
		result = append(result, line)
	}
	return result, nil
}

func mapInvoiceReviewRow(row sqlc.InvoiceListForManagerReviewRow) domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:                            pgtypeUUIDToUUID(row.ID),
		InvoiceKind:                   row.InvoiceKind,
		InvoiceNumber:                 pgtypeTextToStrPtr(row.InvoiceNumber),
		Status:                        row.Status,
		ChildID:                       pgtypeUUIDToUUID(row.ChildID),
		ChildFirstName:                row.ChildFirstName,
		ChildMiddleName:               pgtypeTextToStrPtr(row.ChildMiddleName),
		ChildLastName:                 pgtypeTextToStrPtr(row.ChildLastName),
		BillingMonth:                  pgtypeDateToTime(row.BillingMonth),
		PeriodStartDate:               pgtypeDateToTime(row.PeriodStartDate),
		PeriodEndDate:                 pgtypeDateToTime(row.PeriodEndDate),
		CurrencyCode:                  row.CurrencyCode,
		Subtotal:                      domain.MustGBP(int(row.SubtotalMinor)),
		FundedDeduction:               domain.MustGBP(int(row.FundedDeductionMinor)),
		TotalDue:                      domain.MustGBP(int(row.TotalDueMinor)),
		AmountPaid:                    domain.MustGBP(int(row.AmountPaidMinor)),
		DueAt:                         pgtypeTimestamptzToTimePtr(row.DueAt),
		IssuedAt:                      pgtypeTimestamptzToTimePtr(row.IssuedAt),
		LockedAt:                      pgtypeTimestamptzToTimePtr(row.LockedAt),
		PaidAt:                        pgtypeTimestamptzToTimePtr(row.PaidAt),
		PaymentFailedAt:               pgtypeTimestamptzToTimePtr(row.PaymentFailedAt),
		PaymentStatusUpdatedAt:        pgtypeTimestamptzToTimePtr(row.PaymentStatusUpdatedAt),
		AdjustsInvoiceID:              pgtypeUUIDToUUIDPtr(row.AdjustsInvoiceID),
		AdjustmentReasonCode:          pgtypeTextToStrPtr(row.AdjustmentReasonCode),
		AdjustmentReasonNote:          pgtypeTextToStrPtr(row.AdjustmentReasonNote),
		GeneratedRunID:                pgtypeUUIDToUUIDPtr(row.GeneratedRunID),
		GeneratedRunStatus:            pgtypeTextToStrPtr(row.GeneratedRunStatus),
		GeneratedRunStartedAt:         pgtypeTimestamptzToTimePtr(row.GeneratedRunStartedAt),
		GeneratedRunCompletedAt:       pgtypeTimestamptzToTimePtr(row.GeneratedRunCompletedAt),
		GeneratedRunDetails:           json.RawMessage(row.GeneratedRunDetails),
		CalculationDetails:            json.RawMessage(row.CalculationDetails),
		RoomName:                      nil,
		ChildPhotoPath:                pgtypeTextToStrPtr(row.ChildProfilePhotoPath),
		CreatedAt:                     pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                     pgtypeTimestamptzToTime(row.UpdatedAt),
		LatestPaymentAttemptStatus:    row.LatestPaymentAttemptStatus,
		LatestPaymentAttemptCreatedAt: pgtypeTimestamptzToTimePtr(row.LatestPaymentAttemptCreatedAt),
	}
}

func mapInvoiceReviewRowSort(row interface{}) domain.InvoiceReviewRow {
	type fields struct {
		ID                            pgtype.UUID
		InvoiceKind                   string
		InvoiceNumber                 pgtype.Text
		Status                        string
		ChildID                       pgtype.UUID
		ChildFirstName                string
		ChildMiddleName               pgtype.Text
		ChildLastName                 pgtype.Text
		BillingMonth                  pgtype.Date
		PeriodStartDate               pgtype.Date
		PeriodEndDate                 pgtype.Date
		CurrencyCode                  string
		SubtotalMinor                 int32
		FundedDeductionMinor          int32
		TotalDueMinor                 int32
		AmountPaidMinor               int32
		DueAt                         pgtype.Timestamptz
		IssuedAt                      pgtype.Timestamptz
		LockedAt                      pgtype.Timestamptz
		PaidAt                        pgtype.Timestamptz
		PaymentFailedAt               pgtype.Timestamptz
		PaymentStatusUpdatedAt        pgtype.Timestamptz
		AdjustsInvoiceID              pgtype.UUID
		AdjustmentReasonCode          pgtype.Text
		AdjustmentReasonNote          pgtype.Text
		GeneratedRunID                pgtype.UUID
		GeneratedRunStatus            pgtype.Text
		GeneratedRunStartedAt         pgtype.Timestamptz
		GeneratedRunCompletedAt       pgtype.Timestamptz
		GeneratedRunDetails           []byte
		CalculationDetails            []byte
		ChildProfilePhotoPath         pgtype.Text
		CreatedAt                     pgtype.Timestamptz
		UpdatedAt                     pgtype.Timestamptz
		LatestPaymentAttemptStatus    string
		LatestPaymentAttemptCreatedAt pgtype.Timestamptz
	}
	var f fields
	switch v := row.(type) {
	case sqlc.InvoiceListForManagerReviewSortByBillingMonthAscRow:
		f = fields{v.ID, v.InvoiceKind, v.InvoiceNumber, v.Status, v.ChildID, v.ChildFirstName, v.ChildMiddleName, v.ChildLastName, v.BillingMonth, v.PeriodStartDate, v.PeriodEndDate, v.CurrencyCode, v.SubtotalMinor, v.FundedDeductionMinor, v.TotalDueMinor, v.AmountPaidMinor, v.DueAt, v.IssuedAt, v.LockedAt, v.PaidAt, v.PaymentFailedAt, v.PaymentStatusUpdatedAt, v.AdjustsInvoiceID, v.AdjustmentReasonCode, v.AdjustmentReasonNote, v.GeneratedRunID, v.GeneratedRunStatus, v.GeneratedRunStartedAt, v.GeneratedRunCompletedAt, v.GeneratedRunDetails, v.CalculationDetails, v.ChildProfilePhotoPath, v.CreatedAt, v.UpdatedAt, v.LatestPaymentAttemptStatus, v.LatestPaymentAttemptCreatedAt}
	case sqlc.InvoiceListForManagerReviewSortByDueAtAscRow:
		f = fields{v.ID, v.InvoiceKind, v.InvoiceNumber, v.Status, v.ChildID, v.ChildFirstName, v.ChildMiddleName, v.ChildLastName, v.BillingMonth, v.PeriodStartDate, v.PeriodEndDate, v.CurrencyCode, v.SubtotalMinor, v.FundedDeductionMinor, v.TotalDueMinor, v.AmountPaidMinor, v.DueAt, v.IssuedAt, v.LockedAt, v.PaidAt, v.PaymentFailedAt, v.PaymentStatusUpdatedAt, v.AdjustsInvoiceID, v.AdjustmentReasonCode, v.AdjustmentReasonNote, v.GeneratedRunID, v.GeneratedRunStatus, v.GeneratedRunStartedAt, v.GeneratedRunCompletedAt, v.GeneratedRunDetails, v.CalculationDetails, v.ChildProfilePhotoPath, v.CreatedAt, v.UpdatedAt, v.LatestPaymentAttemptStatus, v.LatestPaymentAttemptCreatedAt}
	case sqlc.InvoiceListForManagerReviewSortByDueAtDescRow:
		f = fields{v.ID, v.InvoiceKind, v.InvoiceNumber, v.Status, v.ChildID, v.ChildFirstName, v.ChildMiddleName, v.ChildLastName, v.BillingMonth, v.PeriodStartDate, v.PeriodEndDate, v.CurrencyCode, v.SubtotalMinor, v.FundedDeductionMinor, v.TotalDueMinor, v.AmountPaidMinor, v.DueAt, v.IssuedAt, v.LockedAt, v.PaidAt, v.PaymentFailedAt, v.PaymentStatusUpdatedAt, v.AdjustsInvoiceID, v.AdjustmentReasonCode, v.AdjustmentReasonNote, v.GeneratedRunID, v.GeneratedRunStatus, v.GeneratedRunStartedAt, v.GeneratedRunCompletedAt, v.GeneratedRunDetails, v.CalculationDetails, v.ChildProfilePhotoPath, v.CreatedAt, v.UpdatedAt, v.LatestPaymentAttemptStatus, v.LatestPaymentAttemptCreatedAt}
	case sqlc.InvoiceListForManagerReviewSortByTotalAmountAscRow:
		f = fields{v.ID, v.InvoiceKind, v.InvoiceNumber, v.Status, v.ChildID, v.ChildFirstName, v.ChildMiddleName, v.ChildLastName, v.BillingMonth, v.PeriodStartDate, v.PeriodEndDate, v.CurrencyCode, v.SubtotalMinor, v.FundedDeductionMinor, v.TotalDueMinor, v.AmountPaidMinor, v.DueAt, v.IssuedAt, v.LockedAt, v.PaidAt, v.PaymentFailedAt, v.PaymentStatusUpdatedAt, v.AdjustsInvoiceID, v.AdjustmentReasonCode, v.AdjustmentReasonNote, v.GeneratedRunID, v.GeneratedRunStatus, v.GeneratedRunStartedAt, v.GeneratedRunCompletedAt, v.GeneratedRunDetails, v.CalculationDetails, v.ChildProfilePhotoPath, v.CreatedAt, v.UpdatedAt, v.LatestPaymentAttemptStatus, v.LatestPaymentAttemptCreatedAt}
	case sqlc.InvoiceListForManagerReviewSortByTotalAmountDescRow:
		f = fields{v.ID, v.InvoiceKind, v.InvoiceNumber, v.Status, v.ChildID, v.ChildFirstName, v.ChildMiddleName, v.ChildLastName, v.BillingMonth, v.PeriodStartDate, v.PeriodEndDate, v.CurrencyCode, v.SubtotalMinor, v.FundedDeductionMinor, v.TotalDueMinor, v.AmountPaidMinor, v.DueAt, v.IssuedAt, v.LockedAt, v.PaidAt, v.PaymentFailedAt, v.PaymentStatusUpdatedAt, v.AdjustsInvoiceID, v.AdjustmentReasonCode, v.AdjustmentReasonNote, v.GeneratedRunID, v.GeneratedRunStatus, v.GeneratedRunStartedAt, v.GeneratedRunCompletedAt, v.GeneratedRunDetails, v.CalculationDetails, v.ChildProfilePhotoPath, v.CreatedAt, v.UpdatedAt, v.LatestPaymentAttemptStatus, v.LatestPaymentAttemptCreatedAt}
	default:
		return domain.InvoiceReviewRow{}
	}
	return domain.InvoiceReviewRow{
		ID:                            pgtypeUUIDToUUID(f.ID),
		InvoiceKind:                   f.InvoiceKind,
		InvoiceNumber:                 pgtypeTextToStrPtr(f.InvoiceNumber),
		Status:                        f.Status,
		ChildID:                       pgtypeUUIDToUUID(f.ChildID),
		ChildFirstName:                f.ChildFirstName,
		ChildMiddleName:               pgtypeTextToStrPtr(f.ChildMiddleName),
		ChildLastName:                 pgtypeTextToStrPtr(f.ChildLastName),
		BillingMonth:                  pgtypeDateToTime(f.BillingMonth),
		PeriodStartDate:               pgtypeDateToTime(f.PeriodStartDate),
		PeriodEndDate:                 pgtypeDateToTime(f.PeriodEndDate),
		CurrencyCode:                  f.CurrencyCode,
		Subtotal:                      domain.MustGBP(int(f.SubtotalMinor)),
		FundedDeduction:               domain.MustGBP(int(f.FundedDeductionMinor)),
		TotalDue:                      domain.MustGBP(int(f.TotalDueMinor)),
		AmountPaid:                    domain.MustGBP(int(f.AmountPaidMinor)),
		DueAt:                         pgtypeTimestamptzToTimePtr(f.DueAt),
		IssuedAt:                      pgtypeTimestamptzToTimePtr(f.IssuedAt),
		LockedAt:                      pgtypeTimestamptzToTimePtr(f.LockedAt),
		PaidAt:                        pgtypeTimestamptzToTimePtr(f.PaidAt),
		PaymentFailedAt:               pgtypeTimestamptzToTimePtr(f.PaymentFailedAt),
		PaymentStatusUpdatedAt:        pgtypeTimestamptzToTimePtr(f.PaymentStatusUpdatedAt),
		AdjustsInvoiceID:              pgtypeUUIDToUUIDPtr(f.AdjustsInvoiceID),
		AdjustmentReasonCode:          pgtypeTextToStrPtr(f.AdjustmentReasonCode),
		AdjustmentReasonNote:          pgtypeTextToStrPtr(f.AdjustmentReasonNote),
		GeneratedRunID:                pgtypeUUIDToUUIDPtr(f.GeneratedRunID),
		GeneratedRunStatus:            pgtypeTextToStrPtr(f.GeneratedRunStatus),
		GeneratedRunStartedAt:         pgtypeTimestamptzToTimePtr(f.GeneratedRunStartedAt),
		GeneratedRunCompletedAt:       pgtypeTimestamptzToTimePtr(f.GeneratedRunCompletedAt),
		GeneratedRunDetails:           json.RawMessage(f.GeneratedRunDetails),
		CalculationDetails:            json.RawMessage(f.CalculationDetails),
		ChildPhotoPath:                pgtypeTextToStrPtr(f.ChildProfilePhotoPath),
		CreatedAt:                     pgtypeTimestamptzToTime(f.CreatedAt),
		UpdatedAt:                     pgtypeTimestamptzToTime(f.UpdatedAt),
		LatestPaymentAttemptStatus:    f.LatestPaymentAttemptStatus,
		LatestPaymentAttemptCreatedAt: pgtypeTimestamptzToTimePtr(f.LatestPaymentAttemptCreatedAt),
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
		Subtotal:                domain.MustGBP(int(row.SubtotalMinor)),
		FundedDeduction:         domain.MustGBP(int(row.FundedDeductionMinor)),
		TotalDue:                domain.MustGBP(int(row.TotalDueMinor)),
		AmountPaid:              domain.MustGBP(int(row.AmountPaidMinor)),
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
		RoomName:                pgtypeTextToStrPtr(row.RoomName),
		ChildPhotoPath:          pgtypeTextToStrPtr(row.ChildProfilePhotoPath),
		CreatedAt:               pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:               pgtypeTimestamptzToTime(row.UpdatedAt),
	}
}

// --- Parent Invoice View (API-21) read-only methods ---

func (r *Repository) ListInvoicesForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID, filters domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceListForParent(ctx, sqlc.InvoiceListForParentParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		ID:               uuidToPgtype(membershipID),
		BillingMonth:     timeToPgtypeDatePtr(filters.BillingMonth),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
		Statuses:         strSliceToPgtypeTextArray(filters.Statuses),
		ChildID:          uuidToPgtypePtr(filters.ChildID),
		Limit:            pgtype.Int4{Int32: int32(filters.Limit), Valid: true},
		Offset:           pgtype.Int4{Int32: int32(filters.Offset), Valid: true},
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

func (r *Repository) CountInvoicesForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID, filters domain.ParentInvoiceFilters) (int, error) {
	q := sqlc.New(r.pool)
	count, err := q.InvoiceCountForParent(ctx, sqlc.InvoiceCountForParentParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		ID:               uuidToPgtype(membershipID),
		BillingMonth:     timeToPgtypeDatePtr(filters.BillingMonth),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
		Statuses:         strSliceToPgtypeTextArray(filters.Statuses),
		ChildID:          uuidToPgtypePtr(filters.ChildID),
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
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
		line := domain.ParentInvoiceLineRow{
			LineKind:        row.LineKind,
			Description:     row.Description,
			SortOrder:       int(row.SortOrder),
			QuantityMinutes: pgtypeInt4ToIntPtr(row.QuantityMinutes),
			UnitAmount:      pgtypeInt4ToMoneyPtr(row.UnitAmountMinor),
			LineAmount:      domain.MustGBP(int(row.LineAmountMinor)),
		}
		if len(row.Details) > 0 {
			var details struct {
				FundingModel string `json:"funding_model"`
			}
			if json.Unmarshal(row.Details, &details) == nil && details.FundingModel != "" {
				line.FundingModel = &details.FundingModel
			}
		}
		result = append(result, line)
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
		Subtotal:               domain.MustGBP(int(subtotalMinor)),
		FundedDeduction:        domain.MustGBP(int(fundedDeductionMinor)),
		TotalDue:               domain.MustGBP(int(totalDueMinor)),
		AmountPaid:             domain.MustGBP(int(amountPaidMinor)),
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

// --- Invoice Void ---

func (r *Repository) MarkInvoiceVoid(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID uuid.UUID, reason string, voidedAt time.Time) (int64, error) {
	n, err := r.queriesTx(tx).MarkInvoiceVoid(ctx, sqlc.MarkInvoiceVoidParams{
		ID:         uuidToPgtype(invoiceID),
		TenantID:   uuidToPgtype(tenantID),
		BranchID:   uuidToPgtype(branchID),
		VoidedAt:   pgtype.Timestamptz{Time: voidedAt, Valid: true},
		VoidReason: pgtype.Text{String: reason, Valid: true},
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

// --- Pre-Overdue Reminders transactional methods ---

func (r *Repository) TryAcquireReminderJobLock(ctx context.Context, tx domain.Tx) (bool, error) {
	return r.queriesTx(tx).TryAcquireReminderJobLock(ctx)
}

func (r *Repository) ListInvoicesDueSoon(ctx context.Context, tx domain.Tx) ([]domain.InvoiceReminderRow, error) {
	rows, err := r.queriesTx(tx).ListInvoicesDueSoon(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.InvoiceReminderRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceReminderRow{
			ID:       pgtypeUUIDToUUID(row.ID),
			TenantID: pgtypeUUIDToUUID(row.TenantID),
			BranchID: pgtypeUUIDToUUID(row.BranchID),
			DueDate:  pgtypeTimestamptzToTime(row.DueAt),
		})
	}
	return result, nil
}

func (r *Repository) ListInvoicesDueToday(ctx context.Context, tx domain.Tx) ([]domain.InvoiceReminderRow, error) {
	rows, err := r.queriesTx(tx).ListInvoicesDueToday(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]domain.InvoiceReminderRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceReminderRow{
			ID:       pgtypeUUIDToUUID(row.ID),
			TenantID: pgtypeUUIDToUUID(row.TenantID),
			BranchID: pgtypeUUIDToUUID(row.BranchID),
			DueDate:  pgtypeTimestamptzToTime(row.DueAt),
		})
	}
	return result, nil
}

func (r *Repository) InsertInvoiceReminderLog(ctx context.Context, tx domain.Tx, tenantID, branchID, invoiceID uuid.UUID, reminderType string) error {
	return r.queriesTx(tx).InsertInvoiceReminderLog(ctx, sqlc.InsertInvoiceReminderLogParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		InvoiceID:    uuidToPgtype(invoiceID),
		ReminderType: reminderType,
	})
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
		TotalDue:        domain.MustGBP(int(row.TotalDueMinor)),
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
			TotalDue:        domain.MustGBP(int(row.TotalDueMinor)),
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
			TotalDue:        domain.MustGBP(int(row.TotalDueMinor)),
		})
	}
	return result
}

// --- Helpers ---

func mapCandidateRows(rows []sqlc.ListCandidateChildrenForUpdateRow) []PreflightChildRow {
	result := make([]PreflightChildRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, PreflightChildRow{
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
			HasParentCarerContact:  row.HasParentCarerContact,
			FundingProfileID:       pgtypeUUIDToUUIDPtr(row.FundingProfileID),
			FundedAllowanceMinutes: pgtypeInt4ToIntPtr(row.FundedAllowanceMinutes),
			ExistingInvoiceID:      pgtypeUUIDToUUIDPtr(row.ExistingInvoiceID),
			ExistingInvoiceStatus:  pgtypeTextToStrPtr(row.ExistingInvoiceStatus),
		})
	}
	return result
}

func mapSelectedRows(rows []sqlc.ListSelectedChildrenForUpdateRow) []PreflightChildRow {
	result := make([]PreflightChildRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, PreflightChildRow{
			ChildID:                pgtypeUUIDToUUID(row.ChildID),
			FirstName:              row.FirstName,
			MiddleName:             pgtypeTextToStrPtr(row.MiddleName),
			LastName:               pgtypeTextToStrPtr(row.LastName),
			DateOfBirth:            pgtypeDateToTime(row.DateOfBirth),
			StartDate:              pgtypeDateToTime(row.StartDate),
			EndDate:                pgtypeDateToTimePtr(row.EndDate),
			CoreHourlyRateMinor:    pgtypeInt4ToIntPtr(row.CoreHourlyRateMinor),
			HasParentCarerContact:  row.HasParentCarerContact,
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

func pgtypeInt4ToInt(i pgtype.Int4) int {
	if !i.Valid {
		return 0
	}
	return int(i.Int32)
}

func pgtypeInt4ToMoney(i pgtype.Int4) domain.Money {
	return domain.MustGBP(pgtypeInt4ToInt(i))
}

func pgtypeInt4ToIntPtr(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
}

func pgtypeInt4ToMoneyPtr(i pgtype.Int4) *domain.Money {
	if !i.Valid {
		return nil
	}
	m := domain.MustGBP(int(i.Int32))
	return &m
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

func strSliceToPgtypeTextArray(ss []string) []string {
	if len(ss) == 0 {
		return nil
	}
	return ss
}

func pgtypeNumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f.Float64
}

func pgtypeNumericToFloat64Ptr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	v := pgtypeNumericToFloat64(n)
	return &v
}

func uuidToPgtypePtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(*u), Valid: true}
}

// --- Invoice Export ---

func (r *Repository) ExportInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceExportForManagerReview(ctx, sqlc.InvoiceExportForManagerReviewParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		BillingMonth:     timeToPgtypeDatePtr(filters.BillingMonth),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
		Statuses:         strSliceToPgtypeTextArray(filters.Statuses),
		ChildID:          uuidToPgtypePtr(filters.ChildID),
		Search:           strToPgtypeTextPtr(filters.Search),
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.InvoiceReviewRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceReviewRow{
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
			Subtotal:                domain.MustGBP(int(row.SubtotalMinor)),
			FundedDeduction:         domain.MustGBP(int(row.FundedDeductionMinor)),
			TotalDue:                domain.MustGBP(int(row.TotalDueMinor)),
			AmountPaid:              domain.MustGBP(int(row.AmountPaidMinor)),
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
			ChildPhotoPath:          pgtypeTextToStrPtr(row.ChildProfilePhotoPath),
			CreatedAt:               pgtypeTimestamptzToTime(row.CreatedAt),
			UpdatedAt:               pgtypeTimestamptzToTime(row.UpdatedAt),
		})
	}
	return result, nil
}

func (r *Repository) ExportInvoiceDetailsForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceExportDetailForManagerReview(ctx, sqlc.InvoiceExportDetailForManagerReviewParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		BillingMonth:     timeToPgtypeDatePtr(filters.BillingMonth),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
		Statuses:         strSliceToPgtypeTextArray(filters.Statuses),
		ChildID:          uuidToPgtypePtr(filters.ChildID),
		Search:           strToPgtypeTextPtr(filters.Search),
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.InvoiceExportLineRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceExportLineRow{
			InvoiceNumber:   pgtypeTextToStrPtr(row.InvoiceNumber),
			ChildFirstName:  row.ChildFirstName,
			ChildLastName:   pgtypeTextToStrPtr(row.ChildLastName),
			BillingMonth:    pgtypeDateToTime(row.BillingMonth),
			Status:          row.Status,
			LineKind:        row.LineKind,
			Description:     row.Description,
			QuantityMinutes: pgtypeInt4ToIntPtr(row.QuantityMinutes),
			UnitAmountMinor: pgtypeInt4ToIntPtr(row.UnitAmountMinor),
			LineAmountMinor: int(row.LineAmountMinor),
		})
	}
	return result, nil
}

func (r *Repository) InvoiceSummaryByMonth(ctx context.Context, tenantID, branchID uuid.UUID, filters domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceSummaryByMonth(ctx, sqlc.InvoiceSummaryByMonthParams{
		TenantID:         uuidToPgtype(tenantID),
		BranchID:         uuidToPgtype(branchID),
		BillingMonthFrom: timeToPgtypeDatePtr(filters.BillingMonthFrom),
		BillingMonthTo:   timeToPgtypeDatePtr(filters.BillingMonthTo),
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.InvoiceMonthSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InvoiceMonthSummary{
			BillingMonth:          pgtypeDateToTime(row.BillingMonth),
			TotalInvoicedMinor:    int(row.TotalInvoicedMinor),
			TotalCollectedMinor:   int(row.TotalCollectedMinor),
			TotalOutstandingMinor: int(row.TotalOutstandingMinor),
			TotalOverdueMinor:     int(row.TotalOverdueMinor),
			InvoiceCount:          int(row.InvoiceCount),
		})
	}
	return result, nil
}

func (r *Repository) InvoiceOverdueSummary(ctx context.Context, tenantID, branchID uuid.UUID) (domain.OverdueSummary, error) {
	q := sqlc.New(r.pool)
	row, err := q.InvoiceOverdueSummary(ctx, sqlc.InvoiceOverdueSummaryParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return domain.OverdueSummary{}, err
	}
	return domain.OverdueSummary{
		TotalOverdueMinor: int(row.TotalOverdueMinor),
		OverdueCount:      int(row.OverdueCount),
	}, nil
}

func (r *Repository) InvoiceOverdueTopItems(ctx context.Context, tenantID, branchID uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	q := sqlc.New(r.pool)
	rows, err := q.InvoiceOverdueTopItems(ctx, sqlc.InvoiceOverdueTopItemsParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.OverdueSummaryItem, 0, len(rows))
	for _, row := range rows {
		invoiceNumber := ""
		if row.InvoiceNumber.Valid {
			invoiceNumber = row.InvoiceNumber.String
		}
		result = append(result, domain.OverdueSummaryItem{
			ID:               pgtypeUUIDToUUID(row.ID),
			InvoiceNumber:    invoiceNumber,
			ChildID:          pgtypeUUIDToUUID(row.ChildID),
			ChildName:        row.ChildName,
			OutstandingMinor: int(row.OutstandingMinor),
			DueDate:          pgtypeTimestamptzToTime(row.DueAt),
			DaysOverdue:      int(row.DaysOverdue),
		})
	}
	return result, nil
}

// --- BranchSettingsRepository implementation ---

func (r *Repository) GetOverdueGraceDays(ctx context.Context, tenantID, branchID uuid.UUID) (int, error) {
	q := sqlc.New(r.pool)
	row, err := q.GetOverdueGraceDays(ctx, sqlc.GetOverdueGraceDaysParams{
		TenantID: uuidToPgtype(tenantID),
		ID:       uuidToPgtype(branchID),
	})
	if err != nil {
		return 0, err
	}
	return int(row), nil
}

func (r *Repository) UpdateOverdueGraceDays(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, days int) error {
	return r.queriesTx(tx).UpdateOverdueGraceDays(ctx, sqlc.UpdateOverdueGraceDaysParams{
		TenantID:         uuidToPgtype(tenantID),
		ID:               uuidToPgtype(branchID),
		OverdueGraceDays: int32(days),
	})
}

func (r *Repository) GetReminderDaysBefore(ctx context.Context, tenantID, branchID uuid.UUID) (int, error) {
	q := sqlc.New(r.pool)
	row, err := q.GetReminderDaysBefore(ctx, sqlc.GetReminderDaysBeforeParams{
		TenantID: uuidToPgtype(tenantID),
		ID:       uuidToPgtype(branchID),
	})
	if err != nil {
		return 0, err
	}
	return int(row), nil
}

func (r *Repository) UpdateReminderDaysBefore(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, days int) error {
	return r.queriesTx(tx).UpdateReminderDaysBefore(ctx, sqlc.UpdateReminderDaysBeforeParams{
		TenantID:           uuidToPgtype(tenantID),
		ID:                 uuidToPgtype(branchID),
		ReminderDaysBefore: int32(days),
	})
}
