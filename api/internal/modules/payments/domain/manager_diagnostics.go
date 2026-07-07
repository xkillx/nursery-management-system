package domain

import (
	"context"
	"time"
)

type ManagerInvoicePaymentStatus struct {
	InvoiceID              string
	InvoiceKind            string
	InvoiceNumber          string
	InvoiceNumberDisplay   string
	ChildID                string
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	BillingMonth           string
	Status                 string
	DueStatus              string
	CurrencyCode           string
	TotalDueMinor          int
	AmountPaidMinor        int
	IssuedAt               *time.Time
	DueAt                  *time.Time
	PaidAt                 *time.Time
	PaymentFailedAt        *time.Time
	PaymentStatusUpdatedAt *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type PaymentAttemptDiagnostic struct {
	PaymentAttemptID        string
	Status                  string
	AmountMinor             int
	CurrencyCode            string
	StripeCheckoutSessionID *string
	StripePaymentIntentID   *string
	StripeExpiresAt         *time.Time
	FailureReason           *string
	ProviderErrorCode       *string
	ProviderErrorMessage    *string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type PaymentEventDiagnostic struct {
	PaymentEventID          string
	PaymentAttemptID        string
	StripeEventID           string
	StripeEventType         string
	StripeCheckoutSessionID string
	StripePaymentIntentID   string
	Outcome                 string
	ReasonCode              string
	PreviousInvoiceStatus   string
	NewInvoiceStatus        string
	AttemptPreviousStatus   string
	AttemptNewStatus        string
	AmountMinor             int
	CurrencyCode            string
	WebhookProcessingStatus string
	WebhookProcessingReason string
	WebhookReceivedAt       *time.Time
	WebhookProcessedAt      *time.Time
	CreatedAt               time.Time
}

type PaymentEventFilters struct {
	Limit  int
	Offset int
}

type ManagerPaymentRepository interface {
	GetManagerInvoicePaymentStatus(ctx context.Context, tenantID, branchID, invoiceID string) (ManagerInvoicePaymentStatus, bool, error)
	GetLatestPaymentAttemptForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*PaymentAttemptDiagnostic, error)
	GetLatestPaymentEventForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*PaymentEventDiagnostic, error)
	ListPaymentEventsForInvoice(ctx context.Context, tenantID, branchID, invoiceID string, filters PaymentEventFilters) ([]PaymentEventDiagnostic, error)
	CountPaymentEventsForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (int, error)
}
