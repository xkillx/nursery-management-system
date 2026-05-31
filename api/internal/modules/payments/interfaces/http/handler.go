package httpayment

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/application"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	httpserver "nursery-management-system/api/internal/platform/http"
)

type CreateCheckoutSessionUseCase interface {
	Execute(ctx context.Context, tenantID, branchID, membershipID, userID, invoiceIDRaw, requestID string) (application.CreateCheckoutSessionResult, error)
}

type HandleWebhookUseCase interface {
	Execute(ctx context.Context, payload []byte, signatureHeader, requestID string) (application.WebhookResult, error)
}

type Handler struct {
	createCheckoutSession CreateCheckoutSessionUseCase
	handleWebhook         HandleWebhookUseCase
}

func NewHandler(createCheckoutSession *application.CreateCheckoutSession, handleWebhook *application.HandleStripeWebhook) *Handler {
	var hw HandleWebhookUseCase
	if handleWebhook != nil {
		hw = handleWebhook
	}
	return &Handler{
		createCheckoutSession: createCheckoutSession,
		handleWebhook:         hw,
	}
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.POST("/invoices/:invoice_id/checkout-sessions", h.createCheckoutSessionHandler)
}

func (h *Handler) RegisterStripeRoutes(api *gin.RouterGroup) {
	api.POST("/stripe/webhooks", h.stripeWebhookHandler)
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
