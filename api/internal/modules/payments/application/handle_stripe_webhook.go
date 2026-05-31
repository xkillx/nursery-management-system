package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type SystemAuditWriter interface {
	WriteSystemWithTx(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, requestID string, params SystemAuditParams) error
}

type SystemAuditParams struct {
	ActionType string
	EntityType string
	EntityID   uuid.UUID
	ReasonCode *string
	Details    map[string]any
}

type HandleStripeWebhook struct {
	repo     domain.WebhookRepository
	verifier domain.WebhookVerifier
	txMgr    domain.TxManager
	auditW   SystemAuditWriter
	newUUID  func() uuid.UUID
}

func NewHandleStripeWebhook(
	repo domain.WebhookRepository,
	verifier domain.WebhookVerifier,
	txMgr domain.TxManager,
	auditW SystemAuditWriter,
) *HandleStripeWebhook {
	return &HandleStripeWebhook{
		repo:     repo,
		verifier: verifier,
		txMgr:    txMgr,
		auditW:   auditW,
		newUUID:  uid.NewUUID,
	}
}

type WebhookResult struct {
	Status string
}

func (uc *HandleStripeWebhook) Execute(ctx context.Context, payload []byte, signatureHeader, requestID string) (WebhookResult, error) {
	if uc.verifier == nil {
		return WebhookResult{}, domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
	}

	event, err := uc.verifier.VerifyAndParse(ctx, payload, signatureHeader)
	if err != nil {
		return WebhookResult{}, domainerrors.New("payment_webhook_invalid_signature", "Webhook signature verification failed.")
	}

	var result WebhookResult
	txErr := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		eventID := uc.newUUID().String()
		event.ID = eventID

		insertedID, inserted, err := uc.repo.InsertWebhookEvent(ctx, tx, *event, requestID, domain.WebhookStatusReceived, "")
		if err != nil {
			return fmt.Errorf("insert webhook event: %w", err)
		}
		if !inserted {
			result = WebhookResult{Status: "duplicate"}
			return nil
		}
		event.ID = insertedID

		if !domain.CheckoutMutatingEventTypes[event.EventType] {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonUnsupportedEventType, "")
			result = WebhookResult{Status: "ignored"}
			return nil
		}

		if event.CheckoutSession == nil {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonMetadataMissing, "")
			result = WebhookResult{Status: "rejected"}
			return nil
		}

		cs := event.CheckoutSession

		if event.EventType == domain.EventTypeCheckoutCompleted && cs.PaymentStatus != "paid" {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonAwaitingAsyncPayment, "")
			result = WebhookResult{Status: "ignored"}
			return nil
		}

		meta := cs.Metadata
		tenantID, branchID, invoiceID, attemptID, ok := extractScope(meta)
		if !ok {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonMetadataMissing, "")
			result = WebhookResult{Status: "rejected"}
			return nil
		}

		row, err := uc.repo.GetPaymentAttemptAndInvoiceForWebhook(ctx, tx, tenantID, branchID, invoiceID, attemptID, cs.CheckoutSessionID)
		if err != nil {
			return fmt.Errorf("get payment attempt and invoice: %w", err)
		}
		if row == nil {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonUnknownPaymentAttempt, "")
			result = WebhookResult{Status: "rejected"}
			return nil
		}

		isSuccess := event.EventType == domain.EventTypeCheckoutCompleted || event.EventType == domain.EventTypeCheckoutAsyncSucceeded

		if isSuccess {
			result, err = uc.handleSuccess(ctx, tx, event, row, tenantID, branchID, invoiceID, attemptID, requestID, cs)
		} else {
			isExpired := event.EventType == domain.EventTypeCheckoutExpired
			result, err = uc.handleFailure(ctx, tx, event, row, tenantID, branchID, invoiceID, attemptID, requestID, isExpired)
		}
		return err
	})
	if txErr != nil {
		return WebhookResult{}, txErr
	}

	return result, nil
}

func (uc *HandleStripeWebhook) handleSuccess(
	ctx context.Context, tx pgx.Tx, event *domain.StripeWebhookEvent, row *domain.WebhookAttemptInvoice,
	tenantID, branchID, invoiceID, attemptID, requestID string, cs *domain.CheckoutSessionWebhookData,
) (WebhookResult, error) {
	prevInvoiceStatus := row.InvoiceStatus
	prevAttemptStatus := row.AttemptStatus

	if cs.AmountTotal != int64(row.AttemptAmountMinor) {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonAmountMismatch, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, prevInvoiceStatus, prevAttemptStatus, prevAttemptStatus, "rejected", domain.ReasonAmountMismatch, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		return WebhookResult{Status: "rejected"}, nil
	}

	if !strings.EqualFold(cs.Currency, row.AttemptCurrencyCode) {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonCurrencyMismatch, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, prevInvoiceStatus, prevAttemptStatus, prevAttemptStatus, "rejected", domain.ReasonCurrencyMismatch, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		return WebhookResult{Status: "rejected"}, nil
	}

	if prevInvoiceStatus == "paid" {
		if err := uc.repo.MarkPaymentAttemptPaid(ctx, tx, tenantID, branchID, attemptID); err != nil {
			return WebhookResult{}, fmt.Errorf("mark attempt paid (already-paid invoice): %w", err)
		}
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, domain.ReasonAlreadyPaid, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "paid", prevAttemptStatus, "paid", "ignored", domain.ReasonAlreadyPaid, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		return WebhookResult{Status: "processed"}, nil
	}

	if err := uc.repo.MarkPaymentAttemptPaid(ctx, tx, tenantID, branchID, attemptID); err != nil {
		return WebhookResult{}, fmt.Errorf("mark attempt paid: %w", err)
	}
	if err := uc.repo.MarkInvoicePaid(ctx, tx, tenantID, branchID, invoiceID); err != nil {
		return WebhookResult{}, fmt.Errorf("mark invoice paid: %w", err)
	}

	reasonCode := domain.ReasonPaid
	_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "paid", prevAttemptStatus, "paid", "paid", reasonCode, row.AttemptAmountMinor, row.AttemptCurrencyCode))
	_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, reasonCode, "")

	tenantUUID, _ := uuid.Parse(tenantID)
	branchUUID, _ := uuid.Parse(branchID)
	invoiceUUID, _ := uuid.Parse(invoiceID)
	_ = uc.auditW.WriteSystemWithTx(ctx, tx, tenantUUID, branchUUID, requestID, SystemAuditParams{
		ActionType: domain.AuditActionInvoicePaymentStatusUpdated,
		EntityType: domain.AuditEntityInvoice,
		EntityID:   invoiceUUID,
		ReasonCode: &reasonCode,
		Details: map[string]any{
			"previous_status": prevInvoiceStatus,
			"new_status":      "paid",
			"attempt_id":      attemptID,
		},
	})

	return WebhookResult{Status: "processed"}, nil
}

func (uc *HandleStripeWebhook) handleFailure(
	ctx context.Context, tx pgx.Tx, event *domain.StripeWebhookEvent, row *domain.WebhookAttemptInvoice,
	tenantID, branchID, invoiceID, attemptID, requestID string, isExpired bool,
) (WebhookResult, error) {
	prevInvoiceStatus := row.InvoiceStatus
	prevAttemptStatus := row.AttemptStatus

	var attemptOutcome, attemptNewStatus string
	if isExpired {
		if err := uc.repo.MarkPaymentAttemptExpired(ctx, tx, tenantID, branchID, attemptID); err != nil {
			return WebhookResult{}, fmt.Errorf("mark attempt expired: %w", err)
		}
		attemptOutcome = domain.ReasonExpired
		attemptNewStatus = domain.AttemptStatusExpired
	} else {
		if err := uc.repo.MarkPaymentAttemptFailed(ctx, tx, tenantID, branchID, attemptID); err != nil {
			return WebhookResult{}, fmt.Errorf("mark attempt failed: %w", err)
		}
		attemptOutcome = domain.ReasonPaymentFailed
		attemptNewStatus = domain.AttemptStatusPaymentFailed
	}

	if prevInvoiceStatus == "paid" {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonAlreadyPaid, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "paid", prevAttemptStatus, attemptNewStatus, "ignored", domain.ReasonAlreadyPaid, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		return WebhookResult{Status: "processed"}, nil
	}

	if prevInvoiceStatus == "issued" || prevInvoiceStatus == "overdue" {
		if err := uc.repo.MarkInvoicePaymentFailed(ctx, tx, tenantID, branchID, invoiceID); err != nil {
			return WebhookResult{}, fmt.Errorf("mark invoice payment_failed: %w", err)
		}

		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "payment_failed", prevAttemptStatus, attemptNewStatus, attemptOutcome, attemptOutcome, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, attemptOutcome, "")

		tenantUUID, _ := uuid.Parse(tenantID)
		branchUUID, _ := uuid.Parse(branchID)
		invoiceUUID, _ := uuid.Parse(invoiceID)
		_ = uc.auditW.WriteSystemWithTx(ctx, tx, tenantUUID, branchUUID, requestID, SystemAuditParams{
			ActionType: domain.AuditActionInvoicePaymentStatusUpdated,
			EntityType: domain.AuditEntityInvoice,
			EntityID:   invoiceUUID,
			ReasonCode: &attemptOutcome,
			Details: map[string]any{
				"previous_status": prevInvoiceStatus,
				"new_status":      "payment_failed",
				"attempt_id":      attemptID,
			},
		})

		return WebhookResult{Status: "processed"}, nil
	}

	// Invoice is already payment_failed — record history without rewriting timestamps
	_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "payment_failed", prevAttemptStatus, attemptNewStatus, attemptOutcome, domain.ReasonAlreadyPaymentFailed, row.AttemptAmountMinor, row.AttemptCurrencyCode))
	_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, domain.ReasonAlreadyPaymentFailed, "")

	return WebhookResult{Status: "processed"}, nil
}

func (uc *HandleStripeWebhook) makeReconciliationParams(
	tenantID, branchID, invoiceID, attemptID string,
	event *domain.StripeWebhookEvent,
	prevInvoiceStatus, newInvoiceStatus, prevAttemptStatus, newAttemptStatus string,
	outcome, reasonCode string, amountMinor int32, currencyCode string,
) domain.ReconciliationRecordParams {
	var checkoutSessionID, paymentIntentID string
	if event.CheckoutSession != nil {
		checkoutSessionID = event.CheckoutSession.CheckoutSessionID
		paymentIntentID = event.CheckoutSession.PaymentIntentID
	}

	details, _ := json.Marshal(map[string]string{"stripe_event_id": event.StripeEventID})

	return domain.ReconciliationRecordParams{
		ID:                    uc.newUUID().String(),
		TenantID:              tenantID,
		BranchID:              branchID,
		InvoiceID:             invoiceID,
		PaymentAttemptID:      attemptID,
		WebhookEventID:        event.ID,
		StripeEventID:         event.StripeEventID,
		StripeEventType:       event.EventType,
		CheckoutSessionID:     checkoutSessionID,
		PaymentIntentID:       paymentIntentID,
		Outcome:               outcome,
		ReasonCode:            reasonCode,
		PreviousInvoiceStatus: prevInvoiceStatus,
		NewInvoiceStatus:      newInvoiceStatus,
		AttemptPreviousStatus: prevAttemptStatus,
		AttemptNewStatus:      newAttemptStatus,
		AmountMinor:           amountMinor,
		CurrencyCode:          currencyCode,
		Details:               string(details),
	}
}

func extractScope(meta map[string]string) (tenantID, branchID, invoiceID, attemptID string, ok bool) {
	if meta == nil {
		return "", "", "", "", false
	}
	tenantID = meta["tenant_id"]
	branchID = meta["branch_id"]
	invoiceID = meta["invoice_id"]
	attemptID = meta["payment_attempt_id"]
	if tenantID == "" || branchID == "" || invoiceID == "" || attemptID == "" {
		return "", "", "", "", false
	}
	if _, err := uuid.Parse(tenantID); err != nil {
		return "", "", "", "", false
	}
	if _, err := uuid.Parse(branchID); err != nil {
		return "", "", "", "", false
	}
	if _, err := uuid.Parse(invoiceID); err != nil {
		return "", "", "", "", false
	}
	if _, err := uuid.Parse(attemptID); err != nil {
		return "", "", "", "", false
	}
	return tenantID, branchID, invoiceID, attemptID, true
}
