package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term/domain"
)

func TestTermEndDateFor(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	got := domain.TermEndDateFor(start)
	want := time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("TermEndDateFor(%v) = %v, want %v", start, got, want)
	}

	// Crossing year boundary
	start = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	got = domain.TermEndDateFor(start)
	want = time.Date(2027, 11, 30, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("TermEndDateFor crossing year: got %v, want %v", got, want)
	}

	// Leap year
	start = time.Date(2028, 2, 1, 0, 0, 0, 0, time.UTC)
	got = domain.TermEndDateFor(start)
	want = time.Date(2029, 1, 31, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("TermEndDateFor leap: got %v, want %v", got, want)
	}
}

func TestValidateTermStartDate(t *testing.T) {
	good := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := domain.ValidateTermStartDate(good); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	bad := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if err := domain.ValidateTermStartDate(bad); err == nil {
		t.Fatal("expected error for day=15")
	} else if !strings.Contains(err.Error(), "1st of a calendar month") {
		t.Fatalf("error message wrong: %v", err)
	}

	notMidnight := time.Date(2026, 7, 1, 1, 0, 0, 0, time.UTC)
	if err := domain.ValidateTermStartDate(notMidnight); err == nil {
		t.Fatal("expected error for non-midnight")
	}
}

func TestNewTerm_RejectsBadInputs(t *testing.T) {
	_, err := domain.NewTerm(uuid.Nil, uuid.New(), uuid.New(), uuid.New(), time.Now(), uuid.New(), 100, uuid.New())
	if err == nil {
		t.Error("expected error for nil id")
	}
	_, err = domain.NewTerm(uuid.New(), uuid.Nil, uuid.New(), uuid.New(), time.Now(), uuid.New(), 100, uuid.New())
	if err == nil {
		t.Error("expected error for nil tenant")
	}
	_, err = domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), uuid.Nil, 100, uuid.New())
	if err == nil {
		t.Error("expected error for nil pattern")
	}
	_, err = domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Now(), uuid.New(), -1, uuid.New())
	if err == nil {
		t.Error("expected error for negative rate")
	}
	_, err = domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), uuid.New(), 100, uuid.New())
	if err == nil {
		t.Error("expected error for mid-month start")
	}
}

func TestNewTerm_FutureStartIsPreTerm(t *testing.T) {
	t1 := uuid.New()
	t2 := uuid.New()
	t3 := uuid.New()
	t4 := uuid.New()
	t5 := uuid.New()
	t6 := uuid.New()
	pattern := uuid.New()
	membership := uuid.New()
	// 2027-01-01 is in the future (today is post-2026-01-01).
	start := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(t1, t2, t3, t4, start, pattern, 100, membership)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if term.Status != domain.TermStatusPreTerm {
		t.Errorf("expected pre_term, got %s", term.Status)
	}
	if term.TermStartDate.IsZero() || term.TermEndDate.IsZero() {
		t.Error("dates not set")
	}
	_ = t5
	_ = t6
}

func TestNewTerm_PastStartIsActive(t *testing.T) {
	// 2025-01-01 is in the past
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), start, uuid.New(), 100, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if term.Status != domain.TermStatusActive {
		t.Errorf("expected active, got %s", term.Status)
	}
}

func TestTermShouldBeEnded(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), start, uuid.New(), 100, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	today := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) // day after term_end_date (2026-01-01 + 12 months - 1 day = 2025-12-31)
	if !term.ShouldBeEnded(today) {
		t.Error("expected ShouldBeEnded true")
	}
}

func TestTermShouldBePendingRenewal(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), start, uuid.New(), 100, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	// Today = 2025-12-02 (one day after term_end_date would be 2026-01-01, so the renewal window is days 30-0 before 2025-12-31)
	today := time.Date(2025, 12, 2, 0, 0, 0, 0, time.UTC)
	if !term.ShouldBePendingRenewal(today) {
		t.Error("expected pending_renewal true")
	}
	today = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if term.ShouldBePendingRenewal(today) {
		t.Error("expected pending_renewal false far from end date")
	}
}

func TestTermDeriveStatus(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(uuid.New(), uuid.New(), uuid.New(), uuid.New(), start, uuid.New(), 100, uuid.New())
	if err != nil {
		t.Fatal(err)
	}

	today := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	if got := term.DeriveStatus(today); got != domain.TermStatusPreTerm {
		t.Errorf("before start: got %s, want pre_term", got)
	}

	today = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if got := term.DeriveStatus(today); got != domain.TermStatusActive {
		t.Errorf("during: got %s, want active", got)
	}

	today = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if got := term.DeriveStatus(today); got != domain.TermStatusEnded {
		t.Errorf("after: got %s, want ended", got)
	}

	// Terminated is sticky
	term.Status = domain.TermStatusTerminated
	today = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	if got := term.DeriveStatus(today); got != domain.TermStatusTerminated {
		t.Errorf("terminated is sticky: got %s, want terminated", got)
	}
}
