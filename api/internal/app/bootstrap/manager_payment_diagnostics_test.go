package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/platform/dbtest"
)

func TestManagerPaymentStatus_Success_200(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-status", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["invoice_id"] != invoiceID.String() {
		t.Errorf("expected invoice_id %s, got %v", invoiceID, resp["invoice_id"])
	}
	if resp["status"] != "issued" {
		t.Errorf("expected status issued, got %v", resp["status"])
	}
	if resp["due_status"] != "due" {
		t.Errorf("expected due_status due, got %v", resp["due_status"])
	}
	if resp["checkout_retry_available"] != true {
		t.Errorf("expected checkout_retry_available true, got %v", resp["checkout_retry_available"])
	}
	if resp["checkout_retry_reason_code"] != "available" {
		t.Errorf("expected checkout_retry_reason_code available, got %v", resp["checkout_retry_reason_code"])
	}
	if resp["latest_payment_attempt"] != nil {
		t.Errorf("expected nil latest_payment_attempt, got %v", resp["latest_payment_attempt"])
	}
	if resp["latest_payment_event"] != nil {
		t.Errorf("expected nil latest_payment_event, got %v", resp["latest_payment_event"])
	}
}

func TestManagerPaymentStatus_MalformedInvoiceID_400(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/not-a-uuid/payment-status", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestManagerPaymentStatus_NotFound_404(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+uuid.New().String()+"/payment-status", h.managerToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

func TestManagerPaymentStatus_RoleGuards(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/invoices/%s/payment-status", invoiceID)

	for _, token := range []string{h.parentToken, h.pracToken} {
		w := doRequest(t, h.router, http.MethodGet, path, token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestManagerPaymentStatus_Unauthenticated(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/invoices/%s/payment-status", invoiceID)

	w := doRequest(t, h.router, http.MethodGet, path, "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestManagerPaymentStatus_PaymentFailed_RetryAvailable(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	childID := uuid.MustParse("d5000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("d6000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("d7000000-0000-0000-0000-000000000001")
	mappingID := uuid.MustParse("d8000000-0000-0000-0000-000000000001")
	runID := uuid.MustParse("d9000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("d9000000-0000-0000-0000-000000000002")
	billingMonth := dbtest.DateAt(2026, 5, 1)
	periodEnd := dbtest.DateAt(2026, 5, 31)

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "PF Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "PF Guardian", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	dbtest.InsertParentMapping(t, h.pool, mappingID, h.tenantID, h.branchID, h.parentMID, guardianID)

	_, err := h.pool.Exec(context.Background(),
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-pf')`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	_, err = h.pool.Exec(context.Background(),
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		  payment_failed_at, payment_status_updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'payment_failed', 'GBP', $6, $6, 5000, 0, 5000, $7, $8, 'INV-PF-001', 1, now(), $9, $10, now(), now() + interval '7 days', now(), now())`,
		invoiceID, h.tenantID, h.branchID, childID, billingMonth,
		runID, billingMonth, periodEnd, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-status", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["checkout_retry_available"] != true {
		t.Errorf("expected retry available for payment_failed, got %v", resp["checkout_retry_available"])
	}
	if resp["due_status"] != "due" {
		t.Errorf("expected due_status due for payment_failed, got %v", resp["due_status"])
	}
}

func TestManagerPaymentEvents_Success_200(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["limit"] != float64(50) {
		t.Errorf("expected limit 50, got %v", resp["limit"])
	}
	if resp["offset"] != float64(0) {
		t.Errorf("expected offset 0, got %v", resp["offset"])
	}
	items, ok := resp["items"].([]interface{})
	if !ok {
		t.Fatalf("expected items array, got %T", resp["items"])
	}
	if len(items) != 0 {
		t.Errorf("expected empty items for invoice with no reconciliation records, got %d", len(items))
	}
}

func TestManagerPaymentEvents_CustomPagination(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events?limit=10&offset=5", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["limit"] != float64(10) {
		t.Errorf("expected limit 10, got %v", resp["limit"])
	}
	if resp["offset"] != float64(5) {
		t.Errorf("expected offset 5, got %v", resp["offset"])
	}
}

func TestManagerPaymentEvents_MalformedInvoiceID_400(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/not-a-uuid/payment-events", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestManagerPaymentEvents_NotFound_404(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+uuid.New().String()+"/payment-events", h.managerToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

func TestManagerPaymentEvents_RoleGuards(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events", invoiceID)

	for _, token := range []string{h.parentToken, h.pracToken} {
		w := doRequest(t, h.router, http.MethodGet, path, token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestManagerPaymentEvents_Unauthenticated(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events", invoiceID)

	w := doRequest(t, h.router, http.MethodGet, path, "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestManagerPaymentEvents_InvalidLimit_400(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events?limit=0", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestManagerPaymentEvents_InvalidOffset_400(t *testing.T) {
	h := setupPaymentsHarness(t, nil)
	invoiceID := seedPayableInvoice(t, h)

	path := fmt.Sprintf("/api/v1/invoices/%s/payment-events?offset=-1", invoiceID)
	w := doRequest(t, h.router, http.MethodGet, path, h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestManagerPaymentDiagnostics_RouteInventory(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/invoices/:invoice_id/payment-status",
		"GET /api/v1/invoices/:invoice_id/payment-events",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Errorf("expected route %s to be registered", want)
		}
	}
}
