package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/dbtest"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
)

func TestHealthRoutesRespondWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(httpserver.RequestIDMiddleware())
	pinger := &fakeHealthPinger{}
	registerHealthRoutes(router, "/api/v1", pinger)

	for _, path := range []string{"/health", "/api/v1/health"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("X-Request-ID", "req-health")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
			}
			if got := w.Header().Get("X-Request-ID"); got != "req-health" {
				t.Fatalf("expected request id header req-health, got %q", got)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("parse response body: %v", err)
			}
			if body["status"] != "ok" {
				t.Fatalf("expected status ok, got %v", body["status"])
			}
			if body["request_id"] != "req-health" {
				t.Fatalf("expected request_id req-health, got %v", body["request_id"])
			}
		})
	}
}

func TestHealthRouteDatabaseUnavailableIncludesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(httpserver.RequestIDMiddleware())
	registerHealthRoutes(router, "/api/v1", &fakeHealthPinger{err: errors.New("down")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if body["code"] != "db_unavailable" {
		t.Fatalf("expected db_unavailable, got %v", body["code"])
	}
	if body["request_id"] == "" {
		t.Fatal("expected generated request_id in error body")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected generated X-Request-ID response header")
	}
}

type fakeHealthPinger struct {
	err error
}

func (p *fakeHealthPinger) Ping(context.Context) error {
	return p.err
}

func TestMetricsRouteExposedWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	t.Cleanup(func() { pool.Close() })

	cfg := baseTestConfig(t)
	cfg.MetricsEnabled = true

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := metrics.NewRegistry()
	recorder := metrics.NewRecorder(registry)

	router := BootstrapWithOptions(cfg, logger, pool, BootstrapOptions{
		MetricsRegistry: registry,
		MetricsRecorder: recorder,
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /metrics, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "nursery_api_http_requests_total") {
		t.Fatal("expected nursery_api metrics in response")
	}
}

func TestMetricsRouteNotExposedWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	t.Cleanup(func() { pool.Close() })

	cfg := baseTestConfig(t)
	cfg.MetricsEnabled = false

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := BootstrapWithOptions(cfg, logger, pool, BootstrapOptions{})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /metrics when disabled, got %d", w.Code)
	}
}

func TestMetricsRouteNotUnderAPIV1(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	t.Cleanup(func() { pool.Close() })

	cfg := baseTestConfig(t)
	cfg.MetricsEnabled = true

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := metrics.NewRegistry()
	recorder := metrics.NewRecorder(registry)

	router := BootstrapWithOptions(cfg, logger, pool, BootstrapOptions{
		MetricsRegistry: registry,
		MetricsRecorder: recorder,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /api/v1/metrics, got %d", w.Code)
	}
}

func baseTestConfig(t *testing.T) config.Config {
	t.Helper()
	cfg := config.Config{
		AppEnv:                       "local",
		APIPort:                      "8080",
		APIBasePath:                  "/api/v1",
		DatabaseURL:                  "",
		JWTAccessSecret:              "access-secret-for-testing",
		JWTRefreshSecret:             "refresh-secret-for-testing",
		JWTAccessTTLMin:              15,
		JWTRefreshTTLHours:           720,
		WebBaseURL:                   "http://localhost:4200",
		EmailProvider:                "smtp",
		SMTPHost:                     "localhost",
		SMTPPort:                     1025,
		SMTPFrom:                     "test@example.com",
		PasswordResetTokenSecret:     "reset-secret",
		PasswordResetTokenTTLMinutes: 60,
		InviteTokenSecret:            "invite-secret",
		InviteTokenTTLHours:          168,
		LogLevel:                     "info",
		MetricsEnabled:               true,
	}
	return cfg
}
