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
	repo      domain.Repository
	childLook ParentChildLookupForFunding
}

func NewGetParentFunding(repo domain.Repository, childLook ParentChildLookupForFunding) *GetParentFunding {
	return &GetParentFunding{repo: repo, childLook: childLook}
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

	var results []ParentFundingEntitlement
	for _, childID := range childIDs {
		profile, found, err := uc.repo.Get(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
		if err != nil {
			return nil, domainerrors.Internal(err)
		}
		if !found {
			results = append(results, ParentFundingEntitlement{
				ChildID: childID,
			})
			continue
		}
		results = append(results, ParentFundingEntitlement{
			ChildID:                childID,
			FundingType:            profile.FundingType,
			FundedHoursPerWeek:     profile.FundedHoursPerWeek,
			FundedAllowanceMinutes: profile.FundedAllowanceMinutes,
		})
	}

	return results, nil
}
