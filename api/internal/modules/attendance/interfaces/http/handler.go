package httpattendance

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	logger       *slog.Logger
	checkIn      *application.CheckInChild
	checkOut     *application.CheckOutChild
	correct      *application.CorrectAttendance
	listSessions *application.ListCorrectionSessions
	listHistory  *application.ListCorrectionHistory
}

func NewHandler(
	checkIn *application.CheckInChild,
	checkOut *application.CheckOutChild,
	correct *application.CorrectAttendance,
	listSessions *application.ListCorrectionSessions,
	listHistory *application.ListCorrectionHistory,
) *Handler {
	return &Handler{checkIn: checkIn, checkOut: checkOut, correct: correct, listSessions: listSessions, listHistory: listHistory}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		checkIn:      h.checkIn,
		checkOut:     h.checkOut,
		correct:      h.correct,
		listSessions: h.listSessions,
		listHistory:  h.listHistory,
		logger:       logger,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	g := protected.Group("")
	g.Use(requireRoles("manager", "practitioner"))
	g.POST("/attendance/check-ins", h.checkInHandler)
	g.POST("/attendance/check-outs", h.checkOutHandler)

	managerOnly := protected.Group("")
	managerOnly.Use(requireRoles("manager"))
	managerOnly.POST("/attendance/corrections", h.correctionHandler)
	managerOnly.GET("/attendance/sessions", h.listSessionsHandler)
	managerOnly.GET("/attendance/sessions/:session_id/history", h.listHistoryHandler)
}

func (h *Handler) checkInHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req checkInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	session, err := h.checkIn.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toSessionResponse(session))
}

func (h *Handler) checkOutHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req checkOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	childID, err := parseChildID(req.ChildID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	session, err := h.checkOut.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toSessionResponse(session))
}

func (h *Handler) correctionHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req correctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params, err := parseCorrectionRequest(req)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	result, err := h.correct.Execute(c.Request.Context(), actor, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if result.Created {
		c.JSON(http.StatusCreated, toSessionResponse(result.Session))
	} else {
		c.JSON(http.StatusOK, toSessionResponse(result.Session))
	}
}

func (h *Handler) listSessionsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	childIDStr := c.Query("child_id")
	if childIDStr == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	localDateStr := c.Query("local_date")
	if localDateStr == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}
	localDate, err := time.Parse("2006-01-02", localDateStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	ctx, err := h.listSessions.Execute(c.Request.Context(), actor, childID, localDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCorrectionSessionContextResponse(ctx))
}

func (h *Handler) listHistoryHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	result, err := h.listHistory.Execute(c.Request.Context(), actor, sessionID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toCorrectionHistoryResponse(result))
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

		if _, exists := allowed[authCtx.Role]; !exists {
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.")
			return
		}

		c.Next()
	}
}
