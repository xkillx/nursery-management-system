package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListChildren struct {
	repo domain.Repository
}

func NewListChildren(repo domain.Repository) *ListChildren {
	return &ListChildren{repo: repo}
}

func (uc *ListChildren) Execute(ctx context.Context, actor tenant.ActorContext, statusFilter string, limit, offset int) ([]domain.Child, error) {
	sf, err := ValidateStatusFilter(statusFilter)
	if err != nil {
		return nil, err
	}

	limit, offset, err = ValidatePagination(limit, offset)
	if err != nil {
		return nil, err
	}

	children, err := uc.repo.List(ctx, actor.TenantID, actor.BranchID, sf, limit, offset)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list children: %w", err))
	}

	return children, nil
}

func (uc *ListChildren) Count(ctx context.Context, actor tenant.ActorContext, statusFilter string) (int, error) {
	sf, err := ValidateStatusFilter(statusFilter)
	if err != nil {
		return 0, err
	}

	n, err := uc.repo.Count(ctx, actor.TenantID, actor.BranchID, sf)
	if err != nil {
		return 0, domainerrors.Internal(fmt.Errorf("count children: %w", err))
	}

	return n, nil
}
