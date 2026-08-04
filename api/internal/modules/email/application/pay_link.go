package application

import (
	"context"

	"github.com/google/uuid"
)

// InvoicePayLinkProvider resolves a Stripe Checkout URL for an invoice email at
// send time. Defined consumer-side in the email module (KTD2) so the email
// application never imports the payments module. The concrete implementation is
// wired in bootstrap. OK=false means "no pay link available — use the app link".
type InvoicePayLinkProvider interface {
	CreateEmailCheckoutSession(ctx context.Context, tenantID, branchID uuid.UUID, invoiceID, requestID string) (url string, ok bool, err error)
}

// invoicePayLinkTemplates are the outbox template names whose emails carry a
// "Pay Now" Stripe link. Receipt and all non-invoice templates never call the
// provider (R2).
var invoicePayLinkTemplates = map[string]bool{
	"issued":       true,
	"due-soon":     true,
	"due-reminder": true,
	"overdue":      true,
}
