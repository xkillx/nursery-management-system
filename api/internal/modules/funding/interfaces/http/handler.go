package httpfunding

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/application"
	"nursery-management-system/api/internal/modules/funding/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger           *slog.Logger
	get              *application.GetProfile
	upsert           *application.UpsertProfile
	overview         *application.ListOverview
	enhancedOverview *application.GetEnhancedOverview
	enhancedDetail   *application.GetEnhancedChildDetail
	expiring         *application.ListExpiring
	parentFunding    *application.GetParentFunding
	parentBreakdown  *application.GetParentFundingBreakdown
}

func NewHandler(
	get *application.GetProfile,
	upsert *application.UpsertProfile,
	overview *application.ListOverview,
	enhancedOverview *application.GetEnhancedOverview,
	enhancedDetail *application.GetEnhancedChildDetail,
	expiring *application.ListExpiring,
	parentFunding *application.GetParentFunding,
	parentBreakdown *application.GetParentFundingBreakdown,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:           logger,
		get:              get,
		upsert:           upsert,
		overview:         overview,
		enhancedOverview: enhancedOverview,
		enhancedDetail:   enhancedDetail,
		expiring:         expiring,
		parentFunding:    parentFunding,
		parentBreakdown:  parentBreakdown,
	}
}

func (h *Handler) RegisterRoutes(manager *gin.RouterGroup) {
	g := manager.Group("/funding")
	g.GET("/overview", h.overviewHandler)
	g.GET("/children/:child_id", h.getProfileHandler)
	g.PUT("/children/:child_id", h.upsertProfileHandler)
	g.GET("/expiring", h.expiringHandler)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/funding", h.parentFundingHandler)
	parent.GET("/funding/:child_id/breakdown", h.parentFundingBreakdownHandler)
}

// parentFundingHandler returns funding entitlement and usage for the parent's children.
//
//	@Summary		Parent funding entitlement
//	@Description	Get funding entitlement and usage for the authenticated parent's children.
//	@Tags			parent-funding
//	@Produce		json
//	@Success		200	{object}	object{items=[]parentFundingEntitlementResponse}
//	@Failure		401	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/funding [get]
func (h *Handler) parentFundingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	results, err := h.parentFunding.Execute(c.Request.Context(), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]parentFundingEntitlementResponse, 0, len(results))
	for _, r := range results {
		items = append(items, parentFundingEntitlementResponse{
			ChildID:                r.ChildID.String(),
			ChildFirstName:         r.ChildFirstName,
			ChildMiddleName:        r.ChildMiddleName,
			ChildLastName:          r.ChildLastName,
			FundingType:            r.FundingType,
			FundedHoursPerWeek:     r.FundedHoursPerWeek,
			FundedAllowanceMinutes: r.FundedAllowanceMinutes,
			BookedHoursThisWeek:    r.BookedHoursThisWeek,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// parentFundingBreakdownHandler returns detailed funding breakdown for a child.
//
//	@Summary		Parent funding breakdown
//	@Description	Get detailed funding breakdown for a specific child.
//	@Tags			parent-funding
//	@Produce		json
//	@Param			child_id		path		string	true	"Child ID"		format(uuid)
//	@Param			billing_month	query		string	true	"Billing month"	format(month)
//	@Success		200				{object}	parentFundingBreakdownResponse
//	@Failure		400				{object}	object{code=string,message=string}
//	@Failure		401				{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["parent"]
//	@Router			/parent/funding/{child_id}/breakdown [get]
func (h *Handler) parentFundingBreakdownHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID, err := uuid.Parse(c.Param("child_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid child_id.", nil)
		return
	}

	billingMonth := c.Query("billing_month")
	if billingMonth == "" {
		billingMonth = time.Now().Format("2006-01")
	}

	result, err := h.parentBreakdown.Execute(c.Request.Context(), actor, childID, billingMonth)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toParentFundingBreakdownResponse(result))
}

// overviewHandler returns the funding overview for a billing month.
//
//	@Summary		Funding overview
//	@Description	Get the funding overview for a billing month, including enhanced metrics.
//	@Tags			funding
//	@Produce		json
//	@Param			billing_month	query		string	true	"Billing month"		format(month)
//	@Param			page			query		int		false	"Page number"		default(1)	minimum(1)
//	@Param			page_size		query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200				{object}	object{items=[]overviewItemResponse,total=int,page=int,page_size=int,metrics=enhancedOverviewMetricsResponse}
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

	metrics, err := h.enhancedOverview.Execute(c.Request.Context(), actor, billingMonth, 30)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     toOverviewResponse(result).Items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"metrics": enhancedOverviewMetricsResponse{
			TotalFundedChildren: metrics.TotalFundedChildren,
			FifteenHourCount:    metrics.FifteenHourCount,
			ThirtyHourCount:     metrics.ThirtyHourCount,
			BookedHoursThisWeek: metrics.BookedHoursThisWeek,
			ExpiringSoonCount:   metrics.ExpiringSoonCount,
		},
	})
}

// getProfileHandler returns the funding profile for a child with allocation and history.
//
//	@Summary		Get funding profile
//	@Description	Get the funding profile for a child for a billing month, including allocation table and history.
//	@Tags			funding
//	@Produce		json
//	@Param			child_id		path		string	true	"Child ID"		format(uuid)
//	@Param			billing_month	query		string	true	"Billing month"	format(month)
//	@Success		200				{object}	enhancedChildDetailResponse
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

	detail, err := h.enhancedDetail.Execute(c.Request.Context(), actor, c.Param("child_id"), billingMonth)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toEnhancedChildDetailResponse(detail))
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

// expiringHandler returns children with funding expiring within N days.
//
//	@Summary		Funding expiring soon
//	@Description	Get children with funding expiring within N days.
//	@Tags			funding
//	@Produce		json
//	@Param			within	query		int	true	"Number of days"	default(30)	minimum(1)
//	@Success		200		{object}	object{items=[]expiringFundingResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/funding/expiring [get]
func (h *Handler) expiringHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	withinStr := c.DefaultQuery("within", "30")
	within, err := strconv.Atoi(withinStr)
	if err != nil || within < 0 {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "within must be a non-negative integer.", nil)
		return
	}

	records, err := h.expiring.Execute(c.Request.Context(), actor, within)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]expiringFundingResponse, 0, len(records))
	for _, r := range records {
		items = append(items, expiringFundingResponse{
			FundingRecordID:    r.FundingRecordID.String(),
			ChildID:            r.ChildID.String(),
			ChildFirstName:     r.ChildFirstName,
			ChildMiddleName:    r.ChildMiddleName,
			ChildLastName:      r.ChildLastName,
			FundingType:        r.FundingType,
			FundedHoursPerWeek: r.FundedHoursPerWeek,
			FundingEndDate:     r.FundingEndDate.Format("2006-01-02"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
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
		ChildID:          row.ChildID.String(),
		ChildFirstName:   row.ChildFirstName,
		ChildMiddleName:  row.ChildMiddleName,
		ChildLastName:    row.ChildLastName,
		IsActive:         row.IsActive,
		StartDate:        row.StartDate,
		Flags:            flags,
		RemainingMinutes: item.RemainingMinutes,
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
	if row.ChildPhotoPath != nil {
		url := "/api/v1/children/" + row.ChildID.String() + "/photo"
		resp.ChildPhotoURL = &url
	}
	return resp
}

func toEnhancedChildDetailResponse(d domain.EnhancedChildDetail) enhancedChildDetailResponse {
	allocation := make([]allocationEntryResponse, 0, len(d.Allocation))
	for _, a := range d.Allocation {
		var endDate *string
		if a.EffectiveEndDate != nil {
			s := a.EffectiveEndDate.Format("2006-01-02")
			endDate = &s
		}
		allocation = append(allocation, allocationEntryResponse{
			BookingID:              a.BookingID.String(),
			EffectiveStartDate:     a.EffectiveStartDate.Format("2006-01-02"),
			EffectiveEndDate:       endDate,
			DaysOfWeek:             a.DaysOfWeek,
			SessionTypeName:        a.SessionTypeName,
			SessionDurationMinutes: a.SessionDurationMinutes,
		})
	}

	history := make([]fundingHistoryResponse, 0, len(d.History))
	for _, h := range d.History {
		var startDate *string
		if h.FundingStartDate != nil {
			s := h.FundingStartDate.Format("2006-01-02")
			startDate = &s
		}
		var endDate *string
		if h.FundingEndDate != nil {
			s := h.FundingEndDate.Format("2006-01-02")
			endDate = &s
		}
		history = append(history, fundingHistoryResponse{
			ID:                 h.ID.String(),
			FundingType:        h.FundingType,
			FundingModel:       h.FundingModel,
			FundedHoursPerWeek: h.FundedHoursPerWeek,
			FundingStartDate:   startDate,
			FundingEndDate:     endDate,
			ChangedAt:          h.ChangedAt,
		})
	}

	return enhancedChildDetailResponse{
		Profile:    toResponse(d.Profile),
		Allocation: allocation,
		History:    history,
	}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
