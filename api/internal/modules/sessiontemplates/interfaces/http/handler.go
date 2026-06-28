package httpsessiontemplates

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/sessiontemplates/application"
	httpserver "nursery-management-system/api/internal/platform/http"
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

	templates, err := h.list.Execute(c.Request.Context(), actor, siteID, includeArchived)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"session_templates": toSessionTemplateListResponse(templates)})
}

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
