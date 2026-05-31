package domain

import "time"

const (
	AttemptStatusCheckoutCreationStarted = "checkout_creation_started"
	AttemptStatusCheckoutCreated         = "checkout_created"
	AttemptStatusCheckoutCreationFailed  = "checkout_creation_failed"
	AttemptStatusPaid                    = "paid"
	AttemptStatusPaymentFailed           = "payment_failed"
	AttemptStatusCancelled               = "cancelled"
	AttemptStatusExpired                 = "expired"
)

const (
	CurrencyGBP = "GBP"
)

const (
	FailureReasonStripeError                 = "stripe_error"
	FailureReasonInvoiceNoLongerPayable      = "invoice_no_longer_payable"
	FailureReasonPaymentProviderUnconfigured = "payment_provider_unconfigured"
)

type CheckoutInvoiceCandidate struct {
	ID              string
	InvoiceKind     string
	InvoiceNumber   string
	Status          string
	CurrencyCode    string
	TotalDueMinor   int
	AmountPaidMinor int
	ChildID         string
}

type InvoicePaymentState struct {
	InvoiceKind     string
	Status          string
	CurrencyCode    string
	TotalDueMinor   int
	AmountPaidMinor int
}

type PaymentAttemptCreateParams struct {
	ID                      string
	TenantID                string
	BranchID                string
	InvoiceID               string
	InitiatedByUserID       string
	InitiatedByMembershipID string
	RequestID               string
	Status                  string
	AmountMinor             int
	CurrencyCode            string
}

type PaymentAttemptCheckoutCreatedParams struct {
	TenantID                string
	BranchID                string
	AttemptID               string
	StripeCheckoutSessionID string
	StripeCheckoutURL       string
	StripePaymentIntentID   string
	StripeExpiresAt         *time.Time
}

type PaymentAttemptCheckoutCreationFailedParams struct {
	TenantID             string
	BranchID             string
	AttemptID            string
	FailureReason        string
	ProviderErrorCode    string
	ProviderErrorMessage string
}
