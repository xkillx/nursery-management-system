package domain

import (
	"testing"
	"time"
)

func TestComputeAllowanceMinutes_TermTime(t *testing.T) {
	billingMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	termDates := []TermDateRange{
		{StartDate: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC)},
	}

	// 15h/week, 20 weekdays in term (Jan 5-30 has 20 weekdays)
	minutes, err := ComputeAllowanceMinutes(15, FundingModelTermTimeOnly, termDates, billingMonth, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 15 * 60 * 20 / 5 = 3600
	if minutes != 3600 {
		t.Errorf("expected 3600, got %d", minutes)
	}
}

func TestComputeAllowanceMinutes_Stretched(t *testing.T) {
	billingMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	minutes, err := ComputeAllowanceMinutes(15, FundingModelStretched, nil, billingMonth, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 15 * 60 * 38 / 12 = 2850
	if minutes != 2850 {
		t.Errorf("expected 2850, got %d", minutes)
	}
}

func TestComputeAllowanceMinutes_ProRating(t *testing.T) {
	billingMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	termDates := []TermDateRange{
		{StartDate: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC)},
	}

	minutes, err := ComputeAllowanceMinutes(15, FundingModelTermTimeOnly, termDates, billingMonth, nil, &startDate, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Jan 15-30 weekdays: 12 weekdays
	// 15 * 60 * 12 / 5 = 2160
	if minutes != 2160 {
		t.Errorf("expected 2160, got %d", minutes)
	}
}

func TestComputeAllowanceMinutes_ZeroHours(t *testing.T) {
	billingMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	minutes, err := ComputeAllowanceMinutes(0, FundingModelTermTimeOnly, nil, billingMonth, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 0 {
		t.Errorf("expected 0, got %d", minutes)
	}
}

func TestComputeAllowanceMinutes_FundingEndsBeforeMonth(t *testing.T) {
	billingMonth := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	minutes, err := ComputeAllowanceMinutes(15, FundingModelStretched, nil, billingMonth, nil, nil, &endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if minutes != 0 {
		t.Errorf("expected 0, got %d", minutes)
	}
}

func TestComputeAllowanceMinutes_UnknownModel(t *testing.T) {
	billingMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := ComputeAllowanceMinutes(15, FundingModelUnknown, nil, billingMonth, nil, nil, nil)
	if err == nil {
		t.Error("expected error for unknown funding model")
	}
}
