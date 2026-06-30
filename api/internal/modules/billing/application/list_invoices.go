package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListInvoices struct {
	repo domain.BillingRepository
}

func NewListInvoices(repo domain.BillingRepository) *ListInvoices {
	return &ListInvoices{repo: repo}
}

type ListInvoicesParams struct {
	BillingMonth     *string
	BillingMonthFrom *string
	BillingMonthTo   *string
	Status           *string
	ChildID          *string
	Limit            *string
	Offset           *string
}

type ListInvoicesResult struct {
	Items  []domain.InvoiceReviewRow
	Limit  int
	Offset int
}

var validStatuses = map[string]bool{
	domain.InvoiceStatusDraft:         true,
	domain.InvoiceStatusIssued:        true,
	domain.InvoiceStatusPaymentFailed: true,
	domain.InvoiceStatusPaid:          true,
	domain.InvoiceStatusOverdue:       true,
}

func (uc *ListInvoices) Execute(ctx context.Context, actor tenant.ActorContext, params ListInvoicesParams) (ListInvoicesResult, error) {
	filters := domain.InvoiceReviewFilters{
		Limit:  50,
		Offset: 0,
	}

	if params.BillingMonth != nil {
		bm, err := ParseBillingMonth(*params.BillingMonth)
		if err != nil {
			return ListInvoicesResult{}, domainerrors.Validation("Invalid billing_month format. Use YYYY-MM.", "billing_month")
		}
		filters.BillingMonth = &bm
	}

	if params.BillingMonthFrom != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthFrom)
		if err != nil {
			return ListInvoicesResult{}, domainerrors.Validation("Invalid billing_month_from format. Use YYYY-MM.", "billing_month_from")
		}
		filters.BillingMonthFrom = &bm
	}

	if params.BillingMonthTo != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthTo)
		if err != nil {
			return ListInvoicesResult{}, domainerrors.Validation("Invalid billing_month_to format. Use YYYY-MM.", "billing_month_to")
		}
		filters.BillingMonthTo = &bm
	}

	if params.Status != nil {
		s := strings.TrimSpace(*params.Status)
		if !validStatuses[s] {
			return ListInvoicesResult{}, domainerrors.Validation(fmt.Sprintf("Invalid status filter: %s.", s), "status")
		}
		filters.Status = &s
	}

	if params.ChildID != nil {
		cid, err := uuid.Parse(strings.TrimSpace(*params.ChildID))
		if err != nil {
			return ListInvoicesResult{}, domainerrors.Validation("Invalid child_id format.", "child_id")
		}
		filters.ChildID = &cid
	}

	if params.Limit != nil {
		l, err := strconv.Atoi(*params.Limit)
		if err != nil || l < 1 || l > 200 {
			return ListInvoicesResult{}, domainerrors.Validation("Limit must be between 1 and 200.", "limit")
		}
		filters.Limit = l
	}

	if params.Offset != nil {
		o, err := strconv.Atoi(*params.Offset)
		if err != nil || o < 0 {
			return ListInvoicesResult{}, domainerrors.Validation("Offset must be 0 or greater.", "offset")
		}
		filters.Offset = o
	}

	rows, err := uc.repo.ListInvoicesForManagerReview(ctx, actor.TenantID, actor.BranchID, filters)
	if err != nil {
		return ListInvoicesResult{}, domainerrors.Internal(err)
	}

	if rows == nil {
		rows = []domain.InvoiceReviewRow{}
	}

	return ListInvoicesResult{
		Items:  rows,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}
