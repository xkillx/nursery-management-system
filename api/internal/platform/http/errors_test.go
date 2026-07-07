package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	domainerrors "nursery-management-system/api/internal/platform/errors"
)

func TestWriteError_IncludesTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, http.StatusBadRequest, "test_code", "test message", nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	if body.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
	}

	parsed, err := time.Parse(time.RFC3339, body.Timestamp)
	if err != nil {
		t.Fatalf("expected valid RFC3339 timestamp, got %q: %v", body.Timestamp, err)
	}

	if parsed.Location() != time.UTC {
		t.Fatalf("expected UTC timezone, got %v", parsed.Location())
	}
}

func TestMapDomainError_IncludesTimestamp(t *testing.T) {
	err := domainerrors.Conflict("test_conflict", "test message")
	_, resp := MapDomainError(err, "req-1")

	if resp.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
	}

	parsed, parseErr := time.Parse(time.RFC3339, resp.Timestamp)
	if parseErr != nil {
		t.Fatalf("expected valid RFC3339 timestamp, got %q: %v", resp.Timestamp, parseErr)
	}

	if parsed.Location() != time.UTC {
		t.Fatalf("expected UTC timezone, got %v", parsed.Location())
	}
}

func TestMapDomainError_NonDomainError_IncludesTimestamp(t *testing.T) {
	_, resp := MapDomainError(http.ErrHandlerTimeout, "req-1")

	if resp.Timestamp == "" {
		t.Fatal("expected timestamp to be set for non-domain error")
	}

	_, parseErr := time.Parse(time.RFC3339, resp.Timestamp)
	if parseErr != nil {
		t.Fatalf("expected valid RFC3339 timestamp, got %q: %v", resp.Timestamp, parseErr)
	}
}

func TestErrorResponseShape_FullShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, http.StatusBadRequest, "test_code", "test message", nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "req-shape")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	code, ok := raw["code"].(string)
	if !ok || code != "test_code" {
		t.Fatalf("expected code 'test_code' (string), got %v (%T)", raw["code"], raw["code"])
	}

	message, ok := raw["message"].(string)
	if !ok || message != "test message" {
		t.Fatalf("expected message 'test message' (string), got %v (%T)", raw["message"], raw["message"])
	}

	requestID, ok := raw["request_id"].(string)
	if !ok || requestID != "req-shape" {
		t.Fatalf("expected request_id 'req-shape' (string), got %v (%T)", raw["request_id"], raw["request_id"])
	}

	timestamp, ok := raw["timestamp"].(string)
	if !ok {
		t.Fatalf("expected timestamp (string), got %v (%T)", raw["timestamp"], raw["timestamp"])
	}
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		t.Fatalf("expected valid RFC3339 timestamp, got %q", timestamp)
	}
}

func TestErrorResponseShape_DetailsAbsentWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, http.StatusBadRequest, "test_code", "test message", nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	if _, exists := raw["details"]; exists {
		t.Fatal("expected 'details' key to be absent when nil")
	}
}

func TestErrorResponseShape_DetailsPresentWhenProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, http.StatusBadRequest, "test_code", "test message", map[string]string{"field": "name"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	details, exists := raw["details"]
	if !exists {
		t.Fatal("expected 'details' key to be present")
	}

	detailsMap, ok := details.(map[string]any)
	if !ok {
		t.Fatalf("expected details to be object, got %T", details)
	}
	if detailsMap["field"] != "name" {
		t.Fatalf("expected field=name, got %v", detailsMap["field"])
	}
}

func TestErrorResponseShape_JSONKeysMatchStructTags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		WriteError(c, http.StatusBadRequest, "test_code", "test message", map[string]string{"field": "name"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	wantKeys := []string{"code", "message", "request_id", "timestamp", "details"}
	for _, key := range wantKeys {
		if _, exists := raw[key]; !exists {
			t.Fatalf("expected JSON key %q to be present", key)
		}
	}

	unexpectedKeys := []string{"Code", "Message", "Details", "RequestID", "Timestamp"}
	for _, key := range unexpectedKeys {
		if _, exists := raw[key]; exists {
			t.Fatalf("unexpected PascalCase key %q found — should use snake_case", key)
		}
	}
}
