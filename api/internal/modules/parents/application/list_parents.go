package application

import (
	"context"

	"nursery-management-system/api/internal/modules/parents/domain"
)

type ListParentsUseCase struct {
	repo domain.Repository
}

func NewListParentsUseCase(repo domain.Repository) *ListParentsUseCase {
	return &ListParentsUseCase{repo: repo}
}

type ListParentsResult struct {
	Parents    []domain.Parent
	TotalCount int
}

func (uc *ListParentsUseCase) Execute(ctx context.Context, actor ActorContext, isActive *bool, search *string, page, pageSize int) (ListParentsResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	parents, total, err := uc.repo.ListFiltered(ctx, nil, actor.TenantID, actor.BranchID, isActive, search, pageSize, offset)
	if err != nil {
		return ListParentsResult{}, err
	}

	return ListParentsResult{
		Parents:    parents,
		TotalCount: total,
	}, nil
}
