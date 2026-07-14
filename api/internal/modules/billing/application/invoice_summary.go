package application

import (
	"context"
	"strings"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type InvoiceSummary struct {
	repo domain.BillingRepository
}

func NewInvoiceSummary(repo domain.BillingRepository) *InvoiceSummary {
	return &InvoiceSummary{repo: repo}
}

type InvoiceSummaryParams struct {
	BillingMonthFrom *string
	BillingMonthTo   *string
}

type InvoiceSummaryResult struct {
	Months []domain.InvoiceMonthSummary
}

func (uc *InvoiceSummary) Execute(ctx context.Context, actor tenant.ActorContext, params InvoiceSummaryParams) (InvoiceSummaryResult, error) {
	filters := domain.InvoiceExportFilters{}

	if params.BillingMonthFrom != nil {
		bm, err := ParseBillingMonth(strings.TrimSpace(*params.BillingMonthFrom))
		if err != nil {
			return InvoiceSummaryResult{}, domainerrors.Validation("Invalid billing_month_from format. Use YYYY-MM.", "billing_month_from")
		}
		filters.BillingMonthFrom = &bm
	}

	if params.BillingMonthTo != nil {
		bm, err := ParseBillingMonth(strings.TrimSpace(*params.BillingMonthTo))
		if err != nil {
			return InvoiceSummaryResult{}, domainerrors.Validation("Invalid billing_month_to format. Use YYYY-MM.", "billing_month_to")
		}
		filters.BillingMonthTo = &bm
	}

	if filters.BillingMonthFrom != nil && filters.BillingMonthTo != nil {
		months := monthsBetween(*filters.BillingMonthFrom, *filters.BillingMonthTo)
		if months > maxExportMonths {
			return InvoiceSummaryResult{}, domainerrors.Validation("Date range exceeds maximum of 24 months.", "billing_month_from")
		}
	}

	rows, err := uc.repo.InvoiceSummaryByMonth(ctx, actor.TenantID, actor.BranchID, filters)
	if err != nil {
		return InvoiceSummaryResult{}, domainerrors.Internal(err)
	}

	if rows == nil {
		rows = []domain.InvoiceMonthSummary{}
	}

	return InvoiceSummaryResult{Months: rows}, nil
}
