package httpparents

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	app "nursery-management-system/api/internal/modules/parents/application"
	"nursery-management-system/api/internal/modules/parents/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/lifecycle"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	createParent    *app.CreateParentUseCase
	updateParent    *app.UpdateParentUseCase
	getParent       *app.GetParentUseCase
	listParents     *app.ListParentsUseCase
	softDelete      *app.SoftDeleteParentUseCase
	linkChild       *app.LinkChildUseCase
	unlinkChild     *app.UnlinkChildUseCase
	inviteToPortal  *app.InviteToPortalUseCase
	revokeAccess    *app.RevokePortalAccessUseCase
	logger          *slog.Logger
}

func NewHandler(
	createParent *app.CreateParentUseCase,
	updateParent *app.UpdateParentUseCase,
	getParent *app.GetParentUseCase,
	listParents *app.ListParentsUseCase,
	softDelete *app.SoftDeleteParentUseCase,
	linkChild *app.LinkChildUseCase,
	unlinkChild *app.UnlinkChildUseCase,
	inviteToPortal *app.InviteToPortalUseCase,
	revokeAccess *app.RevokePortalAccessUseCase,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		createParent:   createParent,
		updateParent:   updateParent,
		getParent:      getParent,
		listParents:    listParents,
		softDelete:     softDelete,
		linkChild:      linkChild,
		unlinkChild:    unlinkChild,
		inviteToPortal: inviteToPortal,
		revokeAccess:   revokeAccess,
		logger:         logger,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/parents", h.listParentsHandler)
	group.POST("/parents", h.createParentHandler)
	group.GET("/parents/:parent_id", h.getParentHandler)
	group.PUT("/parents/:parent_id", h.updateParentHandler)
	group.DELETE("/parents/:parent_id", h.softDeleteHandler)
	group.POST("/parents/:parent_id/link-child", h.linkChildHandler)
	group.DELETE("/parents/:parent_id/link-child/:child_id", h.unlinkChildHandler)
	group.POST("/parents/:parent_id/invite", h.inviteToPortalHandler)
	group.POST("/parents/:parent_id/revoke-access", h.revokeAccessHandler)
}

func (h *Handler) listParentsHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var isActive *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true"
		isActive = &b
	}
	var search *string
	if v := strings.TrimSpace(c.Query("search")); v != "" {
		search = &v
	}

	page := 1
	pageSize := 20
	if v := c.Query("page"); v != "" {
		if parsed, err := parsePositiveInt(v); err == nil {
			page = parsed
		}
	}
	if v := c.Query("page_size"); v != "" {
		if parsed, err := parsePositiveInt(v); err == nil {
			pageSize = parsed
		}
	}

	result, err := h.listParents.Execute(c.Request.Context(), actor, isActive, search, page, pageSize)
	if err != nil {
		httpserver.WriteMappedError(c, h.logger, err)
		return
	}

	parents := make([]parentResponse, 0, len(result.Parents))
	for _, p := range result.Parents {
		parents = append(parents, toParentResponse(p))
	}

	c.JSON(http.StatusOK, parentListResponse{
		Parents:    parents,
		TotalCount: result.TotalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (h *Handler) createParentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req struct {
		FirstName                 string  `json:"first_name"`
		LastName                  *string `json:"last_name"`
		Email                     *string `json:"email"`
		Phone                     *string `json:"phone"`
		AddressLine1              *string `json:"address_line1"`
		AddressLine2              *string `json:"address_line2"`
		AddressCity               *string `json:"address_city"`
		AddressPostcode           *string `json:"address_postcode"`
		RelationshipToChild       *string `json:"relationship_to_child"`
		HasParentalResponsibility bool    `json:"has_parental_responsibility"`
		CanPickUp                 bool    `json:"can_pick_up"`
		IsEmergencyContact        bool    `json:"is_emergency_contact"`
		Notes                     *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	result, err := h.createParent.Execute(c.Request.Context(), actor, app.CreateParentParams{
		FirstName:                 req.FirstName,
		LastName:                  req.LastName,
		Email:                     req.Email,
		Phone:                     req.Phone,
		AddressLine1:              req.AddressLine1,
		AddressLine2:              req.AddressLine2,
		AddressCity:               req.AddressCity,
		AddressPostcode:           req.AddressPostcode,
		RelationshipToChild:       req.RelationshipToChild,
		HasParentalResponsibility: req.HasParentalResponsibility,
		CanPickUp:                 req.CanPickUp,
		IsEmergencyContact:        req.IsEmergencyContact,
		Notes:                     req.Notes,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toParentResponse(result))
}

func (h *Handler) getParentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	result, err := h.getParent.Execute(c.Request.Context(), actor, parentID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	children := make([]parentChildLinkResponse, 0, len(result.Children))
	for _, ch := range result.Children {
		children = append(children, toParentChildLinkResponse(ch))
	}

	resp := parentWithChildrenResponse{
		parentResponse: toParentResponse(result.Parent),
		Children:       children,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) updateParentHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	var req struct {
		FirstName                 *string `json:"first_name"`
		LastName                  *string `json:"last_name"`
		Email                     *string `json:"email"`
		Phone                     *string `json:"phone"`
		AddressLine1              *string `json:"address_line1"`
		AddressLine2              *string `json:"address_line2"`
		AddressCity               *string `json:"address_city"`
		AddressPostcode           *string `json:"address_postcode"`
		RelationshipToChild       *string `json:"relationship_to_child"`
		HasParentalResponsibility *bool   `json:"has_parental_responsibility"`
		CanPickUp                 *bool   `json:"can_pick_up"`
		IsEmergencyContact        *bool   `json:"is_emergency_contact"`
		Notes                     *string `json:"notes"`
		IsActive                  *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	result, err := h.updateParent.Execute(c.Request.Context(), actor, parentID, app.UpdateParentParams{
		FirstName:                 req.FirstName,
		LastName:                  req.LastName,
		Email:                     req.Email,
		Phone:                     req.Phone,
		AddressLine1:              req.AddressLine1,
		AddressLine2:              req.AddressLine2,
		AddressCity:               req.AddressCity,
		AddressPostcode:           req.AddressPostcode,
		RelationshipToChild:       req.RelationshipToChild,
		HasParentalResponsibility: req.HasParentalResponsibility,
		CanPickUp:                 req.CanPickUp,
		IsEmergencyContact:        req.IsEmergencyContact,
		Notes:                     req.Notes,
		IsActive:                  req.IsActive,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toParentResponse(result))
}

func (h *Handler) softDeleteHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	if err := h.softDelete.Execute(c.Request.Context(), actor, parentID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parent deactivated successfully."})
}

func (h *Handler) linkChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	var req struct {
		ChildID string `json:"child_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	childID, err := uuid.Parse(strings.TrimSpace(req.ChildID))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", []map[string]string{{"field": "child_id"}})
		return
	}

	result, err := h.linkChild.Execute(c.Request.Context(), actor, parentID, childID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toParentChildLinkResponse(result))
}

func (h *Handler) unlinkChildHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	childID, ok := parseUUIDParam(c, "child_id")
	if !ok {
		return
	}

	var req struct {
		ReasonCode string `json:"reason_code"`
		ReasonNote string `json:"reason_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	req.ReasonCode = strings.TrimSpace(req.ReasonCode)
	req.ReasonNote = strings.TrimSpace(req.ReasonNote)

	if req.ReasonCode == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "relationship_reason_required", "Invalid request payload.", []map[string]string{{"field": "reason_code"}})
		return
	}
	if !lifecycle.IsValidReasonCode(req.ReasonCode) {
		httpserver.WriteError(c, http.StatusBadRequest, "lifecycle_reason_invalid", "Invalid request payload.", []map[string]string{{"field": "reason_code"}})
		return
	}
	if req.ReasonCode == lifecycle.ReasonOther && req.ReasonNote == "" {
		httpserver.WriteError(c, http.StatusBadRequest, "reason_note_required_for_other", "Invalid request payload.", []map[string]string{{"field": "reason_note"}})
		return
	}

	if err := h.unlinkChild.Execute(c.Request.Context(), actor, parentID, childID, req.ReasonCode, req.ReasonNote); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parent unlinked from child successfully."})
}

func (h *Handler) inviteToPortalHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	if err := h.inviteToPortal.Execute(c.Request.Context(), actor, parentID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Portal invite sent successfully."})
}

func (h *Handler) revokeAccessHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	parentID, ok := parseUUIDParam(c, "parent_id")
	if !ok {
		return
	}

	if err := h.revokeAccess.Execute(c.Request.Context(), actor, parentID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Portal access revoked successfully."})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrParentNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "parent_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, domain.ErrParentInactive) {
		httpserver.WriteError(c, http.StatusBadRequest, "parent_inactive", "Parent is inactive.", nil)
		return
	}
	if errors.Is(err, domain.ErrChildNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, domain.ErrLinkNotFound) {
		httpserver.WriteError(c, http.StatusNotFound, "parent_child_link_not_found", "Resource not found.", nil)
		return
	}
	if errors.Is(err, domain.ErrLinkAlreadyExists) {
		httpserver.WriteError(c, http.StatusConflict, "parent_child_link_exists", "Parent is already linked to this child.", nil)
		return
	}
	if errors.Is(err, domain.ErrUserAlreadyLinked) {
		httpserver.WriteError(c, http.StatusConflict, "user_already_linked", "User account is already linked to another parent.", nil)
		return
	}

	httpserver.WriteMappedError(c, h.logger, err)
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(name)))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", []map[string]string{{"field": name}})
		return uuid.UUID{}, false
	}
	return id, true
}

func parsePositiveInt(s string) (int, error) {
	var v int
	_, err := parsePositiveIntFromString(s, &v)
	return v, err
}

func parsePositiveIntFromString(s string, v *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("not a positive integer")
		}
		n = n*10 + int(c-'0')
	}
	if n <= 0 {
		return 0, errors.New("not a positive integer")
	}
	*v = n
	return n, nil
}
