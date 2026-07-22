package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetEnhancedChildDetail struct {
	recordRepo domain.FundingRecordRepository
	queryRepo  domain.FundingQueryRepository
	historyR   domain.HistoryRepository
	termDates  domain.TermDateProvider
}

func NewGetEnhancedChildDetail(recordRepo domain.FundingRecordRepository, queryRepo domain.FundingQueryRepository, historyR domain.HistoryRepository, termDates domain.TermDateProvider) *GetEnhancedChildDetail {
	return &GetEnhancedChildDetail{recordRepo: recordRepo, queryRepo: queryRepo, historyR: historyR, termDates: termDates}
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

	record, found, err := uc.recordRepo.GetFundingRecord(ctx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return domain.EnhancedChildDetail{}, domainerrors.Internal(err)
	}
	if !found {
		return domain.EnhancedChildDetail{}, domainerrors.NotFound("funding_record", "No funding record found for this child.")
	}

	billingMonthEnd := billingMonth.AddDate(0, 1, -1)
	allocation, err := uc.queryRepo.GetChildAllocation(ctx, actor.TenantID, actor.BranchID, childID, billingMonth, billingMonthEnd)
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

	allowance := 0
	if record.FundingEnabled && record.FundedHoursPerWeek != nil && *record.FundedHoursPerWeek > 0 {
		termDateRanges, _ := uc.termDates.GetTermDatesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)
		allowance, _ = domain.ComputeAllowanceMinutes(*record.FundedHoursPerWeek, record.FundingModel, termDateRanges, billingMonth, nil, record.FundingStartDate, record.FundingEndDate)
	}

	return domain.EnhancedChildDetail{
		Record:                 record,
		FundedAllowanceMinutes: allowance,
		Allocation:             allocation,
		History:                history,
	}, nil
}
