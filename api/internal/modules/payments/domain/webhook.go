package domain

import (
	"context"
	"time"
)

// Webhook event processing statuses
const (
	WebhookStatusReceived  = "received"
	WebhookStatusProcessed = "processed"
	WebhookStatusIgnored   = "ignored"
	WebhookStatusRejected  = "rejected"
)

// Reason codes
const (
	ReasonPaid                  = "paid"
	ReasonPaymentFailed         = "payment_failed"
	ReasonExpired               = "expired"
	ReasonAwaitingAsyncPayment  = "awaiting_async_payment"
	ReasonUnsupportedEventType  = "unsupported_event_type"
	ReasonMetadataMissing       = "metadata_missing"
	ReasonUnknownPaymentAttempt = "unknown_payment_attempt"
	ReasonInvoiceMismatch       = "invoice_mismatch"
	ReasonSessionMismatch       = "session_mismatch"
	ReasonAmountMismatch        = "amount_mismatch"
	ReasonCurrencyMismatch      = "currency_mismatch"
	ReasonAlreadyPaid           = "already_paid"
	ReasonAlreadyPaymentFailed  = "already_payment_failed"
)

// Audit constants for webhook-driven payment status updates
const (
	AuditActionInvoicePaymentStatusUpdated = "invoice_payment_status_updated"
	AuditEntityInvoice                     = "invoice"
)

// Checkout session event types
const (
	EventTypeCheckoutCompleted      = "checkout.session.completed"
	EventTypeCheckoutAsyncSucceeded = "checkout.session.async_payment_succeeded"
	EventTypeCheckoutAsyncFailed    = "checkout.session.async_payment_failed"
	EventTypeCheckoutExpired        = "checkout.session.expired"
)

var CheckoutMutatingEventTypes = map[string]bool{
	EventTypeCheckoutCompleted:      true,
	EventTypeCheckoutAsyncSucceeded: true,
	EventTypeCheckoutAsyncFailed:    true,
	EventTypeCheckoutExpired:        true,
}

type CheckoutSessionWebhookData struct {
	CheckoutSessionID string
	PaymentStatus     string
	AmountTotal       int64
	Currency          string
	PaymentIntentID   string
	Metadata          map[string]string
}

type StripeWebhookEvent struct {
	ID                string
	StripeEventID     string
	EventType         string
	Livemode          bool
	APIVersion        string
	ProviderCreatedAt *time.Time
	RawPayload        []byte
	CheckoutSession   *CheckoutSessionWebhookData
}

type WebhookProcessResult struct {
	Status string // processed, ignored, rejected, duplicate
}

type WebhookRepository interface {
	InsertWebhookEvent(ctx context.Context, tx Tx, event StripeWebhookEvent, requestID string, processingStatus, processingReason string) (string, bool, error)
	UpdateWebhookEventStatus(ctx context.Context, tx Tx, eventID string, status, reason, errorMsg string) error
	GetPaymentAttemptAndInvoiceForWebhook(ctx context.Context, tx Tx, tenantID, branchID, invoiceID, attemptID, sessionID string) (*WebhookAttemptInvoice, error)
	MarkPaymentAttemptPaid(ctx context.Context, tx Tx, tenantID, branchID, attemptID string) error
	MarkPaymentAttemptFailed(ctx context.Context, tx Tx, tenantID, branchID, attemptID string) error
	MarkPaymentAttemptExpired(ctx context.Context, tx Tx, tenantID, branchID, attemptID string) error
	MarkInvoicePaid(ctx context.Context, tx Tx, tenantID, branchID, invoiceID string) error
	MarkInvoicePaymentFailed(ctx context.Context, tx Tx, tenantID, branchID, invoiceID string) error
	InsertReconciliationRecord(ctx context.Context, tx Tx, params ReconciliationRecordParams) error
}

type WebhookAttemptInvoice struct {
	AttemptID              string
	AttemptStatus          string
	AttemptAmountMinor     int32
	AttemptCurrencyCode    string
	AttemptSessionID       string
	InvoiceID              string
	InvoiceStatus          string
	InvoiceTotalDueMinor   int32
	InvoiceAmountPaidMinor int32
	InvoiceCurrencyCode    string
	InvoicePaidAt          *time.Time
	InvoicePaymentFailedAt *time.Time
}

type ReconciliationRecordParams struct {
	ID                    string
	TenantID              string
	BranchID              string
	InvoiceID             string
	PaymentAttemptID      string
	WebhookEventID        string
	StripeEventID         string
	StripeEventType       string
	CheckoutSessionID     string
	PaymentIntentID       string
	Outcome               string
	ReasonCode            string
	PreviousInvoiceStatus string
	NewInvoiceStatus      string
	AttemptPreviousStatus string
	AttemptNewStatus      string
	AmountMinor           int32
	CurrencyCode          string
	Details               string
}

type WebhookVerifier interface {
	VerifyAndParse(ctx context.Context, payload []byte, signatureHeader string) (*StripeWebhookEvent, error)
}
