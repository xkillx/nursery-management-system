package httpclosure

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger *slog.Logger
	create *application.CreateClosureDay
	list   *application.ListClosureDays
	delete *application.DeleteClosureDay
}

func NewHandler(
	create *application.CreateClosureDay,
	list *application.ListClosureDays,
	delete *application.DeleteClosureDay,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger: logger,
		create: create,
		list:   list,
		delete: delete,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.POST("/sites/:site_id/closure-days", h.createClosureDay)
	protected.GET("/sites/:site_id/closure-days", h.listClosureDays)
	protected.DELETE("/sites/:site_id/closure-days/:id", h.deleteClosureDay)
}

func (h *Handler) resolveActor(c *gin.Context) (tenantID, branchID uuid.UUID, ok bool) {
	actor, actorOk := tenant.ActorFromGinContext(c)
	if !actorOk || actor.BranchID == uuid.Nil {
		return uuid.Nil, uuid.Nil, false
	}
	return actor.TenantID, actor.BranchID, true
}

func (h *Handler) createClosureDay(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req createClosureDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	date, err := parseDate(req.Date)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid date format. Use YYYY-MM-DD.")
		return
	}

	params := application.CreateClosureDayParams{
		Date:   date,
		Reason: req.Reason,
	}

	closure, err := h.create.Execute(c.Request.Context(), tenantID, branchID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"closure_day": toClosureDayResponse(closure)})
}

func (h *Handler) listClosureDays(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Both from and to query parameters are required.")
		return
	}

	from, err := parseDate(fromStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid from date format. Use YYYY-MM-DD.")
		return
	}
	to, err := parseDate(toStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid to date format. Use YYYY-MM-DD.")
		return
	}

	closures, err := h.list.Execute(c.Request.Context(), tenantID, branchID, from, to)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"closure_days": toClosureDayListResponse(closures)})
}

func (h *Handler) deleteClosureDay(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid closure day ID.")
		return
	}

	if err := h.delete.Execute(c.Request.Context(), tenantID, branchID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
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
