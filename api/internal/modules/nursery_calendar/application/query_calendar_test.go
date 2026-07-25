package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/nursery_calendar/domain"
)

type mockClosureLookup struct {
	dates   []time.Time
	listErr error
}

func (m *mockClosureLookup) GetClosureDatesForBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]time.Time, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var out []time.Time
	for _, d := range m.dates {
		if !d.Before(from) && !d.After(to) {
			out = append(out, d)
		}
	}
	return out, nil
}

type mockHolidayLookup struct {
	periods []domain.HolidayPeriodDateRange
	listErr error
}

func (m *mockHolidayLookup) GetHolidayPeriodsForBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]domain.HolidayPeriodDateRange, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var out []domain.HolidayPeriodDateRange
	for _, p := range m.periods {
		if !p.EndDate.Before(from) && !p.StartDate.After(to) {
			out = append(out, p)
		}
	}
	return out, nil
}

func TestQueryCalendarDay_ClosureDay(t *testing.T) {
	closureDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	uc := NewQueryCalendarDay(
		&mockClosureLookup{dates: []time.Time{closureDate}},
		&mockHolidayLookup{},
	)

	result, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), closureDate, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsOpen {
		t.Error("expected closed on closure day")
	}
	if result.Reason != domain.ClosureReasonClosureDay {
		t.Errorf("expected closure_day reason, got %s", result.Reason)
	}
}

func TestQueryCalendarDay_HolidayPeriod_TermTime(t *testing.T) {
	holidayStart := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	holidayEnd := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	checkDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)

	uc := NewQueryCalendarDay(
		&mockClosureLookup{},
		&mockHolidayLookup{periods: []domain.HolidayPeriodDateRange{
			{StartDate: holidayStart, EndDate: holidayEnd},
		}},
	)

	result, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), checkDate, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsOpen {
		t.Error("expected closed during holiday period for term-time booking")
	}
	if result.Reason != domain.ClosureReasonHolidayPeriod {
		t.Errorf("expected holiday_period reason, got %s", result.Reason)
	}
}

func TestQueryCalendarDay_HolidayPeriod_Stretched(t *testing.T) {
	holidayStart := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	holidayEnd := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	checkDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)

	uc := NewQueryCalendarDay(
		&mockClosureLookup{},
		&mockHolidayLookup{periods: []domain.HolidayPeriodDateRange{
			{StartDate: holidayStart, EndDate: holidayEnd},
		}},
	)

	result, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), checkDate, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsOpen {
		t.Error("expected open during holiday period for stretched booking")
	}
	if result.Reason != domain.ClosureReasonNone {
		t.Errorf("expected none reason, got %s", result.Reason)
	}
}

func TestQueryCalendarDay_OpenDay(t *testing.T) {
	uc := NewQueryCalendarDay(
		&mockClosureLookup{},
		&mockHolidayLookup{},
	)

	checkDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	result, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), checkDate, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsOpen {
		t.Error("expected open on regular day")
	}
	if result.Reason != domain.ClosureReasonNone {
		t.Errorf("expected none reason, got %s", result.Reason)
	}
}

func TestQueryCalendarDay_ClosureOverridesHoliday(t *testing.T) {
	closureDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	holidayStart := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	holidayEnd := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	uc := NewQueryCalendarDay(
		&mockClosureLookup{dates: []time.Time{closureDate}},
		&mockHolidayLookup{periods: []domain.HolidayPeriodDateRange{
			{StartDate: holidayStart, EndDate: holidayEnd},
		}},
	)

	result, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), closureDate, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsOpen {
		t.Error("expected closed")
	}
	if result.Reason != domain.ClosureReasonClosureDay {
		t.Errorf("closure day should take priority over holiday period, got %s", result.Reason)
	}
}

func TestQueryDateRange_MixedDays(t *testing.T) {
	closureDate := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	holidayStart := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	holidayEnd := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	uc := NewQueryDateRange(
		&mockClosureLookup{dates: []time.Time{closureDate}},
		&mockHolidayLookup{periods: []domain.HolidayPeriodDateRange{
			{StartDate: holidayStart, EndDate: holidayEnd},
		}},
	)

	from := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC)

	results, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 9 {
		t.Fatalf("expected 9 days, got %d", len(results))
	}

	// July 22 is a closure day
	closureDay := results[2]
	if closureDay.IsOpen || closureDay.Reason != domain.ClosureReasonClosureDay {
		t.Errorf("July 22 should be closure day, got open=%v reason=%s", closureDay.IsOpen, closureDay.Reason)
	}

	// July 24-26 are holiday period
	for i := 4; i <= 6; i++ {
		if results[i].IsOpen || results[i].Reason != domain.ClosureReasonHolidayPeriod {
			t.Errorf("July %d should be holiday period, got open=%v reason=%s", 20+i, results[i].IsOpen, results[i].Reason)
		}
	}

	// July 20 is open
	if !results[0].IsOpen || results[0].Reason != domain.ClosureReasonNone {
		t.Errorf("July 20 should be open, got open=%v reason=%s", results[0].IsOpen, results[0].Reason)
	}
}

func TestQueryDateRange_EmptyClosures(t *testing.T) {
	uc := NewQueryDateRange(
		&mockClosureLookup{},
		&mockHolidayLookup{},
	)

	from := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)

	results, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), from, to, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 days, got %d", len(results))
	}
	for _, r := range results {
		if !r.IsOpen || r.Reason != domain.ClosureReasonNone {
			t.Errorf("expected all days open, got open=%v reason=%s", r.IsOpen, r.Reason)
		}
	}
}
