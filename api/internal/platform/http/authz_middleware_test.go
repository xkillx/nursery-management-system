package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	authdomain "nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

func TestAuthorizationGuardsMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokens := authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720)
	tokenParser := &testTokenParser{tm: tokens}
	router := gin.New()
	router.Use(RequestIDMiddleware())

	protected := router.Group("/protected")
	protected.Use(AuthnMiddleware(tokenParser))
	protected.GET("/manager", RequireRoles("manager"), meHandler())
	protected.GET("/parent-link/:child_id", RequireRoles("parent"), parentLinkProbeHandler())
	protected.GET("/scope/:tenant_id/:branch_id", RequireRoles("manager", "practitioner", "parent"), scopeProbeHandler())

	managerToken := mustAccessToken(t, tokens, authdomain.ScopeClaims{
		MembershipID: uuid.NewString(),
		TenantID:     uuid.NewString(),
		BranchID:     uuid.NewString(),
		Role:         "manager",
	})
	managerClaims := mustParseClaims(t, tokens, managerToken)
	practitionerToken := mustAccessToken(t, tokens, authdomain.ScopeClaims{
		MembershipID: uuid.NewString(),
		TenantID:     uuid.NewString(),
		BranchID:     uuid.NewString(),
		Role:         "practitioner",
	})
	parentToken := mustAccessToken(t, tokens, authdomain.ScopeClaims{
		MembershipID: uuid.NewString(),
		TenantID:     uuid.NewString(),
		BranchID:     uuid.NewString(),
		Role:         "parent",
	})
	unknownRoleToken := mustAccessToken(t, tokens, authdomain.ScopeClaims{
		MembershipID: uuid.NewString(),
		TenantID:     uuid.NewString(),
		BranchID:     uuid.NewString(),
		Role:         "director",
	})

	tests := []struct {
		name       string
		path       string
		token      string
		statusCode int
		errorCode  string
	}{
		{name: "unauthenticated rejected", path: "/protected/manager", statusCode: http.StatusUnauthorized, errorCode: "unauthorized"},
		{name: "wrong role rejected", path: "/protected/manager", token: practitionerToken, statusCode: http.StatusForbidden, errorCode: "forbidden_role"},
		{name: "unknown role rejected", path: "/protected/manager", token: unknownRoleToken, statusCode: http.StatusForbidden, errorCode: "forbidden_role_unknown"},
		{name: "correct role allowed", path: "/protected/manager", token: managerToken, statusCode: http.StatusOK},
		{name: "wrong tenant rejected", path: "/protected/scope/" + uuid.NewString() + "/" + managerClaims.BranchID, token: managerToken, statusCode: http.StatusForbidden, errorCode: "forbidden_scope"},
		{name: "wrong branch rejected", path: "/protected/scope/" + managerClaims.TenantID + "/" + uuid.NewString(), token: managerToken, statusCode: http.StatusForbidden, errorCode: "forbidden_scope"},
		{name: "correct scope allowed", path: "/protected/scope/" + managerClaims.TenantID + "/" + managerClaims.BranchID, token: managerToken, statusCode: http.StatusOK},
		{name: "parent link rejected", path: "/protected/parent-link/" + uuid.NewString() + "?linked_child_id=" + uuid.NewString(), token: parentToken, statusCode: http.StatusForbidden, errorCode: "forbidden_parent_child_link"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.statusCode {
				t.Fatalf("expected status %d, got %d", tc.statusCode, w.Code)
			}

			if tc.errorCode != "" {
				var body map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to parse response body: %v", err)
				}
				code, _ := body["code"].(string)
				if code != tc.errorCode {
					t.Fatalf("expected code %q, got %q", tc.errorCode, code)
				}
			}
		})
	}

	t.Run("parent link allowed", func(t *testing.T) {
		childID := uuid.NewString()
		req := httptest.NewRequest(http.MethodGet, "/protected/parent-link/"+childID+"?linked_child_id="+childID, nil)
		req.Header.Set("Authorization", "Bearer "+parentToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

type testTokenParser struct {
	tm *authtokens.TokenManager
}

func (p *testTokenParser) ParseAccessToken(raw string) (tenant.AuthorizationContext, error) {
	claims, err := p.tm.ParseAccessToken(raw)
	if err != nil {
		return tenant.AuthorizationContext{}, err
	}
	return tenant.AuthorizationContext{
		UserID:       claims.Subject,
		MembershipID: claims.MembershipID,
		TenantID:     claims.TenantID,
		BranchID:     claims.BranchID,
		Role:         claims.Role,
	}, nil
}

func mustAccessToken(t *testing.T, tokens *authtokens.TokenManager, scope authdomain.ScopeClaims) string {
	t.Helper()
	userID := uuid.New()
	raw, _, err := tokens.GenerateAccessToken(userID, "test@example.com", scope)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}
	return raw
}

func mustParseClaims(t *testing.T, tokens *authtokens.TokenManager, raw string) authtokens.AccessClaims {
	t.Helper()
	parsed, err := tokens.ParseAccessToken(raw)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	return parsed
}
