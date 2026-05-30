package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/dbtest"
	"nursery-management-system/api/internal/platform/uid"
)

type testHarness struct {
	pool       *pgxpool.Pool
	tenantID   uuid.UUID
	branchID   uuid.UUID
	parentUID  uuid.UUID
	parentMID  uuid.UUID
	childID    uuid.UUID
	guardianID uuid.UUID
	linkID     uuid.UUID
	mappingID  uuid.UUID
	managerUID uuid.UUID
	managerMID uuid.UUID
}

func setupTestHarness(t *testing.T) *testHarness {
	t.Helper()

	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)

	h := &testHarness{
		pool:       pool,
		tenantID:   uid.NewUUID(),
		branchID:   uid.NewUUID(),
		parentUID:  uid.NewUUID(),
		parentMID:  uid.NewUUID(),
		childID:    uid.NewUUID(),
		guardianID: uid.NewUUID(),
		linkID:     uid.NewUUID(),
		mappingID:  uid.NewUUID(),
		managerUID: uid.NewUUID(),
		managerMID: uid.NewUUID(),
	}

	dbtest.InsertTenant(t, pool, h.tenantID, "Payments Test")
	dbtest.InsertBranch(t, pool, h.tenantID, h.branchID, "Payments Branch")
	dbtest.InsertUser(t, pool, h.parentUID, "parent@payments.test", "hash", true)
	dbtest.InsertUser(t, pool, h.managerUID, "mgr@payments.test", "hash", true)
	dbtest.InsertMembership(t, pool, h.parentMID, h.tenantID, h.branchID, h.parentUID, "parent", true)
	dbtest.InsertMembership(t, pool, h.managerMID, h.tenantID, h.branchID, h.managerUID, "manager", true)
	dbtest.InsertChild(t, pool, h.childID, h.tenantID, h.branchID, "Test Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		500, true)
	dbtest.InsertGuardian(t, pool, h.guardianID, h.tenantID, h.branchID, "Test Guardian", true)
	dbtest.InsertGuardianLink(t, pool, h.linkID, h.tenantID, h.branchID, h.guardianID, h.childID)
	dbtest.InsertParentMapping(t, pool, h.mappingID, h.tenantID, h.branchID, h.parentMID, h.guardianID)

	return h
}

func seedIssuedInvoice(t *testing.T, h *testHarness, suffix string, status string, totalDueMinor int) uuid.UUID {
	t.Helper()
	return seedIssuedInvoiceForMonth(t, h, suffix, status, totalDueMinor, 2026, 5)
}

func seedIssuedInvoiceForMonth(t *testing.T, h *testHarness, suffix string, status string, totalDueMinor int, year, month int) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	invoiceID := uid.NewUUID()
	runID := uid.NewUUID()
	billingMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)

	_, err := h.pool.Exec(ctx,
		`INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, started_at, completed_at, requested_by_user_id, requested_by_membership_id, request_id)
		 VALUES ($1, $2, $3, $4, 'draft_generation', 'completed', now(), now(), $5, $6, $7)`,
		runID, h.tenantID, h.branchID, billingMonth, h.managerUID, h.managerMID, "req-"+suffix)
	if err != nil {
		t.Fatalf("insert invoice run: %v", err)
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
		extraVals = fmt.Sprintf(", now(), %d", totalDueMinor)
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
		invoiceID, h.tenantID, h.branchID, h.childID, billingMonth, status,
		runID, totalDueMinor, totalDueMinor,
		billingMonth, periodEnd,
		fmt.Sprintf("INV-PT-%s", suffix),
		h.managerUID, h.managerMID)
	if err != nil {
		t.Fatalf("insert invoice (%s): %v", status, err)
	}

	return invoiceID
}

func TestRepository_GetParentInvoiceForCheckout_IssuedVisible(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "issued", "issued", 5000)

	row, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if !found {
		t.Fatal("expected issued invoice to be found")
	}
	if row.Status != "issued" {
		t.Fatalf("expected issued, got %s", row.Status)
	}
	if row.TotalDueMinor != 5000 {
		t.Fatalf("expected 5000, got %d", row.TotalDueMinor)
	}
	if row.CurrencyCode != "GBP" {
		t.Fatalf("expected GBP, got %s", row.CurrencyCode)
	}
}

func TestRepository_GetParentInvoiceForCheckout_PaymentFailedVisible(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "pfailed", "payment_failed", 3000)

	row, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if !found {
		t.Fatal("expected payment_failed invoice to be found")
	}
	if row.Status != "payment_failed" {
		t.Fatalf("expected payment_failed, got %s", row.Status)
	}
}

func TestRepository_GetParentInvoiceForCheckout_OverdueVisible(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "overdue", "overdue", 7000)

	row, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if !found {
		t.Fatal("expected overdue invoice to be found")
	}
	if row.Status != "overdue" {
		t.Fatalf("expected overdue, got %s", row.Status)
	}
}

func TestRepository_GetParentInvoiceForCheckout_DraftNotFound(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	// Insert draft invoice directly
	ctx2 := context.Background()
	invoiceID := uid.NewUUID()
	_, err := h.pool.Exec(ctx2,
		`INSERT INTO invoices (id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status, currency_code,
		  subtotal_minor, funded_deduction_minor, total_due_minor, period_start_date, period_end_date)
		 VALUES ($1, $2, $3, $4, $5, 'monthly', 'draft', 'GBP', 5000, 0, 5000, $6, $7)`,
		invoiceID, h.tenantID, h.branchID, h.childID,
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert draft: %v", err)
	}

	_, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if found {
		t.Fatal("draft invoice should not be found for checkout")
	}
}

func TestRepository_GetParentInvoiceForCheckout_PaidNotFound(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "paid", "paid", 5000)

	_, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if found {
		t.Fatal("paid invoice should not be found for checkout")
	}
}

func TestRepository_GetParentInvoiceForCheckout_WrongParent(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "wrong", "issued", 5000)
	wrongMembership := uid.NewUUID().String()

	_, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), wrongMembership, invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if found {
		t.Fatal("invoice should not be visible to wrong parent")
	}
}

func TestRepository_GetParentInvoiceForCheckout_NoParentMapping(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	tx := dbtest.BeginTx(t, h.pool)
	ctx := context.Background()

	// End the parent mapping
	_, err := h.pool.Exec(context.Background(),
		"UPDATE parent_membership_guardians SET ended_at = now(), ended_reason_code = 'access_revoked' WHERE id = $1", h.mappingID)
	if err != nil {
		t.Fatalf("end mapping: %v", err)
	}

	invoiceID := seedIssuedInvoice(t, h, "nomap", "issued", 5000)

	_, found, err := repo.GetParentInvoiceForCheckoutForUpdate(ctx, tx, h.tenantID.String(), h.branchID.String(), h.parentMID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get for checkout: %v", err)
	}
	if found {
		t.Fatal("invoice should not be visible without active parent mapping")
	}
}

func TestRepository_CreatePaymentAttempt(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "create", "issued", 5000)

	tx := dbtest.BeginTx(t, h.pool)
	attemptID := uid.NewUUID().String()
	err := repo.CreatePaymentAttempt(ctx, tx, domain.PaymentAttemptCreateParams{
		ID:                      attemptID,
		TenantID:                h.tenantID.String(),
		BranchID:                h.branchID.String(),
		InvoiceID:               invoiceID.String(),
		InitiatedByUserID:       h.parentUID.String(),
		InitiatedByMembershipID: h.parentMID.String(),
		RequestID:               "req-test-1",
		Status:                  domain.AttemptStatusCheckoutCreationStarted,
		AmountMinor:             5000,
		CurrencyCode:            domain.CurrencyGBP,
	})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	dbtest.CommitTx(t, tx)

	// Verify row
	var status string
	err = h.pool.QueryRow(context.Background(),
		"SELECT status FROM payment_attempts WHERE id = $1", attemptID).Scan(&status)
	if err != nil {
		t.Fatalf("query attempt: %v", err)
	}
	if status != domain.AttemptStatusCheckoutCreationStarted {
		t.Fatalf("expected checkout_creation_started, got %s", status)
	}
}

var testMonthCounter = 5

func seedPaymentAttempt(t *testing.T, h *testHarness, repo *Repository, suffix string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	testMonthCounter++
	invoiceID := seedIssuedInvoiceForMonth(t, h, "attempt-"+suffix, "issued", 5000, 2026, testMonthCounter)
	attemptID := uid.NewUUID()

	tx := dbtest.BeginTx(t, h.pool)
	err := repo.CreatePaymentAttempt(ctx, tx, domain.PaymentAttemptCreateParams{
		ID:                      attemptID.String(),
		TenantID:                h.tenantID.String(),
		BranchID:                h.branchID.String(),
		InvoiceID:               invoiceID.String(),
		InitiatedByUserID:       h.parentUID.String(),
		InitiatedByMembershipID: h.parentMID.String(),
		RequestID:               "req-" + suffix,
		Status:                  domain.AttemptStatusCheckoutCreationStarted,
		AmountMinor:             5000,
		CurrencyCode:            domain.CurrencyGBP,
	})
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	dbtest.CommitTx(t, tx)

	return attemptID, invoiceID
}

func TestRepository_MarkPaymentAttemptCheckoutCreated(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	attemptID, _ := seedPaymentAttempt(t, h, repo, "created")

	expiresAt := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	err := repo.MarkPaymentAttemptCheckoutCreated(ctx, domain.PaymentAttemptCheckoutCreatedParams{
		TenantID:                h.tenantID.String(),
		BranchID:                h.branchID.String(),
		AttemptID:               attemptID.String(),
		StripeCheckoutSessionID: "cs_test_abc123",
		StripeCheckoutURL:       "https://checkout.stripe.com/test",
		StripePaymentIntentID:   "pi_test_xyz",
		StripeExpiresAt:         &expiresAt,
	})
	if err != nil {
		t.Fatalf("mark created: %v", err)
	}

	var status, sessionID string
	err = h.pool.QueryRow(ctx,
		"SELECT status, stripe_checkout_session_id FROM payment_attempts WHERE id = $1", attemptID).Scan(&status, &sessionID)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != domain.AttemptStatusCheckoutCreated {
		t.Fatalf("expected checkout_created, got %s", status)
	}
	if sessionID != "cs_test_abc123" {
		t.Fatalf("expected cs_test_abc123, got %s", sessionID)
	}
}

func TestRepository_MarkPaymentAttemptCheckoutCreationFailed(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	attemptID, _ := seedPaymentAttempt(t, h, repo, "failed")

	err := repo.MarkPaymentAttemptCheckoutCreationFailed(ctx, domain.PaymentAttemptCheckoutCreationFailedParams{
		TenantID:             h.tenantID.String(),
		BranchID:             h.branchID.String(),
		AttemptID:            attemptID.String(),
		FailureReason:        domain.FailureReasonStripeError,
		ProviderErrorCode:    "resource_missing",
		ProviderErrorMessage: "No such customer",
	})
	if err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	var status, reason string
	err = h.pool.QueryRow(ctx,
		"SELECT status, failure_reason FROM payment_attempts WHERE id = $1", attemptID).Scan(&status, &reason)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if status != domain.AttemptStatusCheckoutCreationFailed {
		t.Fatalf("expected checkout_creation_failed, got %s", status)
	}
	if reason != domain.FailureReasonStripeError {
		t.Fatalf("expected stripe_error, got %s", reason)
	}
}

func TestRepository_DuplicateStripeSessionIDRejected(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	sessionID := "cs_test_dup_" + uid.NewUUID().String()

	a1, _ := seedPaymentAttempt(t, h, repo, "dup1")
	_ = repo.MarkPaymentAttemptCheckoutCreated(ctx, domain.PaymentAttemptCheckoutCreatedParams{
		TenantID:                h.tenantID.String(),
		BranchID:                h.branchID.String(),
		AttemptID:               a1.String(),
		StripeCheckoutSessionID: sessionID,
		StripeCheckoutURL:       "https://checkout.stripe.com/1",
	})

	a2, _ := seedPaymentAttempt(t, h, repo, "dup2")
	err := repo.MarkPaymentAttemptCheckoutCreated(ctx, domain.PaymentAttemptCheckoutCreatedParams{
		TenantID:                h.tenantID.String(),
		BranchID:                h.branchID.String(),
		AttemptID:               a2.String(),
		StripeCheckoutSessionID: sessionID,
		StripeCheckoutURL:       "https://checkout.stripe.com/2",
	})
	if err == nil {
		t.Fatal("expected duplicate session ID to be rejected")
	}
}

func TestRepository_GetInvoicePaymentState(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	invoiceID := seedIssuedInvoice(t, h, "state", "issued", 5000)

	state, found, err := repo.GetInvoicePaymentState(ctx, h.tenantID.String(), h.branchID.String(), invoiceID.String())
	if err != nil {
		t.Fatalf("get payment state: %v", err)
	}
	if !found {
		t.Fatal("expected invoice state to be found")
	}
	if state.Status != "issued" {
		t.Fatalf("expected issued, got %s", state.Status)
	}
	if state.TotalDueMinor != 5000 {
		t.Fatalf("expected 5000, got %d", state.TotalDueMinor)
	}
	if state.CurrencyCode != "GBP" {
		t.Fatalf("expected GBP, got %s", state.CurrencyCode)
	}
}

func TestRepository_GetInvoicePaymentState_NotFound(t *testing.T) {
	h := setupTestHarness(t)
	repo := NewRepository(h.pool)
	ctx := context.Background()

	_, found, err := repo.GetInvoicePaymentState(ctx, h.tenantID.String(), h.branchID.String(), uid.NewUUID().String())
	if err != nil {
		t.Fatalf("get payment state: %v", err)
	}
	if found {
		t.Fatal("expected not found for nonexistent invoice")
	}
}
