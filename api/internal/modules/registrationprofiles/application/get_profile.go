package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetProfile struct {
	repo domain.Repository
}

func NewGetProfile(repo domain.Repository) *GetProfile {
	return &GetProfile{repo: repo}
}

func (uc *GetProfile) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (domain.ProfileWithChild, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return domain.ProfileWithChild{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.ProfileWithChild{}, domainerrors.Internal(fmt.Errorf("get child summary: %w", err))
	}
	if !found {
		return domain.ProfileWithChild{}, domainerrors.NotFound("child", "Child not found.")
	}

	profile, err := uc.repo.GetByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.ProfileWithChild{}, domainerrors.Internal(fmt.Errorf("get profile by child: %w", err))
	}

	result := domain.ProfileWithChild{
		Child:    child,
		Contacts: make([]domain.ContactEntry, 0),
	}

	if profile == nil {
		result.Profile = defaultProfile(childID, actor.TenantID, actor.BranchID)
		result.ProfileExists = false
	} else {
		result.Profile = profile
		result.ProfileExists = true
		contacts, err := uc.repo.ListContactsByProfile(ctx, actor.TenantID, actor.BranchID, profile.ID)
		if err != nil {
			return domain.ProfileWithChild{}, domainerrors.Internal(fmt.Errorf("list contacts: %w", err))
		}
		result.Contacts = contacts
	}

	return result, nil
}

func (uc *GetProfile) ExecuteGetChildSummary(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (domain.ChildSummary, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return domain.ChildSummary{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.repo.GetChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.ChildSummary{}, domainerrors.Internal(fmt.Errorf("get child summary: %w", err))
	}
	if !found {
		return domain.ChildSummary{}, domainerrors.NotFound("child", "Child not found.")
	}
	return child, nil
}

func defaultProfile(childID, tenantID, branchID uuid.UUID) *domain.Profile {
	return &domain.Profile{
		ID:                           uuid.Nil,
		TenantID:                     tenantID,
		BranchID:                     branchID,
		ChildID:                      childID,
		OtherLanguages:               []string{},
		HomeAddress:                  map[string]any{},
		DisabilityStatus:             domain.YesNoUnknownUnknown,
		MedicalConditionsStatus:      domain.YesNoUnknownUnknown,
		PrescribedMedicationStatus:   domain.YesNoUnknownUnknown,
		ImmunisationStatus:           domain.ImmunisationUnknown,
		DietaryRequirementsStatus:    domain.YesNoUnknownUnknown,
		SocialServicesStatus:         domain.YesNoUnknownUnknown,
		ConcernWalking:               domain.YesNoUnknownUnknown,
		ConcernSpeechLanguage:        domain.YesNoUnknownUnknown,
		ConcernHearing:               domain.YesNoUnknownUnknown,
		ConcernSight:                 domain.YesNoUnknownUnknown,
		ConcernEmotionalWellbeing:    domain.YesNoUnknownUnknown,
		ConcernBehaviour:             domain.YesNoUnknownUnknown,
		ProfessionalReferrals:        []domain.ProfessionalReferral{},
		BenefitsContributeToFees:     domain.YesNoUnknownUnknown,
		WorkingTaxCredit:             domain.YesNoUnknownUnknown,
		CollegeUniPaidToParent:       domain.YesNoUnknownUnknown,
		CollegeUniPaidToNursery:      domain.YesNoUnknownUnknown,
		Funding3yoTermTime:           domain.YesNoUnknownUnknown,
		Funding2yoTermTime:           domain.YesNoUnknownUnknown,
	}
}
