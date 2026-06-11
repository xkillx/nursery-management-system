package httpserver

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
)

type AuthorizationContext = tenant.AuthorizationContext

type TokenParser interface {
	ParseAccessToken(raw string) (AuthorizationContext, error)
}

func AuthnMiddleware(tokens TokenParser) gin.HandlerFunc {
	return authnMiddleware(tokens, nil, nil)
}

func AuthnMiddlewareWithObservability(tokens TokenParser, logger *slog.Logger, recorder *metrics.Recorder) gin.HandlerFunc {
	return authnMiddleware(tokens, logger, recorder)
}

func authnMiddleware(tokens TokenParser, logger *slog.Logger, recorder *metrics.Recorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if rawToken == "" || !strings.HasPrefix(strings.TrimSpace(c.GetHeader("Authorization")), "Bearer ") {
			recordAuthFailure(c, logger, recorder, "bearer_auth", "missing_bearer_token")
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		claims, err := tokens.ParseAccessToken(rawToken)
		if err != nil {
			recordAuthFailure(c, logger, recorder, "bearer_auth", "invalid_access_token")
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		c.Set(tenant.AuthContextKey, AuthorizationContext{
			UserID:        claims.UserID,
			MembershipID:  claims.MembershipID,
			TenantID:      claims.TenantID,
			BranchID:      claims.BranchID,
			Role:          claims.Role,
			RequestID:     requestIDFromContext(c),
			CorrelationID: CorrelationIDFromContext(c),
			TraceID:       TraceIDFromContext(c),
		})
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return requireRoles(roles...)
}

func RequireRolesWithObservability(logger *slog.Logger, recorder *metrics.Recorder, roles ...string) gin.HandlerFunc {
	return requireRolesWithObservability(logger, recorder, roles...)
}

func requireRoles(roles ...string) gin.HandlerFunc {
	return requireRolesWithObservability(nil, nil, roles...)
}

func requireRolesWithObservability(logger *slog.Logger, recorder *metrics.Recorder, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		authCtx, ok := AuthContextFromContext(c)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		switch authCtx.Role {
		case "owner", "manager", "practitioner", "parent":
		default:
			RecordAuthorizationDenial(c, logger, recorder, "require_roles", "forbidden_role_unknown")
			writeError(c, http.StatusForbidden, "forbidden_role_unknown", "Access denied.", nil)
			return
		}

		if _, exists := allowed[authCtx.Role]; !exists {
			RecordAuthorizationDenial(c, logger, recorder, "require_roles", "forbidden_role")
			writeError(c, http.StatusForbidden, "forbidden_role", "Access denied.", nil)
			return
		}

		c.Next()
	}
}

func AuthContextFromContext(c *gin.Context) (AuthorizationContext, bool) {
	v, ok := c.Get(tenant.AuthContextKey)
	if !ok {
		return AuthorizationContext{}, false
	}

	authCtx, castOK := v.(AuthorizationContext)
	if !castOK {
		return AuthorizationContext{}, false
	}

	return authCtx, true
}

func authContextFromContext(c *gin.Context) (AuthorizationContext, bool) {
	return AuthContextFromContext(c)
}

func recordAuthFailure(c *gin.Context, logger *slog.Logger, recorder *metrics.Recorder, operation, reason string) {
	if logger == nil && recorder == nil {
		return
	}
	args := []any{
		"operation", operation,
		"reason", reason,
		"request_id", requestIDFromContext(c),
		"correlation_id", CorrelationIDFromContext(c),
		"route", c.FullPath(),
		"method", c.Request.Method,
	}
	if logger != nil {
		logger.Warn("auth_failure", args...)
	}
	recorder.AuthFailure(operation, reason)
}

func RecordAuthorizationDenial(c *gin.Context, logger *slog.Logger, recorder *metrics.Recorder, operation, denialCode string) {
	args := []any{
		"operation", operation,
		"denial_code", denialCode,
		"request_id", requestIDFromContext(c),
		"correlation_id", CorrelationIDFromContext(c),
		"route", c.FullPath(),
		"method", c.Request.Method,
	}
	if traceID := TraceIDFromContext(c); traceID != "" {
		args = append(args, "trace_id", traceID)
	}
	if authCtx, ok := AuthContextFromContext(c); ok {
		args = append(args,
			"user_id", authCtx.UserID,
			"membership_id", authCtx.MembershipID,
			"tenant_id", authCtx.TenantID,
			"branch_id", authCtx.BranchID,
			"role", authCtx.Role,
		)
	}
	if logger != nil {
		logger.Warn("authorization_denial", args...)
	}
	recorder.AuthorizationDenial(operation, denialCode)
}
