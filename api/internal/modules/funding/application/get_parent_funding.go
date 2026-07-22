package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ParentFundingEntitlement struct {
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	FundingType            *string
	FundedHoursPerWeek     *float64
	FundingStartDate       *string
	FundingEndDate         *string
	FundedAllowanceMinutes int
	BookedHoursThisWeek    float64
}

type ParentChildLookupForFunding interface {
	ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error)
}

type GetParentFunding struct {
	recordRepo domain.FundingRecordRepository
	childLook  ParentChildLookupForFunding
	termDates  domain.TermDateProvider
}

func NewGetParentFunding(recordRepo domain.FundingRecordRepository, childLook ParentChildLookupForFunding, termDates domain.TermDateProvider) *GetParentFunding {
	return &GetParentFunding{recordRepo: recordRepo, childLook: childLook, termDates: termDates}
}

func (uc *GetParentFunding) Execute(ctx context.Context, actor tenant.ActorContext) ([]ParentFundingEntitlement, error) {
	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}
	if len(childIDs) == 0 {
		return nil, nil
	}

	billingMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	termDateRanges, _ := uc.termDates.GetTermDatesForBranchAndMonth(ctx, actor.TenantID, actor.BranchID, billingMonth)

	var results []ParentFundingEntitlement
	for _, childID := range childIDs {
		record, found, err := uc.recordRepo.GetFundingRecord(ctx, actor.TenantID, actor.BranchID, childID)
		if err != nil {
			return nil, domainerrors.Internal(err)
		}
		if !found || !record.FundingEnabled {
			results = append(results, ParentFundingEntitlement{
				ChildID: childID,
			})
			continue
		}

		allowance := 0
		if record.FundedHoursPerWeek != nil && *record.FundedHoursPerWeek > 0 {
			allowance, _ = domain.ComputeAllowanceMinutes(*record.FundedHoursPerWeek, record.FundingModel, termDateRanges, billingMonth, nil, record.FundingStartDate, record.FundingEndDate)
		}

		ft := string(record.FundingType)
		results = append(results, ParentFundingEntitlement{
			ChildID:                childID,
			FundingType:            &ft,
			FundedHoursPerWeek:     record.FundedHoursPerWeek,
			FundedAllowanceMinutes: allowance,
		})
	}

	return results, nil
}
