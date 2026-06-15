package httpserver

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

func TestLogMappedError_NilLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	LogMappedError(c, nil, http.StatusInternalServerError, "internal_error", nil)
}

func TestLogMappedError_SilentFor4xx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	LogMappedError(c, logger, http.StatusBadRequest, "validation_error", nil)
	LogMappedError(c, logger, http.StatusConflict, "attendance_session_already_open", nil)

	if logs.Len() > 0 {
		t.Fatalf("expected no logs for 4xx, got: %s", logs.String())
	}
}

func TestLogMappedError_LogsStructuredFieldsFor5xx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	router := gin.New()
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.Set(requestIDContextKey, "req-abc")
		c.Set(correlationIDContextKey, "corr-xyz")
		LogMappedError(c, logger, http.StatusInternalServerError, "internal_error", nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("X-Request-ID", "req-abc")
	router.ServeHTTP(w, req)

	logOutput := logs.String()
	for _, want := range []string{
		`"msg":"request_failed"`,
		`"method":"GET"`,
		`"route":"/api/v1/test"`,
		`"status":500`,
		`"status_class":"5xx"`,
		`"error_code":"internal_error"`,
		`"request_id":"req-abc"`,
		`"correlation_id":"corr-xyz"`,
	} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("expected log to contain %s, got %s", want, logOutput)
		}
	}
}

func TestLogMappedError_IncludesWrappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	original := domainerrors.Internal(domainerrors.NotFound("child", "child not found"))
	LogMappedError(c, logger, http.StatusInternalServerError, "internal_error", original)

	logOutput := logs.String()
	if !strings.Contains(logOutput, `"error":`) {
		t.Fatalf("expected log to include wrapped error, got %s", logOutput)
	}
}

func TestLogMappedError_NilGinContext(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	LogMappedError(nil, logger, http.StatusInternalServerError, "internal_error", nil)

	if logs.Len() > 0 {
		t.Fatalf("expected no logs for nil context, got: %s", logs.String())
	}
}

func TestLogMappedError_UnmatchedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/unknown", nil)

	LogMappedError(c, logger, http.StatusInternalServerError, "internal_error", nil)

	logOutput := logs.String()
	if !strings.Contains(logOutput, `"route":"unmatched"`) {
		t.Fatalf("expected route to be 'unmatched', got %s", logOutput)
	}
}

func TestLogMappedError_IncludesActorFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(tenant.AuthContextKey, AuthorizationContext{
		UserID:       "user-1",
		MembershipID: "mem-1",
		TenantID:     "tenant-1",
		BranchID:     "branch-1",
		Role:         "manager",
	})

	LogMappedError(c, logger, http.StatusInternalServerError, "internal_error", nil)

	logOutput := logs.String()
	for _, want := range []string{
		`"user_id":"user-1"`,
		`"membership_id":"mem-1"`,
		`"tenant_id":"tenant-1"`,
		`"branch_id":"branch-1"`,
		`"role":"manager"`,
	} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("expected log to contain %s, got %s", want, logOutput)
		}
	}
}
