package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

const (
	RetryAvailable            = "available"
	RetryNotIssued            = "invoice_not_issued"
	RetryAlreadyPaid          = "invoice_already_paid"
	RetryZeroTotal            = "invoice_zero_total"
	RetryPartialPaid          = "invoice_partial_paid"
	RetryNotMonthly           = "invoice_not_monthly"
	RetryCurrencyNotSupported = "currency_not_supported"
)

type GetManagerPaymentStatus struct {
	repo domain.ManagerPaymentRepository
}

func NewGetManagerPaymentStatus(repo domain.ManagerPaymentRepository) *GetManagerPaymentStatus {
	return &GetManagerPaymentStatus{repo: repo}
}

type ManagerPaymentStatusResult struct {
	InvoiceID               string
	InvoiceKind             string
	InvoiceNumber           string
	InvoiceNumberDisplay    string
	ChildID                 string
	ChildFirstName          string
	ChildMiddleName         *string
	ChildLastName           *string
	BillingMonth            string
	Status                  string
	DueStatus               string
	CurrencyCode            string
	TotalDueMinor           int
	AmountPaidMinor         int
	IssuedAt                *string
	DueAt                   *string
	PaidAt                  *string
	PaymentFailedAt         *string
	PaymentStatusUpdatedAt  *string
	CheckoutRetryAvailable  bool
	CheckoutRetryReasonCode string
	LatestPaymentAttempt    *PaymentAttemptResult
	LatestPaymentEvent      *PaymentEventResult
}

type PaymentAttemptResult struct {
	PaymentAttemptID        string
	Status                  string
	AmountMinor             int
	CurrencyCode            string
	StripeCheckoutSessionID *string
	StripePaymentIntentID   *string
	StripeExpiresAt         *string
	FailureReason           *string
	ProviderErrorCode       *string
	ProviderErrorMessage    *string
	CreatedAt               string
	UpdatedAt               string
}

type PaymentEventResult struct {
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
	WebhookReceivedAt       *string
	WebhookProcessedAt      *string
	CreatedAt               string
}

func (uc *GetManagerPaymentStatus) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (ManagerPaymentStatusResult, error) {
	if _, err := uuid.Parse(invoiceIDRaw); err != nil {
		return ManagerPaymentStatusResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	invoice, found, err := uc.repo.GetManagerInvoicePaymentStatus(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw)
	if err != nil {
		return ManagerPaymentStatusResult{}, domainerrors.Internal(err)
	}
	if !found {
		return ManagerPaymentStatusResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
	}

	attempt, err := uc.repo.GetLatestPaymentAttemptForInvoice(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw)
	if err != nil {
		return ManagerPaymentStatusResult{}, domainerrors.Internal(err)
	}

	event, err := uc.repo.GetLatestPaymentEventForInvoice(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw)
	if err != nil {
		return ManagerPaymentStatusResult{}, domainerrors.Internal(err)
	}

	retryAvailable, retryReason := computeRetryAvailability(invoice)

	result := ManagerPaymentStatusResult{
		InvoiceID:               invoice.InvoiceID,
		InvoiceKind:             invoice.InvoiceKind,
		InvoiceNumber:           invoice.InvoiceNumber,
		InvoiceNumberDisplay:    invoice.InvoiceNumberDisplay,
		ChildID:                 invoice.ChildID,
		ChildFirstName:          invoice.ChildFirstName,
		ChildMiddleName:         invoice.ChildMiddleName,
		ChildLastName:           invoice.ChildLastName,
		BillingMonth:            invoice.BillingMonth,
		Status:                  invoice.Status,
		DueStatus:               computeDueStatus(invoice.Status),
		CurrencyCode:            invoice.CurrencyCode,
		TotalDueMinor:           invoice.TotalDueMinor,
		AmountPaidMinor:         invoice.AmountPaidMinor,
		CheckoutRetryAvailable:  retryAvailable,
		CheckoutRetryReasonCode: retryReason,
		IssuedAt:                formatTimePtr(invoice.IssuedAt),
		DueAt:                   formatTimePtr(invoice.DueAt),
		PaidAt:                  formatTimePtr(invoice.PaidAt),
		PaymentFailedAt:         formatTimePtr(invoice.PaymentFailedAt),
		PaymentStatusUpdatedAt:  formatTimePtr(invoice.PaymentStatusUpdatedAt),
	}

	if attempt != nil {
		result.LatestPaymentAttempt = &PaymentAttemptResult{
			PaymentAttemptID:        attempt.PaymentAttemptID,
			Status:                  attempt.Status,
			AmountMinor:             attempt.AmountMinor,
			CurrencyCode:            attempt.CurrencyCode,
			StripeCheckoutSessionID: attempt.StripeCheckoutSessionID,
			StripePaymentIntentID:   attempt.StripePaymentIntentID,
			StripeExpiresAt:         formatTimePtr(attempt.StripeExpiresAt),
			FailureReason:           attempt.FailureReason,
			ProviderErrorCode:       attempt.ProviderErrorCode,
			ProviderErrorMessage:    attempt.ProviderErrorMessage,
			CreatedAt:               formatTime(attempt.CreatedAt),
			UpdatedAt:               formatTime(attempt.UpdatedAt),
		}
	}

	if event != nil {
		result.LatestPaymentEvent = &PaymentEventResult{
			PaymentEventID:          event.PaymentEventID,
			PaymentAttemptID:        event.PaymentAttemptID,
			StripeEventID:           event.StripeEventID,
			StripeEventType:         event.StripeEventType,
			StripeCheckoutSessionID: event.StripeCheckoutSessionID,
			StripePaymentIntentID:   event.StripePaymentIntentID,
			Outcome:                 event.Outcome,
			ReasonCode:              event.ReasonCode,
			PreviousInvoiceStatus:   event.PreviousInvoiceStatus,
			NewInvoiceStatus:        event.NewInvoiceStatus,
			AttemptPreviousStatus:   event.AttemptPreviousStatus,
			AttemptNewStatus:        event.AttemptNewStatus,
			AmountMinor:             event.AmountMinor,
			CurrencyCode:            event.CurrencyCode,
			WebhookProcessingStatus: event.WebhookProcessingStatus,
			WebhookProcessingReason: event.WebhookProcessingReason,
			WebhookReceivedAt:       formatTimePtr(event.WebhookReceivedAt),
			WebhookProcessedAt:      formatTimePtr(event.WebhookProcessedAt),
			CreatedAt:               formatTime(event.CreatedAt),
		}
	}

	return result, nil
}

func computeDueStatus(status string) string {
	switch status {
	case "draft":
		return "not_due"
	case "issued", "payment_failed":
		return "due"
	case "overdue":
		return "overdue"
	case "paid":
		return "paid"
	default:
		return "not_due"
	}
}

func computeRetryAvailability(invoice domain.ManagerInvoicePaymentStatus) (bool, string) {
	if invoice.InvoiceKind != "monthly" {
		return false, RetryNotMonthly
	}
	if invoice.Status == "paid" {
		return false, RetryAlreadyPaid
	}
	payableStatuses := map[string]bool{
		"issued":         true,
		"payment_failed": true,
		"overdue":        true,
	}
	if !payableStatuses[invoice.Status] {
		return false, RetryNotIssued
	}
	if invoice.TotalDueMinor <= 0 {
		return false, RetryZeroTotal
	}
	if invoice.AmountPaidMinor > 0 {
		return false, RetryPartialPaid
	}
	if invoice.CurrencyCode != "GBP" {
		return false, RetryCurrencyNotSupported
	}
	return true, RetryAvailable
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
