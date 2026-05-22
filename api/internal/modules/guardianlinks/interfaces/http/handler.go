package httplink

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	app "nursery-management-system/api/internal/modules/guardianlinks/application"
	"nursery-management-system/api/internal/platform/lifecycle"
	"nursery-management-system/api/internal/platform/tenant"
	httpserver "nursery-management-system/api/internal/platform/http"
)

type Handler struct {
	createLink *app.CreateLinkUseCase
	endLink    *app.EndLinkUseCase
}

func NewHandler(createLink *app.CreateLinkUseCase, endLink *app.EndLinkUseCase) *Handler {
	return &Handler{createLink: createLink, endLink: endLink}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/guardian-child-links", h.createLinkHandler)
	group.POST("/guardian-child-links/:link_id/actions/end", h.endLinkHandler)
}

func (h *Handler) createLinkHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req struct {
		GuardianID string `json:"guardian_id"`
		ChildID    string `json:"child_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	guardianID, err := uuid.Parse(strings.TrimSpace(req.GuardianID))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": "guardian_id"})
		return
	}
	childID, err := uuid.Parse(strings.TrimSpace(req.ChildID))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", map[string]string{"field": "child_id"})
		return
	}

	result, err := h.createLink.Execute(c.Request.Context(), toActor(actor), app.CreateLinkParams{
		TenantID:   actor.TenantID,
		BranchID:   actor.BranchID,
		GuardianID: guardianID,
		ChildID:    childID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toLinkResponse(result))
}

func (h *Handler) endLinkHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	linkID, ok := parseUUID(c, "link_id")
	if !ok {
		return
	}

	reasonCode, reasonNote, ok := parseReasonPayload(c, "relationship_reason_required")
	if !ok {
		return
	}

	result, err := h.endLink.Execute(c.Request.Context(), toActor(actor), linkID, reasonCode, reasonNote)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toLinkResponse(result))
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

func handleError(c *gin.Context, err error) {
	requestID := httpserver.RequestIDFromContext(c)
	status, resp := httpserver.MapDomainError(err, requestID)

	if errors.Is(err, app.ErrGuardianNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, app.ErrGuardianNotActive) {
		httpserver.WriteError(c, http.StatusBadRequest, "guardian_not_active", "Invalid request payload.", nil)
		return
	}
	if errors.Is(err, app.ErrChildNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, app.ErrLinkNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "guardian_child_link_not_found", "Resource not found.", nil)
		return
	}

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
