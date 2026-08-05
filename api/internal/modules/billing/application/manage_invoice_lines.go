package application

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type ManageInvoiceLines struct {
	repo   domain.BillingRepository
	txMgr  DraftInvoiceTxManager
	auditW LineAuditWriter
}

// LineAuditWriter is the audit surface ManageInvoiceLines needs. Defined in
// the consumer so tests can inject a stub without touching the audit module.
type LineAuditWriter interface {
	WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error
}

func NewManageInvoiceLines(
	repo domain.BillingRepository,
	txMgr DraftInvoiceTxManager,
	auditW LineAuditWriter,
) *ManageInvoiceLines {
	return &ManageInvoiceLines{repo: repo, txMgr: txMgr, auditW: auditW}
}

type InvoiceLineResult struct {
	LineID          uuid.UUID
	LineKind        string
	Description     string
	SortOrder       int
	QuantityMinutes int
	UnitAmountMinor int
	LineAmountMinor int
	SubtotalMinor   int
	TotalDueMinor   int
}

type DeleteLineResult struct {
	LineID        uuid.UUID
	SubtotalMinor int
	TotalDueMinor int
}

type AddLineInput struct {
	LineKind        string
	Description     string
	QuantityMinutes int
	UnitAmountMinor int
	LineAmountMinor int
}

type UpdateLineInput struct {
	Description     string
	QuantityMinutes int
	UnitAmountMinor int
	LineAmountMinor int
}

func (uc *ManageInvoiceLines) AddLine(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, input AddLineInput) (InvoiceLineResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return InvoiceLineResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	if input.LineKind != domain.LineKindExtra && input.LineKind != domain.LineKindAdHoc {
		return InvoiceLineResult{}, domainerrors.Conflict("invoice_line_kind_immutable", "Only 'extra' and 'ad_hoc' line kinds can be added.")
	}

	description, descErr := validateLineDescription(input.Description)
	if descErr != nil {
		return InvoiceLineResult{}, descErr
	}
	input.Description = description

	var result InvoiceLineResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		inv, found, getErr := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
		if getErr != nil {
			return fmt.Errorf("get invoice: %w", getErr)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}
		if inv.Status != domain.InvoiceStatusDraft {
			return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be modified.")
		}

		lineID := uid.NewUUID()
		unitAmount := domain.MustGBP(input.UnitAmountMinor)
		lineAmount := domain.MustGBP(input.LineAmountMinor)

		nextSortOrder, sortErr := uc.nextSortOrder(ctx, actor.TenantID, actor.BranchID, invoiceID)
		if sortErr != nil {
			return fmt.Errorf("determine sort order: %w", sortErr)
		}

		if insErr := uc.repo.InsertInvoiceLine(ctx, tx, domain.InvoiceLineCreateParams{
			ID:              lineID,
			TenantID:        actor.TenantID,
			BranchID:        actor.BranchID,
			InvoiceID:       invoiceID,
			LineKind:        input.LineKind,
			Description:     input.Description,
			SortOrder:       nextSortOrder,
			QuantityMinutes: input.QuantityMinutes,
			UnitAmount:      unitAmount,
			LineAmount:      lineAmount,
		}); insErr != nil {
			return fmt.Errorf("insert invoice line: %w", insErr)
		}

		subtotal, totalDue, recalcErr := uc.recalculateTotal(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, inv)
		if recalcErr != nil {
			return fmt.Errorf("recalculate total: %w", recalcErr)
		}

		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invoice_line_added",
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details: map[string]any{
				"line_id":   lineID.String(),
				"line_kind": input.LineKind,
				"amount":    input.LineAmountMinor,
				"subtotal":  subtotal,
				"total_due": totalDue,
			},
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		result = InvoiceLineResult{
			LineID:          lineID,
			LineKind:        input.LineKind,
			Description:     input.Description,
			SortOrder:       nextSortOrder,
			QuantityMinutes: input.QuantityMinutes,
			UnitAmountMinor: input.UnitAmountMinor,
			LineAmountMinor: input.LineAmountMinor,
			SubtotalMinor:   subtotal,
			TotalDueMinor:   totalDue,
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return InvoiceLineResult{}, txErr
		}
		return InvoiceLineResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func (uc *ManageInvoiceLines) UpdateLine(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw, lineIDRaw string, input UpdateLineInput) (InvoiceLineResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return InvoiceLineResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}
	lineID, err := uuid.Parse(lineIDRaw)
	if err != nil {
		return InvoiceLineResult{}, domainerrors.Validation("Invalid line ID format.", "line_id")
	}

	description, descErr := validateLineDescription(input.Description)
	if descErr != nil {
		return InvoiceLineResult{}, descErr
	}

	var result InvoiceLineResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		inv, found, getErr := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
		if getErr != nil {
			return fmt.Errorf("get invoice: %w", getErr)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}
		if inv.Status != domain.InvoiceStatusDraft {
			return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be modified.")
		}

		line, lineFound, lineErr := uc.repo.GetInvoiceLine(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, lineID)
		if lineErr != nil {
			return fmt.Errorf("get invoice line: %w", lineErr)
		}
		if !lineFound {
			return domainerrors.NotFound("invoice_line", "Invoice line not found.")
		}

		isSystemLine := line.LineKind != domain.LineKindExtra && line.LineKind != domain.LineKindAdHoc

		// System lines (core childcare, funded deduction, hourly) are
		// description-editable only; any quantity/unit/amount change is
		// rejected server-side (KTD4). Extra/ad-hoc lines keep full value
		// editability.
		if isSystemLine {
			if input.QuantityMinutes != line.QuantityMinutes ||
				input.UnitAmountMinor != line.UnitAmount.Minor() ||
				input.LineAmountMinor != line.LineAmount.Minor() {
				return domainerrors.Conflict("invoice_line_values_immutable", "Only the description of this line can be updated.")
			}
		}

		descriptionChanged := description != line.Description

		// Persist the override marker whenever the description changes, so a
		// later regeneration preserves the human-renamed label (KTD2/KTD6).
		details := line.Details
		if descriptionChanged {
			merged, mergeErr := domain.SetLineDescriptionOverride(line.Details)
			if mergeErr != nil {
				return fmt.Errorf("merge description override: %w", mergeErr)
			}
			details = merged
		}

		quantityMinutes := line.QuantityMinutes
		unitAmount := line.UnitAmount
		lineAmount := line.LineAmount
		if !isSystemLine {
			quantityMinutes = input.QuantityMinutes
			unitAmount = domain.MustGBP(input.UnitAmountMinor)
			lineAmount = domain.MustGBP(input.LineAmountMinor)
		}

		n, updErr := uc.repo.UpdateInvoiceLine(ctx, tx, actor.TenantID, actor.BranchID, lineID, description, quantityMinutes, unitAmount, lineAmount, details)
		if updErr != nil {
			return fmt.Errorf("update invoice line: %w", updErr)
		}
		if n == 0 {
			return domainerrors.Conflict("invoice_line_kind_immutable", "Line could not be updated — it may not be an editable type.")
		}

		subtotal, totalDue, recalcErr := uc.recalculateTotal(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, inv)
		if recalcErr != nil {
			return fmt.Errorf("recalculate total: %w", recalcErr)
		}

		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invoice_line_updated",
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details: map[string]any{
				"line_id":   lineID.String(),
				"line_kind": line.LineKind,
				"amount":    lineAmount.Minor(),
				"subtotal":  subtotal,
				"total_due": totalDue,
			},
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		result = InvoiceLineResult{
			LineID:          lineID,
			LineKind:        line.LineKind,
			Description:     description,
			SortOrder:       line.SortOrder,
			QuantityMinutes: quantityMinutes,
			UnitAmountMinor: unitAmount.Minor(),
			LineAmountMinor: lineAmount.Minor(),
			SubtotalMinor:   subtotal,
			TotalDueMinor:   totalDue,
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return InvoiceLineResult{}, txErr
		}
		return InvoiceLineResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func (uc *ManageInvoiceLines) DeleteLine(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw, lineIDRaw string) (DeleteLineResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return DeleteLineResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}
	lineID, err := uuid.Parse(lineIDRaw)
	if err != nil {
		return DeleteLineResult{}, domainerrors.Validation("Invalid line ID format.", "line_id")
	}

	var result DeleteLineResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		inv, found, getErr := uc.repo.GetInvoiceForManagerReview(ctx, actor.TenantID, actor.BranchID, invoiceID)
		if getErr != nil {
			return fmt.Errorf("get invoice: %w", getErr)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}
		if inv.Status != domain.InvoiceStatusDraft {
			return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be modified.")
		}

		line, lineFound, lineErr := uc.repo.GetInvoiceLine(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, lineID)
		if lineErr != nil {
			return fmt.Errorf("get invoice line: %w", lineErr)
		}
		if !lineFound {
			return domainerrors.NotFound("invoice_line", "Invoice line not found.")
		}
		if line.LineKind != domain.LineKindExtra && line.LineKind != domain.LineKindAdHoc {
			return domainerrors.Conflict("invoice_line_kind_immutable", "Only 'extra' and 'ad_hoc' lines can be deleted.")
		}

		n, delErr := uc.repo.DeleteInvoiceLine(ctx, tx, actor.TenantID, actor.BranchID, lineID)
		if delErr != nil {
			return fmt.Errorf("delete invoice line: %w", delErr)
		}
		if n == 0 {
			return domainerrors.Conflict("invoice_line_kind_immutable", "Line could not be deleted — it may not be an editable type.")
		}

		subtotal, totalDue, recalcErr := uc.recalculateTotal(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, inv)
		if recalcErr != nil {
			return fmt.Errorf("recalculate total: %w", recalcErr)
		}

		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "invoice_line_deleted",
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceID,
			Details: map[string]any{
				"line_id":   lineID.String(),
				"line_kind": line.LineKind,
				"subtotal":  subtotal,
				"total_due": totalDue,
			},
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		result = DeleteLineResult{
			LineID:        lineID,
			SubtotalMinor: subtotal,
			TotalDueMinor: totalDue,
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return DeleteLineResult{}, txErr
		}
		return DeleteLineResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func (uc *ManageInvoiceLines) recalculateTotal(ctx context.Context, tx pgx.Tx, tenantID, branchID, invoiceID uuid.UUID, inv domain.InvoiceReviewRow) (subtotal, totalDue int, err error) {
	lines, lineErr := uc.repo.ListInvoiceLinesForManagerReview(ctx, tenantID, branchID, invoiceID)
	if lineErr != nil {
		return 0, 0, fmt.Errorf("list lines for recalculation: %w", lineErr)
	}

	newSubtotal := 0
	for _, l := range lines {
		newSubtotal += l.LineAmount.Minor()
	}

	fundedDeduction := inv.FundedDeduction.Minor()
	newTotalDue := newSubtotal - fundedDeduction
	if newTotalDue < 0 {
		newTotalDue = 0
	}

	var generatedRunID uuid.UUID
	if inv.GeneratedRunID != nil {
		generatedRunID = *inv.GeneratedRunID
	}

	if updErr := uc.repo.UpdateDraftInvoice(ctx, tx, domain.DraftInvoiceUpdateParams{
		ID:              invoiceID,
		TenantID:        tenantID,
		BranchID:        branchID,
		GeneratedRunID:  generatedRunID,
		Subtotal:        domain.MustGBP(newSubtotal),
		FundedDeduction: domain.MustGBP(fundedDeduction),
		TotalDue:        domain.MustGBP(newTotalDue),
	}); updErr != nil {
		return 0, 0, fmt.Errorf("update draft invoice totals: %w", updErr)
	}

	return newSubtotal, newTotalDue, nil
}

func (uc *ManageInvoiceLines) nextSortOrder(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) (int, error) {
	lines, err := uc.repo.ListInvoiceLinesForManagerReview(ctx, tenantID, branchID, invoiceID)
	if err != nil {
		return 0, err
	}
	maxOrder := 0
	for _, l := range lines {
		if l.SortOrder > maxOrder {
			maxOrder = l.SortOrder
		}
	}
	return maxOrder + 1, nil
}

// maxLineDescriptionRunes is the maximum length (in runes) of a line
// description accepted anywhere a description is entered (R9).
const maxLineDescriptionRunes = 120

// validateLineDescription trims and validates a line description: non-empty
// after trimming, and at most 120 runes. It returns the trimmed value so the
// caller persists the canonical form.
func validateLineDescription(raw string) (string, *domainerrors.DomainError) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", domainerrors.Validation("Description must not be empty.", "description")
	}
	if utf8.RuneCountInString(trimmed) > maxLineDescriptionRunes {
		return "", domainerrors.Validation("Description must be at most 120 characters.", "description")
	}
	return trimmed, nil
}
