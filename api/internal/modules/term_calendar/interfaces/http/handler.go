package httptermcalendar

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term_calendar/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger  *slog.Logger
	create  *application.CreateTerm
	list    *application.ListTerms
	update  *application.UpdateTerm
	archive *application.ArchiveTerm
}

func NewHandler(
	create *application.CreateTerm,
	list *application.ListTerms,
	update *application.UpdateTerm,
	archive *application.ArchiveTerm,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:  logger,
		create:  create,
		list:    list,
		update:  update,
		archive: archive,
	}
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.GET("/sites/:site_id/academic-terms", h.listTerms)
	manager.POST("/sites/:site_id/academic-terms", h.createTerm)
	manager.PATCH("/sites/:site_id/academic-terms/:term_id", h.updateTerm)
	manager.POST("/sites/:site_id/academic-terms/:term_id/actions/archive", h.archiveTerm)
}

func (h *Handler) resolveActor(c *gin.Context) (application.TermCalendarActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerTermCalendarActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	if actor.BranchID == uuid.Nil {
		return nil, false
	}

	return application.NewManagerTermCalendarActor(actor), true
}

// listTerms returns a paginated list of academic terms for a site.
//
//	@Summary		List academic terms
//	@Description	Get a paginated list of academic terms for a site.
//	@Tags			academic-terms
//	@Produce		json
//	@Param			site_id				path		string	true	"Site ID"	format(uuid)
//	@Param			include_archived	query		bool	false	"Include archived terms"
//	@Param			page				query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size			query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200					{object}	object{items=[]academicTermResponse,total=int,page=int,page_size=int}
//	@Failure		401					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/academic-terms [get]
func (h *Handler) listTerms(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	includeArchived := c.Query("include_archived") == "true"

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	terms, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, siteID, includeArchived, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toTermListResponse(terms), total, page, pageSize))
}

// createTerm creates a new academic term for a site.
//
//	@Summary		Create academic term
//	@Description	Create a new academic term for a site.
//	@Tags			academic-terms
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string			true	"Site ID"	format(uuid)
//	@Param			body	body		createTermRequest	true	"Term data"
//	@Success		201		{object}	object{academic_term=academicTermResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/academic-terms [post]
func (h *Handler) createTerm(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	var req createTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid start_date format.", nil)
		return
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid end_date format.", nil)
		return
	}

	params := application.CreateTermParams{
		Name:      req.Name,
		Kind:      req.Kind,
		StartDate: startDate,
		EndDate:   endDate,
	}

	term, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	resp := toTermResponse(term)
	c.Header("Location", fmt.Sprintf("/api/sites/%s/academic-terms/%s", siteID, resp.ID))
	c.JSON(http.StatusCreated, gin.H{"academic_term": resp})
}

// updateTerm updates an existing academic term.
//
//	@Summary		Update academic term
//	@Description	Update an existing academic term.
//	@Tags			academic-terms
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string			true	"Site ID"	format(uuid)
//	@Param			term_id	path		string			true	"Term ID"	format(uuid)
//	@Param			body	body		updateTermRequest	true	"Term data"
//	@Success		200		{object}	object{academic_term=academicTermResponse}
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Failure		404		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/academic-terms/{term_id} [patch]
func (h *Handler) updateTerm(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	termID, err := uuid.Parse(c.Param("term_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	var req updateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params := application.UpdateTermParams{}

	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Kind != nil {
		params.Kind = req.Kind
	}
	if req.StartDate != nil {
		sd, err := parseDate(*req.StartDate)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid start_date format.", nil)
			return
		}
		params.StartDate = &sd
	}
	if req.EndDate != nil {
		ed, err := parseDate(*req.EndDate)
		if err != nil {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid end_date format.", nil)
			return
		}
		params.EndDate = &ed
	}

	term, err := h.update.Execute(c.Request.Context(), actor, siteID, termID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"academic_term": toTermResponse(term)})
}

// archiveTerm archives an academic term.
//
//	@Summary		Archive academic term
//	@Description	Archive an academic term.
//	@Tags			academic-terms
//	@Produce		json
//	@Param			site_id	path	string	true	"Site ID"	format(uuid)
//	@Param			term_id	path	string	true	"Term ID"	format(uuid)
//	@Success		204
//	@Failure		401	{object}	object{code=string,message=string}
//	@Failure		404	{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/academic-terms/{term_id}/actions/archive [post]
func (h *Handler) archiveTerm(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	termID, err := uuid.Parse(c.Param("term_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	if err := h.archive.Execute(c.Request.Context(), actor, siteID, termID); err != nil {
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
