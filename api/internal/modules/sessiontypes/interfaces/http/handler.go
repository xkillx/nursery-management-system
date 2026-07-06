package httpsessiontypes

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontypes/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger     *slog.Logger
	create     *application.CreateSessionType
	update     *application.UpdateSessionType
	list       *application.ListSessionTypes
	get        *application.GetSessionType
	archive    *application.ArchiveSessionType
	reactivate *application.ReactivateSessionType
}

func NewHandler(
	create *application.CreateSessionType,
	update *application.UpdateSessionType,
	list *application.ListSessionTypes,
	get *application.GetSessionType,
	archive *application.ArchiveSessionType,
	reactivate *application.ReactivateSessionType,
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
	readOnly.GET("/sites/:site_id/session-types", h.listSessionTypes)
	readOnly.GET("/sites/:site_id/session-types/:session_type_id", h.getSessionType)

	writeOps := protected.Group("")
	writeOps.Use(requireRoles("manager", "owner"))
	writeOps.POST("/sites/:site_id/session-types", h.createSessionType)
	writeOps.PATCH("/sites/:site_id/session-types/:session_type_id", h.updateSessionType)
	writeOps.POST("/sites/:site_id/session-types/:session_type_id/actions/archive", h.archiveSessionType)
	writeOps.POST("/sites/:site_id/session-types/:session_type_id/actions/activate", h.reactivateSessionType)
}

func (h *Handler) resolveActor(c *gin.Context) (application.SessionTypeActor, bool) {
	if owner, ok := tenant.OwnerActorFromGinContext(c); ok {
		return application.NewOwnerSessionTypeActor(owner), true
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
		return application.NewManagerSessionTypeActor(actor), true
	case "practitioner":
		return application.NewPractitionerSessionTypeActor(actor), true
	}

	return nil, false
}

// listSessionTypes returns a paginated list of session types for a site.
//
//	@Summary		List session types
//	@Description	Get a paginated list of session types for a site.
//	@Tags			session-types
//	@Produce		json
//	@Param			site_id				path		string	true	"Site ID"	format(uuid)
//	@Param			include_archived	query		bool	false	"Include archived session types"
//	@Param			page				query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size			query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200					{object}	object{items=[]sessionTypeResponse,total=int,page=int,page_size=int}
//	@Failure		401					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/session-types [get]
func (h *Handler) listSessionTypes(c *gin.Context) {
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

	types, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, siteID, includeArchived, pageSize, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toSessionTypeListResponse(types), total, page, pageSize))
}

// createSessionType creates a new session type for a site.
//
//	@Summary		Create session type
//	@Description	Create a new session type for a site.
//	@Tags			session-types
//	@Accept			json
//	@Produce		json
//	@Param			site_id	path		string					true	"Site ID"	format(uuid)
//	@Param			body	body		createSessionTypeRequest	true	"Session type data"
//	@Success		201		{object}	sessionTypeResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-types [post]
func (h *Handler) createSessionType(c *gin.Context) {
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

	var req createSessionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params := application.CreateSessionTypeParams{
		Name:         req.Name,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Kind:         req.Kind,
		FlatFeeMinor: req.FlatFeeMinor,
	}

	st, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSessionTypeResponse(st))
}

// getSessionType returns a single session type by ID.
//
//	@Summary		Get session type
//	@Description	Get a single session type by ID.
//	@Tags			session-types
//	@Produce		json
//	@Param			site_id				path		string	true	"Site ID"			format(uuid)
//	@Param			session_type_id		path		string	true	"Session Type ID"	format(uuid)
//	@Success		200					{object}	sessionTypeResponse
//	@Failure		401					{object}	object{code=string,message=string}
//	@Failure		404					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner","practitioner"]
//	@Router			/sites/{site_id}/session-types/{session_type_id} [get]
func (h *Handler) getSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	st, err := h.get.Execute(c.Request.Context(), actor, siteID, stID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTypeResponse(st))
}

// updateSessionType updates an existing session type.
//
//	@Summary		Update session type
//	@Description	Update an existing session type.
//	@Tags			session-types
//	@Accept			json
//	@Produce		json
//	@Param			site_id				path		string						true	"Site ID"			format(uuid)
//	@Param			session_type_id		path		string						true	"Session Type ID"	format(uuid)
//	@Param			body				body		updateSessionTypeRequest	true	"Session type data"
//	@Success		200					{object}	sessionTypeResponse
//	@Failure		400					{object}	object{code=string,message=string}
//	@Failure		401					{object}	object{code=string,message=string}
//	@Failure		404					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-types/{session_type_id} [patch]
func (h *Handler) updateSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	var req updateSessionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	params := application.UpdateSessionTypeParams{
		Name:         req.Name,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Kind:         req.Kind,
		FlatFeeMinor: req.FlatFeeMinor,
	}

	st, err := h.update.Execute(c.Request.Context(), actor, siteID, stID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTypeResponse(st))
}

// archiveSessionType archives a session type.
//
//	@Summary		Archive session type
//	@Description	Archive a session type.
//	@Tags			session-types
//	@Produce		json
//	@Param			site_id				path	string	true	"Site ID"			format(uuid)
//	@Param			session_type_id		path	string	true	"Session Type ID"	format(uuid)
//	@Success		200					{object}	object
//	@Failure		401					{object}	object{code=string,message=string}
//	@Failure		404					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-types/{session_type_id}/actions/archive [post]
func (h *Handler) archiveSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	if err := h.archive.Execute(c.Request.Context(), actor, siteID, stID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// reactivateSessionType reactivates an archived session type.
//
//	@Summary		Reactivate session type
//	@Description	Reactivate an archived session type.
//	@Tags			session-types
//	@Produce		json
//	@Param			site_id				path	string	true	"Site ID"			format(uuid)
//	@Param			session_type_id		path	string	true	"Session Type ID"	format(uuid)
//	@Success		200					{object}	sessionTypeResponse
//	@Failure		401					{object}	object{code=string,message=string}
//	@Failure		404					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","owner"]
//	@Router			/sites/{site_id}/session-types/{session_type_id}/actions/activate [post]
func (h *Handler) reactivateSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	st, err := h.reactivate.Execute(c.Request.Context(), actor, siteID, stID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTypeResponse(st))
}

func (h *Handler) handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
	httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
	c.AbortWithStatusJSON(status, resp)
}

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		v, ok := c.Get(tenant.AuthContextKey)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		authCtx, ok := v.(tenant.AuthorizationContext)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		switch authCtx.Role {
		case "owner", "manager", "practitioner", "parent":
		default:
			httpserver.WriteError(c, http.StatusForbidden, "forbidden_role_unknown", "Access denied.", nil)
			return
		}

		if _, exists := allowed[authCtx.Role]; !exists {
			httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
			return
		}

		c.Next()
	}
}
