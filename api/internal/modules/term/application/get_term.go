package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetTermUseCase struct {
	repo domain.Repository
}

func NewGetTermUseCase(repo domain.Repository) *GetTermUseCase {
	return &GetTermUseCase{repo: repo}
}

func (uc *GetTermUseCase) Execute(ctx context.Context, actor tenant.ActorContext, termIDRaw string) (*domain.Term, error) {
	id, err := uuid.Parse(termIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "term_id")
	}
	term, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get term: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("term", "Resource not found.")
	}
	return term, nil
}

type GetCurrentTermForChildUseCase struct {
	repo domain.Repository
}

func NewGetCurrentTermForChildUseCase(repo domain.Repository) *GetCurrentTermForChildUseCase {
	return &GetCurrentTermForChildUseCase{repo: repo}
}

func (uc *GetCurrentTermForChildUseCase) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (*domain.Term, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	term, found, err := uc.repo.GetActiveForChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get current term: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("term", "No active term for this child.")
	}
	return term, nil
}
