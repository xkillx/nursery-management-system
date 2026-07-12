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

type stubReminderRepo struct {
	lockAcquired bool
	lockErr      error
	dueSoon      []domain.InvoiceReminderRow
	dueSoonErr   error
	dueToday     []domain.InvoiceReminderRow
	dueTodayErr  error
	logEntries   []reminderLogEntry
}

type reminderLogEntry struct {
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	InvoiceID    uuid.UUID
	ReminderType string
}

func (s *stubReminderRepo) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	return s.lockAcquired, s.lockErr
}

func (s *stubReminderRepo) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	return s.dueSoon, s.dueSoonErr
}

func (s *stubReminderRepo) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	return s.dueToday, s.dueTodayErr
}

func (s *stubReminderRepo) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, tenantID, branchID, invoiceID uuid.UUID, reminderType string) error {
	s.logEntries = append(s.logEntries, reminderLogEntry{
		TenantID:     tenantID,
		BranchID:     branchID,
		InvoiceID:    invoiceID,
		ReminderType: reminderType,
	})
	return nil
}

// Satisfy the rest of BillingRepository with panics (unused in these tests).

func (s *stubReminderRepo) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListActiveTermsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListActiveTerms(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListBookingPatternEntries(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.BookingPatternEntryRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("stub")
}
func (s *stubReminderRepo) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("stub")
}
func (s *stubReminderRepo) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("stub")
}
func (s *stubReminderRepo) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("stub")
}
func (s *stubReminderRepo) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("stub")
}
func (s *stubReminderRepo) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) InsertInvoiceLine(_ context.Context, _ domain.Tx, _ domain.InvoiceLineCreateParams) error {
	panic("stub")
}
func (s *stubReminderRepo) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("stub")
}
func (s *stubReminderRepo) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("stub")
}
func (s *stubReminderRepo) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("stub")
}
func (s *stubReminderRepo) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("stub")
}
func (s *stubReminderRepo) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("stub")
}
func (s *stubReminderRepo) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("stub")
}
func (s *stubReminderRepo) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	panic("stub")
}
func (s *stubReminderRepo) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("stub")
}

type stubReminderTxMgr struct {
	repo *stubReminderRepo
}

func (m *stubReminderTxMgr) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

// --- Tests ---

func TestSendDueSoonReminders_HappyPath(t *testing.T) {
	now := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	invoice1ID := uuid.New()
	invoice2ID := uuid.New()
	invoice3ID := uuid.New()

	repo := &stubReminderRepo{
		lockAcquired: true,
		dueSoon: []domain.InvoiceReminderRow{
			{ID: invoice1ID, TenantID: uuid.New(), BranchID: uuid.New(), DueDate: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
			{ID: invoice2ID, TenantID: uuid.New(), BranchID: uuid.New(), DueDate: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)},
		},
		dueToday: []domain.InvoiceReminderRow{
			{ID: invoice3ID, TenantID: uuid.New(), BranchID: uuid.New(), DueDate: time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)},
		},
	}
	uc := NewSendDueSoonReminders(repo, events.NewEventDispatcher(&stubReminderTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.LockAcquired {
		t.Fatal("expected lock acquired")
	}
	if len(result.DueSoon) != 2 {
		t.Fatalf("expected 2 due soon, got %d", len(result.DueSoon))
	}
	if len(result.DueToday) != 1 {
		t.Fatalf("expected 1 due today, got %d", len(result.DueToday))
	}
	if len(repo.logEntries) != 3 {
		t.Fatalf("expected 3 log entries, got %d", len(repo.logEntries))
	}
}

func TestSendDueSoonReminders_NoInvoices(t *testing.T) {
	now := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	repo := &stubReminderRepo{
		lockAcquired: true,
		dueSoon:      []domain.InvoiceReminderRow{},
		dueToday:     []domain.InvoiceReminderRow{},
	}
	uc := NewSendDueSoonReminders(repo, events.NewEventDispatcher(&stubReminderTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.LockAcquired {
		t.Fatal("expected lock acquired")
	}
	if len(result.DueSoon) != 0 {
		t.Fatalf("expected 0 due soon, got %d", len(result.DueSoon))
	}
	if len(result.DueToday) != 0 {
		t.Fatalf("expected 0 due today, got %d", len(result.DueToday))
	}
	if len(repo.logEntries) != 0 {
		t.Fatalf("expected 0 log entries, got %d", len(repo.logEntries))
	}
}

func TestSendDueSoonReminders_LockNotAcquired(t *testing.T) {
	now := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	repo := &stubReminderRepo{lockAcquired: false}
	uc := NewSendDueSoonReminders(repo, events.NewEventDispatcher(&stubReminderTxMgr{repo: repo}), func() time.Time { return now })

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LockAcquired {
		t.Fatal("expected lock not acquired")
	}
	if len(result.DueSoon) != 0 {
		t.Fatalf("expected 0 due soon when lock not acquired, got %d", len(result.DueSoon))
	}
}

func TestSendDueSoonReminders_RepositoryError(t *testing.T) {
	now := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	repo := &stubReminderRepo{
		lockAcquired: true,
		dueSoonErr:   context.DeadlineExceeded,
	}
	uc := NewSendDueSoonReminders(repo, events.NewEventDispatcher(&stubReminderTxMgr{repo: repo}), func() time.Time { return now })

	_, err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error from repository")
	}
}
