package httpfunding

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/funding/application"
	"nursery-management-system/api/internal/modules/funding/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	get    *application.GetProfile
	upsert *application.UpsertProfile
}

func NewHandler(get *application.GetProfile, upsert *application.UpsertProfile) *Handler {
	return &Handler{get: get, upsert: upsert}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	g := manager.Group("/funding")
	g.GET("/children/:child_id", h.getProfileHandler)
	g.PUT("/children/:child_id", h.upsertProfileHandler)
}

func (h *Handler) getProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	billingMonth := c.Query("billing_month")
	if billingMonth == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "billing_month query parameter is required.")
		return
	}

	profile, err := h.get.Execute(c.Request.Context(), actor, c.Param("child_id"), billingMonth)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(profile))
}

func (h *Handler) upsertProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req fundingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	result, err := h.upsert.Execute(c.Request.Context(), actor, c.Param("child_id"), application.UpsertProfileParams{
		BillingMonth:           req.BillingMonth,
		FundedAllowanceMinutes: req.FundedAllowanceMinutes,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	if result.Created {
		c.JSON(http.StatusCreated, toResponse(result.Profile))
	} else {
		c.JSON(http.StatusOK, toResponse(result.Profile))
	}
}

func toResponse(p domain.FundingProfile) fundingProfileResponse {
	return fundingProfileResponse{
		ID:                     p.ID.String(),
		ChildID:                p.ChildID.String(),
		BillingMonth:           p.BillingMonth.Format("2006-01"),
		FundedAllowanceMinutes: p.FundedAllowanceMinutes,
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	c.AbortWithStatusJSON(status, resp)
}

func writeError(c *gin.Context, status int, code, message string) {
	requestID := httpserver.RequestIDFromContext(c)
	c.AbortWithStatusJSON(status, httpserver.ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}
