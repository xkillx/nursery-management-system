package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type SetCollectionPassword struct {
	repo        domain.Repository
	auditWriter *audit.Writer
	txManager   *transaction.Manager
}

func NewSetCollectionPassword(repo domain.Repository, auditWriter *audit.Writer, txManager *transaction.Manager) *SetCollectionPassword {
	return &SetCollectionPassword{repo: repo, auditWriter: auditWriter, txManager: txManager}
}

type SetCollectionPasswordResult struct {
	ProfileWithChild domain.ProfileWithChild
	Completeness     domain.Completeness
}

func (uc *SetCollectionPassword) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string, password string) (*SetCollectionPasswordResult, error) {
	childID, err := uuid.Parse(strings.TrimSpace(childIDRaw))
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	password = strings.TrimSpace(password)
	if len(password) < MinPasswordLen || len(password) > MaxPasswordLen {
		return nil, domainerrors.Validation("Invalid request payload.", "password")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("hash password: %w", err))
	}

	var result *SetCollectionPasswordResult

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

		wasPreviouslySet := false

		if existingProfile == nil {
			profile := defaultProfile(childID, actor.TenantID, actor.BranchID)
			profile.ID = uid.NewUUID()
			if _, createErr := uc.repo.Create(ctx, tx, profile); createErr != nil {
				return domainerrors.Internal(fmt.Errorf("create profile: %w", createErr))
			}
		} else {
			wasPreviouslySet = existingProfile.CollectionPasswordHash != nil
		}

		now := time.Now().UTC()
		if err := uc.repo.SetCollectionPassword(ctx, tx, actor.TenantID, actor.BranchID, childID, hash, now, actor.UserID, actor.MembershipID); err != nil {
			return domainerrors.Internal(fmt.Errorf("set collection password: %w", err))
		}

		profile, err := uc.repo.GetByChild(ctx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get profile after password set: %w", err))
		}

		contacts, listErr := uc.repo.ListContactsByProfile(ctx, actor.TenantID, actor.BranchID, profile.ID)
		if listErr != nil {
			return domainerrors.Internal(fmt.Errorf("list contacts: %w", listErr))
		}

		completeness := domain.ComputeCompleteness(profile, contacts, true)

		if auditErr := uc.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "registration_collection_password_set",
			EntityType: "child_registration_profile",
			EntityID:   profile.ID,
			Details: map[string]any{
				"child_id":                  childID.String(),
				"profile_id":                profile.ID.String(),
				"password_was_previously_set": wasPreviouslySet,
			},
		}); auditErr != nil {
			return domainerrors.Internal(fmt.Errorf("audit password set: %w", auditErr))
		}

		result = &SetCollectionPasswordResult{
			ProfileWithChild: domain.ProfileWithChild{
				Profile:       profile,
				Child:         child,
				Contacts:      contacts,
				ProfileExists: true,
			},
			Completeness: completeness,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
