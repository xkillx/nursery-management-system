package httpreset

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/passwordreset/application"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/ratelimit"
)

type Handler struct {
	requestReset *application.RequestResetUseCase
	setPassword  *application.SetNewPasswordUseCase
	emailLimiter *ratelimit.FixedWindowLimiter
	ipLimiter    *ratelimit.FixedWindowLimiter
	logger       *slog.Logger
	recorder     *metrics.Recorder
}

func NewHandler(
	requestReset *application.RequestResetUseCase,
	setPassword *application.SetNewPasswordUseCase,
	emailLimiter *ratelimit.FixedWindowLimiter,
	ipLimiter *ratelimit.FixedWindowLimiter,
	recorder *metrics.Recorder,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		requestReset: requestReset,
		setPassword:  setPassword,
		emailLimiter: emailLimiter,
		ipLimiter:    ipLimiter,
		recorder:     recorder,
		logger:       logger,
	}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	auth := group.Group("/auth")
	auth.POST("/password-reset-requests", h.requestResetHandler)
	auth.POST("/password-resets", h.resetPasswordHandler)
}

type resetRequestPayload struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordPayload struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *Handler) requestResetHandler(c *gin.Context) {
	var req resetRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(req.Email))
	clientIP := c.ClientIP()

	if !h.emailLimiter.Allow("password_reset_email:"+emailNormalized) ||
		!h.ipLimiter.Allow("password_reset_ip:"+clientIP) {
		httpserver.WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many requests.", nil)
		return
	}

	result, err := h.requestReset.Execute(c.Request.Context(), req.Email)
	if err != nil {
		httpserver.WriteInternalError(c)
		return
	}

	if result.Accepted {
		c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
	}
}

func (h *Handler) resetPasswordHandler(c *gin.Context) {
	var req resetPasswordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	_, err := h.setPassword.Execute(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		var domainErr interface{ Error() string }
		if errors.As(err, &domainErr) {
			errMsg := domainErr.Error()
			switch {
			case strings.Contains(errMsg, "password_reset_token_invalid"):
				h.recordAuthFailure(c, "password_reset", "password_reset_invalid_token")
				httpserver.WriteError(c, http.StatusBadRequest, "password_reset_token_invalid", "Invalid reset token.", nil)
			case strings.Contains(errMsg, "password_reset_token_expired"):
				h.recordAuthFailure(c, "password_reset", "password_reset_expired_token")
				httpserver.WriteError(c, http.StatusBadRequest, "password_reset_token_expired", "Reset token has expired.", nil)
			case strings.Contains(errMsg, "password_reset_token_used"):
				h.recordAuthFailure(c, "password_reset", "password_reset_used_token")
				httpserver.WriteError(c, http.StatusBadRequest, "password_reset_token_used", "Reset token has already been used.", nil)
			case strings.Contains(errMsg, "validation_error"):
				httpserver.WriteError(c, http.StatusBadRequest, "validation_error", "Password must be at least 8 characters.", nil)
			default:
				httpserver.WriteInternalError(c)
			}
		} else {
			httpserver.WriteInternalError(c)
		}
		return
	}

	c.Status(http.StatusNoContent)
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
