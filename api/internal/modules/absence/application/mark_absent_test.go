package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/absence/domain"
	"nursery-management-system/api/internal/modules/attendance/application"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeAbsenceRepo struct {
	marker        domain.AbsenceMarker
	activeMarker  *domain.AbsenceMarker
	found         bool
	hasAttendance bool
	clearOk       bool
}

func (f *fakeAbsenceRepo) Create(ctx context.Context, tx domain.Tx, marker domain.AbsenceMarker) (domain.AbsenceMarker, error) {
	return marker, nil
}

func (f *fakeAbsenceRepo) FindActiveByChildDate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (domain.AbsenceMarker, bool, error) {
	if f.activeMarker != nil {
		return *f.activeMarker, true, nil
	}
	return domain.AbsenceMarker{}, false, nil
}

func (f *fakeAbsenceRepo) GetByID(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (domain.AbsenceMarker, bool, error) {
	if f.found {
		return f.marker, true, nil
	}
	return domain.AbsenceMarker{}, false, nil
}

func (f *fakeAbsenceRepo) Clear(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID, clearedAt time.Time, clearedByUserID, clearedByMembershipID uuid.UUID) (domain.AbsenceMarker, bool, error) {
	if f.clearOk {
		f.marker.ClearedAt = &clearedAt
		return f.marker, true, nil
	}
	return domain.AbsenceMarker{}, false, nil
}

func (f *fakeAbsenceRepo) HasAttendanceForChildDate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (bool, error) {
	return f.hasAttendance, nil
}

type fakeAbsenceChildChecker struct {
	err error
}

func (f *fakeAbsenceChildChecker) CheckEnrollmentForAttendance(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error {
	return f.err
}

type fakeTxMgr struct{}

func (*fakeTxMgr) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

var _ txManager = (*fakeTxMgr)(nil)

type fakeAuditWriter struct{}

func (*fakeAuditWriter) WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error {
	return nil
}

var _ auditWriter = (*fakeAuditWriter)(nil)

func makeAbsenceActor() tenant.ActorContext {
	return tenant.ActorContext{
		UserID:       uuid.New(),
		MembershipID: uuid.New(),
		TenantID:     uuid.New(),
		BranchID:     uuid.New(),
		RequestID:    "test-request",
	}
}

func fixedAbsenceClock(t time.Time) *application.AttendanceClock {
	return application.NewAttendanceClock(func() time.Time { return t })
}

func newTestMarkAbsent(repo *fakeAbsenceRepo, checker *fakeAbsenceChildChecker) *MarkAbsent {
	return NewMarkAbsent(repo, checker, &fakeTxMgr{}, &fakeAuditWriter{}, fixedAbsenceClock(time.Now().UTC()))
}

func newTestClearMarker(repo *fakeAbsenceRepo) *ClearMarker {
	return NewClearMarker(repo, &fakeTxMgr{}, &fakeAuditWriter{}, fixedAbsenceClock(time.Now().UTC()))
}

func TestMarkAbsent_CreatesForEligibleChild(t *testing.T) {
	uc := newTestMarkAbsent(&fakeAbsenceRepo{}, &fakeAbsenceChildChecker{})
	result, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true")
	}
	if result.Marker.ChildID == uuid.Nil {
		t.Fatal("expected non-nil child ID")
	}
}

func TestMarkAbsent_IdempotentForExistingActiveMarker(t *testing.T) {
	existing := domain.AbsenceMarker{ID: uuid.New(), ChildID: uuid.New()}
	uc := newTestMarkAbsent(&fakeAbsenceRepo{activeMarker: &existing}, &fakeAbsenceChildChecker{})
	result, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Created {
		t.Fatal("expected Created=false for idempotent")
	}
	if result.Marker.ID != existing.ID {
		t.Fatal("expected existing marker returned")
	}
}

func TestMarkAbsent_RejectsInvalidChildID(t *testing.T) {
	uc := newTestMarkAbsent(&fakeAbsenceRepo{}, &fakeAbsenceChildChecker{})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), "not-a-uuid")
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %v", err)
	}
}

func TestMarkAbsent_RejectsChildNotFound(t *testing.T) {
	uc := newTestMarkAbsent(&fakeAbsenceRepo{}, &fakeAbsenceChildChecker{err: attendancedomain.ErrChildNotFound})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "child_not_found" {
		t.Fatalf("expected child_not_found, got %v", err)
	}
}

func TestMarkAbsent_RejectsEnrollmentIncomplete(t *testing.T) {
	uc := newTestMarkAbsent(&fakeAbsenceRepo{}, &fakeAbsenceChildChecker{err: attendancedomain.ErrChildEnrollmentIncomplete})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "child_enrollment_incomplete" {
		t.Fatalf("expected child_enrollment_incomplete, got %v", err)
	}
}

func TestMarkAbsent_RejectsWhenAttendanceExists(t *testing.T) {
	uc := newTestMarkAbsent(&fakeAbsenceRepo{hasAttendance: true}, &fakeAbsenceChildChecker{})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "absence_attendance_exists" {
		t.Fatalf("expected absence_attendance_exists, got %v", err)
	}
}

func TestClearMarker_ClearsActiveMarker(t *testing.T) {
	marker := domain.AbsenceMarker{ID: uuid.New(), ChildID: uuid.New()}
	uc := newTestClearMarker(&fakeAbsenceRepo{marker: marker, found: true, clearOk: true})
	result, err := uc.Execute(context.Background(), makeAbsenceActor(), marker.ID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ClearedAt == nil {
		t.Fatal("expected cleared_at to be set")
	}
}

func TestClearMarker_IdempotentWhenAlreadyCleared(t *testing.T) {
	cleared := time.Now().UTC()
	marker := domain.AbsenceMarker{ID: uuid.New(), ChildID: uuid.New(), ClearedAt: &cleared}
	uc := newTestClearMarker(&fakeAbsenceRepo{marker: marker, found: true})
	result, err := uc.Execute(context.Background(), makeAbsenceActor(), marker.ID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ClearedAt == nil {
		t.Fatal("expected cleared_at")
	}
}

func TestClearMarker_NotFoundForMissingMarker(t *testing.T) {
	uc := newTestClearMarker(&fakeAbsenceRepo{found: false})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "absence_marker_not_found" {
		t.Fatalf("expected absence_marker_not_found, got %v", err)
	}
}

func TestClearMarker_RejectsInvalidMarkerID(t *testing.T) {
	uc := newTestClearMarker(&fakeAbsenceRepo{})
	_, err := uc.Execute(context.Background(), makeAbsenceActor(), "not-a-uuid")
	de, ok := err.(*domainerrors.DomainError)
	if !ok || de.Code != "validation_error" {
		t.Fatalf("expected validation_error, got %v", err)
	}
}

func TestMarkAbsent_AllowsNewAfterClear(t *testing.T) {
	// No active marker (cleared), no attendance → should create new
	uc := newTestMarkAbsent(&fakeAbsenceRepo{}, &fakeAbsenceChildChecker{})
	result, err := uc.Execute(context.Background(), makeAbsenceActor(), uuid.New().String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true for new marker after clear")
	}
}
