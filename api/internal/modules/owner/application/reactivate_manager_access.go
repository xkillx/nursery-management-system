package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/domain"
)

type ReactivateManagerAccessUseCase struct {
	repo domain.ManagerAccessRepository
}

func NewReactivateManagerAccessUseCase(repo domain.ManagerAccessRepository) *ReactivateManagerAccessUseCase {
	return &ReactivateManagerAccessUseCase{repo: repo}
}

func (uc *ReactivateManagerAccessUseCase) Execute(ctx context.Context, actor domain.OwnerActor, siteID, membershipID uuid.UUID) error {
	if _, err := uc.repo.GetActiveSite(ctx, actor.TenantID, siteID); err != nil {
		return err
	}

	if err := uc.repo.ReactivateManagerMembership(ctx, membershipID, actor.TenantID); err != nil {
		return fmt.Errorf("reactivate: %w", err)
	}

	return nil
}
