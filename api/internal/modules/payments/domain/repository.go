package domain

import (
	"context"
)

type Tx = any

type PaymentRepository interface {
	GetParentInvoiceForCheckoutForUpdate(ctx context.Context, tx Tx, tenantID, branchID, membershipID, invoiceID string) (CheckoutInvoiceCandidate, bool, error)
	CreatePaymentAttempt(ctx context.Context, tx Tx, params PaymentAttemptCreateParams) error
	GetInvoicePaymentState(ctx context.Context, tenantID, branchID, invoiceID string) (InvoicePaymentState, bool, error)
	MarkPaymentAttemptCheckoutCreated(ctx context.Context, params PaymentAttemptCheckoutCreatedParams) error
	MarkPaymentAttemptCheckoutCreationFailed(ctx context.Context, params PaymentAttemptCheckoutCreationFailedParams) error
	GetActiveCheckoutForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*ActiveCheckoutSession, bool, error)
}

// ActiveCheckoutSession represents an existing active checkout session for an invoice.
type ActiveCheckoutSession struct {
	AttemptID         string
	CheckoutSessionID string
	CheckoutURL       string
	PaymentIntentID   string
	AmountMinor       int
	CurrencyCode      string
}
