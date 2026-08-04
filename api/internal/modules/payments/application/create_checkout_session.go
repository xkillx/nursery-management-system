package application

import (
	"context"
	"errors"
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

// Sentinel errors returned by the shared checkout pipeline. Each use case maps
// them to its own surface (portal: DomainError codes; email: ok=false fallback).
var (
	errCheckoutInvoiceNotFound   = errors.New("checkout invoice not found")
	errCheckoutInvoiceNotPayable = errors.New("checkout invoice not payable")
	errCheckoutStateNotPayable   = errors.New("checkout invoice no longer payable")
	errCheckoutMarkCreated       = errors.New("mark payment attempt checkout created")
)

// checkoutSessionSteps carries the Stripe checkout-session creation pipeline
// shared by the portal and email checkout use cases (KTD3): invoice resolution
// under the row lock, attempt creation, the provider call, the post-Stripe state
// re-check, and marking the attempt created (or failed).
type checkoutSessionSteps struct {
	repo     domain.PaymentRepository
	txMgr    domain.TxManager
	provider domain.CheckoutProvider
	newUUID  func() uuid.UUID
}

// checkoutCandidateResolver resolves the invoice candidate for a checkout flow.
// Portal passes the membership-joined lookup; email passes the non-membership one.
type checkoutCandidateResolver func(ctx context.Context, tx domain.Tx) (domain.CheckoutInvoiceCandidate, bool, error)

// createAttempt resolves the invoice under the row lock (FOR UPDATE), checks
// payability, and creates a checkout_creation_started attempt. Returns the
// resolved candidate and attempt id. Errors: errCheckoutInvoiceNotFound,
// errCheckoutInvoiceNotPayable, or the wrapped repo error.
func (s *checkoutSessionSteps) createAttempt(
	ctx context.Context,
	tenantID, branchID, invoiceID, requestID, userID, membershipID string,
	resolve checkoutCandidateResolver,
	logf func(msg string, args ...any),
) (domain.CheckoutInvoiceCandidate, uuid.UUID, error) {
	var candidate domain.CheckoutInvoiceCandidate
	var attemptID uuid.UUID

	txErr := s.txMgr.ExecTx(ctx, func(tx domain.Tx) error {
		row, found, err := resolve(ctx, tx)
		if err != nil {
			if logf != nil {
				logf("checkout_session_repo",
					"operation", "resolve_invoice_for_checkout",
					"invoice_id", invoiceID,
					"error", logging.SafeErr(err),
				)
			}
			return fmt.Errorf("resolve invoice for checkout: %w", err)
		}
		if !found {
			return errCheckoutInvoiceNotFound
		}
		if !isPayableCandidate(row) {
			return errCheckoutInvoiceNotPayable
		}

		candidate = row
		attemptID = s.newUUID()

		return s.repo.CreatePaymentAttempt(ctx, tx, domain.PaymentAttemptCreateParams{
			ID:                      attemptID.String(),
			TenantID:                tenantID,
			BranchID:                branchID,
			InvoiceID:               invoiceID,
			InitiatedByUserID:       userID,
			InitiatedByMembershipID: membershipID,
			RequestID:               requestID,
			Status:                  domain.AttemptStatusCheckoutCreationStarted,
			AmountMinor:             candidate.TotalDueMinor,
			CurrencyCode:            domain.CurrencyGBP,
		})
	})
	return candidate, attemptID, txErr
}

// callProviderAndMark calls the provider, re-checks the invoice payment state,
// and marks the attempt checkout_created. On provider or state failure the
// attempt is marked checkout_creation_failed and an error is returned:
// errCheckoutStateNotPayable for the state re-check, errCheckoutMarkCreated for
// a failed mark-created, and a wrapped provider error otherwise.
func (s *checkoutSessionSteps) callProviderAndMark(
	ctx context.Context,
	tenantID, branchID, invoiceID, attemptID, requestID string,
	candidate domain.CheckoutInvoiceCandidate,
	successURL, cancelURL string,
) (domain.CheckoutSessionResult, error) {
	result, providerErr := s.provider.CreateCheckoutSession(ctx, domain.CheckoutSessionCreateParams{
		PaymentAttemptID: attemptID,
		InvoiceID:        invoiceID,
		InvoiceNumber:    candidate.InvoiceNumber,
		AmountMinor:      candidate.TotalDueMinor,
		Currency:         "gbp",
		ProductName:      "Nursery invoice payment",
		ProductDesc:      invoiceProductDesc(candidate.InvoiceNumber),
		SuccessURL:       successURL,
		CancelURL:        cancelURL,
		TenantID:         tenantID,
		BranchID:         branchID,
	})
	if providerErr != nil {
		_ = s.repo.MarkPaymentAttemptCheckoutCreationFailed(ctx, domain.PaymentAttemptCheckoutCreationFailedParams{
			TenantID:             tenantID,
			BranchID:             branchID,
			AttemptID:            attemptID,
			FailureReason:        domain.FailureReasonStripeError,
			ProviderErrorCode:    safeProviderCode(providerErr),
			ProviderErrorMessage: safeProviderMessage(providerErr),
		})
		return domain.CheckoutSessionResult{}, fmt.Errorf("payment provider: %w", providerErr)
	}

	state, found, err := s.repo.GetInvoicePaymentState(ctx, tenantID, branchID, invoiceID)
	if err != nil || !found || !isStatePayable(state) {
		_ = s.repo.MarkPaymentAttemptCheckoutCreationFailed(ctx, domain.PaymentAttemptCheckoutCreationFailedParams{
			TenantID:      tenantID,
			BranchID:      branchID,
			AttemptID:     attemptID,
			FailureReason: domain.FailureReasonInvoiceNoLongerPayable,
		})
		return domain.CheckoutSessionResult{}, errCheckoutStateNotPayable
	}

	var expiresAt *time.Time
	if result.ExpiresAt != "" {
		if ts, parseErr := parseTimestamp(result.ExpiresAt); parseErr == nil {
			expiresAt = &ts
		}
	}

	if markErr := s.repo.MarkPaymentAttemptCheckoutCreated(ctx, domain.PaymentAttemptCheckoutCreatedParams{
		TenantID:                tenantID,
		BranchID:                branchID,
		AttemptID:               attemptID,
		StripeCheckoutSessionID: result.CheckoutSessionID,
		StripeCheckoutURL:       result.CheckoutURL,
		StripePaymentIntentID:   result.PaymentIntentID,
		StripeExpiresAt:         expiresAt,
	}); markErr != nil {
		return domain.CheckoutSessionResult{}, fmt.Errorf("%w: %v", errCheckoutMarkCreated, markErr)
	}

	return result, nil
}

func isPayableCandidate(c domain.CheckoutInvoiceCandidate) bool {
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

func isStatePayable(s domain.InvoicePaymentState) bool {
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

func invoiceProductDesc(invoiceNumber string) string {
	if invoiceNumber == "" {
		return ""
	}
	return "Invoice " + invoiceNumber
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

	// Idempotency check: return existing active checkout session if one exists.
	if active, found, _ := uc.repo.GetActiveCheckoutForInvoice(ctx, tenantID, branchID, invoiceID.String()); found && active != nil {
		uc.logDebug("checkout_session_idempotent",
			"operation", "get_active_checkout_for_invoice",
			"request_id", requestID,
			"invoice_id", invoiceID.String(),
			"attempt_id", active.AttemptID,
		)
		return CreateCheckoutSessionResult{
			CheckoutSessionID: active.CheckoutSessionID,
			CheckoutURL:       active.CheckoutURL,
			PaymentAttemptID:  active.AttemptID,
		}, nil
	}

	steps := checkoutSessionSteps{repo: uc.repo, txMgr: uc.txMgr, provider: uc.provider, newUUID: uc.newUUID}
	candidate, attemptID, txErr := steps.createAttempt(ctx, tenantID, branchID, invoiceID.String(), requestID, userID, membershipID,
		func(ctx context.Context, tx domain.Tx) (domain.CheckoutInvoiceCandidate, bool, error) {
			return uc.repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, tenantID, branchID, membershipID, invoiceID.String())
		},
		uc.logDebug,
	)
	if txErr != nil {
		if errors.Is(txErr, errCheckoutInvoiceNotFound) {
			return CreateCheckoutSessionResult{}, domainerrors.NotFound("invoice", "Invoice not found.")
		}
		if errors.Is(txErr, errCheckoutInvoiceNotPayable) {
			return CreateCheckoutSessionResult{}, domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
		}
		return CreateCheckoutSessionResult{}, txErr
	}

	uc.recordTransition("parent_checkout", "payment_attempt", "none", "checkout_creation_started", "parent_checkout_requested")

	successURL := fmt.Sprintf("%s/parent/invoices/%s?checkout=success&session_id={CHECKOUT_SESSION_ID}", uc.webBaseURL, invoiceID.String())
	cancelURL := fmt.Sprintf("%s/parent/invoices/%s?checkout=cancelled", uc.webBaseURL, invoiceID.String())

	result, providerErr := steps.callProviderAndMark(ctx, tenantID, branchID, invoiceID.String(), attemptID.String(), requestID, candidate, successURL, cancelURL)
	if providerErr != nil {
		if errors.Is(providerErr, errCheckoutStateNotPayable) {
			uc.recordTransition("parent_checkout", "payment_attempt", "checkout_creation_started", "checkout_creation_failed", "invoice_no_longer_payable")
			uc.logDebug("checkout_session_state",
				"operation", "check_invoice_payment_state",
				"request_id", requestID,
				"invoice_id", invoiceID.String(),
				"attempt_id", attemptID.String(),
			)
			return CreateCheckoutSessionResult{}, domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")
		}
		if errors.Is(providerErr, errCheckoutMarkCreated) {
			return CreateCheckoutSessionResult{}, domainerrors.Internal(fmt.Errorf("mark payment attempt created: %w", providerErr))
		}
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
