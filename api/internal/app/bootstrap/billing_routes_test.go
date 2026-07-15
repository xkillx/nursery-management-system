package bootstrap

import (
	"context"
	"encoding/json"
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

type billingHarness struct {
	router            *gin.Engine
	pool              *pgxpool.Pool
	tokens            *authtokens.TokenManager
	tenantID          uuid.UUID
	branchID          uuid.UUID
	managerUID        uuid.UUID
	managerMID        uuid.UUID
	practitionerUID   uuid.UUID
	practitionerMID   uuid.UUID
	parentUID         uuid.UUID
	parentMID         uuid.UUID
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
		router:          Bootstrap(testConfig(), slog.New(slog.NewTextHandler(io.Discard, nil)), pool),
		pool:            pool,
		tokens:          authtokens.NewTokenManager("access-secret", "refresh-secret", 15, 720, 24),
		tenantID:        uuid.MustParse("b1000000-0000-0000-0000-000000000001"),
		branchID:        uuid.MustParse("b2000000-0000-0000-0000-000000000001"),
		managerUID:      uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
		managerMID:      uuid.MustParse("b4000000-0000-0000-0000-000000000001"),
		practitionerUID: uuid.MustParse("b3000000-0000-0000-0000-000000000002"),
		practitionerMID: uuid.MustParse("b4000000-0000-0000-0000-000000000002"),
		parentUID:       uuid.MustParse("b3000000-0000-0000-0000-000000000003"),
		parentMID:       uuid.MustParse("b4000000-0000-0000-0000-000000000003"),
	}

	dbtest.InsertTenant(t, pool, h.tenantID, "Billing Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Billing Branch")
	dbtest.InsertUser(t, pool, h.managerUID, "billing-mgr@example.com", "hash", true)
	dbtest.InsertUser(t, pool, h.practitionerUID, "billing-prac@example.com", "hash", true)
	dbtest.InsertUser(t, pool, h.parentUID, "billing-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertMembership(t, pool, h.practitionerMID, h.tenantID, h.branchID, h.practitionerUID, "practitioner", true)
	dbtest.InsertMembership(t, pool, h.parentMID, h.tenantID, h.branchID, h.parentUID, "parent", true)

	h.managerToken = mustAccessToken(t, h.tokens, h.managerUID, h.managerMID, h.tenantID, h.branchID, "manager")
	h.practitionerToken = mustAccessToken(t, h.tokens, h.practitionerUID, h.practitionerMID, h.tenantID, h.branchID, "practitioner")
	h.parentToken = mustAccessToken(t, h.tokens, h.parentUID, h.parentMID, h.tenantID, h.branchID, "parent")

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
		"POST /api/v1/invoice-runs/drafts",
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
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")

	// Another malformed format
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-13", h.managerToken, "")
	requireStatus(t, w, http.StatusUnprocessableEntity)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian3, h.tenantID, h.branchID, "Guardian Three", true)
	dbtest.InsertGuardianLink(t, h.pool, link3, h.tenantID, h.branchID, guardian3, child3)

	// Child 4: blocked - missing guardian link
	child4 := uuid.MustParse("b5000000-0000-0000-0000-000000000004")
	fundingProfile4 := uuid.MustParse("b8000000-0000-0000-0000-000000000004")

	dbtest.InsertChild(t, h.pool, child4, h.tenantID, h.branchID, "Diana No Guardian",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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
	if resp.EligibleChildren[0].ChildFirstName != "Alice Eligible" {
		t.Fatalf("first eligible = %s, want Alice Eligible", resp.EligibleChildren[0].ChildFirstName)
	}
	if resp.EligibleChildren[0].RawAttendedMinutes != 480 {
		t.Fatalf("alice raw = %d, want 480", resp.EligibleChildren[0].RawAttendedMinutes)
	}

	// Bob: zero attendance
	if resp.EligibleChildren[1].ChildFirstName != "Bob Zero Attendance" {
		t.Fatalf("second eligible = %s, want Bob Zero Attendance", resp.EligibleChildren[1].ChildFirstName)
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
	if resp.EligibleChildren[2].ChildFirstName != "Frank Draft Invoice" {
		t.Fatalf("third eligible = %s, want Frank Draft Invoice", resp.EligibleChildren[2].ChildFirstName)
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
	if resp.BlockedChildren[0].ChildFirstName != "Charlie No Funding" {
		t.Fatalf("first blocked = %s, want Charlie No Funding", resp.BlockedChildren[0].ChildFirstName)
	}
	if resp.BlockedChildren[0].Blockers[0].Code != "missing_funding_profile" {
		t.Fatalf("charlie blocker = %s, want missing_funding_profile", resp.BlockedChildren[0].Blockers[0].Code)
	}

	if resp.BlockedChildren[1].ChildFirstName != "Diana No Guardian" {
		t.Fatalf("second blocked = %s, want Diana No Guardian", resp.BlockedChildren[1].ChildFirstName)
	}
	if resp.BlockedChildren[1].Blockers[0].Code != "missing_guardian_link" {
		t.Fatalf("diana blocker = %s, want missing_guardian_link", resp.BlockedChildren[1].Blockers[0].Code)
	}

	if resp.BlockedChildren[2].ChildFirstName != "Eve Open Session" {
		t.Fatalf("third blocked = %s, want Eve Open Session", resp.BlockedChildren[2].ChildFirstName)
	}
	if resp.BlockedChildren[2].Blockers[0].Code != "incomplete_attendance" {
		t.Fatalf("eve blocker = %s, want incomplete_attendance", resp.BlockedChildren[2].Blockers[0].Code)
	}
	if resp.BlockedChildren[2].Blockers[0].SessionID == nil {
		t.Fatal("eve blocker should have session_id")
	}

	if resp.BlockedChildren[3].ChildFirstName != "Grace Issued Invoice" {
		t.Fatalf("fourth blocked = %s, want Grace Issued Invoice", resp.BlockedChildren[3].ChildFirstName)
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
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
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

// --- Draft Generation Tests (API-17) ---

func TestBillingGenerationRouteInventory(t *testing.T) {
	h := setupBillingHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"POST /api/v1/invoice-runs/drafts",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestBillingGenerationUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", "", `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestBillingGenerationRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)

	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", token, `{"billing_month":"2026-05"}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestBillingGenerationValidationErrors(t *testing.T) {
	h := setupBillingHarness(t)

	// Missing billing_month
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{}`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")

	// Malformed billing_month
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"invalid"}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")

	// Invalid child_id
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05","child_ids":["not-a-uuid"]}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")
}

func TestBillingGenerationFullMonthCreatesDrafts(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Seed one eligible child
	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000001")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000001")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000001")
	session1 := uuid.MustParse("c9000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Alice Eligible",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Alice", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}
	checkIn1 := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	checkOut1 := time.Date(2026, 5, 15, 16, 0, 0, 0, time.UTC)
	localDate1 := dbtest.DateAt(2026, 5, 15)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session1, h.tenantID, h.branchID, child1, checkIn1, checkOut1, localDate1, localDate1)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	// Generate
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success_count = %d, want 1", resp.Summary.SuccessCount)
	}
	if resp.Summary.BlockedCount != 0 {
		t.Fatalf("blocked_count = %d, want 0", resp.Summary.BlockedCount)
	}
	if resp.Status != "completed" {
		t.Fatalf("status = %s, want completed", resp.Status)
	}
	if len(resp.Generated) != 1 {
		t.Fatalf("generated = %d, want 1", len(resp.Generated))
	}
	if resp.Generated[0].Action != "created" {
		t.Fatalf("action = %s, want created", resp.Generated[0].Action)
	}
	if resp.Generated[0].ChildFirstName != "Alice Eligible" {
		t.Fatalf("child_first_name = %s, want Alice Eligible", resp.Generated[0].ChildFirstName)
	}

	// Verify DB state: 1 run, 1 invoice, 2 lines, 1 audit
	var runCount, invoiceCount, lineCount, auditCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_runs WHERE tenant_id = $1", h.tenantID).Scan(&runCount)
	if runCount != 1 {
		t.Fatalf("invoice_runs = %d, want 1", runCount)
	}
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1 AND status = 'draft'", h.tenantID).Scan(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("draft invoices = %d, want 1", invoiceCount)
	}
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_lines WHERE tenant_id = $1", h.tenantID).Scan(&lineCount)
	if lineCount != 2 {
		t.Fatalf("invoice_lines = %d, want 2 (core + deduction)", lineCount)
	}
	h.pool.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action_type = 'invoice_draft_generated'", h.tenantID).Scan(&auditCount)
	if auditCount != 1 {
		t.Fatalf("audit logs = %d, want 1", auditCount)
	}

	// Verify totals: 480min * 500/hr = 4000, funded 300min * 500/hr = 2500, total = 1500
	if resp.Generated[0].SubtotalMinor != 4000 {
		t.Fatalf("subtotal = %d, want 4000", resp.Generated[0].SubtotalMinor)
	}
	if resp.Generated[0].FundedDeductionMinor != 2500 {
		t.Fatalf("funded_deduction = %d, want 2500", resp.Generated[0].FundedDeductionMinor)
	}
	if resp.Generated[0].TotalDueMinor != 1500 {
		t.Fatalf("total_due = %d, want 1500", resp.Generated[0].TotalDueMinor)
	}
}

func TestBillingGenerationRerunUpdatesSameDraft(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000002")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000002")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000002")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000002")
	session1 := uuid.MustParse("c9000000-0000-0000-0000-000000000002")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Bob Rerun",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Bob", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}
	checkIn1 := time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC)
	checkOut1 := time.Date(2026, 5, 10, 16, 0, 0, 0, time.UTC)
	localDate1 := dbtest.DateAt(2026, 5, 10)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session1, h.tenantID, h.branchID, child1, checkIn1, checkOut1, localDate1, localDate1)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	// First generation
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp1 genDraftsResponse
	decodeJSON(t, w, &resp1)
	invoiceID := resp1.Generated[0].InvoiceID

	// Second generation (rerun)
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp2 genDraftsResponse
	decodeJSON(t, w, &resp2)

	if resp2.Generated[0].Action != "updated" {
		t.Fatalf("rerun action = %s, want updated", resp2.Generated[0].Action)
	}
	if resp2.Generated[0].InvoiceID != invoiceID {
		t.Fatalf("rerun invoice_id = %s, want %s (same as first)", resp2.Generated[0].InvoiceID, invoiceID)
	}

	// Still only 1 invoice
	var invoiceCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1 AND child_id = $2", h.tenantID, child1).Scan(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("invoices = %d, want 1 (not duplicated)", invoiceCount)
	}

	// Still 2 lines (replaced, not accumulated)
	var lineCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_lines il JOIN invoices i ON il.invoice_id = i.id WHERE i.tenant_id = $1 AND i.child_id = $2", h.tenantID, child1).Scan(&lineCount)
	if lineCount != 2 {
		t.Fatalf("lines = %d, want 2 (core + deduction)", lineCount)
	}

	// Two runs
	var runCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_runs WHERE tenant_id = $1", h.tenantID).Scan(&runCount)
	if runCount != 2 {
		t.Fatalf("runs = %d, want 2", runCount)
	}

	// Audit has both generated and regenerated
	var genCount, regenCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE action_type = 'invoice_draft_generated' AND action_entity_id::text = $1", invoiceID).Scan(&genCount)
	h.pool.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE action_type = 'invoice_draft_regenerated' AND action_entity_id::text = $1", invoiceID).Scan(&regenCount)
	if genCount != 1 {
		t.Fatalf("draft_generated audits = %d, want 1", genCount)
	}
	if regenCount != 1 {
		t.Fatalf("draft_regenerated audits = %d, want 1", regenCount)
	}
}

func TestBillingGenerationIssuedInvoiceBlocked(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000003")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000003")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000003")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000003")
	invoice1 := uuid.MustParse("ca000000-0000-0000-0000-000000000003")
	run1 := uuid.MustParse("cb000000-0000-0000-0000-000000000003")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Grace Issued",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Grace", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 300)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}
	managerUID := uuid.MustParse("b3000000-0000-0000-0000-000000000001")
	managerMID := uuid.MustParse("b4000000-0000-0000-0000-000000000001")
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-issued-test')`,
		run1, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), managerUID, managerMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  period_start_date, period_end_date, invoice_number, issued_sequence, generated_run_id, issued_run_id,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $7, 'INV-TEST', 1, $8, $8, now(), $9, $10, now(), now() + interval '7 days')`,
		invoice1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31),
		run1, managerUID, managerMID)
	if err != nil {
		t.Fatalf("insert issued invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 0 {
		t.Fatalf("success_count = %d, want 0", resp.Summary.SuccessCount)
	}
	if resp.Summary.BlockedCount != 1 {
		t.Fatalf("blocked_count = %d, want 1", resp.Summary.BlockedCount)
	}
	if len(resp.Blocked) != 1 {
		t.Fatalf("blocked = %d, want 1", len(resp.Blocked))
	}
	if resp.Blocked[0].Blockers[0].Code != "invoice_already_issued" {
		t.Fatalf("blocker = %s, want invoice_already_issued", resp.Blocked[0].Blockers[0].Code)
	}
	if resp.Status != "completed_with_exceptions" {
		t.Fatalf("status = %s, want completed_with_exceptions", resp.Status)
	}

	// Issued invoice not mutated
	var status string
	h.pool.QueryRow(ctx, "SELECT status FROM invoices WHERE id = $1", invoice1).Scan(&status)
	if status != "issued" {
		t.Fatalf("issued invoice status = %s, want issued (unchanged)", status)
	}
}

func TestBillingGenerationSelectedChildren(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Two eligible children
	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000010")
	child2 := uuid.MustParse("c5000000-0000-0000-0000-000000000011")
	for i, cid := range []uuid.UUID{child1, child2} {
		name := fmt.Sprintf("Child %d", i+1)
		gid := uuid.MustParse(fmt.Sprintf("c6000000-0000-0000-0000-%012d", i+10))
		lid := uuid.MustParse(fmt.Sprintf("c7000000-0000-0000-0000-%012d", i+10))
		fid := uuid.MustParse(fmt.Sprintf("c8000000-0000-0000-0000-%012d", i+10))

		dbtest.InsertChild(t, h.pool, cid, h.tenantID, h.branchID, name,
			dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
		dbtest.InsertGuardian(t, h.pool, gid, h.tenantID, h.branchID, "G "+name, true)
		dbtest.InsertGuardianLink(t, h.pool, lid, h.tenantID, h.branchID, gid, cid)
		_, err := h.pool.Exec(ctx,
			"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
			fid, h.tenantID, h.branchID, cid, dbtest.DateAt(2026, 5, 1), 0)
		if err != nil {
			t.Fatalf("insert funding %d: %v", i, err)
		}
		sid := uuid.MustParse(fmt.Sprintf("c9000000-0000-0000-0000-%012d", i+10))
		checkIn := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
		checkOut := time.Date(2026, 5, 15, 16, 0, 0, 0, time.UTC)
		localDate := dbtest.DateAt(2026, 5, 15)
		_, err = h.pool.Exec(ctx,
			"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
			sid, h.tenantID, h.branchID, cid, checkIn, checkOut, localDate, localDate)
		if err != nil {
			t.Fatalf("insert session %d: %v", i, err)
		}
	}

	// Generate only child1
	body := fmt.Sprintf(`{"billing_month":"2026-05","child_ids":["%s"]}`, child1)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success_count = %d, want 1", resp.Summary.SuccessCount)
	}
	if resp.Generated[0].ChildID != child1.String() {
		t.Fatalf("generated child = %s, want %s", resp.Generated[0].ChildID, child1)
	}

	// Only child1 has an invoice
	var invoiceCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1", h.tenantID).Scan(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("invoices = %d, want 1 (only selected child)", invoiceCount)
	}
}

func TestBillingGenerationSelectedUnknownChild(t *testing.T) {
	h := setupBillingHarness(t)

	unknownID := uuid.MustParse("d0000000-0000-0000-0000-000000000001")
	body := fmt.Sprintf(`{"billing_month":"2026-05","child_ids":["%s"]}`, unknownID)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.BlockedCount != 1 {
		t.Fatalf("blocked_count = %d, want 1", resp.Summary.BlockedCount)
	}
	if len(resp.Blocked) != 1 {
		t.Fatalf("blocked = %d, want 1", len(resp.Blocked))
	}
	if resp.Blocked[0].Blockers[0].Code != "child_not_found" {
		t.Fatalf("blocker = %s, want child_not_found", resp.Blocked[0].Blockers[0].Code)
	}
}

func TestBillingGenerationEmptyChildIDsNoop(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05","child_ids":[]}`)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 0 {
		t.Fatalf("success_count = %d, want 0", resp.Summary.SuccessCount)
	}
	if resp.Summary.BlockedCount != 0 {
		t.Fatalf("blocked_count = %d, want 0", resp.Summary.BlockedCount)
	}
}

func TestBillingGenerationDuplicateChildIDsDeduped(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000020")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000020")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000020")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000020")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Dedup Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Dedup", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	body := fmt.Sprintf(`{"billing_month":"2026-05","child_ids":["%s","%s"]}`, child1, child1)
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, body)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success_count = %d, want 1 (deduped)", resp.Summary.SuccessCount)
	}

	var invoiceCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1 AND child_id = $2", h.tenantID, child1).Scan(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("invoices = %d, want 1 (deduped)", invoiceCount)
	}
}

func TestBillingGenerationZeroAttendanceGetsDraft(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000030")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000030")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000030")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000030")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Zero Attendance",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Zero", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}
	// No attendance sessions

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success_count = %d, want 1", resp.Summary.SuccessCount)
	}
	if resp.Generated[0].TotalDueMinor != 0 {
		t.Fatalf("total_due = %d, want 0", resp.Generated[0].TotalDueMinor)
	}
}

func TestBillingGenerationPreservesExtraLines(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	child1 := uuid.MustParse("c5000000-0000-0000-0000-000000000040")
	guardian1 := uuid.MustParse("c6000000-0000-0000-0000-000000000040")
	link1 := uuid.MustParse("c7000000-0000-0000-0000-000000000040")
	funding1 := uuid.MustParse("c8000000-0000-0000-0000-000000000040")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Extra Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardian1, h.tenantID, h.branchID, "Guardian Extra", true)
	dbtest.InsertGuardianLink(t, h.pool, link1, h.tenantID, h.branchID, guardian1, child1)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		funding1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	// First generation
	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)
	var resp1 genDraftsResponse
	decodeJSON(t, w, &resp1)
	invoiceID := resp1.Generated[0].InvoiceID

	// Insert an extra line manually
	extraLineID := uuid.MustParse("cc000000-0000-0000-0000-000000000040")
	extraInvoiceID, _ := uuid.Parse(invoiceID)
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, line_amount_minor, details)
		 VALUES ($1, $2, $3, $4, 'extra', 'Late pickup fee', 3, 500, '{}'::jsonb)`,
		extraLineID, h.tenantID, h.branchID, extraInvoiceID)
	if err != nil {
		t.Fatalf("insert extra line: %v", err)
	}

	// Rerun generation
	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	// Extra line preserved
	var extraCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_lines WHERE invoice_id = $1 AND line_kind = 'extra'", extraInvoiceID).Scan(&extraCount)
	if extraCount != 1 {
		t.Fatalf("extra lines = %d, want 1 (preserved)", extraCount)
	}

	// Total lines = 2 system + 1 extra = 3
	var totalLines int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoice_lines WHERE invoice_id = $1", extraInvoiceID).Scan(&totalLines)
	if totalLines != 3 {
		t.Fatalf("total lines = %d, want 3", totalLines)
	}

	// Verify the response includes the extras in total
	var resp2 genDraftsResponse
	decodeJSON(t, w, &resp2)
	// core_subtotal = 0 (zero attendance), extras = 500, subtotal = 500, funded_deduction = 0, total = 500
	if resp2.Generated[0].TotalDueMinor != 500 {
		t.Fatalf("total_due = %d, want 500 (0 core + 500 extra)", resp2.Generated[0].TotalDueMinor)
	}
}

// JSON response types for test deserialization
type preflightResponse struct {
	BillingMonth     string                  `json:"billing_month"`
	CurrencyCode     string                  `json:"currency_code"`
	Period           periodResponse          `json:"period"`
	Summary          summaryResponse         `json:"summary"`
	EligibleChildren []eligibleChildResponse `json:"eligible_children"`
	BlockedChildren  []blockedChildResponse  `json:"blocked_children"`
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
	ChildFirstName         string              `json:"child_first_name"`
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
	ChildID        string            `json:"child_id"`
	ChildFirstName string            `json:"child_first_name"`
	Blockers       []blockerResponse `json:"blockers"`
}

type blockerResponse struct {
	Code             string  `json:"code"`
	Message          string  `json:"message"`
	SessionID        *string `json:"session_id,omitempty"`
	CheckInAt        *string `json:"check_in_at,omitempty"`
	CheckInLocalDate *string `json:"check_in_local_date,omitempty"`
	InvoiceID        *string `json:"invoice_id,omitempty"`
	InvoiceStatus    *string `json:"invoice_status,omitempty"`
	Field            *string `json:"field,omitempty"`
}

type blockerCountResponse struct {
	Code          string `json:"code"`
	ChildrenCount int    `json:"children_count"`
}

// Generation response types
type genDraftsResponse struct {
	RunID        string                    `json:"run_id"`
	BillingMonth string                    `json:"billing_month"`
	Status       string                    `json:"status"`
	Summary      genDraftsSummary          `json:"summary"`
	Generated    []genDraftChildResponse   `json:"generated"`
	Blocked      []genBlockedChildResponse `json:"blocked"`
}

type genDraftsSummary struct {
	EligibleCount int `json:"eligible_count"`
	SuccessCount  int `json:"success_count"`
	BlockedCount  int `json:"blocked_count"`
	TotalDueMinor int `json:"total_due_minor"`
}

type genDraftChildResponse struct {
	ChildID              string `json:"child_id"`
	ChildFirstName       string `json:"child_first_name"`
	Action               string `json:"action"`
	InvoiceID            string `json:"invoice_id"`
	SubtotalMinor        int    `json:"subtotal_minor"`
	FundedDeductionMinor int    `json:"funded_deduction_minor"`
	TotalDueMinor        int    `json:"total_due_minor"`
}

type genBlockedChildResponse struct {
	ChildID        string               `json:"child_id"`
	ChildFirstName string               `json:"child_first_name,omitempty"`
	Blockers       []genBlockerResponse `json:"blockers"`
}

type genBlockerResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestBillingGenerationCriticalCalculationSnapshot(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("f5000000-0000-0000-0000-000000000001")
	guardianID := uuid.MustParse("f6000000-0000-0000-0000-000000000001")
	linkID := uuid.MustParse("f7000000-0000-0000-0000-000000000001")
	fundingID := uuid.MustParse("f8000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Snapshot Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Guardian Snapshot", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), 30)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	session1 := uuid.MustParse("f9000000-0000-0000-0000-000000000001")
	session2 := uuid.MustParse("f9000000-0000-0000-0000-000000000002")

	checkIn1 := time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC)
	checkOut1 := time.Date(2026, 5, 12, 9, 1, 0, 0, time.UTC)
	localDate1 := dbtest.DateAt(2026, 5, 12)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session1, h.tenantID, h.branchID, childID, checkIn1, checkOut1, localDate1, localDate1)
	if err != nil {
		t.Fatalf("insert session 1: %v", err)
	}

	checkIn2 := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	checkOut2 := time.Date(2026, 5, 12, 10, 15, 0, 0, time.UTC)
	localDate2 := dbtest.DateAt(2026, 5, 12)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session2, h.tenantID, h.branchID, childID, checkIn2, checkOut2, localDate2, localDate2)
	if err != nil {
		t.Fatalf("insert session 2: %v", err)
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp genDraftsResponse
	decodeJSON(t, w, &resp)

	if resp.Summary.SuccessCount != 1 {
		t.Fatalf("success_count = %d, want 1", resp.Summary.SuccessCount)
	}
	gen := resp.Generated[0]
	if gen.Action != "created" {
		t.Fatalf("action = %s, want created", gen.Action)
	}
	if gen.SubtotalMinor != 602 {
		t.Fatalf("subtotal_minor = %d, want 602", gen.SubtotalMinor)
	}
	if gen.FundedDeductionMinor != 201 {
		t.Fatalf("funded_deduction_minor = %d, want 201", gen.FundedDeductionMinor)
	}
	if gen.TotalDueMinor != 401 {
		t.Fatalf("total_due_minor = %d, want 401", gen.TotalDueMinor)
	}

	invoiceUUID, _ := uuid.Parse(gen.InvoiceID)

	var subtotalMinor, fundedDeductMinor, totalDueMinor int
	var calcDetails string
	err = h.pool.QueryRow(ctx,
		`SELECT subtotal_minor, funded_deduction_minor, total_due_minor, calculation_details
		 FROM invoices WHERE id = $1`, invoiceUUID).
		Scan(&subtotalMinor, &fundedDeductMinor, &totalDueMinor, &calcDetails)
	if err != nil {
		t.Fatalf("query invoice: %v", err)
	}

	if subtotalMinor != 602 {
		t.Fatalf("subtotal_minor = %d, want 602", subtotalMinor)
	}
	if fundedDeductMinor != 201 {
		t.Fatalf("funded_deduction_minor = %d, want 201", fundedDeductMinor)
	}
	if totalDueMinor != 401 {
		t.Fatalf("total_due_minor = %d, want 401", totalDueMinor)
	}

	type calcDetailJSON struct {
		RawAttendedMinutes     int `json:"raw_attended_minutes"`
		RoundedAttendedMinutes int `json:"rounded_attended_minutes"`
		FundedAllowanceMinutes int `json:"funded_allowance_minutes"`
		FundedDeductionMinutes int `json:"funded_deduction_minutes"`
		CoreBillableMinutes    int `json:"core_billable_minutes"`
		IncludedSessionCount   int `json:"included_session_count"`
		SourceSessions         []struct {
			SessionID string `json:"session_id"`
		} `json:"source_sessions"`
	}
	var cd calcDetailJSON
	if err := json.Unmarshal([]byte(calcDetails), &cd); err != nil {
		t.Fatalf("parse calculation_details: %v", err)
	}
	if cd.RawAttendedMinutes != 76 {
		t.Fatalf("calc raw_attended_minutes = %d, want 76", cd.RawAttendedMinutes)
	}
	if cd.RoundedAttendedMinutes != 90 {
		t.Fatalf("calc rounded_attended_minutes = %d, want 90", cd.RoundedAttendedMinutes)
	}
	if cd.FundedAllowanceMinutes != 30 {
		t.Fatalf("calc funded_allowance_minutes = %d, want 30", cd.FundedAllowanceMinutes)
	}
	if cd.FundedDeductionMinutes != 30 {
		t.Fatalf("calc funded_deduction_minutes = %d, want 30", cd.FundedDeductionMinutes)
	}
	if cd.CoreBillableMinutes != 60 {
		t.Fatalf("calc core_billable_minutes = %d, want 60", cd.CoreBillableMinutes)
	}
	if cd.IncludedSessionCount != 2 {
		t.Fatalf("included_session_count = %d, want 2", cd.IncludedSessionCount)
	}
	if len(cd.SourceSessions) != 2 {
		t.Fatalf("source_sessions length = %d, want 2", len(cd.SourceSessions))
	}

	var coreLineCount int
	err = h.pool.QueryRow(ctx,
		`SELECT count(*) FROM invoice_lines WHERE invoice_id = $1 AND line_kind = 'core_childcare'`, invoiceUUID).Scan(&coreLineCount)
	if err != nil {
		t.Fatalf("count core lines: %v", err)
	}
	if coreLineCount != 1 {
		t.Fatalf("core_childcare lines = %d, want 1", coreLineCount)
	}

	var coreQty, coreAmount, coreRaw, coreRounded, coreBillableLine, coreSessions int
	err = h.pool.QueryRow(ctx,
		`SELECT quantity_minutes, line_amount_minor, raw_attended_minutes, rounded_attended_minutes, core_billable_minutes, session_count
		 FROM invoice_lines WHERE invoice_id = $1 AND line_kind = 'core_childcare'`, invoiceUUID).
		Scan(&coreQty, &coreAmount, &coreRaw, &coreRounded, &coreBillableLine, &coreSessions)
	if err != nil {
		t.Fatalf("query core line: %v", err)
	}
	if coreQty != 90 {
		t.Fatalf("core quantity_minutes = %d, want 90", coreQty)
	}
	if coreAmount != 602 {
		t.Fatalf("core line_amount_minor = %d, want 602", coreAmount)
	}
	if coreRaw != 76 {
		t.Fatalf("core raw_attended_minutes = %d, want 76", coreRaw)
	}
	if coreRounded != 90 {
		t.Fatalf("core rounded_attended_minutes = %d, want 90", coreRounded)
	}
	if coreBillableLine != 0 {
		t.Fatalf("core core_billable_minutes = %d, want 0 (populated only on funded_deduction lines)", coreBillableLine)
	}
	if coreSessions != 2 {
		t.Fatalf("core session_count = %d, want 2", coreSessions)
	}

	var fundAllow, fundDeduct, fundBillable, fundAmount int
	err = h.pool.QueryRow(ctx,
		`SELECT funded_allowance_minutes, funded_deduction_minutes, core_billable_minutes, line_amount_minor
		 FROM invoice_lines WHERE invoice_id = $1 AND line_kind = 'funded_deduction'`, invoiceUUID).
		Scan(&fundAllow, &fundDeduct, &fundBillable, &fundAmount)
	if err != nil {
		t.Fatalf("query funded line: %v", err)
	}
	if fundAllow != 30 {
		t.Fatalf("funded allowance = %d, want 30", fundAllow)
	}
	if fundDeduct != 30 {
		t.Fatalf("funded deduction = %d, want 30", fundDeduct)
	}
	if fundBillable != 60 {
		t.Fatalf("funded core_billable = %d, want 60", fundBillable)
	}
	if fundAmount != -201 {
		t.Fatalf("funded line_amount = %d, want -201", fundAmount)
	}
}

func TestBillingGenerationRerunRecalculatesSameDraftAndReplacesSystemLines(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("f5000000-0000-0000-0000-000000000010")
	guardianID := uuid.MustParse("f6000000-0000-0000-0000-000000000010")
	linkID := uuid.MustParse("f7000000-0000-0000-0000-000000000010")
	fundingID := uuid.MustParse("f8000000-0000-0000-0000-000000000010")
	session1 := uuid.MustParse("f9000000-0000-0000-0000-000000000010")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Rerun Calc Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Guardian Rerun", true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	_, err := h.pool.Exec(ctx,
		"INSERT INTO funding_profiles (id, tenant_id, branch_id, child_id, billing_month, funded_allowance_minutes) VALUES ($1, $2, $3, $4, $5, $6)",
		fundingID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), 0)
	if err != nil {
		t.Fatalf("insert funding: %v", err)
	}

	checkIn1 := time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC)
	checkOut1 := time.Date(2026, 5, 10, 9, 0, 0, 0, time.UTC)
	localDate1 := dbtest.DateAt(2026, 5, 10)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session1, h.tenantID, h.branchID, childID, checkIn1, checkOut1, localDate1, localDate1)
	if err != nil {
		t.Fatalf("insert session 1: %v", err)
	}

	w := doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp1 genDraftsResponse
	decodeJSON(t, w, &resp1)

	if resp1.Summary.SuccessCount != 1 {
		t.Fatalf("first run success = %d, want 1", resp1.Summary.SuccessCount)
	}
	invoiceID := resp1.Generated[0].InvoiceID
	if resp1.Generated[0].TotalDueMinor != 600 {
		t.Fatalf("first total_due = %d, want 600", resp1.Generated[0].TotalDueMinor)
	}

	var initialSystemLines int
	h.pool.QueryRow(ctx,
		"SELECT count(*) FROM invoice_lines WHERE invoice_id = $1 AND line_kind IN ('core_childcare','funded_deduction')", invoiceID).
		Scan(&initialSystemLines)
	if initialSystemLines != 2 {
		t.Fatalf("initial system lines = %d, want 2", initialSystemLines)
	}

	session2 := uuid.MustParse("f9000000-0000-0000-0000-000000000011")
	checkIn2 := time.Date(2026, 5, 11, 8, 0, 0, 0, time.UTC)
	checkOut2 := time.Date(2026, 5, 11, 8, 15, 0, 0, time.UTC)
	localDate2 := dbtest.DateAt(2026, 5, 11)
	_, err = h.pool.Exec(ctx,
		"INSERT INTO attendance_sessions (id, tenant_id, branch_id, child_id, status, check_in_at, check_out_at, check_in_local_date, check_out_local_date) VALUES ($1, $2, $3, $4, 'complete', $5, $6, $7, $8)",
		session2, h.tenantID, h.branchID, childID, checkIn2, checkOut2, localDate2, localDate2)
	if err != nil {
		t.Fatalf("insert session 2: %v", err)
	}

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.managerToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusOK)

	var resp2 genDraftsResponse
	decodeJSON(t, w, &resp2)

	if resp2.Generated[0].Action != "updated" {
		t.Fatalf("rerun action = %s, want updated", resp2.Generated[0].Action)
	}
	if resp2.Generated[0].InvoiceID != invoiceID {
		t.Fatalf("rerun invoice_id changed: %s vs %s", resp2.Generated[0].InvoiceID, invoiceID)
	}

	var invoiceCount int
	h.pool.QueryRow(ctx, "SELECT count(*) FROM invoices WHERE tenant_id = $1 AND child_id = $2", h.tenantID, childID).Scan(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("invoice count = %d, want 1 (no duplication)", invoiceCount)
	}

	var systemLines int
	h.pool.QueryRow(ctx,
		"SELECT count(*) FROM invoice_lines WHERE invoice_id = $1 AND line_kind IN ('core_childcare','funded_deduction')", invoiceID).
		Scan(&systemLines)
	if systemLines != 2 {
		t.Fatalf("system lines = %d, want 2 (replaced, not accumulated)", systemLines)
	}

	var roundedAtt int
	h.pool.QueryRow(ctx,
		"SELECT rounded_attended_minutes FROM invoice_lines WHERE invoice_id = $1 AND line_kind = 'core_childcare'", invoiceID).
		Scan(&roundedAtt)
	if roundedAtt != 75 {
		t.Fatalf("core rounded = %d, want 75 (60+15)", roundedAtt)
	}

	if resp2.Generated[0].TotalDueMinor != 750 {
		t.Fatalf("rerun total_due = %d, want 750", resp2.Generated[0].TotalDueMinor)
	}

	invoiceUUID, _ := uuid.Parse(invoiceID)
	var genAuditCount, regenAuditCount int
	h.pool.QueryRow(ctx,
		"SELECT count(*) FROM audit_logs WHERE action_type = 'invoice_draft_generated' AND action_entity_id = $1", invoiceUUID).Scan(&genAuditCount)
	h.pool.QueryRow(ctx,
		"SELECT count(*) FROM audit_logs WHERE action_type = 'invoice_draft_regenerated' AND action_entity_id = $1", invoiceUUID).Scan(&regenAuditCount)
	if genAuditCount != 1 {
		t.Fatalf("draft_generated audits = %d, want 1", genAuditCount)
	}
	if regenAuditCount != 1 {
		t.Fatalf("draft_regenerated audits = %d, want 1", regenAuditCount)
	}
}

// --- API-18 Manager Invoice Review tests ---

func TestInvoiceReviewRouteInventory(t *testing.T) {
	h := setupBillingHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/invoices",
		"GET /api/v1/invoices/:invoice_id",
	}
	for _, want := range expected {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestInvoiceListUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestInvoiceListRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)
	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestInvoiceDetailUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+uuid.New().String(), "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestInvoiceDetailRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)
	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+uuid.New().String(), token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestInvoiceListValidationErrors(t *testing.T) {
	h := setupBillingHarness(t)

	cases := []struct {
		name string
		url  string
	}{
		{"bad billing_month", "/api/v1/invoices?billing_month=invalid"},
		{"bad status", "/api/v1/invoices?status=unknown"},
		{"bad child_id", "/api/v1/invoices?child_id=not-a-uuid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(t, h.router, http.MethodGet, tc.url, h.managerToken, "")
			requireStatus(t, w, http.StatusUnprocessableEntity)
			requireErrorCode(t, w, "validation_error")
		})
	}
}

func TestInvoiceDetailValidationErrors(t *testing.T) {
	h := setupBillingHarness(t)
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/not-a-uuid", h.managerToken, "")
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")
}

func TestInvoiceListEmpty(t *testing.T) {
	h := setupBillingHarness(t)
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp invoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(resp.Items))
	}
	if resp.Limit != 50 {
		t.Fatalf("expected default limit 50, got %d", resp.Limit)
	}
	if resp.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", resp.Offset)
	}
}

func TestInvoiceListDraftInvoice(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Seed child, guardian, link, run, and draft invoice
	childID := uuid.MustParse("c1000000-0000-0000-0000-000000000001")
	runID := uuid.MustParse("c2000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("c3000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Draft Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id, details)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed_with_exceptions', now(), now(), $5, $6, 'req-inv', $7)`,
		runID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1),
		uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
		uuid.MustParse("b4000000-0000-0000-0000-000000000001"),
		`{"blocked_children":[{"child_id":"b5000000-0000-0000-0000-000000000099","child_first_name":"Blocked Kid","blocker_codes":["missing_funding_profile"]}]}`)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	calcDetails := `{"billing_month":"2026-05","child_id":"` + childID.String() + `","core_hourly_rate_minor":500,"core_subtotal_minor":4000,"extras_total_minor":0,"manual_extras_supported":false,"funded_allowance_minutes":300,"funded_deduction_minutes":300,"core_billable_minutes":180,"raw_attended_minutes":480,"rounded_attended_minutes":480,"included_session_count":1,"source_sessions":[{"session_id":"00000000-0000-0000-0000-000000000001","status":"complete","check_in_at":"2026-05-15T08:00:00Z","check_out_at":"2026-05-15T16:00:00Z","raw_elapsed_minutes":480,"rounded_billable_minutes":480}]}`

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, generated_run_id, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date, calculation_details)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, 4000, 2500, 1500, $7, $8, $9)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), runID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), calcDetails)
	if err != nil {
		t.Fatalf("insert draft invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp invoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}

	item := resp.Items[0]
	if item.InvoiceNumber != nil {
		t.Fatalf("expected invoice_number null for draft, got %v", item.InvoiceNumber)
	}
	if item.InvoiceNumberDisplay != "Draft" {
		t.Fatalf("expected invoice_number_display 'Draft', got %s", item.InvoiceNumberDisplay)
	}
	if item.DueAt != nil {
		t.Fatalf("expected due_at null for draft, got %v", item.DueAt)
	}
	if item.DueStatus != "not_due" {
		t.Fatalf("expected due_status 'not_due', got %s", item.DueStatus)
	}
	if item.Status != "draft" {
		t.Fatalf("expected status 'draft', got %s", item.Status)
	}
	if item.GeneratedRunExceptionCount != 1 {
		t.Fatalf("expected exception_count 1, got %d", item.GeneratedRunExceptionCount)
	}
	if item.GeneratedRunID == nil || *item.GeneratedRunID != runID.String() {
		t.Fatalf("expected generated_run_id %s", runID.String())
	}
	if item.ChildFirstName != "Draft Child" {
		t.Fatalf("expected child_first_name 'Draft Child', got %s", item.ChildFirstName)
	}
}

func TestInvoiceListFilters(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Seed two draft invoices in different months
	child1 := uuid.MustParse("c4000000-0000-0000-0000-000000000001")
	child2 := uuid.MustParse("c4000000-0000-0000-0000-000000000002")
	inv1 := uuid.MustParse("c5000000-0000-0000-0000-000000000001")
	inv2 := uuid.MustParse("c5000000-0000-0000-0000-000000000002")

	dbtest.InsertChild(t, h.pool, child1, h.tenantID, h.branchID, "Filter Child 1",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	dbtest.InsertChild(t, h.pool, child2, h.tenantID, h.branchID, "Filter Child 2",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 1000, 0, 1000, $6, $7)`,
		inv1, h.tenantID, h.branchID, child1, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert inv1: %v", err)
	}

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 2000, 0, 2000, $6, $7)`,
		inv2, h.tenantID, h.branchID, child2, dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 30))
	if err != nil {
		t.Fatalf("insert inv2: %v", err)
	}

	// Filter by billing_month
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices?billing_month=2026-05", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	var resp invoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("billing_month filter: expected 1, got %d", len(resp.Items))
	}

	// Filter by child_id
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices?child_id="+child2.String(), h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("child_id filter: expected 1, got %d", len(resp.Items))
	}

	// Filter by status
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices?status=draft", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 2 {
		t.Fatalf("status filter: expected 2, got %d", len(resp.Items))
	}

	// Filter by non-matching status
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices?status=paid", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("status=paid filter: expected 0, got %d", len(resp.Items))
	}
}

func TestInvoiceDetailDraft(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("c6000000-0000-0000-0000-000000000001")
	runID := uuid.MustParse("c7000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("c8000000-0000-0000-0000-000000000001")
	lineID1 := uuid.MustParse("c9000000-0000-0000-0000-000000000001")
	lineID2 := uuid.MustParse("c9000000-0000-0000-0000-000000000002")
	sessionID := uuid.MustParse("c9000000-0000-0000-0000-000000000003")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Detail Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id, details)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed_with_exceptions', now(), now(), $5, $6, 'req-detail', $7)`,
		runID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1),
		uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
		uuid.MustParse("b4000000-0000-0000-0000-000000000001"),
		`{"blocked_children":[{"child_id":"`+childID.String()+`","child_first_name":"Detail Child","blocker_codes":["missing_funding_profile"]}]}`)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	calcDetails := `{"billing_month":"2026-05","child_id":"` + childID.String() + `","core_hourly_rate_minor":500,"core_subtotal_minor":4000,"extras_total_minor":0,"manual_extras_supported":false,"funded_allowance_minutes":300,"funded_deduction_minutes":300,"core_billable_minutes":180,"raw_attended_minutes":480,"rounded_attended_minutes":480,"included_session_count":1,"source_sessions":[{"session_id":"` + sessionID.String() + `","status":"complete","check_in_at":"2026-05-15T08:00:00Z","check_out_at":"2026-05-15T16:00:00Z","raw_elapsed_minutes":480,"rounded_billable_minutes":480}]}`

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, generated_run_id, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date, calculation_details)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, 4000, 2500, 1500, $7, $8, $9)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), runID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), calcDetails)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	// Insert invoice lines
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, raw_attended_minutes, rounded_attended_minutes, session_count)
		 VALUES ($1, $2, $3, $4, 'core_childcare', 'Core childcare', 1, 480, 500, 4000, 480, 480, 1)`,
		lineID1, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line 1: %v", err)
	}

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, line_amount_minor, funded_allowance_minutes, funded_deduction_minutes, core_billable_minutes)
		 VALUES ($1, $2, $3, $4, 'funded_deduction', 'Funded hours deduction', 2, -2500, 300, 300, 180)`,
		lineID2, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line 2: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+invoiceID.String(), h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp invoiceDetailResponseTest
	decodeJSON(t, w, &resp)

	// Header checks
	if resp.InvoiceNumber != nil {
		t.Fatalf("expected invoice_number null for draft, got %v", resp.InvoiceNumber)
	}
	if resp.InvoiceNumberDisplay != "Draft" {
		t.Fatalf("expected invoice_number_display 'Draft', got %s", resp.InvoiceNumberDisplay)
	}
	if resp.DueStatus != "not_due" {
		t.Fatalf("expected due_status 'not_due', got %s", resp.DueStatus)
	}
	if resp.Status != "draft" {
		t.Fatalf("expected status 'draft', got %s", resp.Status)
	}
	if resp.SubtotalMinor != 4000 {
		t.Fatalf("expected subtotal_minor 4000, got %d", resp.SubtotalMinor)
	}
	if resp.TotalDueMinor != 1500 {
		t.Fatalf("expected total_due_minor 1500, got %d", resp.TotalDueMinor)
	}

	// Lines
	if len(resp.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(resp.Lines))
	}
	if resp.Lines[0].LineKind != "core_childcare" {
		t.Fatalf("expected line 0 kind 'core_childcare', got %s", resp.Lines[0].LineKind)
	}
	if resp.Lines[0].SortOrder != 1 {
		t.Fatalf("expected line 0 sort_order 1, got %d", resp.Lines[0].SortOrder)
	}
	if resp.Lines[1].LineKind != "funded_deduction" {
		t.Fatalf("expected line 1 kind 'funded_deduction', got %s", resp.Lines[1].LineKind)
	}

	// Calculation
	if resp.Calculation.CoreHourlyRateMinor != 500 {
		t.Fatalf("expected core_hourly_rate_minor 500, got %d", resp.Calculation.CoreHourlyRateMinor)
	}
	if resp.Calculation.RawAttendedMinutes != 480 {
		t.Fatalf("expected raw_attended_minutes 480, got %d", resp.Calculation.RawAttendedMinutes)
	}
	if len(resp.Calculation.SourceSessions) != 1 {
		t.Fatalf("expected 1 source session, got %d", len(resp.Calculation.SourceSessions))
	}
	if resp.Calculation.SourceSessions[0].CheckInAt != "2026-05-15T08:00:00Z" {
		t.Fatalf("expected source session check_in_at '2026-05-15T08:00:00Z', got %s", resp.Calculation.SourceSessions[0].CheckInAt)
	}

	// Exceptions
	if resp.GeneratedRunExceptionCount != 1 {
		t.Fatalf("expected exception_count 1, got %d", resp.GeneratedRunExceptionCount)
	}
	if len(resp.GeneratedRunExceptions) != 1 {
		t.Fatalf("expected 1 exception, got %d", len(resp.GeneratedRunExceptions))
	}
	if len(resp.GeneratedRunExceptions[0].BlockerCodes) != 1 || resp.GeneratedRunExceptions[0].BlockerCodes[0] != "missing_funding_profile" {
		t.Fatalf("expected blocker_code 'missing_funding_profile', got %v", resp.GeneratedRunExceptions[0].BlockerCodes)
	}
}

func TestInvoiceDetailNotFound(t *testing.T) {
	h := setupBillingHarness(t)
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+uuid.New().String(), h.managerToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

func TestInvoiceListScopeIsolation(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Create invoice in a different tenant/branch
	otherTenant := uuid.MustParse("d1000000-0000-0000-0000-000000000001")
	otherBranch := uuid.MustParse("d2000000-0000-0000-0000-000000000001")
	otherChild := uuid.MustParse("d3000000-0000-0000-0000-000000000001")
	otherInvoice := uuid.MustParse("d4000000-0000-0000-0000-000000000001")

	dbtest.InsertTenant(t, h.pool, otherTenant, "Other Tenant")
	dbtest.InsertBranch(t, h.pool, otherTenant, otherBranch, "Other Branch")
	dbtest.InsertChild(t, h.pool, otherChild, otherTenant, otherBranch, "Other Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 1000, 0, 1000, $6, $7)`,
		otherInvoice, otherTenant, otherBranch, otherChild, dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert other invoice: %v", err)
	}

	// List should not include other-tenant invoice
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)
	var resp invoiceListResponseTest
	decodeJSON(t, w, &resp)
	for _, item := range resp.Items {
		if item.InvoiceID == otherInvoice.String() {
			t.Fatal("other-tenant invoice should not appear in list")
		}
	}

	// Detail should 404 for other-tenant invoice
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/"+otherInvoice.String(), h.managerToken, "")
	requireStatus(t, w, http.StatusNotFound)
}

func TestInvoiceListIssuedInvoice(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("ca000000-0000-0000-0000-000000000001")
	runID := uuid.MustParse("cb000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("cc000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Issued Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)

	issuedByUID := uuid.MustParse("b3000000-0000-0000-0000-000000000001")
	issuedByMID := uuid.MustParse("b4000000-0000-0000-0000-000000000001")

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-issued')`,
		runID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), issuedByUID, issuedByMID)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $6, 4000, 0, 4000, $7, $8, 'INV-2026-05-001', 1, now(), $9, $10, now(), now() + interval '7 days')`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), runID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), issuedByUID, issuedByMID)
	if err != nil {
		t.Fatalf("insert issued invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp invoiceListResponseTest
	decodeJSON(t, w, &resp)

	var found *invoiceListItemResponseTest
	for i := range resp.Items {
		if resp.Items[i].InvoiceID == invoiceID.String() {
			found = &resp.Items[i]
			break
		}
	}
	if found == nil {
		t.Fatal("issued invoice not found in list")
	}
	if found.DueStatus != "due" {
		t.Fatalf("expected due_status 'due', got %s", found.DueStatus)
	}
	if found.InvoiceNumber == nil || *found.InvoiceNumber != "INV-2026-05-001" {
		t.Fatalf("expected invoice_number 'INV-2026-05-001', got %v", found.InvoiceNumber)
	}
	if found.InvoiceNumberDisplay != "INV-2026-05-001" {
		t.Fatalf("expected invoice_number_display 'INV-2026-05-001', got %s", found.InvoiceNumberDisplay)
	}
}

// --- API-18 test response types ---

type invoiceListResponseTest struct {
	Items  []invoiceListItemResponseTest `json:"items"`
	Limit  int                           `json:"limit"`
	Offset int                           `json:"offset"`
}

type invoiceListItemResponseTest struct {
	InvoiceID                     string  `json:"invoice_id"`
	InvoiceKind                   string  `json:"invoice_kind"`
	InvoiceNumber                 *string `json:"invoice_number"`
	InvoiceNumberDisplay          string  `json:"invoice_number_display"`
	ChildID                       string  `json:"child_id"`
	ChildFirstName                string  `json:"child_first_name"`
	BillingMonth                  string  `json:"billing_month"`
	Status                        string  `json:"status"`
	DueStatus                     string  `json:"due_status"`
	CurrencyCode                  string  `json:"currency_code"`
	SubtotalMinor                 int     `json:"subtotal_minor"`
	FundedDeductionMinor          int     `json:"funded_deduction_minor"`
	TotalDueMinor                 int     `json:"total_due_minor"`
	AmountPaidMinor               int     `json:"amount_paid_minor"`
	DueAt                         *string `json:"due_at"`
	IssuedAt                      *string `json:"issued_at"`
	PaidAt                        *string `json:"paid_at"`
	PaymentFailedAt               *string `json:"payment_failed_at"`
	PaymentStatusUpdatedAt        *string `json:"payment_status_updated_at"`
	GeneratedRunID                *string `json:"generated_run_id"`
	GeneratedRunStatus            *string `json:"generated_run_status"`
	GeneratedRunStartedAt         *string `json:"generated_run_started_at"`
	GeneratedRunCompletedAt       *string `json:"generated_run_completed_at"`
	GeneratedRunExceptionCount    int     `json:"generated_run_exception_count"`
	CreatedAt                     string  `json:"created_at"`
	UpdatedAt                     string  `json:"updated_at"`
	CheckoutRetryAvailable        bool    `json:"checkout_retry_available"`
	LatestPaymentAttemptStatus    *string `json:"latest_payment_attempt_status"`
	LatestPaymentAttemptCreatedAt *string `json:"latest_payment_attempt_created_at"`
}

type invoiceDetailResponseTest struct {
	InvoiceID                  string                            `json:"invoice_id"`
	InvoiceKind                string                            `json:"invoice_kind"`
	InvoiceNumber              *string                           `json:"invoice_number"`
	InvoiceNumberDisplay       string                            `json:"invoice_number_display"`
	ChildID                    string                            `json:"child_id"`
	ChildFirstName             string                            `json:"child_first_name"`
	BillingMonth               string                            `json:"billing_month"`
	Status                     string                            `json:"status"`
	DueStatus                  string                            `json:"due_status"`
	CurrencyCode               string                            `json:"currency_code"`
	SubtotalMinor              int                               `json:"subtotal_minor"`
	FundedDeductionMinor       int                               `json:"funded_deduction_minor"`
	TotalDueMinor              int                               `json:"total_due_minor"`
	AmountPaidMinor            int                               `json:"amount_paid_minor"`
	GeneratedRunExceptionCount int                               `json:"generated_run_exception_count"`
	GeneratedRunExceptions     []invoiceRunExceptionResponseTest `json:"generated_run_exceptions"`
	Calculation                invoiceCalculationResponseTest    `json:"calculation"`
	Lines                      []invoiceLineResponseTest         `json:"lines"`
	CreatedAt                  string                            `json:"created_at"`
	UpdatedAt                  string                            `json:"updated_at"`
}

type invoiceLineResponseTest struct {
	LineID          string `json:"line_id"`
	LineKind        string `json:"line_kind"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
	LineAmountMinor int    `json:"line_amount_minor"`
}

type invoiceCalculationResponseTest struct {
	CoreHourlyRateMinor    int                         `json:"core_hourly_rate_minor"`
	RawAttendedMinutes     int                         `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                         `json:"rounded_attended_minutes"`
	CoreBillableMinutes    int                         `json:"core_billable_minutes"`
	IncludedSessionCount   int                         `json:"included_session_count"`
	SourceSessions         []sourceSessionResponseTest `json:"source_sessions"`
}

type sourceSessionResponseTest struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	CheckInAt string `json:"check_in_at"`
}

type invoiceRunExceptionResponseTest struct {
	ChildID        string   `json:"child_id"`
	ChildFirstName string   `json:"child_first_name"`
	BlockerCodes   []string `json:"blocker_codes"`
}

// ── Billing Setup Tests ────────────────────────────────────────────────────

type billingSetupResponse struct {
	CoreHourlyRateMinor int  `json:"core_hourly_rate_minor"`
	HasRate             bool `json:"has_rate"`
}

func TestBillingSetupRouteInventory(t *testing.T) {
	h := setupBillingHarness(t)

	have := make(map[string]struct{})
	for _, route := range h.router.Routes() {
		have[route.Method+" "+route.Path] = struct{}{}
	}

	for _, want := range []string{
		"GET /api/v1/billing-setup",
		"PUT /api/v1/billing-setup",
	} {
		if _, ok := have[want]; !ok {
			t.Fatalf("expected route %s to be registered", want)
		}
	}
}

func TestBillingSetupUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/billing-setup", "", "")
	requireStatus(t, w, http.StatusUnauthorized)

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", "", `{"core_hourly_rate_minor":1500}`)
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestBillingSetupRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)

	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/billing-setup", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")

		w = doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", token, `{"core_hourly_rate_minor":1500}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestBillingSetupGetInitialNoRate(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/billing-setup", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp billingSetupResponse
	decodeJSON(t, w, &resp)
	if resp.HasRate {
		t.Fatal("expected has_rate=false when no rate is set")
	}
	if resp.CoreHourlyRateMinor != 0 {
		t.Fatalf("expected core_hourly_rate_minor=0, got %d", resp.CoreHourlyRateMinor)
	}
}

func TestBillingSetupPutValidRate(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `{"core_hourly_rate_minor":1500}`)
	requireStatus(t, w, http.StatusNoContent)
}

func TestBillingSetupPutZeroRate(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `{"core_hourly_rate_minor":0}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")

	w = doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `{"core_hourly_rate_minor":-100}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")
}

func TestBillingSetupPutInvalidJSON(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `not-json`)
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

func TestBillingSetupGetAfterPutReturnsRate(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `{"core_hourly_rate_minor":2500}`)
	requireStatus(t, w, http.StatusNoContent)

	w = doRequest(t, h.router, http.MethodGet, "/api/v1/billing-setup", h.managerToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp billingSetupResponse
	decodeJSON(t, w, &resp)
	if !resp.HasRate {
		t.Fatal("expected has_rate=true after setting rate")
	}
	if resp.CoreHourlyRateMinor != 2500 {
		t.Fatalf("expected core_hourly_rate_minor=2500, got %d", resp.CoreHourlyRateMinor)
	}
}

func TestBillingSetupAuditLogged(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	w := doRequest(t, h.router, http.MethodPut, "/api/v1/billing-setup", h.managerToken, `{"core_hourly_rate_minor":3000}`)
	requireStatus(t, w, http.StatusNoContent)

	var auditCount int
	err := h.pool.QueryRow(ctx,
		"SELECT count(*) FROM audit_logs WHERE tenant_id = $1 AND action_type = 'site_core_hourly_rate_updated'",
		h.tenantID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected 1 audit log entry, got %d", auditCount)
	}
}
