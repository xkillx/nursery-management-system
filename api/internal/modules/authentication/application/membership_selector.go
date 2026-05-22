package application

import (
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

// SelectLoginMembership chooses the active membership for a login request.
// If selectedMembershipID is empty and the user has exactly one membership, it is used automatically.
// If the user has multiple memberships, an explicit selection is required.
func SelectLoginMembership(memberships []domain.Membership, selectedMembershipID string) (domain.Membership, error) {
	if len(memberships) == 0 {
		return domain.Membership{}, domain.ErrInvalidMembership
	}

	selectedMembershipID = strings.TrimSpace(selectedMembershipID)

	if len(memberships) == 1 {
		only := memberships[0]
		if selectedMembershipID == "" {
			return only, nil
		}
		selectedID, err := uuid.Parse(selectedMembershipID)
		if err != nil {
			return domain.Membership{}, domain.ErrInvalidMembership
		}
		if selectedID != only.ID {
			return domain.Membership{}, domain.ErrInvalidMembership
		}
		return only, nil
	}

	if selectedMembershipID == "" {
		return domain.Membership{}, domain.ErrInvalidMembership
	}

	return SelectExplicitMembership(memberships, selectedMembershipID)
}

// SelectExplicitMembership finds a membership by ID from the user's membership list.
func SelectExplicitMembership(memberships []domain.Membership, selectedMembershipID string) (domain.Membership, error) {
	selectedID, err := uuid.Parse(strings.TrimSpace(selectedMembershipID))
	if err != nil {
		return domain.Membership{}, domain.ErrInvalidMembership
	}

	for _, m := range memberships {
		if m.ID == selectedID {
			return m, nil
		}
	}

	return domain.Membership{}, domain.ErrInvalidMembership
}

func containsMembership(memberships []domain.Membership, membershipID uuid.UUID) bool {
	for _, m := range memberships {
		if m.ID == membershipID {
			return true
		}
	}
	return false
}
