package bootstrap

import (
	"context"
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

type billingHarness struct {
	router            *gin.Engine
	pool              *pgxpool.Pool
	tokens            *authtokens.TokenManager
	tenantID          uuid.UUID
	branchID          uuid.UUID
	managerToken      string
	practitionerToken string
	parentToken       string
}

func setupBillingHarness(t *testing.T) *billingHarness {
	t.Helper()

	gin.SetMode(gin.TestMode)

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &billingHarness{
		router:   Bootstrap(testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), pool),
		pool:     pool,
		tokens:   authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720),
		tenantID: uuid.MustParse("b1000000-0000-0000-0000-000000000001"),
		branchID: uuid.MustParse("b2000000-0000-0000-0000-000000000001"),
	}

	managerUID := uuid.MustParse("b3000000-0000-0000-0000-000000000001")
	managerMID := uuid.MustParse("b4000000-0000-0000-0000-000000000001")
	practitionerUID := uuid.MustParse("b3000000-0000-0000-0000-000000000002")
	practitionerMID := uuid.MustParse("b4000000-0000-0000-0000-000000000002")
	parentUID := uuid.MustParse("b3000000-0000-0000-0000-000000000003")
	parentMID := uuid.MustParse("b4000000-0000-0000-0000-000000000003")

	dbtest.InsertTenant(t, pool, h.tenantID, "Billing Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Billing Branch")
	dbtest.InsertUser(t, pool, managerUID, "billing-mgr@example.com", "hash", true)
	dbtest.InsertUser(t, pool, practitionerUID, "billing-prac@example.com", "hash", true)
	dbtest.InsertUser(t, pool, parentUID, "billing-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, managerMID, h.tenantID, h.branchID, managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, practitionerMID, h.tenantID, h.branchID, practitionerUID, "practitioner", true)
	dbtest.InsertMembership(t, pool, parentMID, h.tenantID, h.branchID, parentUID, "parent", true)

	h.managerToken = mustAccessToken(t, h.tokens, managerUID, managerMID, h.tenantID, h.branchID, "manager")
	h.practitionerToken = mustAccessToken(t, h.tokens, practitionerUID, practitionerMID, h.tenantID, h.branchID, "practitioner")
	h.parentToken = mustAccessToken(t, h.tokens, parentUID, parentMID, h.tenantID, h.branchID, "parent")

	return h
}

func TestBillingRouteInventory(t *testing.T) {
	h := setupBillingHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/invoices/drafts/preflight",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestBillingPreflightUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestBillingPreflightRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)

	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestBillingPreflightValidationErrors(t *testing.T) {
	h := setupBillingHarness(t)

	// Missing billing_month
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Malformed billing_month
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=invalid", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Another malformed format
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-13", h.managerToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestBillingPreflightEmpty(t *testing.T) {
	h := setupBillingHarness(t)

	// No children at all
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp preflightResponse
	decodeJSON(t, w, &resp)
	if resp.Summary.TotalChildrenCount != 0 {
		t.Fatalf("expected 0 children, got %d", resp.Summary.TotalChildrenCount)
	}
}

func TestBillingPreflightMixed(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Child 1: eligible with complete attendance and funding
	child1 := uuid.MustParse("b5000000-0000-0000-0000-000000000001")
	guardian1 := uuid.MustParse("b6000000-0000-0000-0000-000000000001")
	link1 := uuid.MustParse("b7000000-0000-0000-0000-000000000001")
	fundingProfile1 := uuid.MustParse("b8000000-0000-0000-0000-000000000001")
	session1 := uuid.MustParse("b9000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Alice Eligible",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian One", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	// Funding profile with 300 minutes
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding profile 1: %v", err)
	}
	// Complete attendance session: 8:00 to 16:00 = 480 min raw, 480 rounded
	checkIn1 := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	checkOut1 := time.Date(2026, 5, 15, 16, 0, 0, 0, time.UTC)
	localDate1 := dbtest.DateAt(2026, 5, 15)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session1, h.tenantID, h.branchID, child1, checkIn1, checkOut1, localDate1, localDate1)
	if err != nil {
		t.Fatalf("insert session 1: %v", err)
	}

	// Child 2: eligible with zero attendance and zero funding
	child2 := uuid.MustParse("b5000000-0000-0000-0000-000000000002")
	guardian2 := uuid.MustParse("b6000000-0000-0000-0000-000000000002")
	link2 := uuid.MustParse("b7000000-0000-0000-0000-000000000002")
	fundingProfile2 := uuid.MustParse("b8000000-0000-0000-0000-000000000002")

	dbtest.InsertChild(t, h.pool, child2, h.tenantID, h.branchID, "Bob Zero Attendance",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 400, true)
	dbtest.InsertGuardian(t, h.pool, guardian2, h.tenantID, h.branchID, "Guardian Two", true)
	dbtest.InsertGuardianLink(t, h.pool, link2, h.tenantID, h.branchID, guardian2, child2)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile2, h.tenantID, h.branchID, child2, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding profile 2: %v", err)
	}

	// Child 3: blocked - missing funding profile
	child3 := uuid.MustParse("b5000000-0000-0000-0000-000000000003")
	guardian3 := uuid.MustParse("b6000000-0000-0000-0000-000000000003")
	link3 := uuid.MustParse("b7000000-0000-0000-0000-000000000003")

	dbtest.InsertChild(t, h.pool, child3, h.tenantID, h.branchID, "Charlie No Funding",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardian3, h.tenantID, h.branchID, "Guardian Three", true)
	dbtest.InsertGuardianLink(t, h.pool, link3, h.tenantID, h.branchID, guardian3, child3)

	// Child 4: blocked - missing guardian link
	child4 := uuid.MustParse("b5000000-0000-0000-0000-000000000004")
	fundingProfile4 := uuid.MustParse("b8000000-0000-0000-0000-000000000004")

	dbtest.InsertChild(t, h.pool, child4, h.tenantID, h.branchID, "Diana No Guardian",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile4, h.tenantID, h.branchID, child4, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding profile 4: %v", err)
	}

	// Child 5: blocked - open attendance session (no checkout)
	child5 := uuid.MustParse("b5000000-0000-0000-0000-000000000005")
	guardian5 := uuid.MustParse("b6000000-0000-0000-0000-000000000005")
	link5 := uuid.MustParse("b7000000-0000-0000-0000-000000000005")
	fundingProfile5 := uuid.MustParse("b8000000-0000-0000-0000-000000000005")
	session5 := uuid.MustParse("b9000000-0000-0000-0000-000000000005")

	dbtest.InsertChild(t, h.pool, child5, h.tenantID, h.branchID, "Eve Open Session",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardian5, h.tenantID, h.branchID, "Guardian Five", true)
	dbtest.InsertGuardianLink(t, h.pool, link5, h.tenantID, h.branchID, guardian5, child5)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile5, h.tenantID, h.branchID, child5, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding profile 5: %v", err)
	}
	// Open session (no checkout)
	dbtest.InsertAttendanceSession(t, h.pool, session5, h.tenantID, h.branchID, child5, "open",
		time.Date(2026, 5, 3, 8, 30, 0, 0, time.UTC),
		dbtest.DateAt(2026, 5, 3))

	// Child 6: eligible with existing draft invoice
	child6 := uuid.MustParse("b5000000-0000-0000-0000-000000000006")
	guardian6 := uuid.MustParse("b6000000-0000-0000-0000-000000000006")
	link6 := uuid.MustParse("b7000000-0000-0000-0000-000000000006")
	fundingProfile6 := uuid.MustParse("b8000000-0000-0000-0000-000000000006")
	invoice6 := uuid.MustParse("ba000000-0000-0000-0000-000000000006")

	dbtest.InsertChild(t, h.pool, child6, h.tenantID, h.branchID, "Frank Draft Invoice",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardian6, h.tenantID, h.branchID, "Guardian Six", true)
	dbtest.InsertGuardianLink(t, h.pool, link6, h.tenantID, h.branchID, guardian6, child6)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile6, h.tenantID, h.branchID, child6, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding profile 6: %v", err)
	}
	// Existing draft invoice
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, $7)`,
		invoice6, h.tenantID, h.branchID, child6, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert draft invoice: %v", err)
	}

	// Child 7: blocked with existing issued invoice
	child7 := uuid.MustParse("b5000000-0000-0000-0000-000000000007")
	guardian7 := uuid.MustParse("b6000000-0000-0000-0000-000000000007")
	link7 := uuid.MustParse("b7000000-0000-0000-0000-000000000007")
	fundingProfile7 := uuid.MustParse("b8000000-0000-0000-0000-000000000007")
	invoice7 := uuid.MustParse("ba000000-0000-0000-0000-000000000007")

	dbtest.InsertChild(t, h.pool, child7, h.tenantID, h.branchID, "Grace Issued Invoice",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardian7, h.tenantID, h.branchID, "Guardian Seven", true)
	dbtest.InsertGuardianLink(t, h.pool, link7, h.tenantID, h.branchID, guardian7, child7)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingProfile7, h.tenantID, h.branchID, child7, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding profile 7: %v", err)
	}
	// Issued invoice (needs all required fields per schema constraint)
	issuedByUID := uuid.MustParse("b3000000-0000-0000-0000-000000000001")
	issuedByMID := uuid.MustParse("b4000000-0000-0000-0000-000000000001")
	issuedRunID := uuid.MustParse("bb000000-0000-0000-0000-000000000001")
	// Need an invoice_run for the issued invoice FK
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, $7)`,
		issuedRunID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), issuedByUID, issuedByMID, "req-1")
	if err != nil {
		t.Fatalf("insert invoice run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  period_start_date, period_end_date, invoice_number, issued_sequence, generated_run_id, issued_run_id,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $7, 'INV-001', 1, $8, $8, now(), $9, $10, now(), now() + interval '7 days')`,
		invoice7, h.tenantID, h.branchID, child7, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31),
		issuedRunID, issuedByUID, issuedByMID)
	if err != nil {
		t.Fatalf("insert issued invoice: %v", err)
	}

	// Execute preflight
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp preflightResponse
	decodeJSON(t, w, &resp)

	// Verify basic structure
	if resp.BillingMonth != "2026-05" {
		t.Fatalf("billing_month = %s, want 2026-05", resp.BillingMonth)
	}
	if resp.CurrencyCode != "GBP" {
		t.Fatalf("currency_code = %s, want GBP", resp.CurrencyCode)
	}
	if resp.Period.StartDate != "2026-05-01" {
		t.Fatalf("period.start_date = %s, want 2026-05-01", resp.Period.StartDate)
	}
	if resp.Period.EndDate != "2026-05-31" {
		t.Fatalf("period.end_date = %s, want 2026-05-31", resp.Period.EndDate)
	}
	if resp.Period.EndExclusiveDate != "2026-06-01" {
		t.Fatalf("period.end_exclusive_date = %s, want 2026-06-01", resp.Period.EndExclusiveDate)
	}

	// Counts
	if resp.Summary.TotalChildrenCount != 7 {
		t.Fatalf("total = %d, want 7", resp.Summary.TotalChildrenCount)
	}
	if resp.Summary.EligibleChildrenCount != 3 {
		t.Fatalf("eligible = %d, want 3 (Alice, Bob, Frank)", resp.Summary.EligibleChildrenCount)
	}
	if resp.Summary.BlockedChildrenCount != 4 {
		t.Fatalf("blocked = %d, want 4 (Charlie, Diana, Eve, Grace)", resp.Summary.BlockedChildrenCount)
	}

	// Alice: 480 min raw, 480 rounded, rate 500, funded 300
	// subtotal = ceil(480 * 500 / 60) = ceil(4000) = 4000
	// funded_deduction = ceil(300 * 500 / 60) = ceil(2500) = 2500
	// total_due = max(0, 4000 - 2500) = 1500
	if resp.Summary.RawAttendedMinutes != 480 {
		t.Fatalf("summary raw_attended = %d, want 480", resp.Summary.RawAttendedMinutes)
	}
	if resp.Summary.RoundedAttendedMinutes != 480 {
		t.Fatalf("summary rounded_attended = %d, want 480", resp.Summary.RoundedAttendedMinutes)
	}
	if resp.Summary.IncludedSessionCount != 1 {
		t.Fatalf("summary included_sessions = %d, want 1", resp.Summary.IncludedSessionCount)
	}
	if resp.Summary.SubtotalMinor != 4000 {
		t.Fatalf("summary subtotal = %d, want 4000", resp.Summary.SubtotalMinor)
	}
	if resp.Summary.FundedDeductionMinor != 2500 {
		t.Fatalf("summary funded_deduction = %d, want 2500", resp.Summary.FundedDeductionMinor)
	}
	if resp.Summary.TotalDueMinor != 1500 {
		t.Fatalf("summary total_due = %d, want 1500", resp.Summary.TotalDueMinor)
	}

	// Verify eligible children sorted by name
	if len(resp.EligibleChildren) != 3 {
		t.Fatalf("eligible children = %d, want 3", len(resp.EligibleChildren))
	}
	if resp.EligibleChildren[0].ChildName != "Alice Eligible" {
		t.Fatalf("first eligible = %s, want Alice Eligible", resp.EligibleChildren[0].ChildName)
	}
	if resp.EligibleChildren[0].RawAttendedMinutes != 480 {
		t.Fatalf("alice raw = %d, want 480", resp.EligibleChildren[0].RawAttendedMinutes)
	}

	// Bob: zero attendance
	if resp.EligibleChildren[1].ChildName != "Bob Zero Attendance" {
		t.Fatalf("second eligible = %s, want Bob Zero Attendance", resp.EligibleChildren[1].ChildName)
	}
	if resp.EligibleChildren[1].RawAttendedMinutes != 0 {
		t.Fatalf("bob raw = %d, want 0", resp.EligibleChildren[1].RawAttendedMinutes)
	}
	if resp.EligibleChildren[1].SubtotalMinor != 0 {
		t.Fatalf("bob subtotal = %d, want 0", resp.EligibleChildren[1].SubtotalMinor)
	}
	if resp.EligibleChildren[1].TotalDueMinor != 0 {
		t.Fatalf("bob total_due = %d, want 0", resp.EligibleChildren[1].TotalDueMinor)
	}

	// Frank: has existing draft invoice
	if resp.EligibleChildren[2].ChildName != "Frank Draft Invoice" {
		t.Fatalf("third eligible = %s, want Frank Draft Invoice", resp.EligibleChildren[2].ChildName)
	}
	if resp.EligibleChildren[2].ExistingInvoice == nil {
		t.Fatal("frank should have existing_invoice")
	}
	if resp.EligibleChildren[2].ExistingInvoice.Status != "draft" {
		t.Fatalf("frank invoice status = %s, want draft", resp.EligibleChildren[2].ExistingInvoice.Status)
	}

	// Blocked children sorted by name
	if len(resp.BlockedChildren) != 4 {
		t.Fatalf("blocked children = %d, want 4", len(resp.BlockedChildren))
	}
	if resp.BlockedChildren[0].ChildName != "Charlie No Funding" {
		t.Fatalf("first blocked = %s, want Charlie No Funding", resp.BlockedChildren[0].ChildName)
	}
	if resp.BlockedChildren[0].Blockers[0].Code != "missing_funding_profile" {
		t.Fatalf("charlie blocker = %s, want missing_funding_profile", resp.BlockedChildren[0].Blockers[0].Code)
	}

	if resp.BlockedChildren[1].ChildName != "Diana No Guardian" {
		t.Fatalf("second blocked = %s, want Diana No Guardian", resp.BlockedChildren[1].ChildName)
	}
	if resp.BlockedChildren[1].Blockers[0].Code != "missing_guardian_link" {
		t.Fatalf("diana blocker = %s, want missing_guardian_link", resp.BlockedChildren[1].Blockers[0].Code)
	}

	if resp.BlockedChildren[2].ChildName != "Eve Open Session" {
		t.Fatalf("third blocked = %s, want Eve Open Session", resp.BlockedChildren[2].ChildName)
	}
	if resp.BlockedChildren[2].Blockers[0].Code != "incomplete_attendance" {
		t.Fatalf("eve blocker = %s, want incomplete_attendance", resp.BlockedChildren[2].Blockers[0].Code)
	}
	if resp.BlockedChildren[2].Blockers[0].SessionID == nil {
		t.Fatal("eve blocker should have session_id")
	}

	if resp.BlockedChildren[3].ChildName != "Grace Issued Invoice" {
		t.Fatalf("fourth blocked = %s, want Grace Issued Invoice", resp.BlockedChildren[3].ChildName)
	}
	if resp.BlockedChildren[3].Blockers[0].Code != "invoice_already_issued" {
		t.Fatalf("grace blocker = %s, want invoice_already_issued", resp.BlockedChildren[3].Blockers[0].Code)
	}
	if resp.BlockedChildren[3].Blockers[0].InvoiceID == nil {
		t.Fatal("grace blocker should have invoice_id")
	}

	// Blocker counts
	blockerCountMap := make(map[string]int)
	for _, bc := range resp.Summary.BlockerCounts {
		blockerCountMap[bc.Code] = bc.ChildrenCount
	}
	if blockerCountMap["missing_funding_profile"] != 1 {
		t.Fatalf("missing_funding_profile count = %d, want 1", blockerCountMap["missing_funding_profile"])
	}
	if blockerCountMap["missing_guardian_link"] != 1 {
		t.Fatalf("missing_guardian_link count = %d, want 1", blockerCountMap["missing_guardian_link"])
	}
	if blockerCountMap["incomplete_attendance"] != 1 {
		t.Fatalf("incomplete_attendance count = %d, want 1", blockerCountMap["incomplete_attendance"])
	}
	if blockerCountMap["invoice_already_issued"] != 1 {
		t.Fatalf("invoice_already_issued count = %d, want 1", blockerCountMap["invoice_already_issued"])
	}

	// Verify no invoice_runs, invoices, or invoice_lines were created
	var runCount, invoiceCount, lineCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_runs WHERE tenant_id = $1", h.tenantID).Scan(&runCount)
	// The issued invoice test creates 1 run, so we expect 1
	if runCount != 1 {
		t.Fatalf("invoice_runs = %d, want 1 (only the test data)", runCount)
	}
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1", h.tenantID).Scan(&invoiceCount)
	// draft + issued = 2
	if invoiceCount != 2 {
		t.Fatalf("invoices = %d, want 2 (only test data)", invoiceCount)
	}
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_lines WHERE tenant_id = $1", h.tenantID).Scan(&lineCount)
	if lineCount != 0 {
		t.Fatalf("invoice_lines = %d, want 0 (preflight is read-only)", lineCount)
	}
}

func TestBillingPreflightCrossMonthAllocation(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Child enrolled for May 2026
	childID := uuid.MustParse("b5000000-0000-0000-0000-000000000010")
	guardianID := uuid.MustParse("b6000000-0000-0000-0000-000000000010")
	linkID := uuid.MustParse("b7000000-0000-0000-0000-000000000010")
	fundingID := uuid.MustParse("b8000000-0000-0000-0000-000000000010")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Cross Month Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Guardian Cross", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	// Session checking in on May 31 (London) and checking out June 1 - allocated to May
	london, _ := time.LoadLocation("Europe/London")
	checkInAt := time.Date(2026, 5, 31, 20, 0, 0, 0, london).UTC()
	checkOutAt := time.Date(2026, 6, 1, 8, 0, 0, 0, london).UTC()
	sessionID := uuid.MustParse("b9000000-0000-0000-0000-000000000010")
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		sessionID, h.tenantID, h.branchID, childID, checkInAt, checkOutAt, dbtest.DateAt(2026, 5, 31), dbtest.DateAt(2026, 6, 1))
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	// April funding profile so April preflight includes the child
	fundingID2 := uuid.MustParse("b8000000-0000-0000-0000-000000000011")
	_, err = h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingID2, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 4, 1), 0)
	if err != nil {
		t.Fatalf("insert april funding: %v", err)
	}

	// Session checking in on Apr 30 (London) and checking out May 1 - allocated to April, NOT May
	sessionID2 := uuid.MustParse("b9000000-0000-0000-0000-000000000011")
	checkInAt2 := time.Date(2026, 4, 30, 20, 0, 0, 0, london).UTC()
	checkOutAt2 := time.Date(2026, 5, 1, 8, 0, 0, 0, london).UTC()
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		sessionID2, h.tenantID, h.branchID, childID, checkInAt2, checkOutAt2, dbtest.DateAt(2026, 4, 30), dbtest.DateAt(2026, 5, 1))
	if err != nil {
		t.Fatalf("insert session 2: %v", err)
	}

	// May preflight should include only the May 31 session
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp preflightResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.TotalChildrenCount != 1 {
		t.Fatalf("total = %d, want 1", resp.Summary.TotalChildrenCount)
	}
	if resp.Summary.EligibleChildrenCount != 1 {
		t.Fatalf("eligible = %d, want 1", resp.Summary.EligibleChildrenCount)
	}
	if resp.Summary.IncludedSessionCount != 1 {
		t.Fatalf("included sessions = %d, want 1 (only the May 31 session)", resp.Summary.IncludedSessionCount)
	}

	// April preflight should include the Apr 30 session
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-04", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)

	if resp.Summary.IncludedSessionCount != 1 {
		t.Fatalf("april included sessions = %d, want 1 (only the Apr 30 session)", resp.Summary.IncludedSessionCount)
	}
}

// JSON response types for test deserialization
type preflightResponse struct {
	BillingMonth     string                   `json:"billing_month"`
	CurrencyCode     string                   `json:"currency_code"`
	Period           periodResponse           `json:"period"`
	Summary          summaryResponse          `json:"summary"`
	EligibleChildren []eligibleChildResponse  `json:"eligible_children"`
	BlockedChildren  []blockedChildResponse   `json:"blocked_children"`
}

type periodResponse struct {
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	EndExclusiveDate string `json:"end_exclusive_date"`
}

type summaryResponse struct {
	TotalChildrenCount     int                    `json:"total_children_count"`
	EligibleChildrenCount  int                    `json:"eligible_children_count"`
	BlockedChildrenCount   int                    `json:"blocked_children_count"`
	IncludedSessionCount   int                    `json:"included_session_count"`
	RawAttendedMinutes     int                    `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                    `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int                    `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                    `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                    `json:"core_billable_minutes"`
	SubtotalMinor          int                    `json:"subtotal_minor"`
	FundedDeductionMinor   int                    `json:"funded_deduction_minor"`
	TotalDueMinor          int                    `json:"total_due_minor"`
	BlockerCounts          []blockerCountResponse `json:"blocker_counts"`
}

type eligibleChildResponse struct {
	ChildID                string              `json:"child_id"`
	ChildName              string              `json:"child_name"`
	CoreHourlyRateMinor    int                 `json:"core_hourly_rate_minor"`
	FundingProfileID       *string             `json:"funding_profile_id"`
	FundedAllowanceMinutes int                 `json:"funded_allowance_minutes"`
	RawAttendedMinutes     int                 `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                 `json:"rounded_attended_minutes"`
	IncludedSessionCount   int                 `json:"included_session_count"`
	FundedDeductionMinutes int                 `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                 `json:"core_billable_minutes"`
	SubtotalMinor          int                 `json:"subtotal_minor"`
	FundedDeductionMinor   int                 `json:"funded_deduction_minor"`
	TotalDueMinor          int                 `json:"total_due_minor"`
	ExistingInvoice        *existingInvoiceRef `json:"existing_invoice"`
}

type existingInvoiceRef struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type blockedChildResponse struct {
	ChildID   string            `json:"child_id"`
	ChildName string            `json:"child_name"`
	Blockers  []blockerResponse `json:"blockers"`
}

type blockerResponse struct {
	Code             string   `json:"code"`
	Message          string   `json:"message"`
	SessionID        *string  `json:"session_id,omitempty"`
	CheckInAt        *string  `json:"check_in_at,omitempty"`
	CheckInLocalDate *string  `json:"check_in_local_date,omitempty"`
	InvoiceID        *string  `json:"invoice_id,omitempty"`
	InvoiceStatus    *string  `json:"invoice_status,omitempty"`
	Field            *string  `json:"field,omitempty"`
}

type blockerCountResponse struct {
	Code          string `json:"code"`
	ChildrenCount int    `json:"children_count"`
}
