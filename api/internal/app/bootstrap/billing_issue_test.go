package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	"nursery-management-system/api/internal/platform/dbtest"
)

type billingIssueHarness struct {
	router            *gin.Engine
	pool              *pgxpool.Pool
	tokens            *authtokens.TokenManager
	tenantID          uuid.UUID
	branchID          uuid.UUID
	managerUID        uuid.UUID
	managerMID        uuid.UUID
	managerToken      string
	practitionerToken string
	parentToken       string
}

func setupBillingIssueHarness(t *testing.T) *billingIssueHarness {
	t.Helper()

	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &billingIssueHarness{
		router:     Bootstrap(testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), pool),
		pool:       pool,
		tokens:     authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720),
		tenantID:   uuid.MustParse("e1000000-0000-0000-0000-000000000001"),
		branchID:   uuid.MustParse("e2000000-0000-0000-0000-000000000001"),
		managerUID: uuid.MustParse("e3000000-0000-0000-0000-000000000001"),
		managerMID: uuid.MustParse("e4000000-0000-0000-0000-000000000001"),
	}

	practitionerUID := uuid.MustParse("e3000000-0000-0000-0000-000000000002")
	practitionerMID := uuid.MustParse("e4000000-0000-0000-0000-000000000002")
	parentUID := uuid.MustParse("e3000000-0000-0000-0000-000000000003")
	parentMID := uuid.MustParse("e4000000-0000-0000-0000-000000000003")

	dbtest.InsertTenant(t, pool, h.tenantID, "Issue Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Issue Branch")
	dbtest.InsertUser(t, pool, h.managerUID, "issue-mgr@example.com", "hash", true)
	dbtest.InsertUser(t, pool, practitionerUID, "issue-prac@example.com", "hash", true)
	dbtest.InsertUser(t, pool, parentUID, "issue-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, practitionerMID, h.tenantID, h.branchID, practitionerUID, "practitioner", true)
	dbtest.InsertMembership(t, pool, parentMID, h.tenantID, h.branchID, parentUID, "parent", true)

	h.managerToken = mustAccessToken(t, h.tokens, h.managerUID, h.managerMID, h.tenantID, h.branchID, "manager")
	h.practitionerToken = mustAccessToken(t, h.tokens, practitionerUID, practitionerMID, h.tenantID, h.branchID, "practitioner")
	h.parentToken = mustAccessToken(t, h.tokens, parentUID, parentMID, h.tenantID, h.branchID, "parent")

	return h
}

func seedDraftInvoice(t *testing.T, h *billingIssueHarness, suffix string, childName string, totalDueMinor int) (childID, invoiceID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	childID = uuid.MustParse(fmt.Sprintf("e5000000-0000-0000-0000-%012s", suffix))
	invoiceID = uuid.MustParse(fmt.Sprintf("e6000000-0000-0000-0000-%012s", suffix))

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, childName,
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, 0, $7, $8, $9)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1),
		totalDueMinor, totalDueMinor, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert draft invoice: %v", err)
	}

	return childID, invoiceID
}

func seedDraftInvoiceForMonth(t *testing.T, h *billingIssueHarness, suffix string, childName string, totalDueMinor int, year int, month time.Month) (childID, invoiceID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	childID = uuid.MustParse(fmt.Sprintf("e5000000-0000-0000-0000-%012s", suffix))
	invoiceID = uuid.MustParse(fmt.Sprintf("e6000000-0000-0000-0000-%012s", suffix))

	billingMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	periodStart := billingMonth
	periodEnd := billingMonth.AddDate(0, 1, -1)

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, childName,
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, 0, $7, $8, $9)`,
		invoiceID, h.tenantID, h.branchID, childID, billingMonth,
		totalDueMinor, totalDueMinor, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("insert draft invoice for month: %v", err)
	}

	return childID, invoiceID
}

// --- Route Registration ---

func TestInvoiceIssueRouteInventory(t *testing.T) {
	h := setupBillingIssueHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"POST /api/v1/invoices/:invoice_id/issue",
		"POST /api/v1/invoices/bulk-issue",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

// --- Authorization ---

func TestInvoiceIssueUnauthenticated(t *testing.T) {
	h := setupBillingIssueHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", "", `{"confirm":true}`)
	requireStatus(t, w, http.StatusUnauthorized)

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", "", `{"billing_month":"2026-05","confirm":true}`)
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestInvoiceIssueRoleGuards(t *testing.T) {
	h := setupBillingIssueHarness(t)

	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", token, `{"confirm":true}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")

		w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", token, `{"billing_month":"2026-05","confirm":true}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

// --- Single Issue Validation ---

func TestInvoiceIssueValidationErrors(t *testing.T) {
	h := setupBillingIssueHarness(t)

	// Bad invoice ID
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/not-a-uuid/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Missing body
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Missing confirm
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", h.managerToken, `{}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// confirm: false
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", h.managerToken, `{"confirm":false}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

// --- Single Issue Success ---

func TestInvoiceIssueSingleSuccess(t *testing.T) {
	h := setupBillingIssueHarness(t)
	ctx := context.Background()

	_, invoiceID := seedDraftInvoice(t, h, "000000000001", "Alice Issue", 1500)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp issueInvoiceResponseTest
	decodeJSON(t, w, &resp)

	if resp.InvoiceID != invoiceID.String() {
		t.Fatalf("invoice_id = %s, want %s", resp.InvoiceID, invoiceID)
	}
	if resp.InvoiceNumber != "INV-202605-0001" {
		t.Fatalf("invoice_number = %s, want INV-202605-0001", resp.InvoiceNumber)
	}
	if resp.Status != "issued" {
		t.Fatalf("status = %s, want issued", resp.Status)
	}
	if resp.IssuedAt == "" {
		t.Fatal("issued_at should not be empty")
	}
	if resp.LockedAt == "" {
		t.Fatal("locked_at should not be empty")
	}
	if resp.DueAt == "" {
		t.Fatal("due_at should not be empty")
	}
	if resp.IssuedRunID == "" {
		t.Fatal("issued_run_id should not be empty")
	}
	if resp.TotalDueMinor != 1500 {
		t.Fatalf("total_due_minor = %d, want 1500", resp.TotalDueMinor)
	}

	// Verify DB state
	var status, invoiceNumber string
	var issuedSequence int
	var issuedAt, lockedAt, dueAt time.Time
	var issuedByUserID, issuedByMembershipID, issuedRunID uuid.UUID
	err := h.pool.QueryRow(ctx,
		`SELECT status, invoice_number, issued_sequence, issued_at, locked_at, due_at, issued_by_user_id, issued_by_membership_id, issued_run_id
		 FROM invoices WHERE id = $1`, invoiceID).
		Scan(&status, &invoiceNumber, &issuedSequence, &issuedAt, &lockedAt, &dueAt, &issuedByUserID, &issuedByMembershipID, &issuedRunID)
	if err != nil {
		t.Fatalf("query invoice: %v", err)
	}
	if status != "issued" {
		t.Fatalf("db status = %s, want issued", status)
	}
	if invoiceNumber != "INV-202605-0001" {
		t.Fatalf("db invoice_number = %s, want INV-202605-0001", invoiceNumber)
	}
	if issuedSequence != 1 {
		t.Fatalf("db issued_sequence = %d, want 1", issuedSequence)
	}
	if !issuedAt.Equal(lockedAt) || !issuedAt.Equal(dueAt) {
		t.Fatal("issued_at, locked_at, due_at should be equal")
	}
	if issuedByUserID != h.managerUID {
		t.Fatalf("issued_by_user_id = %s, want %s", issuedByUserID, h.managerUID)
	}
	if issuedByMembershipID != h.managerMID {
		t.Fatalf("issued_by_membership_id = %s, want %s", issuedByMembershipID, h.managerMID)
	}

	// One issue run
	var runCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_runs WHERE tenant_id = $1 AND run_type = 'issue'", h.tenantID).Scan(&runCount)
	if runCount != 1 {
		t.Fatalf("issue runs = %d, want 1", runCount)
	}

	var runStatus string
	var eligibleCount, successCount, blockedCount int
	h.pool.QueryRow(ctx,
		"SELECT status, eligible_count, success_count, blocked_count FROM invoice_runs WHERE id = $1", issuedRunID).
		Scan(&runStatus, &eligibleCount, &successCount, &blockedCount)
	if runStatus != "completed" {
		t.Fatalf("run status = %s, want completed", runStatus)
	}
	if eligibleCount != 1 || successCount != 1 || blockedCount != 0 {
		t.Fatalf("run counts = %d/%d/%d, want 1/1/0", eligibleCount, successCount, blockedCount)
	}

	// One audit row
	var auditCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action_type = 'invoice_issued'", h.tenantID).Scan(&auditCount)
	if auditCount != 1 {
		t.Fatalf("audit logs = %d, want 1", auditCount)
	}
}

// --- Single Issue Blocked ---

func TestInvoiceIssueNotFound(t *testing.T) {
	h := setupBillingIssueHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+uuid.New().String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

func TestInvoiceIssueNotDraft(t *testing.T) {
	h := setupBillingIssueHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e5000000-0000-0000-0000-000000000010")
	invoiceID := uuid.MustParse("e6000000-0000-0000-0000-000000000010")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Issued Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)

	genRunID := uuid.MustParse("e7000000-0000-0000-0000-000000000010")
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-10')`,
		genRunID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  period_start_date, period_end_date, invoice_number, issued_sequence, generated_run_id, issued_run_id,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at, subtotal_minor, funded_deduction_minor, total_due_minor)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $7, 'INV-OLD', 1, $8, $8, now(), $9, $10, now(), now(), 1000, 0, 1000)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1),
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31),
		genRunID, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert issued invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusConflict)
	requireErrorCode(t, w, "invoice_not_draft")

	// No issue run created
	var issueRunCount int
	h.pool.QueryRow(context.Background(), "SELECT count(*) FROM invoice_runs WHERE tenant_id = $1 AND run_type = 'issue'", h.tenantID).Scan(&issueRunCount)
	if issueRunCount != 0 {
		t.Fatalf("issue runs = %d, want 0 (no run on failed validation)", issueRunCount)
	}
}

// --- Bulk Issue Validation ---

func TestBulkIssueValidationErrors(t *testing.T) {
	h := setupBillingIssueHarness(t)

	// Missing billing_month
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Malformed billing_month
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"invalid","confirm":true}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Malformed invoice ID
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","invoice_ids":["not-a-uuid"],"confirm":true}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Missing confirm
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// confirm: false
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","confirm":false}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

// --- Bulk Issue All Drafts ---

func TestBulkIssueAllDrafts(t *testing.T) {
	h := setupBillingIssueHarness(t)

	seedDraftInvoice(t, h, "000000000100", "Alpha Child", 1000)
	seedDraftInvoice(t, h, "000000000101", "Beta Child", 2000)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Status != "completed" {
		t.Fatalf("status = %s, want completed", resp.Status)
	}
	if resp.Summary.EligibleCount != 2 {
		t.Fatalf("eligible = %d, want 2", resp.Summary.EligibleCount)
	}
	if resp.Summary.SuccessCount != 2 {
		t.Fatalf("success = %d, want 2", resp.Summary.SuccessCount)
	}
	if resp.Summary.BlockedCount != 0 {
		t.Fatalf("blocked = %d, want 0", resp.Summary.BlockedCount)
	}
	if resp.Summary.TotalDueMinor != 3000 {
		t.Fatalf("total_due = %d, want 3000", resp.Summary.TotalDueMinor)
	}

	// Verify sequence order: Alpha first, Beta second
	if len(resp.Issued) != 2 {
		t.Fatalf("issued = %d, want 2", len(resp.Issued))
	}
	if resp.Issued[0].ChildName != "Alpha Child" {
		t.Fatalf("first issued = %s, want Alpha Child", resp.Issued[0].ChildName)
	}
	if resp.Issued[0].InvoiceNumber != "INV-202605-0001" {
		t.Fatalf("first number = %s, want INV-202605-0001", resp.Issued[0].InvoiceNumber)
	}
	if resp.Issued[1].ChildName != "Beta Child" {
		t.Fatalf("second issued = %s, want Beta Child", resp.Issued[1].ChildName)
	}
	if resp.Issued[1].InvoiceNumber != "INV-202605-0002" {
		t.Fatalf("second number = %s, want INV-202605-0002", resp.Issued[1].InvoiceNumber)
	}
}

// --- Bulk Issue Selected Subset ---

func TestBulkIssueSelectedSubset(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, inv1 := seedDraftInvoice(t, h, "000000000200", "Selected One", 1000)
	_, inv2 := seedDraftInvoice(t, h, "000000000201", "Selected Two", 2000)

	body := fmt.Sprintf(`{"billing_month":"2026-05","invoice_ids":["%s"],"confirm":true}`, inv1)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success = %d, want 1", resp.Summary.SuccessCount)
	}
	if resp.Issued[0].InvoiceID != inv1.String() {
		t.Fatalf("issued invoice = %s, want %s", resp.Issued[0].InvoiceID, inv1)
	}

	// inv2 still draft
	var status string
	h.pool.QueryRow(context.Background(), "SELECT status FROM invoices WHERE id = $1", inv2).Scan(&status)
	if status != "draft" {
		t.Fatalf("inv2 status = %s, want draft (not selected)", status)
	}
}

// --- Bulk Issue Deduplication ---

func TestBulkIssueDedupeIDs(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, inv1 := seedDraftInvoice(t, h, "000000000300", "Dedup Child", 1000)

	body := fmt.Sprintf(`{"billing_month":"2026-05","invoice_ids":["%s","%s"],"confirm":true}`, inv1, inv1)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success = %d, want 1 (deduped)", resp.Summary.SuccessCount)
	}
}

// --- Bulk Issue Partial Exceptions ---

func TestBulkIssuePartialExceptions(t *testing.T) {
	h := setupBillingIssueHarness(t)
	ctx := context.Background()

	// Eligible draft
	_, eligibleInv := seedDraftInvoice(t, h, "000000000400", "Eligible Draft", 1000)

	// Already-issued invoice
	childIssued := uuid.MustParse("e5000000-0000-0000-0000-000000000410")
	issuedInv := uuid.MustParse("e6000000-0000-0000-0000-000000000410")
	genRunID := uuid.MustParse("e7000000-0000-0000-0000-000000000410")
	dbtest.InsertChild(t, h.pool, childIssued, h.tenantID, h.branchID, "Already Issued",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-exc')`,
		genRunID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  period_start_date, period_end_date, invoice_number, issued_sequence, generated_run_id, issued_run_id,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at, subtotal_minor, funded_deduction_minor, total_due_minor)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $7, 'INV-OLD', 1, $8, $8, now(), $9, $10, now(), now(), 500, 0, 500)`,
		issuedInv, h.tenantID, h.branchID, childIssued, dbtest.DateAt(2026, 5, 1),
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31),
		genRunID, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert issued: %v", err)
	}

	// Unknown invoice ID
	unknownID := uuid.New()

	body := fmt.Sprintf(`{"billing_month":"2026-05","invoice_ids":["%s","%s","%s"],"confirm":true}`,
		eligibleInv, issuedInv, unknownID)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Status != "completed_with_exceptions" {
		t.Fatalf("status = %s, want completed_with_exceptions", resp.Status)
	}
	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success = %d, want 1", resp.Summary.SuccessCount)
	}
	if resp.Summary.BlockedCount != 2 {
		t.Fatalf("blocked = %d, want 2", resp.Summary.BlockedCount)
	}
	if len(resp.Blocked) != 2 {
		t.Fatalf("blocked entries = %d, want 2", len(resp.Blocked))
	}

	// Check blocker codes
	blockedCodes := make(map[string]string)
	for _, b := range resp.Blocked {
		if len(b.Blockers) > 0 {
			blockedCodes[b.InvoiceID] = b.Blockers[0].Code
		}
	}
	if blockedCodes[issuedInv.String()] != "invoice_not_draft" {
		t.Fatalf("issued blocker = %s, want invoice_not_draft", blockedCodes[issuedInv.String()])
	}
	if blockedCodes[unknownID.String()] != "invoice_not_found" {
		t.Fatalf("unknown blocker = %s, want invoice_not_found", blockedCodes[unknownID.String()])
	}
}

// --- Bulk Issue No-Op ---

func TestBulkIssueNoop(t *testing.T) {
	h := setupBillingIssueHarness(t)

	// Omitted invoice_ids with no drafts
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Status != "completed" {
		t.Fatalf("status = %s, want completed", resp.Status)
	}
	if resp.Summary.EligibleCount != 0 {
		t.Fatalf("eligible = %d, want 0", resp.Summary.EligibleCount)
	}
}

func TestBulkIssueEmptyIDsNoop(t *testing.T) {
	h := setupBillingIssueHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","invoice_ids":[],"confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Summary.EligibleCount != 0 {
		t.Fatalf("eligible = %d, want 0", resp.Summary.EligibleCount)
	}
}

// --- Zero Total Issue ---

func TestInvoiceIssueZeroTotal(t *testing.T) {
	h := setupBillingIssueHarness(t)

	_, invoiceID := seedDraftInvoice(t, h, "000000000500", "Zero Child", 0)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp issueInvoiceResponseTest
	decodeJSON(t, w, &resp)

	if resp.InvoiceNumber == "" {
		t.Fatal("zero-total invoice should receive invoice number")
	}
	if resp.TotalDueMinor != 0 {
		t.Fatalf("total_due_minor = %d, want 0", resp.TotalDueMinor)
	}
}

// --- Regeneration Guard ---

func TestInvoiceIssueRegenerationGuard(t *testing.T) {
	h := setupBillingIssueHarness(t)
	ctx := context.Background()

	childID, invoiceID := seedDraftInvoice(t, h, "000000000600", "Guard Child", 1000)

	// Issue the draft
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+invoiceID.String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	// Add guardian and funding so child is eligible for generation
	guardianID := uuid.MustParse("e8000000-0000-0000-0000-000000000600")
	linkID := uuid.MustParse("e9000000-0000-0000-0000-000000000600")
	fundingID := uuid.MustParse("ea000000-0000-0000-0000-000000000600")
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Guardian Guard", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	// Try draft generation for same month
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05","child_ids":["`+childID.String()+`"]}`)
	requireStatus(t, w, http.StatusOK)

	var genResp genDraftsResponse
	decodeJSON(t, w, &genResp)

	if genResp.Summary.SuccessCount != 0 {
		t.Fatalf("success_count = %d, want 0 (issued should block)", genResp.Summary.SuccessCount)
	}
	if genResp.Summary.BlockedCount != 1 {
		t.Fatalf("blocked_count = %d, want 1", genResp.Summary.BlockedCount)
	}
	if genResp.Blocked[0].Blockers[0].Code != "invoice_already_issued" {
		t.Fatalf("blocker = %s, want invoice_already_issued", genResp.Blocked[0].Blockers[0].Code)
	}

	// Issued invoice not mutated
	var status string
	h.pool.QueryRow(ctx, "SELECT status FROM invoices WHERE id = $1", invoiceID).Scan(&status)
	if status != "issued" {
		t.Fatalf("invoice status = %s, want issued (unchanged)", status)
	}
}

func TestInvoiceNumberingContinuesAcrossSingleAndBulkAndResetsByMonth(t *testing.T) {
	h := setupBillingIssueHarness(t)
	ctx := context.Background()

	betaChildID, betaInvID := seedDraftInvoice(t, h, "000000000700", "Beta Numbering", 1000)
	_, alphaInvID := seedDraftInvoice(t, h, "000000000701", "Alpha Numbering", 2000)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/"+betaInvID.String()+"/issue", h.managerToken, `{"confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var singleResp issueInvoiceResponseTest
	decodeJSON(t, w, &singleResp)
	if singleResp.InvoiceNumber != "INV-202605-0001" {
		t.Fatalf("single issue number = %s, want INV-202605-0001", singleResp.InvoiceNumber)
	}

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var bulkResp bulkIssueResponseTest
	decodeJSON(t, w, &bulkResp)

	if bulkResp.Summary.SuccessCount != 1 {
		t.Fatalf("bulk success = %d, want 1 (only Alpha remaining)", bulkResp.Summary.SuccessCount)
	}
	if bulkResp.Issued[0].InvoiceID != alphaInvID.String() {
		t.Fatalf("bulk issued = %s, want %s", bulkResp.Issued[0].InvoiceID, alphaInvID)
	}
	if bulkResp.Issued[0].InvoiceNumber != "INV-202605-0002" {
		t.Fatalf("alpha number = %s, want INV-202605-0002", bulkResp.Issued[0].InvoiceNumber)
	}

	_, juneInvID := seedDraftInvoiceForMonth(t, h, "000000000710", "June Child", 1000, 2026, 6)

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-06","confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var juneResp bulkIssueResponseTest
	decodeJSON(t, w, &juneResp)

	if juneResp.Issued[0].InvoiceNumber != "INV-202606-0001" {
		t.Fatalf("june number = %s, want INV-202606-0001", juneResp.Issued[0].InvoiceNumber)
	}

	var betaNumber, alphaNumber, juneNumber string
	h.pool.QueryRow(ctx, "SELECT invoice_number FROM invoices WHERE id = $1", betaInvID).Scan(&betaNumber)
	h.pool.QueryRow(ctx, "SELECT invoice_number FROM invoices WHERE id = $1", alphaInvID).Scan(&alphaNumber)
	h.pool.QueryRow(ctx, "SELECT invoice_number FROM invoices WHERE id = $1", juneInvID).Scan(&juneNumber)
	if betaNumber != "INV-202605-0001" {
		t.Fatalf("beta db number = %s, want INV-202605-0001", betaNumber)
	}
	if alphaNumber != "INV-202605-0002" {
		t.Fatalf("alpha db number = %s, want INV-202605-0002", alphaNumber)
	}
	if juneNumber != "INV-202606-0001" {
		t.Fatalf("june db number = %s, want INV-202606-0001", juneNumber)
	}

	_ = betaChildID
}

func TestBulkIssueInvoiceNumberingTieBreaksSameNameByInvoiceID(t *testing.T) {
	h := setupBillingIssueHarness(t)

	lowID := uuid.MustParse("e6000000-0000-0000-0000-000000000800")
	highID := uuid.MustParse("e6000000-0000-0000-0000-000000000801")
	lowChild := uuid.MustParse("e5000000-0000-0000-0000-000000000800")
	highChild := uuid.MustParse("e5000000-0000-0000-0000-000000000801")

	for i, pair := range []struct {
		childID, invoiceID uuid.UUID
	}{
		{lowChild, lowID},
		{highChild, highID},
	} {
		dbtest.InsertChild(t, h.pool, pair.childID, h.tenantID, h.branchID, "Same Name Child",
			dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
		_, err := h.pool.Exec(context.Background(),
			`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
			 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 1000, 0, 1000, $6, $7)`,
			pair.invoiceID, h.tenantID, h.branchID, pair.childID, dbtest.DateAt(2026, 5, 1),
			dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
		if err != nil {
			t.Fatalf("insert draft %d: %v", i, err)
		}
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoices/bulk-issue", h.managerToken, `{"billing_month":"2026-05","confirm":true}`)
	requireStatus(t, w, http.StatusOK)

	var resp bulkIssueResponseTest
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 2 {
		t.Fatalf("success = %d, want 2", resp.Summary.SuccessCount)
	}

	lowNum, highNum := "", ""
	for _, issued := range resp.Issued {
		if issued.InvoiceID == lowID.String() {
			lowNum = issued.InvoiceNumber
		}
		if issued.InvoiceID == highID.String() {
			highNum = issued.InvoiceNumber
		}
	}
	if lowNum != "INV-202605-0001" {
		t.Fatalf("lower ID number = %s, want INV-202605-0001", lowNum)
	}
	if highNum != "INV-202605-0002" {
		t.Fatalf("higher ID number = %s, want INV-202605-0002", highNum)
	}
}

// --- Test response types for API-19 ---

type issueInvoiceResponseTest struct {
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
	Status        string `json:"status"`
	IssuedAt      string `json:"issued_at"`
	LockedAt      string `json:"locked_at"`
	DueAt         string `json:"due_at"`
	IssuedRunID   string `json:"issued_run_id"`
	TotalDueMinor int    `json:"total_due_minor"`
}

type bulkIssueResponseTest struct {
	RunID        string                      `json:"run_id"`
	BillingMonth string                      `json:"billing_month"`
	Status       string                      `json:"status"`
	Summary      bulkIssueSummaryTest        `json:"summary"`
	Issued       []issuedInvoiceResponseTest `json:"issued"`
	Blocked      []blockedInvoiceTest        `json:"blocked"`
}

type bulkIssueSummaryTest struct {
	EligibleCount int `json:"eligible_count"`
	SuccessCount  int `json:"success_count"`
	BlockedCount  int `json:"blocked_count"`
	TotalDueMinor int `json:"total_due_minor"`
}

type issuedInvoiceResponseTest struct {
	InvoiceID     string `json:"invoice_id"`
	ChildID       string `json:"child_id"`
	ChildName     string `json:"child_name"`
	InvoiceNumber string `json:"invoice_number"`
	IssuedAt      string `json:"issued_at"`
	DueAt         string `json:"due_at"`
	TotalDueMinor int    `json:"total_due_minor"`
}

type blockedInvoiceTest struct {
	InvoiceID string             `json:"invoice_id"`
	ChildID   *string            `json:"child_id,omitempty"`
	ChildName string             `json:"child_name,omitempty"`
	Blockers  []issueBlockerTest `json:"blockers"`
}

type issueBlockerTest struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
