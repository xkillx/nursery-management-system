package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAllocateLineAmount_EqualWeightsExact(t *testing.T) {
	// 8 equal weights and a £300.00 total distribute £37.50 each.
	weights := []int{1, 1, 1, 1, 1, 1, 1, 1}
	got := AllocateLineAmount(30000, weights)
	want := []int{3750, 3750, 3750, 3750, 3750, 3750, 3750, 3750}
	if len(got) != len(want) {
		t.Fatalf("got %d allocations, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestAllocateLineAmount_LargestRemainder(t *testing.T) {
	// weights [3,3,3] over total 10 -> [4,3,3] (largest remainder).
	got := AllocateLineAmount(10, []int{3, 3, 3})
	want := []int{4, 3, 3}
	if len(got) != len(want) {
		t.Fatalf("got %d allocations, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %d, want %d", i, got[i], want[i])
		}
	}
	sum := 0
	for _, g := range got {
		sum += g
	}
	if sum != 10 {
		t.Errorf("allocations sum to %d, want 10", sum)
	}
}

func TestAllocateLineAmount_ZeroTotal(t *testing.T) {
	got := AllocateLineAmount(0, []int{3, 3, 3})
	for i, g := range got {
		if g != 0 {
			t.Errorf("index %d: got %d, want 0", i, g)
		}
	}
}

func TestAllocateLineAmount_NegativeWeightsAbsolute(t *testing.T) {
	// Negative weights are treated as absolute values.
	got := AllocateLineAmount(10, []int{-3, -3, -3})
	want := []int{4, 3, 3}
	if len(got) != len(want) {
		t.Fatalf("got %d allocations, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestBuildSessionRows_SortsChronologically(t *testing.T) {
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedCoreMinutes: 900,
		BookedSessions: []BookedSession{
			{OccurrenceDate: timeMustParse("2026-07-20"), DurationMinutes: 300, SessionTypeName: "Late", StartMinutes: 12 * 60, EndMinutes: 17 * 60},
			{OccurrenceDate: timeMustParse("2026-07-06"), DurationMinutes: 300, SessionTypeName: "Early Week", StartMinutes: 8 * 60, EndMinutes: 13 * 60},
			{OccurrenceDate: timeMustParse("2026-07-20"), DurationMinutes: 300, SessionTypeName: "Morning", StartMinutes: 8 * 60, EndMinutes: 13 * 60},
		},
	})

	rows := BuildSessionRows(details, LineKindCoreChildcare, nil, 30000)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if !rows[0].OccurrenceDate.Equal(timeMustParse("2026-07-06")) {
		t.Errorf("row 0 date = %v, want 2026-07-06", rows[0].OccurrenceDate)
	}
	if !rows[1].OccurrenceDate.Equal(timeMustParse("2026-07-20")) || rows[1].StartMinutes != 8*60 {
		t.Errorf("row 1 = %v start=%d, want 2026-07-20 start=480", rows[1].OccurrenceDate, rows[1].StartMinutes)
	}
	if rows[2].StartMinutes != 12*60 {
		t.Errorf("row 2 start = %d, want 720", rows[2].StartMinutes)
	}
	// Equal durations -> equal split.
	for i, r := range rows {
		if r.SessionAmountMinor != 10000 {
			t.Errorf("row %d amount = %d, want 10000", i, r.SessionAmountMinor)
		}
	}
}

func TestBuildSessionRows_LegacyUntimedRow(t *testing.T) {
	// A legacy row lacking start/end minutes still renders (name-only) and
	// sorts after timed rows on the same date.
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedCoreMinutes: 600,
		BookedSessions: []BookedSession{
			{OccurrenceDate: timeMustParse("2026-07-20"), DurationMinutes: 300, SessionTypeName: "Legacy"},
			{OccurrenceDate: timeMustParse("2026-07-20"), DurationMinutes: 300, SessionTypeName: "Timed", StartMinutes: 8 * 60, EndMinutes: 13 * 60},
		},
	})

	rows := BuildSessionRows(details, LineKindCoreChildcare, nil, 20000)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].SessionTypeName != "Timed" {
		t.Errorf("row 0 = %q, want timed row first", rows[0].SessionTypeName)
	}
	if rows[1].SessionTypeName != "Legacy" {
		t.Errorf("row 1 = %q, want legacy row after timed", rows[1].SessionTypeName)
	}
	if rows[1].StartMinutes != 0 || rows[1].EndMinutes != 0 {
		t.Errorf("legacy row start/end = %d/%d, want 0/0", rows[1].StartMinutes, rows[1].EndMinutes)
	}
	if rows[0].SessionAmountMinor != rows[1].SessionAmountMinor {
		t.Errorf("amounts differ: %d vs %d", rows[0].SessionAmountMinor, rows[1].SessionAmountMinor)
	}
}

func TestBuildSessionRows_SkipsInvalidDateAndReallocates(t *testing.T) {
	// A row with a zero/invalid date is skipped; the line total is
	// re-allocated across the remaining rows.
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedCoreMinutes: 600,
		BookedSessions: []BookedSession{
			{OccurrenceDate: time.Time{}, DurationMinutes: 300, SessionTypeName: "Invalid"},
			{OccurrenceDate: timeMustParse("2026-07-06"), DurationMinutes: 300, SessionTypeName: "Valid"},
		},
	})

	rows := BuildSessionRows(details, LineKindCoreChildcare, nil, 20000)
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].SessionTypeName != "Valid" {
		t.Errorf("row name = %q, want Valid", rows[0].SessionTypeName)
	}
	if rows[0].SessionAmountMinor != 20000 {
		t.Errorf("amount = %d, want 20000 (whole line re-allocated)", rows[0].SessionAmountMinor)
	}
}

func TestBuildSessionRows_EmptyMalformedDetails(t *testing.T) {
	if rows := BuildSessionRows(nil, LineKindCoreChildcare, nil, 1000); rows != nil {
		t.Errorf("nil details: got %d rows, want nil", len(rows))
	}
	if rows := BuildSessionRows(json.RawMessage("not json"), LineKindCoreChildcare, nil, 1000); rows != nil {
		t.Errorf("malformed details: got %d rows, want nil", len(rows))
	}
	empty := marshalCoreLineDetailsForTest(t, CoreLineDetails{BookedCoreMinutes: 300})
	if rows := BuildSessionRows(empty, LineKindCoreChildcare, nil, 1000); rows != nil {
		t.Errorf("empty sessions: got %d rows, want nil", len(rows))
	}
}

func TestBuildSessionRows_NonCoreLine(t *testing.T) {
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedSessions: []BookedSession{{OccurrenceDate: timeMustParse("2026-07-06"), DurationMinutes: 300, SessionTypeName: "X"}},
	})
	if rows := BuildSessionRows(details, LineKindFundedDeduction, nil, 1000); rows != nil {
		t.Errorf("non-core line: got %d rows, want nil", len(rows))
	}
}

func TestBuildSessionRows_ReallocatesAgainstLineTotal(t *testing.T) {
	// Stored per-session amounts disagree with the passed line total; the
	// helper must re-normalize against the line total (largest remainder).
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedCoreMinutes: 600,
		BookedSessions: []BookedSession{
			{OccurrenceDate: timeMustParse("2026-07-06"), DurationMinutes: 300, SessionTypeName: "A", StartMinutes: 480, EndMinutes: 780, SessionAmountMinor: 1},
			{OccurrenceDate: timeMustParse("2026-07-13"), DurationMinutes: 300, SessionTypeName: "B", StartMinutes: 480, EndMinutes: 780, SessionAmountMinor: 1},
		},
	})

	rows := BuildSessionRows(details, LineKindCoreChildcare, nil, 40000)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].SessionAmountMinor != 20000 || rows[1].SessionAmountMinor != 20000 {
		t.Errorf("amounts = %d/%d, want 20000/20000", rows[0].SessionAmountMinor, rows[1].SessionAmountMinor)
	}
}

func TestBuildSessionRows_NonDivisibleAllocation(t *testing.T) {
	// weights [300, 300, 300] over £10 total -> deterministic largest-remainder.
	details := marshalCoreLineDetailsForTest(t, CoreLineDetails{
		BookedCoreMinutes: 900,
		BookedSessions: []BookedSession{
			{OccurrenceDate: timeMustParse("2026-07-06"), DurationMinutes: 300, SessionTypeName: "A"},
			{OccurrenceDate: timeMustParse("2026-07-13"), DurationMinutes: 300, SessionTypeName: "B"},
			{OccurrenceDate: timeMustParse("2026-07-20"), DurationMinutes: 300, SessionTypeName: "C"},
		},
	})

	rows := BuildSessionRows(details, LineKindCoreChildcare, nil, 10)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if rows[0].SessionAmountMinor != 4 || rows[1].SessionAmountMinor != 3 || rows[2].SessionAmountMinor != 3 {
		t.Errorf("amounts = %d/%d/%d, want 4/3/3", rows[0].SessionAmountMinor, rows[1].SessionAmountMinor, rows[2].SessionAmountMinor)
	}
}

func marshalCoreLineDetailsForTest(t *testing.T, d CoreLineDetails) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal details: %v", err)
	}
	return b
}
