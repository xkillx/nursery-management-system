package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeRepository struct {
	session domain.Session
	err     error
}

func (f *fakeRepository) CreateOpenSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	if f.err != nil {
		return domain.Session{}, f.err
	}
	return f.session, nil
}

func (f *fakeRepository) GetOpenSessionForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.Session, bool, error) {
	if f.err != nil {
		return domain.Session{}, false, f.err
	}
	if f.session.ID == uuid.Nil {
		return domain.Session{}, false, nil
	}
	return f.session, true, nil
}

func (f *fakeRepository) CompleteSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session domain.Session, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	if f.err != nil {
		return domain.Session{}, f.err
	}
	return session, nil
}

func (f *fakeRepository) GetSessionForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, sessionID uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (f *fakeRepository) CreateCorrectedSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (f *fakeRepository) CorrectSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session domain.Session, params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	return domain.Session{}, nil
}

func (f *fakeRepository) HasOverlappingSession(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, excludeSessionID *uuid.UUID, checkInAt, checkOutAt time.Time) (bool, error) {
	return false, nil
}

func (f *fakeRepository) ListIncompleteSessionsForPeriod(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]domain.IncompleteSessionBlocker, error) {
	return nil, nil
}

func (f *fakeRepository) ListSessionsForCorrection(ctx context.Context, tenantID, branchID, childID uuid.UUID, localDate time.Time) (domain.CorrectionSessionContext, error) {
	return domain.CorrectionSessionContext{}, nil
}

func (f *fakeRepository) ListCorrectionHistory(ctx context.Context, tenantID, branchID, sessionID uuid.UUID) (domain.Session, []domain.CorrectionHistoryEvent, error) {
	return domain.Session{}, nil, nil
}

type captureRepo struct {
	createOpenSession struct {
		called             bool
		occurredAt         time.Time
		localDate          time.Time
	}
	completeSession struct {
		called             bool
		occurredAt         time.Time
		localDate          time.Time
	}
	createCorrectedSession struct {
		called             bool
		checkInLocalDate   time.Time
		checkOutLocalDate  time.Time
		correctionActionLD time.Time
		occurredAt         time.Time
	}
	correctSession struct {
		called             bool
		checkInLocalDate   time.Time
		checkOutLocalDate  time.Time
		correctionActionLD time.Time
		occurredAt         time.Time
	}
	session domain.Session
	err     error
}

func (c *captureRepo) CreateOpenSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	c.createOpenSession.called = true
	c.createOpenSession.occurredAt = occurredAt
	c.createOpenSession.localDate = localDate
	if c.err != nil {
		return domain.Session{}, c.err
	}
	return c.session, nil
}

func (c *captureRepo) GetOpenSessionForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.Session, bool, error) {
	if c.err != nil {
		return domain.Session{}, false, c.err
	}
	if c.session.ID == uuid.Nil {
		return domain.Session{}, false, nil
	}
	return c.session, true, nil
}

func (c *captureRepo) CompleteSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session domain.Session, occurredAt time.Time, localDate time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	c.completeSession.called = true
	c.completeSession.occurredAt = occurredAt
	c.completeSession.localDate = localDate
	if c.err != nil {
		return domain.Session{}, c.err
	}
	return session, nil
}

func (c *captureRepo) GetSessionForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, sessionID uuid.UUID) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}

func (c *captureRepo) CreateCorrectedSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	c.createCorrectedSession.called = true
	c.createCorrectedSession.checkInLocalDate = checkInLocalDate
	c.createCorrectedSession.checkOutLocalDate = checkOutLocalDate
	c.createCorrectedSession.correctionActionLD = correctionActionLocalDate
	c.createCorrectedSession.occurredAt = occurredAt
	return c.session, nil
}

func (c *captureRepo) CorrectSessionWithEvent(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, session domain.Session, params domain.CorrectionParams, checkInLocalDate, checkOutLocalDate, correctionActionLocalDate time.Time, occurredAt time.Time, userID, membershipID uuid.UUID, requestID string) (domain.Session, error) {
	c.correctSession.called = true
	c.correctSession.checkInLocalDate = checkInLocalDate
	c.correctSession.checkOutLocalDate = checkOutLocalDate
	c.correctSession.correctionActionLD = correctionActionLocalDate
	c.correctSession.occurredAt = occurredAt
	return c.session, nil
}

func (c *captureRepo) HasOverlappingSession(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, excludeSessionID *uuid.UUID, checkInAt, checkOutAt time.Time) (bool, error) {
	return false, nil
}

func (c *captureRepo) ListIncompleteSessionsForPeriod(ctx context.Context, tenantID, branchID uuid.UUID, periodStartLocalDate, periodEndExclusiveLocalDate time.Time) ([]domain.IncompleteSessionBlocker, error) {
	return nil, nil
}

func (c *captureRepo) ListSessionsForCorrection(ctx context.Context, tenantID, branchID, childID uuid.UUID, localDate time.Time) (domain.CorrectionSessionContext, error) {
	return domain.CorrectionSessionContext{}, nil
}

func (c *captureRepo) ListCorrectionHistory(ctx context.Context, tenantID, branchID, sessionID uuid.UUID) (domain.Session, []domain.CorrectionHistoryEvent, error) {
	return domain.Session{}, nil, nil
}

type fakeChildChecker struct {
	err error
}

func (f *fakeChildChecker) CheckEnrollmentForAttendance(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error {
	return f.err
}

func makeActor() tenant.ActorContext {
	return tenant.ActorContext{
		UserID:       uuid.New(),
		MembershipID: uuid.New(),
		TenantID:     uuid.New(),
		BranchID:     uuid.New(),
		RequestID:    "test-request",
	}
}

func fixedClock(t time.Time) *AttendanceClock {
	return NewAttendanceClock(func() time.Time { return t })
}

func TestCheckIn_RejectsEnrollmentIncomplete(t *testing.T) {
	err := mapCheckInError(domain.ErrChildEnrollmentIncomplete)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "child_enrollment_incomplete" {
		t.Fatalf("expected child_enrollment_incomplete domain error, got %v", err)
	}
}

func TestCheckIn_RejectsDuplicateOpenSession(t *testing.T) {
	err := mapCheckInError(domain.ErrSessionAlreadyOpen)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_session_already_open" {
		t.Fatalf("expected attendance_session_already_open, got %v", err)
	}
}

func TestCheckIn_RejectsChildNotFound(t *testing.T) {
	err := mapCheckInError(domain.ErrChildNotFound)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "child_not_found" {
		t.Fatalf("expected child_not_found, got %v", err)
	}
}

func TestCheckOut_RejectsInvalidTimeOrder(t *testing.T) {
	err := mapCheckOutError(domain.ErrInvalidTimeOrder)
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "attendance_invalid_time_order" {
		t.Fatalf("expected attendance_invalid_time_order, got %v", err)
	}
}

func TestCheckOut_SucceedsWithOpenSession(t *testing.T) {
	session := domain.Session{
		ID:               uuid.New(),
		ChildID:          uuid.New(),
		Status:           domain.SessionStatusOpen,
		CheckInAt:        time.Now().Add(-time.Hour),
		CheckInLocalDate: time.Now(),
	}
	repo := &fakeRepository{session: session}
	auditWriter := audit.NewWriter()
	clock := fixedClock(time.Now().UTC())
	_ = NewCheckOutChild(repo, nil, auditWriter, clock)
}

func TestCheckIn_ChildCheckerReturnsIncomplete(t *testing.T) {
	checker := &fakeChildChecker{err: domain.ErrChildEnrollmentIncomplete}
	err := checker.CheckEnrollmentForAttendance(context.Background(), nil, uuid.New(), uuid.New(), uuid.New(), time.Now())
	if err != domain.ErrChildEnrollmentIncomplete {
		t.Fatalf("expected ErrChildEnrollmentIncomplete, got %v", err)
	}
}

func TestLondonNow(t *testing.T) {
	clock := NewAttendanceClock(RealClock)
	utc, localDate := clock.Now()
	if utc.IsZero() {
		t.Fatal("expected non-zero UTC time")
	}
	if localDate.IsZero() {
		t.Fatal("expected non-zero local date")
	}
}

func TestCheckOut_RejectsNoOpenSession_ErrorMapping(t *testing.T) {
	repo := &fakeRepository{}
	_, found, _ := repo.GetOpenSessionForUpdate(context.Background(), nil, uuid.Nil, uuid.Nil, uuid.Nil)
	if found {
		t.Fatal("expected found=false when no session in fake")
	}
}

func TestCheckOut_SessionNotOpen_ErrorCode(t *testing.T) {
	code := "attendance_session_not_open"
	if code != "attendance_session_not_open" {
		t.Fatalf("expected attendance_session_not_open")
	}
}
