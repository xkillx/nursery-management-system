package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/logging"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/uid"
)

type SystemAuditWriter interface {
	WriteSystemWithTx(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, requestID string, params SystemAuditParams) error
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
	logger   *slog.Logger
	recorder *metrics.Recorder
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

func (uc *HandleStripeWebhook) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *HandleStripeWebhook {
	return &HandleStripeWebhook{
		repo:     uc.repo,
		verifier: uc.verifier,
		txMgr:    uc.txMgr,
		auditW:   uc.auditW,
		newUUID:  uc.newUUID,
		logger:   logger,
		recorder: recorder,
	}
}

func (uc *HandleStripeWebhook) recordWebhookOutcome(provider, eventType, outcome, reason string) {
	if uc.recorder != nil {
		uc.recorder.WebhookOutcome(provider, eventType, outcome, reason)
	}
}

func (uc *HandleStripeWebhook) recordTransition(operation, entityType, previousStatus, newStatus, reason string) {
	if uc.recorder != nil {
		uc.recorder.PaymentStateTransition(operation, entityType, previousStatus, newStatus, reason)
	}
}

func (uc *HandleStripeWebhook) logInfo(msg string, args ...any) {
	if uc.logger != nil {
		uc.logger.Info(msg, args...)
	}
}

func (uc *HandleStripeWebhook) logDebug(msg string, args ...any) {
	if uc.logger != nil {
		uc.logger.Debug(msg, args...)
	}
}

type WebhookResult struct {
	Status            string
	EventType         string
	StripeEventID     string
	CheckoutSessionID string
	PaymentIntentID   string
}

func (uc *HandleStripeWebhook) Execute(ctx context.Context, payload []byte, signatureHeader, requestID string) (WebhookResult, error) {
	if uc.verifier == nil {
		uc.recordWebhookOutcome("stripe", "unknown", "unconfigured", "payment_provider_unconfigured")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "unconfigured",
			"reason", "payment_provider_unconfigured",
			"request_id", requestID,
		)
		return WebhookResult{}, domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
	}

	event, err := uc.verifier.VerifyAndParse(ctx, payload, signatureHeader)
	if err != nil {
		uc.recordWebhookOutcome("stripe", "unknown", "invalid_signature", "payment_webhook_invalid_signature")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "invalid_signature",
			"reason", "payment_webhook_invalid_signature",
			"request_id", requestID,
		)
		return WebhookResult{}, domainerrors.New("payment_webhook_invalid_signature", "Webhook signature verification failed.")
	}

	var result WebhookResult
	txErr := uc.txMgr.ExecTx(ctx, func(tx domain.Tx) error {
		eventID := uc.newUUID().String()
		event.ID = eventID

		insertedID, inserted, err := uc.repo.InsertWebhookEvent(ctx, tx, *event, requestID, domain.WebhookStatusReceived, "")
		if err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "insert_webhook_event",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
			return fmt.Errorf("insert webhook event: %w", err)
		}
		if !inserted {
			uc.recordWebhookOutcome("stripe", event.EventType, "duplicate", "duplicate_event")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "duplicate",
				"reason", "duplicate_event",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "duplicate", EventType: event.EventType, StripeEventID: event.StripeEventID}
			return nil
		}
		event.ID = insertedID

		if !domain.CheckoutMutatingEventTypes[event.EventType] {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonUnsupportedEventType, "")
			uc.recordWebhookOutcome("stripe", event.EventType, "ignored", "unsupported_event_type")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "ignored",
				"reason", "unsupported_event_type",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "ignored", EventType: event.EventType, StripeEventID: event.StripeEventID}
			return nil
		}

		if event.CheckoutSession == nil {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonMetadataMissing, "")
			uc.recordWebhookOutcome("stripe", event.EventType, "rejected", "metadata_missing")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "rejected",
				"reason", "metadata_missing",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "rejected", EventType: event.EventType, StripeEventID: event.StripeEventID}
			return nil
		}

		cs := event.CheckoutSession

		if event.EventType == domain.EventTypeCheckoutCompleted && cs.PaymentStatus != "paid" {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonAwaitingAsyncPayment, "")
			uc.recordWebhookOutcome("stripe", event.EventType, "ignored", "awaiting_async_payment")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "ignored",
				"reason", "awaiting_async_payment",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "ignored", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID}
			return nil
		}

		meta := cs.Metadata
		tenantID, branchID, invoiceID, attemptID, ok := extractScope(meta)
		if !ok {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonMetadataMissing, "")
			uc.recordWebhookOutcome("stripe", event.EventType, "rejected", "metadata_missing")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "rejected",
				"reason", "metadata_missing_scope",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "rejected", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID}
			return nil
		}

		row, err := uc.repo.GetPaymentAttemptAndInvoiceForWebhook(ctx, tx, tenantID, branchID, invoiceID, attemptID, cs.CheckoutSessionID)
		if err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "get_payment_attempt_and_invoice",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
			return fmt.Errorf("get payment attempt and invoice: %w", err)
		}
		if row == nil {
			_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonUnknownPaymentAttempt, "")
			uc.recordWebhookOutcome("stripe", event.EventType, "rejected", "unknown_payment_attempt")
			uc.logInfo("stripe_webhook",
				"operation", "stripe_webhook",
				"outcome", "rejected",
				"reason", "unknown_payment_attempt",
				"stripe_event_id", event.StripeEventID,
				"stripe_event_type", event.EventType,
				"request_id", requestID,
			)
			result = WebhookResult{Status: "rejected", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID}
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
		uc.recordWebhookOutcome("stripe", event.EventType, "error", "internal_error")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "error",
			"reason", "internal_error",
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
			"error", logging.SafeErr(txErr),
		)
		return WebhookResult{}, txErr
	}

	return result, nil
}

func (uc *HandleStripeWebhook) handleSuccess(
	ctx context.Context, tx domain.Tx, event *domain.StripeWebhookEvent, row *domain.WebhookAttemptInvoice,
	tenantID, branchID, invoiceID, attemptID, requestID string, cs *domain.CheckoutSessionWebhookData,
) (WebhookResult, error) {
	prevInvoiceStatus := row.InvoiceStatus
	prevAttemptStatus := row.AttemptStatus

	if cs.AmountTotal != int64(row.AttemptAmountMinor) {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonAmountMismatch, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, prevInvoiceStatus, prevAttemptStatus, prevAttemptStatus, "rejected", domain.ReasonAmountMismatch, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		uc.recordWebhookOutcome("stripe", event.EventType, "rejected", "amount_mismatch")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "rejected",
			"reason", "amount_mismatch",
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
		)
		return WebhookResult{Status: "rejected", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID, PaymentIntentID: cs.PaymentIntentID}, nil
	}

	if !strings.EqualFold(cs.Currency, row.AttemptCurrencyCode) {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusRejected, domain.ReasonCurrencyMismatch, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, prevInvoiceStatus, prevAttemptStatus, prevAttemptStatus, "rejected", domain.ReasonCurrencyMismatch, row.AttemptAmountMinor, row.AttemptCurrencyCode))
		uc.recordWebhookOutcome("stripe", event.EventType, "rejected", "currency_mismatch")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "rejected",
			"reason", "currency_mismatch",
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
		)
		return WebhookResult{Status: "rejected", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID, PaymentIntentID: cs.PaymentIntentID}, nil
	}

	if prevInvoiceStatus == "paid" {
		if err := uc.repo.MarkPaymentAttemptPaid(ctx, tx, tenantID, branchID, attemptID); err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "mark_attempt_paid_already_paid",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
			return WebhookResult{}, fmt.Errorf("mark attempt paid (already-paid invoice): %w", err)
		}
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, domain.ReasonAlreadyPaid, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "paid", prevAttemptStatus, "paid", "ignored", domain.ReasonAlreadyPaid, row.AttemptAmountMinor, row.AttemptCurrencyCode))

		uc.recordTransition("stripe_webhook", "payment_attempt", "checkout_created", "paid", "already_paid")
		uc.recordWebhookOutcome("stripe", event.EventType, "processed", "already_paid")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "processed",
			"reason", "already_paid",
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
		)

		return WebhookResult{Status: "processed", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID, PaymentIntentID: cs.PaymentIntentID}, nil
	}

	if err := uc.repo.MarkPaymentAttemptPaid(ctx, tx, tenantID, branchID, attemptID); err != nil {
		uc.logDebug("stripe_webhook_repo",
			"operation", "mark_attempt_paid",
			"request_id", requestID,
			"stripe_event_id", event.StripeEventID,
			"error", logging.SafeErr(err),
		)
		return WebhookResult{}, fmt.Errorf("mark attempt paid: %w", err)
	}
	if err := uc.repo.MarkInvoicePaid(ctx, tx, tenantID, branchID, invoiceID); err != nil {
		uc.logDebug("stripe_webhook_repo",
			"operation", "mark_invoice_paid",
			"request_id", requestID,
			"stripe_event_id", event.StripeEventID,
			"error", logging.SafeErr(err),
		)
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

	uc.recordTransition("stripe_webhook", "payment_attempt", "checkout_created", "paid", "paid")
	uc.recordTransition("stripe_webhook", "invoice", prevInvoiceStatus, "paid", "paid")
	uc.recordWebhookOutcome("stripe", event.EventType, "processed", "paid")
	uc.logInfo("stripe_webhook",
		"operation", "stripe_webhook",
		"outcome", "processed",
		"reason", "paid",
		"stripe_event_id", event.StripeEventID,
		"stripe_event_type", event.EventType,
		"request_id", requestID,
		"previous_invoice_status", prevInvoiceStatus,
	)

	return WebhookResult{Status: "processed", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: cs.CheckoutSessionID, PaymentIntentID: cs.PaymentIntentID}, nil
}

func (uc *HandleStripeWebhook) handleFailure(
	ctx context.Context, tx domain.Tx, event *domain.StripeWebhookEvent, row *domain.WebhookAttemptInvoice,
	tenantID, branchID, invoiceID, attemptID, requestID string, isExpired bool,
) (WebhookResult, error) {
	prevInvoiceStatus := row.InvoiceStatus
	prevAttemptStatus := row.AttemptStatus

	var attemptOutcome, attemptNewStatus string
	if isExpired {
		if err := uc.repo.MarkPaymentAttemptExpired(ctx, tx, tenantID, branchID, attemptID); err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "mark_attempt_expired",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
			return WebhookResult{}, fmt.Errorf("mark attempt expired: %w", err)
		}
		attemptOutcome = domain.ReasonExpired
		attemptNewStatus = domain.AttemptStatusExpired
	} else {
		if err := uc.repo.MarkPaymentAttemptFailed(ctx, tx, tenantID, branchID, attemptID); err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "mark_attempt_failed",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
			return WebhookResult{}, fmt.Errorf("mark attempt failed: %w", err)
		}
		attemptOutcome = domain.ReasonPaymentFailed
		attemptNewStatus = domain.AttemptStatusPaymentFailed
	}

	var csID, piID string
	if event.CheckoutSession != nil {
		csID = event.CheckoutSession.CheckoutSessionID
		piID = event.CheckoutSession.PaymentIntentID
	}

	if prevInvoiceStatus == "paid" {
		_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusIgnored, domain.ReasonAlreadyPaid, "")
		_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "paid", prevAttemptStatus, attemptNewStatus, "ignored", domain.ReasonAlreadyPaid, row.AttemptAmountMinor, row.AttemptCurrencyCode))

		uc.recordTransition("stripe_webhook", "payment_attempt", "checkout_created", attemptNewStatus, attemptOutcome)
		uc.recordWebhookOutcome("stripe", event.EventType, "processed", "already_paid")
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "processed",
			"reason", "already_paid",
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
		)

		return WebhookResult{Status: "processed", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: csID, PaymentIntentID: piID}, nil
	}

	if prevInvoiceStatus == "issued" || prevInvoiceStatus == "overdue" {
		if err := uc.repo.MarkInvoicePaymentFailed(ctx, tx, tenantID, branchID, invoiceID); err != nil {
			uc.logDebug("stripe_webhook_repo",
				"operation", "mark_invoice_payment_failed",
				"request_id", requestID,
				"stripe_event_id", event.StripeEventID,
				"error", logging.SafeErr(err),
			)
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

		uc.recordTransition("stripe_webhook", "payment_attempt", "checkout_created", attemptNewStatus, attemptOutcome)
		uc.recordTransition("stripe_webhook", "invoice", prevInvoiceStatus, "payment_failed", attemptOutcome)
		uc.recordWebhookOutcome("stripe", event.EventType, "processed", attemptOutcome)
		uc.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "processed",
			"reason", attemptOutcome,
			"stripe_event_id", event.StripeEventID,
			"stripe_event_type", event.EventType,
			"request_id", requestID,
			"previous_invoice_status", prevInvoiceStatus,
		)

		return WebhookResult{Status: "processed", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: csID, PaymentIntentID: piID}, nil
	}

	// Invoice is already payment_failed -- record history without rewriting timestamps
	_ = uc.repo.InsertReconciliationRecord(ctx, tx, uc.makeReconciliationParams(tenantID, branchID, invoiceID, attemptID, event, prevInvoiceStatus, "payment_failed", prevAttemptStatus, attemptNewStatus, attemptOutcome, domain.ReasonAlreadyPaymentFailed, row.AttemptAmountMinor, row.AttemptCurrencyCode))
	_ = uc.repo.UpdateWebhookEventStatus(ctx, tx, event.ID, domain.WebhookStatusProcessed, domain.ReasonAlreadyPaymentFailed, "")

	uc.recordTransition("stripe_webhook", "payment_attempt", "checkout_created", attemptNewStatus, attemptOutcome)
	uc.recordWebhookOutcome("stripe", event.EventType, "processed", "already_payment_failed")
	uc.logInfo("stripe_webhook",
		"operation", "stripe_webhook",
		"outcome", "processed",
		"reason", "already_payment_failed",
		"stripe_event_id", event.StripeEventID,
		"stripe_event_type", event.EventType,
		"request_id", requestID,
	)

	return WebhookResult{Status: "processed", EventType: event.EventType, StripeEventID: event.StripeEventID, CheckoutSessionID: csID, PaymentIntentID: piID}, nil
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
