package httpsessiontemplates

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontemplates/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger     *slog.Logger
	create     *application.CreateSessionTemplate
	update     *application.UpdateSessionTemplate
	list       *application.ListSessionTemplates
	get        *application.GetSessionTemplate
	archive    *application.ArchiveSessionTemplate
	reactivate *application.ReactivateSessionTemplate
}

func NewHandler(
	create *application.CreateSessionTemplate,
	update *application.UpdateSessionTemplate,
	list *application.ListSessionTemplates,
	get *application.GetSessionTemplate,
	archive *application.ArchiveSessionTemplate,
	reactivate *application.ReactivateSessionTemplate,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:     logger,
		create:     create,
		update:     update,
		list:       list,
		get:        get,
		archive:    archive,
		reactivate: reactivate,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	readOnly := protected.Group("")
	readOnly.Use(requireRoles("manager", "owner", "practitioner"))
	readOnly.GET("/sites/:site_id/session-templates", h.listTemplates)
	readOnly.GET("/sites/:site_id/session-templates/:template_id", h.getTemplate)

	writeOps := protected.Group("")
	writeOps.Use(requireRoles("manager", "owner"))
	writeOps.POST("/sites/:site_id/session-templates", h.createTemplate)
	writeOps.PATCH("/sites/:site_id/session-templates/:template_id", h.updateTemplate)
	writeOps.POST("/sites/:site_id/session-templates/:template_id/actions/archive", h.archiveTemplate)
	writeOps.POST("/sites/:site_id/session-templates/:template_id/actions/reactivate", h.reactivateTemplate)
}

func (h *Handler) resolveActor(c *gin.Context) (application.SessionTemplateActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerSessionTemplateActor(owner), true
	}

	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		return nil, false
	}

	if actor.BranchID == uuid.Nil {
		return nil, false
	}

	role := ""
	if v, authOk := c.Get(tenant.AuthContextKey); authOk {
		if authCtx, authOk := v.(tenant.AuthorizationContext); authOk {
			role = authCtx.Role
		}
	}

	switch role {
	case "manager":
		return application.NewManagerSessionTemplateActor(actor), true
	case "practitioner":
		return application.NewPractitionerSessionTemplateActor(actor), true
	}

	return nil, false
}

// listTemplates returns a paginated list of session templates for a site.
//
//	@Summary		List session templates
//	@Description	Get a paginated list of session templates for a site.
//	@Tags			session-templates
//	@Produce		json
//	@Param			site_id				path		string	true	"Site ID"	format(uuid)
//	@Param			include_archived	query		bool	false	"Include archived templates"
//	@Param			page				query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size			query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200					{object}	object{items=[]sessionTemplateResponse,total=int,page=int,page_size=int}
//	@Failure		401					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/session-templates [get]
func (h *Handler) listTemplates(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	includeArchived := c.Query("include_archived") == "true"

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	templates, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, siteID, includeArchived, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toSessionTemplateListResponse(templates), total, page, pageSize))
}

// getTemplate returns a single session template by ID.
//
//	@Summary		Get session template
//	@Description	Get a single session template by ID.
//	@Tags			session-templates
//	@Produce		json
//	@Param			site_id		path		string	true	"Site ID"		format(uuid)
//	@Param			template_id	path		string	true	"Template ID"	format(uuid)
//	@Success		200			{object}	sessionTemplateResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/session-templates/{template_id} [get]
func (h *Handler) getTemplate(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	templateID, err := uuid.Parse(c.Param("template_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	t, err := h.get.Execute(c.Request.Context(), actor, siteID, templateID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTemplateResponse(t))
}

// createTemplate creates a new session template for a site.
//
//	@Summary		Create session template
//	@Description	Create a new session template for a site.
//	@Tags			session-templates
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string						true	"Site ID"	format(uuid)
//	@Param			body	body		createSessionTemplateRequest	true	"Template data"
//	@Success		201		{object}	sessionTemplateResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-templates [post]
func (h *Handler) createTemplate(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req createSessionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	entries := make([]application.SessionTemplateEntryInput, 0, len(req.Entries))
	for _, e := range req.Entries {
		stID, perr := uuid.Parse(e.SessionTypeID)
		if perr != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
			return
		}
		entries = append(entries, application.SessionTemplateEntryInput{
			DayOfWeek:     e.DayOfWeek,
			SessionTypeID: stID,
		})
	}

	params := application.CreateSessionTemplateParams{
		Name:        req.Name,
		Description: req.Description,
		Entries:     entries,
	}

	t, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSessionTemplateResponse(t))
}

// updateTemplate updates an existing session template.
//
//	@Summary		Update session template
//	@Description	Update an existing session template.
//	@Tags			session-templates
//	@Accept			json
//	@Produce		json
//	@Param			site_id		path		string						true	"Site ID"		format(uuid)
//	@Param			template_id	path		string						true	"Template ID"	format(uuid)
//	@Param			body		body		updateSessionTemplateRequest	true	"Template data"
//	@Success		200			{object}	sessionTemplateResponse
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-templates/{template_id} [patch]
func (h *Handler) updateTemplate(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	templateID, err := uuid.Parse(c.Param("template_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req updateSessionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.UpdateSessionTemplateParams{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.Entries != nil {
		entries := make([]application.SessionTemplateEntryInput, 0, len(*req.Entries))
		for _, e := range *req.Entries {
			stID, perr := uuid.Parse(e.SessionTypeID)
			if perr != nil {
				writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
				return
			}
			entries = append(entries, application.SessionTemplateEntryInput{
				DayOfWeek:     e.DayOfWeek,
				SessionTypeID: stID,
			})
		}
		params.Entries = &entries
	}

	t, err := h.update.Execute(c.Request.Context(), actor, siteID, templateID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTemplateResponse(t))
}

// archiveTemplate archives a session template.
//
//	@Summary		Archive session template
//	@Description	Archive a session template.
//	@Tags			session-templates
//	@Produce		json
//	@Param			site_id		path	string	true	"Site ID"		format(uuid)
//	@Param			template_id	path	string	true	"Template ID"	format(uuid)
//	@Success		200			{object}	object
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-templates/{template_id}/actions/archive [post]
func (h *Handler) archiveTemplate(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	templateID, err := uuid.Parse(c.Param("template_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if err := h.archive.Execute(c.Request.Context(), actor, siteID, templateID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// reactivateTemplate reactivates an archived session template.
//
//	@Summary		Reactivate session template
//	@Description	Reactivate an archived session template.
//	@Tags			session-templates
//	@Produce		json
//	@Param			site_id		path	string	true	"Site ID"		format(uuid)
//	@Param			template_id	path	string	true	"Template ID"	format(uuid)
//	@Success		200			{object}	sessionTemplateResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-templates/{template_id}/actions/reactivate [post]
func (h *Handler) reactivateTemplate(c *gin.Context) {
	actor, ok := h.resolveActor(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	templateID, err := uuid.Parse(c.Param("template_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	t, err := h.reactivate.Execute(c.Request.Context(), actor, siteID, templateID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTemplateResponse(t))
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

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		v, ok := c.Get(tenant.AuthContextKey)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
			return
		}

		authCtx, ok := v.(tenant.AuthorizationContext)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
			return
		}

		switch authCtx.Role {
		case "owner", "manager", "practitioner", "parent":
		default:
			writeError(c, http.StatusForbidden, "forbidden_role_unknown", "Access denied.")
			return
		}

		if _, exists := allowed[authCtx.Role]; !exists {
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.")
			return
		}

		c.Next()
	}
}
