package domain

import "context"

type CheckoutSessionCreateParams struct {
	PaymentAttemptID string
	InvoiceID        string
	InvoiceNumber    string
	AmountMinor      int
	Currency         string
	ProductName      string
	ProductDesc      string
	SuccessURL       string
	CancelURL        string
	TenantID         string
	BranchID         string
}

type CheckoutSessionResult struct {
	CheckoutSessionID string
	CheckoutURL       string
	PaymentIntentID   string
	ExpiresAt         string
}

type CheckoutProvider interface {
	CreateCheckoutSession(ctx context.Context, params CheckoutSessionCreateParams) (CheckoutSessionResult, error)
}

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx Tx) error) error
}
