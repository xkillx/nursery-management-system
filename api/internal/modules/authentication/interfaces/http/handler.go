package httpauth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/authentication/application"
	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/config"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/uid"
)

const refreshCookieName = "refresh_token"
const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

type Handler struct {
	login    *application.LoginUseCase
	refresh  *application.RefreshUseCase
	logout   *application.LogoutUseCase
	switch_  *application.SwitchMembershipUseCase
	cfg      config.Config
	logger   *slog.Logger
	recorder *metrics.Recorder
}

func NewHandler(
	login *application.LoginUseCase,
	refresh *application.RefreshUseCase,
	logout *application.LogoutUseCase,
	switch_ *application.SwitchMembershipUseCase,
	cfg config.Config,
) *Handler {
	return &Handler{
		login:   login,
		refresh: refresh,
		logout:  logout,
		switch_: switch_,
		cfg:     cfg,
	}
}

func (h *Handler) WithObservability(logger *slog.Logger, recorder *metrics.Recorder) *Handler {
	return &Handler{
		login:    h.login,
		refresh:  h.refresh,
		logout:   h.logout,
		switch_:  h.switch_,
		cfg:      h.cfg,
		logger:   logger,
		recorder: recorder,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	auth := group.Group("/auth")
	auth.POST("/login", h.loginHandler)
	auth.POST("/refresh", h.refreshHandler)
	auth.POST("/logout", h.logoutHandler)
	auth.POST("/switch-membership", h.switchMembershipHandler)
}

type loginRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	MembershipID string `json:"membership_id"`
}

type switchMembershipRequest struct {
	MembershipID string `json:"membership_id" binding:"required"`
}

type membershipResponse struct {
	MembershipID string `json:"membership_id"`
	TenantID     string `json:"tenant_id"`
	TenantName   string `json:"tenant_name"`
	BranchID     string `json:"branch_id"`
	BranchName   string `json:"branch_name"`
	Role         string `json:"role"`
}

type authResponse struct {
	AccessToken          string               `json:"access_token"`
	TokenType            string               `json:"token_type"`
	ExpiresIn            int64                `json:"expires_in_seconds"`
	User                 userResponse         `json:"user"`
	ActiveMembership     membershipResponse   `json:"active_membership"`
	AvailableMemberships []membershipResponse `json:"available_memberships"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) loginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"details":    err.Error(),
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	ctx := application.ContextWithRequestMeta(c.Request.Context(), c.Request.UserAgent(), c.ClientIP())

	result, err := h.login.Execute(ctx, req.Email, req.Password, req.MembershipID)
	if err != nil {
		var selErr *domain.MembershipSelectionRequiredError
		var valErr *domain.ValidationError
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			h.recordAuthFailure(c, "login", "login_invalid_credentials")
			h.unauthorized(c)
		case errors.As(err, &valErr):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":       "validation_error",
				"message":    "Invalid request payload.",
				"request_id": c.Writer.Header().Get("X-Request-ID"),
				"details":    map[string]string{"field": valErr.Field, "message": valErr.Message},
			})
		case errors.As(err, &selErr):
			msg := "Choose a nursery to continue."
			if selErr.IsStaleChoice {
				msg = "That access is no longer available. Choose another nursery or contact your manager."
			}
			choices := make([]membershipResponse, 0, len(selErr.Memberships))
			for _, m := range selErr.Memberships {
				choices = append(choices, toMembershipResponse(m))
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":                  "membership_selection_required",
				"message":               msg,
				"request_id":            c.Writer.Header().Get("X-Request-ID"),
				"available_memberships": choices,
			})
		default:
			h.internalError(c)
		}
		return
	}

	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	h.setCSRFCookie(c, uid.NewCSRFToken(), result.RefreshExpiresAt)
	c.JSON(http.StatusOK, h.buildAuthResponse(result.AccessToken, result.User, result.Memberships, result.ActiveMembership))
}

func (h *Handler) refreshHandler(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || strings.TrimSpace(rawRefresh) == "" {
		h.recordAuthFailure(c, "refresh", "refresh_invalid_or_missing_token")
		h.unauthorized(c)
		return
	}

	if !h.validateCSRFSessionAction(c) {
		return
	}

	ctx := application.ContextWithRequestMeta(c.Request.Context(), c.Request.UserAgent(), c.ClientIP())

	result, err := h.refresh.Execute(ctx, rawRefresh)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) {
			h.recordAuthFailure(c, "refresh", "refresh_invalid_or_missing_token")
			h.unauthorized(c)
		} else {
			h.internalError(c)
		}
		return
	}

	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	h.setCSRFCookie(c, uid.NewCSRFToken(), result.RefreshExpiresAt)
	c.JSON(http.StatusOK, h.buildAuthResponse(result.AccessToken, result.User, result.Memberships, result.ActiveMembership))
}

func (h *Handler) logoutHandler(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err == nil && strings.TrimSpace(rawRefresh) != "" {
		if !h.validateCSRFSessionAction(c) {
			h.recordAuthFailure(c, "logout", "logout_csrf_denied")
			return
		}

		_ = h.logout.Execute(c.Request.Context(), rawRefresh)
	}

	h.clearRefreshCookie(c)
	h.clearCSRFCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) switchMembershipHandler(c *gin.Context) {
	var req switchMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":       "validation_error",
			"message":    "Invalid request payload.",
			"details":    err.Error(),
			"request_id": c.Writer.Header().Get("X-Request-ID"),
		})
		return
	}

	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || strings.TrimSpace(rawRefresh) == "" {
		h.recordAuthFailure(c, "switch_membership", "switch_membership_invalid_or_missing_token")
		h.unauthorized(c)
		return
	}

	if !h.validateCSRFSessionAction(c) {
		h.recordAuthFailure(c, "switch_membership", "switch_membership_csrf_denied")
		return
	}

	ctx := application.ContextWithRequestMeta(c.Request.Context(), c.Request.UserAgent(), c.ClientIP())
	requestID := c.Writer.Header().Get("X-Request-ID")

	result, err := h.switch_.Execute(ctx, rawRefresh, req.MembershipID, requestID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			h.recordAuthFailure(c, "switch_membership", "switch_membership_invalid_or_missing_token")
			h.unauthorized(c)
		case errors.Is(err, domain.ErrInvalidMembership):
			h.forbiddenScopeSelection(c, "Invalid membership selection.")
		default:
			h.internalError(c)
		}
		return
	}

	h.setRefreshCookie(c, result.RefreshToken, result.RefreshExpiresAt)
	h.setCSRFCookie(c, uid.NewCSRFToken(), result.RefreshExpiresAt)
	c.JSON(http.StatusOK, h.buildAuthResponse(result.AccessToken, result.User, result.Memberships, result.ActiveMembership))
}

func (h *Handler) setRefreshCookie(c *gin.Context, value string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		refreshCookieName,
		value,
		int(time.Until(expiresAt).Seconds()),
		"/",
		"",
		h.cfg.AppEnv == "prod",
		true,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(refreshCookieName, "", -1, "/", "", h.cfg.AppEnv == "prod", true)
}

func (h *Handler) setCSRFCookie(c *gin.Context, value string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		csrfCookieName,
		value,
		int(time.Until(expiresAt).Seconds()),
		"/",
		"",
		h.cfg.AppEnv == "prod",
		false,
	)
}

func (h *Handler) clearCSRFCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(csrfCookieName, "", -1, "/", "", h.cfg.AppEnv == "prod", false)
}

func (h *Handler) validateCSRFSessionAction(c *gin.Context) bool {
	csrfCookie, err := c.Cookie(csrfCookieName)
	if err != nil || strings.TrimSpace(csrfCookie) == "" {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	csrfHeader := strings.TrimSpace(c.GetHeader(csrfHeaderName))
	if csrfHeader == "" || csrfHeader != csrfCookie {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	if !isTrustedOriginOrReferer(c) {
		h.forbiddenScopeSelection(c, "Invalid session action.")
		return false
	}

	return true
}

func isTrustedOriginOrReferer(c *gin.Context) bool {
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return false
	}

	if origin := strings.TrimSpace(c.GetHeader("Origin")); origin != "" {
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, host)
	}

	if referer := strings.TrimSpace(c.GetHeader("Referer")); referer != "" {
		u, err := url.Parse(referer)
		if err != nil {
			return false
		}
		return strings.EqualFold(u.Host, host)
	}

	return false
}

func (h *Handler) buildAuthResponse(accessToken string, user domain.User, memberships []domain.Membership, activeMembership domain.Membership) authResponse {
	out := make([]membershipResponse, 0, len(memberships))
	for _, m := range memberships {
		out = append(out, toMembershipResponse(m))
	}

	return authResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.cfg.JWTAccessTTLMin) * 60,
		User: userResponse{
			ID:    user.ID.String(),
			Email: user.Email,
		},
		ActiveMembership:     toMembershipResponse(activeMembership),
		AvailableMemberships: out,
	}
}

func toMembershipResponse(m domain.Membership) membershipResponse {
	return membershipResponse{
		MembershipID: m.ID.String(),
		TenantID:     m.TenantID.String(),
		TenantName:   m.TenantName,
		BranchID:     m.BranchID.String(),
		BranchName:   m.BranchName,
		Role:         m.Role,
	}
}

func (h *Handler) unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":       "unauthorized",
		"message":    "Invalid credentials or session.",
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func (h *Handler) internalError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"code":       "internal_error",
		"message":    "Something went wrong.",
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func (h *Handler) forbiddenScopeSelection(c *gin.Context, message string) {
	if h.logger != nil || h.recorder != nil {
		httpserver.RecordAuthorizationDenial(c, h.logger, h.recorder, "session_action", "forbidden_scope_selection")
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"code":       "forbidden_scope_selection",
		"message":    message,
		"request_id": c.Writer.Header().Get("X-Request-ID"),
	})
}

func (h *Handler) recordAuthFailure(c *gin.Context, operation, reason string) {
	if h.logger != nil {
		h.logger.Warn("auth_failure",
			"operation", operation,
			"reason", reason,
			"request_id", httpserver.RequestIDFromContext(c),
			"correlation_id", httpserver.CorrelationIDFromContext(c),
		)
	}
	if h.recorder != nil {
		h.recorder.AuthFailure(operation, reason)
	}
}
