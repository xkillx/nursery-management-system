package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/uid"
)

// EmailCheckoutResult is the outcome of an email-initiated checkout session
// creation. OK=false is the explicit "use the fallback link" signal: the caller
// (email send worker) must never fail a send because of a payment-link problem
// (R3, R8).
type EmailCheckoutResult struct {
	CheckoutURL string
	OK          bool
}

// CreateEmailCheckoutSession creates or reuses a Stripe Checkout session for an
// invoice on behalf of an email recipient, without a portal membership. It
// mirrors CreateCheckoutSession minus the membership resolution and is wired to
// the email worker through the consumer-side InvoicePayLinkProvider (KTD2, KTD3).
type CreateEmailCheckoutSession struct {
	repo             domain.PaymentRepository
	txMgr            domain.TxManager
	provider         domain.CheckoutProvider
	webBaseURL       string
	stripeConfigured bool
	newUUID          func() uuid.UUID
	logger           *slog.Logger
	recorder         *metrics.Recorder
}

func NewCreateEmailCheckoutSession(
	repo domain.PaymentRepository,
	txMgr domain.TxManager,
	provider domain.CheckoutProvider,
	webBaseURL string,
	stripeConfigured bool,
) *CreateEmailCheckoutSession {
	return &CreateEmailCheckoutSession{
		repo:             repo,
		txMgr:            txMgr,
		provider:         provider,
		webBaseURL:       strings.TrimRight(webBaseURL, "/"),
		stripeConfigured: stripeConfigured,
		newUUID:          uid.NewUUID,
	}
}

func (uc *CreateEmailCheckoutSession) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *CreateEmailCheckoutSession {
	return &CreateEmailCheckoutSession{
		repo:             uc.repo,
		txMgr:            uc.txMgr,
		provider:         uc.provider,
		webBaseURL:       uc.webBaseURL,
		stripeConfigured: uc.stripeConfigured,
		newUUID:          uc.newUUID,
		logger:           logger,
		recorder:         recorder,
	}
}

func (uc *CreateEmailCheckoutSession) recordTransition(operation, entityType, previousStatus, newStatus, reason string) {
	if uc.recorder != nil {
		uc.recorder.PaymentStateTransition(operation, entityType, previousStatus, newStatus, reason)
	}
}

func (uc *CreateEmailCheckoutSession) logDebug(msg string, args ...any) {
	if uc.logger != nil {
		uc.logger.Debug(msg, args...)
	}
}

// Execute creates (or reuses) a checkout session for an email recipient. It
// returns (url, ok=true) when a session is available, and ok=false when the
// provider is unconfigured, the invoice is not payable, or a provider/state
// error occurred. The caller must treat ok=false as "use the app link" (R3, R8).
func (uc *CreateEmailCheckoutSession) Execute(ctx context.Context, tenantID, branchID, invoiceIDRaw, requestID string) (EmailCheckoutResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return EmailCheckoutResult{OK: false}, nil
	}

	if !uc.stripeConfigured {
		return EmailCheckoutResult{OK: false}, nil
	}

	// Idempotency: reuse a live email-flow session (KTD5, KTD8). Expired sessions
	// fall out of the active lookup, so a fresh one is created below.
	if active, found, _ := uc.repo.GetActiveEmailCheckoutForInvoice(ctx, tenantID, branchID, invoiceID.String()); found && active != nil {
		uc.logDebug("email_checkout_session_idempotent",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", active.AttemptID,
		)
		return EmailCheckoutResult{CheckoutURL: active.CheckoutURL, OK: true}, nil
	}

	steps := checkoutSessionSteps{repo: uc.repo, txMgr: uc.txMgr, provider: uc.provider, newUUID: uc.newUUID}
	candidate, attemptID, txErr := steps.createAttempt(ctx, tenantID, branchID, invoiceID.String(), requestID, "", "",
		func(ctx context.Context, tx domain.Tx) (domain.CheckoutInvoiceCandidate, bool, error) {
			return uc.repo.GetInvoiceForEmailCheckoutForUpdate(ctx, tx, tenantID, branchID, invoiceID.String())
		},
		uc.logDebug,
	)
	if txErr != nil {
		if errors.Is(txErr, errCheckoutInvoiceNotFound) || errors.Is(txErr, errCheckoutInvoiceNotPayable) {
			return EmailCheckoutResult{OK: false}, nil
		}
		return EmailCheckoutResult{OK: false}, txErr
	}

	uc.recordTransition("email_checkout", "payment_attempt", "none", "checkout_creation_started", "email_send_requested")

	// Email sessions land the payer on the public outcome page (R5), carrying the
	// outcome, invoice id, and the Stripe session id for the success redirect.
	successURL := fmt.Sprintf("%s/payment/result?outcome=success&invoice_id=%s&session_id={CHECKOUT_SESSION_ID}", uc.webBaseURL, invoiceID.String())
	cancelURL := fmt.Sprintf("%s/payment/result?outcome=cancelled&invoice_id=%s", uc.webBaseURL, invoiceID.String())

	result, providerErr := steps.callProviderAndMark(ctx, tenantID, branchID, invoiceID.String(), attemptID.String(), requestID, candidate, successURL, cancelURL)
	if providerErr != nil {
		uc.recordTransition("email_checkout", "payment_attempt", "checkout_creation_started", "checkout_creation_failed", "email_checkout_fallback")
		uc.logDebug("email_checkout_session_fallback",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", attemptID.String(),
			"error", providerErr,
		)
		return EmailCheckoutResult{OK: false}, nil
	}

	uc.recordTransition("email_checkout", "payment_attempt", "checkout_creation_started", "checkout_created", "email_checkout_created")
	uc.logDebug("email_checkout_session_created",
		"request_id", requestID,
		"invoice_id", invoiceID.String(),
		"attempt_id", attemptID.String(),
		"checkout_session_id", result.CheckoutSessionID,
	)

	return EmailCheckoutResult{CheckoutURL: result.CheckoutURL, OK: true}, nil
}
