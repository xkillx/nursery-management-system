package httpsessiontypes

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontypes/application"
	httpserver "nursery-management-system/api/internal/platform/http"
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
) *Handler {
	return &Handler{
		create:     create,
		update:     update,
		list:       list,
		get:        get,
		archive:    archive,
		reactivate: reactivate,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		create:     h.create,
		update:     h.update,
		list:       h.list,
		get:        h.get,
		archive:    h.archive,
		reactivate: h.reactivate,
		logger:     logger,
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

func (h *Handler) listSessionTypes(c *gin.Context) {
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

	types, err := h.list.Execute(c.Request.Context(), actor, siteID, includeArchived)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"session_types": toSessionTypeListResponse(types)})
}

func (h *Handler) createSessionType(c *gin.Context) {
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

	var req createSessionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.CreateSessionTypeParams{
		Name:      req.Name,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	st, err := h.create.Execute(c.Request.Context(), actor, siteID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSessionTypeResponse(st))
}

func (h *Handler) getSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	st, err := h.get.Execute(c.Request.Context(), actor, siteID, stID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTypeResponse(st))
}

func (h *Handler) updateSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req updateSessionTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.UpdateSessionTypeParams{
		Name:      req.Name,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	st, err := h.update.Execute(c.Request.Context(), actor, siteID, stID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionTypeResponse(st))
}

func (h *Handler) archiveSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if err := h.archive.Execute(c.Request.Context(), actor, siteID, stID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) reactivateSessionType(c *gin.Context) {
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

	stID, err := uuid.Parse(c.Param("session_type_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
