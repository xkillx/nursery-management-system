package domain

import (
	"testing"
	"time"
)

func TestComputeInvoicePrefill(t *testing.T) {
	july2026 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	// 3 sessions/week: Mon, Wed, Fri, each 2 hours (120 min)
	entries := []BookedPatternEntry{
		{
			DayOfWeek: 1, // Monday
			SessionType: BookedSessionType{
				ID:              "st1",
				Name:            "Full Day",
				StartMinutes:    480, // 08:00
				EndMinutes:      720, // 12:00
				DurationMinutes: 240,
			},
		},
		{
			DayOfWeek: 3, // Wednesday
			SessionType: BookedSessionType{
				ID:              "st1",
				Name:            "Full Day",
				StartMinutes:    480,
				EndMinutes:      720,
				DurationMinutes: 240,
			},
		},
		{
			DayOfWeek: 5, // Friday
			SessionType: BookedSessionType{
				ID:              "st1",
				Name:            "Full Day",
				StartMinutes:    480,
				EndMinutes:      720,
				DurationMinutes: 240,
			},
		},
	}

	t.Run("no funding profile", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:                entries,
			BillingMonthStart:      july2026,
			SiteHourlyRateMinor:    500,
			FundedAllowanceMinutes: 0,
			HasFunding:             false,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.SubtotalMinor <= 0 {
			t.Errorf("SubtotalMinor = %d, want > 0", result.SubtotalMinor)
		}
		if result.FundedDeductionMinor != 0 {
			t.Errorf("FundedDeductionMinor = %d, want 0", result.FundedDeductionMinor)
		}
		if result.TotalDueMinor != result.SubtotalMinor {
			t.Errorf("TotalDueMinor = %d, want %d", result.TotalDueMinor, result.SubtotalMinor)
		}
		if len(result.Lines) != 1 {
			t.Errorf("lines count = %d, want 1", len(result.Lines))
		}
		if result.Lines[0].LineKind != LineKindCoreChildcare {
			t.Errorf("line kind = %q, want %q", result.Lines[0].LineKind, LineKindCoreChildcare)
		}

		hasMissingFunding := false
		for _, w := range result.Warnings {
			if w == "missing_funding_record" {
				hasMissingFunding = true
			}
		}
		if !hasMissingFunding {
			t.Error("expected missing_funding_record warning")
		}
	})

	t.Run("with funded minutes deduction", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:                entries,
			BillingMonthStart:      july2026,
			SiteHourlyRateMinor:    500,
			FundedHourlyRateMinor:  500,
			FundedAllowanceMinutes: 600, // 10 hours funded
			HasFunding:             true,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.FundedDeductionMinor <= 0 {
			t.Errorf("FundedDeductionMinor = %d, want > 0", result.FundedDeductionMinor)
		}
		if result.TotalDueMinor >= result.SubtotalMinor {
			t.Errorf("TotalDueMinor = %d, want < SubtotalMinor %d", result.TotalDueMinor, result.SubtotalMinor)
		}
		if len(result.Lines) != 2 {
			t.Errorf("lines count = %d, want 2", len(result.Lines))
		}
		if result.Lines[1].LineKind != LineKindFundedDeduction {
			t.Errorf("second line kind = %q, want %q", result.Lines[1].LineKind, LineKindFundedDeduction)
		}

		fundedLine := result.Lines[1]
		if fundedLine.QuantityMinutes != 600 {
			t.Errorf("funded QuantityMinutes = %d, want 600", fundedLine.QuantityMinutes)
		}
		if fundedLine.UnitAmountMinor != -500 {
			t.Errorf("funded UnitAmountMinor = %d, want -500", fundedLine.UnitAmountMinor)
		}
		if fundedLine.LineAmountMinor != result.FundedDeductionMinor {
			t.Errorf("funded LineAmountMinor = %d, want %d", fundedLine.LineAmountMinor, result.FundedDeductionMinor)
		}

		if result.SubtotalMinor != result.Lines[0].LineAmountMinor {
			t.Errorf("SubtotalMinor = %d, want %d (positive lines only)", result.SubtotalMinor, result.Lines[0].LineAmountMinor)
		}
		if result.TotalDueMinor != result.SubtotalMinor-result.FundedDeductionMinor {
			t.Errorf("TotalDueMinor = %d, want %d", result.TotalDueMinor, result.SubtotalMinor-result.FundedDeductionMinor)
		}
	})

	t.Run("funded minutes exceed booked minutes", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:                entries,
			BillingMonthStart:      july2026,
			SiteHourlyRateMinor:    500,
			FundedHourlyRateMinor:  500,
			FundedAllowanceMinutes: 99999, // way more than booked
			HasFunding:             true,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TotalDueMinor != 0 {
			t.Errorf("TotalDueMinor = %d, want 0", result.TotalDueMinor)
		}
	})

	t.Run("warning threshold uses pre-deduction subtotal", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:                entries,
			BillingMonthStart:      july2026,
			SiteHourlyRateMinor:    500,
			FundedHourlyRateMinor:  500,
			FundedAllowanceMinutes: 600, // 10 hours funded
			HasFunding:             true,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		preDeductionSubtotal := result.SubtotalMinor + result.FundedDeductionMinor
		threshold := preDeductionSubtotal / 4
		if result.FundedDeductionMinor > threshold {
			hasWarning := false
			for _, w := range result.Warnings {
				if w == "significant_funding_deduction" {
					hasWarning = true
				}
			}
			if !hasWarning {
				t.Error("expected significant_funding_deduction warning when deduction > 25% of pre-deduction subtotal")
			}
		}
	})

	t.Run("zero funded minutes", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:                entries,
			BillingMonthStart:      july2026,
			SiteHourlyRateMinor:    500,
			FundedAllowanceMinutes: 0,
			HasFunding:             true,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.FundedDeductionMinor != 0 {
			t.Errorf("FundedDeductionMinor = %d, want 0", result.FundedDeductionMinor)
		}
		if result.TotalDueMinor != result.SubtotalMinor {
			t.Errorf("TotalDueMinor = %d, want %d", result.TotalDueMinor, result.SubtotalMinor)
		}
	})

	t.Run("negative hourly rate returns error", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: -1,
		}

		_, err := ComputeInvoicePrefill(params)
		if err == nil {
			t.Fatal("expected error for negative rate")
		}
	})

	t.Run("transparency fields populated with term dates", func(t *testing.T) {
		termDates := []TermDateRange{
			{StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)},
		}
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: 500,
			TermDates:           termDates,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.TermDatesUsed) != 1 {
			t.Fatalf("expected 1 TermDatesUsed, got %d", len(result.TermDatesUsed))
		}
		if result.TermDatesUsed[0] != "2026-07-01 to 2026-07-31" {
			t.Errorf("unexpected TermDatesUsed: %s", result.TermDatesUsed[0])
		}
	})

	t.Run("transparency fields populated with closure dates", func(t *testing.T) {
		closureDates := []time.Time{
			time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC),
		}
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: 500,
			ClosureDates:        closureDates,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.ClosureDaysExcluded) != 1 {
			t.Fatalf("expected 1 ClosureDaysExcluded, got %d", len(result.ClosureDaysExcluded))
		}
		if result.ClosureDaysExcluded[0] != "2026-07-06" {
			t.Errorf("unexpected ClosureDaysExcluded: %s", result.ClosureDaysExcluded[0])
		}
	})

	t.Run("transparency fields populated with holiday periods", func(t *testing.T) {
		holidayPeriods := []HolidayPeriodDateRange{
			{StartDate: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)},
		}
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: 500,
			HolidayPeriods:      holidayPeriods,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.HolidayPeriodsExcluded) != 1 {
			t.Fatalf("expected 1 HolidayPeriodsExcluded, got %d", len(result.HolidayPeriodsExcluded))
		}
		if result.HolidayPeriodsExcluded[0] != "2026-07-20 to 2026-07-24" {
			t.Errorf("unexpected HolidayPeriodsExcluded: %s", result.HolidayPeriodsExcluded[0])
		}
	})

	t.Run("no_term_dates_for_month warning when TermTimeOnly and no term dates", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: 500,
			TermTimeOnly:        true,
			TermDates:           nil,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasWarning := false
		for _, w := range result.Warnings {
			if w == "no_term_dates_for_month" {
				hasWarning = true
			}
		}
		if !hasWarning {
			t.Error("expected no_term_dates_for_month warning")
		}
	})

	t.Run("no no_term_dates_for_month warning when not TermTimeOnly", func(t *testing.T) {
		params := InvoicePrefillParams{
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: 500,
			TermTimeOnly:        false,
			TermDates:           nil,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hasWarning := false
		for _, w := range result.Warnings {
			if w == "no_term_dates_for_month" {
				hasWarning = true
			}
		}
		if hasWarning {
			t.Error("did not expect no_term_dates_for_month warning when TermTimeOnly is false")
		}
	})
}

func TestCoreChildcareDefaultDescription(t *testing.T) {
	cases := []struct {
		month time.Time
		want  string
	}{
		{time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "May 2026 Recurring Booking"},
		{time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), "December 2026 Recurring Booking"},
		{time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC), "January 2027 Recurring Booking"},
	}
	for _, tc := range cases {
		if got := CoreChildcareDefaultDescription(tc.month); got != tc.want {
			t.Errorf("CoreChildcareDefaultDescription(%v) = %q, want %q", tc.month, got, tc.want)
		}
	}
}

func TestComputeInvoicePrefill_DefaultDescriptionIsBookingTypeLabel(t *testing.T) {
	entries := []BookedPatternEntry{
		{
			DayOfWeek: 1,
			SessionType: BookedSessionType{
				ID:              "st1",
				Name:            "Full Day",
				StartMinutes:    480,
				EndMinutes:      720,
				DurationMinutes: 240,
			},
		},
	}
	params := InvoicePrefillParams{
		Entries:                entries,
		BillingMonthStart:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		SiteHourlyRateMinor:    500,
		FundedAllowanceMinutes: 0,
		HasFunding:             false,
	}
	result, err := ComputeInvoicePrefill(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Lines) == 0 {
		t.Fatal("expected at least one line")
	}
	core := result.Lines[0]
	if core.Description != "May 2026 Recurring Booking" {
		t.Errorf("core line description = %q, want %q", core.Description, "May 2026 Recurring Booking")
	}
}
