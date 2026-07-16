package application

import (
	"context"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type GetEnhancedOverview struct {
	repo domain.Repository
}

func NewGetEnhancedOverview(repo domain.Repository) *GetEnhancedOverview {
	return &GetEnhancedOverview{repo: repo}
}

func (uc *GetEnhancedOverview) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string, expiringDays int) (domain.EnhancedOverviewMetrics, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.EnhancedOverviewMetrics{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	metrics, err := uc.repo.GetFundedChildrenCount(ctx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return domain.EnhancedOverviewMetrics{}, domainerrors.Internal(err)
	}

	bookedHours, err := uc.repo.GetBookedHoursThisWeek(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return domain.EnhancedOverviewMetrics{}, domainerrors.Internal(err)
	}
	metrics.BookedHoursThisWeek = bookedHours

	expiringCount, err := uc.repo.GetExpiringSoonCount(ctx, actor.TenantID, actor.BranchID, expiringDays)
	if err != nil {
		return domain.EnhancedOverviewMetrics{}, domainerrors.Internal(err)
	}
	metrics.ExpiringSoonCount = expiringCount

	return metrics, nil
}
