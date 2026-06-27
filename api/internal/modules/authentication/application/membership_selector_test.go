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
		wantSelection  bool
		wantStale      bool
		wantMembership domain.Membership
	}{
		{
			name:        "zero memberships returns invalid credentials",
			memberships: nil,
			selectedID:  "",
			wantErr:     domain.ErrInvalidCredentials,
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
			name:          "one membership malformed selection",
			memberships:   []domain.Membership{m1},
			selectedID:    "not-a-uuid",
			wantSelection: false,
			wantErr:       &domain.ErrMalformedMembershipID,
		},
		{
			name:          "one membership wrong uuid",
			memberships:   []domain.Membership{m1},
			selectedID:    fixtureMembership2.String(),
			wantSelection: true,
			wantStale:     true,
		},
		{
			name:          "multiple memberships empty selection requires selection",
			memberships:   []domain.Membership{m1, m2},
			selectedID:    "",
			wantSelection: true,
			wantStale:     false,
		},
		{
			name:          "multiple memberships malformed selection",
			memberships:   []domain.Membership{m1, m2},
			selectedID:    "garbage",
			wantSelection: false,
			wantErr:       &domain.ErrMalformedMembershipID,
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
			name:          "multiple memberships uuid not in list",
			memberships:   []domain.Membership{m1, m2},
			selectedID:    fixtureMembership3.String(),
			wantSelection: true,
			wantStale:     true,
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

			if tt.wantSelection {
				if err == nil {
					t.Fatalf("expected MembershipSelectionRequiredError, got nil")
				}
				var selErr *domain.MembershipSelectionRequiredError
				if !isMembershipSelectionRequired(err) {
					t.Fatalf("expected MembershipSelectionRequiredError, got %T: %v", err, err)
				}
				selErr = err.(*domain.MembershipSelectionRequiredError)
				if selErr.IsStaleChoice != tt.wantStale {
					t.Fatalf("expected IsStaleChoice=%v, got %v", tt.wantStale, selErr.IsStaleChoice)
				}
				if len(selErr.Memberships) != len(tt.memberships) {
					t.Fatalf("expected %d choices, got %d", len(tt.memberships), len(selErr.Memberships))
				}
				return
			}

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

func isMembershipSelectionRequired(err error) bool {
	_, ok := err.(*domain.MembershipSelectionRequiredError)
	return ok
}
