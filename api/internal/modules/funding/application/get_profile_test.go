package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type mockRepo struct {
	profile   domain.FundingProfile
	found     bool
	err       error
	enrollment domain.ChildEnrollment
	enrFound  bool
}

func (m *mockRepo) Get(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (domain.FundingProfile, bool, error) {
	return m.profile, m.found, m.err
}

func (m *mockRepo) GetForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (domain.FundingProfile, bool, error) {
	return m.profile, m.found, m.err
}

func (m *mockRepo) Create(ctx context.Context, tx domain.Tx, profile domain.FundingProfile) (domain.FundingProfile, error) {
	return profile, nil
}

func (m *mockRepo) UpdateAllowance(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time, minutes int) (domain.FundingProfile, error) {
	return domain.FundingProfile{FundedAllowanceMinutes: minutes}, nil
}

func (m *mockRepo) GetChildEnrollmentForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildEnrollment, bool, error) {
	return m.enrollment, m.enrFound, m.err
}

func (m *mockRepo) ListOverview(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]domain.OverviewRow, error) {
	return nil, nil
}

func TestParseBillingMonth(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Time
		wantErr bool
	}{
		{"2026-05", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), false},
		{"2026-01", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{"2026-12", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), false},
		{"invalid", time.Time{}, true},
		{"2026/05", time.Time{}, true},
		{"2026-5", time.Time{}, true},
		{"", time.Time{}, true},
	}

	for _, tt := range tests {
		got, err := ParseBillingMonth(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseBillingMonth(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseBillingMonth(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestGetProfile_InvalidChildID(t *testing.T) {
	uc := NewGetProfile(&mockRepo{})
	_, err := uc.Execute(context.Background(), tenant.ActorContext{}, "not-a-uuid", "2026-05")
	if err == nil {
		t.Fatal("expected error for invalid child ID")
	}
	if err.Error()[:16] != "validation_error" {
		t.Errorf("error = %v, want validation_error", err)
	}
}

func TestGetProfile_InvalidBillingMonth(t *testing.T) {
	uc := NewGetProfile(&mockRepo{})
	_, err := uc.Execute(context.Background(), tenant.ActorContext{}, uuid.New().String(), "invalid")
	if err == nil {
		t.Fatal("expected error for invalid billing month")
	}
	if err.Error()[:16] != "validation_error" {
		t.Errorf("error = %v, want validation_error", err)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	uc := NewGetProfile(&mockRepo{found: false})
	_, err := uc.Execute(context.Background(), tenant.ActorContext{}, uuid.New().String(), "2026-05")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if err.Error()[:23] != "funding_profile_not_fou" {
		t.Errorf("error = %v, want funding_profile_not_found", err)
	}
}

func TestGetProfile_Success(t *testing.T) {
	id := uuid.New()
	uc := NewGetProfile(&mockRepo{
		found:   true,
		profile: domain.FundingProfile{ID: id, FundedAllowanceMinutes: 570},
	})
	got, err := uc.Execute(context.Background(), tenant.ActorContext{}, uuid.New().String(), "2026-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
}

func TestValidateMonthOverlap(t *testing.T) {
	tests := []struct {
		name      string
		month     time.Time
		enrollment domain.ChildEnrollment
		want      bool
	}{
		{
			"partial start overlap",
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			domain.ChildEnrollment{StartDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
			true,
		},
		{
			"partial end overlap",
			time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			domain.ChildEnrollment{
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   ptrTime(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)),
			},
			true,
		},
		{
			"fully before start",
			time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			domain.ChildEnrollment{StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			false,
		},
		{
			"fully after end",
			time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			domain.ChildEnrollment{
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   ptrTime(time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)),
			},
			false,
		},
		{
			"nil end date always valid",
			time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
			domain.ChildEnrollment{StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateMonthOverlap(tt.month, tt.enrollment)
			if got != tt.want {
				t.Errorf("validateMonthOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
