package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetEnhancedChildDetail struct {
	repo     domain.Repository
	historyR domain.HistoryRepository
}

func NewGetEnhancedChildDetail(repo domain.Repository, historyR domain.HistoryRepository) *GetEnhancedChildDetail {
	return &GetEnhancedChildDetail{repo: repo, historyR: historyR}
}

func (uc *GetEnhancedChildDetail) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw, billingMonthRaw string) (domain.EnhancedChildDetail, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Validation("Invalid child ID.", "child_id")
	}

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	profile, found, err := uc.repo.Get(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Internal(err)
	}
	if !found {
		return domain.EnhancedChildDetail{}, domainerrors.NotFound("funding_profile", "No funding profile found for this child and billing month.")
	}

	billingMonthEnd := billingMonth.AddDate(0, 1, -1)
	allocation, err := uc.repo.GetChildAllocation(ctx, actor.TenantID, actor.BranchID, childID, billingMonth, billingMonthEnd)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Internal(err)
	}
	if allocation == nil {
		allocation = []domain.AllocationEntry{}
	}

	history, err := uc.historyR.ListByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Internal(err)
	}
	if history == nil {
		history = []domain.FundingHistory{}
	}

	return domain.EnhancedChildDetail{
		Profile:    profile,
		Allocation: allocation,
		History:    history,
	}, nil
}

func billingMonthEnd(billingMonth time.Time) time.Time {
	return billingMonth.AddDate(0, 1, -1)
}
