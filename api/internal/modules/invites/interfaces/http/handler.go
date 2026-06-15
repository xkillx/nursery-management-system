package httpinvite

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/invites/application"
	"nursery-management-system/api/internal/modules/invites/domain"
	"nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	create     *application.CreateInviteUseCase
	list       *application.ListInvitesUseCase
	resend     *application.ResendInviteUseCase
	revoke     *application.RevokeInviteUseCase
	accept     *application.AcceptInviteUseCase
	tokenMgr   *tokens.Manager
	ipLimiter  *ratelimit.FixedWindowLimiter
	logger     *slog.Logger
}

func NewHandler(
	create *application.CreateInviteUseCase,
	list *application.ListInvitesUseCase,
	resend *application.ResendInviteUseCase,
	revoke *application.RevokeInviteUseCase,
	accept *application.AcceptInviteUseCase,
	tokenMgr *tokens.Manager,
	ipLimiter *ratelimit.FixedWindowLimiter,
) *Handler {
	return &Handler{
		create:    create,
		list:      list,
		resend:    resend,
		revoke:    revoke,
		accept:    accept,
		tokenMgr:  tokenMgr,
		ipLimiter: ipLimiter,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger) *Handler {
	return &Handler{
		create:    h.create,
		list:      h.list,
		resend:    h.resend,
		revoke:    h.revoke,
		accept:    h.accept,
		tokenMgr:  h.tokenMgr,
		ipLimiter: h.ipLimiter,
		logger:    logger,
	}
}

func (h *Handler) RegisterPublicRoutes(api *gin.RouterGroup) {
	api.POST("/invites/accept", h.acceptHandler)
}

func (h *Handler) RegisterManagerRoutes(manager *gin.RouterGroup) {
	manager.POST("/invites", h.createHandler)
	manager.GET("/invites", h.listHandler)
	manager.POST("/invites/:invite_id/resend", h.resendHandler)
	manager.POST("/invites/:invite_id/revoke", h.revokeHandler)
}

type createInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

func (h *Handler) createHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req createInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	result, err := h.create.Execute(c.Request.Context(), actor, req.Email, req.Role)
	if err != nil {
		status, resp := httpserver.MapDomainError(err, httpserver.RequestIDFromContext(c))
		httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	status := http.StatusCreated
	if !result.IsNew {
		status = http.StatusOK
	}
	c.JSON(status, toInviteResponse(result.Invite))
}

func (h *Handler) listHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	statusVal := strings.TrimSpace(c.Query("status"))
	status, valid := application.ParseStatus(statusVal)
	if !valid {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "status"})
		return
	}

	result, err := h.list.Execute(c.Request.Context(), actor, status)
	if err != nil {
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": toInviteResponseList(result.Invites)})
}

func (h *Handler) resendHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "invite_id"})
		return
	}

	result, err := h.resend.Execute(c.Request.Context(), actor, inviteID)
	if err != nil {
		status, resp := httpserver.MapDomainError(err, httpserver.RequestIDFromContext(c))
		httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, toInviteResponse(result.Invite))
}

func (h *Handler) revokeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "invite_id"})
		return
	}

	result, err := h.revoke.Execute(c.Request.Context(), actor, inviteID)
	if err != nil {
		status, resp := httpserver.MapDomainError(err, httpserver.RequestIDFromContext(c))
		httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, toInviteResponse(result.Invite))
}

type acceptInviteRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *Handler) acceptHandler(c *gin.Context) {
	clientIP := c.ClientIP()
	if !h.ipLimiter.Allow("invite_accept_ip:" + clientIP) {
		httpserver.WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many requests.", nil)
		return
	}

	var req acceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	tokenHash := h.tokenMgr.Hash(req.Token)

	_, err := h.accept.Execute(c.Request.Context(), tokenHash, req.NewPassword)
	if err != nil {
		status, resp := httpserver.MapDomainError(err, httpserver.RequestIDFromContext(c))
		httpserver.LogMappedError(c, h.logger, status, resp.Code, err)
		c.AbortWithStatusJSON(status, resp)
		return
	}

	c.Status(http.StatusNoContent)
}

type inviteResponse struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	Role       string  `json:"role"`
	Status     string  `json:"status"`
	ExpiresAt  string  `json:"expires_at"`
	AcceptedAt *string `json:"accepted_at"`
	RevokedAt  *string `json:"revoked_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func toInviteResponse(inv domain.Invite) inviteResponse {
	return inviteResponse{
		ID:         inv.ID.String(),
		Email:      inv.Email,
		Role:       inv.Role,
		Status:     string(inv.Status()),
		ExpiresAt:  inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		AcceptedAt: formatTimePtr(inv.AcceptedAt),
		RevokedAt:  formatTimePtr(inv.RevokedAt),
		CreatedAt:  inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toInviteResponseList(invites []domain.Invite) []inviteResponse {
	out := make([]inviteResponse, len(invites))
	for i, inv := range invites {
		out[i] = toInviteResponse(inv)
	}
	return out
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02T15:04:05Z")
	return &s
}
