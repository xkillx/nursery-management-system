package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type PrefillTxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type ComputeInvoicePrefill struct {
	repo                 domain.BillingRepository
	txMgr                PrefillTxManager
	bookingEntriesLookup domain.BookingEntriesLookup
	fundingLookup        domain.FundingLookup
	termDateLookup       domain.TermDateLookup
	closureDateLookup    domain.ClosureDateLookup
	holidayPeriodLookup  domain.HolidayPeriodLookup
}

func NewComputeInvoicePrefill(repo domain.BillingRepository, txMgr PrefillTxManager, bookingEntriesLookup domain.BookingEntriesLookup, fundingLookup domain.FundingLookup, termDateLookup domain.TermDateLookup, closureDateLookup domain.ClosureDateLookup, holidayPeriodLookup domain.HolidayPeriodLookup) *ComputeInvoicePrefill {
	return &ComputeInvoicePrefill{repo: repo, txMgr: txMgr, bookingEntriesLookup: bookingEntriesLookup, fundingLookup: fundingLookup, termDateLookup: termDateLookup, closureDateLookup: closureDateLookup, holidayPeriodLookup: holidayPeriodLookup}
}

type ComputeInvoicePrefillResult struct {
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	BillingMonth           string
	FundedAllowanceMinutes int
	Lines                  []PrefillLine
	SubtotalMinor          int
	FundedDeductionMinor   int
	TotalDueMinor          int
	Warnings               []string
	TermDatesUsed          []string
	ClosureDaysExcluded    []string
	HolidayPeriodsExcluded []string
}

type PrefillLine struct {
	LineKind               string
	Description            string
	SortOrder              int
	QuantityMinutes        int
	UnitAmountMinor        int
	LineAmountMinor        int
	FundedAllowanceMinutes int
	FundedDeductionMinutes int
	CoreBillableMinutes    int
	SessionCount           int
}

func (uc *ComputeInvoicePrefill) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw, billingMonthRaw string) (ComputeInvoicePrefillResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return ComputeInvoicePrefillResult{}, domainerrors.Validation("Invalid child ID format.", "child_id")
	}

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return ComputeInvoicePrefillResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	var result ComputeInvoicePrefillResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		bookings, listErr := uc.repo.ListActiveBookingsForGeneration(ctx, tx, actor.TenantID, actor.BranchID, billingMonth)
		if listErr != nil {
			return fmt.Errorf("list active bookings: %w", listErr)
		}

		var bookingRow *domain.BillableChildRow
		for i, b := range bookings {
			if b.ChildID == childID {
				bookingRow = &bookings[i]
				break
			}
		}
		if bookingRow == nil {
			return domainerrors.NotFound("child", "Child not found for this billing month.")
		}

		warnings := prefillWarnings(*bookingRow)

		entries, entriesErr := uc.bookingEntriesLookup.GetEntriesForChildInMonth(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
		if entriesErr != nil {
			return fmt.Errorf("lookup booking entries for child: %w", entriesErr)
		}

		domainEntries := entries

		fundedAllowance := 0
		hasFunding := false
		fundedHourlyRateMinor := 0

		if uc.fundingLookup != nil {
			fundingInfo, fundErr := uc.fundingLookup.GetChildFunding(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
			if fundErr != nil {
				return fmt.Errorf("lookup funding for child: %w", fundErr)
			}
			if fundingInfo.HasFunding {
				hasFunding = true
				fundedAllowance = fundingInfo.FundedAllowanceMinutes
				fundedHourlyRateMinor = fundingInfo.FundedHourlyRateMinor
			}
		}

		var termDates []domain.TermDateRange
		if bookingRow.TermTimeOnly && uc.termDateLookup != nil {
			var termErr error
			termDates, termErr = uc.termDateLookup.GetTermDateRangesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
			if termErr != nil {
				return fmt.Errorf("lookup term dates: %w", termErr)
			}
		}

		var closureDates []time.Time
		if uc.closureDateLookup != nil {
			var closureErr error
			closureDates, closureErr = uc.closureDateLookup.GetClosureDatesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
			if closureErr != nil {
				return fmt.Errorf("lookup closure dates: %w", closureErr)
			}
		}

		var holidayPeriods []domain.HolidayPeriodDateRange
		if bookingRow.TermTimeOnly && uc.holidayPeriodLookup != nil {
			var holidayErr error
			holidayPeriods, holidayErr = uc.holidayPeriodLookup.GetHolidayPeriodsForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
			if holidayErr != nil {
				return fmt.Errorf("lookup holiday periods: %w", holidayErr)
			}
		}

		prefillResult, prefillErr := domain.ComputeInvoicePrefill(domain.InvoicePrefillParams{
			Entries:                domainEntries,
			BillingMonthStart:      billingMonth,
			SiteHourlyRateMinor:    bookingRow.SiteHourlyRateMinor,
			FundedHourlyRateMinor:  fundedHourlyRateMinor,
			FundedAllowanceMinutes: fundedAllowance,
			HasFunding:             hasFunding,
			TermDates:              termDates,
			ClosureDates:           closureDates,
			HolidayPeriods:         holidayPeriods,
			TermTimeOnly:           bookingRow.TermTimeOnly,
		})
		if prefillErr != nil {
			return fmt.Errorf("compute invoice prefill: %w", prefillErr)
		}

		lines := make([]PrefillLine, 0, len(prefillResult.Lines))
		for _, l := range prefillResult.Lines {
			lines = append(lines, PrefillLine{
				LineKind:               l.LineKind,
				Description:            l.Description,
				SortOrder:              l.SortOrder,
				QuantityMinutes:        l.QuantityMinutes,
				UnitAmountMinor:        l.UnitAmountMinor,
				LineAmountMinor:        l.LineAmountMinor,
				FundedAllowanceMinutes: l.FundedAllowanceMinutes,
				FundedDeductionMinutes: l.FundedDeductionMinutes,
				CoreBillableMinutes:    l.CoreBillableMinutes,
				SessionCount:           l.SessionCount,
			})
		}

		warnings = append(warnings, prefillResult.Warnings...)

		result = ComputeInvoicePrefillResult{
			ChildID:                bookingRow.ChildID,
			ChildFirstName:         bookingRow.FirstName,
			ChildMiddleName:        bookingRow.MiddleName,
			ChildLastName:          bookingRow.LastName,
			BillingMonth:           billingMonthRaw,
			FundedAllowanceMinutes: fundedAllowance,
			Lines:                  lines,
			SubtotalMinor:          prefillResult.SubtotalMinor,
			FundedDeductionMinor:   prefillResult.FundedDeductionMinor,
			TotalDueMinor:          prefillResult.TotalDueMinor,
			Warnings:               warnings,
			TermDatesUsed:          prefillResult.TermDatesUsed,
			ClosureDaysExcluded:    prefillResult.ClosureDaysExcluded,
			HolidayPeriodsExcluded: prefillResult.HolidayPeriodsExcluded,
		}

		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return ComputeInvoicePrefillResult{}, txErr
		}
		return ComputeInvoicePrefillResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func prefillWarnings(t domain.BillableChildRow) []string {
	var w []string
	if t.FirstName == "" {
		w = append(w, "missing_child_name")
	}
	if t.DateOfBirth.IsZero() {
		w = append(w, "missing_date_of_birth")
	}
	if t.StartDate.IsZero() {
		w = append(w, "missing_start_date")
	}
	if !t.HasParentCarerContact {
		w = append(w, "missing_guardian_link")
	}
	return w
}
