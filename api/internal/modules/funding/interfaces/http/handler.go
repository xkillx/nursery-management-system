package httpfunding

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/funding/application"
	"nursery-management-system/api/internal/modules/funding/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
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

// overviewHandler returns the funding overview for a billing month.
//
//	@Summary		Funding overview
//	@Description	Get the funding overview for a billing month.
//	@Tags			funding
//	@Produce		json
//	@Param			billing_month	query		string	true	"Billing month"	format(month)
//	@Param			page			query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size		query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200				{object}	object{items=[]overviewItemResponse,total=int,page=int,page_size=int}
//	@Failure		400				{object}	object{code=string,message=string}
//	@Failure		401				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/funding/overview [get]
func (h *Handler) overviewHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	billingMonth := c.Query("billing_month")
	if billingMonth == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "billing_month query parameter is required.", nil)
		return
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	result, total, err := h.overview.ExecutePaginated(c.Request.Context(), actor, billingMonth, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toOverviewResponse(result), total, page, pageSize))
}

// getProfileHandler returns the funding profile for a child.
//
//	@Summary		Get funding profile
//	@Description	Get the funding profile for a child for a billing month.
//	@Tags			funding
//	@Produce		json
//	@Param			child_id		path		string	true	"Child ID"		format(uuid)
//	@Param			billing_month	query		string	true	"Billing month"	format(month)
//	@Success		200				{object}	fundingProfileResponse
//	@Failure		400				{object}	object{code=string,message=string}
//	@Failure		401				{object}	object{code=string,message=string}
//	@Failure		404				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/funding/children/{child_id} [get]
func (h *Handler) getProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	billingMonth := c.Query("billing_month")
	if billingMonth == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "billing_month query parameter is required.", nil)
		return
	}

	profile, err := h.get.Execute(c.Request.Context(), actor, c.Param("child_id"), billingMonth)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(profile))
}

// upsertProfileHandler creates or updates the funding profile for a child.
//
//	@Summary		Upsert funding profile
//	@Description	Create or update the funding profile for a child.
//	@Tags			funding
//	@Accept			json
//	@Produce		json
//	@Param			child_id	path		string					true	"Child ID"	format(uuid)
//	@Param			body		body		fundingProfileRequest	true	"Funding profile data"
//	@Success		200			{object}	fundingProfileResponse
//	@Success		201			{object}	fundingProfileResponse
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/funding/children/{child_id} [put]
func (h *Handler) upsertProfileHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req fundingProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
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
		resp := toResponse(result.Profile)
		c.Header("Location", fmt.Sprintf("/api/children/%s/funding/%s", resp.ChildID, resp.ID))
		c.JSON(http.StatusCreated, resp)
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
	httpserver.WriteMappedError(c, h.logger, err)
}
