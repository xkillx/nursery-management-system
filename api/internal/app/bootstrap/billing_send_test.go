package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/platform/dbtest"
)

// seedIssuedInvoiceForSend inserts a site profile, a parent (optionally with an
// email), a parent-child link, and an invoice in the given status. Used by the
// manager-triggered resend route tests.
func seedIssuedInvoiceForSend(t *testing.T, h *billingIssueHarness, suffix string, childName string, status string, parentEmail string) (childID, invoiceID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	childID = uuid.MustParse(fmt.Sprintf("f5000000-0000-0000-0000-%012s", suffix))
	invoiceID = uuid.MustParse(fmt.Sprintf("f6000000-0000-0000-0000-%012s", suffix))

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, childName,
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	siteID := uuid.MustParse("f1000000-0000-0000-0000-000000000001")
	_, err := h.pool.Exec(ctx, `INSERT INTO site_profiles (id, tenant_id, branch_id, nursery_name, phone, email, website, address_street, address_city, address_postcode)
		VALUES ($1, $2, $3, 'Send Test Nursery', '020 0000 0000', 'billing@nursery.example.com', 'https://nursery.example.com', '1 High St', 'London', 'N1 1AA')`,
		siteID, h.tenantID, h.branchID)
	if err != nil {
		t.Fatalf("insert site profile: %v", err)
	}

	parentID := uuid.MustParse(fmt.Sprintf("f2000000-0000-0000-0000-%012s", suffix))
	_, err = h.pool.Exec(ctx, `INSERT INTO parents (id, tenant_id, branch_id, first_name, last_name, email, is_active)
		VALUES ($1, $2, $3, 'Jane', 'Doe', $4, true)`,
		parentID, h.tenantID, h.branchID, parentEmail)
	if err != nil {
		t.Fatalf("insert parent: %v", err)
	}

	linkID := uuid.MustParse(fmt.Sprintf("f3000000-0000-0000-0000-%012s", suffix))
	_, err = h.pool.Exec(ctx, `INSERT INTO parent_children (id, tenant_id, branch_id, parent_id, child_id) VALUES ($1, $2, $3, $4, $5)`,
		linkID, h.tenantID, h.branchID, parentID, childID)
	if err != nil {
		t.Fatalf("insert parent_children: %v", err)
	}

	billingMonth := dbtest.DateAt(2026, 5, 1)
	periodStart := billingMonth
	periodEnd := billingMonth.AddDate(0, 1, -1)

	if status == "draft" {
		_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
			period_start_date, period_end_date, subtotal_minor, funded_deduction_minor, total_due_minor)
			VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, $7, 1000, 0, 1000)`,
			invoiceID, h.tenantID, h.branchID, childID, billingMonth, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("insert draft invoice: %v", err)
		}
		return childID, invoiceID
	}

	runID := uuid.MustParse(fmt.Sprintf("f7000000-0000-0000-0000-%012s", suffix))
	_, err = h.pool.Exec(ctx, `INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-send-' || $7)`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID, suffix)
	if err != nil {
		t.Fatalf("insert invoice run: %v", err)
	}

	var voidedAt any
	var voidReason any
	if status == "void" {
		voidedAt = time.Now()
		voidReason = "test void"
	}
	var paymentFailedAt any
	if status == "payment_failed" {
		paymentFailedAt = time.Now()
	}

	_, err = h.pool.Exec(ctx, `INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		period_start_date, period_end_date, invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		voided_at, void_reason, payment_failed_at, subtotal_minor, funded_deduction_minor, total_due_minor)
		VALUES ($1, $2, $3, $4, $5, 'monthly', $6, 'GBP', $7, $8, $9, 1, $10, $10, now(), $11, $12, now(), now(), $13, $14, $15, 1000, 0, 1000)`,
		invoiceID, h.tenantID, h.branchID, childID, billingMonth,
		status, periodStart, periodEnd, "INV-SEND-"+suffix, runID, h.managerUID, h.managerMID,
		voidedAt, voidReason, paymentFailedAt)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	return childID, invoiceID
}

func countInvoiceResendRows(t *testing.T, h *billingIssueHarness, invoiceID uuid.UUID) int {
	t.Helper()
	var count int
	err := h.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM email_outbox WHERE tenant_id = $1 AND branch_id = $2 AND event_type = 'invoice_resend' AND entity_id = $3`,
		h.tenantID, h.branchID, invoiceID.String()).Scan(&count)
	if err != nil {
		t.Fatalf("count invoice_resend rows: %v", err)
	}
	return count
}

func TestInvoiceSendRouteInventory(t *testing.T) {
	h := setupBillingIssueHarness(t)

	found := false
	for _, route := range h.router.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/invoices/:invoice_id/send" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected route POST /api/v1/invoices/:invoice_id/send to be registered")
	}
}

func TestInvoiceSendUnauthenticated(t *testing.T) {
	h := setupBillingIssueHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/send", "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestInvoiceSendRoleGuards(t *testing.T) {
	h := setupBillingIssueHarness(t)

	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/send", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestInvoiceSendSuccess(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, invoiceID := seedIssuedInvoiceForSend(t, h, "000000000001", "Send Child", "issued", "parent-send@example.com")

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusAccepted)

	var resp sendInvoiceResponseTest
	decodeJSON(t, w, &resp)
	if resp.Status != "queued" {
		t.Fatalf("status = %q, want queued", resp.Status)
	}

	if n := countInvoiceResendRows(t, h, invoiceID); n != 1 {
		t.Fatalf("invoice_resend rows = %d, want 1", n)
	}

	var eventType, recipient, templateName string
	var templateVersion int
	var entityID string
	err := h.pool.QueryRow(context.Background(),
		`SELECT event_type, recipient, template_name, template_version, entity_id
		 FROM email_outbox WHERE tenant_id = $1 AND branch_id = $2 AND event_type = 'invoice_resend' AND entity_id = $3`,
		h.tenantID, h.branchID, invoiceID.String()).
		Scan(&eventType, &recipient, &templateName, &templateVersion, &entityID)
	if err != nil {
		t.Fatalf("query outbox row: %v", err)
	}
	if eventType != "invoice_resend" {
		t.Fatalf("event_type = %q, want invoice_resend", eventType)
	}
	if recipient != "parent-send@example.com" {
		t.Fatalf("recipient = %q, want parent-send@example.com", recipient)
	}
	if templateName != "issued" {
		t.Fatalf("template_name = %q, want issued", templateName)
	}
	if templateVersion != 2 {
		t.Fatalf("template_version = %d, want 2", templateVersion)
	}
	if entityID != invoiceID.String() {
		t.Fatalf("entity_id = %q, want %s", entityID, invoiceID)
	}
}

func TestInvoiceSendRejectsDraftAndVoid(t *testing.T) {
	for _, status := range []string{"draft", "void"} {
		h := setupBillingIssueHarness(t)
		_, invoiceID := seedIssuedInvoiceForSend(t, h, "000000000010", "Draft Void Child", status, "parent-dv@example.com")

		w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
		requireStatus(t, w, http.StatusConflict)
		requireErrorCode(t, w, "invoice_not_payable")

		if n := countInvoiceResendRows(t, h, invoiceID); n != 0 {
			t.Fatalf("status %s: invoice_resend rows = %d, want 0", status, n)
		}
	}
}

func TestInvoiceSendParentNoEmail(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, invoiceID := seedIssuedInvoiceForSend(t, h, "000000000020", "No Email Child", "issued", "")

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "parent_no_email")

	if n := countInvoiceResendRows(t, h, invoiceID); n != 0 {
		t.Fatalf("invoice_resend rows = %d, want 0", n)
	}
}

func TestInvoiceSendThrottledWithinCooldown(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, invoiceID := seedIssuedInvoiceForSend(t, h, "000000000030", "Throttle Child", "issued", "parent-throttle@example.com")

	_, err := h.pool.Exec(context.Background(), `INSERT INTO email_outbox (id, tenant_id, branch_id, idempotency_key, event_type, recipient, subject, template_name, template_version, payload_json, status, entity_id)
		VALUES ($1, $2, $3, 'invoice_resend_seed_1', 'invoice_resend', 'parent-throttle@example.com', 'subject', 'issued', 2, '{}', 'pending', $4)`,
		uuid.New(), h.tenantID, h.branchID, invoiceID.String())
	if err != nil {
		t.Fatalf("insert seed resend row: %v", err)
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusConflict)
	requireErrorCode(t, w, "invoice_resend_throttled")

	if n := countInvoiceResendRows(t, h, invoiceID); n != 1 {
		t.Fatalf("invoice_resend rows = %d, want 1 (no new row enqueued)", n)
	}
}

func TestInvoiceSendSecondResendAfterCooldown(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, invoiceID := seedIssuedInvoiceForSend(t, h, "000000000040", "Cooled Child", "issued", "parent-cooled@example.com")

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusAccepted)
	if n := countInvoiceResendRows(t, h, invoiceID); n != 1 {
		t.Fatalf("invoice_resend rows = %d, want 1", n)
	}

	// Backdate the first resend so the cooldown window no longer applies.
	_, err := h.pool.Exec(context.Background(),
		`UPDATE email_outbox SET created_at = now() - interval '1 hour' WHERE event_type = 'invoice_resend' AND entity_id = $1`, invoiceID.String())
	if err != nil {
		t.Fatalf("backdate resend row: %v", err)
	}

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusAccepted)

	if n := countInvoiceResendRows(t, h, invoiceID); n != 2 {
		t.Fatalf("invoice_resend rows = %d, want 2", n)
	}

	var keys []string
	rows, err := h.pool.Query(context.Background(),
		`SELECT idempotency_key FROM email_outbox WHERE event_type = 'invoice_resend' AND entity_id = $1 ORDER BY created_at`, invoiceID.String())
	if err != nil {
		t.Fatalf("query idempotency keys: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatalf("scan key: %v", err)
		}
		keys = append(keys, k)
	}
	if len(keys) != 2 {
		t.Fatalf("keys = %d, want 2", len(keys))
	}
	if keys[0] == keys[1] {
		t.Fatalf("idempotency keys must be distinct per attempt, got %q twice", keys[0])
	}
}

func TestInvoiceSendUnknownInvoice(t *testing.T) {
	h := setupBillingIssueHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/send", h.managerToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

type sendInvoiceResponseTest struct {
	Status string `json:"status"`
}
