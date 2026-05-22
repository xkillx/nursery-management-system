package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/tenant"
)

type AuthorizationContext = tenant.AuthorizationContext

type TokenParser interface {
	ParseAccessToken(raw string) (AuthorizationContext, error)
}

func AuthnMiddleware(tokens TokenParser) gin.HandlerFunc {
	return authnMiddleware(tokens)
}

func authnMiddleware(tokens TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		if rawToken == "" || !strings.HasPrefix(strings.TrimSpace(c.GetHeader("Authorization")), "Bearer ") {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		claims, err := tokens.ParseAccessToken(rawToken)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		c.Set(tenant.AuthContextKey, AuthorizationContext{
			UserID:       claims.UserID,
			MembershipID: claims.MembershipID,
			TenantID:     claims.TenantID,
			BranchID:     claims.BranchID,
			Role:         claims.Role,
			RequestID:    requestIDFromContext(c),
		})
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return requireRoles(roles...)
}

func requireRoles(roles ...string) gin.HandlerFunc {
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
		case "manager", "practitioner", "parent":
		default:
			writeError(c, http.StatusForbidden, "forbidden_role_unknown", "Access denied.", nil)
			return
		}

		if _, exists := allowed[authCtx.Role]; !exists {
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
