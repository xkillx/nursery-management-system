package application

import (
	"testing"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

func TestSelectLoginMembership(t *testing.T) {
	m1 := makeMembership(fixtureMembership1, fixtureTenantID, fixtureBranchID, "manager")
	m2 := makeMembership(fixtureMembership2, fixtureTenantID, fixtureBranchID, "practitioner")

	tests := []struct {
		name           string
		memberships    []domain.Membership
		selectedID     string
		wantErr        error
		wantMembership domain.Membership
	}{
		{
			name:        "zero memberships returns error",
			memberships: nil,
			selectedID:  "",
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:           "one membership empty selection auto selects",
			memberships:    []domain.Membership{m1},
			selectedID:     "",
			wantMembership: m1,
		},
		{
			name:           "one membership matching selection",
			memberships:    []domain.Membership{m1},
			selectedID:     fixtureMembership1.String(),
			wantMembership: m1,
		},
		{
			name:        "one membership malformed selection",
			memberships: []domain.Membership{m1},
			selectedID:  "not-a-uuid",
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:        "one membership wrong uuid",
			memberships: []domain.Membership{m1},
			selectedID:  fixtureMembership2.String(),
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:        "multiple memberships empty selection requires explicit",
			memberships: []domain.Membership{m1, m2},
			selectedID:  "",
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:        "multiple memberships malformed selection",
			memberships: []domain.Membership{m1, m2},
			selectedID:  "garbage",
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:           "multiple memberships valid selection first",
			memberships:    []domain.Membership{m1, m2},
			selectedID:     fixtureMembership1.String(),
			wantMembership: m1,
		},
		{
			name:           "multiple memberships valid selection second",
			memberships:    []domain.Membership{m1, m2},
			selectedID:     fixtureMembership2.String(),
			wantMembership: m2,
		},
		{
			name:        "multiple memberships uuid not in list",
			memberships: []domain.Membership{m1, m2},
			selectedID:  fixtureMembership3.String(),
			wantErr:     domain.ErrInvalidMembership,
		},
		{
			name:           "whitespace trimmed around membership_id",
			memberships:    []domain.Membership{m1},
			selectedID:     " " + fixtureMembership1.String() + " ",
			wantMembership: m1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectLoginMembership(tt.memberships, tt.selectedID)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.wantMembership.ID {
				t.Fatalf("expected membership %s, got %s", tt.wantMembership.ID, got.ID)
			}
		})
	}
}
