package application

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type GetProfile struct {
	repo domain.Repository
}

func NewGetProfile(repo domain.Repository) *GetProfile {
	return &GetProfile{repo: repo}
}

func (uc *GetProfile) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildProfile, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	p, err := uc.repo.GetProfileByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child profile: %w", err))
	}
	return p, nil
}

type UpdateProfile struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateProfile(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateProfile {
	return &UpdateProfile{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *UpdateProfile) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildProfileInput) (*domain.ChildProfile, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}

	var result *domain.ChildProfile
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		existing, eerr := uc.repo.GetProfileForUpdate(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("get profile for update: %w", eerr))
		}
		var out *domain.ChildProfile
		if existing == nil {
			p := buildChildProfileFromInput(actor.TenantID, actor.BranchID, id, in)
			created, eerr := uc.repo.InsertProfile(ctx, tx, p)
			if eerr != nil {
				return domainerrors.Internal(fmt.Errorf("insert child profile: %w", eerr))
			}
			out = created
		} else {
			updated := buildChildProfileFromInput(actor.TenantID, actor.BranchID, id, in)
			updated.ID = existing.ID
			saved, eerr := uc.repo.UpdateProfile(ctx, tx, updated)
			if eerr != nil {
				return domainerrors.Internal(fmt.Errorf("update child profile: %w", eerr))
			}
			out = saved
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_profile_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_profile_updated: %w", aerr))
		}
		result = out
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
