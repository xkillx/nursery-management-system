package application

import (
	"context"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListExpiring struct {
	repo domain.Repository
}

func NewListExpiring(repo domain.Repository) *ListExpiring {
	return &ListExpiring{repo: repo}
}

func (uc *ListExpiring) Execute(ctx context.Context, actor tenant.ActorContext, withinDays int) ([]domain.ExpiringFundingRecord, error) {
	if withinDays < 0 {
		return nil, domainerrors.Validation("within must be a non-negative integer.", "within")
	}

	records, err := uc.repo.ListExpiringSoon(ctx, actor.TenantID, actor.BranchID, withinDays)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}
	if records == nil {
		records = []domain.ExpiringFundingRecord{}
	}
	return records, nil
}
