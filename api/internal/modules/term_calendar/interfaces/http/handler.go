package httptermcalendar

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/term_calendar/application"
	httpserver "nursery-management-system/api/internal/platform/http"
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

func (h *Handler) listTerms(c *gin.Context) {
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

	terms, err := h.list.Execute(c.Request.Context(), actor, siteID, includeArchived)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"academic_terms": toTermListResponse(terms)})
}

func (h *Handler) createTerm(c *gin.Context) {
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

	var req createTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid start_date format.")
		return
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid end_date format.")
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

	c.JSON(http.StatusCreated, gin.H{"academic_term": toTermResponse(term)})
}

func (h *Handler) updateTerm(c *gin.Context) {
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

	termID, err := uuid.Parse(c.Param("term_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	var req updateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid start_date format.")
			return
		}
		params.StartDate = &sd
	}
	if req.EndDate != nil {
		ed, err := parseDate(*req.EndDate)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid end_date format.")
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

func (h *Handler) archiveTerm(c *gin.Context) {
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

	termID, err := uuid.Parse(c.Param("term_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
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
