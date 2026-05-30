package httpayment

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/application"
	"nursery-management-system/api/internal/platform/tenant"
	httpserver "nursery-management-system/api/internal/platform/http"
)

type CreateCheckoutSessionUseCase interface {
	Execute(ctx context.Context, tenantID, branchID, membershipID, userID, invoiceIDRaw, requestID string) (application.CreateCheckoutSessionResult, error)
}

type Handler struct {
	createCheckoutSession CreateCheckoutSessionUseCase
}

func NewHandler(createCheckoutSession *application.CreateCheckoutSession) *Handler {
	return &Handler{createCheckoutSession: createCheckoutSession}
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.POST("/invoices/:invoice_id/checkout-sessions", h.createCheckoutSessionHandler)
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

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	c.AbortWithStatusJSON(status, resp)
}
