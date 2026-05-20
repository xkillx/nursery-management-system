package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/auth"
)

const authContextKey = "auth_context"

type AuthorizationContext struct {
	UserID       string `json:"user_id"`
	MembershipID string `json:"membership_id"`
	TenantID     string `json:"tenant_id"`
	BranchID     string `json:"branch_id"`
	Role         string `json:"role"`
	RequestID    string `json:"request_id"`
}

func authnMiddleware(tokens *auth.TokenManager) gin.HandlerFunc {
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

		c.Set(authContextKey, AuthorizationContext{
			UserID:       claims.Subject,
			MembershipID: claims.MembershipID,
			TenantID:     claims.TenantID,
			BranchID:     claims.BranchID,
			Role:         claims.Role,
			RequestID:    requestIDFromContext(c),
		})
		c.Next()
	}
}

func requireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		authCtx, ok := authContextFromContext(c)
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

func authContextFromContext(c *gin.Context) (AuthorizationContext, bool) {
	v, ok := c.Get(authContextKey)
	if !ok {
		return AuthorizationContext{}, false
	}

	authCtx, castOK := v.(AuthorizationContext)
	if !castOK {
		return AuthorizationContext{}, false
	}

	return authCtx, true
}
