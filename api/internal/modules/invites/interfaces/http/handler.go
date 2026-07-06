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
	"nursery-management-system/api/internal/platform/http/pagination"
	"nursery-management-system/api/internal/platform/http/queryparams"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	create    *application.CreateInviteUseCase
	list      *application.ListInvitesUseCase
	resend    *application.ResendInviteUseCase
	revoke    *application.RevokeInviteUseCase
	accept    *application.AcceptInviteUseCase
	tokenMgr  *tokens.Manager
	ipLimiter *ratelimit.FixedWindowLimiter
	logger    *slog.Logger
}

func NewHandler(
	create *application.CreateInviteUseCase,
	list *application.ListInvitesUseCase,
	resend *application.ResendInviteUseCase,
	revoke *application.RevokeInviteUseCase,
	accept *application.AcceptInviteUseCase,
	tokenMgr *tokens.Manager,
	ipLimiter *ratelimit.FixedWindowLimiter,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		logger:    logger,
		create:    create,
		list:      list,
		resend:    resend,
		revoke:    revoke,
		accept:    accept,
		tokenMgr:  tokenMgr,
		ipLimiter: ipLimiter,
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

// createHandler creates a new invite.
//
//	@Summary		Create invite
//	@Description	Create a new invite for a manager or practitioner.
//	@Tags			invites
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createInviteRequest	true	"Invite data"
//	@Success		201		{object}	inviteResponse
//	@Success		200		{object}	inviteResponse
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		401		{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/invites [post]
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

// listHandler returns a paginated list of invites.
//
//	@Summary		List invites
//	@Description	Get a paginated list of invites with optional status filter.
//	@Tags			invites
//	@Produce		json
//	@Param			status		query		string	false	"Filter by status"	Enums(pending, accepted, revoked, all)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			page_size	query		int		false	"Items per page"	default(50)	minimum(1)	maximum(200)
//	@Success		200			{object}	object{items=[]inviteResponse,total=int,page=int,page_size=int}
//	@Failure		400			{object}	object{code=string,message=string}
//	@Failure		401			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/invites [get]
func (h *Handler) listHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	filters := queryparams.ParseFilterParams(c, map[string]string{
		"status": "string",
		"role":   "string",
	})

	statusVal := strings.TrimSpace(filters["status"])
	status, valid := application.ParseStatus(statusVal)
	if !valid {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Validation failed.", []map[string]string{{"field": "status", "message": "must be a valid status (pending, accepted, revoked, all)"}})
		return
	}

	var role *string
	if r, ok := filters["role"]; ok {
		role = &r
	}

	page, pageSize := pagination.ParsePageParams(c)
	offset := (page - 1) * pageSize

	result, total, err := h.list.ExecutePaginated(c.Request.Context(), actor, status, pageSize, offset, role)
	if err != nil {
		httpserver.WriteInternalError(c)
		return
	}

	c.JSON(http.StatusOK, pagination.PaginatedResponse(toInviteResponseList(result.Invites), total, page, pageSize))
}

// resendHandler resends an invite.
//
//	@Summary		Resend invite
//	@Description	Resend an invite email.
//	@Tags			invites
//	@Produce		json
//	@Param			invite_id	path		string	true	"Invite ID"	format(uuid)
//	@Success		200			{object}	inviteResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/invites/{invite_id}/resend [post]
func (h *Handler) resendHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Validation failed.", []map[string]string{{"field": "invite_id", "message": "must be a valid UUID"}})
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

// revokeHandler revokes an invite.
//
//	@Summary		Revoke invite
//	@Description	Revoke an invite.
//	@Tags			invites
//	@Produce		json
//	@Param			invite_id	path		string	true	"Invite ID"	format(uuid)
//	@Success		200			{object}	inviteResponse
//	@Failure		401			{object}	object{code=string,message=string}
//	@Failure		404			{object}	object{code=string,message=string}
//	@Security		BearerAuth
//	@x-roles		["manager"]
//	@Router			/invites/{invite_id}/revoke [post]
func (h *Handler) revokeHandler(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Validation failed.", []map[string]string{{"field": "invite_id", "message": "must be a valid UUID"}})
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

// acceptHandler accepts an invite and creates a user account.
//
//	@Summary		Accept invite
//	@Description	Accept an invite and create a user account with the provided password.
//	@Tags			invites
//	@Accept			json
//	@Produce		json
//	@Param			body	body		acceptInviteRequest	true	"Invite token and new password"
//	@Success		204
//	@Failure		400		{object}	object{code=string,message=string}
//	@Failure		429		{object}	object{code=string,message=string}
//	@Router			/invites/accept [post]
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
