package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// InvoiceResendSender enqueues a manager-triggered invoice email inside the
// caller's transaction. It is a consumer-side interface so the billing
// application never imports the notifications module; bootstrap binds it to the
// billing notification adapter (KTD-1).
type InvoiceResendSender interface {
	SendInvoiceResendEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error
}

// SendInvoiceResult reports the outcome of a manager-triggered resend.
type SendInvoiceResult struct {
	Status string
}

// resendCooldown is the minimum window between two manual resends of the same
// invoice. Rapid duplicates are rejected with invoice_resend_throttled (R11).
const resendCooldown = 5 * time.Minute

// SendInvoiceTxManager runs the resend flow inside one transaction.
type SendInvoiceTxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

// SendInvoiceEmail validates a manager-triggered invoice resend and enqueues
// the email via the strict sender inside one transaction (KTD-4).
type SendInvoiceEmail struct {
	repo         domain.BillingRepository
	parentLookup ParentContactLookup
	sender       InvoiceResendSender
	txMgr        SendInvoiceTxManager
}

func NewSendInvoiceEmail(
	repo domain.BillingRepository,
	parentLookup ParentContactLookup,
	sender InvoiceResendSender,
	txMgr SendInvoiceTxManager,
) *SendInvoiceEmail {
	return &SendInvoiceEmail{repo: repo, parentLookup: parentLookup, sender: sender, txMgr: txMgr}
}

func (uc *SendInvoiceEmail) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (SendInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return SendInvoiceResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	var result SendInvoiceResult

	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		locked, err := uc.repo.LockInvoiceForResendTx(ctx, tx, actor.TenantID, actor.BranchID, invoiceID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("lock invoice for resend: %w", err))
		}
		if !locked {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}

		invoice, found, err := uc.repo.GetInvoiceForManagerReviewTx(ctx, tx, actor.TenantID, actor.BranchID, invoiceID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("get invoice for resend: %w", err))
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}

		if !resendableStatus(invoice.Status) {
			return domainerrors.Conflict("invoice_not_payable", "Only issued, overdue, or payment-failed invoices can be emailed to the parent.")
		}

		recent, err := uc.repo.CountRecentInvoiceResendsTx(ctx, tx, actor.TenantID, actor.BranchID, invoiceID, time.Now().UTC().Add(-resendCooldown))
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("check resend cooldown: %w", err))
		}
		if recent > 0 {
			return domainerrors.Conflict("invoice_resend_throttled", "This invoice was emailed recently. Please wait a moment before sending again.")
		}

		parent, err := uc.parentLookup.GetForInvoice(ctx, actor.TenantID, actor.BranchID, invoice.ChildID)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("look up parent contact: %w", err))
		}
		if parent == nil || parent.Email == "" {
			return domainerrors.New("parent_no_email", "The parent has no email address on file.", "parent_email")
		}

		if err := uc.sender.SendInvoiceResendEmail(ctx, tx, invoiceID, actor.TenantID, actor.BranchID); err != nil {
			return err
		}

		result = SendInvoiceResult{Status: "queued"}
		return nil
	})

	if txErr != nil {
		if _, ok := txErr.(*domainerrors.DomainError); ok {
			return SendInvoiceResult{}, txErr
		}
		return SendInvoiceResult{}, domainerrors.Internal(txErr)
	}

	return result, nil
}

func resendableStatus(status string) bool {
	switch status {
	case domain.InvoiceStatusIssued, domain.InvoiceStatusOverdue, domain.InvoiceStatusPaymentFailed:
		return true
	}
	return false
}
