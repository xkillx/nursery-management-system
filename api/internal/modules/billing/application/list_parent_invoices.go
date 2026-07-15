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
	BillingMonth     *string
	BillingMonthFrom *string
	BillingMonthTo   *string
	Statuses         *string
	ChildID          *string
	Limit            *string
	Offset           *string
}

type ListParentInvoicesResult struct {
	Items  []domain.ParentInvoiceRow
	Total  int
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

	if params.BillingMonthFrom != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthFrom)
		if err != nil {
			return ListParentInvoicesResult{}, domainerrors.Validation("Invalid billing_month_from format. Use YYYY-MM.", "billing_month_from")
		}
		filters.BillingMonthFrom = &bm
	}

	if params.BillingMonthTo != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthTo)
		if err != nil {
			return ListParentInvoicesResult{}, domainerrors.Validation("Invalid billing_month_to format. Use YYYY-MM.", "billing_month_to")
		}
		filters.BillingMonthTo = &bm
	}

	if params.Statuses != nil {
		raw := strings.TrimSpace(*params.Statuses)
		if raw != "" {
			parts := strings.Split(raw, ",")
			statuses := make([]string, 0, len(parts))
			for _, p := range parts {
				s := strings.TrimSpace(p)
				if s == "" {
					continue
				}
				if s == domain.InvoiceStatusDraft {
					return ListParentInvoicesResult{}, domainerrors.Validation("Invalid status filter: draft.", "status")
				}
				if !parentValidStatuses[s] {
					return ListParentInvoicesResult{}, domainerrors.Validation(fmt.Sprintf("Invalid status filter: %s.", s), "status")
				}
				statuses = append(statuses, s)
			}
			if len(statuses) > 0 {
				filters.Statuses = statuses
			}
		}
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

	total, err := uc.repo.CountInvoicesForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID, filters)
	if err != nil {
		return ListParentInvoicesResult{}, domainerrors.Internal(err)
	}

	if rows == nil {
		rows = []domain.ParentInvoiceRow{}
	}

	return ListParentInvoicesResult{
		Items:  rows,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}
