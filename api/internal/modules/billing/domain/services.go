package domain

import (
	"fmt"
	"time"
)

// InvoicePrefillParams holds the inputs for the billing prefill calculation.
type InvoicePrefillParams struct {
	BookingPatternID    string
	Entries             []BookedPatternEntry
	BillingMonthStart   time.Time
	SiteHourlyRateMinor int
	FundedAllowance     int
	HasFundingProfile   bool
	TermDates           []TermDateRange
	ClosureDates        []time.Time
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
	Lines                []InvoicePrefillLine
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
	TotalMinutes         int
	Warnings             []string
}

// ComputeInvoicePrefill is a pure domain service that computes invoice line
// items and totals from booking pattern data, hourly rate, and funding info.
// It has no side effects and no infrastructure dependencies.
func ComputeInvoicePrefill(params InvoicePrefillParams) (InvoicePrefillResult, error) {
	if params.SiteHourlyRateMinor < 0 {
		return InvoicePrefillResult{}, fmt.Errorf("site hourly rate must not be negative")
	}

	calc, err := CalculateBookedCoreMinutesInMonth(
		params.BookingPatternID,
		params.Entries,
		params.BillingMonthStart,
		params.SiteHourlyRateMinor,
		params.TermDates,
		params.ClosureDates,
	)
	if err != nil {
		return InvoicePrefillResult{}, fmt.Errorf("calculate booked minutes: %w", err)
	}

	subtotalMinor := calc.Subtotal.Minor()
	fundedDeductionMinor := 0
	fundedDeductionMinutes := 0
	billableMinutes := calc.TotalMinutes

	if params.HasFundingProfile {
		var fundErr error
		fundedDeductionMinutes, billableMinutes, fundedDeductionMinor, _, fundErr = ComputeFundedDeductionMinor(
			calc.TotalMinutes, params.FundedAllowance, params.SiteHourlyRateMinor,
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
		Description:            "Core childcare",
		SortOrder:              1,
		QuantityMinutes:        calc.TotalMinutes,
		UnitAmountMinor:        params.SiteHourlyRateMinor,
		LineAmountMinor:        subtotalMinor,
		FundedAllowanceMinutes: params.FundedAllowance,
		FundedDeductionMinutes: fundedDeductionMinutes,
		CoreBillableMinutes:    billableMinutes,
		SessionCount:           len(calc.Sessions),
	})

	if params.HasFundingProfile && fundedDeductionMinor > 0 {
		lines = append(lines, InvoicePrefillLine{
			LineKind:               LineKindFundedDeduction,
			Description:            "Funded hours deduction",
			SortOrder:              2,
			FundedAllowanceMinutes: params.FundedAllowance,
			FundedDeductionMinutes: fundedDeductionMinutes,
			CoreBillableMinutes:    billableMinutes,
			LineAmountMinor:        fundedDeductionMinor,
		})
	}

	var warnings []string
	if params.SiteHourlyRateMinor <= 0 {
		warnings = append(warnings, "site_rate_not_set")
	}
	if !params.HasFundingProfile {
		warnings = append(warnings, "missing_funding_profile")
	}
	if fundedDeductionMinor > 0 && subtotalMinor > 0 {
		threshold := subtotalMinor / 4
		if fundedDeductionMinor > threshold {
			warnings = append(warnings, "significant_funding_deduction")
		}
	}

	return InvoicePrefillResult{
		Lines:                lines,
		SubtotalMinor:        subtotalMinor,
		FundedDeductionMinor: fundedDeductionMinor,
		TotalDueMinor:        totalDueMinor,
		TotalMinutes:         calc.TotalMinutes,
		Warnings:             warnings,
	}, nil
}
