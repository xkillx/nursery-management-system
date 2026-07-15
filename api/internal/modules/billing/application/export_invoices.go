package application

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

const maxExportMonths = 24

type ExportInvoices struct {
	repo domain.BillingRepository
}

func NewExportInvoices(repo domain.BillingRepository) *ExportInvoices {
	return &ExportInvoices{repo: repo}
}

type ExportInvoicesParams struct {
	BillingMonthFrom *string
	BillingMonthTo   *string
	Statuses         *string
	Format           string
}

func (uc *ExportInvoices) Execute(ctx context.Context, actor tenant.ActorContext, w io.Writer, params ExportInvoicesParams) error {
	filters := domain.InvoiceExportFilters{}

	if params.BillingMonthFrom != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthFrom)
		if err != nil {
			return domainerrors.Validation("Invalid billing_month_from format. Use YYYY-MM.", "billing_month_from")
		}
		filters.BillingMonthFrom = &bm
	}

	if params.BillingMonthTo != nil {
		bm, err := ParseBillingMonth(*params.BillingMonthTo)
		if err != nil {
			return domainerrors.Validation("Invalid billing_month_to format. Use YYYY-MM.", "billing_month_to")
		}
		filters.BillingMonthTo = &bm
	}

	if filters.BillingMonthFrom != nil && filters.BillingMonthTo != nil {
		months := monthsBetween(*filters.BillingMonthFrom, *filters.BillingMonthTo)
		if months > maxExportMonths {
			return domainerrors.Validation(fmt.Sprintf("Date range exceeds maximum of %d months.", maxExportMonths), "billing_month_from")
		}
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
				if !validStatuses[s] {
					return domainerrors.Validation(fmt.Sprintf("Invalid status filter: %s.", s), "status")
				}
				statuses = append(statuses, s)
			}
			if len(statuses) > 0 {
				filters.Statuses = statuses
			}
		}
	}

	format := strings.TrimSpace(params.Format)
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "csv-detail" {
		return domainerrors.Validation("Format must be 'csv' or 'csv-detail'.", "format")
	}

	if format == "csv" {
		return uc.executeCSV(ctx, actor, w, filters)
	}
	return uc.executeCSVDetail(ctx, actor, w, filters)
}

func (uc *ExportInvoices) executeCSV(ctx context.Context, actor tenant.ActorContext, w io.Writer, filters domain.InvoiceExportFilters) error {
	rows, err := uc.repo.ExportInvoicesForManagerReview(ctx, actor.TenantID, actor.BranchID, filters)
	if err != nil {
		return domainerrors.Internal(err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"invoice_number", "child_name", "billing_month", "status",
		"subtotal_minor", "funded_deduction_minor", "total_due_minor",
		"issued_at", "due_date", "paid_at",
	}
	if err := writer.Write(header); err != nil {
		return domainerrors.Internal(err)
	}

	for _, row := range rows {
		childName := row.ChildFirstName
		if row.ChildLastName != nil {
			childName += " " + *row.ChildLastName
		}
		record := []string{
			derefStr(row.InvoiceNumber),
			childName,
			row.BillingMonth.Format("2006-01"),
			row.Status,
			fmt.Sprintf("%d", row.Subtotal.Minor()),
			fmt.Sprintf("%d", row.FundedDeduction.Minor()),
			fmt.Sprintf("%d", row.TotalDue.Minor()),
			formatTimePtr(row.IssuedAt),
			formatTimePtr(row.DueAt),
			formatTimePtr(row.PaidAt),
		}
		if err := writer.Write(record); err != nil {
			return domainerrors.Internal(err)
		}
	}

	return nil
}

func (uc *ExportInvoices) executeCSVDetail(ctx context.Context, actor tenant.ActorContext, w io.Writer, filters domain.InvoiceExportFilters) error {
	rows, err := uc.repo.ExportInvoiceDetailsForManagerReview(ctx, actor.TenantID, actor.BranchID, filters)
	if err != nil {
		return domainerrors.Internal(err)
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"invoice_number", "child_name", "billing_month", "status",
		"line_kind", "description", "quantity_minutes", "unit_amount_minor", "line_amount_minor",
	}
	if err := writer.Write(header); err != nil {
		return domainerrors.Internal(err)
	}

	for _, row := range rows {
		childName := row.ChildFirstName
		if row.ChildLastName != nil {
			childName += " " + *row.ChildLastName
		}
		record := []string{
			derefStr(row.InvoiceNumber),
			childName,
			row.BillingMonth.Format("2006-01"),
			row.Status,
			row.LineKind,
			row.Description,
			derefInt(row.QuantityMinutes),
			derefInt(row.UnitAmountMinor),
			fmt.Sprintf("%d", row.LineAmountMinor),
		}
		if err := writer.Write(record); err != nil {
			return domainerrors.Internal(err)
		}
	}

	return nil
}

func monthsBetween(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	return years*12 + months + 1
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
