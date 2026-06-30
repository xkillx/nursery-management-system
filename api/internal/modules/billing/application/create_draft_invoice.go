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
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateDraftInvoice struct {
	repo   domain.BillingRepository
	txMgr  DraftInvoiceTxManager
	auditW *audit.Writer
}

type DraftInvoiceTxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

func NewCreateDraftInvoice(repo domain.BillingRepository, txMgr DraftInvoiceTxManager, auditW *audit.Writer) *CreateDraftInvoice {
	return &CreateDraftInvoice{repo: repo, txMgr: txMgr, auditW: auditW}
}

type CreateDraftInvoiceInput struct {
	ChildID       uuid.UUID
	BillingMonth  string
	Lines         []DraftInvoiceLineInput
	PaymentTerms  string
	InternalNotes string
}

type DraftInvoiceLineInput struct {
	LineKind        string
	Description     string
	SortOrder       int
	QuantityMinutes int
	UnitAmountMinor int
	LineAmountMinor int
}

type CreateDraftInvoiceResult struct {
	InvoiceID     uuid.UUID
	ChildID       uuid.UUID
	BillingMonth  string
	Status        string
	Lines         []DraftLineResult
	SubtotalMinor int
	TotalDueMinor int
	PaymentTerms  string
	InternalNotes string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DraftLineResult struct {
	LineID          uuid.UUID
	LineKind        string
	Description     string
	SortOrder       int
	QuantityMinutes int
	UnitAmountMinor int
	LineAmountMinor int
}

func (uc *CreateDraftInvoice) Execute(ctx context.Context, actor tenant.ActorContext, input CreateDraftInvoiceInput) (CreateDraftInvoiceResult, error) {
	billingMonth, err := ParseBillingMonth(input.BillingMonth)
	if err != nil {
		return CreateDraftInvoiceResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	if len(input.Lines) == 0 {
		return CreateDraftInvoiceResult{}, domainerrors.Validation("At least one line is required.", "lines")
	}

	for _, line := range input.Lines {
		if line.LineKind != domain.LineKindFundedDeduction && line.LineKind != "" {
			if line.QuantityMinutes < 0 {
				return CreateDraftInvoiceResult{}, domainerrors.Validation("Quantity must be non-negative.", "lines")
			}
			if line.UnitAmountMinor < 0 {
				return CreateDraftInvoiceResult{}, domainerrors.Validation("Unit price must be non-negative.", "lines")
			}
		}
	}

	var result CreateDraftInvoiceResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
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
		now := time.Now().UTC()

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

		lineResults := make([]DraftLineResult, 0, len(input.Lines))
		for _, line := range input.Lines {
			lineID := uid.NewUUID()
			unitAmount := domain.MustGBP(line.UnitAmountMinor)
			lineAmount := domain.MustGBP(line.LineAmountMinor)
			if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
				ID:              lineID,
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
			lineResults = append(lineResults, DraftLineResult{
				LineID:          lineID,
				LineKind:        line.LineKind,
				Description:     line.Description,
				SortOrder:       line.SortOrder,
				QuantityMinutes: line.QuantityMinutes,
				UnitAmountMinor: line.UnitAmountMinor,
				LineAmountMinor: line.LineAmountMinor,
			})
		}

		auditDetails := map[string]any{
			"billing_month":   input.BillingMonth,
			"child_id":        input.ChildID.String(),
			"line_count":      len(input.Lines),
			"subtotal_minor":  subtotalMinor,
			"total_due_minor": totalDueMinor,
		}
		if input.PaymentTerms != "" {
			auditDetails["payment_terms"] = input.PaymentTerms
		}
		if input.InternalNotes != "" {
			auditDetails["has_internal_notes"] = true
		}
		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invoice_draft_created",
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details:    auditDetails,
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		result = CreateDraftInvoiceResult{
			InvoiceID:     invoiceID,
			ChildID:       input.ChildID,
			BillingMonth:  input.BillingMonth,
			Status:        domain.InvoiceStatusDraft,
			Lines:         lineResults,
			SubtotalMinor: subtotalMinor,
			TotalDueMinor: totalDueMinor,
			PaymentTerms:  input.PaymentTerms,
			InternalNotes: input.InternalNotes,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return CreateDraftInvoiceResult{}, txErr
		}
		return CreateDraftInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}
