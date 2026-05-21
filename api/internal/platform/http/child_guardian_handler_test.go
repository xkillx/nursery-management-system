package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseReasonPayloadValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing reason code", func(t *testing.T) {
		c, w := reasonTestContext(t, `{}`)
		_, ok := parseReasonPayload(c, "relationship_reason_required")
		if ok {
			t.Fatal("expected parseReasonPayload to fail")
		}
		assertErrorCode(t, w, http.StatusBadRequest, "relationship_reason_required")
	})

	t.Run("invalid reason code", func(t *testing.T) {
		c, w := reasonTestContext(t, `{"reason_code":"bad_code"}`)
		_, ok := parseReasonPayload(c, "relationship_reason_required")
		if ok {
			t.Fatal("expected parseReasonPayload to fail")
		}
		assertErrorCode(t, w, http.StatusBadRequest, "lifecycle_reason_invalid")
	})

	t.Run("other requires note", func(t *testing.T) {
		c, w := reasonTestContext(t, `{"reason_code":"other"}`)
		_, ok := parseReasonPayload(c, "relationship_reason_required")
		if ok {
			t.Fatal("expected parseReasonPayload to fail")
		}
		assertErrorCode(t, w, http.StatusBadRequest, "reason_note_required_for_other")
	})

	t.Run("valid reason payload", func(t *testing.T) {
		c, _ := reasonTestContext(t, `{"reason_code":"left_nursery","reason_note":"Moved away"}`)
		reason, ok := parseReasonPayload(c, "relationship_reason_required")
		if !ok {
			t.Fatal("expected parseReasonPayload to succeed")
		}
		if reason.Code != "left_nursery" {
			t.Fatalf("expected reason code left_nursery, got %s", reason.Code)
		}
		if reason.Note != "Moved away" {
			t.Fatalf("expected reason note Moved away, got %s", reason.Note)
		}
	})
}

func TestParseStatusFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("defaults to active", func(t *testing.T) {
		c := statusTestContext(t, "")
		status, ok := parseStatusFilter(c)
		if !ok {
			t.Fatal("expected parseStatusFilter to succeed")
		}
		if status != statusFilterActive {
			t.Fatalf("expected active, got %s", status)
		}
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		c := statusTestContext(t, "paused")
		status, ok := parseStatusFilter(c)
		if ok {
			t.Fatalf("expected failure, got %s", status)
		}
	})
}

func reasonTestContext(t *testing.T, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func statusTestContext(t *testing.T, status string) *gin.Context {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	path := "/"
	if status != "" {
		path = "/?status=" + status
	}
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	return c
}

func assertErrorCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	t.Helper()
	if w.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, w.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("parse response: %v", err)
	}

	code, _ := payload["code"].(string)
	if code != expectedCode {
		t.Fatalf("expected code %q, got %q", expectedCode, code)
	}
}
