package httpreset

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/passwordreset/application"
	"nursery-management-system/api/internal/modules/passwordreset/domain"
	"nursery-management-system/api/internal/platform/ratelimit"
)

func setupRouter() (*gin.Engine, *mockHandlerRepo, *mockHandlerEmail) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := &mockHandlerRepo{}
	email := &mockHandlerEmail{}
	tok := &mockHandlerTokenGen{raw: "raw-token", hash: "hash123", expiresAt: time.Now().Add(60 * time.Minute)}

	requestReset := application.NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", slog.New(slog.NewTextHandler(io.Discard, nil)))
	setPassword := application.NewSetNewPasswordUseCase(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))

	emailLimiter := ratelimit.NewFixedWindowLimiter(5, 15*time.Minute)
	ipLimiter := ratelimit.NewFixedWindowLimiter(20, 15*time.Minute)

	handler := NewHandler(requestReset, setPassword, emailLimiter, ipLimiter, nil, nil)
	api := r.Group("/api/v1")
	handler.RegisterRoutes(api)

	return r, repo, email
}

func TestRequestReset_InvalidEmail_Returns400(t *testing.T) {
	r, _, _ := setupRouter()

	body, _ := json.Marshal(map[string]string{"email": "not-an-email"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset-requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "validation_error") {
		t.Fatalf("expected validation_error, got %s", w.Body.String())
	}
}

func TestRequestReset_UnknownEmail_Returns202(t *testing.T) {
	r, _, _ := setupRouter()

	body, _ := json.Marshal(map[string]string{"email": "unknown@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset-requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "accepted") {
		t.Fatalf("expected accepted, got %s", w.Body.String())
	}
}

func TestRequestReset_ActiveEmail_Returns202(t *testing.T) {
	r, _, email := setupRouter()

	body, _ := json.Marshal(map[string]string{"email": "active@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset-requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	if len(email.captured) != 1 {
		t.Fatal("expected email to be sent")
	}
	if !strings.Contains(email.captured[0], "raw-token") {
		t.Fatal("expected reset URL to contain raw token")
	}
}

func TestResetPassword_InvalidPayload_Returns400(t *testing.T) {
	r, _, _ := setupRouter()

	body, _ := json.Marshal(map[string]string{"token": "", "new_password": "short"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-resets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestResetPassword_ValidToken_Returns204(t *testing.T) {
	r, _, _ := setupRouter()

	body, _ := json.Marshal(map[string]string{"token": "valid-token", "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-resets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestResetPassword_InvalidToken_Returns400(t *testing.T) {
	r, repo, _ := setupRouter()
	repo.resetErr = domain.ErrTokenInvalid

	body, _ := json.Marshal(map[string]string{"token": "bad-token", "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-resets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "password_reset_token_invalid") {
		t.Fatalf("expected token invalid code, got %s", w.Body.String())
	}
}

func TestResetPassword_ExpiredToken_Returns400(t *testing.T) {
	r, repo, _ := setupRouter()
	repo.resetErr = domain.ErrTokenExpired

	body, _ := json.Marshal(map[string]string{"token": "expired-token", "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-resets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "password_reset_token_expired") {
		t.Fatalf("expected token expired code, got %s", w.Body.String())
	}
}

func TestResetPassword_UsedToken_Returns400(t *testing.T) {
	r, repo, _ := setupRouter()
	repo.resetErr = domain.ErrTokenUsed

	body, _ := json.Marshal(map[string]string{"token": "used-token", "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-resets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "password_reset_token_used") {
		t.Fatalf("expected token used code, got %s", w.Body.String())
	}
}

// Mocks for handler tests

type mockHandlerRepo struct {
	user     domain.User
	found    bool
	resetErr error
}

func (m *mockHandlerRepo) FindUserByEmail(_ context.Context, email string) (domain.User, bool, error) {
	if m.found {
		return m.user, true, nil
	}
	if strings.Contains(email, "active") {
		return domain.User{ID: uuid.New(), Email: "active@example.com", IsActive: true}, true, nil
	}
	return domain.User{}, false, nil
}

func (m *mockHandlerRepo) IssueResetToken(_ context.Context, _ uuid.UUID, _ string, _ time.Time, sendEmail func() error) error {
	return sendEmail()
}

func (m *mockHandlerRepo) ResetPassword(_ context.Context, _ string, _ string) error {
	return m.resetErr
}

type mockHandlerEmail struct {
	captured []string
}

func (m *mockHandlerEmail) SendPasswordReset(_ context.Context, to string, resetURL string) error {
	m.captured = append(m.captured, to+"|"+resetURL)
	return nil
}

type mockHandlerTokenGen struct {
	raw       string
	hash      string
	expiresAt time.Time
}

func (m *mockHandlerTokenGen) Generate() (string, string, time.Time, error) {
	return m.raw, m.hash, m.expiresAt, nil
}
