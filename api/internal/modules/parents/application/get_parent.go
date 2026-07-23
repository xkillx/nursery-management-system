package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/parents/domain"
)

type GetParentUseCase struct {
	repo domain.Repository
}

func NewGetParentUseCase(repo domain.Repository) *GetParentUseCase {
	return &GetParentUseCase{repo: repo}
}

type ParentWithChildren struct {
	Parent   domain.Parent
	Children []domain.ParentChild
}

func (uc *GetParentUseCase) Execute(ctx context.Context, actor ActorContext, parentID uuid.UUID) (ParentWithChildren, error) {
	parent, found, err := uc.repo.GetByID(ctx, nil, actor.TenantID, actor.BranchID, parentID)
	if err != nil {
		return ParentWithChildren{}, err
	}
	if !found {
		return ParentWithChildren{}, domain.ErrParentNotFound
	}

	children, err := uc.repo.ListChildrenByParent(ctx, nil, actor.TenantID, actor.BranchID, parentID)
	if err != nil {
		return ParentWithChildren{}, err
	}

	return ParentWithChildren{
		Parent:   parent,
		Children: children,
	}, nil
}
