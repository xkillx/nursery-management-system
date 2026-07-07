package httpabsence

import (
	"fmt"
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
	g.Use(httpserver.RequireRolesWithObservability(h.logger, nil, "manager", "practitioner"))
	g.POST("/attendance/absence-markers", h.markAbsentHandler)
	g.POST("/attendance/absence-markers/:absence_marker_id/clear", h.clearMarkerHandler)
}

// markAbsentHandler marks a child as absent.
//
//	@Summary		Mark child absent
//	@Description	Mark a child as absent for today.
//	@Tags			absence
//	@Accept			json
//	@Produce		json
//	@Param			body	body		markAbsentRequest	true	"Child ID"
//	@Success		201		{object}	absenceMarkerResponse
//	@Success		200		{object}	absenceMarkerResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","practitioner"]
//	@Router			/attendance/absence-markers [post]
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

	resp := toMarkerResponse(result.Marker)
	if result.Created {
		c.Header("Location", fmt.Sprintf("/api/attendance/absence-markers/%s", resp.ID))
		c.JSON(http.StatusCreated, resp)
	} else {
		c.JSON(http.StatusOK, resp)
	}
}

// clearMarkerHandler clears an absence marker.
//
//	@Summary		Clear absence marker
//	@Description	Clear an absence marker for a child.
//	@Tags			absence
//	@Produce		json
//	@Param			absence_marker_id	path		string	true	"Absence Marker ID"	format(uuid)
//	@Success		200					{object}	absenceMarkerResponse
//	@Failure		400					{object}	object{code=string,message=string}
//	@Failure		401					{object}	object{code=string,message=string}
//	@Failure		404					{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager","practitioner"]
//	@Router			/attendance/absence-markers/{absence_marker_id}/clear [post]
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
