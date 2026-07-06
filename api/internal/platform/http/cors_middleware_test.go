package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware_PreflightAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://app.example.com"}, ""))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Errorf("expected ACAO header, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected credentials header")
	}
}

func TestCORSMiddleware_SimpleRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://app.example.com"}, ""))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Errorf("expected ACAO header, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://app.example.com"}, ""))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no ACAO header for disallowed origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_DefaultOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware(nil, "https://default.example.com"))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://default.example.com")
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://default.example.com" {
		t.Errorf("expected ACAO header from default, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_AllowMethodsHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://app.example.com"}, ""))
	router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Error("expected Allow-Methods header on preflight")
	}
}

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"empty", "", nil},
		{"single", "https://app.example.com", []string{"https://app.example.com"}},
		{"multiple", "https://a.com, https://b.com", []string{"https://a.com", "https://b.com"}},
		{"with spaces", "  https://a.com , https://b.com  ", []string{"https://a.com", "https://b.com"}},
		{"trailing comma", "https://a.com,", []string{"https://a.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseAllowedOrigins(tt.input)
			if len(result) != len(tt.expect) {
				t.Fatalf("expected %d origins, got %d: %v", len(tt.expect), len(result), result)
			}
			for i, v := range result {
				if v != tt.expect[i] {
					t.Errorf("origin[%d] = %q, want %q", i, v, tt.expect[i])
				}
			}
		})
	}
}
