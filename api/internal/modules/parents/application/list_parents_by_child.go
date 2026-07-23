package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/parents/domain"
)

type ListParentsByChildUseCase struct {
	repo domain.Repository
}

func NewListParentsByChildUseCase(repo domain.Repository) *ListParentsByChildUseCase {
	return &ListParentsByChildUseCase{repo: repo}
}

type ParentForChild struct {
	Parent       domain.Parent
	LinkID       uuid.UUID
	Relationship string
}

func (uc *ListParentsByChildUseCase) Execute(ctx context.Context, actor ActorContext, childID uuid.UUID) ([]ParentForChild, error) {
	links, err := uc.repo.ListParentsByChild(ctx, nil, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return nil, err
	}

	var result []ParentForChild
	for _, link := range links {
		parent, found, err := uc.repo.GetByID(ctx, nil, actor.TenantID, actor.BranchID, link.ParentID)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		var rel string
		if parent.RelationshipToChild != nil {
			rel = *parent.RelationshipToChild
		}
		result = append(result, ParentForChild{
			Parent:       parent,
			LinkID:       link.ID,
			Relationship: rel,
		})
	}

	return result, nil
}
