package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/domain"
)

type DeactivateManagerAccessUseCase struct {
	repo domain.ManagerAccessRepository
}

func NewDeactivateManagerAccessUseCase(repo domain.ManagerAccessRepository) *DeactivateManagerAccessUseCase {
	return &DeactivateManagerAccessUseCase{repo: repo}
}

func (uc *DeactivateManagerAccessUseCase) Execute(ctx context.Context, actor domain.OwnerActor, siteID, membershipID uuid.UUID) error {
	if _, err := uc.repo.GetActiveSite(ctx, actor.TenantID, siteID); err != nil {
		return err
	}

	rows, err := uc.repo.DeactivateManagerMembership(ctx, membershipID, actor.TenantID)
	if err != nil {
		return fmt.Errorf("deactivate: %w", err)
	}
	if rows == 0 {
		return domain.ErrMembershipNotFound
	}

	if err := uc.repo.RevokeRefreshTokensByMembership(ctx, membershipID); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}

	return nil
}
