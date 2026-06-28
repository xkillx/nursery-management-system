package httpfunding

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/funding/application"
	"nursery-management-system/api/internal/modules/funding/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger   *slog.Logger
	get      *application.GetProfile
	upsert   *application.UpsertProfile
	overview *application.ListOverview
}

func NewHandler(get *application.GetProfile, upsert *application.UpsertProfile, overview *application.ListOverview, logger *slog.Logger) *Handler {
	return &Handler{logger: logger, get: get, upsert: upsert, overview: overview}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	g := manager.Group("/funding")
	g.GET("/overview", h.overviewHandler)
	g.GET("/children/:child_id", h.getProfileHandler)
	g.PUT("/children/:child_id", h.upsertProfileHandler)
}

func (h *Handler) overviewHandler(c *gin.Context) {
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

	result, err := h.overview.Execute(c.Request.Context(), actor, billingMonth)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toOverviewResponse(result))
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
		h.handleError(c, err)
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
		h.handleError(c, err)
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

func toOverviewResponse(r domain.OverviewResult) overviewResponse {
	items := make([]overviewItemResponse, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, toOverviewItemResponse(item))
	}
	return overviewResponse{
		BillingMonth: r.BillingMonth.Format("2006-01"),
		Summary: overviewSummaryResponse{
			IncludedChildCount:  r.Summary.IncludedChildCount,
			FlaggedChildCount:   r.Summary.FlaggedChildCount,
			MissingProfileCount: r.Summary.MissingProfileCount,
			ExplicitZeroCount:   r.Summary.ExplicitZeroCount,
			UnderOneHourCount:   r.Summary.UnderOneHourCount,
			Above160HoursCount:  r.Summary.Above160HoursCount,
		},
		Items: items,
	}
}

func toOverviewItemResponse(item domain.OverviewItem) overviewItemResponse {
	row := item.Row
	flags := make([]string, 0, len(item.Flags))
	for _, f := range item.Flags {
		flags = append(flags, string(f))
	}

	resp := overviewItemResponse{
		ChildID:         row.ChildID.String(),
		ChildFirstName:  row.ChildFirstName,
		ChildMiddleName: row.ChildMiddleName,
		ChildLastName:   row.ChildLastName,
		IsActive:        row.IsActive,
		StartDate:       row.StartDate,
		Flags:           flags,
	}
	if row.EndDate != nil {
		resp.EndDate = row.EndDate
	}
	if row.FundingProfileID != nil {
		resp.FundingProfileID = row.FundingProfileID.String()
	}
	if row.FundedAllowanceMinutes != nil {
		resp.FundedAllowanceMinutes = row.FundedAllowanceMinutes
	}
	if row.FundingUpdatedAt != nil {
		resp.FundingUpdatedAt = row.FundingUpdatedAt
	}
	return resp
}

func (h *Handler) handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
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
