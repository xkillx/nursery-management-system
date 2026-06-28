package httpmapping

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	app "nursery-management-system/api/internal/modules/parentchildmappings/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/lifecycle"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	createMapping *app.CreateMappingUseCase
	endMapping    *app.EndMappingUseCase
	logger        *slog.Logger
}

func NewHandler(createMapping *app.CreateMappingUseCase, endMapping *app.EndMappingUseCase, logger *slog.Logger) *Handler {
	return &Handler{logger: logger, createMapping: createMapping, endMapping: endMapping}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/parent-membership-children", h.createMappingHandler)
	group.POST("/parent-membership-children/:mapping_id/actions/end", h.endMappingHandler)
}

func (h *Handler) createMappingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req struct {
		MembershipID string `json:"membership_id"`
		ChildID      string `json:"child_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	membershipID, err := uuid.Parse(strings.TrimSpace(req.MembershipID))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": "membership_id"})
		return
	}
	childID, err := uuid.Parse(strings.TrimSpace(req.ChildID))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": "child_id"})
		return
	}

	result, err := h.createMapping.Execute(c.Request.Context(), toActor(actor), app.CreateMappingParams{
		TenantID:     actor.TenantID,
		BranchID:     actor.BranchID,
		MembershipID: membershipID,
		ChildID:      childID,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toMappingResponse(result))
}

func (h *Handler) endMappingHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	mappingID, ok := parseUUID(c, "mapping_id")
	if !ok {
		return
	}

	reasonCode, reasonNote, ok := parseReasonPayload(c, "relationship_reason_required")
	if !ok {
		return
	}

	result, err := h.endMapping.Execute(c.Request.Context(), toActor(actor), mappingID, reasonCode, reasonNote)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toMappingResponse(result))
}

func toActor(a tenant.ActorContext) app.ActorContext {
	return app.ActorContext{
		UserID:       a.UserID,
		MembershipID: a.MembershipID,
		TenantID:     a.TenantID,
		BranchID:     a.BranchID,
		RequestID:    a.RequestID,
	}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)

	if errors.Is(err, app.ErrMembershipNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "membership_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, app.ErrMembershipNotParent) {
		httpserver.WriteError(c, http.StatusBadRequest, "membership_not_parent", "Invalid request payload.", nil)
		return
	}
	if errors.Is(err, app.ErrMembershipNotActive) {
		httpserver.WriteError(c, http.StatusBadRequest, "membership_not_active", "Invalid request payload.", nil)
		return
	}
	if errors.Is(err, app.ErrChildNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, app.ErrMappingNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "parent_child_mapping_not_found", "Resource not found.", nil)
		return
	}

	status, resp := httpserver.MapDomainError(err, requestID)
	httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
	c.AbortWithStatusJSON(status, resp)
}

func parseUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(name)))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": name})
		return uuid.UUID{}, false
	}
	return id, true
}

func parseReasonPayload(c *gin.Context, missingCode string) (string, string, bool) {
	var req struct {
		ReasonCode string `json:"reason_code"`
		ReasonNote string `json:"reason_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return "", "", false
	}

	req.ReasonCode = strings.TrimSpace(req.ReasonCode)
	req.ReasonNote = strings.TrimSpace(req.ReasonNote)

	if req.ReasonCode == "" {
		httpserver.WriteError(c, http.StatusBadRequest, missingCode, "Invalid request payload.", map[string]string{"field": "reason_code"})
		return "", "", false
	}
	if !lifecycle.IsValidReasonCode(req.ReasonCode) {
		httpserver.WriteError(c, http.StatusBadRequest, "lifecycle_reason_invalid", "Invalid request payload.", map[string]string{"field": "reason_code"})
		return "", "", false
	}
	if len(req.ReasonNote) > lifecycle.MaxReasonNoteLen {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": "reason_note"})
		return "", "", false
	}
	if req.ReasonCode == lifecycle.ReasonOther && req.ReasonNote == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "reason_note_required_for_other", "Invalid request payload.", map[string]string{"field": "reason_note"})
		return "", "", false
	}

	return req.ReasonCode, req.ReasonNote, true
}
