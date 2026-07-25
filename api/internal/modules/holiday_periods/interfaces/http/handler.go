package httphp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/holiday_periods/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger *slog.Logger
	create *application.CreateHolidayPeriod
	update *application.UpdateHolidayPeriod
	delete *application.DeleteHolidayPeriod
	list   *application.ListHolidayPeriods
}

func NewHandler(
	create *application.CreateHolidayPeriod,
	update *application.UpdateHolidayPeriod,
	delete *application.DeleteHolidayPeriod,
	list *application.ListHolidayPeriods,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger: logger,
		create: create,
		update: update,
		delete: delete,
		list:   list,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.POST("/sites/:site_id/holiday-periods", h.createHolidayPeriod)
	protected.GET("/sites/:site_id/holiday-periods", h.listHolidayPeriods)
	protected.PATCH("/sites/:site_id/holiday-periods/:id", h.updateHolidayPeriod)
	protected.DELETE("/sites/:site_id/holiday-periods/:id", h.deleteHolidayPeriod)
}

func (h *Handler) resolveActor(c *gin.Context) (tenantID, branchID uuid.UUID, ok bool) {
	actor, actorOk := tenant.ActorFromGinContext(c)
	if !actorOk || actor.BranchID == uuid.Nil {
		return uuid.Nil, uuid.Nil, false
	}
	return actor.TenantID, actor.BranchID, true
}

func (h *Handler) createHolidayPeriod(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req createHolidayPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid start date format. Use YYYY-MM-DD.", nil)
		return
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid end date format. Use YYYY-MM-DD.", nil)
		return
	}

	params := application.CreateHolidayPeriodParams{
		Name:      req.Name,
		Type:      req.Type,
		StartDate: startDate,
		EndDate:   endDate,
	}

	hp, err := h.create.Execute(c.Request.Context(), tenantID, branchID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toHolidayPeriodResponse(hp)
	c.Header("Location", fmt.Sprintf("/api/sites/%s/holiday-periods/%s", branchID, resp.ID))
	c.JSON(http.StatusCreated, gin.H{"holiday_period": resp})
}

func (h *Handler) listHolidayPeriods(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	periods, err := h.list.Execute(c.Request.Context(), tenantID, branchID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": toHolidayPeriodListResponse(periods)})
}

func (h *Handler) updateHolidayPeriod(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid holiday period ID.", nil)
		return
	}

	var req updateHolidayPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params := application.UpdateHolidayPeriodParams{
		Name: req.Name,
		Type: req.Type,
	}
	if req.StartDate != nil {
		t, err := parseDate(*req.StartDate)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid start date format. Use YYYY-MM-DD.", nil)
			return
		}
		params.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := parseDate(*req.EndDate)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid end date format. Use YYYY-MM-DD.", nil)
			return
		}
		params.EndDate = &t
	}

	hp, err := h.update.Execute(c.Request.Context(), tenantID, branchID, id, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"holiday_period": toHolidayPeriodResponse(hp)})
}

func (h *Handler) deleteHolidayPeriod(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid holiday period ID.", nil)
		return
	}

	if err := h.delete.Execute(c.Request.Context(), tenantID, branchID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	httpserver.WriteMappedError(c, h.logger, err)
}
