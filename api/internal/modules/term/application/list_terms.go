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

type ListTermsForChildResult struct {
	Items []domain.Term
	Total int
}

func (uc *ListTermsForChildUseCase) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string, limit, offset int) (ListTermsForChildResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return ListTermsForChildResult{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	terms, err := uc.repo.ListByChildPaginated(ctx, actor.TenantID, actor.BranchID, childID, limit, offset)
	if err != nil {
		return ListTermsForChildResult{}, domainerrors.Internal(fmt.Errorf("list terms: %w", err))
	}
	total, err := uc.repo.CountByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return ListTermsForChildResult{}, domainerrors.Internal(fmt.Errorf("count terms: %w", err))
	}
	return ListTermsForChildResult{Items: terms, Total: total}, nil
}

type ListExpiringTermsUseCase struct {
	repo domain.Repository
	now  func() time.Time
}

func NewListExpiringTermsUseCase(repo domain.Repository) *ListExpiringTermsUseCase {
	return &ListExpiringTermsUseCase{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

type ListExpiringTermsResult struct {
	Items []domain.Term
	Total int
}

func (uc *ListExpiringTermsUseCase) Execute(ctx context.Context, actor tenant.ActorContext, expiringWithinDays, limit, offset int) (ListExpiringTermsResult, error) {
	if expiringWithinDays <= 0 {
		expiringWithinDays = 30
	}
	today := uc.now().UTC().Truncate(24 * time.Hour)
	maxEnd := today.AddDate(0, 0, expiringWithinDays)
	terms, err := uc.repo.ListExpiringWithinPaginated(ctx, actor.TenantID, actor.BranchID, maxEnd, limit, offset)
	if err != nil {
		return ListExpiringTermsResult{}, domainerrors.Internal(fmt.Errorf("list expiring terms: %w", err))
	}
	total, err := uc.repo.CountExpiringWithin(ctx, actor.TenantID, actor.BranchID, maxEnd)
	if err != nil {
		return ListExpiringTermsResult{}, domainerrors.Internal(fmt.Errorf("count expiring terms: %w", err))
	}
	return ListExpiringTermsResult{Items: terms, Total: total}, nil
}
