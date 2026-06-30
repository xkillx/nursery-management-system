package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateAndIssueInvoiceFromForm struct {
	repo       domain.BillingRepository
	dispatcher *events.EventDispatcher
	auditW     *audit.Writer
	issueUC    *IssueInvoice
}

func NewCreateAndIssueInvoiceFromForm(
	repo domain.BillingRepository,
	dispatcher *events.EventDispatcher,
	auditW *audit.Writer,
	issueUC *IssueInvoice,
) *CreateAndIssueInvoiceFromForm {
	return &CreateAndIssueInvoiceFromForm{
		repo:       repo,
		dispatcher: dispatcher,
		auditW:     auditW,
		issueUC:    issueUC,
	}
}

type CreateAndIssueInvoiceInput struct {
	ChildID       uuid.UUID
	BillingMonth  string
	Lines         []DraftInvoiceLineInput
	PaymentTerms  string
	InternalNotes string
}

func (uc *CreateAndIssueInvoiceFromForm) Execute(ctx context.Context, actor tenant.ActorContext, input CreateAndIssueInvoiceInput) (domain.IssueInvoiceResult, error) {
	billingMonth, err := ParseBillingMonth(input.BillingMonth)
	if err != nil {
		return domain.IssueInvoiceResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	if len(input.Lines) == 0 {
		return domain.IssueInvoiceResult{}, domainerrors.Validation("At least one line is required.", "lines")
	}

	for _, line := range input.Lines {
		if line.LineKind != domain.LineKindFundedDeduction && line.LineKind != "" {
			if line.QuantityMinutes < 0 {
				return domain.IssueInvoiceResult{}, domainerrors.Validation("Quantity must be non-negative.", "lines")
			}
			if line.UnitAmountMinor < 0 {
				return domain.IssueInvoiceResult{}, domainerrors.Validation("Unit price must be non-negative.", "lines")
			}
		}
	}

	var result domain.IssueInvoiceResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		existingInvoice, found, findErr := uc.repo.GetMonthlyInvoiceForUpdate(ctx, tx, actor.TenantID, actor.BranchID, input.ChildID, billingMonth)
		if findErr != nil {
			return fmt.Errorf("check existing invoice: %w", findErr)
		}
		if found {
			return domainerrors.Conflict("duplicate_child_month", fmt.Sprintf(
				"A monthly invoice already exists for this child and billing month (status: %s, id: %s).",
				existingInvoice.Status, existingInvoice.ID.String(),
			))
		}

		invoiceID := uid.NewUUID()
		formRunID := uid.NewUUID()

		subtotalMinor := 0
		for _, line := range input.Lines {
			subtotalMinor += line.LineAmountMinor
		}

		fundedDeductionMinor := 0
		totalDueMinor := subtotalMinor

		periodStart := time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, -1)

		if createErr := uc.repo.CreateDraftInvoice(ctx, tx, domain.DraftInvoiceCreateParams{
			ID:                 invoiceID,
			TenantID:           actor.TenantID,
			BranchID:           actor.BranchID,
			ChildID:            input.ChildID,
			BillingMonth:       billingMonth,
			GeneratedRunID:     formRunID,
			CurrencyCode:       "GBP",
			Subtotal:           domain.MustGBP(subtotalMinor),
			FundedDeduction:    domain.MustGBP(fundedDeductionMinor),
			TotalDue:           domain.MustGBP(totalDueMinor),
			PeriodStartDate:    periodStart,
			PeriodEndDate:      periodEnd,
			CalculationDetails: nil,
		}); createErr != nil {
			return fmt.Errorf("create draft invoice: %w", createErr)
		}

		for _, line := range input.Lines {
			unitAmount := domain.MustGBP(line.UnitAmountMinor)
			lineAmount := domain.MustGBP(line.LineAmountMinor)
			if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
				ID:              uid.NewUUID(),
				TenantID:        actor.TenantID,
				BranchID:        actor.BranchID,
				InvoiceID:       invoiceID,
				LineKind:        line.LineKind,
				Description:     line.Description,
				SortOrder:       line.SortOrder,
				QuantityMinutes: line.QuantityMinutes,
				UnitAmount:      unitAmount,
				LineAmount:      lineAmount,
			}); insErr != nil {
				return fmt.Errorf("insert invoice line: %w", insErr)
			}
		}

		issueResult, issueErr := uc.issueUC.executeIssue(ctx, tx, emitter, actor, invoiceID, billingMonth, domain.MustGBP(totalDueMinor))
		if issueErr != nil {
			return issueErr
		}
		result = issueResult
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return domain.IssueInvoiceResult{}, txErr
		}
		return domain.IssueInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}
