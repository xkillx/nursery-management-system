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
			BookingPatternID:      "pattern-1",
			Entries:               entries,
			BillingMonthStart:     july2026,
			SiteHourlyRateMinor:   500,
			FundedAllowanceMinutes: 0,
			HasFunding:            false,
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
			if w == "missing_funding_profile" {
				hasMissingFunding = true
			}
		}
		if !hasMissingFunding {
			t.Error("expected missing_funding_profile warning")
		}
	})

	t.Run("with funded minutes deduction", func(t *testing.T) {
		params := InvoicePrefillParams{
			BookingPatternID:      "pattern-1",
			Entries:               entries,
			BillingMonthStart:     july2026,
			SiteHourlyRateMinor:   500,
			FundedHourlyRateMinor: 500,
			FundedAllowanceMinutes: 600, // 10 hours funded
			HasFunding:            true,
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
	})

	t.Run("funded minutes exceed booked minutes", func(t *testing.T) {
		params := InvoicePrefillParams{
			BookingPatternID:      "pattern-1",
			Entries:               entries,
			BillingMonthStart:     july2026,
			SiteHourlyRateMinor:   500,
			FundedHourlyRateMinor: 500,
			FundedAllowanceMinutes: 99999, // way more than booked
			HasFunding:            true,
		}

		result, err := ComputeInvoicePrefill(params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TotalDueMinor != 0 {
			t.Errorf("TotalDueMinor = %d, want 0", result.TotalDueMinor)
		}
	})

	t.Run("zero funded minutes", func(t *testing.T) {
		params := InvoicePrefillParams{
			BookingPatternID:      "pattern-1",
			Entries:               entries,
			BillingMonthStart:     july2026,
			SiteHourlyRateMinor:   500,
			FundedAllowanceMinutes: 0,
			HasFunding:            true,
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
			BookingPatternID:    "pattern-1",
			Entries:             entries,
			BillingMonthStart:   july2026,
			SiteHourlyRateMinor: -1,
		}

		_, err := ComputeInvoicePrefill(params)
		if err == nil {
			t.Fatal("expected error for negative rate")
		}
	})
}
