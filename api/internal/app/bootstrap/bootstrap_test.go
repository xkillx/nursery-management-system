package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	httpserver "nursery-management-system/api/internal/platform/http"
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
