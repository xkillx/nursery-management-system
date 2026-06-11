package application

import (
	"context"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetConsents struct {
	repo domain.ConsentRepository
}

func NewGetConsents(repo domain.ConsentRepository) *GetConsents {
	return &GetConsents{repo: repo}
}

func (uc *GetConsents) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (domain.ConsentWithCompleteness, error) {
	cid, err := uuid.Parse(childID)
	if err != nil {
		return domain.ConsentWithCompleteness{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	current, err := uc.repo.GetLatestByChild(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return domain.ConsentWithCompleteness{}, domainerrors.Internal(err)
	}

	history, err := uc.repo.ListByChild(ctx, actor.TenantID, actor.BranchID, cid)
	if err != nil {
		return domain.ConsentWithCompleteness{}, domainerrors.Internal(err)
	}

	completeness := domain.ComputeConsentCompleteness(current)

	return domain.ConsentWithCompleteness{
		Current:      current,
		History:      history,
		Completeness: completeness,
	}, nil
}
