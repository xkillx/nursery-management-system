package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/platform/dbtest"
	"nursery-management-system/api/internal/platform/uid"
)

func TestOverdueTransition_IssuedUnpaidBeforeCutoffBecomesOverdue(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	dueAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC) // before cutoff

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertIssuedInvoiceRow(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, dueAt)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)

	acquired, err := repo.TryAcquireOverdueTransitionJobLock(ctx, tx)
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	if !acquired {
		t.Fatal("expected lock acquired")
	}

	rows, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].ID != invoiceID {
		t.Fatalf("invoice id: got %v, want %v", rows[0].ID, invoiceID)
	}

	status := getInvoiceStatus(t, pool, invoiceID)
	if status != "overdue" {
		t.Fatalf("status: got %q, want %q", status, "overdue")
	}
}

func TestOverdueTransition_IssuedUnpaidAtOrAfterCutoffStaysIssued(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	dueAt := cutoff // exactly at cutoff — should NOT transition (< not <=)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertIssuedInvoiceRow(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, dueAt)

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)
	rows, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}

	status := getInvoiceStatus(t, pool, invoiceID)
	if status != "issued" {
		t.Fatalf("status: got %q, want %q", status, "issued")
	}
}

func TestOverdueTransition_ZeroTotalIssuedStaysIssued(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	dueAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertIssuedInvoiceRow(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 0, 0, dueAt) // zero total

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)
	rows, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows for zero-total, got %d", len(rows))
	}
	status := getInvoiceStatus(t, pool, invoiceID)
	if status != "issued" {
		t.Fatalf("status: got %q, want %q", status, "issued")
	}
}

func TestOverdueTransition_PaymentFailedStaysPaymentFailed(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertInvoice(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, "payment_failed", time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC))

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)
	rows, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
	status := getInvoiceStatus(t, pool, invoiceID)
	if status != "payment_failed" {
		t.Fatalf("status: got %q, want %q", status, "payment_failed")
	}
}

func TestOverdueTransition_PaidStaysPaid(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertInvoice(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 2000, "paid", time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC))

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)
	rows, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
	status := getInvoiceStatus(t, pool, invoiceID)
	if status != "paid" {
		t.Fatalf("status: got %q, want %q", status, "paid")
	}
}

func TestOverdueTransition_IdempotentDoubleRun(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	dueAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertIssuedInvoiceRow(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, dueAt)

	repo := NewRepository(pool)

	// First run
	tx1 := dbtest.BeginTx(t, pool)
	rows1, err := repo.MarkIssuedInvoicesOverdue(ctx, tx1, cutoff)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	dbtest.CommitTx(t, tx1)
	if len(rows1) != 1 {
		t.Fatalf("first run: expected 1, got %d", len(rows1))
	}

	// Second run
	tx2 := dbtest.BeginTx(t, pool)
	rows2, err := repo.MarkIssuedInvoicesOverdue(ctx, tx2, cutoff)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	dbtest.CommitTx(t, tx2)
	if len(rows2) != 0 {
		t.Fatalf("second run: expected 0, got %d", len(rows2))
	}
}

func TestOverdueTransition_TimestampsUpdated(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	cutoff := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	dueAt := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertIssuedInvoiceRow(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, dueAt)

	before := time.Now().UTC()

	repo := NewRepository(pool)
	tx := dbtest.BeginTx(t, pool)
	_, err := repo.MarkIssuedInvoicesOverdue(ctx, tx, cutoff)
	if err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	dbtest.CommitTx(t, tx)

	psu, updated := getPaymentTimestamps(t, pool, invoiceID)
	if psu.Before(before) {
		t.Fatalf("payment_status_updated_at %v should be >= %v", psu, before)
	}
	if updated.Before(before) {
		t.Fatalf("updated_at %v should be >= %v", updated, before)
	}
}

func TestOverdueTransition_TriggerBlocksPaymentFailedToOverdue(t *testing.T) {
	pool := dbtest.RequirePostgres(t)
	dbtest.Reset(t, pool)
	ctx := context.Background()

	tenantID, branchID, userID, membershipID := seedTenantBranchUserMembership(t, pool)
	childID := seedChild(t, pool, tenantID, branchID)
	runID := uid.NewUUID()
	invoiceID := uid.NewUUID()

	insertInvoiceRun(t, pool, runID, tenantID, branchID, userID, membershipID)
	insertInvoice(t, pool, invoiceID, tenantID, branchID, childID, userID, membershipID, runID, 2000, 0, "payment_failed", time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC))

	// Direct SQL attempt — trigger should block payment_failed → overdue
	_, err := pool.Exec(ctx, `UPDATE invoices SET status = 'overdue', updated_at = now() WHERE id = $1`, invoiceID)
	if err == nil {
		t.Fatal("expected trigger to block payment_failed → overdue")
	}
}

// --- Helpers ---

func seedTenantBranchUserMembership(t testing.TB, pool *pgxpool.Pool) (tenantID, branchID, userID, membershipID uuid.UUID) {
	t.Helper()
	tenantID = uid.NewUUID()
	branchID = uid.NewUUID()
	userID = uid.NewUUID()
	membershipID = uid.NewUUID()

	dbtest.InsertTenant(t, pool, tenantID, "Overdue Test Nursery")
	dbtest.InsertBranch(t, pool, tenantID, branchID, "Main Branch")
	dbtest.InsertUser(t, pool, userID, "manager@test.local", "$2a$10$hash", true)
	dbtest.InsertMembership(t, pool, membershipID, tenantID, branchID, userID, "manager", true)
	return
}

func seedChild(t testing.TB, pool *pgxpool.Pool, tenantID, branchID uuid.UUID) uuid.UUID {
	t.Helper()
	childID := uid.NewUUID()
	dbtest.InsertChild(t, pool, childID, tenantID, branchID, "Test Child",
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		500, true)
	return childID
}

func insertInvoiceRun(t testing.TB, pool *pgxpool.Pool, runID, tenantID, branchID, userID, membershipID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `INSERT INTO invoice_runs (id, tenant_id, branch_id, billing_month, run_type, status, requested_by_user_id, requested_by_membership_id, request_id, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())`,
		runID, tenantID, branchID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "issue", "completed", userID, membershipID, "test-req")
	if err != nil {
		t.Fatalf("insert invoice run: %v", err)
	}
}

func insertIssuedInvoiceRow(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, childID, userID, membershipID, runID uuid.UUID, totalDueMinor, amountPaidMinor int, dueAt time.Time) {
	t.Helper()
	ctx := context.Background()
	issuedAt := dueAt
	_, err := pool.Exec(ctx, `INSERT INTO invoices (
		id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		period_start_date, period_end_date, calculation_details
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,
		$8,$9,$10,$11,
		$12,$13,$14,$15,$16,
		$17,$18,$19,$20,$21,
		$22,$23,$24
	)`,
		id, tenantID, branchID, childID,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "monthly", "issued",
		"INV-202606-0001", 1, runID, runID,
		issuedAt, userID, membershipID, issuedAt, dueAt,
		"GBP", totalDueMinor, 0, totalDueMinor, amountPaidMinor,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "{}",
	)
	if err != nil {
		t.Fatalf("insert issued invoice: %v", err)
	}
}

func insertInvoice(t testing.TB, pool *pgxpool.Pool, id, tenantID, branchID, childID, userID, membershipID, runID uuid.UUID, totalDueMinor, amountPaidMinor int, status string, dueAt time.Time) {
	t.Helper()
	ctx := context.Background()
	issuedAt := dueAt
	paymentFailedAt := time.Time{}
	paidAt := time.Time{}
	paymentStatusUpdatedAt := time.Time{}
	if status == "payment_failed" {
		paymentFailedAt = time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
		paymentStatusUpdatedAt = paymentFailedAt
	} else if status == "paid" {
		paidAt = time.Date(2026, 6, 12, 10, 0, 0, 0, time.UTC)
		paymentStatusUpdatedAt = paidAt
	}
	_, err := pool.Exec(ctx, `INSERT INTO invoices (
		id, tenant_id, branch_id, child_id, billing_month, invoice_kind, status,
		invoice_number, issued_sequence, generated_run_id, issued_run_id,
		issued_at, issued_by_user_id, issued_by_membership_id, locked_at, due_at,
		currency_code, subtotal_minor, funded_deduction_minor, total_due_minor, amount_paid_minor,
		paid_at, payment_failed_at, payment_status_updated_at,
		period_start_date, period_end_date, calculation_details
	) VALUES (
		$1,$2,$3,$4,$5,$6,$7,
		$8,$9,$10,$11,
		$12,$13,$14,$15,$16,
		$17,$18,$19,$20,$21,
		$22,$23,$24,
		$25,$26,$27
	)`,
		id, tenantID, branchID, childID,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), "monthly", status,
		"INV-202606-0002", 2, runID, runID,
		issuedAt, userID, membershipID, issuedAt, dueAt,
		"GBP", totalDueMinor, 0, totalDueMinor, amountPaidMinor,
		paidAt, paymentFailedAt, paymentStatusUpdatedAt,
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "{}",
	)
	if err != nil {
		t.Fatalf("insert invoice (%s): %v", status, err)
	}
}

func getInvoiceStatus(t testing.TB, pool *pgxpool.Pool, id uuid.UUID) string {
	t.Helper()
	ctx := context.Background()
	var status string
	err := pool.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1`, id).Scan(&status)
	if err != nil {
		t.Fatalf("get invoice status: %v", err)
	}
	return status
}

func getPaymentTimestamps(t testing.TB, pool *pgxpool.Pool, id uuid.UUID) (paymentStatusUpdatedAt, updatedAt time.Time) {
	t.Helper()
	ctx := context.Background()
	err := pool.QueryRow(ctx, `SELECT payment_status_updated_at, updated_at FROM invoices WHERE id = $1`, id).Scan(&paymentStatusUpdatedAt, &updatedAt)
	if err != nil {
		t.Fatalf("get payment timestamps: %v", err)
	}
	return
}
