package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func TestListCorrectionHistory_ReturnsSessionAndEvents(t *testing.T) {
	sessionID := uuid.New()
	actor := makeActor()

	session := domain.Session{ID: sessionID, Status: domain.SessionStatusCorrected}
	events := []domain.CorrectionHistoryEvent{
		{ID: uuid.New(), EventType: domain.EventCheckIn},
		{ID: uuid.New(), EventType: domain.EventCorrection},
	}

	repo := &stubHistoryRepo{session: session, events: events}
	uc := NewListCorrectionHistory(repo)

	got, err := uc.Execute(context.Background(), actor, sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Session.ID != sessionID {
		t.Fatalf("expected session id %s, got %s", sessionID, got.Session.ID)
	}
	if len(got.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got.Events))
	}
}

func TestListCorrectionHistory_MapsNotFound(t *testing.T) {
	sessionID := uuid.New()
	actor := makeActor()

	repo := &stubHistoryRepo{err: domain.ErrSessionNotFound}
	uc := NewListCorrectionHistory(repo)

	_, err := uc.Execute(context.Background(), actor, sessionID)
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_session_not_found" {
		t.Fatalf("expected attendance_session_not_found, got %v", err)
	}
}

type stubHistoryRepo struct {
	session domain.Session
	events  []domain.CorrectionHistoryEvent
	err     error
}

func (s *stubHistoryRepo) ListCorrectionHistory(_ context.Context, _, _, _ uuid.UUID) (domain.Session, []domain.CorrectionHistoryEvent, error) {
	return s.session, s.events, s.err
}

func (s *stubHistoryRepo) ListSessionsForCorrection(_ context.Context, _, _, _ uuid.UUID, _ time.Time) (domain.CorrectionSessionContext, error) {
	return domain.CorrectionSessionContext{}, nil
}

func (s *stubHistoryRepo) CreateOpenSessionWithEvent(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubHistoryRepo) GetOpenSessionForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (s *stubHistoryRepo) CompleteSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.Session, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubHistoryRepo) GetSessionForCorrection(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (s *stubHistoryRepo) CreateCorrectedSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.CorrectionParams, _, _, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubHistoryRepo) CorrectSessionWithEvent(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ domain.Session, _ domain.CorrectionParams, _, _, _ time.Time, _ time.Time, _, _ uuid.UUID, _ string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s *stubHistoryRepo) HasOverlappingSession(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ *uuid.UUID, _, _ time.Time) (bool, error) {
	return false, nil
}

func (s *stubHistoryRepo) ListIncompleteSessionsForPeriod(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]domain.IncompleteSessionBlocker, error) {
	return nil, nil
}

func (s *stubHistoryRepo) GetRegister(_ context.Context, _, _ uuid.UUID, _ time.Time, _ []int32) ([]domain.RegisterEntry, error) {
	return nil, nil
}

func (s *stubHistoryRepo) GetRegisterSummary(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]domain.RegisterSummaryEntry, error) {
	return nil, nil
}
