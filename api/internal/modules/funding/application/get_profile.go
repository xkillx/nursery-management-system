package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetProfile struct {
	repo domain.Repository
}

func NewGetProfile(repo domain.Repository) *GetProfile {
	return &GetProfile{repo: repo}
}

func (uc *GetProfile) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw, billingMonthRaw string) (domain.FundingProfile, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return domain.FundingProfile{}, domainerrors.Validation("Invalid child ID.", "child_id")
	}

	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.FundingProfile{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	profile, found, err := uc.repo.Get(ctx, actor.TenantID, actor.BranchID, childID, billingMonth)
	if err != nil {
		return domain.FundingProfile{}, domainerrors.Internal(err)
	}
	if !found {
		return domain.FundingProfile{}, domainerrors.NotFound("funding_profile", "No funding profile found for this child and billing month.")
	}

	return profile, nil
}

func validateMonthOverlap(billingMonth time.Time, enrollment domain.ChildEnrollment) bool {
	nextMonth := billingMonth.AddDate(0, 1, 0)
	if !enrollment.StartDate.Before(nextMonth) {
		return false
	}
	if enrollment.EndDate != nil && !enrollment.EndDate.After(billingMonth.AddDate(0, 0, -1)) {
		return false
	}
	return true
}
