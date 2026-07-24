package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetChildFunding struct {
	repo       domain.Repository
	fundingRdr domain.ChildFundingReader
}

func NewGetChildFunding(repo domain.Repository, fundingRdr domain.ChildFundingReader) *GetChildFunding {
	return &GetChildFunding{repo: repo, fundingRdr: fundingRdr}
}

func (uc *GetChildFunding) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildFundingData, error) {
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
	data, found, err := uc.fundingRdr.GetFundingForChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child funding: %w", err))
	}
	if !found {
		return nil, nil
	}
	return data, nil
}
