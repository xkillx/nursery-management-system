package application

import (
	"context"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type OverdueSummary struct {
	repo domain.BillingRepository
}

func NewOverdueSummary(repo domain.BillingRepository) *OverdueSummary {
	return &OverdueSummary{repo: repo}
}

type OverdueSummaryResult struct {
	TotalOverdueMinor int
	OverdueCount      int
	Items             []domain.OverdueSummaryItem
}

func (uc *OverdueSummary) Execute(ctx context.Context, actor tenant.ActorContext) (OverdueSummaryResult, error) {
	summary, err := uc.repo.InvoiceOverdueSummary(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return OverdueSummaryResult{}, domainerrors.Internal(err)
	}

	items, err := uc.repo.InvoiceOverdueTopItems(ctx, actor.TenantID, actor.BranchID)
	if err != nil {
		return OverdueSummaryResult{}, domainerrors.Internal(err)
	}

	if items == nil {
		items = []domain.OverdueSummaryItem{}
	}

	return OverdueSummaryResult{
		TotalOverdueMinor: summary.TotalOverdueMinor,
		OverdueCount:      summary.OverdueCount,
		Items:             items,
	}, nil
}
