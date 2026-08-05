package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

func TestCoreSessionsComputer_PersistsEnrichedSessions(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600
	// June 2026: Mondays (4) and Tuesdays (4), each 300 min.
	entries := []domain.BookedPatternEntry{
		{DayOfWeek: 1, SessionType: domain.BookedSessionType{ID: uuid.New().String(), Name: "Full Day", StartMinutes: 480, EndMinutes: 780, DurationMinutes: 300}},
		{DayOfWeek: 2, SessionType: domain.BookedSessionType{ID: uuid.New().String(), Name: "Full Day", StartMinutes: 480, EndMinutes: 780, DurationMinutes: 300}},
	}

	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{bookingRow}},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: entries},
	}

	actor := tenant.ActorContext{TenantID: bookingRow.TenantID, BranchID: bookingRow.BranchID}
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare",
		QuantityMinutes: 3000,
		UnitAmountMinor: 600,
		LineAmountMinor: 30000, // 3000 min * £6/hr
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details == nil {
		t.Fatal("expected persisted details, got nil")
	}

	var core domain.CoreLineDetails
	if err := json.Unmarshal(details, &core); err != nil {
		t.Fatalf("unmarshal details: %v", err)
	}
	if len(core.BookedSessions) != 10 {
		t.Fatalf("expected 10 booked sessions, got %d", len(core.BookedSessions))
	}
	if core.BookedSessions[0].StartMinutes != 480 || core.BookedSessions[0].EndMinutes != 780 {
		t.Errorf("session start/end = %d/%d, want 480/780", core.BookedSessions[0].StartMinutes, core.BookedSessions[0].EndMinutes)
	}
	if core.BookedSessions[0].SessionAmountMinor != 3000 {
		t.Errorf("session amount = %d, want 3000", core.BookedSessions[0].SessionAmountMinor)
	}
	sum := 0
	for _, s := range core.BookedSessions {
		sum += s.SessionAmountMinor
	}
	if sum != line.LineAmountMinor {
		t.Errorf("session amounts sum to %d, want line amount %d", sum, line.LineAmountMinor)
	}
}

func TestCoreSessionsComputer_EnteredLineDivergesCollapses(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600
	entries := []domain.BookedPatternEntry{
		{DayOfWeek: 1, SessionType: domain.BookedSessionType{ID: uuid.New().String(), Name: "Full Day", DurationMinutes: 300}},
	}

	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{bookingRow}},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: entries},
	}

	actor := tenant.ActorContext{TenantID: bookingRow.TenantID, BranchID: bookingRow.BranchID}
	// Quantity edited to differ from the pattern-derived 1500 minutes.
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare",
		QuantityMinutes: 1800,
		UnitAmountMinor: 600,
		LineAmountMinor: 18000,
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != nil {
		t.Fatal("expected nil details when entered line diverges from pattern (breakdown collapsed)")
	}
}

func TestCoreSessionsComputer_NoBookingRowCollapses(t *testing.T) {
	childID := uuid.New()
	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{}},
		bookingEntriesLookup: &stubBookingEntriesLookup{},
	}

	actor := tenant.ActorContext{}
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare",
		QuantityMinutes: 1200,
		UnitAmountMinor: 600,
		LineAmountMinor: 12000,
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != nil {
		t.Fatal("expected nil details when child has no booking row")
	}
}

func TestCoreSessionsComputer_ZeroBookedSessionsCollapses(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600

	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{bookingRow}},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: []domain.BookedPatternEntry{}},
	}

	actor := tenant.ActorContext{TenantID: bookingRow.TenantID, BranchID: bookingRow.BranchID}
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare",
		QuantityMinutes: 0,
		UnitAmountMinor: 600,
		LineAmountMinor: 0,
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != nil {
		t.Fatal("expected nil details when month has zero booked sessions")
	}
}

func TestCoreSessionsComputer_DescriptionOnlyEditPreservesSessions(t *testing.T) {
	// Q2: a description-only edit does not invalidate the pattern-derived
	// breakdown, so sessions are preserved.
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600
	entries := []domain.BookedPatternEntry{
		{DayOfWeek: 1, SessionType: domain.BookedSessionType{ID: uuid.New().String(), Name: "Full Day", StartMinutes: 480, EndMinutes: 780, DurationMinutes: 300}},
	}

	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{bookingRow}},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: entries},
	}

	actor := tenant.ActorContext{TenantID: bookingRow.TenantID, BranchID: bookingRow.BranchID}
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare (renamed)",
		QuantityMinutes: 1500,
		UnitAmountMinor: 600,
		LineAmountMinor: 15000,
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details == nil {
		t.Fatal("expected persisted details on description-only edit (sessions preserved)")
	}
	var core domain.CoreLineDetails
	if err := json.Unmarshal(details, &core); err != nil {
		t.Fatalf("unmarshal details: %v", err)
	}
	if len(core.BookedSessions) == 0 {
		t.Fatal("expected booked sessions to be preserved on description-only edit")
	}
}

func TestCoreSessionsComputer_NoOpEditPreservesSessions(t *testing.T) {
	// A no-op save (identical quantity/unit/amount) preserves the breakdown.
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600
	entries := []domain.BookedPatternEntry{
		{DayOfWeek: 1, SessionType: domain.BookedSessionType{ID: uuid.New().String(), Name: "Full Day", StartMinutes: 480, EndMinutes: 780, DurationMinutes: 300}},
	}

	computer := &coreSessionsComputer{
		repo:                 &stubPrefillRepo{bookings: []domain.BillableChildRow{bookingRow}},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: entries},
	}

	actor := tenant.ActorContext{TenantID: bookingRow.TenantID, BranchID: bookingRow.BranchID}
	line := DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     "Core childcare",
		QuantityMinutes: 1500,
		UnitAmountMinor: 600,
		LineAmountMinor: 15000,
	}

	details, err := computer.computeCoreLineDetails(context.Background(), nil, actor, childID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details == nil {
		t.Fatal("expected persisted details on no-op edit (sessions preserved)")
	}
}
