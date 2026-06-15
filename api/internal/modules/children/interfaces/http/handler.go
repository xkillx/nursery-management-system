package httpchild

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/platform/tenant"

	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	logger         *slog.Logger
	listChildren   *application.ListChildren
	getChild       *application.GetChild
	createChild    *application.CreateChild
	updateChild    *application.UpdateChild
	markInactive   *application.MarkInactive
	listAttendance *application.ListAttendance
}

func NewHandler(
	listChildren *application.ListChildren,
	getChild *application.GetChild,
	createChild *application.CreateChild,
	updateChild *application.UpdateChild,
	markInactive *application.MarkInactive,
	listAttendance *application.ListAttendance,
) *Handler {
	return &Handler{
		listChildren:   listChildren,
		getChild:       getChild,
		createChild:    createChild,
		updateChild:    updateChild,
		markInactive:   markInactive,
		listAttendance: listAttendance,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		listChildren:   h.listChildren,
		getChild:       h.getChild,
		createChild:    h.createChild,
		updateChild:    h.updateChild,
		markInactive:   h.markInactive,
		listAttendance: h.listAttendance,
		logger:         logger,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/children/attendance", requireRoles("manager", "practitioner"), h.listAttendanceHandler)

	manager := protected.Group("")
	manager.Use(requireRoles("manager"))

	manager.GET("/children", h.listChildrenHandler)
	manager.GET("/children/:child_id", h.getChildHandler)
	manager.POST("/children", h.createChildHandler)
	manager.PATCH("/children/:child_id", h.updateChildHandler)
	manager.POST("/children/:child_id/actions/mark-inactive", h.markInactiveHandler)
}

func (h *Handler) listChildrenHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	statusFilter := strings.TrimSpace(c.Query("status"))
	limit := defaultListLimit
	offset := 0

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maxListLimit {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
			return
		}
		limit = parsed
	}

	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
			return
		}
		offset = parsed
	}

	children, err := h.listChildren.Execute(c.Request.Context(), actor, statusFilter, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]childResponse, 0, len(children))
	for _, child := range children {
		out = append(out, toChildResponse(child))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) getChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	childID := c.Param("child_id")

	child, err := h.getChild.Execute(c.Request.Context(), actor, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(child))
}

func (h *Handler) createChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if req.CoreHourlyRateMinor != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "child-specific billing rates are not supported; site rates are configured by owners")
		return
	}

	params := application.CreateChildParams{
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Notes:       req.Notes,
	}

	child, err := h.createChild.Execute(c.Request.Context(), actor, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toChildResponse(child))
}

func (h *Handler) updateChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	childID := c.Param("child_id")

	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	if req.CoreHourlyRateMinor != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "child-specific billing rates are not supported; site rates are configured by owners")
		return
	}

	params := application.UpdateChildParams{
		FullName:    req.FullName,
		DateOfBirth: req.DateOfBirth,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Notes:       req.Notes,
	}

	child, err := h.updateChild.Execute(c.Request.Context(), actor, childID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(child))
}

func (h *Handler) markInactiveHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	childID := c.Param("child_id")

	var req reasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.")
		return
	}

	params := application.MarkInactiveParams{
		ReasonCode: req.ReasonCode,
		ReasonNote: req.ReasonNote,
	}

	child, err := h.markInactive.Execute(c.Request.Context(), actor, childID, params)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(child))
}

func (h *Handler) listAttendanceHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.")
		return
	}

	children, err := h.listAttendance.Execute(c.Request.Context(), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]attendanceChildResponse, 0, len(children))
	for _, child := range children {
		out = append(out, toAttendanceResponse(child))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

// Request/response types

type childWriteRequest struct {
	FullName            string `json:"full_name"`
	DateOfBirth         string `json:"date_of_birth"`
	StartDate           string `json:"start_date"`
	EndDate             string `json:"end_date"`
	CoreHourlyRateMinor *int   `json:"core_hourly_rate_minor"`
	Notes               string `json:"notes"`
}

type reasonRequest struct {
	ReasonCode string `json:"reason_code"`
	ReasonNote string `json:"reason_note"`
}

// Helpers

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

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

// requireRoles checks that the authenticated user has one of the allowed roles.
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
