package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"

	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

const testWebhookSecret = "whsec_test_integration_secret_key_12345"

type webhookHarness struct {
	router   *gin.Engine
	pool     *pgxpool.Pool
	tenantID uuid.UUID
	branchID uuid.UUID
	parentUID uuid.UUID
	parentMID uuid.UUID
	managerUID uuid.UUID
	managerMID uuid.UUID
}

func setupWebhookHarness(t *testing.T) *webhookHarness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &webhookHarness{
		pool:       pool,
		tenantID:   uuid.MustParse("d1000000-0000-0000-0000-000000000001"),
		branchID:   uuid.MustParse("d2000000-0000-0000-0000-000000000001"),
		parentUID:  uuid.MustParse("d3000000-0000-0000-0000-000000000001"),
		parentMID:  uuid.MustParse("d4000000-0000-0000-0000-000000000001"),
		managerUID: uuid.MustParse("d3000000-0000-0000-0000-000000000002"),
		managerMID: uuid.MustParse("d4000000-0000-0000-0000-000000000002"),
	}

	dbtest.InsertTenant(t, pool, h.tenantID, "Webhook Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Webhook Branch")
	dbtest.InsertUser(t, pool, h.parentUID, "wh-parent@test.com", "hash", true)
	dbtest.InsertUser(t, pool, h.managerUID, "wh-mgr@test.com", "hash", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, h.parentMID, h.tenantID, h.branchID, h.parentUID, "parent", true)

	cfg := testConfig()
	cfg.StripeSecretKey = "sk_test_wh"
	cfg.StripeWebhookSecret = testWebhookSecret

	fakeProvider := &fakeCheckoutProvider{
		result: domain.CheckoutSessionResult{
			CheckoutSessionID: "cs_test_wh",
			CheckoutURL:       "https://checkout.stripe.com/wh",
			PaymentIntentID:   "pi_test_wh",
		},
	}

	h.router = BootstrapWithOptions(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), pool, BootstrapOptions{
		CheckoutProvider: fakeProvider,
	})

	return h
}

func seedInvoiceWithAttempt(t *testing.T, h *webhookHarness, invoiceStatus string, amountMinor int) (invoiceID, attemptID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	childID := uuid.MustParse("d5000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("d6000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("d7000000-0000-0000-0000-000000000001")
	mappingID := uuid.MustParse("d8000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "WH Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "WH Guardian", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	dbtest.InsertParentMapping(t, h.pool, mappingID, h.tenantID, h.branchID, h.parentMID, guardianID)

	runID := uuid.New()
	invoiceID = uuid.New()
	billingMonth := dbtest.DateAt(2026, 6, 1)
	periodEnd := dbtest.DateAt(2026, 6, 30)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-wh')`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	dueAt := "now() + interval '7 days'"
	if invoiceStatus == "overdue" {
		dueAt = "now() - interval '7 days'"
	}
	extraCols := ""
	extraVals := ""
	if invoiceStatus == "paid" {
		extraCols = ", paid_at, amount_paid_minor"
		extraVals = fmt.Sprintf(", now(), %d", amountMinor)
	}
	if invoiceStatus == "payment_failed" {
		extraCols = ", payment_failed_at"
		extraVals = ", now()"
	}

	_, err = h.pool.Exec(ctx, fmt.Sprintf(
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at%s)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', $6, 'GBP', $7, $7, %d, 0, %d, $8, $9, 'INV-WH-001', 1, now(), $10, $11, now(), %s%s)`,
		extraCols, amountMinor, amountMinor, dueAt, extraVals),
		invoiceID, h.tenantID, h.branchID, childID, billingMonth, invoiceStatus,
		runID, billingMonth, periodEnd, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert invoice (%s): %v", invoiceStatus, err)
	}

	attemptID = uuid.New()
	_, err = h.pool.Exec(ctx,
		`INSERT INTO payment_attempts (id, tenant_id, branch_id, invoice_id, initiated_by_user_id, initiated_by_membership_id, request_id, status, amount_minor, currency_code, stripe_checkout_session_id, stripe_checkout_url)
		 VALUES ($1, $2, $3, $4, $5, $6, 'req-wh-attempt', 'checkout_created', $7, 'GBP', $8, $9)`,
		attemptID, h.tenantID, h.branchID, invoiceID, h.parentUID, h.parentMID, amountMinor, "cs_test_wh", "https://checkout.stripe.com/test")
	if err != nil {
		t.Fatalf("insert attempt: %v", err)
	}

	return invoiceID, attemptID
}

func buildCheckoutCompletedPayload(t *testing.T, eventID, sessionID string, paymentStatus stripe.CheckoutSessionPaymentStatus, amountTotal int64, currency string, metadata map[string]string) ([]byte, string) {
	t.Helper()

	metaJSON := "{}"
	if metadata != nil {
		mj, _ := json.Marshal(metadata)
		metaJSON = string(mj)
	}

	payload := []byte(fmt.Sprintf(`{"id":%q,"object":"event","type":"checkout.session.completed","api_version":%q,"livemode":false,"created":%d,"data":{"object":{"id":%q,"payment_status":%q,"amount_total":%d,"currency":%q,"metadata":%s}}}`,
		eventID, stripe.APIVersion, time.Now().Unix(), sessionID, string(paymentStatus), amountTotal, currency, metaJSON))

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    testWebhookSecret,
		Timestamp: time.Now(),
	})

	return payload, signed.Header
}

func buildCheckoutEventPayload(t *testing.T, eventType stripe.EventType, eventID, sessionID string, amountTotal int64, currency string, metadata map[string]string) ([]byte, string) {
	t.Helper()

	metaJSON := "{}"
	if metadata != nil {
		mj, _ := json.Marshal(metadata)
		metaJSON = string(mj)
	}

	payload := []byte(fmt.Sprintf(`{"id":%q,"object":"event","type":%q,"api_version":%q,"livemode":false,"created":%d,"data":{"object":{"id":%q,"payment_status":"paid","amount_total":%d,"currency":%q,"metadata":%s}}}`,
		eventID, string(eventType), stripe.APIVersion, time.Now().Unix(), sessionID, amountTotal, currency, metaJSON))

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    testWebhookSecret,
		Timestamp: time.Now(),
	})

	return payload, signed.Header
}

func doWebhookRequest(t *testing.T, router *gin.Engine, payload []byte, sigHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhooks", nil)
	req.Header.Set("Stripe-Signature", sigHeader)
	req.Body = io.NopCloser(newBytesReader(payload))
	req.ContentLength = int64(len(payload))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func doWebhookRawRequest(t *testing.T, router *gin.Engine, body []byte, sigHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhooks", newBytesReader(body))
	req.Header.Set("Stripe-Signature", sigHeader)
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

type bytesReader struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReader { return &bytesReader{data: data} }
func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
func (r *bytesReader) Close() error { return nil }

func requireWebhookStatus(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode webhook response: %v", err)
	}
	if resp.Status != want {
		t.Fatalf("expected webhook status %q, got %q (body: %s)", want, resp.Status, w.Body.String())
	}
}

func fetchInvoiceStatus(t *testing.T, pool *pgxpool.Pool, invoiceID uuid.UUID) string {
	t.Helper()
	var status string
	err := pool.QueryRow(context.Background(), "SELECT status FROM invoices WHERE id = $1", invoiceID).Scan(&status)
	if err != nil {
		t.Fatalf("fetch invoice status: %v", err)
	}
	return status
}

func fetchAttemptStatus(t *testing.T, pool *pgxpool.Pool, attemptID uuid.UUID) string {
	t.Helper()
	var status string
	err := pool.QueryRow(context.Background(), "SELECT status FROM payment_attempts WHERE id = $1", attemptID).Scan(&status)
	if err != nil {
		t.Fatalf("fetch attempt status: %v", err)
	}
	return status
}

func countReconciliationRecords(t *testing.T, pool *pgxpool.Pool, invoiceID uuid.UUID) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM payment_reconciliation_records WHERE invoice_id = $1", invoiceID).Scan(&count)
	if err != nil {
		t.Fatalf("count reconciliation records: %v", err)
	}
	return count
}

func countWebhookEvents(t *testing.T, pool *pgxpool.Pool, stripeEventID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM stripe_webhook_events WHERE stripe_event_id = $1", stripeEventID).Scan(&count)
	if err != nil {
		t.Fatalf("count webhook events: %v", err)
	}
	return count
}

func countSystemAuditRows(t *testing.T, pool *pgxpool.Pool, invoiceID uuid.UUID) int {
	t.Helper()
	var count int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FROM audit_logs WHERE action_type = 'invoice_payment_status_updated' AND action_entity_id = $1 AND actor_user_id IS NULL",
		invoiceID).Scan(&count)
	if err != nil {
		t.Fatalf("count system audit rows: %v", err)
	}
	return count
}

func TestWebhook_RouteRegistered(t *testing.T) {
	h := setupWebhookHarness(t)
	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}
	if _, ok := have["POST /api/v1/stripe/webhooks"]; !ok {
		t.Fatal("expected POST /api/v1/stripe/webhooks to be registered")
	}
}

func TestWebhook_WorksNoBearerToken(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, _ := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": uuid.New().String(),
	}
	payload, sig := buildCheckoutCompletedPayload(t, "evt_noauth_"+uuid.New().String()[:8], "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)
	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
}

func TestWebhook_UnconfiguredReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	cfg := testConfig()
	cfg.StripeSecretKey = ""
	cfg.StripeWebhookSecret = ""
	router := BootstrapWithOptions(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), pool, BootstrapOptions{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stripe/webhooks", newBytesReader([]byte(`{}`)))
	req.Header.Set("Stripe-Signature", "t=1,v1=abc")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	requireStatus(t, w, http.StatusServiceUnavailable)
	requireErrorCode(t, w, "payment_provider_unconfigured")
}

func TestWebhook_InvalidSignatureReturns400(t *testing.T) {
	h := setupWebhookHarness(t)
	w := doWebhookRawRequest(t, h.router, []byte(`{"id":"evt_bad_sig"}`), "t=1,v1=invalidsignature")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "payment_webhook_invalid_signature")

	if countWebhookEvents(t, h.pool, "evt_bad_sig") != 0 {
		t.Fatal("invalid signature should not store webhook event")
	}
}

func TestWebhook_CheckoutCompletedPaid_MarksPaid(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_completed_paid_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "paid" {
		t.Fatalf("expected invoice paid, got %s", s)
	}
	if s := fetchAttemptStatus(t, h.pool, attemptID); s != "paid" {
		t.Fatalf("expected attempt paid, got %s", s)
	}
	if n := countReconciliationRecords(t, h.pool, invoiceID); n != 1 {
		t.Fatalf("expected 1 reconciliation record, got %d", n)
	}
	if n := countSystemAuditRows(t, h.pool, invoiceID); n != 1 {
		t.Fatalf("expected 1 system audit row, got %d", n)
	}
}

func TestWebhook_CheckoutCompletedUnpaid_Ignores(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_completed_unpaid_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusUnpaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "ignored")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "issued" {
		t.Fatalf("expected invoice unchanged, got %s", s)
	}
}

func TestWebhook_AsyncPaymentSucceeded_MarksPaid(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_async_success_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutEventPayload(t, stripe.EventTypeCheckoutSessionAsyncPaymentSucceeded, evtID, "cs_test_wh", 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "paid" {
		t.Fatalf("expected paid, got %s", s)
	}
}

func TestWebhook_AsyncPaymentFailed_MarksPaymentFailed(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_async_fail_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutEventPayload(t, stripe.EventTypeCheckoutSessionAsyncPaymentFailed, evtID, "cs_test_wh", 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "payment_failed" {
		t.Fatalf("expected payment_failed, got %s", s)
	}
	if s := fetchAttemptStatus(t, h.pool, attemptID); s != "payment_failed" {
		t.Fatalf("expected attempt payment_failed, got %s", s)
	}
}

func TestWebhook_Expired_MarksPaymentFailed(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_expired_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutEventPayload(t, stripe.EventTypeCheckoutSessionExpired, evtID, "cs_test_wh", 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "payment_failed" {
		t.Fatalf("expected payment_failed, got %s", s)
	}
	if s := fetchAttemptStatus(t, h.pool, attemptID); s != "expired" {
		t.Fatalf("expected attempt expired, got %s", s)
	}
}

func TestWebhook_DuplicateEvent_ReturnsDuplicate(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_dup_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w1 := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w1, http.StatusOK)
	requireWebhookStatus(t, w1, "processed")

	w2 := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w2, http.StatusOK)
	requireWebhookStatus(t, w2, "duplicate")

	if n := countReconciliationRecords(t, h.pool, invoiceID); n != 1 {
		t.Fatalf("expected 1 reconciliation record (no duplicate), got %d", n)
	}
	if n := countSystemAuditRows(t, h.pool, invoiceID); n != 1 {
		t.Fatalf("expected 1 audit row (no duplicate), got %d", n)
	}
}

func TestWebhook_AlreadyPaid_IgnoresFailure(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "paid", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_fail_paid_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutEventPayload(t, stripe.EventTypeCheckoutSessionAsyncPaymentFailed, evtID, "cs_test_wh", 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "paid" {
		t.Fatalf("expected paid unchanged, got %s", s)
	}
}

func TestWebhook_PaymentFailedToPaid(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "payment_failed", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_recover_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "paid" {
		t.Fatalf("expected payment_failed -> paid, got %s", s)
	}
}

func TestWebhook_AlreadyPaymentFailed_NoTimestampRewrite(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "payment_failed", 5000)

	var originalFailedAt time.Time
	err := h.pool.QueryRow(context.Background(), "SELECT payment_failed_at FROM invoices WHERE id = $1", invoiceID).Scan(&originalFailedAt)
	if err != nil {
		t.Fatalf("fetch original payment_failed_at: %v", err)
	}

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_dup_fail_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutEventPayload(t, stripe.EventTypeCheckoutSessionAsyncPaymentFailed, evtID, "cs_test_wh", 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	var newFailedAt time.Time
	err = h.pool.QueryRow(context.Background(), "SELECT payment_failed_at FROM invoices WHERE id = $1", invoiceID).Scan(&newFailedAt)
	if err != nil {
		t.Fatalf("fetch new payment_failed_at: %v", err)
	}
	if !newFailedAt.Equal(originalFailedAt) {
		t.Fatalf("payment_failed_at should not be rewritten: original %v, new %v", originalFailedAt, newFailedAt)
	}
}

func TestWebhook_AmountMismatch_Rejected(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_amt_mismatch_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 9999, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "rejected")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "issued" {
		t.Fatalf("expected invoice unchanged, got %s", s)
	}
}

func TestWebhook_CurrencyMismatch_Rejected(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "issued", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_cur_mismatch_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "usd", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "rejected")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "issued" {
		t.Fatalf("expected invoice unchanged, got %s", s)
	}
}

func TestWebhook_MissingMetadata_Rejected(t *testing.T) {
	h := setupWebhookHarness(t)

	evtID := "evt_no_meta_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", nil)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	requireWebhookStatus(t, w, "rejected")
}

func TestWebhook_UnknownAttempt_Rejected(t *testing.T) {
	h := setupWebhookHarness(t)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        uuid.New().String(),
		"payment_attempt_id": uuid.New().String(),
	}
	evtID := "evt_unknown_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "rejected")
}

func TestWebhook_PaymentIntentEvent_Ignored(t *testing.T) {
	h := setupWebhookHarness(t)

	eventID := "evt_pi_" + uuid.New().String()[:8]
	payload := []byte(fmt.Sprintf(`{"id":%q,"object":"event","type":"payment_intent.succeeded","api_version":%q,"livemode":false,"created":%d,"data":{"object":{"id":"pi_test"}}}`,
		eventID, stripe.APIVersion, time.Now().Unix()))

	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    testWebhookSecret,
		Timestamp: time.Now(),
	})

	w := doWebhookRawRequest(t, h.router, payload, signed.Header)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "ignored")
}

func TestWebhook_OverdueToPaid(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "overdue", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_overdue_paid_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	if s := fetchInvoiceStatus(t, h.pool, invoiceID); s != "paid" {
		t.Fatalf("expected overdue -> paid, got %s", s)
	}
}

func TestWebhook_AlreadyPaid_IgnoresLaterSuccess(t *testing.T) {
	h := setupWebhookHarness(t)
	invoiceID, attemptID := seedInvoiceWithAttempt(t, h, "paid", 5000)

	meta := map[string]string{
		"tenant_id":          h.tenantID.String(),
		"branch_id":         h.branchID.String(),
		"invoice_id":        invoiceID.String(),
		"payment_attempt_id": attemptID.String(),
	}
	evtID := "evt_late_success_" + uuid.New().String()[:8]
	payload, sig := buildCheckoutCompletedPayload(t, evtID, "cs_test_wh", stripe.CheckoutSessionPaymentStatusPaid, 5000, "gbp", meta)

	w := doWebhookRawRequest(t, h.router, payload, sig)
	requireStatus(t, w, http.StatusOK)
	requireWebhookStatus(t, w, "processed")

	// invoice stays paid, no second reconciliation
	if n := countReconciliationRecords(t, h.pool, invoiceID); n != 1 {
		t.Fatalf("expected 1 reconciliation (ignored), got %d", n)
	}
}
