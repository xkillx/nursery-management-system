package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/domain"
)

type ManagerAccessItem struct {
	MembershipID uuid.UUID
	UserID       uuid.UUID
	Email        string
	IsActive     bool
}

type ListManagerAccessUseCase struct {
	repo domain.ManagerAccessRepository
}

func NewListManagerAccessUseCase(repo domain.ManagerAccessRepository) *ListManagerAccessUseCase {
	return &ListManagerAccessUseCase{repo: repo}
}

func (uc *ListManagerAccessUseCase) Execute(ctx context.Context, actor domain.OwnerActor, siteID uuid.UUID, statusFilter string) ([]ManagerAccessItem, error) {
	if _, err := uc.repo.GetActiveSite(ctx, actor.TenantID, siteID); err != nil {
		return nil, err
	}

	entries, err := uc.repo.ListManagerAccess(ctx, actor.TenantID, siteID, statusFilter)
	if err != nil {
		return nil, err
	}

	items := make([]ManagerAccessItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, ManagerAccessItem{
			MembershipID: e.MembershipID,
			UserID:       e.UserID,
			Email:        e.Email,
			IsActive:     e.IsActive,
		})
	}
	return items, nil
}
