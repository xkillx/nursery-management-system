package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/domain"
)

type mockClosureRepo struct {
	closures      []domain.BranchClosureDay
	dateExists    bool
	createErr     error
	deleteErr     error
	listErr       error
	dateExistsErr error
}

func (m *mockClosureRepo) Create(ctx context.Context, closure domain.BranchClosureDay) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.closures = append(m.closures, closure)
	return nil
}

func (m *mockClosureRepo) ListByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) ([]domain.BranchClosureDay, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var out []domain.BranchClosureDay
	for _, c := range m.closures {
		if !c.Date.Before(from) && !c.Date.After(to) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *mockClosureRepo) ListByBranchAndDateRangePaginated(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time, limit, offset int) ([]domain.BranchClosureDay, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var out []domain.BranchClosureDay
	for _, c := range m.closures {
		if !c.Date.Before(from) && !c.Date.After(to) {
			out = append(out, c)
		}
	}
	if offset >= len(out) {
		return nil, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func (m *mockClosureRepo) CountByBranchAndDateRange(ctx context.Context, tenantID, branchID uuid.UUID, from, to time.Time) (int, error) {
	if m.listErr != nil {
		return 0, m.listErr
	}
	count := 0
	for _, c := range m.closures {
		if !c.Date.Before(from) && !c.Date.After(to) {
			count++
		}
	}
	return count, nil
}

func (m *mockClosureRepo) Delete(ctx context.Context, tenantID, branchID, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, c := range m.closures {
		if c.ID == id {
			m.closures = append(m.closures[:i], m.closures[i+1:]...)
			return nil
		}
	}
	return domain.ErrClosureDayNotFound
}

func (m *mockClosureRepo) DateExists(ctx context.Context, tenantID, branchID uuid.UUID, date time.Time) (bool, error) {
	if m.dateExistsErr != nil {
		return false, m.dateExistsErr
	}
	return m.dateExists, nil
}

func (m *mockClosureRepo) ListClosureDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]time.Time, error) {
	return nil, nil
}

func TestCreateClosureDay_HappyPath(t *testing.T) {
	repo := &mockClosureRepo{}
	uc := NewCreateClosureDay(repo)
	tenantID := uuid.New()
	branchID := uuid.New()
	date := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)

	closure, err := uc.Execute(context.Background(), tenantID, branchID, CreateClosureDayParams{Date: date})
	if err != nil {
		t.Fatal(err)
	}
	if closure.TenantID != tenantID {
		t.Errorf("tenant ID mismatch")
	}
	if closure.BranchID != branchID {
		t.Errorf("branch ID mismatch")
	}
	if closure.Date != date {
		t.Errorf("date mismatch")
	}
}

func TestCreateClosureDay_WithReason(t *testing.T) {
	repo := &mockClosureRepo{}
	uc := NewCreateClosureDay(repo)
	tenantID := uuid.New()
	branchID := uuid.New()
	date := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	reason := "Inset day"

	closure, err := uc.Execute(context.Background(), tenantID, branchID, CreateClosureDayParams{Date: date, Reason: &reason})
	if err != nil {
		t.Fatal(err)
	}
	if closure.Reason == nil || *closure.Reason != reason {
		t.Errorf("reason mismatch")
	}
}

func TestCreateClosureDay_EmptyDate(t *testing.T) {
	repo := &mockClosureRepo{}
	uc := NewCreateClosureDay(repo)
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), CreateClosureDayParams{})
	if err == nil {
		t.Error("expected error for zero date")
	}
}

func TestCreateClosureDay_DuplicateDate(t *testing.T) {
	repo := &mockClosureRepo{dateExists: true}
	uc := NewCreateClosureDay(repo)
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), CreateClosureDayParams{
		Date: time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Error("expected error for duplicate date")
	}
}

func TestListClosureDays_HappyPath(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	repo := &mockClosureRepo{
		closures: []domain.BranchClosureDay{
			{ID: uuid.New(), TenantID: tenantID, BranchID: branchID, Date: time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New(), TenantID: tenantID, BranchID: branchID, Date: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New(), TenantID: tenantID, BranchID: branchID, Date: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	uc := NewListClosureDays(repo)

	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	closures, err := uc.Execute(context.Background(), tenantID, branchID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(closures) != 2 {
		t.Errorf("expected 2, got %d", len(closures))
	}
}

func TestListClosureDays_InvalidDateRange(t *testing.T) {
	repo := &mockClosureRepo{}
	uc := NewListClosureDays(repo)
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), time.Time{}, time.Time{})
	if err == nil {
		t.Error("expected error for zero dates")
	}
}

func TestDeleteClosureDay_HappyPath(t *testing.T) {
	id := uuid.New()
	repo := &mockClosureRepo{
		closures: []domain.BranchClosureDay{
			{ID: id, TenantID: uuid.New(), BranchID: uuid.New(), Date: time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)},
		},
	}
	uc := NewDeleteClosureDay(repo)
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteClosureDay_NotFound(t *testing.T) {
	repo := &mockClosureRepo{}
	uc := NewDeleteClosureDay(repo)
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}
