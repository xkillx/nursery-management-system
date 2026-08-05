package bootstrap

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/dbtest"
)

// seedDraftInvoiceWithRun inserts a draft invoice referencing a draft_generation
// run so the manager line-update recalculate step's UpdateDraftInvoice write
// satisfies the generated_run_id foreign key.
func seedDraftInvoiceWithRun(t *testing.T, h *billingHarness, childID, invoiceID uuid.UUID, subtotalMinor int) {
	t.Helper()
	ctx := context.Background()
	runID := uuid.MustParse("e4000000-0000-0000-0000-000000000001")

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, 'req-line-update')`,
		runID, h.tenantID, h.branchID, dbtest.DateAt(2026, 5, 1),
		uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
		uuid.MustParse("b4000000-0000-0000-0000-000000000001"))
	if err != nil {
		t.Fatalf("insert run: %v", err)
	}

	_, err = h.pool.Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code, generated_run_id, subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date, calculation_details)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', $6, $7, 0, $7, $8, $9, '{}'::jsonb)`,
		invoiceID, h.tenantID, h.branchID, childID, dbtest.DateAt(2026, 5, 1), runID, subtotalMinor,
		dbtest.DateAt(2026, 5, 1), dbtest.DateAt(2026, 5, 31))
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}
}

func TestInvoiceLineUpdate_DescriptionOnlyOnCoreLine(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e1000000-0000-0000-0000-000000000001")
	invoiceID := uuid.MustParse("e2000000-0000-0000-0000-000000000001")
	lineID := uuid.MustParse("e3000000-0000-0000-0000-000000000001")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Line Child",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	seedDraftInvoiceWithRun(t, h, childID, invoiceID, 4000)

	coreDetails := `{"booked_core_minutes":480,"booked_sessions":[{"DayOfWeek":1,"OccurrenceDate":"2026-05-04T00:00:00Z","DurationMinutes":480,"SessionTypeID":"st1","SessionTypeName":"Full Day","StartMinutes":480,"EndMinutes":960,"SessionAmountMinor":4000}],"booked_per_entry":[]}`
	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, session_count, details)
		 VALUES ($1, $2, $3, $4, 'core_childcare', 'May 2026 Recurring Booking', 1, 480, 500, 4000, 1, $5::jsonb)`,
		lineID, h.tenantID, h.branchID, invoiceID, coreDetails)
	if err != nil {
		t.Fatalf("insert line: %v", err)
	}

	// Description-only update on a core line.
	w := doRequest(t, h.router, http.MethodPut,
		"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
		h.managerToken,
		`{"description":"Wrap-around care","quantity_minutes":480,"unit_amount_minor":500,"line_amount_minor":4000}`)
	requireStatus(t, w, http.StatusOK)

	var dbDesc string
	var dbDetails []byte
	err = h.pool.QueryRow(ctx,
		"SELECT description, details FROM invoice_lines WHERE id = $1", lineID).Scan(&dbDesc, &dbDetails)
	if err != nil {
		t.Fatalf("query line: %v", err)
	}
	if dbDesc != "Wrap-around care" {
		t.Fatalf("description = %q, want %q", dbDesc, "Wrap-around care")
	}
	if !domain.HasLineDescriptionOverride(dbDetails) {
		t.Fatalf("expected description_override marker, got %s", dbDetails)
	}
	if !bytes.Contains(dbDetails, []byte("booked_sessions")) {
		t.Fatalf("expected booked_sessions preserved, got %s", dbDetails)
	}

	// Invoice totals unchanged (description-only edit).
	var subtotal, totalDue int
	err = h.pool.QueryRow(ctx, "SELECT subtotal_minor, total_due_minor FROM invoices WHERE id = $1", invoiceID).Scan(&subtotal, &totalDue)
	if err != nil {
		t.Fatalf("query invoice totals: %v", err)
	}
	if subtotal != 4000 || totalDue != 4000 {
		t.Fatalf("totals changed: subtotal=%d total=%d, want 4000/4000", subtotal, totalDue)
	}
}

func TestInvoiceLineUpdate_ValueChangeOnCoreLineRejected(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e1000000-0000-0000-0000-000000000002")
	invoiceID := uuid.MustParse("e2000000-0000-0000-0000-000000000002")
	lineID := uuid.MustParse("e3000000-0000-0000-0000-000000000002")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Line Child 2",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	seedDraftInvoiceWithRun(t, h, childID, invoiceID, 4000)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, session_count, details)
		 VALUES ($1, $2, $3, $4, 'core_childcare', 'May 2026 Recurring Booking', 1, 480, 500, 4000, 1, '{}'::jsonb)`,
		lineID, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line: %v", err)
	}

	// Quantity change on a core line is rejected server-side.
	w := doRequest(t, h.router, http.MethodPut,
		"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
		h.managerToken,
		`{"description":"Wrap-around care","quantity_minutes":600,"unit_amount_minor":500,"line_amount_minor":5000}`)
	requireStatus(t, w, http.StatusConflict)
	requireErrorCode(t, w, "invoice_line_values_immutable")

	// Line unchanged.
	var dbDesc string
	err = h.pool.QueryRow(ctx, "SELECT description FROM invoice_lines WHERE id = $1", lineID).Scan(&dbDesc)
	if err != nil {
		t.Fatalf("query line: %v", err)
	}
	if dbDesc != "May 2026 Recurring Booking" {
		t.Fatalf("line description changed despite rejection: %q", dbDesc)
	}
}

func TestInvoiceLineUpdate_DescriptionOnExtraLineStillEditable(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e1000000-0000-0000-0000-000000000003")
	invoiceID := uuid.MustParse("e2000000-0000-0000-0000-000000000003")
	lineID := uuid.MustParse("e3000000-0000-0000-0000-000000000003")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Line Child 3",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	seedDraftInvoiceWithRun(t, h, childID, invoiceID, 500)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, details)
		 VALUES ($1, $2, $3, $4, 'extra', 'Late pick-up', 1, 60, 500, 500, '{}'::jsonb)`,
		lineID, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line: %v", err)
	}

	// Full editability on an extra line: description + value changes apply.
	w := doRequest(t, h.router, http.MethodPut,
		"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
		h.managerToken,
		`{"description":"Late pick-up (updated)","quantity_minutes":90,"unit_amount_minor":600,"line_amount_minor":900}`)
	requireStatus(t, w, http.StatusOK)

	var dbDesc string
	var dbQty, dbLine int
	err = h.pool.QueryRow(ctx,
		"SELECT description, quantity_minutes, line_amount_minor FROM invoice_lines WHERE id = $1", lineID).Scan(&dbDesc, &dbQty, &dbLine)
	if err != nil {
		t.Fatalf("query line: %v", err)
	}
	if dbDesc != "Late pick-up (updated)" || dbQty != 90 || dbLine != 900 {
		t.Fatalf("extra line not fully updated: desc=%q qty=%d line=%d", dbDesc, dbQty, dbLine)
	}
}

func TestInvoiceLineUpdate_ValidationRejectsEmptyAndTooLong(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e1000000-0000-0000-0000-000000000004")
	invoiceID := uuid.MustParse("e2000000-0000-0000-0000-000000000004")
	lineID := uuid.MustParse("e3000000-0000-0000-0000-000000000004")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Line Child 4",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	seedDraftInvoiceWithRun(t, h, childID, invoiceID, 500)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, details)
		 VALUES ($1, $2, $3, $4, 'extra', 'Late pick-up', 1, 60, 500, 500, '{}'::jsonb)`,
		lineID, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line: %v", err)
	}

	// Whitespace-only description.
	w := doRequest(t, h.router, http.MethodPut,
		"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
		h.managerToken,
		`{"description":"   ","quantity_minutes":60,"unit_amount_minor":500,"line_amount_minor":500}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")

	// Over-120-char description.
	longDesc := strings.Repeat("x", 121)
	w = doRequest(t, h.router, http.MethodPut,
		"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
		h.managerToken,
		`{"description":"`+longDesc+`","quantity_minutes":60,"unit_amount_minor":500,"line_amount_minor":500}`)
	requireStatus(t, w, http.StatusUnprocessableEntity)
	requireErrorCode(t, w, "validation_error")
}

func TestInvoiceLineUpdate_NonManagerRejected(t *testing.T) {
	h := setupBillingHarness(t)
	ctx := context.Background()

	childID := uuid.MustParse("e1000000-0000-0000-0000-000000000005")
	invoiceID := uuid.MustParse("e2000000-0000-0000-0000-000000000005")
	lineID := uuid.MustParse("e3000000-0000-0000-0000-000000000005")

	dbtest.InsertChild(t, h.pool, childID, h.tenantID, h.branchID, "Line Child 5",
		dbtest.DateAt(2023, 1, 1), dbtest.DateAt(2026, 1, 1), true)
	seedDraftInvoiceWithRun(t, h, childID, invoiceID, 500)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_lines (id, tenant_id, branch_id, invoice_id, line_kind, description, sort_order, quantity_minutes, unit_amount_minor, line_amount_minor, details)
		 VALUES ($1, $2, $3, $4, 'extra', 'Late pick-up', 1, 60, 500, 500, '{}'::jsonb)`,
		lineID, h.tenantID, h.branchID, invoiceID)
	if err != nil {
		t.Fatalf("insert line: %v", err)
	}

	// R10: renaming is manager-only; parent/practitioner are rejected.
	for _, token := range []string{h.practitionerToken, h.parentToken} {
		w := doRequest(t, h.router, http.MethodPut,
			"/api/v1/invoices/"+invoiceID.String()+"/lines/"+lineID.String(),
			token,
			`{"description":"Renamed","quantity_minutes":60,"unit_amount_minor":500,"line_amount_minor":500}`)
		requireStatus(t, w, http.StatusForbidden)
		requireErrorCode(t, w, "forbidden_role")
	}
}
