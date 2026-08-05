package domain

import (
	"fmt"
	"time"
)

// InvoicePrefillParams holds the inputs for the billing prefill calculation.
type InvoicePrefillParams struct {
	Entries                []BookedPatternEntry
	BillingMonthStart      time.Time
	SiteHourlyRateMinor    int
	FundedHourlyRateMinor  int
	FundedAllowanceMinutes int
	HasFunding             bool
	TermDates              []TermDateRange
	ClosureDates           []time.Time
	HolidayPeriods         []HolidayPeriodDateRange
	TermTimeOnly           bool
}

// InvoicePrefillLine is a computed line item from the prefill calculation.
type InvoicePrefillLine struct {
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

// InvoicePrefillResult holds the output of the billing prefill calculation.
type InvoicePrefillResult struct {
	Lines                  []InvoicePrefillLine
	SubtotalMinor          int
	FundedDeductionMinor   int
	TotalDueMinor          int
	TotalMinutes           int
	Warnings               []string
	TermDatesUsed          []string
	ClosureDaysExcluded    []string
	HolidayPeriodsExcluded []string
	Sessions               []BookedSession
}

// CoreChildcareDefaultDescription returns the default description for the
// recurring core-childcare line, scoped to the invoice's billing month (e.g.
// "May 2026 Recurring Booking"). Shared by generation, prefill, and the
// manual-draft create paths so they always agree (KTD5).
func CoreChildcareDefaultDescription(billingMonth time.Time) string {
	return fmt.Sprintf("%s Recurring Booking", billingMonth.Format("January 2006"))
}

// ComputeInvoicePrefill is a pure domain service that computes invoice line
// items and totals from booking pattern data, hourly rate, and funding info.
// It has no side effects and no infrastructure dependencies.
func ComputeInvoicePrefill(params InvoicePrefillParams) (InvoicePrefillResult, error) {
	if params.SiteHourlyRateMinor < 0 {
		return InvoicePrefillResult{}, fmt.Errorf("site hourly rate must not be negative")
	}

	calc, err := CalculateBookedCoreMinutesInMonth(
		"",
		params.Entries,
		params.BillingMonthStart,
		params.SiteHourlyRateMinor,
		params.TermDates,
		params.ClosureDates,
		params.HolidayPeriods,
	)
	if err != nil {
		return InvoicePrefillResult{}, fmt.Errorf("calculate booked minutes: %w", err)
	}

	subtotalMinor := calc.Subtotal.Minor()
	fundedDeductionMinor := 0
	fundedDeductionMinutes := 0
	billableMinutes := calc.TotalMinutes

	if params.HasFunding {
		var fundErr error
		fundedDeductionMinutes, billableMinutes, fundedDeductionMinor, _, fundErr = ComputeFundedDeductionMinor(
			calc.TotalMinutes, params.FundedAllowanceMinutes, params.FundedHourlyRateMinor,
		)
		if fundErr != nil {
			return InvoicePrefillResult{}, fmt.Errorf("compute funded deduction: %w", fundErr)
		}
	}

	totalDueMinor := subtotalMinor - fundedDeductionMinor
	if totalDueMinor < 0 {
		totalDueMinor = 0
	}

	lines := make([]InvoicePrefillLine, 0, 2)
	lines = append(lines, InvoicePrefillLine{
		LineKind:               LineKindCoreChildcare,
		Description:            CoreChildcareDefaultDescription(params.BillingMonthStart),
		SortOrder:              1,
		QuantityMinutes:        calc.TotalMinutes,
		UnitAmountMinor:        params.SiteHourlyRateMinor,
		LineAmountMinor:        subtotalMinor,
		FundedAllowanceMinutes: params.FundedAllowanceMinutes,
		FundedDeductionMinutes: fundedDeductionMinutes,
		CoreBillableMinutes:    billableMinutes,
		SessionCount:           len(calc.Sessions),
	})

	if params.HasFunding && fundedDeductionMinor > 0 {
		lines = append(lines, InvoicePrefillLine{
			LineKind:               LineKindFundedDeduction,
			Description:            "Funded hours deduction",
			SortOrder:              2,
			QuantityMinutes:        fundedDeductionMinutes,
			UnitAmountMinor:        -params.SiteHourlyRateMinor,
			FundedAllowanceMinutes: params.FundedAllowanceMinutes,
			FundedDeductionMinutes: fundedDeductionMinutes,
			CoreBillableMinutes:    billableMinutes,
			LineAmountMinor:        fundedDeductionMinor,
		})
	}

	var warnings []string
	if params.SiteHourlyRateMinor <= 0 {
		warnings = append(warnings, "site_rate_not_set")
	}
	if !params.HasFunding {
		warnings = append(warnings, "missing_funding_record")
	}
	if fundedDeductionMinor > 0 && subtotalMinor > 0 {
		threshold := (subtotalMinor + fundedDeductionMinor) / 4
		if fundedDeductionMinor > threshold {
			warnings = append(warnings, "significant_funding_deduction")
		}
	}
	if params.TermTimeOnly && len(params.TermDates) == 0 {
		warnings = append(warnings, "no_term_dates_for_month")
	}

	return InvoicePrefillResult{
		Lines:                  lines,
		SubtotalMinor:          subtotalMinor,
		FundedDeductionMinor:   fundedDeductionMinor,
		TotalDueMinor:          totalDueMinor,
		TotalMinutes:           calc.TotalMinutes,
		Warnings:               warnings,
		TermDatesUsed:          calc.TermDatesUsed,
		ClosureDaysExcluded:    calc.ClosureDaysExcluded,
		HolidayPeriodsExcluded: calc.HolidayPeriodsExcluded,
		Sessions:               calc.Sessions,
	}, nil
}
