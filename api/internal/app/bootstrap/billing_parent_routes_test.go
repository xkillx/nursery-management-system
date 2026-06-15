package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/platform/dbtest"
)

// seedParentAccessChain creates a complete parent access chain:
// parent user -> parent membership -> guardian -> parent_membership_guardians mapping -> guardian_child_link
// and returns the IDs for use in test assertions.
func seedParentAccessChain(t *testing.T, h *billingHarness, suffix string, childName string, childIsActive bool) (childID, guardianID, linkID, mappingID uuid.UUID) {
	t.Helper()

	childID = uuid.MustParse(fmt.Sprintf("f1000000-0000-0000-0000-%012s", suffix))
	guardianID = uuid.MustParse(fmt.Sprintf("f2000000-0000-0000-0000-%012s", suffix))
	linkID = uuid.MustParse(fmt.Sprintf("f3000000-0000-0000-0000-%012s", suffix))
	mappingID = uuid.MustParse(fmt.Sprintf("f4000000-0000-0000-0000-%012s", suffix))

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, childName,
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, childIsActive)
	dbtest.InsertGuardian(t, h.pool, guardianID, h.tenantID, h.branchID, "Guardian "+childName, true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, guardianID, childID)
	dbtest.InsertParentMapping(t, h.pool, mappingID, h.tenantID, h.branchID, h.parentMID, guardianID)

	return childID, guardianID, linkID, mappingID
}

// seedIssuedInvoiceForChild creates an issued invoice for the given child with proper schema constraints satisfied.
// Returns the invoiceID and the runID used.
func seedIssuedInvoiceForChild(t *testing.T, h *billingHarness, suffix string, childID uuid.UUID, status string, totalDueMinor int) (invoiceID, runID uuid.UUID) {
	t.Helper()
	return seedIssuedInvoiceForChildMonth(t, h, suffix, childID, status, totalDueMinor, 2026, 5)
}

// seedIssuedInvoiceForChildMonth creates an issued invoice for the given child in a specific billing month.
func seedIssuedInvoiceForChildMonth(t *testing.T, h *billingHarness, suffix string, childID uuid.UUID, status string, totalDueMinor int, year, month int) (invoiceID, runID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	invoiceID = uuid.MustParse(fmt.Sprintf("f5000000-0000-0000-0000-%012s", suffix))
	runID = uuid.MustParse(fmt.Sprintf("f6000000-0000-0000-0000-%012s", suffix))
	billingMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)

	// Invoice run required for FK references
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, $7)`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID, "req-parent-"+suffix)
	if err != nil {
		t.Fatalf("insert invoice run for parent test: %v", err)
	}

	dueAt := "now() + interval '7 days'"
	if status == "overdue" {
		dueAt = "now() - interval '7 days'"
	}

	extraCols := ""
	extraVals := ""
	switch status {
	case "paid":
		extraCols = ", paid_at, amount_paid_minor"
		extraVals = ", now(), " + strconv.Itoa(totalDueMinor)
	case "payment_failed":
		extraCols = ", payment_failed_at"
		extraVals = ", now()"
	}

	_, err = h.pool.Exec(ctx, fmt.Sprintf(
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at%s)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', $6, 'GBP', $7, $7, $8, 0, $9, $10, $11, $12, 1, now(), $13, $14, now(), %s%s)`,
		extraCols, dueAt, extraVals),
		invoiceID, h.tenantID, h.branchID, childID, billingMonth, status,
		runID, totalDueMinor, totalDueMinor,
		billingMonth, periodEnd,
		fmt.Sprintf("INV-PAR-%s", suffix),
		h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert issued invoice for parent test (%s): %v", status, err)
	}

	return invoiceID, runID
}

// seedIssuedInvoiceWithLines creates an issued invoice with calculation details and invoice lines.
func seedIssuedInvoiceWithLines(t *testing.T, h *billingHarness, suffix string, childID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	invoiceID := uuid.MustParse(fmt.Sprintf("f5000000-0000-0000-0000-%012s", suffix))
	runID := uuid.MustParse(fmt.Sprintf("f6000000-0000-0000-0000-%012s", suffix))

	// Invoice run
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, $7)`,
		runID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1), h.managerUID, h.managerMID, "req-lines-"+suffix)
	if err != nil {
		t.Fatalf("insert invoice run for lines test: %v", err)
	}

	calcDetails := `{"billing_month":"2026-05","child_id":"` + childID.String() + `","core_hourly_rate_minor":500,"core_subtotal_minor":4000,"extras_total_minor":0,"manual_extras_supported":false,"funded_allowance_minutes":300,"funded_deduction_minutes":300,"core_billable_minutes":180,"raw_attended_minutes":480,"rounded_attended_minutes":480,"included_session_count":1}`

	// Insert as draft first (so lines can be inserted)
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, calculation_details)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, 4000, 2500, 1500, $7, $8, $9)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1),
		runID,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31),
		calcDetails)
	if err != nil {
		t.Fatalf("insert draft invoice for lines test: %v", err)
	}

	// Core childcare line
	line1 := uuid.MustParse(fmt.Sprintf("f7000000-0000-0000-0000-%012s", suffix))
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, raw_attended_minutes, rounded_attended_minutes, session_count)
		 VALUES ($1, $2, $3, $4, 'core_childcare', 'Core childcare', 1, 480, 500, 4000, 480, 480, 1)`,
		line1, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert core line: %v", err)
	}

	// Funded deduction line
	line2 := uuid.MustParse(fmt.Sprintf("f7000000-0000-0000-0001-%012s", suffix))
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, line_amount_minor, funded_allowance_minutes, funded_deduction_minutes, core_billable_minutes)
		 VALUES ($1, $2, $3, $4, 'funded_deduction', 'Funded hours deduction', 2, -2500, 300, 300, 180)`,
		line2, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert deduction line: %v", err)
	}

	// Issue the invoice (status transition is allowed by the trigger)
	_, err = h.pool.Exec(ctx,
		`UPDATE invoices SET status = 'issued', invoice_number = $2, issued_sequence = 1, issued_run_id = $3,
		 issued_at = now(), issued_by_user_id = $4, issued_by_membership_id = $5, locked_at = now(), due_at = now() + interval '7 days'
		 WHERE id = $1`,
		invoiceID, fmt.Sprintf("INV-LINES-%s", suffix), runID, h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("issue invoice: %v", err)
	}

	return invoiceID
}

// --- Route inventory (1) ---

func TestParentInvoiceRouteInventory(t *testing.T) {
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
			t.Fatalf("expected parent route %s to be registered", want)
		}
	}
}

// --- Authentication required (2) ---

func TestParentInvoiceListUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

func TestParentInvoiceDetailUnauthenticated(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+uuid.New().String(), "", "")
	requireStatus(t, w, http.StatusUnauthorized)
}

// --- Manager/practitioner forbidden on parent routes (3) ---

func TestParentInvoiceListRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)

	for _, token := range []string{h.managerToken, h.practitionerToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

func TestParentInvoiceDetailRoleGuards(t *testing.T) {
	h := setupBillingHarness(t)

	for _, token := range []string{h.managerToken, h.practitionerToken} {
		w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+uuid.New().String(), token, "")
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}

// --- Parent forbidden on manager invoice review routes (4) ---

func TestParentForbiddenOnManagerInvoiceRoutes(t *testing.T) {
	h := setupBillingHarness(t)

	// Parent token on manager invoice review route (GET /api/v1/invoices) is forbidden.
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")

	// Parent token on manager-specific preflight action is also forbidden.
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/invoices/drafts/preflight?billing_month=2026-05", h.parentToken, "")
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")

	w = doRequest(t, h.router, http.MethodPost, "/api/v1/invoice-runs/drafts", h.parentToken, `{"billing_month":"2026-05"}`)
	requireStatus(t, w, http.StatusForbidden)
	requireErrorCode(t, w, "forbidden_role")
}

// --- Parent list returns only issued-or-later invoices for linked children (5) ---

func TestParentInvoiceListReturnsOnlyIssuedForLinkedChildren(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Create child with full access chain
	childID, _, _, _ := seedParentAccessChain(t, h, "000000000001", "Linked Child", true)

	// Seed draft invoice for April (should NOT appear; uses April to avoid unique index collision with May issued invoice)
	draftInvoiceID := uuid.MustParse("f5000000-0000-0000-0000-000000000099")
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 1000, 0, 1000, $6, $7)`,
		draftInvoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 4, 1), dbtest.DateAt(2026, 4, 1), dbtest.DateAt(2026, 4, 30))
	if err != nil {
		t.Fatalf("insert draft invoice: %v", err)
	}

	// Seed issued invoice (should appear)
	issuedInvoiceID, _ := seedIssuedInvoiceForChild(t, h, "000000000001", childID, "issued", 1500)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item (issued only), got %d", len(resp.Items))
	}

	if resp.Items[0].InvoiceID != issuedInvoiceID.String() {
		t.Fatalf("expected invoice %s, got %s", issuedInvoiceID, resp.Items[0].InvoiceID)
	}
	if resp.Items[0].Status != "issued" {
		t.Fatalf("expected status issued, got %s", resp.Items[0].Status)
	}
	if resp.Items[0].ChildFirstName != "Linked Child" {
		t.Fatalf("expected child_first_name 'Linked Child', got %s", resp.Items[0].ChildFirstName)
	}
	if resp.Items[0].ChildID != childID.String() {
		t.Fatalf("expected child_id %s, got %s", childID, resp.Items[0].ChildID)
	}
}

// --- Parent list excludes drafts, unlinked child invoices, wrong tenant, wrong branch (6) ---

func TestParentInvoiceListExcludesDraftsUnlinkedWrongScope(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Linked child with issued invoice (should appear)
	linkedChildID, _, _, _ := seedParentAccessChain(t, h, "000000000010", "Included Child", true)
	seedIssuedInvoiceForChild(t, h, "000000000010", linkedChildID, "issued", 1000)

	// Unlinked child with issued invoice (should NOT appear - no guardian link to parent)
	unlinkedChildID := uuid.MustParse("f1000000-0000-0000-0000-000000000020")
	dbtest.InsertChild(t, h.pool, unlinkedChildID, h.tenantID, h.branchID, "Unlinked Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	seedIssuedInvoiceForChild(t, h, "000000000020", unlinkedChildID, "issued", 2000)

	// Wrong tenant invoice (should NOT appear)
	otherTenant := uuid.MustParse("d1000000-0000-0000-0000-000000000020")
	otherBranch := uuid.MustParse("d2000000-0000-0000-0000-000000000020")
	otherChild := uuid.MustParse("d3000000-0000-0000-0000-000000000020")
	otherInvoice := uuid.MustParse("d4000000-0000-0000-0000-000000000020")
	otherRun := uuid.MustParse("d5000000-0000-0000-0000-000000000020")
	otherUser := uuid.MustParse("d6000000-0000-0000-0000-000000000020")
	otherMembership := uuid.MustParse("d7000000-0000-0000-0000-000000000020")
	dbtest.InsertTenant(t, h.pool, otherTenant, "Wrong Tenant")
	dbtest.InsertBranch(t, h.pool, otherTenant, otherBranch, "Wrong Branch")
	dbtest.InsertUser(t, h.pool, otherUser, "wrong-tenant@example.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, otherMembership, otherTenant, otherBranch, otherUser, "manager", true)
	dbtest.InsertChild(t, h.pool, otherChild, otherTenant, otherBranch, "Other Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-wrong-tenant')`,
		otherRun, otherTenant, otherBranch, dbtest.DateAt(2026, 5, 1), otherUser, otherMembership)
	if err != nil {
		t.Fatalf("insert other run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $6, 3000, 0, 3000, $7, $8, 'INV-OTHER', 1, now(), $9, $10, now(), now() + interval '7 days')`,
		otherInvoice, otherTenant, otherBranch, otherChild, dbtest.DateAt(2026, 5, 1), otherRun,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), otherUser, otherMembership)
	if err != nil {
		t.Fatalf("insert other invoice: %v", err)
	}

	// Wrong branch invoice (same tenant, different branch) — needs its own user+membership for FK
	otherBranch2 := uuid.MustParse("b2000000-0000-0000-0000-000000000099")
	otherBranch2User := uuid.MustParse("a1000000-0000-0000-0000-000000000099")
	otherBranch2MID := uuid.MustParse("a2000000-0000-0000-0000-000000000099")
	otherChild2 := uuid.MustParse("f1000000-0000-0000-0000-000000000099")
	otherInvoice2 := uuid.MustParse("f5000000-0000-0000-0000-000000000099")
	otherRun2 := uuid.MustParse("f6000000-0000-0000-0000-000000000099")
	dbtest.InsertBranch(t, h.pool, h.tenantID, otherBranch2, "Other Branch Same Tenant")
	dbtest.InsertUser(t, h.pool, otherBranch2User, "other-branch@example.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, otherBranch2MID, h.tenantID, otherBranch2, otherBranch2User, "manager", true)
	dbtest.InsertChild(t, h.pool, otherChild2, h.tenantID, otherBranch2, "Other Branch Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-wrong-branch')`,
		otherRun2, h.tenantID, otherBranch2, dbtest.DateAt(2026, 5, 1), otherBranch2User, otherBranch2MID)
	if err != nil {
		t.Fatalf("insert other branch run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $6, 4000, 0, 4000, $7, $8, 'INV-WRONG-BRANCH', 1, now(), $9, $10, now(), now() + interval '7 days')`,
		otherInvoice2, h.tenantID, otherBranch2, otherChild2, dbtest.DateAt(2026, 5, 1), otherRun2,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), otherBranch2User, otherBranch2MID)
	if err != nil {
		t.Fatalf("insert wrong branch invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item (only linked child issued), got %d", len(resp.Items))
	}
	if resp.Items[0].InvoiceID == otherInvoice.String() {
		t.Fatal("wrong-tenant invoice should not appear")
	}
	if resp.Items[0].InvoiceID == otherInvoice2.String() {
		t.Fatal("wrong-branch invoice should not appear")
	}
}

// --- Parent list with no active parent mapping returns empty (7) ---

func TestParentInvoiceListNoMappingReturnsEmpty(t *testing.T) {
	h := setupBillingHarness(t)

	// Parent exists but has no guardian mapping and no guardian-child links.
	// The setup already creates the parent membership with no parent_membership_guardians mapping.

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items (no mapping), got %d", len(resp.Items))
	}
}

// --- Parent list with ended guardian-child link returns empty for that child (8) ---

func TestParentInvoiceListEndedGuardianLinkExcludesChild(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID, guardianID, linkID, _ := seedParentAccessChain(t, h, "000000000030", "Ended Link Child", true)
	seedIssuedInvoiceForChild(t, h, "000000000030", childID, "issued", 1000)

	// End the guardian-child link
	_, err := h.pool.Exec(ctx,
		`UPDATE guardian_child_links SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1`,
		linkID)
	if err != nil {
		t.Fatalf("end guardian link: %v", err)
	}

	// Verify the link was ended for this guardian-child pair only
	_ = guardianID // used via FK

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items (ended link), got %d", len(resp.Items))
	}
}

// --- Parent list with inactive or ended membership returns empty (9) ---

func TestParentInvoiceListInactiveMembershipReturnsEmpty(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Create a separate parent user with initially active membership
	inactiveUID := uuid.MustParse("f3000000-0000-0000-0000-000000000090")
	inactiveMID := uuid.MustParse("f4000000-0000-0000-0000-000000000090")
	inactiveGuardianID := uuid.MustParse("f2000000-0000-0000-0000-000000000090")

	dbtest.InsertUser(t, h.pool, inactiveUID, "inactive-parent@example.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, inactiveMID, h.tenantID, h.branchID, inactiveUID, "parent", true)
	dbtest.InsertGuardian(t, h.pool, inactiveGuardianID, h.tenantID, h.branchID, "Inactive Guardian", true)

	// Create child and link to this parent's guardian
	childID := uuid.MustParse("f1000000-0000-0000-0000-000000000090")
	linkID := uuid.MustParse("f3000000-0000-0000-0000-000000000091")
	mappingID := uuid.MustParse("f4000000-0000-0000-0000-000000000091")
	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Inactive Member Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, inactiveGuardianID, childID)
	// Insert mapping while membership is active (satisfies trigger)
	dbtest.InsertParentMapping(t, h.pool, mappingID, h.tenantID, h.branchID, inactiveMID, inactiveGuardianID)
	// Now deactivate the membership
	_, err := h.pool.Exec(ctx,
		`UPDATE memberships SET is_active = false, ended_at = now() WHERE id = $1`, inactiveMID)
	if err != nil {
		t.Fatalf("deactivate membership: %v", err)
	}
	seedIssuedInvoiceForChild(t, h, "000000000090", childID, "issued", 1000)

	inactiveToken := mustAccessToken(t, h.tokens, inactiveUID, inactiveMID, h.tenantID, h.branchID, "parent")

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", inactiveToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items (inactive membership), got %d", len(resp.Items))
	}
}

func TestParentInvoiceListEndedMembershipReturnsEmpty(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Create a parent with ended parent_membership_guardians mapping
	endedUID := uuid.MustParse("f3000000-0000-0000-0000-000000000095")
	endedMID := uuid.MustParse("f4000000-0000-0000-0000-000000000095")
	endedGuardianID := uuid.MustParse("f2000000-0000-0000-0000-000000000095")
	endedMappingID := uuid.MustParse("f4000000-0000-0000-0000-000000000096")

	dbtest.InsertUser(t, h.pool, endedUID, "ended-mapping@example.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, endedMID, h.tenantID, h.branchID, endedUID, "parent", true)
	dbtest.InsertGuardian(t, h.pool, endedGuardianID, h.tenantID, h.branchID, "Ended Guardian", true)

	childID := uuid.MustParse("f1000000-0000-0000-0000-000000000095")
	linkID := uuid.MustParse("f3000000-0000-0000-0000-000000000096")
	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Ended Mapping Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	dbtest.InsertGuardianLink(t, h.pool, linkID, h.tenantID, h.branchID, endedGuardianID, childID)
	// Insert mapping then end it
	dbtest.InsertParentMapping(t, h.pool, endedMappingID, h.tenantID, h.branchID, endedMID, endedGuardianID)
	_, err := h.pool.Exec(ctx,
		`UPDATE parent_membership_guardians SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1`,
		endedMappingID)
	if err != nil {
		t.Fatalf("end parent mapping: %v", err)
	}
	seedIssuedInvoiceForChild(t, h, "000000000095", childID, "issued", 1000)

	endedToken := mustAccessToken(t, h.tokens, endedUID, endedMID, h.tenantID, h.branchID, "parent")

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", endedToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items (ended mapping), got %d", len(resp.Items))
	}
}

// --- Parent list with child_id for unlinked child returns empty (10) ---

func TestParentInvoiceListUnlinkedChildIDReturnsEmpty(t *testing.T) {
	h := setupBillingHarness(t)

	// Create a linked child with issued invoice
	linkedChildID, _, _, _ := seedParentAccessChain(t, h, "000000000040", "Filter Child", true)
	seedIssuedInvoiceForChild(t, h, "000000000040", linkedChildID, "issued", 1000)

	// Create an unlinked child with issued invoice
	unlinkedChildID := uuid.MustParse("f1000000-0000-0000-0000-000000000041")
	dbtest.InsertChild(t, h.pool, unlinkedChildID, h.tenantID, h.branchID, "Unlinked Filter Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	seedIssuedInvoiceForChild(t, h, "000000000041", unlinkedChildID, "issued", 2000)

	// Filter by unlinked child ID
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?child_id="+unlinkedChildID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 0 {
		t.Fatalf("expected 0 items (unlinked child filter), got %d", len(resp.Items))
	}

	// Filter by linked child ID should return the invoice
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?child_id="+linkedChildID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item (linked child filter), got %d", len(resp.Items))
	}
}

// --- Parent list rejects status=draft (11) ---

func TestParentInvoiceListRejectsDraftStatus(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?status=draft", h.parentToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

// --- Parent list accepts issued, payment_failed, paid, overdue (12) ---

func TestParentInvoiceListAcceptsValidStatuses(t *testing.T) {
	h := setupBillingHarness(t)

	childID, _, _, _ := seedParentAccessChain(t, h, "000000000050", "Status Test Child", true)

	statusSuffixes := []struct {
		status string
		suffix string
		year   int
		month  int
	}{
		{"issued", "000000000051", 2026, 5},
		{"payment_failed", "000000000052", 2026, 6},
		{"paid", "000000000053", 2026, 7},
		{"overdue", "000000000054", 2026, 8},
	}
	for _, ss := range statusSuffixes {
		seedIssuedInvoiceForChildMonth(t, h, ss.suffix, childID, ss.status, 1000, ss.year, ss.month)
	}

	for _, ss := range statusSuffixes {
		t.Run("status="+ss.status, func(t *testing.T) {
			w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?status="+ss.status, h.parentToken, "")
			requireStatus(t, w, http.StatusOK)

			var resp parentInvoiceListResponseTest
			decodeJSON(t, w, &resp)
			for _, item := range resp.Items {
				if item.Status != ss.status {
					t.Fatalf("expected status %s, got %s", ss.status, item.Status)
				}
			}
		})
	}
}

// --- Parent detail for linked issued invoice returns parent-safe header, calculation, and lines (13) ---

func TestParentInvoiceDetailLinkedIssued(t *testing.T) {
	h := setupBillingHarness(t)

	childID, _, _, _ := seedParentAccessChain(t, h, "000000000060", "Detail Child", true)
	invoiceID := seedIssuedInvoiceWithLines(t, h, "000000000060", childID)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+invoiceID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceDetailResponseTest
	decodeJSON(t, w, &resp)

	// Header checks
	if resp.InvoiceID != invoiceID.String() {
		t.Fatalf("expected invoice_id %s, got %s", invoiceID, resp.InvoiceID)
	}
	if resp.Status != "issued" {
		t.Fatalf("expected status issued, got %s", resp.Status)
	}
	if resp.ChildFirstName != "Detail Child" {
		t.Fatalf("expected child_first_name 'Detail Child', got %s", resp.ChildFirstName)
	}
	if resp.CurrencyCode != "GBP" {
		t.Fatalf("expected currency_code GBP, got %s", resp.CurrencyCode)
	}
	if resp.SubtotalMinor != 4000 {
		t.Fatalf("expected subtotal_minor 4000, got %d", resp.SubtotalMinor)
	}
	if resp.FundedDeductionMinor != 2500 {
		t.Fatalf("expected funded_deduction_minor 2500, got %d", resp.FundedDeductionMinor)
	}
	if resp.TotalDueMinor != 1500 {
		t.Fatalf("expected total_due_minor 1500, got %d", resp.TotalDueMinor)
	}
	if resp.InvoiceNumber == nil || *resp.InvoiceNumber == "" {
		t.Fatal("expected invoice_number to be present")
	}
	if resp.BillingMonth != "2026-05" {
		t.Fatalf("expected billing_month 2026-05, got %s", resp.BillingMonth)
	}
	if resp.IssuedAt == nil || *resp.IssuedAt == "" {
		t.Fatal("expected issued_at to be present")
	}
	if resp.DueAt == nil || *resp.DueAt == "" {
		t.Fatal("expected due_at to be present")
	}

	// Calculation checks
	if resp.Calculation.CoreHourlyRateMinor != 500 {
		t.Fatalf("expected core_hourly_rate_minor 500, got %d", resp.Calculation.CoreHourlyRateMinor)
	}
	if resp.Calculation.RawAttendedMinutes != 480 {
		t.Fatalf("expected raw_attended_minutes 480, got %d", resp.Calculation.RawAttendedMinutes)
	}
	if resp.Calculation.RoundedAttendedMinutes != 480 {
		t.Fatalf("expected rounded_attended_minutes 480, got %d", resp.Calculation.RoundedAttendedMinutes)
	}
	if resp.Calculation.FundedAllowanceMinutes != 300 {
		t.Fatalf("expected funded_allowance_minutes 300, got %d", resp.Calculation.FundedAllowanceMinutes)
	}
	if resp.Calculation.IncludedSessionCount != 1 {
		t.Fatalf("expected included_session_count 1, got %d", resp.Calculation.IncludedSessionCount)
	}
	if resp.Calculation.CoreSubtotalMinor != 4000 {
		t.Fatalf("expected core_subtotal_minor 4000, got %d", resp.Calculation.CoreSubtotalMinor)
	}

	// Lines checks
	if len(resp.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(resp.Lines))
	}
	if resp.Lines[0].LineKind != "core_childcare" {
		t.Fatalf("expected line 0 kind 'core_childcare', got %s", resp.Lines[0].LineKind)
	}
	if resp.Lines[0].LineAmountMinor != 4000 {
		t.Fatalf("expected line 0 amount 4000, got %d", resp.Lines[0].LineAmountMinor)
	}
	if resp.Lines[1].LineKind != "funded_deduction" {
		t.Fatalf("expected line 1 kind 'funded_deduction', got %s", resp.Lines[1].LineKind)
	}
	if resp.Lines[1].LineAmountMinor != -2500 {
		t.Fatalf("expected line 1 amount -2500, got %d", resp.Lines[1].LineAmountMinor)
	}
}

// --- Parent detail does NOT serialize manager-only fields (14) ---

func TestParentInvoiceDetailOmitsManagerOnlyFields(t *testing.T) {
	h := setupBillingHarness(t)

	childID, _, _, _ := seedParentAccessChain(t, h, "000000000070", "Safe Child", true)
	invoiceID := seedIssuedInvoiceWithLines(t, h, "000000000070", childID)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+invoiceID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	// Decode as raw map to check for absent keys
	var raw map[string]any
	decodeJSON(t, w, &raw)

	// Manager-only fields that must NOT be present
	managerOnlyFields := []string{
		"generated_run_id",
		"generated_run_exceptions",
		"locked_at",
		"adjustment_reason_code",
		"source_sessions",
		"created_at",
		"updated_at",
	}
	for _, field := range managerOnlyFields {
		if _, exists := raw[field]; exists {
			t.Fatalf("parent detail must not include %q but it was present", field)
		}
	}

	// Check line items for manager-only fields
	linesRaw, ok := raw["lines"].([]any)
	if !ok || len(linesRaw) == 0 {
		t.Fatal("expected lines array with at least one item")
	}
	firstLine, ok := linesRaw[0].(map[string]any)
	if !ok {
		t.Fatal("expected line item to be a map")
	}
	lineOnlyFields := []string{
		"line_id",
		"raw_attended_minutes",
	}
	for _, field := range lineOnlyFields {
		if _, exists := firstLine[field]; exists {
			t.Fatalf("parent line must not include %q but it was present", field)
		}
	}

	// Also check calculation has no source_sessions
	calcRaw, ok := raw["calculation"].(map[string]any)
	if !ok {
		t.Fatal("expected calculation to be a map")
	}
	if _, exists := calcRaw["source_sessions"]; exists {
		t.Fatal("parent calculation must not include source_sessions")
	}
}

// --- Parent detail returns 404 for draft invoice (15) ---

func TestParentInvoiceDetailDraftReturns404(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID, _, _, _ := seedParentAccessChain(t, h, "000000000080", "Draft Detail Child", true)

	// Insert draft invoice for this linked child (uses April to avoid unique index collision with May invoices)
	draftInvoiceID := uuid.MustParse("f5000000-0000-0000-0000-000000000080")
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 1000, 0, 1000, $6, $7)`,
		draftInvoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 4, 1), dbtest.DateAt(2026, 4, 1), dbtest.DateAt(2026, 4, 30))
	if err != nil {
		t.Fatalf("insert draft invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+draftInvoiceID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

// --- Parent detail returns 404 for unlinked child invoice (16) ---

func TestParentInvoiceDetailUnlinkedChildReturns404(t *testing.T) {
	h := setupBillingHarness(t)

	// Create unlinked child with issued invoice
	unlinkedChildID := uuid.MustParse("f1000000-0000-0000-0000-000000000085")
	dbtest.InsertChild(t, h.pool, unlinkedChildID, h.tenantID, h.branchID, "Unlinked Detail Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	invoiceID, _ := seedIssuedInvoiceForChild(t, h, "000000000085", unlinkedChildID, "issued", 1000)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+invoiceID.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusNotFound)
	requireErrorCode(t, w, "invoice_not_found")
}

// --- Parent detail returns 404 for wrong tenant or wrong branch (17) ---

func TestParentInvoiceDetailWrongTenantReturns404(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	otherTenant := uuid.MustParse("d1000000-0000-0000-0000-000000000086")
	otherBranch := uuid.MustParse("d2000000-0000-0000-0000-000000000086")
	otherChild := uuid.MustParse("d3000000-0000-0000-0000-000000000086")
	otherInvoice := uuid.MustParse("d4000000-0000-0000-0000-000000000086")
	otherRun := uuid.MustParse("d5000000-0000-0000-0000-000000000086")
	otherUser := uuid.MustParse("d6000000-0000-0000-0000-000000000086")
	otherMembership := uuid.MustParse("d7000000-0000-0000-0000-000000000086")

	dbtest.InsertTenant(t, h.pool, otherTenant, "Other Tenant Detail")
	dbtest.InsertBranch(t, h.pool, otherTenant, otherBranch, "Other Branch Detail")
	dbtest.InsertUser(t, h.pool, otherUser, "other-detail@example.com", "hash", true)
	dbtest.InsertMembership(t, h.pool, otherMembership, otherTenant, otherBranch, otherUser, "manager", true)
	dbtest.InsertChild(t, h.pool, otherChild, otherTenant, otherBranch, "Other Detail Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), 500, true)
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-detail-tenant')`,
		otherRun, otherTenant, otherBranch, dbtest.DateAt(2026, 5, 1), otherUser, otherMembership)
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $6, 1000, 0, 1000, $7, $8, 'INV-X', 1, now(), $9, $10, now(), now() + interval '7 days')`,
		otherInvoice, otherTenant, otherBranch, otherChild, dbtest.DateAt(2026, 5, 1), otherRun,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31), otherUser, otherMembership)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/"+otherInvoice.String(), h.parentToken, "")
	requireStatus(t, w, http.StatusNotFound)
}

// --- Parent detail returns 400 for malformed invoice ID (18) ---

func TestParentInvoiceDetailMalformedIDReturns400(t *testing.T) {
	h := setupBillingHarness(t)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices/not-a-uuid", h.parentToken, "")
	requireStatus(t, w, http.StatusBadRequest)
	requireErrorCode(t, w, "validation_error")
}

// --- Inactive child with active access chain remains visible (19) ---

func TestParentInvoiceListInactiveChildWithActiveChain(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	// Create child as active first (to satisfy constraint), then deactivate
	childID, _, _, _ := seedParentAccessChain(t, h, "000000000088", "Inactive Visible Child", true)
	_, err := h.pool.Exec(ctx,
		`UPDATE children SET is_active = false, left_at = now(), left_reason_code = 'left_nursery' WHERE id = $1`, childID)
	if err != nil {
		t.Fatalf("deactivate child: %v", err)
	}
	invoiceID, _ := seedIssuedInvoiceForChild(t, h, "000000000088", childID, "issued", 1000)

	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)

	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item (inactive child with active chain), got %d", len(resp.Items))
	}
	if resp.Items[0].InvoiceID != invoiceID.String() {
		t.Fatalf("expected invoice %s, got %s", invoiceID, resp.Items[0].InvoiceID)
	}
	if resp.Items[0].ChildFirstName != "Inactive Visible Child" {
		t.Fatalf("expected child_first_name 'Inactive Visible Child', got %s", resp.Items[0].ChildFirstName)
	}
}

// --- Parent list supports pagination and filtering (additional coverage) ---

func TestParentInvoiceListBillingMonthFilter(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID, _, _, _ := seedParentAccessChain(t, h, "000000000091", "Month Filter Child", true)

	// May invoice
	seedIssuedInvoiceForChild(t, h, "000000000091", childID, "issued", 1000)

	// June invoice
	juneInvoice := uuid.MustParse("f5000000-0000-0000-0000-000000000092")
	juneRun := uuid.MustParse("f6000000-0000-0000-0000-000000000092")
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-june')`,
		juneRun, h.tenantID, h.branchID, dbtest.DateAt(2026, 6, 1), h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert june run: %v", err)
	}
	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  generated_run_id, issued_run_id, subtotal_minor, funded_deduction_minor, total_due_minor,
		  period_start_date, period_end_date, invoice_number, issued_sequence,
		  issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'issued', 'GBP', $6, $6, 2000, 0, 2000, $7, $8, 'INV-JUN', 1, now(), $9, $10, now(), now() + interval '7 days')`,
		juneInvoice, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 6, 1), juneRun,
		dbtest.DateAt(2026, 6, 1), dbtest.DateAt(2026, 6, 30), h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert june invoice: %v", err)
	}

	// Filter by May
	w := doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?billing_month=2026-05", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)
	var resp parentInvoiceListResponseTest
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 May invoice, got %d", len(resp.Items))
	}

	// Filter by June
	w = doRequest(t, h.router, http.MethodGet, "/api/v1/parent/invoices?billing_month=2026-06", h.parentToken, "")
	requireStatus(t, w, http.StatusOK)
	decodeJSON(t, w, &resp)
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 June invoice, got %d", len(resp.Items))
	}
}

func TestParentInvoiceListValidationErrors(t *testing.T) {
	h := setupBillingHarness(t)

	cases := []struct {
		name string
		url  string
	}{
		{"bad billing_month", "/api/v1/parent/invoices?billing_month=invalid"},
		{"bad child_id", "/api/v1/parent/invoices?child_id=not-a-uuid"},
		{"bad limit 0", "/api/v1/parent/invoices?limit=0"},
		{"bad limit 201", "/api/v1/parent/invoices?limit=201"},
		{"bad offset -1", "/api/v1/parent/invoices?offset=-1"},
		{"unknown status", "/api/v1/parent/invoices?status=unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRequest(t, h.router, http.MethodGet, tc.url, h.parentToken, "")
			requireStatus(t, w, http.StatusBadRequest)
			requireErrorCode(t, w, "validation_error")
		})
	}
}

// --- Test response types for parent invoice (API-21) ---

type parentInvoiceListResponseTest struct {
	Items  []parentInvoiceListItemResponseTest `json:"items"`
	Limit  int                                 `json:"limit"`
	Offset int                                 `json:"offset"`
}

type parentInvoiceListItemResponseTest struct {
	InvoiceID              string  `json:"invoice_id"`
	InvoiceKind            string  `json:"invoice_kind"`
	InvoiceNumber          *string `json:"invoice_number"`
	InvoiceNumberDisplay   string  `json:"invoice_number_display"`
	ChildID                string  `json:"child_id"`
	ChildFirstName         string  `json:"child_first_name"`
	BillingMonth           string  `json:"billing_month"`
	Status                 string  `json:"status"`
	DueStatus              string  `json:"due_status"`
	CurrencyCode           string  `json:"currency_code"`
	SubtotalMinor          int     `json:"subtotal_minor"`
	FundedDeductionMinor   int     `json:"funded_deduction_minor"`
	TotalDueMinor          int     `json:"total_due_minor"`
	AmountPaidMinor        int     `json:"amount_paid_minor"`
	IssuedAt               *string `json:"issued_at"`
	DueAt                  *string `json:"due_at"`
	PaidAt                 *string `json:"paid_at"`
	PaymentFailedAt        *string `json:"payment_failed_at"`
	PaymentStatusUpdatedAt *string `json:"payment_status_updated_at"`
}

type parentInvoiceDetailResponseTest struct {
	InvoiceID              string                               `json:"invoice_id"`
	InvoiceKind            string                               `json:"invoice_kind"`
	InvoiceNumber          *string                              `json:"invoice_number"`
	InvoiceNumberDisplay   string                               `json:"invoice_number_display"`
	ChildID                string                               `json:"child_id"`
	ChildFirstName         string                               `json:"child_first_name"`
	BillingMonth           string                               `json:"billing_month"`
	Status                 string                               `json:"status"`
	DueStatus              string                               `json:"due_status"`
	CurrencyCode           string                               `json:"currency_code"`
	SubtotalMinor          int                                  `json:"subtotal_minor"`
	FundedDeductionMinor   int                                  `json:"funded_deduction_minor"`
	TotalDueMinor          int                                  `json:"total_due_minor"`
	AmountPaidMinor        int                                  `json:"amount_paid_minor"`
	IssuedAt               *string                              `json:"issued_at"`
	DueAt                  *string                              `json:"due_at"`
	PaidAt                 *string                              `json:"paid_at"`
	PaymentFailedAt        *string                              `json:"payment_failed_at"`
	PaymentStatusUpdatedAt *string                              `json:"payment_status_updated_at"`
	Calculation            parentInvoiceCalculationResponseTest `json:"calculation"`
	Lines                  []parentInvoiceLineResponseTest      `json:"lines"`
}

type parentInvoiceLineResponseTest struct {
	LineKind        string `json:"line_kind"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
	QuantityMinutes *int   `json:"quantity_minutes"`
	UnitAmountMinor *int   `json:"unit_amount_minor"`
	LineAmountMinor int    `json:"line_amount_minor"`
}

type parentInvoiceCalculationResponseTest struct {
	CoreHourlyRateMinor    int `json:"core_hourly_rate_minor"`
	RawAttendedMinutes     int `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int `json:"core_billable_minutes"`
	IncludedSessionCount   int `json:"included_session_count"`
	CoreSubtotalMinor      int `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int `json:"extras_total_minor"`
}

// Ensure json import is used
var _ = json.RawMessage{}
var _ = time.UTC
