package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

type fakeCheckoutProvider struct {
	result domain.CheckoutSessionResult
	err    error
}

func (f *fakeCheckoutProvider) CreateCheckoutSession(_ context.Context, _ domain.CheckoutSessionCreateParams) (domain.CheckoutSessionResult, error) {
	return f.result, f.err
}

type paymentsHarness struct {
	router       *gin.Engine
	pool         *pgxpool.Pool
	tokens       *tokens.TokenManager
	tenantID     uuid.UUID
	branchID     uuid.UUID
	parentUID    uuid.UUID
	parentMID    uuid.UUID
	managerUID   uuid.UUID
	managerMID   uuid.UUID
	parentToken  string
	managerToken string
	pracToken    string
}

func setupPaymentsHarness(t *testing.T, provider *fakeCheckoutProvider) *paymentsHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &paymentsHarness{
		pool:       pool,
		tokens:     tokens.NewTokenManager("access-secret", "refresh-secret", 15, 720),
		tenantID:   uuid.MustParse("c1000000-0000-0000-0000-000000000001"),
		branchID:   uuid.MustParse("c2000000-0000-0000-0000-000000000001"),
		parentUID:  uuid.MustParse("c3000000-0000-0000-0000-000000000001"),
		parentMID:  uuid.MustParse("c4000000-0000-0000-0000-000000000001"),
		managerUID: uuid.MustParse("c3000000-0000-0000-0000-000000000002"),
		managerMID: uuid.MustParse("c4000000-0000-0000-0000-000000000002"),
	}

	dbtest.InsertTenant(t, pool, h.tenantID, "Payments Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Payments Branch")
	dbtest.InsertUser(t, pool, h.parentUID, "pay-parent@test.com", "hash", true)
	dbtest.InsertUser(t, pool, h.managerUID, "pay-mgr@test.com", "hash", true)
	pracUID := uuid.MustParse("c3000000-0000-0000-0000-000000000003")
	pracMID := uuid.MustParse("c4000000-0000-0000-0000-000000000003")
	dbtest.InsertUser(t, pool, pracUID, "pay-prac@test.com", "hash", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, h.parentMID, h.tenantID, h.branchID, h.parentUID, "parent", true)
	dbtest.InsertMembership(t, pool, pracMID, h.tenantID, h.branchID, pracUID, "practitioner", true)

	cfg := testConfig()
	cfg.StripeSecretKey = "" // no real Stripe

	opts := BootstrapOptions{}
	if provider != nil {
		opts.CheckoutProvider = provider
	}

	h.router = BootstrapWithOptions(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), pool, opts)

	h.parentToken = mustAccessToken(t, h.tokens, h.parentUID, h.parentMID, h.tenantID, h.branchID, "parent")
	h.managerToken = mustAccessToken(t, h.tokens, h.managerUID, h.managerMID, h.tenantID, h.branchID, "manager")
	h.pracToken = mustAccessToken(t, h.tokens, pracUID, pracMID, h.tenantID, h.branchID, "practitioner")

	return h
}

func seedPayableInvoice(t *testing.T, h *paymentsHarness) uuid.UUID {
	t.Helper()
	return seedPayableInvoiceWithStatus(t, h, "issued")
}

func seedPayableInvoiceWithStatus(t *testing.T, h *paymentsHarness, status string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	childID := uuid.MustParse("c5000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("c7000000-0000-0000-0000-000000000001")
	mappingID := uuid.MustParse("c8000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Pay Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Pay Guardian", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	dbtest.InsertParentMapping(t, h.pool, mappingID, h.tenantID, h.branchID, h.parentMID, guardianID)

	runID := uuid.MustParse("c9000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("c9000000-0000-0000-0000-000000000002")
	billingMonth := dbtest.DateAt(2026, 5, 1)
	periodEnd := dbtest.DateAt(2026, 5, 31)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-pay')`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	dueAt := "now() + interval '7 days'"
	if status == "overdue" {
		dueAt = "now() - interval '7 days'"
	}
	extraCols := ""
	extraVals := ""
	if status == "paid" {
		extraCols = ", paid_at, amount_paid_minor"
		extraVals = ", now(), 5000"
	}

	_, err = h.pool.Exec(ctx, fmt.Sprintf(
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at%s)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', $6, 'GBP', $7, $7, 5000, 0, 5000, $8, $9, 'INV-PAY-001', 1, now(), $10, $11, now(), %s%s)`,
		extraCols, dueAt, extraVals),
		invoiceID, h.tenantID, h.branchID, childID, billingMonth, status,
		runID, billingMonth, periodEnd, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert invoice (%s): %v", status, err)
	}

	return invoiceID
}

func TestPaymentsRouteInventory(t *testing.T) {
	h := setupBillingHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"POST /api/v1/parent/invoices/:invoice_id/checkout-sessions",
		"POST /api/v1/stripe/webhooks",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestPaymentsCheckoutSessionRoleGuards(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	for _, token := range []string{h.managerToken, h.pracToken} {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/parent/invoices/00000000-0000-0000-0000-000000000099/checkout-sessions", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestPaymentsCheckoutSessionUnauthenticated(t *testing.T) {
	h := setupPaymentsHarness(t, nil)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/parent/invoices/00000000-0000-0000-0000-000000000099/checkout-sessions", "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestPaymentsCheckoutSession_StripeUnconfigured_503(t *testing.T) {
	h := setupPaymentsHarness(t, nil) // nil provider = unconfigured

	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/parent/invoices/%s/checkout-sessions", invoiceID)

	w := doRequest(t, h.router, http.MethodPost, path, h.parentToken, "")
	requireStatus(t, w, http.StatusServiceUnavailable)
	requireErrorCode(t, w, "payment_provider_unconfigured")
}

func TestPaymentsCheckoutSession_Success_201(t *testing.T) {
	provider := &fakeCheckoutProvider{
		result: domain.CheckoutSessionResult{
			CheckoutSessionID: "cs_test_success",
			CheckoutURL:       "https://checkout.stripe.com/success",
			PaymentIntentID:   "pi_test_123",
		},
	}
	h := setupPaymentsHarness(t, provider)

	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/parent/invoices/%s/checkout-sessions", invoiceID)

	w := doRequest(t, h.router, http.MethodPost, path, h.parentToken, "")
	requireStatus(t, w, http.StatusCreated)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp["checkout_session_id"] != "cs_test_success" {
		t.Fatalf("expected cs_test_success, got %v", resp["checkout_session_id"])
	}
	if resp["checkout_url"] != "https://checkout.stripe.com/success" {
		t.Fatalf("expected checkout URL, got %v", resp["checkout_url"])
	}
	if resp["payment_attempt_id"] == nil || resp["payment_attempt_id"] == "" {
		t.Fatal("expected payment_attempt_id")
	}
}

func TestPaymentsCheckoutSession_MalformedInvoiceID_400(t *testing.T) {
	h := setupPaymentsHarness(t, &fakeCheckoutProvider{
		result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"},
	})

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/parent/invoices/not-a-uuid/checkout-sessions", h.parentToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestPaymentsCheckoutSession_NotFound_404(t *testing.T) {
	h := setupPaymentsHarness(t, &fakeCheckoutProvider{
		result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"},
	})

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/parent/invoices/"+uuid.New().String()+"/checkout-sessions", h.parentToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

func TestPaymentsCheckoutSession_DraftNotPayable_409(t *testing.T) {
	h := setupPaymentsHarness(t, &fakeCheckoutProvider{
		result: domain.CheckoutSessionResult{CheckoutSessionID: "cs_test", CheckoutURL: "https://checkout.stripe.com/test"},
	})

	// Insert draft invoice directly — need a real child for FK
	ctx := context.Background()
	draftChildID := uuid.MustParse("c5000000-0000-0000-0000-000000000099")
	dbtest.InsertChild(t, h.pool, draftChildID, h.tenantID, h.branchID, "Draft Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)

	invoiceID := uuid.New()
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 5000, 0, 5000, $6, $7)`,
		invoiceID, h.tenantID, h.branchID, draftChildID,
		dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 30))
	if err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	path := fmt.Sprintf("/api/v1/parent/invoices/%s/checkout-sessions", invoiceID)
	w := doRequest(t, h.router, http.MethodPost, path, h.parentToken, "")
	requireStatus(t, w, http.StatusNotFound) // draft is not parent-visible, so 404
	requireErrorCode(t, w, "invoice_not_found")
}

func TestPaymentsCheckoutSession_ProviderError_502(t *testing.T) {
	provider := &fakeCheckoutProvider{
		err: fmt.Errorf("stripe API connection refused"),
	}
	h := setupPaymentsHarness(t, provider)

	invoiceID := seedPayableInvoice(t, h)
	path := fmt.Sprintf("/api/v1/parent/invoices/%s/checkout-sessions", invoiceID)

	w := doRequest(t, h.router, http.MethodPost, path, h.parentToken, "")
	requireStatus(t, w, http.StatusBadGateway)
	requireErrorCode(t, w, "payment_provider_error")
}
