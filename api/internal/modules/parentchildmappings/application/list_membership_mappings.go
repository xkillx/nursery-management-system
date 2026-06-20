package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parentchildmappings/domain"
)

// ListMembershipMappingsUseCase returns the active child mappings for a given
// parent membership. Used by manager-side "this parent has these children" views.
type ListMembershipMappingsUseCase struct {
	repo domain.Repository
}

func NewListMembershipMappingsUseCase(repo domain.Repository) *ListMembershipMappingsUseCase {
	return &ListMembershipMappingsUseCase{repo: repo}
}

func (uc *ListMembershipMappingsUseCase) Execute(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) ([]domain.ParentChildMapping, error) {
	return uc.repo.ListActiveByMembership(ctx, tx, tenantID, branchID, membershipID)
}
