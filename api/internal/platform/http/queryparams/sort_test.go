package queryparams

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseSortParams_HappyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?sort=name:asc", nil)

	allowed := map[string][]string{"name": {"asc", "desc"}}
	expr, err := ParseSortParams(c, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Field != "name" || expr.Direction != "asc" {
		t.Errorf("expected {name, asc}, got %+v", expr)
	}
	if expr.IsZero() {
		t.Error("expected non-zero sort expression")
	}
}

func TestParseSortParams_Descending(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?sort=created_at:desc", nil)

	allowed := map[string][]string{"created_at": {"asc", "desc"}}
	expr, err := ParseSortParams(c, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expr.Field != "created_at" || expr.Direction != "desc" {
		t.Errorf("expected {created_at, desc}, got %+v", expr)
	}
}

func TestParseSortParams_Omitted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	allowed := map[string][]string{"name": {"asc", "desc"}}
	expr, err := ParseSortParams(c, allowed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !expr.IsZero() {
		t.Errorf("expected zero-value sort expression, got %+v", expr)
	}
}

func TestParseSortParams_InvalidField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?sort=invalid:asc", nil)

	allowed := map[string][]string{"name": {"asc", "desc"}}
	_, err := ParseSortParams(c, allowed)
	if err == nil {
		t.Error("expected error for invalid sort field")
	}
}

func TestParseSortParams_InvalidDirection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?sort=name:up", nil)

	allowed := map[string][]string{"name": {"asc", "desc"}}
	_, err := ParseSortParams(c, allowed)
	if err == nil {
		t.Error("expected error for invalid sort direction")
	}
}

func TestParseSortParams_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?sort=badformat", nil)

	allowed := map[string][]string{"name": {"asc", "desc"}}
	_, err := ParseSortParams(c, allowed)
	if err == nil {
		t.Error("expected error for invalid sort format")
	}
}
