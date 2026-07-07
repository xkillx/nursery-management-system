package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetInfo_DefaultValues(t *testing.T) {
	info := GetInfo()

	if info.Version != "dev" {
		t.Fatalf("expected version 'dev', got %q", info.Version)
	}
	if info.Commit != "unknown" {
		t.Fatalf("expected commit 'unknown', got %q", info.Commit)
	}
	if info.Date != "unknown" {
		t.Fatalf("expected date 'unknown', got %q", info.Date)
	}
}

func TestHandler_ReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/version", Handler())

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body Info
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse response body: %v", err)
	}

	if body.Version == "" {
		t.Fatal("expected version field to be present")
	}
	if body.Commit == "" {
		t.Fatal("expected commit field to be present")
	}
	if body.Date == "" {
		t.Fatal("expected date field to be present")
	}
}
