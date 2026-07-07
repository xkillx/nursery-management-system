package application

import (
	"context"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListManagerPaymentEvents struct {
	repo domain.ManagerPaymentRepository
}

func NewListManagerPaymentEvents(repo domain.ManagerPaymentRepository) *ListManagerPaymentEvents {
	return &ListManagerPaymentEvents{repo: repo}
}

type ListPaymentEventsResult struct {
	Items    []PaymentEventResult `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

func (uc *ListManagerPaymentEvents) Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string, page, pageSize int) (ListPaymentEventsResult, error) {
	if _, err := uuid.Parse(invoiceIDRaw); err != nil {
		return ListPaymentEventsResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	offset := (page - 1) * pageSize

	_, found, err := uc.repo.GetManagerInvoicePaymentStatus(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw)
	if err != nil {
		return ListPaymentEventsResult{}, domainerrors.Internal(err)
	}
	if !found {
		return ListPaymentEventsResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
	}

	events, err := uc.repo.ListPaymentEventsForInvoice(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw, domain.PaymentEventFilters{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return ListPaymentEventsResult{}, domainerrors.Internal(err)
	}

	total, err := uc.repo.CountPaymentEventsForInvoice(ctx, actor.TenantID.String(), actor.BranchID.String(), invoiceIDRaw)
	if err != nil {
		return ListPaymentEventsResult{}, domainerrors.Internal(err)
	}

	items := make([]PaymentEventResult, 0, len(events))
	for _, e := range events {
		items = append(items, PaymentEventResult{
			PaymentEventID:          e.PaymentEventID,
			PaymentAttemptID:        e.PaymentAttemptID,
			StripeEventID:           e.StripeEventID,
			StripeEventType:         e.StripeEventType,
			StripeCheckoutSessionID: e.StripeCheckoutSessionID,
			StripePaymentIntentID:   e.StripePaymentIntentID,
			Outcome:                 e.Outcome,
			ReasonCode:              e.ReasonCode,
			PreviousInvoiceStatus:   e.PreviousInvoiceStatus,
			NewInvoiceStatus:        e.NewInvoiceStatus,
			AttemptPreviousStatus:   e.AttemptPreviousStatus,
			AttemptNewStatus:        e.AttemptNewStatus,
			AmountMinor:             e.AmountMinor,
			CurrencyCode:            e.CurrencyCode,
			WebhookProcessingStatus: e.WebhookProcessingStatus,
			WebhookProcessingReason: e.WebhookProcessingReason,
			WebhookReceivedAt:       formatTimePtr(e.WebhookReceivedAt),
			WebhookProcessedAt:      formatTimePtr(e.WebhookProcessedAt),
			CreatedAt:               formatTime(e.CreatedAt),
		})
	}

	return ListPaymentEventsResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
