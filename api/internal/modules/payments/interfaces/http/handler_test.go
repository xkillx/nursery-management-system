package httpayment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/application"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/tenant"
)

type stubUseCase struct {
	result application.CreateCheckoutSessionResult
	err    error
}

func (s *stubUseCase) Execute(_ context.Context, _, _, _, _, _, _ string) (application.CreateCheckoutSessionResult, error) {
	return s.result, s.err
}

func setupRouter(uc CreateCheckoutSessionUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(tenant.AuthContextKey, tenant.AuthorizationContext{
			UserID:       "00000000-0000-0000-0000-000000000001",
			MembershipID: "00000000-0000-0000-0000-000000000002",
			TenantID:     "00000000-0000-0000-0000-000000000003",
			BranchID:     "00000000-0000-0000-0000-000000000004",
			Role:         "parent",
			RequestID:    "test-req-1",
		})
		c.Next()
	})
	r.Use(httpserver.RequestIDMiddleware())
	h := &Handler{createCheckoutSession: uc}
	parent := r.Group("/api/v1/parent")
	h.RegisterParentRoutes(parent)
	return r
}

func TestHandler_ParentSuccess_201(t *testing.T) {
	uc := &stubUseCase{
		result: application.CreateCheckoutSessionResult{
			CheckoutSessionID: "cs_test_123",
			CheckoutURL:       "https://checkout.stripe.com/test",
			PaymentAttemptID:  uuid.New().String(),
		},
	}
	r := setupRouter(uc)

	invoiceID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+invoiceID.String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp createCheckoutSessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.CheckoutSessionID != "cs_test_123" {
		t.Fatalf("expected cs_test_123, got %s", resp.CheckoutSessionID)
	}
	if resp.CheckoutURL != "https://checkout.stripe.com/test" {
		t.Fatalf("expected checkout URL, got %s", resp.CheckoutURL)
	}
}

func TestHandler_MalformedInvoiceID_400(t *testing.T) {
	r := setupRouter(&stubUseCase{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/not-a-uuid/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_NotFound_404(t *testing.T) {
	r := setupRouter(&stubUseCase{err: domainerrors.NotFound("invoice", "Invoice not found.")})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_NotPayable_409(t *testing.T) {
	r := setupRouter(&stubUseCase{err: domainerrors.Conflict("invoice_not_payable", "Invoice is not payable.")})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestHandler_ProviderUnconfigured_503(t *testing.T) {
	r := setupRouter(&stubUseCase{err: domainerrors.New("payment_provider_unconfigured", "Payment provider is not configured.")})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandler_ProviderError_502(t *testing.T) {
	r := setupRouter(&stubUseCase{err: domainerrors.New("payment_provider_error", "Payment provider failed.")})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestHandler_Unauthenticated_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpserver.RequestIDMiddleware())
	uc := &stubUseCase{result: application.CreateCheckoutSessionResult{}}
	h := &Handler{createCheckoutSession: uc}
	parent := r.Group("/api/v1/parent")
	h.RegisterParentRoutes(parent)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", bytes.NewReader(nil))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
