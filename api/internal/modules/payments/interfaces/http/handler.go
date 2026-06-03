package httpayment

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/application"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
)

type CreateCheckoutSessionUseCase interface {
	Execute(ctx context.Context, tenantID, branchID, membershipID, userID, invoiceIDRaw, requestID string) (application.CreateCheckoutSessionResult, error)
}

type HandleWebhookUseCase interface {
	Execute(ctx context.Context, payload []byte, signatureHeader, requestID string) (application.WebhookResult, error)
}

type GetManagerPaymentStatusUseCase interface {
	Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw string) (application.ManagerPaymentStatusResult, error)
}

type ListManagerPaymentEventsUseCase interface {
	Execute(ctx context.Context, actor tenant.ActorContext, invoiceIDRaw, limitRaw, offsetRaw string) (application.ListPaymentEventsResult, error)
}

type Handler struct {
	createCheckoutSession CreateCheckoutSessionUseCase
	handleWebhook         HandleWebhookUseCase
	getManagerStatus      GetManagerPaymentStatusUseCase
	listManagerEvents     ListManagerPaymentEventsUseCase
	logger                *slog.Logger
	recorder              *metrics.Recorder
}

func NewHandler(
	createCheckoutSession *application.CreateCheckoutSession,
	handleWebhook *application.HandleStripeWebhook,
	getManagerStatus *application.GetManagerPaymentStatus,
	listManagerEvents *application.ListManagerPaymentEvents,
) *Handler {
	var hw HandleWebhookUseCase
	if handleWebhook != nil {
		hw = handleWebhook
	}
	return &Handler{
		createCheckoutSession: createCheckoutSession,
		handleWebhook:         hw,
		getManagerStatus:      getManagerStatus,
		listManagerEvents:     listManagerEvents,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *Handler {
	return &Handler{
		createCheckoutSession: h.createCheckoutSession,
		handleWebhook:         h.handleWebhook,
		getManagerStatus:      h.getManagerStatus,
		listManagerEvents:     h.listManagerEvents,
		logger:                logger,
		recorder:              recorder,
	}
}

func (h *Handler) recordWebhookOutcome(provider, eventType, outcome, reason string) {
	if h.recorder != nil {
		h.recorder.WebhookOutcome(provider, eventType, outcome, reason)
	}
}

func (h *Handler) logInfo(msg string, args ...any) {
	if h.logger != nil {
		h.logger.Info(msg, args...)
	}
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.POST("/invoices/:invoice_id/checkout-sessions", h.createCheckoutSessionHandler)
}

func (h *Handler) RegisterStripeRoutes(api *gin.RouterGroup) {
	api.POST("/stripe/webhooks", h.stripeWebhookHandler)
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.GET("/invoices/:invoice_id/payment-status", h.managerPaymentStatusHandler)
	manager.GET("/invoices/:invoice_id/payment-events", h.managerPaymentEventsHandler)
}

func (h *Handler) managerPaymentStatusHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	invoiceIDRaw := c.Param("invoice_id")
	if _, err := uuid.Parse(invoiceIDRaw); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid invoice ID format.", map[string]string{"field": "invoice_id"})
		return
	}

	result, err := h.getManagerStatus.Execute(c.Request.Context(), actor, invoiceIDRaw)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toManagerPaymentStatusResponse(result))
}

func (h *Handler) managerPaymentEventsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	invoiceIDRaw := c.Param("invoice_id")
	if _, err := uuid.Parse(invoiceIDRaw); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid invoice ID format.", map[string]string{"field": "invoice_id"})
		return
	}

	limitRaw := c.Query("limit")
	offsetRaw := c.Query("offset")

	result, err := h.listManagerEvents.Execute(c.Request.Context(), actor, invoiceIDRaw, limitRaw, offsetRaw)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toManagerPaymentEventsResponse(result))
}

func (h *Handler) createCheckoutSessionHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	invoiceIDRaw := c.Param("invoice_id")
	if _, err := uuid.Parse(invoiceIDRaw); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid invoice ID format.", map[string]string{"field": "invoice_id"})
		return
	}

	result, err := h.createCheckoutSession.Execute(
		c.Request.Context(),
		actor.TenantID.String(),
		actor.BranchID.String(),
		actor.MembershipID.String(),
		actor.UserID.String(),
		invoiceIDRaw,
		actor.RequestID,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createCheckoutSessionResponse{
		CheckoutSessionID: result.CheckoutSessionID,
		CheckoutURL:       result.CheckoutURL,
		PaymentAttemptID:  result.PaymentAttemptID,
	})
}

func (h *Handler) stripeWebhookHandler(c *gin.Context) {
	if h.handleWebhook == nil {
		requestID := httpserver.RequestIDFromContext(c)
		h.recordWebhookOutcome("stripe", "unknown", "unconfigured", "payment_provider_unconfigured")
		h.logInfo("stripe_webhook",
			"operation", "stripe_webhook",
			"outcome", "unconfigured",
			"reason", "payment_provider_unconfigured",
			"request_id", requestID,
		)
		status, resp := httpserver.MapDomainError(
			domainErrorPaymentProviderUnconfigured(),
			requestID,
		)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64*1024)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Request body too large or unreadable.", nil)
		return
	}

	requestID := httpserver.RequestIDFromContext(c)
	result, err := h.handleWebhook.Execute(c.Request.Context(), payload, sigHeader, requestID)
	if err != nil {
		if isInvalidSignatureError(err) {
			eventType := "unknown"
			h.recordWebhookOutcome("stripe", eventType, "invalid_signature", "payment_webhook_invalid_signature")
		}
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, webhookResponse{Status: result.Status})
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	c.AbortWithStatusJSON(status, resp)
}

func domainErrorPaymentProviderUnconfigured() *domainerrors.DomainError {
	return domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")
}

func isInvalidSignatureError(err error) bool {
	if de, ok := err.(*domainerrors.DomainError); ok {
		return de.Code == "payment_webhook_invalid_signature"
	}
	return false
}

func toManagerPaymentStatusResponse(r application.ManagerPaymentStatusResult) managerPaymentStatusResponse {
	resp := managerPaymentStatusResponse{
		InvoiceID:               r.InvoiceID,
		InvoiceKind:             r.InvoiceKind,
		InvoiceNumber:           r.InvoiceNumber,
		InvoiceNumberDisplay:    r.InvoiceNumberDisplay,
		ChildID:                 r.ChildID,
		ChildName:               r.ChildName,
		BillingMonth:            r.BillingMonth,
		Status:                  r.Status,
		DueStatus:               r.DueStatus,
		CurrencyCode:            r.CurrencyCode,
		TotalDueMinor:           r.TotalDueMinor,
		AmountPaidMinor:         r.AmountPaidMinor,
		IssuedAt:                r.IssuedAt,
		DueAt:                   r.DueAt,
		PaidAt:                  r.PaidAt,
		PaymentFailedAt:         r.PaymentFailedAt,
		PaymentStatusUpdatedAt:  r.PaymentStatusUpdatedAt,
		CheckoutRetryAvailable:  r.CheckoutRetryAvailable,
		CheckoutRetryReasonCode: r.CheckoutRetryReasonCode,
	}
	if r.LatestPaymentAttempt != nil {
		resp.LatestPaymentAttempt = &paymentAttemptDiagnosticDTO{
			PaymentAttemptID:        r.LatestPaymentAttempt.PaymentAttemptID,
			Status:                  r.LatestPaymentAttempt.Status,
			AmountMinor:             r.LatestPaymentAttempt.AmountMinor,
			CurrencyCode:            r.LatestPaymentAttempt.CurrencyCode,
			StripeCheckoutSessionID: r.LatestPaymentAttempt.StripeCheckoutSessionID,
			StripePaymentIntentID:   r.LatestPaymentAttempt.StripePaymentIntentID,
			StripeExpiresAt:         r.LatestPaymentAttempt.StripeExpiresAt,
			FailureReason:           r.LatestPaymentAttempt.FailureReason,
			ProviderErrorCode:       r.LatestPaymentAttempt.ProviderErrorCode,
			ProviderErrorMessage:    r.LatestPaymentAttempt.ProviderErrorMessage,
			CreatedAt:               r.LatestPaymentAttempt.CreatedAt,
			UpdatedAt:               r.LatestPaymentAttempt.UpdatedAt,
		}
	}
	if r.LatestPaymentEvent != nil {
		resp.LatestPaymentEvent = toPaymentEventDTO(r.LatestPaymentEvent)
	}
	return resp
}

func toPaymentEventDTO(e *application.PaymentEventResult) *paymentEventDiagnosticDTO {
	return &paymentEventDiagnosticDTO{
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
		WebhookReceivedAt:       e.WebhookReceivedAt,
		WebhookProcessedAt:      e.WebhookProcessedAt,
		CreatedAt:               e.CreatedAt,
	}
}

func toManagerPaymentEventsResponse(r application.ListPaymentEventsResult) managerPaymentEventsResponse {
	items := make([]paymentEventDiagnosticDTO, 0, len(r.Items))
	for i := range r.Items {
		items = append(items, *toPaymentEventDTO(&r.Items[i]))
	}
	return managerPaymentEventsResponse{
		Items:  items,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}
