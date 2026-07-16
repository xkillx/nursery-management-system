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

func (m *mockOverviewRepo) Get(_ context.Context, _, _, _ uuid.UUID, _ time.Time) (domain.FundingProfile, bool, error) {
	return domain.FundingProfile{}, false, nil
}
func (m *mockOverviewRepo) GetForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.FundingProfile, bool, error) {
	return domain.FundingProfile{}, false, nil
}
func (m *mockOverviewRepo) Create(_ context.Context, _ domain.Tx, _ domain.FundingProfile) (domain.FundingProfile, error) {
	return domain.FundingProfile{}, nil
}
func (m *mockOverviewRepo) UpdateAllowance(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time, _ int) (domain.FundingProfile, error) {
	return domain.FundingProfile{}, nil
}
func (m *mockOverviewRepo) GetChildEnrollmentForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.ChildEnrollment, bool, error) {
	return domain.ChildEnrollment{}, false, nil
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

func testActor() tenant.ActorContext {
	return tenant.ActorContext{
		TenantID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		BranchID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	}
}

func ptrInt(v int) *int                 { return &v }
func ptrTimeVal(v time.Time) *time.Time { return &v }
func ptrUUID(v uuid.UUID) *uuid.UUID    { return &v }

func makeRow(name string, profileID *uuid.UUID, allowance *int) domain.OverviewRow {
	return domain.OverviewRow{
		ChildID:                uuid.New(),
		ChildFirstName:         name,
		IsActive:               true,
		StartDate:              time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:                nil,
		FundingProfileID:       profileID,
		FundedAllowanceMinutes: allowance,
	}
}

func TestListOverview_InvalidMonth(t *testing.T) {
	uc := NewListOverview(&mockOverviewRepo{}, &mockConsumedMinutesProvider{})
	_, err := uc.Execute(context.Background(), testActor(), "bad")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestListOverview_MissingProfile(t *testing.T) {
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Alice", nil, nil)}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
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
	pid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Bob", &pid, ptrInt(0))}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.ExplicitZeroCount != 1 {
		t.Fatalf("zero = %d, want 1", result.Summary.ExplicitZeroCount)
	}
	if result.Items[0].Flags[0] != domain.FlagExplicitZero {
		t.Fatalf("flag = %v, want explicit_zero_allowance", result.Items[0].Flags)
	}
}

func TestListOverview_UnderOneHour(t *testing.T) {
	pid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Cara", &pid, ptrInt(30))}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.UnderOneHourCount != 1 {
		t.Fatalf("under1h = %d, want 1", result.Summary.UnderOneHourCount)
	}
}

func TestListOverview_ExactlySixtyNotFlagged(t *testing.T) {
	pid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Dana", &pid, ptrInt(60))}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.FlaggedChildCount != 0 {
		t.Fatalf("60 min should not be flagged, got %d flagged", result.Summary.FlaggedChildCount)
	}
}

func TestListOverview_Above160Hours(t *testing.T) {
	pid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Eve", &pid, ptrInt(9601))}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.Above160HoursCount != 1 {
		t.Fatalf("above160 = %d, want 1", result.Summary.Above160HoursCount)
	}
}

func TestListOverview_Exactly9600NotFlagged(t *testing.T) {
	pid := uuid.New()
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{makeRow("Frank", &pid, ptrInt(9600))}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.FlaggedChildCount != 0 {
		t.Fatalf("9600 min (160h) should not be flagged, got %d flagged", result.Summary.FlaggedChildCount)
	}
}

func TestListOverview_EmptyResult(t *testing.T) {
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if result.Summary.IncludedChildCount != 0 {
		t.Fatalf("included = %d, want 0", result.Summary.IncludedChildCount)
	}
	if result.Items == nil {
		t.Fatal("items should be empty slice, not nil")
	}
}

func TestListOverview_MultipleFlagsOnOneChild(t *testing.T) {
	pid := uuid.New()
	// Zero allowance but also not above 160h — zero only
	repo := &mockOverviewRepo{rows: []domain.OverviewRow{
		makeRow("Zero", &pid, ptrInt(0)),
	}}
	uc := NewListOverview(repo, &mockConsumedMinutesProvider{})
	result, _ := uc.Execute(context.Background(), testActor(), "2026-06")
	if len(result.Items[0].Flags) != 1 {
		t.Fatalf("zero should have 1 flag (explicit_zero), got %d", len(result.Items[0].Flags))
	}
}
