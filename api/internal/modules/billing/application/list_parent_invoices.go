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

type ListParentInvoices struct {
	repo domain.BillingRepository
}

func NewListParentInvoices(repo domain.BillingRepository) *ListParentInvoices {
	return &ListParentInvoices{repo: repo}
}

type ListParentInvoicesParams struct {
	BillingMonth *string
	Status       *string
	ChildID      *string
	Limit        *string
	Offset       *string
}

type ListParentInvoicesResult struct {
	Items  []domain.ParentInvoiceRow
	Limit  int
	Offset int
}

var parentValidStatuses = map[string]bool{
	domain.InvoiceStatusIssued:        true,
	domain.InvoiceStatusPaymentFailed: true,
	domain.InvoiceStatusPaid:          true,
	domain.InvoiceStatusOverdue:       true,
}

func (uc *ListParentInvoices) Execute(ctx context.Context, actor tenant.ActorContext, params ListParentInvoicesParams) (ListParentInvoicesResult, error) {
	filters := domain.ParentInvoiceFilters{
		Limit:  50,
		Offset: 0,
	}

	if params.BillingMonth != nil {
		bm, err := ParseBillingMonth(*params.BillingMonth)
		if err != nil {
			return ListParentInvoicesResult{}, domainerrors.Validation("Invalid billing_month format. Use YYYY-MM.", "billing_month")
		}
		filters.BillingMonth = &bm
	}

	if params.Status != nil {
		s := strings.TrimSpace(*params.Status)
		if s == domain.InvoiceStatusDraft {
			return ListParentInvoicesResult{}, domainerrors.Validation("Invalid status filter: draft.", "status")
		}
		if !parentValidStatuses[s] {
			return ListParentInvoicesResult{}, domainerrors.Validation(fmt.Sprintf("Invalid status filter: %s.", s), "status")
		}
		filters.Status = &s
	}

	if params.ChildID != nil {
		cid, err := uuid.Parse(strings.TrimSpace(*params.ChildID))
		if err != nil {
			return ListParentInvoicesResult{}, domainerrors.Validation("Invalid child_id format.", "child_id")
		}
		filters.ChildID = &cid
	}

	if params.Limit != nil {
		l, err := strconv.Atoi(*params.Limit)
		if err != nil || l < 1 || l > 200 {
			return ListParentInvoicesResult{}, domainerrors.Validation("Limit must be between 1 and 200.", "limit")
		}
		filters.Limit = l
	}

	if params.Offset != nil {
		o, err := strconv.Atoi(*params.Offset)
		if err != nil || o < 0 {
			return ListParentInvoicesResult{}, domainerrors.Validation("Offset must be 0 or greater.", "offset")
		}
		filters.Offset = o
	}

	rows, err := uc.repo.ListInvoicesForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID, filters)
	if err != nil {
		return ListParentInvoicesResult{}, domainerrors.Internal(err)
	}

	if rows == nil {
		rows = []domain.ParentInvoiceRow{}
	}

	return ListParentInvoicesResult{
		Items:  rows,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}
