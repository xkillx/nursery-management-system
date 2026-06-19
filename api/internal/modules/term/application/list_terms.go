package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListTermsForChildUseCase struct {
	repo domain.Repository
}

func NewListTermsForChildUseCase(repo domain.Repository) *ListTermsForChildUseCase {
	return &ListTermsForChildUseCase{repo: repo}
}

func (uc *ListTermsForChildUseCase) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) ([]domain.Term, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	terms, err := uc.repo.ListByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list terms: %w", err))
	}
	return terms, nil
}

type ListExpiringTermsUseCase struct {
	repo domain.Repository
	now  func() time.Time
}

func NewListExpiringTermsUseCase(repo domain.Repository) *ListExpiringTermsUseCase {
	return &ListExpiringTermsUseCase{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (uc *ListExpiringTermsUseCase) Execute(ctx context.Context, actor tenant.ActorContext, expiringWithinDays int) ([]domain.Term, error) {
	if expiringWithinDays <= 0 {
		expiringWithinDays = 30
	}
	today := uc.now().UTC().Truncate(24 * time.Hour)
	maxEnd := today.AddDate(0, 0, expiringWithinDays)
	terms, err := uc.repo.ListExpiringWithin(ctx, actor.TenantID, actor.BranchID, maxEnd)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list expiring terms: %w", err))
	}
	return terms, nil
}
