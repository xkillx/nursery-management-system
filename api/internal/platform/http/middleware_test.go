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

	"nursery-management-system/api/internal/platform/metrics"
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
		`"route":"/panic"`,
		`"status":500`,
		`"request_id":"req-panic"`,
	} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("expected logs to contain %s, got %s", want, logOutput)
		}
	}
}

func TestCorrelationIDGeneratedWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID to be generated")
	}
	if w.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("expected X-Correlation-ID to be generated")
	}
}

func TestCorrelationIDPreserved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("X-Correlation-ID"); got != "corr-123" {
		t.Fatalf("expected X-Correlation-ID corr-123, got %q", got)
	}
}

func TestInvalidRequestIDReplaced(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "has\nnewline")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	got := w.Header().Get("X-Request-ID")
	if got == "has\nnewline" || got == "" {
		t.Fatalf("expected X-Request-ID to be replaced, got %q", got)
	}
}

func TestValidTraceparentExtractsTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var traceID string
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID = TraceIDFromContext(c)
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if traceID != "0af7651916cd43dd8448eb211c80319c" {
		t.Fatalf("expected trace ID extracted, got %q", traceID)
	}
}

func TestMalformedTraceparentIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var traceID string
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID = TraceIDFromContext(c)
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "garbage")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if traceID != "" {
		t.Fatalf("expected empty trace ID, got %q", traceID)
	}
}

func TestZeroTraceIDIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var traceID string
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID = TraceIDFromContext(c)
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", "00-00000000000000000000000000000000-0000000000000000-01")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if traceID != "" {
		t.Fatalf("expected empty trace ID for zero trace, got %q", traceID)
	}
}

func TestTooLongHeaderReplaced(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	longID := strings.Repeat("a", 200)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", longID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	got := w.Header().Get("X-Request-ID")
	if got == longID {
		t.Fatal("expected too-long request ID to be replaced")
	}
}

func TestNewRequestIDReturnsValue(t *testing.T) {
	id := NewRequestID()
	if id == "" || len(id) != 32 {
		t.Fatalf("expected 32-char hex ID, got %q (len=%d)", id, len(id))
	}
}

func TestAccessLogMiddlewareWithMetricsIncludesCorrelationID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	registry := metrics.NewRegistry()
	recorder := metrics.NewRecorder(registry)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(AccessLogMiddlewareWithMetrics(logger, recorder))
	router.GET("/test/:id", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	req.Header.Set("X-Correlation-ID", "corr-456")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logs.String()
	for _, want := range []string{
		`"correlation_id":"corr-456"`,
		`"route":"/test/:id"`,
		`"status_class":"2xx"`,
	} {
		if !strings.Contains(logOutput, want) {
			t.Fatalf("expected logs to contain %s, got %s", want, logOutput)
		}
	}
}

func TestAccessLogMiddlewareWithMetricsNilRecorder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelInfo}))

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(AccessLogMiddlewareWithMetrics(logger, nil))
	router.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
