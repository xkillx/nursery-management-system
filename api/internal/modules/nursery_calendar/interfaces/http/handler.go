package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/nursery_calendar/application"
	"nursery-management-system/api/internal/modules/nursery_calendar/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	queryDay   *application.QueryCalendarDay
	queryRange *application.QueryDateRange
	logger     *slog.Logger
}

func NewHandler(queryDay *application.QueryCalendarDay, queryRange *application.QueryDateRange, logger *slog.Logger) *Handler {
	return &Handler{queryDay: queryDay, queryRange: queryRange, logger: logger}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/sites/:site_id/calendar/check", h.checkDateHandler)
	protected.GET("/sites/:site_id/calendar/range", h.dateRangeHandler)
}

func (h *Handler) RegisterParentRoutes(parent *gin.RouterGroup) {
	parent.GET("/closure-days", h.parentClosureDaysHandler)
	parent.GET("/holiday-periods", h.parentHolidayPeriodsHandler)
	parent.GET("/calendar/check", h.parentCheckDateHandler)
	parent.GET("/calendar/range", h.parentDateRangeHandler)
}

func (h *Handler) checkDateHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := parseUUID(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site ID.", nil)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Date parameter is required.", nil)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid date format. Use YYYY-MM-DD.", nil)
		return
	}

	isTermTime := c.Query("is_term_time") == "true"

	result, err := h.queryDay.Execute(c.Request.Context(), actor.TenantID, siteID, date, isTermTime)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCalendarDayResponse(result.Date, result.IsOpen, string(result.Reason)))
}

func (h *Handler) dateRangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := parseUUID(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid site ID.", nil)
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Both 'from' and 'to' date parameters are required.", nil)
		return
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid 'from' date format. Use YYYY-MM-DD.", nil)
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid 'to' date format. Use YYYY-MM-DD.", nil)
		return
	}

	isTermTime := c.Query("is_term_time") == "true"

	results, err := h.queryRange.Execute(c.Request.Context(), actor.TenantID, siteID, from, to, isTermTime)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]calendarDayResponse, 0, len(results))
	for _, r := range results {
		items = append(items, toCalendarDayResponse(r.Date, r.IsOpen, string(r.Reason)))
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) parentClosureDaysHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	from, to := h.resolveDateRange(c)

	results, err := h.queryRange.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, from, to, false)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var items []calendarDayResponse
	for _, r := range results {
		if r.Reason == domain.ClosureReasonClosureDay {
			items = append(items, toCalendarDayResponse(r.Date, r.IsOpen, string(r.Reason)))
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) parentHolidayPeriodsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	from, to := h.resolveDateRange(c)

	results, err := h.queryRange.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, from, to, true)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var items []calendarDayResponse
	for _, r := range results {
		if r.Reason == domain.ClosureReasonHolidayPeriod {
			items = append(items, toCalendarDayResponse(r.Date, r.IsOpen, string(r.Reason)))
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) parentCheckDateHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Date parameter is required.", nil)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid date format. Use YYYY-MM-DD.", nil)
		return
	}

	isTermTime := c.Query("is_term_time") == "true"

	result, err := h.queryDay.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, date, isTermTime)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCalendarDayResponse(result.Date, result.IsOpen, string(result.Reason)))
}

func (h *Handler) parentDateRangeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	from, to := h.resolveDateRange(c)
	isTermTime := c.Query("is_term_time") == "true"

	results, err := h.queryRange.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, from, to, isTermTime)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]calendarDayResponse, 0, len(results))
	for _, r := range results {
		items = append(items, toCalendarDayResponse(r.Date, r.IsOpen, string(r.Reason)))
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) resolveDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	defaultFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	defaultTo := defaultFrom.AddDate(0, 3, 0).AddDate(0, 0, -1)

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			defaultFrom = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			defaultTo = t
		}
	}

	return defaultFrom, defaultTo
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
	if err != nil {
		h.logger.Error(fmt.Sprintf("nursery_calendar: %v", err))
	}
}
