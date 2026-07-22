package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type mockOverviewRepo struct {
	rows []domain.OverviewRow
	err  error
}

type mockConsumedMinutesProvider struct {
	consumed map[uuid.UUID]int
}

func (m *mockConsumedMinutesProvider) GetConsumedMinutes(_ context.Context, _, _ uuid.UUID, _ []uuid.UUID, _ time.Time) (map[uuid.UUID]int, error) {
	if m.consumed == nil {
		return map[uuid.UUID]int{}, nil
	}
	return m.consumed, nil
}

type mockTermDateProvider struct {
	ranges []domain.TermDateRange
}

func (m *mockTermDateProvider) GetTermDatesForBranchAndMonth(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.TermDateRange, error) {
	if m.ranges == nil {
		return []domain.TermDateRange{}, nil
	}
	return m.ranges, nil
}

func (m *mockOverviewRepo) ListOverview(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.OverviewRow, error) {
	return m.rows, m.err
}

func (m *mockOverviewRepo) ListOverviewPaginated(_ context.Context, _, _ uuid.UUID, _ time.Time, limit, offset int) ([]domain.OverviewRow, error) {
	if offset >= len(m.rows) {
		return nil, m.err
	}
	end := offset + limit
	if end > len(m.rows) {
		end = len(m.rows)
	}
	return m.rows[offset:end], m.err
}

func (m *mockOverviewRepo) CountOverview(_ context.Context, _, _ uuid.UUID, _ time.Time) (int, error) {
	return len(m.rows), m.err
}

func (m *mockOverviewRepo) ListExpiringSoon(_ context.Context, _, _ uuid.UUID, _ int) ([]domain.ExpiringFundingRecord, error) {
	return nil, nil
}

func (m *mockOverviewRepo) GetFundedChildrenCount(_ context.Context, _, _ uuid.UUID, _ time.Time) (domain.EnhancedOverviewMetrics, error) {
	return domain.EnhancedOverviewMetrics{}, nil
}

func (m *mockOverviewRepo) GetBookedHoursThisWeek(_ context.Context, _, _ uuid.UUID) (float64, error) {
	return 0, nil
}

func (m *mockOverviewRepo) GetExpiringSoonCount(_ context.Context, _, _ uuid.UUID, _ int) (int, error) {
	return 0, nil
}

func (m *mockOverviewRepo) GetChildAllocation(_ context.Context, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AllocationEntry, error) {
	return nil, nil
}

func testActor() tenant.ActorContext {
	return tenant.ActorContext{
		TenantID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		BranchID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	}
}

func ptrInt(v int) *int                 { return &v }
func ptrTimeVal(v time.Time) *time.Time { return &v }
func ptrUUID(v uuid.UUID) *uuid.UUID    { return &v }
func ptrFloat64(v float64) *float64     { return &v }
func ptrString(v string) *string        { return &v }

func makeRow(name string, recordID *uuid.UUID, enabled bool, hoursPerWeek *float64, model *string) domain.OverviewRow {
	return domain.OverviewRow{
		ChildID:            uuid.New(),
		ChildFirstName:     name,
		IsActive:           true,
		StartDate:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:            nil,
		FundingRecordID:    recordID,
		FundingEnabled:     enabled,
		FundedHoursPerWeek: hoursPerWeek,
		FundingModel:       model,
	}
}

func newTestListOverview(repo *mockOverviewRepo) *ListOverview {
	return NewListOverview(repo, &mockConsumedMinutesProvider{}, &mockTermDateProvider{
		ranges: []domain.TermDateRange{
			{StartDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)},
		},
	})
}

func TestListOverview_InvalidMonth(t *testing.T) {
	uc := newTestListOverview(&mockOverviewRepo{})
	_, err := uc.Execute(context.Background(), testActor(), "bad")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestListOverview_MissingProfile(t *testing.T) {
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Alice", nil, false, nil, nil)}}
	uc := newTestListOverview(repo)
	result, err := uc.Execute(context.Background(), testActor(), "2026-06")
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.IncludedChildCount != 1 {
		t.Fatalf("included = %d, want 1", result.Summary.IncludedChildCount)
	}
	if result.Summary.FlaggedChildCount != 1 {
		t.Fatalf("flagged = %d, want 1", result.Summary.FlaggedChildCount)
	}
	if result.Summary.MissingProfileCount != 1 {
		t.Fatalf("missing = %d, want 1", result.Summary.MissingProfileCount)
	}
	if len(result.Items) != 1 || result.Items[0].Flags[0] != domain.FlagMissingProfile {
		t.Fatalf("flags = %v, want [missing_profile]", result.Items)
	}
}

func TestListOverview_ExplicitZero(t *testing.T) {
	rid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Bob", &rid, true, ptrFloat64(0), ptrString("term_time_only"))}}
	uc := newTestListOverview(repo)
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.ExplicitZeroCount != 1 {
		t.Fatalf("zero = %d, want 1", result.Summary.ExplicitZeroCount)
	}
	if result.Items[0].Flags[0] != domain.FlagExplicitZero {
		t.Fatalf("flag = %v, want explicit_zero_allowance", result.Items[0].Flags)
	}
}

func TestListOverview_EmptyResult(t *testing.T) {
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{}}
	uc := newTestListOverview(repo)
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.IncludedChildCount != 0 {
		t.Fatalf("included = %d, want 0", result.Summary.IncludedChildCount)
	}
	if result.Items == nil {
		t.Fatal("items should be empty slice, not nil")
	}
}
