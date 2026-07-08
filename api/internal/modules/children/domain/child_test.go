package domain

import (
	"testing"
	"time"
)

func strPtr(s string) *string { return &s }

func TestActivate(t *testing.T) {
	tests := []struct {
		name          string
		child         Child
		startDate     time.Time
		hourlyRate    int
		wantErr       bool
		wantActive    bool
		wantRate      *int
		wantStartDate time.Time
	}{
		{
			name: "valid future date",
			child: Child{
				IsActive:  false,
				StartDate: time.Time{},
				EndDate:   timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			startDate:     time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
			hourlyRate:    500,
			wantErr:       false,
			wantActive:    true,
			wantRate:      intPtr(500),
			wantStartDate: time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "past date returns error",
			child: Child{
				IsActive: false,
			},
			startDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			hourlyRate: 500,
			wantErr:    true,
		},
		{
			name: "already active returns error",
			child: Child{
				IsActive: true,
			},
			startDate:  time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC),
			hourlyRate: 500,
			wantErr:    true,
			wantActive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.child.Activate(tc.startDate, tc.hourlyRate)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.child.IsActive != tc.wantActive {
				t.Errorf("IsActive = %v, want %v", tc.child.IsActive, tc.wantActive)
			}
			if tc.child.StartDate != tc.wantStartDate {
				t.Errorf("StartDate = %v, want %v", tc.child.StartDate, tc.wantStartDate)
			}
			if tc.wantRate != nil {
				if tc.child.SiteCoreHourlyRateMinor == nil {
					t.Fatal("SiteCoreHourlyRateMinor is nil")
				}
				if *tc.child.SiteCoreHourlyRateMinor != *tc.wantRate {
					t.Errorf("SiteCoreHourlyRateMinor = %d, want %d", *tc.child.SiteCoreHourlyRateMinor, *tc.wantRate)
				}
			}
		})
	}
}

func TestDeactivate(t *testing.T) {
	tests := []struct {
		name        string
		child       Child
		reasonCode  ReasonCode
		deactivated time.Time
		wantErr     bool
		wantActive  bool
		wantEndDate *time.Time
	}{
		{
			name: "valid reason code",
			child: Child{
				IsActive: true,
			},
			reasonCode:  ReasonLeftNursery,
			deactivated: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			wantActive:  false,
			wantEndDate: timePtr(time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "invalid reason code",
			child: Child{
				IsActive: true,
			},
			reasonCode: "invalid_code",
			wantErr:    true,
			wantActive: true,
		},
		{
			name: "already inactive",
			child: Child{
				IsActive: false,
			},
			reasonCode: ReasonLeftNursery,
			wantErr:    true,
			wantActive: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.child.Deactivate(tc.reasonCode, tc.deactivated)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.child.IsActive != tc.wantActive {
				t.Errorf("IsActive = %v, want %v", tc.child.IsActive, tc.wantActive)
			}
			if tc.wantEndDate != nil {
				if tc.child.EndDate == nil {
					t.Fatal("EndDate is nil")
				}
				if !tc.child.EndDate.Equal(*tc.wantEndDate) {
					t.Errorf("EndDate = %v, want %v", tc.child.EndDate, tc.wantEndDate)
				}
			}
		})
	}
}

func TestChangeName(t *testing.T) {
	tests := []struct {
		name      string
		child     Child
		firstName string
		lastName  *string
		wantErr   bool
	}{
		{
			name:      "set first and last name",
			child:     Child{},
			firstName: "Jane",
			lastName:  strPtr("Doe"),
			wantErr:   false,
		},
		{
			name:      "set first name only",
			child:     Child{},
			firstName: "Jane",
			lastName:  nil,
			wantErr:   false,
		},
		{
			name:      "empty first name",
			child:     Child{},
			firstName: "",
			lastName:  strPtr("Doe"),
			wantErr:   true,
		},
		{
			name: "no change returns error",
			child: Child{
				FirstName: "Jane",
				LastName:  strPtr("Doe"),
			},
			firstName: "Jane",
			lastName:  strPtr("Doe"),
			wantErr:   true,
		},
		{
			name: "no change with nil last name",
			child: Child{
				FirstName: "Jane",
				LastName:  nil,
			},
			firstName: "Jane",
			lastName:  nil,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.child.ChangeName(tc.firstName, tc.lastName)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.child.FirstName != tc.firstName {
				t.Errorf("FirstName = %q, want %q", tc.child.FirstName, tc.firstName)
			}
			if !stringPtrEqual(tc.child.LastName, tc.lastName) {
				t.Errorf("LastName = %v, want %v", tc.child.LastName, tc.lastName)
			}
		})
	}
}

func TestIsEligibleForAttendance(t *testing.T) {
	tests := []struct {
		name      string
		child     Child
		localDate time.Time
		want      bool
	}{
		{
			name: "active, after start, no end date",
			child: Child{
				IsActive:  true,
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   nil,
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name: "inactive",
			child: Child{
				IsActive:  false,
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name: "before start date",
			child: Child{
				IsActive:  true,
				StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name: "on start date is eligible",
			child: Child{
				IsActive:  true,
				StartDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
				EndDate:   nil,
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name: "after end date",
			child: Child{
				IsActive:  true,
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name: "on end date is not eligible",
			child: Child{
				IsActive:  true,
				StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   timePtr(time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)),
			},
			localDate: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
			want:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.child.IsEligibleForAttendance(tc.localDate)
			if got != tc.want {
				t.Errorf("IsEligibleForAttendance(%v) = %v, want %v", tc.localDate, got, tc.want)
			}
		})
	}
}

func TestNewReasonCode(t *testing.T) {
	t.Run("valid code", func(t *testing.T) {
		code, err := NewReasonCode("left_nursery")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != ReasonLeftNursery {
			t.Errorf("got %q, want %q", code, ReasonLeftNursery)
		}
	})

	t.Run("all valid codes", func(t *testing.T) {
		for valid := range ValidReasonCodes {
			_, err := NewReasonCode(string(valid))
			if err != nil {
				t.Errorf("NewReasonCode(%q) unexpected error: %v", valid, err)
			}
		}
	})

	t.Run("empty string returns error", func(t *testing.T) {
		_, err := NewReasonCode("")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("invalid code returns error", func(t *testing.T) {
		_, err := NewReasonCode("invalid_code")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func intPtr(i int) *int { return &i }

func timePtr(t time.Time) *time.Time { return &t }
