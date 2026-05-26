package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}

func londonTime(t *testing.T, year int, month time.Month, day, hour, min, sec int) time.Time {
	t.Helper()
	return time.Date(year, month, day, hour, min, sec, 0, londonLoc)
}

func TestNewBillingPeriod(t *testing.T) {
	t.Run("basic month", func(t *testing.T) {
		bp, err := NewBillingPeriod(2025, time.May)
		if err != nil {
			t.Fatal(err)
		}
		if bp.StartLocal.Year() != 2025 || bp.StartLocal.Month() != time.May || bp.StartLocal.Day() != 1 {
			t.Fatalf("start: got %v", bp.StartLocal)
		}
		if bp.EndExclusiveLocal.Year() != 2025 || bp.EndExclusiveLocal.Month() != time.June || bp.EndExclusiveLocal.Day() != 1 {
			t.Fatalf("end: got %v", bp.EndExclusiveLocal)
		}
	})

	t.Run("December wraps year", func(t *testing.T) {
		bp, err := NewBillingPeriod(2025, time.December)
		if err != nil {
			t.Fatal(err)
		}
		if bp.EndExclusiveLocal.Year() != 2026 || bp.EndExclusiveLocal.Month() != time.January {
			t.Fatalf("end: got %v", bp.EndExclusiveLocal)
		}
	})

	t.Run("DST March boundary", func(t *testing.T) {
		bp, err := NewBillingPeriod(2025, time.March)
		if err != nil {
			t.Fatal(err)
		}
		loc, _ := time.LoadLocation("Europe/London")
		start := time.Date(2025, time.March, 1, 0, 0, 0, 0, loc)
		if !bp.StartLocal.Equal(start) {
			t.Fatalf("march start: got %v want %v", bp.StartLocal, start)
		}
		end := time.Date(2025, time.April, 1, 0, 0, 0, 0, loc)
		if !bp.EndExclusiveLocal.Equal(end) {
			t.Fatalf("march end: got %v want %v", bp.EndExclusiveLocal, end)
		}
	})

	t.Run("DST October boundary", func(t *testing.T) {
		bp, err := NewBillingPeriod(2025, time.October)
		if err != nil {
			t.Fatal(err)
		}
		loc, _ := time.LoadLocation("Europe/London")
		start := time.Date(2025, time.October, 1, 0, 0, 0, 0, loc)
		if !bp.StartLocal.Equal(start) {
			t.Fatalf("oct start: got %v want %v", bp.StartLocal, start)
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		cases := []struct {
			year  int
			month time.Month
		}{
			{0, time.January},
			{-1, time.May},
		}
		for _, tc := range cases {
			_, err := NewBillingPeriod(tc.year, tc.month)
			if err == nil {
				t.Errorf("year=%d month=%d: expected error", tc.year, tc.month)
			}
		}
	})
}

func TestCalculateAttendanceMinutes(t *testing.T) {
	sid := func() uuid.UUID { return uuid.Must(uuid.NewV7()) }

	t.Run("same-day complete session", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 16, 0, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RawElapsedMinutes != 480 {
			t.Errorf("raw: got %d want 480", result.RawElapsedMinutes)
		}
		if result.RoundedBillableMinutes != 480 {
			t.Errorf("rounded: got %d want 480", result.RoundedBillableMinutes)
		}
		if result.IncludedSessionCount != 1 {
			t.Errorf("count: got %d want 1", result.IncludedSessionCount)
		}
	})

	t.Run("non-boundary rounding up", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 8, 15, 1)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RoundedBillableMinutes != 30 {
			t.Errorf("rounded: got %d want 30", result.RoundedBillableMinutes)
		}
	})

	t.Run("non-boundary exact 15 minutes", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 8, 15, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RoundedBillableMinutes != 15 {
			t.Errorf("rounded: got %d want 15", result.RoundedBillableMinutes)
		}
	})

	t.Run("14m59s rounds to 15", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 8, 14, 59)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RoundedBillableMinutes != 15 {
			t.Errorf("rounded: got %d want 15", result.RoundedBillableMinutes)
		}
	})

	t.Run("cross-midnight same month", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 23, 30, 0)
		checkOut := londonTime(t, 2025, time.May, 16, 0, 10, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RawElapsedMinutes != 40 {
			t.Errorf("raw: got %d want 40", result.RawElapsedMinutes)
		}
		if result.RoundedBillableMinutes != 45 {
			t.Errorf("rounded: got %d want 45", result.RoundedBillableMinutes)
		}
	})

	t.Run("cross-month check-in-month allocation April session", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.April, 30, 23, 30, 0)
		checkOut := londonTime(t, 2025, time.May, 1, 0, 10, 0)

		// Included for April
		aprResult, err := CalculateAttendanceMinutes(2025, time.April, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if aprResult.IncludedSessionCount != 1 {
			t.Errorf("April count: got %d want 1", aprResult.IncludedSessionCount)
		}

		// Excluded from May
		mayResult, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if mayResult.IncludedSessionCount != 0 {
			t.Errorf("May count: got %d want 0", mayResult.IncludedSessionCount)
		}
	})

	t.Run("session starts before target month ends inside", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.April, 30, 20, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 1, 8, 0, 0)

		mayResult, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if mayResult.IncludedSessionCount != 0 {
			t.Errorf("May count: got %d want 0", mayResult.IncludedSessionCount)
		}
	})

	t.Run("incomplete open session excluded", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusOpen, CheckInAt: checkIn, CheckOutAt: nil},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IncludedSessionCount != 0 {
			t.Errorf("count: got %d want 0", result.IncludedSessionCount)
		}
		if len(result.ExcludedIncompleteSessions) != 1 {
			t.Fatalf("exclusions: got %d want 1", len(result.ExcludedIncompleteSessions))
		}
		if result.ExcludedIncompleteSessions[0].SessionID != id {
			t.Errorf("exclusion session id mismatch")
		}
		if result.RawElapsedMinutes != 0 || result.RoundedBillableMinutes != 0 {
			t.Errorf("totals should be zero: raw=%d rounded=%d", result.RawElapsedMinutes, result.RoundedBillableMinutes)
		}
	})

	t.Run("corrected session uses supplied interval", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 9, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 12, 30, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusCorrected, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RawElapsedMinutes != 210 {
			t.Errorf("raw: got %d want 210", result.RawElapsedMinutes)
		}
		if result.RoundedBillableMinutes != 210 {
			t.Errorf("rounded: got %d want 210", result.RoundedBillableMinutes)
		}
	})

	t.Run("multiple sessions sum individually rounded", func(t *testing.T) {
		id1, id2 := sid(), sid()
		checkIn1 := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut1 := londonTime(t, 2025, time.May, 15, 8, 1, 0)
		checkIn2 := londonTime(t, 2025, time.May, 15, 9, 0, 0)
		checkOut2 := londonTime(t, 2025, time.May, 15, 9, 1, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id1, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn1, CheckOutAt: &checkOut1},
			{SessionID: id2, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn2, CheckOutAt: &checkOut2},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.RoundedBillableMinutes != 30 {
			t.Errorf("rounded total: got %d want 30 (15+15)", result.RoundedBillableMinutes)
		}
		if result.IncludedSessionCount != 2 {
			t.Errorf("count: got %d want 2", result.IncludedSessionCount)
		}
	})

	t.Run("error checkout equal to checkin", func(t *testing.T) {
		id := sid()
		ts := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		_, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: ts, CheckOutAt: &ts},
		})
		if err == nil {
			t.Fatal("expected error for checkout == checkin")
		}
	})

	t.Run("error checkout before checkin", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 7, 0, 0)
		_, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err == nil {
			t.Fatal("expected error for checkout before checkin")
		}
	})

	t.Run("error open with checkout", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 16, 0, 0)
		_, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusOpen, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err == nil {
			t.Fatal("expected error for open with checkout")
		}
	})

	t.Run("error unknown status", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		_, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: "unknown", CheckInAt: checkIn, CheckOutAt: nil},
		})
		if err == nil {
			t.Fatal("expected error for unknown status")
		}
	})

	t.Run("complete without checkout excluded as incomplete", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: nil},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IncludedSessionCount != 0 {
			t.Errorf("count: got %d want 0", result.IncludedSessionCount)
		}
		if len(result.ExcludedIncompleteSessions) != 1 {
			t.Fatalf("exclusions: got %d want 1", len(result.ExcludedIncompleteSessions))
		}
	})

	t.Run("DST March clock-forward allocation", func(t *testing.T) {
		// DST 2025: clocks forward March 30 01:00 → 02:00 London
		id := sid()
		// Session 23:30 March 29 to 00:30 March 30 London
		checkIn := londonTime(t, 2025, time.March, 29, 23, 30, 0)
		checkOut := londonTime(t, 2025, time.March, 30, 0, 30, 0)
		result, err := CalculateAttendanceMinutes(2025, time.March, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IncludedSessionCount != 1 {
			t.Errorf("count: got %d want 1", result.IncludedSessionCount)
		}
		if result.RawElapsedMinutes != 60 {
			t.Errorf("raw: got %d want 60", result.RawElapsedMinutes)
		}
	})

	t.Run("DST October clock-back allocation", func(t *testing.T) {
		// DST 2025: clocks back October 26 02:00 → 01:00 London
		id := sid()
		checkIn := londonTime(t, 2025, time.October, 25, 23, 0, 0)
		checkOut := londonTime(t, 2025, time.October, 26, 1, 0, 0)
		result, err := CalculateAttendanceMinutes(2025, time.October, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IncludedSessionCount != 1 {
			t.Errorf("count: got %d want 1", result.IncludedSessionCount)
		}
	})

	t.Run("empty sessions returns zero totals", func(t *testing.T) {
		result, err := CalculateAttendanceMinutes(2025, time.May, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.RawElapsedMinutes != 0 || result.RoundedBillableMinutes != 0 || result.IncludedSessionCount != 0 {
			t.Errorf("expected all zeros: raw=%d rounded=%d count=%d", result.RawElapsedMinutes, result.RoundedBillableMinutes, result.IncludedSessionCount)
		}
	})

	t.Run("session allocation fields populated", func(t *testing.T) {
		id := sid()
		checkIn := londonTime(t, 2025, time.May, 15, 8, 0, 0)
		checkOut := londonTime(t, 2025, time.May, 15, 9, 0, 0)
		result, err := CalculateAttendanceMinutes(2025, time.May, []AttendanceSessionInput{
			{SessionID: id, Status: AttendanceSessionStatusComplete, CheckInAt: checkIn, CheckOutAt: &checkOut},
		})
		if err != nil {
			t.Fatal(err)
		}
		s := result.Sessions[0]
		if s.AllocationYear != 2025 {
			t.Errorf("alloc year: got %d want 2025", s.AllocationYear)
		}
		if s.AllocationMonth != time.May {
			t.Errorf("alloc month: got %v want May", s.AllocationMonth)
		}
		if s.RawElapsedDuration != time.Hour {
			t.Errorf("duration: got %v want 1h", s.RawElapsedDuration)
		}
	})
}

func TestRoundDurationUpToBlockMinutes(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected int
	}{
		{15 * time.Minute, 15},
		{15*time.Minute + time.Second, 30},
		{14*time.Minute + 59*time.Second, 15},
		{30 * time.Minute, 30},
		{0, 0},
		{time.Nanosecond, 15},
		{480 * time.Minute, 480},
	}
	for _, tc := range tests {
		got := roundDurationUpToBlockMinutes(tc.input)
		if got != tc.expected {
			t.Errorf("roundDurationUp(%v) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}
