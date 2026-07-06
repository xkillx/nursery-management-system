package httpclosure

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/branch_closures/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
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

// createClosureDay creates a new closure day for a site.
//
//	@Summary		Create closure day
//	@Description	Create a new closure day for a site.
//	@Tags			closure-days
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		createClosureDayRequest	true	"Closure day data"
//	@Success		201		{object}	object{closure_day=closureDayResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/sites/{site_id}/closure-days [post]
func (h *Handler) createClosureDay(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req createClosureDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	date, err := parseDate(req.Date)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid date format. Use YYYY-MM-DD.", nil)
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

// listClosureDays returns a paginated list of closure days for a site.
//
//	@Summary		List closure days
//	@Description	Get a paginated list of closure days for a site within a date range.
//	@Tags			closure-days
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"	format(uuid)
//	@Param			from		query		string	true	"From date"	format(date)
//	@Param			to			query		string	true	"To date"	format(date)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]closureDayResponse,total=int,page=int,page_size=int}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/sites/{site_id}/closure-days [get]
func (h *Handler) listClosureDays(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Both from and to query parameters are required.", nil)
		return
	}

	from, err := parseDate(fromStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid from date format. Use YYYY-MM-DD.", nil)
		return
	}
	to, err := parseDate(toStr)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid to date format. Use YYYY-MM-DD.", nil)
		return
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	closures, total, err := h.list.ExecutePaginated(c.Request.Context(), tenantID, branchID, from, to, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toClosureDayListResponse(closures), total, page, pageSize))
}

// deleteClosureDay deletes a closure day.
//
//	@Summary		Delete closure day
//	@Description	Delete a closure day.
//	@Tags			closure-days
//	@Produce		json
//	@Param			site_id	path	string	true	"Site ID"	format(uuid)
//	@Param			id		path	string	true	"Closure Day ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/sites/{site_id}/closure-days/{id} [delete]
func (h *Handler) deleteClosureDay(c *gin.Context) {
	tenantID, branchID, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid closure day ID.", nil)
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
