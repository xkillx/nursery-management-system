package queryparams

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseFilterParams_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?status=active", nil)

	allowed := map[string]string{"status": "string"}
	filters := ParseFilterParams(c, allowed)

	if filters["status"] != "active" {
		t.Errorf("expected status=active, got %q", filters["status"])
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestParseFilterParams_UnknownParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?unknown=value", nil)

	allowed := map[string]string{"status": "string"}
	filters := ParseFilterParams(c, allowed)

	if len(filters) != 0 {
		t.Errorf("expected 0 filters for unknown param, got %d", len(filters))
	}
}

func TestParseFilterParams_Partial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?status=active&bogus=x", nil)

	allowed := map[string]string{"status": "string"}
	filters := ParseFilterParams(c, allowed)

	if filters["status"] != "active" {
		t.Errorf("expected status=active, got %q", filters["status"])
	}
	if len(filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(filters))
	}
}

func TestParseFilterParams_MultipleFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?status=active&room_id=abc-123", nil)

	allowed := map[string]string{"status": "string", "room_id": "uuid"}
	filters := ParseFilterParams(c, allowed)

	if filters["status"] != "active" {
		t.Errorf("expected status=active, got %q", filters["status"])
	}
	if filters["room_id"] != "abc-123" {
		t.Errorf("expected room_id=abc-123, got %q", filters["room_id"])
	}
	if len(filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(filters))
	}
}

func TestParseFilterParams_EmptyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?status=", nil)

	allowed := map[string]string{"status": "string"}
	filters := ParseFilterParams(c, allowed)

	if len(filters) != 0 {
		t.Errorf("expected 0 filters for empty value, got %d", len(filters))
	}
}

func TestParseFilterParams_NoParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	allowed := map[string]string{"status": "string", "room_id": "uuid"}
	filters := ParseFilterParams(c, allowed)

	if len(filters) != 0 {
		t.Errorf("expected 0 filters with no params, got %d", len(filters))
	}
}
