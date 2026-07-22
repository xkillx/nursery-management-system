package httpfunding

import (
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
	getChildFunding  *application.GetChildFunding
	updateFunding    *application.UpdateChildFunding
	overview         *application.ListOverview
	enhancedOverview *application.GetEnhancedOverview
	enhancedDetail   *application.GetEnhancedChildDetail
	expiring         *application.ListExpiring
	parentFunding    *application.GetParentFunding
	parentBreakdown  *application.GetParentFundingBreakdown
}

func NewHandler(
	getChildFunding *application.GetChildFunding,
	updateFunding *application.UpdateChildFunding,
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
		getChildFunding:  getChildFunding,
		updateFunding:    updateFunding,
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
	g.GET("/children/:child_id", h.getChildFundingHandler)
	g.PUT("/children/:child_id", h.updateFundingHandler)
	g.GET("/expiring", h.expiringHandler)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/funding", h.parentFundingHandler)
	parent.GET("/funding/:child_id/breakdown", h.parentFundingBreakdownHandler)
}

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

func (h *Handler) getChildFundingHandler(c *gin.Context) {
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

func (h *Handler) updateFundingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req fundingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params := application.UpdateChildFundingParams{
		FundingEnabled:           req.FundingEnabled,
		FundingType:              domain.FundingType(req.FundingType),
		FundingModel:             domain.FundingModel(req.FundingModel),
		FundedHoursPerWeek:       req.FundedHoursPerWeek,
		FundingStartDate:         req.FundingStartDate,
		FundingEndDate:           req.FundingEndDate,
		EligibilityCode:          req.EligibilityCode,
		EligibilityCodeValidated: req.EligibilityCodeValidated,
		EvidenceReceived:         req.EvidenceReceived,
	}

	record, err := h.updateFunding.Execute(c.Request.Context(), actor, c.Param("child_id"), params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toFundingRecordResponse(record))
}

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

func toFundingRecordResponse(r domain.FundingRecord) fundingRecordResponse {
	resp := fundingRecordResponse{
		ID:                       r.ID.String(),
		ChildID:                  r.ChildID.String(),
		FundingEnabled:           r.FundingEnabled,
		FundingType:              string(r.FundingType),
		FundingModel:             string(r.FundingModel),
		EligibilityCodeValidated: r.EligibilityCodeValidated,
		EvidenceReceived:         r.EvidenceReceived,
		CreatedAt:                r.CreatedAt,
		UpdatedAt:                r.UpdatedAt,
	}
	if r.FundedHoursPerWeek != nil {
		resp.FundedHoursPerWeek = r.FundedHoursPerWeek
	}
	if r.FundingStartDate != nil {
		s := r.FundingStartDate.Format("2006-01-02")
		resp.FundingStartDate = &s
	}
	if r.FundingEndDate != nil {
		s := r.FundingEndDate.Format("2006-01-02")
		resp.FundingEndDate = &s
	}
	if r.EligibilityCode != nil {
		resp.EligibilityCode = r.EligibilityCode
	}
	return resp
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
	if row.FundingRecordID != nil {
		resp.FundingRecordID = row.FundingRecordID.String()
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
		Record:                 toFundingRecordResponse(d.Record),
		FundedAllowanceMinutes: d.FundedAllowanceMinutes,
		Allocation:             allocation,
		History:                history,
	}
}

func toParentFundingBreakdownResponse(d application.ParentFundingBreakdown) parentFundingBreakdownResponse {
	return parentFundingBreakdownResponse{
		Record:                 toFundingRecordResponse(d.Record),
		FundedAllowanceMinutes: d.FundedAllowanceMinutes,
		Allocation:             toAllocationResponse(d.Allocation),
		History:                toHistoryResponse(d.History),
	}
}

func toAllocationResponse(allocation []domain.AllocationEntry) []allocationEntryResponse {
	out := make([]allocationEntryResponse, 0, len(allocation))
	for _, a := range allocation {
		var endDate *string
		if a.EffectiveEndDate != nil {
			s := a.EffectiveEndDate.Format("2006-01-02")
			endDate = &s
		}
		out = append(out, allocationEntryResponse{
			BookingID:              a.BookingID.String(),
			EffectiveStartDate:     a.EffectiveStartDate.Format("2006-01-02"),
			EffectiveEndDate:       endDate,
			DaysOfWeek:             a.DaysOfWeek,
			SessionTypeName:        a.SessionTypeName,
			SessionDurationMinutes: a.SessionDurationMinutes,
		})
	}
	return out
}

func toHistoryResponse(history []domain.FundingHistory) []fundingHistoryResponse {
	out := make([]fundingHistoryResponse, 0, len(history))
	for _, h := range history {
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
		out = append(out, fundingHistoryResponse{
			ID:                 h.ID.String(),
			FundingType:        h.FundingType,
			FundingModel:       h.FundingModel,
			FundedHoursPerWeek: h.FundedHoursPerWeek,
			FundingStartDate:   startDate,
			FundingEndDate:     endDate,
			ChangedAt:          h.ChangedAt,
		})
	}
	return out
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
