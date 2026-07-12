package domain

import "context"

const (
	PaymentLinkStatusActive      = "active"
	PaymentLinkStatusDeactivated = "deactivated"
)

type PaymentLinkCreateParams struct {
	AmountMinor   int
	Currency      string
	ProductName   string
	Description   string
	TenantID      string
	BranchID      string
	InvoiceID     string
	InvoiceNumber string
}

type PaymentLinkResult struct {
	ID  string
	URL string
}

type PaymentLinkProvider interface {
	CreatePaymentLink(ctx context.Context, params PaymentLinkCreateParams) (PaymentLinkResult, error)
}

type PaymentLinkRecord struct {
	ID                    string
	TenantID              string
	BranchID              string
	InvoiceID             string
	StripePaymentLinkID   string
	StripePaymentLinkURL  string
	AmountMinor           int
	CurrencyCode          string
	CreatedByUserID       string
	CreatedByMembershipID string
	Status                string
}

type PaymentLinkRepository interface {
	CreatePaymentLink(ctx context.Context, params PaymentLinkRecord) error
	GetActivePaymentLinkForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*PaymentLinkRecord, bool, error)
}
