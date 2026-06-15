package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetWorkflowStatus struct {
	profileRepo domain.Repository
	consentRepo domain.ConsentRepository
	attestRepo  domain.AttestationRepository
}

func NewGetWorkflowStatus(profileRepo domain.Repository, consentRepo domain.ConsentRepository, attestRepo domain.AttestationRepository) *GetWorkflowStatus {
	return &GetWorkflowStatus{
		profileRepo: profileRepo,
		consentRepo: consentRepo,
		attestRepo:  attestRepo,
	}
}

func (uc *GetWorkflowStatus) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (domain.WorkflowStatus, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return domain.WorkflowStatus{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	child, found, err := uc.profileRepo.GetChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.WorkflowStatus{}, domainerrors.Internal(err)
	}
	if !found {
		return domain.WorkflowStatus{}, domainerrors.NotFound("child", "Child not found.")
	}

	profile, err := uc.profileRepo.GetByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.WorkflowStatus{}, domainerrors.Internal(err)
	}

	var profileCompleteness domain.Completeness
	var profileUpdatedAt *time.Time

	if profile != nil {
		contacts, listErr := uc.profileRepo.ListContactsByProfile(ctx, actor.TenantID, actor.BranchID, profile.ID)
		if listErr != nil {
			return domain.WorkflowStatus{}, domainerrors.Internal(listErr)
		}
		passwordIsSet := profile.CollectionPasswordHash != nil
		profileCompleteness = domain.ComputeCompleteness(profile, contacts, passwordIsSet)
		profileUpdatedAt = &profile.UpdatedAt
	} else {
		profileCompleteness = domain.Completeness{
			IsComplete:      false,
			MissingSections: []domain.SectionCode{domain.SectionDemographicsHome},
			Sections:        []domain.CompletenessSection{},
		}
	}

	currentConsent, err := uc.consentRepo.GetLatestByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.WorkflowStatus{}, domainerrors.Internal(err)
	}

	consentCompleteness := domain.ComputeConsentCompleteness(currentConsent)

	latestAttestation, err := uc.attestRepo.GetLatestAttestationByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.WorkflowStatus{}, domainerrors.Internal(err)
	}

	return domain.ComputeWorkflowStatus(child, profileCompleteness, consentCompleteness, currentConsent, latestAttestation, profileUpdatedAt), nil
}
