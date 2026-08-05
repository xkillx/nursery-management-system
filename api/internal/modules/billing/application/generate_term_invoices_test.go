package application

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
)

func TestEnrichBookedSessions_PersistsAllocatedAmounts(t *testing.T) {
	// Two same-duration sessions on different days get equal allocated
	// amounts; the amounts sum exactly to the line subtotal.
	calc := domain.BookedCoreCalculation{
		TotalMinutes: 600,
		Subtotal:     domain.MustGBP(30000),
		Sessions: []domain.BookedSession{
			{OccurrenceDate: timeMustParseApp("2026-07-06"), DurationMinutes: 300, SessionTypeName: "Full Day", StartMinutes: 480, EndMinutes: 780},
			{OccurrenceDate: timeMustParseApp("2026-07-13"), DurationMinutes: 300, SessionTypeName: "Full Day", StartMinutes: 480, EndMinutes: 780},
		},
	}

	enriched := enrichBookedSessions(calc)
	if len(enriched) != 2 {
		t.Fatalf("got %d sessions, want 2", len(enriched))
	}
	if enriched[0].SessionAmountMinor != 15000 || enriched[1].SessionAmountMinor != 15000 {
		t.Errorf("amounts = %d/%d, want 15000/15000", enriched[0].SessionAmountMinor, enriched[1].SessionAmountMinor)
	}
	if enriched[0].StartMinutes != 480 || enriched[1].EndMinutes != 780 {
		t.Errorf("start/end not preserved: %d/%d, %d/%d", enriched[0].StartMinutes, enriched[0].EndMinutes, enriched[1].StartMinutes, enriched[1].EndMinutes)
	}
	sum := enriched[0].SessionAmountMinor + enriched[1].SessionAmountMinor
	if sum != calc.Subtotal.Minor() {
		t.Errorf("amounts sum to %d, want line subtotal %d", sum, calc.Subtotal.Minor())
	}
}

func TestEnrichBookedSessions_NonDivisibleLargestRemainder(t *testing.T) {
	// A non-divisible case lands on a deterministic largest-remainder split.
	calc := domain.BookedCoreCalculation{
		TotalMinutes: 900,
		Subtotal:     domain.MustGBP(10),
		Sessions: []domain.BookedSession{
			{OccurrenceDate: timeMustParseApp("2026-07-06"), DurationMinutes: 300, SessionTypeName: "A"},
			{OccurrenceDate: timeMustParseApp("2026-07-13"), DurationMinutes: 300, SessionTypeName: "B"},
			{OccurrenceDate: timeMustParseApp("2026-07-20"), DurationMinutes: 300, SessionTypeName: "C"},
		},
	}

	enriched := enrichBookedSessions(calc)
	want := []int{4, 3, 3}
	for i, w := range want {
		if enriched[i].SessionAmountMinor != w {
			t.Errorf("session %d amount = %d, want %d", i, enriched[i].SessionAmountMinor, w)
		}
	}
}

func TestEnrichBookedSessions_MarshalsToGoStyleStorageKeys(t *testing.T) {
	// The persisted details use Go-style storage keys (StartMinutes,
	// EndMinutes, SessionAmountMinor) matching the existing untagged keys.
	calc := domain.BookedCoreCalculation{
		TotalMinutes: 300,
		Subtotal:     domain.MustGBP(20000),
		Sessions: []domain.BookedSession{
			{OccurrenceDate: timeMustParseApp("2026-07-06"), DurationMinutes: 300, SessionTypeName: "Full Day", StartMinutes: 480, EndMinutes: 780},
		},
	}

	coreLineDetails := domain.CoreLineDetails{
		BookedCoreMinutes: calc.TotalMinutes,
		BookedSessions:    enrichBookedSessions(calc),
		BookedPerEntry:    calc.PerEntry,
	}
	detailsJSON, err := json.Marshal(coreLineDetails)
	if err != nil {
		t.Fatalf("marshal details: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(detailsJSON, &raw); err != nil {
		t.Fatalf("unmarshal details: %v", err)
	}
	sessionJSON := string(raw["booked_sessions"])
	for _, key := range []string{"OccurrenceDate", "DurationMinutes", "StartMinutes", "EndMinutes", "SessionAmountMinor", "SessionTypeName"} {
		if !containsJSONKey(sessionJSON, key) {
			t.Errorf("persisted session JSON missing Go-style key %q: %s", key, sessionJSON)
		}
	}

	// Round-trips through BuildSessionRows with identical amounts.
	rows := domain.BuildSessionRows(detailsJSON, domain.LineKindCoreChildcare, nil, calc.Subtotal.Minor())
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].SessionAmountMinor != 20000 {
		t.Errorf("row amount = %d, want 20000", rows[0].SessionAmountMinor)
	}
}

func containsJSONKey(s, key string) bool {
	return strings.Contains(s, key)
}

func timeMustParseApp(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}
