package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ParentFundingBreakdown struct {
	Profile    domain.FundingProfile
	Allocation []domain.AllocationEntry
	History    []domain.FundingHistory
}

type GetParentFundingBreakdown struct {
	repo        domain.Repository
	historyRepo domain.HistoryRepository
	childLook   ParentChildLookupForFunding
}

func NewGetParentFundingBreakdown(repo domain.Repository, historyRepo domain.HistoryRepository, childLook ParentChildLookupForFunding) *GetParentFundingBreakdown {
	return &GetParentFundingBreakdown{repo: repo, historyRepo: historyRepo, childLook: childLook}
}

func (uc *GetParentFundingBreakdown) Execute(ctx context.Context, actor tenant.ActorContext, childID uuid.UUID, billingMonthRaw string) (ParentFundingBreakdown, error) {
	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}

	childAllowed := false
	for _, cid := range childIDs {
		if cid == childID {
			childAllowed = true
			break
		}
	}
	if !childAllowed {
		return ParentFundingBreakdown{}, domainerrors.Forbidden("child_not_in_scope", "Child is not linked to your account.")
	}

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	profile, found, err := uc.repo.Get(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}
	if !found {
		return ParentFundingBreakdown{}, nil
	}

	billingMonthStart := billingMonth
	billingMonthEnd := billingMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
	allocation, err := uc.repo.GetChildAllocation(ctx, actor.TenantID, actor.BranchID, childID, billingMonthStart, billingMonthEnd)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}

	history, err := uc.historyRepo.ListByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}

	return ParentFundingBreakdown{
		Profile:    profile,
		Allocation: allocation,
		History:    history,
	}, nil
}
