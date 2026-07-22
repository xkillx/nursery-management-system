package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ParentFundingBreakdown struct {
	Record                 domain.FundingRecord
	FundedAllowanceMinutes int
	Allocation             []domain.AllocationEntry
	History                []domain.FundingHistory
}

type GetParentFundingBreakdown struct {
	recordRepo domain.FundingRecordRepository
	queryRepo  domain.FundingQueryRepository
	historyR   domain.HistoryRepository
	childLook  ParentChildLookupForFunding
	termDates  domain.TermDateProvider
}

func NewGetParentFundingBreakdown(recordRepo domain.FundingRecordRepository, queryRepo domain.FundingQueryRepository, historyR domain.HistoryRepository, childLook ParentChildLookupForFunding, termDates domain.TermDateProvider) *GetParentFundingBreakdown {
	return &GetParentFundingBreakdown{recordRepo: recordRepo, queryRepo: queryRepo, historyR: historyR, childLook: childLook, termDates: termDates}
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

	record, found, err := uc.recordRepo.GetFundingRecord(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}
	if !found {
		return ParentFundingBreakdown{}, nil
	}

	billingMonthEnd := billingMonth.AddDate(0, 1, 0).AddDate(0, 0, -1)
	allocation, err := uc.queryRepo.GetChildAllocation(ctx, actor.TenantID, actor.BranchID, childID, billingMonth, billingMonthEnd)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}

	history, err := uc.historyR.ListByChild(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return ParentFundingBreakdown{}, domainerrors.Internal(err)
	}

	allowance := 0
	if record.FundingEnabled && record.FundedHoursPerWeek != nil && *record.FundedHoursPerWeek > 0 {
		termDateRanges, _ := uc.termDates.GetTermDatesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
		allowance, _ = domain.ComputeAllowanceMinutes(*record.FundedHoursPerWeek, record.FundingModel, termDateRanges, billingMonth, nil, record.FundingStartDate, record.FundingEndDate)
	}

	return ParentFundingBreakdown{
		Record:                 record,
		FundedAllowanceMinutes: allowance,
		Allocation:             allocation,
		History:                history,
	}, nil
}
