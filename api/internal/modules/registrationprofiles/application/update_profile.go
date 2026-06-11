package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type UpdateProfile struct {
	repo        domain.Repository
	auditWriter *audit.Writer
	txManager   *transaction.Manager
}

func NewUpdateProfile(repo domain.Repository, auditWriter *audit.Writer, txManager *transaction.Manager) *UpdateProfile {
	return &UpdateProfile{repo: repo, auditWriter: auditWriter, txManager: txManager}
}

type UpdateProfileResult struct {
	ProfileWithChild domain.ProfileWithChild
	Completeness     domain.Completeness
	Created          bool
}

func (uc *UpdateProfile) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string, patch PatchSection) (*UpdateProfileResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	if !hasAnySection(patch) {
		return nil, domainerrors.Validation("Invalid request payload.", "patch")
	}

	var result *UpdateProfileResult

	err = uc.txManager.ExecTx(ctx, func(tx domain.Tx) error {
		child, found, err := uc.repo.GetChildSummary(ctx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get child summary: %w", err))
		}
		if !found {
			return domainerrors.NotFound("child", "Child not found.")
		}

		existingProfile, err := uc.repo.GetForUpdateByChild(ctx, tx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get profile for update: %w", err))
		}

		var profile *domain.Profile
		created := false

		if existingProfile == nil {
			profile = defaultProfile(childID, actor.TenantID, actor.BranchID)
			profile.ID = uid.NewUUID()
			created = true
		} else {
			profile = existingProfile
		}

		completenessBefore := domain.ComputeCompleteness(profile, nil, profile.CollectionPasswordHash != nil)

		changedSections, err := MergePatch(profile, patch)
		if err != nil {
			return err
		}

		if len(changedSections) == 0 {
			profileWithChild := domain.ProfileWithChild{
				Profile:       profile,
				Child:         child,
				ProfileExists: !created || existingProfile != nil,
			}
			if profile.ID != uuid.Nil && (existingProfile != nil || !created) {
				contacts, listErr := uc.repo.ListContactsByProfile(ctx, actor.TenantID, actor.BranchID, profile.ID)
				if listErr != nil {
					return domainerrors.Internal(fmt.Errorf("list contacts: %w", listErr))
				}
				profileWithChild.Contacts = contacts
			}
			completeness := domain.ComputeCompleteness(profile, profileWithChild.Contacts, profile.CollectionPasswordHash != nil)
			result = &UpdateProfileResult{
				ProfileWithChild: profileWithChild,
				Completeness:     completeness,
				Created:          false,
			}
			return nil
		}

		if created {
			savedProfile, createErr := uc.repo.Create(ctx, tx, profile)
			if createErr != nil {
				return domainerrors.Internal(fmt.Errorf("create profile: %w", createErr))
			}
			profile = savedProfile
		} else {
			savedProfile, updateErr := uc.repo.Update(ctx, tx, profile)
			if updateErr != nil {
				return domainerrors.Internal(fmt.Errorf("update profile: %w", updateErr))
			}
			profile = savedProfile
		}

		submittedContactTypes := SubmittedContactTypes(patch)
		if len(submittedContactTypes) > 0 {
			allEntries := make([]domain.ContactEntry, 0)
			for _, ct := range submittedContactTypes {
				var patches []ContactEntryPatch
				switch ct {
				case domain.ContactTypeParentCarer:
					patches = *patch.ParentCarers
				case domain.ContactTypeEmergencyContact:
					patches = *patch.EmergencyContacts
				case domain.ContactTypeAuthorisedCollector:
					patches = *patch.AuthorisedCollectors
				}
				for _, p := range patches {
					if strings.TrimSpace(p.FullName) == "" {
						return domainerrors.Validation("Invalid request payload.", "contact.full_name")
					}
				}
				entries := BuildContactEntries(ct, patches, profile.ID, actor.TenantID, actor.BranchID, childID)
				allEntries = append(allEntries, entries...)
			}
			if replaceErr := uc.repo.ReplaceContactsForTypes(ctx, tx, profile.ID, submittedContactTypes, allEntries); replaceErr != nil {
				return domainerrors.Internal(fmt.Errorf("replace contacts: %w", replaceErr))
			}
		}

		contacts, listErr := uc.repo.ListContactsByProfile(ctx, actor.TenantID, actor.BranchID, profile.ID)
		if listErr != nil {
			return domainerrors.Internal(fmt.Errorf("list contacts: %w", listErr))
		}

		completenessAfter := domain.ComputeCompleteness(profile, contacts, profile.CollectionPasswordHash != nil)

		auditAction := "registration_profile_updated"
		if created {
			auditAction = "registration_profile_created"
		}

		auditDetails := map[string]any{
			"child_id":             childID.String(),
			"profile_id":           profile.ID.String(),
			"changed_sections":     sectionCodesToStrings(changedSections),
			"profile_created":      created,
			"completeness_before":  completenessToStrings(completenessBefore),
			"completeness_after":   completenessToStrings(completenessAfter),
		}

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: auditAction,
			EntityType: "child_registration_profile",
			EntityID:   profile.ID,
			Details:    auditDetails,
		}); auditErr != nil {
			return domainerrors.Internal(fmt.Errorf("audit profile update: %w", auditErr))
		}

		result = &UpdateProfileResult{
			ProfileWithChild: domain.ProfileWithChild{
				Profile:       profile,
				Child:         child,
				Contacts:      contacts,
				ProfileExists: true,
			},
			Completeness: completenessAfter,
			Created:      created,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func hasAnySection(p PatchSection) bool {
	return p.DemographicsHome != nil ||
		p.MedicalDietary != nil ||
		p.HealthContacts != nil ||
		p.SocialDevelopment != nil ||
		p.ParentCarers != nil ||
		p.EmergencyContacts != nil ||
		p.AuthorisedCollectors != nil ||
		p.Collection != nil ||
		p.FundingSupport != nil ||
		p.RoutineCare != nil ||
		p.GDPRDeclaration != nil
}

func sectionCodesToStrings(codes []domain.SectionCode) []string {
	s := make([]string, len(codes))
	for i, c := range codes {
		s[i] = string(c)
	}
	return s
}

func completenessToStrings(c domain.Completeness) []string {
	s := make([]string, len(c.MissingSections))
	for i, ms := range c.MissingSections {
		s[i] = string(ms)
	}
	return s
}
