package httpauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/authentication/application"
	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/config"
	httpserver "nursery-management-system/api/internal/platform/http"
)

// --- HTTP test fakes ---

type htFakeUserRepo struct {
	ByEmail map[string]domain.User
	Mships  map[uuid.UUID][]domain.Membership
	capEmail string
}

func (r *htFakeUserRepo) FindUserByEmail(_ context.Context, email string) (domain.User, error) {
	r.capEmail = email
	u, ok := r.ByEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *htFakeUserRepo) ListMembershipsByUserID(_ context.Context, uid uuid.UUID) ([]domain.Membership, error) {
	return r.Mships[uid], nil
}

type htFakeSessionRepo struct {
	tokens map[string]domain.RefreshToken
	users  map[uuid.UUID]domain.User
	mships map[uuid.UUID]domain.Membership
	calls  []string
}

func newHTFakeSessionRepo() *htFakeSessionRepo {
	return &htFakeSessionRepo{
		tokens: make(map[string]domain.RefreshToken),
		users:  make(map[uuid.UUID]domain.User),
		mships: make(map[uuid.UUID]domain.Membership),
	}
}

func (r *htFakeSessionRepo) CreateRefreshToken(_ context.Context, t domain.RefreshToken, _, _ string) error {
	r.calls = append(r.calls, "create")
	r.tokens[t.TokenHash] = t
	return nil
}

func (r *htFakeSessionRepo) FindActiveRefreshToken(_ context.Context, hash string) (domain.RefreshToken, domain.User, domain.Membership, error) {
	r.calls = append(r.calls, "find")
	tok, ok := r.tokens[hash]
	if !ok {
		return domain.RefreshToken{}, domain.User{}, domain.Membership{}, domain.ErrNotFound
	}
	return tok, r.users[tok.UserID], r.mships[tok.MembershipID], nil
}

func (r *htFakeSessionRepo) RotateRefreshToken(_ context.Context, oldID uuid.UUID, rep domain.RefreshToken, _, _ string) error {
	r.calls = append(r.calls, "rotate")
	for h, t := range r.tokens {
		if t.ID == oldID {
			delete(r.tokens, h)
			break
		}
	}
	r.tokens[rep.TokenHash] = rep
	return nil
}

func (r *htFakeSessionRepo) RevokeByTokenHash(_ context.Context, hash string) error {
	r.calls = append(r.calls, "revoke:"+hash)
	delete(r.tokens, hash)
	return nil
}

func (r *htFakeSessionRepo) CreateScopeSwitchAuditLog(_ context.Context, _ uuid.UUID, _, _ domain.Membership, _ string) error {
	r.calls = append(r.calls, "audit")
	return nil
}

func (r *htFakeSessionRepo) seed(tok domain.RefreshToken, u domain.User, m domain.Membership) {
	if r.tokens == nil {
		r.tokens = make(map[string]domain.RefreshToken)
		r.users = make(map[uuid.UUID]domain.User)
		r.mships = make(map[uuid.UUID]domain.Membership)
	}
	r.tokens[tok.TokenHash] = tok
	r.users[u.ID] = u
	r.mships[m.ID] = m
}

type htFakeTokens struct{}

func (htFakeTokens) GenerateRefreshToken(rememberMe bool) (string, string, time.Time, error) {
	if rememberMe {
		return "raw-refresh-token", "hash:raw-refresh-token", time.Now().UTC().Add(30 * 24 * time.Hour), nil
	}
	return "raw-refresh-token", "hash:raw-refresh-token", time.Now().UTC().Add(24 * time.Hour), nil
}

func (htFakeTokens) GenerateAccessToken(_ uuid.UUID, _ string, _ domain.ScopeClaims) (string, time.Time, error) {
	return "fake-access-token", time.Now().UTC().Add(15 * time.Minute), nil
}
func (htFakeTokens) HashRefreshToken(raw string) string { return "hash:" + raw }
func (htFakeTokens) AccessTTLSeconds() int64             { return 900 }

// --- Fixtures ---

var (
	hUID   = uuid.MustParse("10000000-0000-0000-0000-000000000001")
	hTID   = uuid.MustParse("10000000-0000-0000-0000-000000000010")
	hBID   = uuid.MustParse("10000000-0000-0000-0000-000000000100")
	hM1    = uuid.MustParse("10000000-0000-0000-0000-000000001001")
	hM2    = uuid.MustParse("10000000-0000-0000-0000-000000001002")
	hTokID = uuid.MustParse("10000000-0000-0000-0000-000000002001")
)

func htHash(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	return string(h)
}

func htUser() domain.User {
	return domain.User{ID: hUID, Email: "u@test.com", PasswordHash: htHash("password1"), IsActive: true}
}

func htM1() domain.Membership {
	return domain.Membership{ID: hM1, TenantID: hTID, TenantName: "Little Sprouts Nursery", BranchID: hBID, BranchName: "Main Branch", Role: "manager", IsActive: true}
}
func htM2() domain.Membership {
	return domain.Membership{ID: hM2, TenantID: hTID, TenantName: "Little Sprouts Nursery", BranchID: hBID, BranchName: "Main Branch", Role: "practitioner", IsActive: true}
}

// --- Router setup ---

func htRouter(ur *htFakeUserRepo, sr *htFakeSessionRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpserver.RequestIDMiddleware())

	tk := htFakeTokens{}
	login := application.NewLoginUseCase(ur, sr, tk)
	refresh := application.NewRefreshUseCase(ur, sr, tk)
	logout := application.NewLogoutUseCase(sr, tk)
	sw := application.NewSwitchMembershipUseCase(ur, sr, tk)
	cfg := config.Config{AppEnv: "local", JWTAccessTTLMin: 15, JWTRefreshTTLHours: 720}
	h := NewHandler(login, refresh, logout, sw, cfg)
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

// --- Helpers ---

func htDo(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func htJSON(t *testing.T, method, url, body string, cookies ...http.Cookie) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(&c)
	}
	return req
}

func htCSRFGet(t *testing.T, url string, csrf, refreshRaw string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshRaw})
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrf})
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("Origin", "http://api.test")
	req.Host = "api.test"
	return req
}

func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse json: %v", err)
	}
	return m
}

func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// --- Tests ---

func TestLoginHandler(t *testing.T) {
	t.Run("one-membership success returns 200 with token_type Bearer expires_in_seconds 900", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"u@test.com","password":"password1"}`)
		w := htDo(r, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := parseBody(t, w)
		if body["token_type"] != "Bearer" {
			t.Fatalf("expected Bearer, got %v", body["token_type"])
		}
		if body["expires_in_seconds"] != float64(900) {
			t.Fatalf("expected 900, got %v", body["expires_in_seconds"])
		}
		if body["active_membership"] == nil {
			t.Fatal("expected active_membership")
		}
		if body["available_memberships"] == nil {
			t.Fatal("expected available_memberships")
		}
	})

	t.Run("success sets refresh_token HttpOnly and csrf_token not HttpOnly", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"u@test.com","password":"password1"}`)
		w := htDo(r, req)

		rc := findCookie(w, "refresh_token")
		if rc == nil {
			t.Fatal("expected refresh_token cookie")
		}
		if !rc.HttpOnly {
			t.Fatal("refresh_token must be HttpOnly")
		}
		cc := findCookie(w, "csrf_token")
		if cc == nil {
			t.Fatal("expected csrf_token cookie")
		}
		if cc.HttpOnly {
			t.Fatal("csrf_token must not be HttpOnly")
		}
	})

	t.Run("multi-membership without membership_id returns 400 membership_selection_required with choices", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"u@test.com","password":"password1"}`)
		w := htDo(r, req)

		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		body := parseBody(t, w)
		if body["code"] != "membership_selection_required" {
			t.Fatalf("expected membership_selection_required, got %v", body["code"])
		}
		if body["message"] != "Choose a nursery to continue." {
			t.Fatalf("expected picker message, got %v", body["message"])
		}
		choices, ok := body["available_memberships"].([]interface{})
		if !ok || len(choices) != 2 {
			t.Fatalf("expected 2 available_memberships, got %v", body["available_memberships"])
		}
		first := choices[0].(map[string]interface{})
		if first["tenant_name"] != "Little Sprouts Nursery" {
			t.Fatalf("expected tenant_name in choice, got %v", first)
		}
		if first["branch_name"] != "Main Branch" {
			t.Fatalf("expected branch_name in choice, got %v", first)
		}
		rc := findCookie(w, "refresh_token")
		if rc != nil {
			t.Fatal("expected no session cookies")
		}
	})

	t.Run("stale explicit membership returns 400 membership_selection_required with stale message", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"u@test.com","password":"password1","membership_id":"00000000-0000-0000-0000-999999999999"}`)
		w := htDo(r, req)

		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		body := parseBody(t, w)
		if body["code"] != "membership_selection_required" {
			t.Fatalf("expected membership_selection_required, got %v", body["code"])
		}
		if body["message"] != "That access is no longer available. Choose another nursery or contact your manager." {
			t.Fatalf("expected stale message, got %v", body["message"])
		}
		choices, ok := body["available_memberships"].([]interface{})
		if !ok || len(choices) != 2 {
			t.Fatalf("expected 2 available_memberships, got %v", body["available_memberships"])
		}
		rc := findCookie(w, "refresh_token")
		if rc != nil {
			t.Fatal("expected no session cookies")
		}
	})

	t.Run("malformed membership_id returns 400 validation_error", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"u@test.com","password":"password1","membership_id":"not-a-uuid"}`)
		w := htDo(r, req)

		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		body := parseBody(t, w)
		if body["code"] != "validation_error" {
			t.Fatalf("expected validation_error, got %v", body["code"])
		}
	})
}

func TestRefreshHandler(t *testing.T) {
	seed := func(sr *htFakeSessionRepo, ur *htFakeUserRepo) {
		sr.seed(domain.RefreshToken{ID: hTokID, UserID: hUID, MembershipID: hM1, TokenHash: "hash:raw-refresh", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)}, htUser(), htM1())
	}

	t.Run("without refresh cookie returns 401 unauthorized", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htCSRFGet(t, "/api/v1/auth/refresh", "csrf", "")
		req.Header.Del("Cookie")
		req.Header.Del("Origin")
		req.Header.Del("X-CSRF-Token")
		w := htDo(r, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("with refresh cookie but missing CSRF cookie/header returns 403", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		req := htJSON(t, http.MethodPost, "/api/v1/auth/refresh", "")
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-refresh"})
		w := htDo(r, req)

		if w.Code != 403 {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("with CSRF but missing trusted Origin/Referer returns 403", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-refresh"})
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf"})
		req.Header.Set("X-CSRF-Token", "csrf")
		req.Header.Set("Origin", "http://evil.test")
		req.Host = "api.test"
		w := htDo(r, req)

		if w.Code != 403 {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("with valid cookies header and origin returns 200 and sets replacement cookies", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		req := htCSRFGet(t, "/api/v1/auth/refresh", "csrf", "raw-refresh")
		w := htDo(r, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		rc := findCookie(w, "refresh_token")
		if rc == nil {
			t.Fatal("expected replacement refresh cookie")
		}
		cc := findCookie(w, "csrf_token")
		if cc == nil {
			t.Fatal("expected replacement csrf cookie")
		}
	})

	t.Run("Referer fallback same-host passes CSRF", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-refresh"})
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf"})
		req.Header.Set("X-CSRF-Token", "csrf")
		req.Header.Set("Referer", "http://api.test/some/page")
		req.Host = "api.test"
		w := htDo(r, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	t.Run("without refresh cookie returns 204 and clears cookies", func(t *testing.T) {
		ur := &htFakeUserRepo{}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Host = "api.test"
		w := htDo(r, req)

		if w.Code != 204 {
			t.Fatalf("expected 204, got %d", w.Code)
		}
		rc := findCookie(w, "refresh_token")
		if rc == nil || rc.Value != "" {
			t.Fatal("expected cleared refresh_token cookie")
		}
		cc := findCookie(w, "csrf_token")
		if cc == nil || cc.Value != "" {
			t.Fatal("expected cleared csrf_token cookie")
		}
	})

	t.Run("with refresh cookie but missing CSRF returns 403 and does not revoke", func(t *testing.T) {
		ur := &htFakeUserRepo{}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "raw-refresh"})
		req.Host = "api.test"
		w := htDo(r, req)

		if w.Code != 403 {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		hasRevoke := false
		for _, c := range sr.calls {
			if strings.HasPrefix(c, "revoke:") {
				hasRevoke = true
			}
		}
		if hasRevoke {
			t.Fatal("expected no revoke call")
		}
	})

	t.Run("with valid cookies header and origin returns 204 revokes and clears cookies", func(t *testing.T) {
		ur := &htFakeUserRepo{}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htCSRFGet(t, "/api/v1/auth/logout", "csrf", "raw-refresh")
		w := htDo(r, req)

		if w.Code != 204 {
			t.Fatalf("expected 204, got %d", w.Code)
		}
		hasRevoke := false
		for _, c := range sr.calls {
			if strings.HasPrefix(c, "revoke:") {
				hasRevoke = true
			}
		}
		if !hasRevoke {
			t.Fatal("expected revoke call")
		}
		rc := findCookie(w, "refresh_token")
		if rc == nil || rc.Value != "" {
			t.Fatal("expected cleared refresh cookie")
		}
		cc := findCookie(w, "csrf_token")
		if cc == nil || cc.Value != "" {
			t.Fatal("expected cleared csrf cookie")
		}
	})

	t.Run("repeating logout without refresh cookie still returns 204", func(t *testing.T) {
		ur := &htFakeUserRepo{}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Host = "api.test"
		w := htDo(r, req)
		if w.Code != 204 {
			t.Fatalf("expected 204, got %d", w.Code)
		}

		req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req2.Host = "api.test"
		w2 := htDo(r, req2)
		if w2.Code != 204 {
			t.Fatalf("expected 204 on second call, got %d", w2.Code)
		}
	})
}

func TestSwitchMembershipHandler(t *testing.T) {
	seed := func(sr *htFakeSessionRepo, ur *htFakeUserRepo) {
		sr.seed(domain.RefreshToken{ID: hTokID, UserID: hUID, MembershipID: hM1, TokenHash: "hash:raw-refresh", ExpiresAt: time.Now().UTC().Add(24 * time.Hour)}, htUser(), htM1())
	}

	t.Run("missing JSON membership_id returns 400 validation_error", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		req := htCSRFGet(t, "/api/v1/auth/switch-membership", "csrf", "raw-refresh")
		req.Body = nil
		req.ContentLength = 0
		req.Header.Set("Content-Type", "application/json")
		w := htDo(r, req)

		if w.Code != 400 {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		body := parseBody(t, w)
		if body["code"] != "validation_error" {
			t.Fatalf("expected validation_error, got %v", body["code"])
		}
	})

	t.Run("without refresh cookie returns 401 unauthorized", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		r := htRouter(ur, sr)

		body := `{"membership_id":"` + hM2.String() + `"}`
		req := htJSON(t, http.MethodPost, "/api/v1/auth/switch-membership", body)
		w := htDo(r, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid explicit membership with valid CSRF returns 403 forbidden_scope_selection", func(t *testing.T) {
		ur := &htFakeUserRepo{Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		body := `{"membership_id":"00000000-0000-0000-0000-999999999999"}`
		req := htCSRFGet(t, "/api/v1/auth/switch-membership", "csrf", "raw-refresh")
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", "application/json")
		w := htDo(r, req)

		if w.Code != 403 {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("valid explicit membership with CSRF returns 200 and target active membership", func(t *testing.T) {
		ur := &htFakeUserRepo{ByEmail: map[string]domain.User{"u@test.com": htUser()}, Mships: map[uuid.UUID][]domain.Membership{hUID: {htM1(), htM2()}}}
		sr := newHTFakeSessionRepo()
		seed(sr, ur)
		r := htRouter(ur, sr)

		body := `{"membership_id":"` + hM2.String() + `"}`
		req := htCSRFGet(t, "/api/v1/auth/switch-membership", "csrf", "raw-refresh")
		req.Body = io.NopCloser(strings.NewReader(body))
		req.ContentLength = int64(len(body))
		req.Header.Set("Content-Type", "application/json")
		w := htDo(r, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		resp := parseBody(t, w)
		am := resp["active_membership"].(map[string]interface{})
		if am["membership_id"] != hM2.String() {
			t.Fatalf("expected active membership %s, got %v", hM2, am["membership_id"])
		}
		rc := findCookie(w, "refresh_token")
		if rc == nil {
			t.Fatal("expected replacement refresh cookie")
		}
		cc := findCookie(w, "csrf_token")
		if cc == nil {
			t.Fatal("expected replacement csrf cookie")
		}
	})
}
