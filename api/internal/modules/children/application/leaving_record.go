package application

import (
	"context"
	"fmt"

	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetLeavingRecord struct {
	repo domain.Repository
}

func NewGetLeavingRecord(repo domain.Repository) *GetLeavingRecord {
	return &GetLeavingRecord{repo: repo}
}

func (uc *GetLeavingRecord) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildLeavingRecord, error) {
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
	lr, found, err := uc.repo.GetLeavingRecordByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child leaving record: %w", err))
	}
	if !found {
		return nil, nil
	}
	return lr, nil
}
