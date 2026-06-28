package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
)

func TestListCorrectionSessions_DelegatesToRepo(t *testing.T) {
	childID := uuid.New()
	localDate := time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)
	actor := makeActor()

	expected := domain.CorrectionSessionContext{
		ChildID:           childID,
		SelectedLocalDate: localDate,
		Sessions:          []domain.Session{},
	}

	repo := &stubListSessionsRepo{result: expected}
	uc := NewListCorrectionSessions(repo)

	got, err := uc.Execute(context.Background(), actor, childID, localDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ChildID != childID {
		t.Fatalf("expected child_id %s, got %s", childID, got.ChildID)
	}
	if len(got.Sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(got.Sessions))
	}
	if got.InvoiceWarning != nil {
		t.Fatal("expected no invoice warning")
	}
}

func TestListCorrectionSessions_PassesInvoiceWarning(t *testing.T) {
	childID := uuid.New()
	localDate := time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)
	actor := makeActor()

	warning := &domain.IssuedInvoiceWarning{
		InvoiceID:     uuid.New(),
		InvoiceNumber: "INV-202606-0001",
		BillingMonth:  "2026-06",
		Status:        "issued",
	}

	repo := &stubListSessionsRepo{result: domain.CorrectionSessionContext{
		ChildID:           childID,
		SelectedLocalDate: localDate,
		InvoiceWarning:    warning,
		Sessions:          []domain.Session{},
	}}
	uc := NewListCorrectionSessions(repo)

	got, err := uc.Execute(context.Background(), actor, childID, localDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.InvoiceWarning == nil {
		t.Fatal("expected invoice warning")
	}
	if got.InvoiceWarning.InvoiceNumber != "INV-202606-0001" {
		t.Fatalf("expected invoice number INV-202606-0001, got %s", got.InvoiceWarning.InvoiceNumber)
	}
}

type stubListSessionsRepo struct {
	result domain.CorrectionSessionContext
	err    error
}

func (s *stubListSessionsRepo) ListSessionsForCorrection(_ context.Context, _, _, _ uuid.UUID, _ time.Time) (domain.CorrectionSessionContext, error) {
	return s.result, s.err
}

func (s *stubListSessionsRepo) ListCorrectionHistory(_ context.Context, _, _, _ uuid.UUID) (domain.Session, []domain.CorrectionHistoryEvent, error) {
	return domain.Session{}, nil, nil
}

func (s *stubListSessionsRepo) CreateOpenSessionWithEvent(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubListSessionsRepo) GetOpenSessionForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (s *stubListSessionsRepo) CompleteSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.Session, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubListSessionsRepo) GetSessionForCorrection(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (s *stubListSessionsRepo) CreateCorrectedSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.CorrectionParams, _, _, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubListSessionsRepo) CorrectSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.Session, _ domain.CorrectionParams, _, _, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubListSessionsRepo) HasOverlappingSession(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) (bool, error) {
	return false, nil
}

func (s *stubListSessionsRepo) ListIncompleteSessionsForPeriod(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]domain.IncompleteSessionBlocker, error) {
	return nil, nil
}
