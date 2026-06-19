package domain

import "testing"

func TestCalculateBookedCoreMinutesInMonth(t *testing.T) {
	// A pattern with one entry: Monday, 5-hour (300 min) session.
	entries := []BookedPatternEntry{
		{
			DayOfWeek: 1, // Monday
			SessionType: BookedSessionType{
				ID:              "st1",
				Name:            "Full Day",
				StartMinutes:    8 * 60,
				EndMinutes:      13 * 60,
				DurationMinutes: 5 * 60,
			},
		},
	}
	// July 2026 has 4 Mondays: 6, 13, 20, 27.
	calc, err := CalculateBookedCoreMinutesInMonth("p1", entries, timeMustParse("2026-07-01"), 750)
	if err != nil {
		t.Fatal(err)
	}
	if calc.TotalMinutes != 4*300 {
		t.Errorf("expected 4 * 300 = %d, got %d", 4*300, calc.TotalMinutes)
	}
	if len(calc.Sessions) != 4 {
		t.Errorf("expected 4 sessions, got %d", len(calc.Sessions))
	}
	wantSubtotal, _ := CalculateHourlyAmountMinor(4*300, 750)
	if calc.SubtotalMinor != wantSubtotal {
		t.Errorf("subtotal: got %d, want %d", calc.SubtotalMinor, wantSubtotal)
	}
}

func TestCalculateBookedCoreMinutesInMonth_MultipleDays(t *testing.T) {
	// Mon + Wed + Fri, 3-hour sessions.
	entries := []BookedPatternEntry{
		{DayOfWeek: 1, SessionType: BookedSessionType{ID: "st1", Name: "Mon", DurationMinutes: 180}},
		{DayOfWeek: 3, SessionType: BookedSessionType{ID: "st2", Name: "Wed", DurationMinutes: 180}},
		{DayOfWeek: 5, SessionType: BookedSessionType{ID: "st3", Name: "Fri", DurationMinutes: 180}},
	}
	// July 2026: Mondays=4, Wednesdays=5, Fridays=5 → 14 * 180 = 2520
	calc, err := CalculateBookedCoreMinutesInMonth("p1", entries, timeMustParse("2026-07-01"), 1000)
	if err != nil {
		t.Fatal(err)
	}
	if calc.TotalMinutes != 14*180 {
		t.Errorf("expected %d, got %d", 14*180, calc.TotalMinutes)
	}
}

func TestCalculateBookedCoreMinutesInMonth_DayOfWeekConversion(t *testing.T) {
	// Day 7 = Sunday per the schema constraint.
	entries := []BookedPatternEntry{
		{DayOfWeek: 7, SessionType: BookedSessionType{ID: "st1", Name: "Sun", DurationMinutes: 60}},
	}
	// July 2026: Sundays are 5, 12, 19, 26 = 4.
	calc, _ := CalculateBookedCoreMinutesInMonth("p1", entries, timeMustParse("2026-07-01"), 1000)
	if calc.TotalMinutes != 4*60 {
		t.Errorf("Sunday count: got %d, want %d", calc.TotalMinutes, 4*60)
	}
}

func TestCalculateBookedCoreMinutesInMonth_FebruaryLeapYear(t *testing.T) {
	// Feb 2028 has 29 days. Mondays in Feb 2028: 7, 14, 21, 28 = 4.
	entries := []BookedPatternEntry{
		{DayOfWeek: 1, SessionType: BookedSessionType{ID: "st1", Name: "M", DurationMinutes: 60}},
	}
	calc, _ := CalculateBookedCoreMinutesInMonth("p1", entries, timeMustParse("2028-02-01"), 1000)
	if calc.TotalMinutes != 4*60 {
		t.Errorf("Feb leap year: got %d, want %d", calc.TotalMinutes, 4*60)
	}
}

func TestCalculateBookedCoreMinutesInMonth_BadDayOfWeek(t *testing.T) {
	entries := []BookedPatternEntry{
		{DayOfWeek: 0, SessionType: BookedSessionType{ID: "st1", Name: "X", DurationMinutes: 60}},
	}
	_, err := CalculateBookedCoreMinutesInMonth("p1", entries, timeMustParse("2026-07-01"), 1000)
	if err == nil {
		t.Error("expected error for day_of_week=0")
	}
}

func TestComputeFundedDeductionMinor(t *testing.T) {
	// booked == funded: deduction = booked; billable = 0
	fdm, bm, fdmMinor, bmMinor, err := ComputeFundedDeductionMinor(600, 600, 700) // 600/60*700=7000
	if err != nil {
		t.Fatal(err)
	}
	if fdm != 600 || bm != 0 || fdmMinor != 7000 || bmMinor != 0 {
		t.Errorf("got fdm=%d bm=%d fdmMinor=%d bmMinor=%d", fdm, bm, fdmMinor, bmMinor)
	}

	// Booked < funded: deduction = booked; billable = 0
	fdm, bm, fdmMinor, bmMinor, err = ComputeFundedDeductionMinor(300, 600, 700)
	if err != nil {
		t.Fatal(err)
	}
	if fdm != 300 || bm != 0 {
		t.Errorf("booked<funded: got fdm=%d bm=%d, want 300 0", fdm, bm)
	}

	// Booked > funded: deduction = funded; billable = (booked - funded)
	fdm, bm, fdmMinor, bmMinor, err = ComputeFundedDeductionMinor(1200, 600, 700)
	if err != nil {
		t.Fatal(err)
	}
	if fdm != 600 || bm != 600 {
		t.Errorf("expected fdm=600 bm=600, got fdm=%d bm=%d", fdm, bm)
	}
	wantFdm, _ := CalculateHourlyAmountMinor(600, 700)
	wantBm, _ := CalculateHourlyAmountMinor(600, 700)
	if fdmMinor != wantFdm || bmMinor != wantBm {
		t.Errorf("expected fdmMinor=%d bmMinor=%d, got fdmMinor=%d bmMinor=%d", wantFdm, wantBm, fdmMinor, bmMinor)
	}
}
