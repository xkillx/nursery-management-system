package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/logging"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/uid"
)

var payableStatuses = map[string]bool{
	"issued":         true,
	"payment_failed": true,
	"overdue":        true,
}

type CreateCheckoutSession struct {
	repo             domain.PaymentRepository
	txMgr            domain.TxManager
	provider         domain.CheckoutProvider
	webBaseURL       string
	stripeConfigured bool
	newUUID          func() uuid.UUID
	logger           *slog.Logger
	recorder         *metrics.Recorder
}

func NewCreateCheckoutSession(
	repo domain.PaymentRepository,
	txMgr domain.TxManager,
	provider domain.CheckoutProvider,
	webBaseURL string,
	stripeConfigured bool,
) *CreateCheckoutSession {
	return &CreateCheckoutSession{
		repo:             repo,
		txMgr:            txMgr,
		provider:         provider,
		webBaseURL:       strings.TrimRight(webBaseURL, "/"),
		stripeConfigured: stripeConfigured,
		newUUID:          uid.NewUUID,
	}
}

func (uc *CreateCheckoutSession) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *CreateCheckoutSession {
	return &CreateCheckoutSession{
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

func (uc *CreateCheckoutSession) recordTransition(operation, entityType, previousStatus, newStatus, reason string) {
	if uc.recorder != nil {
		uc.recorder.PaymentStateTransition(operation, entityType, previousStatus, newStatus, reason)
	}
}

func (uc *CreateCheckoutSession) logDebug(msg string, args ...any) {
	if uc.logger != nil {
		uc.logger.Debug(msg, args...)
	}
}

type CreateCheckoutSessionResult struct {
	CheckoutSessionID string
	CheckoutURL       string
	PaymentAttemptID  string
}

func (uc *CreateCheckoutSession) Execute(ctx context.Context, tenantID, branchID, membershipID, userID, invoiceIDRaw, requestID string) (CreateCheckoutSessionResult, error) {
	invoiceID, err := uuid.Parse(invoiceIDRaw)
	if err != nil {
		return CreateCheckoutSessionResult{}, domainerrors.Validation("Invalid invoice ID format.", "invoice_id")
	}

	if !uc.stripeConfigured {
		uc.recordTransition("parent_checkout", "payment_attempt", "none", "checkout_creation_failed", "payment_provider_unconfigured")
		return CreateCheckoutSessionResult{}, domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
	}

	var candidate domain.CheckoutInvoiceCandidate
	var attemptID uuid.UUID

	txErr := uc.txMgr.ExecTx(ctx, func(tx domain.Tx) error {
		row, found, err := uc.repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, tenantID, branchID, membershipID, invoiceID.String())
		if err != nil {
			uc.logDebug("checkout_session_repo",
				"operation", "get_parent_invoice_for_checkout",
				"request_id", requestID,
				"invoice_id", invoiceID.String(),
				"error", logging.SafeErr(err),
			)
			return fmt.Errorf("get parent invoice for checkout: %w", err)
		}
		if !found {
			return domainerrors.NotFound("invoice", "Invoice not found.")
		}

		candidate = row

		if !uc.isPayable(row) {
			return domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
		}

		attemptID = uc.newUUID()

		return uc.repo.CreatePaymentAttempt(ctx, tx, domain.PaymentAttemptCreateParams{
			ID:                      attemptID.String(),
			TenantID:                tenantID,
			BranchID:                branchID,
			InvoiceID:               invoiceID.String(),
			InitiatedByUserID:       userID,
			InitiatedByMembershipID: membershipID,
			RequestID:               requestID,
			Status:                  domain.AttemptStatusCheckoutCreationStarted,
			AmountMinor:             candidate.TotalDueMinor,
			CurrencyCode:            domain.CurrencyGBP,
		})
	})
	if txErr != nil {
		return CreateCheckoutSessionResult{}, txErr
	}

	uc.recordTransition("parent_checkout", "payment_attempt", "none", "checkout_creation_started", "parent_checkout_requested")

	productDesc := ""
	if candidate.InvoiceNumber != "" {
		productDesc = "Invoice " + candidate.InvoiceNumber
	}

	successURL := fmt.Sprintf("%s/parent/invoices/%s?checkout=success&session_id={CHECKOUT_SESSION_ID}", uc.webBaseURL, invoiceID.String())
	cancelURL := fmt.Sprintf("%s/parent/invoices/%s?checkout=cancelled", uc.webBaseURL, invoiceID.String())

	result, providerErr := uc.provider.CreateCheckoutSession(ctx, domain.CheckoutSessionCreateParams{
		PaymentAttemptID: attemptID.String(),
		InvoiceID:        invoiceID.String(),
		InvoiceNumber:    candidate.InvoiceNumber,
		AmountMinor:      candidate.TotalDueMinor,
		Currency:         "gbp",
		ProductName:      "Nursery invoice payment",
		ProductDesc:      productDesc,
		SuccessURL:       successURL,
		CancelURL:        cancelURL,
		TenantID:         tenantID,
		BranchID:         branchID,
	})
	if providerErr != nil {
		_ = uc.repo.MarkPaymentAttemptCheckoutCreationFailed(ctx, domain.PaymentAttemptCheckoutCreationFailedParams{
			TenantID:             tenantID,
			BranchID:             branchID,
			AttemptID:            attemptID.String(),
			FailureReason:        domain.FailureReasonStripeError,
			ProviderErrorCode:    safeProviderCode(providerErr),
			ProviderErrorMessage: safeProviderMessage(providerErr),
		})
		uc.recordTransition("parent_checkout", "payment_attempt", "checkout_creation_started", "checkout_creation_failed", "stripe_error")
		uc.logDebug("checkout_session_provider",
			"operation", "create_checkout_session",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", attemptID.String(),
			"error", logging.SafeErr(providerErr),
		)
		return CreateCheckoutSessionResult{}, domainerrors.New("payment_provider_error", "Payment provider failed to create checkout session.")
	}

	state, found, err := uc.repo.GetInvoicePaymentState(ctx, tenantID, branchID, invoiceID.String())
	if err != nil || !found || !uc.isStatePayable(state) {
		_ = uc.repo.MarkPaymentAttemptCheckoutCreationFailed(ctx, domain.PaymentAttemptCheckoutCreationFailedParams{
			TenantID:      tenantID,
			BranchID:      branchID,
			AttemptID:     attemptID.String(),
			FailureReason: domain.FailureReasonInvoiceNoLongerPayable,
		})
		uc.recordTransition("parent_checkout", "payment_attempt", "checkout_creation_started", "checkout_creation_failed", "invoice_no_longer_payable")
		uc.logDebug("checkout_session_state",
			"operation", "check_invoice_payment_state",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", attemptID.String(),
		)
		return CreateCheckoutSessionResult{}, domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
	}

	var expiresAt *time.Time
	if result.ExpiresAt != "" {
		if ts, parseErr := parseTimestamp(result.ExpiresAt); parseErr == nil {
			expiresAt = &ts
		}
	}

	if markErr := uc.repo.MarkPaymentAttemptCheckoutCreated(ctx, domain.PaymentAttemptCheckoutCreatedParams{
		TenantID:                tenantID,
		BranchID:                branchID,
		AttemptID:               attemptID.String(),
		StripeCheckoutSessionID: result.CheckoutSessionID,
		StripeCheckoutURL:       result.CheckoutURL,
		StripePaymentIntentID:   result.PaymentIntentID,
		StripeExpiresAt:         expiresAt,
	}); markErr != nil {
		uc.logDebug("checkout_session_repo",
			"operation", "mark_payment_attempt_checkout_created",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", attemptID.String(),
			"error", logging.SafeErr(markErr),
		)
		return CreateCheckoutSessionResult{}, domainerrors.Internal(fmt.Errorf("mark payment attempt created: %w", markErr))
	}

	uc.recordTransition("parent_checkout", "payment_attempt", "checkout_creation_started", "checkout_created", "checkout_created")
	uc.logDebug("checkout_session_created",
		"operation", "create_checkout_session",
		"request_id", requestID,
		"invoice_id", invoiceID.String(),
		"attempt_id", attemptID.String(),
		"checkout_session_id", result.CheckoutSessionID,
		"payment_intent_id", result.PaymentIntentID,
	)

	return CreateCheckoutSessionResult{
		CheckoutSessionID: result.CheckoutSessionID,
		CheckoutURL:       result.CheckoutURL,
		PaymentAttemptID:  attemptID.String(),
	}, nil
}

func (uc *CreateCheckoutSession) isPayable(c domain.CheckoutInvoiceCandidate) bool {
	if c.InvoiceKind != "monthly" {
		return false
	}
	if !payableStatuses[c.Status] {
		return false
	}
	if c.CurrencyCode != "GBP" {
		return false
	}
	if c.TotalDueMinor <= 0 {
		return false
	}
	if c.AmountPaidMinor != 0 {
		return false
	}
	return true
}

func (uc *CreateCheckoutSession) isStatePayable(s domain.InvoicePaymentState) bool {
	if s.InvoiceKind != "monthly" {
		return false
	}
	if !payableStatuses[s.Status] {
		return false
	}
	if s.CurrencyCode != "GBP" {
		return false
	}
	if s.TotalDueMinor <= 0 {
		return false
	}
	if s.AmountPaidMinor != 0 {
		return false
	}
	return true
}

func safeProviderCode(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 100 {
		msg = msg[:100]
	}
	return msg
}

func safeProviderMessage(err error) string {
	return safeProviderCode(err)
}

func parseTimestamp(s string) (time.Time, error) {
	i, err := time.Parse("2006-01-02 15:04:05 -0700 MST", s)
	if err == nil {
		return i, nil
	}
	ts, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return ts, nil
	}
	epoch, err := time.Parse("1504000000000", s)
	if err == nil {
		return epoch, nil
	}
	// Try unix timestamp
	var sec int64
	fmt.Sscanf(s, "%d", &sec)
	if sec > 0 {
		return time.Unix(sec, 0), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse timestamp: %s", s)
}
