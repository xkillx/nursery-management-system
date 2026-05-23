package application

import (
	"testing"
	"time"
)

func TestLondonLocalDate_MidnightBoundary(t *testing.T) {
	// 2025-01-15 00:00:00 UTC = 2025-01-15 00:00:00 London (no DST)
	utc := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	local := LondonLocalDate(utc)
	y, m, d := local.Date()
	if y != 2025 || m != 1 || d != 15 {
		t.Fatalf("expected 2025-01-15, got %d-%02d-%02d", y, m, d)
	}
}

func TestLondonLocalDate_BeforeMidnightIsPreviousDay(t *testing.T) {
	// 2025-01-15 23:59:59 UTC = 2025-01-15 23:59:59 London (no DST)
	utc := time.Date(2025, 1, 15, 23, 59, 59, 0, time.UTC)
	local := LondonLocalDate(utc)
	y, m, d := local.Date()
	if y != 2025 || m != 1 || d != 15 {
		t.Fatalf("expected 2025-01-15, got %d-%02d-%02d", y, m, d)
	}
}

func TestLondonLocalDate_DSTBoundary(t *testing.T) {
	// BST starts 2025-03-30 01:00 UTC = 2025-03-30 02:00 BST
	// 2025-03-30 00:30 UTC = still 2025-03-30 00:30 GMT (before DST switch)
	utc := time.Date(2025, 3, 30, 0, 30, 0, 0, time.UTC)
	local := LondonLocalDate(utc)
	y, m, d := local.Date()
	if y != 2025 || m != 3 || d != 30 {
		t.Fatalf("expected 2025-03-30, got %d-%02d-%02d", y, m, d)
	}

	// After DST switch: 2025-03-30 01:30 UTC = 2025-03-30 02:30 BST
	utc2 := time.Date(2025, 3, 30, 1, 30, 0, 0, time.UTC)
	local2 := LondonLocalDate(utc2)
	_, _, d2 := local2.Date()
	if d2 != 30 {
		t.Fatalf("expected day 30, got %d", d2)
	}
}

func TestLondonLocalDate_ClockForwardHour(t *testing.T) {
	// 2025-03-30 00:59 UTC = 2025-03-30 00:59 GMT
	before := time.Date(2025, 3, 30, 0, 59, 0, 0, time.UTC)
	local := LondonLocalDate(before)
	_, _, d := local.Date()
	if d != 30 {
		t.Fatalf("expected day 30, got %d", d)
	}

	// 2025-03-30 01:01 UTC = 2025-03-30 02:01 BST (clock jumped forward)
	after := time.Date(2025, 3, 30, 1, 1, 0, 0, time.UTC)
	local2 := LondonLocalDate(after)
	_, _, d2 := local2.Date()
	if d2 != 30 {
		t.Fatalf("expected day 30, got %d", d2)
	}
}

func TestLondonDateOnly_TruncatesTime(t *testing.T) {
	utc := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
	dateOnly := LondonDateOnly(utc)
	h, m, s := dateOnly.Clock()
	if h != 0 || m != 0 || s != 0 {
		t.Fatalf("expected zero time, got %02d:%02d:%02d", h, m, s)
	}
	y, mo, d := dateOnly.Date()
	if y != 2025 || mo != 6 || d != 15 {
		t.Fatalf("expected 2025-06-15, got %d-%02d-%02d", y, mo, d)
	}
}

func TestAttendanceClock_InjectableNow(t *testing.T) {
	fixed := time.Date(2025, 12, 25, 10, 0, 0, 0, time.UTC)
	clock := NewAttendanceClock(func() time.Time { return fixed })
	utc, local := clock.Now()
	if !utc.Equal(fixed) {
		t.Fatalf("expected %v, got %v", fixed, utc)
	}
	y, m, d := local.Date()
	if y != 2025 || m != 12 || d != 25 {
		t.Fatalf("expected 2025-12-25, got %d-%02d-%02d", y, m, d)
	}
}

func TestAttendanceClock_LocalDate(t *testing.T) {
	fixed := time.Date(2025, 6, 1, 23, 0, 0, 0, time.UTC)
	clock := NewAttendanceClock(func() time.Time { return fixed })
	local := clock.LocalDate(fixed)
	y, m, d := local.Date()
	// June 1, 23:00 UTC = June 2, 00:00 BST
	if y != 2025 || m != 6 || d != 2 {
		t.Fatalf("expected 2025-06-02 (BST), got %d-%02d-%02d", y, m, d)
	}
}
