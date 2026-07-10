package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
)

type VoidInvoice struct {
	repo       domain.BillingRepository
	txMgr      *transaction.Manager
	auditW     *audit.Writer
	dispatcher *events.EventDispatcher
}

func NewVoidInvoice(
	repo domain.BillingRepository,
	txMgr *transaction.Manager,
	auditW *audit.Writer,
	dispatcher *events.EventDispatcher,
) *VoidInvoice {
	return &VoidInvoice{repo: repo, txMgr: txMgr, auditW: auditW, dispatcher: dispatcher}
}

type VoidInvoiceResult struct {
	InvoiceID  uuid.UUID
	Status     string
	VoidedAt   time.Time
	VoidReason string
}

func (uc *VoidInvoice) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, reason string) (VoidInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return VoidInvoiceResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		return VoidInvoiceResult{}, domainerrors.Validation("Void reason must not be empty.", "reason")
	}

	var result VoidInvoiceResult

	txErr := uc.dispatcher.DispatchInTx(ctx, func(tx pgx.Tx, emitter events.Emitter) error {
		voidedAt := time.Now().UTC()

		n, markErr := uc.repo.MarkInvoiceVoid(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, reason, voidedAt)
		if markErr != nil {
			return fmt.Errorf("mark invoice void: %w", markErr)
		}
		if n == 0 {
			return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be voided.")
		}

		if auditErr := uc.auditW.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: domain.AuditInvoiceVoided,
			EntityType: "invoice",
			EntityID:   invoiceID,
			Details: map[string]any{
				"reason": reason,
			},
		}); auditErr != nil {
			return fmt.Errorf("write audit: %w", auditErr)
		}

		emitter.Emit(domain.InvoiceVoided{
			InvoiceID: invoiceID,
			Reason:    reason,
			Occurred:  voidedAt,
		})

		result = VoidInvoiceResult{
			InvoiceID:  invoiceID,
			Status:     domain.InvoiceStatusVoid,
			VoidedAt:   voidedAt,
			VoidReason: reason,
		}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return VoidInvoiceResult{}, txErr
		}
		return VoidInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}
