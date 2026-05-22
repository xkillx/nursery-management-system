package httpguardian

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/guardians/application"
	"nursery-management-system/api/internal/modules/guardians/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type Handler struct {
	list       *application.ListGuardians
	get        *application.GetGuardian
	create     *application.CreateGuardian
	update     *application.UpdateGuardian
	deactivate *application.DeactivateGuardian
	reactivate *application.ReactivateGuardian
}

func NewHandler(
	list *application.ListGuardians,
	get *application.GetGuardian,
	create *application.CreateGuardian,
	update *application.UpdateGuardian,
	deactivate *application.DeactivateGuardian,
	reactivate *application.ReactivateGuardian,
) *Handler {
	return &Handler{
		list:       list,
		get:        get,
		create:     create,
		update:     update,
		deactivate: deactivate,
		reactivate: reactivate,
	}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	manager := protected.Group("")
	manager.Use(requireRolesMiddleware("manager"))

	manager.GET("/guardians", h.listGuardians)
	manager.GET("/guardians/:guardian_id", h.getGuardian)
	manager.POST("/guardians", h.createGuardian)
	manager.PATCH("/guardians/:guardian_id", h.updateGuardian)
	manager.POST("/guardians/:guardian_id/actions/deactivate", h.deactivateGuardian)
	manager.POST("/guardians/:guardian_id/actions/reactivate", h.reactivateGuardian)
}

func (h *Handler) listGuardians(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	statusFilter, ok := parseStatusFilter(c)
	if !ok {
		return
	}

	limit, offset, ok := parsePagination(c)
	if !ok {
		return
	}

	guardians, err := h.list.Execute(c.Request.Context(), actor, statusFilter, limit, offset)
	if err != nil {
		httpserver.WriteInternalError(c)
		return
	}

	out := make([]guardianResponse, 0, len(guardians))
	for _, g := range guardians {
		out = append(out, toGuardianResponse(g))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *Handler) getGuardian(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	guardian, err := h.get.Execute(c.Request.Context(), actor, guardianID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpserver.WriteError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
			return
		}
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(guardian))
}

func (h *Handler) createGuardian(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req guardianWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "full_name"})
		return
	}

	guardian, err := h.create.Execute(c.Request.Context(), actor, application.CreateGuardianParams{
		FullName: req.FullName,
		Email:    strings.TrimSpace(req.Email),
		Phone:    strings.TrimSpace(req.Phone),
		Notes:    strings.TrimSpace(req.Notes),
	})
	if err != nil {
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toGuardianResponse(guardian))
}

func (h *Handler) updateGuardian(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	var req guardianWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	guardian, err := h.update.Execute(c.Request.Context(), actor, guardianID, application.UpdateGuardianParams{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Notes:    req.Notes,
	})
	if err != nil {
		errMsg := err.Error()
		if errMsg == "no fields to update" {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "body"})
			return
		}
		if errMsg == "guardian not found" {
			httpserver.WriteError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
			return
		}
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(guardian))
}

func (h *Handler) deactivateGuardian(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	reason, ok := parseReasonPayload(c)
	if !ok {
		return
	}

	guardian, err := h.deactivate.Execute(c.Request.Context(), actor, guardianID, application.DeactivateGuardianParams{
		ReasonCode: reason.Code,
		ReasonNote: reason.Note,
	})
	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "reason_code is required"):
			httpserver.WriteError(c, http.StatusBadRequest, "guardian_deactivation_reason_required", "Invalid request payload.", gin.H{"field": "reason_code"})
		case strings.Contains(errMsg, "invalid reason_code"):
			httpserver.WriteError(c, http.StatusBadRequest, "lifecycle_reason_invalid", "Invalid request payload.", gin.H{"field": "reason_code"})
		case strings.Contains(errMsg, "reason_note too long"):
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "reason_note"})
		case strings.Contains(errMsg, "reason_note required for other"):
			httpserver.WriteError(c, http.StatusBadRequest, "reason_note_required_for_other", "Invalid request payload.", gin.H{"field": "reason_note"})
		case errors.Is(err, domain.ErrNotFound):
			httpserver.WriteError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		default:
			httpserver.WriteInternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(guardian))
}

func (h *Handler) reactivateGuardian(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	guardian, err := h.reactivate.Execute(c.Request.Context(), actor, guardianID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpserver.WriteError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
			return
		}
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(guardian))
}

type parsedReason struct {
	Code string
	Note string
}

func parseReasonPayload(c *gin.Context) (parsedReason, bool) {
	var req reasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return parsedReason{}, false
	}

	req.ReasonCode = strings.TrimSpace(req.ReasonCode)
	req.ReasonNote = strings.TrimSpace(req.ReasonNote)

	if req.ReasonCode == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "guardian_deactivation_reason_required", "Invalid request payload.", gin.H{"field": "reason_code"})
		return parsedReason{}, false
	}

	return parsedReason{Code: req.ReasonCode, Note: req.ReasonNote}, true
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(name)))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": name})
		return uuid.UUID{}, false
	}
	return id, true
}

func parseStatusFilter(c *gin.Context) (domain.StatusFilter, bool) {
	v := strings.TrimSpace(c.Query("status"))
	if v == "" {
		return domain.StatusActive, true
	}
	switch v {
	case string(domain.StatusActive):
		return domain.StatusActive, true
	case string(domain.StatusInactive):
		return domain.StatusInactive, true
	case string(domain.StatusAll):
		return domain.StatusAll, true
	default:
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "status"})
		return "", false
	}
}

func parsePagination(c *gin.Context) (int, int, bool) {
	limit := defaultListLimit
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maxListLimit {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "limit"})
			return 0, 0, false
		}
		limit = parsed
	}

	offset := 0
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "offset"})
			return 0, 0, false
		}
		offset = parsed
	}

	return limit, offset, true
}

func requireRolesMiddleware(roles ...string) gin.HandlerFunc {
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
