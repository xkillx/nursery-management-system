package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/ratelimit"
)

func TestRateLimitMiddlewareUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := ratelimit.NewFixedWindowLimiter(100, 1*time.Minute)

	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("X-RateLimit-Limit"); got != "100" {
		t.Fatalf("expected X-RateLimit-Limit 100, got %q", got)
	}
	if got := w.Header().Get("X-RateLimit-Remaining"); got != "99" {
		t.Fatalf("expected X-RateLimit-Remaining 99, got %q", got)
	}
	if w.Header().Get("X-RateLimit-Reset") == "" {
		t.Fatal("expected X-RateLimit-Reset header to be set")
	}
	if w.Header().Get("Retry-After") != "" {
		t.Fatal("expected Retry-After header to be absent")
	}
}

func TestRateLimitMiddlewareAtLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := ratelimit.NewFixedWindowLimiter(2, 1*time.Minute)

	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
	if got := w.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Fatalf("expected X-RateLimit-Limit 2, got %q", got)
	}
	if got := w.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("expected X-RateLimit-Remaining 0, got %q", got)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header to be set")
	}

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse response body: %v", err)
	}
	if body.Code != "rate_limited" {
		t.Fatalf("expected code rate_limited, got %q", body.Code)
	}
	if body.Message != "Too many requests." {
		t.Fatalf("expected message 'Too many requests.', got %q", body.Message)
	}
}

func TestRateLimitMiddlewareDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := ratelimit.NewFixedWindowLimiter(1, 1*time.Minute)

	router := gin.New()
	router.Use(RateLimitMiddleware(limiter))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("IP1 first request: expected status %d, got %d", http.StatusOK, w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("IP2 first request: expected status %d, got %d", http.StatusOK, w2.Code)
	}
	if got := w2.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Fatalf("IP2: expected X-RateLimit-Remaining 0, got %q", got)
	}
}

func TestRateLimitMiddlewareExemptsHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := ratelimit.NewFixedWindowLimiter(1, 1*time.Minute)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := router.Group("/api")
	api.Use(RateLimitMiddleware(limiter))
	api.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("api request: expected status %d, got %d", http.StatusOK, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("second api request: expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("health request: expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("X-RateLimit-Limit") != "" {
		t.Fatal("health request should not have rate limit headers")
	}
}
