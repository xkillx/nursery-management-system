package httpabsence

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/absence/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	markAbsent  *application.MarkAbsent
	clearMarker *application.ClearMarker
}

func NewHandler(markAbsent *application.MarkAbsent, clearMarker *application.ClearMarker) *Handler {
	return &Handler{markAbsent: markAbsent, clearMarker: clearMarker}
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req markAbsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	result, err := h.markAbsent.Execute(c.Request.Context(), actor, req.ChildID)
	if err != nil {
		handleError(c, err)
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
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	markerID := c.Param("absence_marker_id")
	if _, err := parseMarkerID(markerID); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	marker, err := h.clearMarker.Execute(c.Request.Context(), actor, markerID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toMarkerResponse(marker))
}

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)
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

		if _, exists := allowed[authCtx.Role]; !exists {
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.")
			return
		}

		c.Next()
	}
}
