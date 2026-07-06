package httpabsence

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/absence/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	logger      *slog.Logger
	markAbsent  *application.MarkAbsent
	clearMarker *application.ClearMarker
}

func NewHandler(markAbsent *application.MarkAbsent, clearMarker *application.ClearMarker, logger *slog.Logger) *Handler {
	return &Handler{logger: logger, markAbsent: markAbsent, clearMarker: clearMarker}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	g := protected.Group("")
	g.Use(requireRoles("manager", "practitioner"))
	g.POST("/attendance/absence-markers", h.markAbsentHandler)
	g.POST("/attendance/absence-markers/:absence_marker_id/clear", h.clearMarkerHandler)
}

func (h *Handler) markAbsentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req markAbsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	result, err := h.markAbsent.Execute(c.Request.Context(), actor, req.ChildID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if result.Created {
		c.JSON(http.StatusCreated, toMarkerResponse(result.Marker))
	} else {
		c.JSON(http.StatusOK, toMarkerResponse(result.Marker))
	}
}

func (h *Handler) clearMarkerHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	markerID := c.Param("absence_marker_id")
	if _, err := parseMarkerID(markerID); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", nil)
		return
	}

	marker, err := h.clearMarker.Execute(c.Request.Context(), actor, markerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toMarkerResponse(marker))
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

		if _, exists := allowed[authCtx.Role]; !exists {
			httpserver.WriteError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
			return
		}

		c.Next()
	}
}
