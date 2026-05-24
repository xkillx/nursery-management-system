package httpserver

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDAccessLogAndRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(AccessLogMiddleware(logger))
	router.Use(RecoveryMiddleware(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Request-ID", "req-panic")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if got := w.Header().Get("X-Request-ID"); got != "req-panic" {
		t.Fatalf("expected request id header req-panic, got %q", got)
	}

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if body.Code != "internal_error" {
		t.Fatalf("expected internal_error, got %q", body.Code)
	}
	if body.RequestID != "req-panic" {
		t.Fatalf("expected request_id req-panic, got %q", body.RequestID)
	}

	logOutput := logs.String()
	for _, want := range []string{
		`"msg":"panic recovered"`,
		`"msg":"http_request"`,
		`"path":"/panic"`,
		`"status":500`,
		`"request_id":"req-panic"`,
	} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("expected logs to contain %s, got %s", want, logOutput)
		}
	}
}
