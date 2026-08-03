package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	"nursery-management-system/api/internal/platform/tenant"
)

type IssueInvoiceWithCheckout struct {
	issueUC    *IssueInvoice
	checkoutUC *paymentsapp.CreateCheckoutSession
}

func NewIssueInvoiceWithCheckout(
	issueUC *IssueInvoice,
	checkoutUC *paymentsapp.CreateCheckoutSession,
) *IssueInvoiceWithCheckout {
	return &IssueInvoiceWithCheckout{
		issueUC:    issueUC,
		checkoutUC: checkoutUC,
	}
}

func (uc *IssueInvoiceWithCheckout) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, confirm bool) (domain.IssueInvoiceResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return domain.IssueInvoiceResult{}, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Pre-work: create Stripe checkout session (idempotent)
	var checkoutURL string
	checkoutResult, checkoutErr := uc.checkoutUC.Execute(ctx,
		actor.TenantID.String(),
		actor.BranchID.String(),
		actor.MembershipID.String(),
		actor.UserID.String(),
		invoiceIDRaw,
		actor.RequestID,
	)
	if checkoutErr != nil {
		slog.WarnContext(ctx, "checkout_creation_failed_in_orchestrator",
			"invoice_id", invoiceID,
			"error", checkoutErr,
		)
		// Best-effort: invoice will still be issued, scheduler catches it
	} else {
		checkoutURL = checkoutResult.CheckoutURL
	}

	return uc.issueUC.ExecuteWithContext(ctx, actor, invoiceIDRaw, confirm, IssueContext{
		CheckoutURL: checkoutURL,
	})
}
