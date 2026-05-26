package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeListAttRepo struct {
	children []domain.AttendanceChild
	err      error
	captured struct {
		localDate time.Time
	}
}

func (f *fakeListAttRepo) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Child, error) {
	return nil, nil
}
func (f *fakeListAttRepo) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	return domain.Child{}, false, nil
}
func (f *fakeListAttRepo) Create(ctx context.Context, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	return nil
}
func (f *fakeListAttRepo) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	return 0, nil
}
func (f *fakeListAttRepo) MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, reasonCode, reasonNote string) error {
	return nil
}
func (f *fakeListAttRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	return domain.Child{}, false, nil
}
func (f *fakeListAttRepo) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeListAttRepo) ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]domain.AttendanceChild, error) {
	f.captured.localDate = localDate
	return f.children, f.err
}
func (f *fakeListAttRepo) GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildCorrectionInfo, bool, error) {
	return domain.ChildCorrectionInfo{}, false, nil
}

func TestListAttendance_DerivesLondonDateFromClock(t *testing.T) {
	// 2025-06-15 23:30 UTC = 2025-06-16 00:30 BST
	clockTime := time.Date(2025, 6, 15, 23, 30, 0, 0, time.UTC)
	repo := &fakeListAttRepo{}
	uc := NewListAttendance(repo, func() time.Time { return clockTime })

	_, err := uc.Execute(context.Background(), tenant.ActorContext{
		TenantID: uuid.New(), BranchID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	y, m, d := repo.captured.localDate.Date()
	if y != 2025 || m != 6 || d != 16 {
		t.Fatalf("localDate = %d-%02d-%02d, want 2025-06-16 (BST)", y, m, d)
	}
}

func TestListAttendance_UTCNoonMatchesLondonDate(t *testing.T) {
	// 2025-01-15 12:00 UTC = 2025-01-15 12:00 GMT (no DST)
	clockTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	repo := &fakeListAttRepo{}
	uc := NewListAttendance(repo, func() time.Time { return clockTime })

	_, err := uc.Execute(context.Background(), tenant.ActorContext{
		TenantID: uuid.New(), BranchID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	y, m, d := repo.captured.localDate.Date()
	if y != 2025 || m != 1 || d != 15 {
		t.Fatalf("localDate = %d-%02d-%02d, want 2025-01-15", y, m, d)
	}
}
