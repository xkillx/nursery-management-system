package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/events"
)

// --- Stubs ---

type stubBillingRepo struct {
	lockAcquired    bool
	lockErr         error
	overdueInvoices []domain.OverdueTransitionedInvoice
	overdueErr      error
}

func (s *stubBillingRepo) TryAcquireOverdueTransitionJobLock(ctx context.Context, tx domain.Tx) (bool, error) {
	return s.lockAcquired, s.lockErr
}

func (s *stubBillingRepo) MarkIssuedInvoicesOverdue(ctx context.Context, tx domain.Tx, cutoffUTC time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	return s.overdueInvoices, s.overdueErr
}

// Satisfy the rest of BillingRepository with panics (unused in these tests).

func (s *stubBillingRepo) ListActiveTermsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListActiveTerms(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListBookingPatternEntries(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.BookingPatternEntryRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("stub")
}
func (s *stubBillingRepo) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("stub")
}
func (s *stubBillingRepo) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("stub")
}
func (s *stubBillingRepo) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("stub")
}
func (s *stubBillingRepo) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("stub")
}
func (s *stubBillingRepo) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) InsertInvoiceLine(_ context.Context, _ domain.Tx, _ domain.InvoiceLineCreateParams) error {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("stub")
}

func (s *stubBillingRepo) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("stub")
}
func (s *stubBillingRepo) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("stub")
}
func (s *stubBillingRepo) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("stub")
}
func (s *stubBillingRepo) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("stub")
}
func (s *stubBillingRepo) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("stub")
}

func (s *stubBillingRepo) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("stub")
}
func (s *stubBillingRepo) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("stub")
}
func (s *stubBillingRepo) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	panic("stub")
}
func (s *stubBillingRepo) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("stub")
}
func (s *stubBillingRepo) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("stub")
}
func (s *stubBillingRepo) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("stub")
}
func (s *stubBillingRepo) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("stub")
}
func (s *stubBillingRepo) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("stub")
}

type stubTxMgr struct {
	repo *stubBillingRepo
}

func (m *stubTxMgr) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

// --- Tests ---

func TestMarkOverdue_LondonDateNotUTC(t *testing.T) {
	// 2026-07-15 01:30 UTC = 2026-07-15 02:30 BST (London)
	// London date = July 15, cutoff = 2026-07-15 00:00 London = 2026-07-14 23:00 UTC
	now := time.Date(2026, 7, 15, 1, 30, 0, 0, time.UTC)
	repo := &stubBillingRepo{lockAcquired: true}
	uc := NewMarkOverdueInvoices(repo, events.NewEventDispatcher(&stubTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	london := mustLoadLondon()
	wantCutoff := time.Date(2026, 7, 15, 0, 0, 0, 0, london).UTC()
	if !result.CutoffUTC.Equal(wantCutoff) {
		t.Fatalf("cutoff: got %v, want %v", result.CutoffUTC, wantCutoff)
	}
}

func TestMarkOverdue_BSTBoundary(t *testing.T) {
	// During BST (UTC+1): 2026-07-14 23:30 UTC = 2026-07-15 00:30 London
	// London date = July 15, cutoff = 2026-07-15 00:00 London = 2026-07-14 23:00 UTC
	now := time.Date(2026, 7, 14, 23, 30, 0, 0, time.UTC)
	repo := &stubBillingRepo{lockAcquired: true}
	uc := NewMarkOverdueInvoices(repo, events.NewEventDispatcher(&stubTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	london := mustLoadLondon()
	cutoff := time.Date(2026, 7, 15, 0, 0, 0, 0, london).UTC()
	if !result.CutoffUTC.Equal(cutoff) {
		t.Fatalf("BST boundary cutoff: got %v, want %v", result.CutoffUTC, cutoff)
	}
}

func TestMarkOverdue_GMTBoundary(t *testing.T) {
	// During GMT (UTC+0): 2026-01-15 00:30 UTC = 2026-01-15 00:30 London
	// London date = Jan 15, cutoff = 2026-01-15 00:00 UTC
	now := time.Date(2026, 1, 15, 0, 30, 0, 0, time.UTC)
	repo := &stubBillingRepo{lockAcquired: true}
	uc := NewMarkOverdueInvoices(repo, events.NewEventDispatcher(&stubTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cutoff := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !result.CutoffUTC.Equal(cutoff) {
		t.Fatalf("GMT boundary cutoff: got %v, want %v", result.CutoffUTC, cutoff)
	}
}

func TestMarkOverdue_LockNotAcquired(t *testing.T) {
	repo := &stubBillingRepo{lockAcquired: false}
	uc := NewMarkOverdueInvoices(repo, events.NewEventDispatcher(&stubTxMgr{repo: repo}), func() time.Time { return time.Now() })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LockAcquired {
		t.Fatal("expected LockAcquired false")
	}
	if len(result.Transitioned) != 0 {
		t.Fatal("expected no transitioned invoices when lock not acquired")
	}
}

func TestMarkOverdue_RepositoryError(t *testing.T) {
	repo := &stubBillingRepo{
		lockAcquired: true,
		overdueErr:   context.DeadlineExceeded,
	}
	uc := NewMarkOverdueInvoices(repo, events.NewEventDispatcher(&stubTxMgr{repo: repo}), func() time.Time { return time.Now() })

	_, err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestMarkOverdue_IdempotentDoubleRun(t *testing.T) {
	now := time.Date(2026, 6, 1, 2, 0, 0, 0, time.UTC)

	// First run returns invoices
	repo1 := &stubBillingRepo{
		lockAcquired:    true,
		overdueInvoices: []domain.OverdueTransitionedInvoice{{ID: uuid.New(), TenantID: uuid.New(), BranchID: uuid.New()}},
	}
	uc := NewMarkOverdueInvoices(repo1, events.NewEventDispatcher(&stubTxMgr{repo: repo1}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Transitioned) != 1 {
		t.Fatalf("first run: expected 1 transitioned, got %d", len(result.Transitioned))
	}

	// Second run returns nothing (SQL is idempotent)
	repo2 := &stubBillingRepo{lockAcquired: true}
	uc2 := NewMarkOverdueInvoices(repo2, events.NewEventDispatcher(&stubTxMgr{repo: repo2}), func() time.Time { return now })

	result2, err := uc2.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result2.Transitioned) != 0 {
		t.Fatalf("second run: expected 0 transitioned, got %d", len(result2.Transitioned))
	}
}

func mustLoadLondon() *time.Location {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(err)
	}
	return loc
}
