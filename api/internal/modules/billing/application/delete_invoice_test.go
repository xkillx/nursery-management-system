package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/events"
)

// --- Stub for DeleteInvoice tests ---

type deleteInvoiceRepoStub struct {
	deleteAffectedRows int64
	deleteErr          error
	invoiceFound       bool
}

func (s *deleteInvoiceRepoStub) DeleteInvoice(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	if s.deleteErr != nil {
		return 0, s.deleteErr
	}
	return s.deleteAffectedRows, nil
}

func (s *deleteInvoiceRepoStub) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return domain.InvoiceReviewRow{}, s.invoiceFound, nil
}

// Unused methods required by BillingRepository interface
func (s *deleteInvoiceRepoStub) ListActiveBookingsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListActiveBookings(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) InsertInvoiceLine(_ context.Context, _ domain.Tx, _ domain.InvoiceLineCreateParams) error {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *deleteInvoiceRepoStub) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("unused")
}

// --- Helper to create dispatcher with stub tx manager ---

func newTestDispatcher() *events.EventDispatcher {
	return events.NewEventDispatcher(&stubTxManager{})
}

// --- Tests ---

func TestDeleteInvoice_InvalidID(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "validation_error" {
		t.Errorf("expected validation_error, got %s", de.Code)
	}
}

func TestDeleteInvoice_HappyPath_Draft(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1, invoiceFound: false}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	result, err := uc.Execute(context.Background(), actor, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "deleted" {
		t.Errorf("expected status deleted, got %s", result.Status)
	}
	if result.InvoiceID == uuid.Nil {
		t.Error("expected non-nil invoice ID")
	}
}

func TestDeleteInvoice_HappyPath_Void(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1, invoiceFound: false}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	result, err := uc.Execute(context.Background(), actor, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "deleted" {
		t.Errorf("expected status deleted, got %s", result.Status)
	}
}

func TestDeleteInvoice_NotFound(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 0, invoiceFound: false}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for not found invoice")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_not_found" {
		t.Errorf("expected invoice_not_found, got %s", de.Code)
	}
}

func TestDeleteInvoice_IneligibleStatus(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 0, invoiceFound: true}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for ineligible status")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_not_deletable" {
		t.Errorf("expected invoice_not_deletable, got %s", de.Code)
	}
}

func TestDeleteInvoice_InfraError(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteErr: context.DeadlineExceeded}
	uc := &DeleteInvoice{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for infra failure")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "internal_error" {
		t.Errorf("expected internal_error, got %s", de.Code)
	}
}

// --- BulkDeleteInvoices Tests ---

func TestBulkDeleteInvoices_EmptyList(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1}
	uc := &BulkDeleteInvoices{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, []string{})
	if err == nil {
		t.Fatal("expected error for empty list")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "validation_error" {
		t.Errorf("expected validation_error, got %s", de.Code)
	}
}

func TestBulkDeleteInvoices_InvalidID(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1}
	uc := &BulkDeleteInvoices{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, []string{"not-a-uuid"})
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "validation_error" {
		t.Errorf("expected validation_error, got %s", de.Code)
	}
}

func TestBulkDeleteInvoices_HappyPath_AllDeleted(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1, invoiceFound: false}
	uc := &BulkDeleteInvoices{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	ids := []string{uuid.New().String(), uuid.New().String(), uuid.New().String()}
	result, err := uc.Execute(context.Background(), actor, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 3 {
		t.Errorf("expected 3 deleted, got %d", len(result.Deleted))
	}
	if len(result.Blocked) != 0 {
		t.Errorf("expected 0 blocked, got %d", len(result.Blocked))
	}
}

func TestBulkDeleteInvoices_MixedResults(t *testing.T) {
	// First call returns 1 (deleted), second returns 0 (blocked).
	callCount := 0
	repo := &deleteInvoiceRepoStub{invoiceFound: false}
	repo.deleteAffectedRows = 1 // Default

	// We need a custom stub that returns different values per call.
	uc := &BulkDeleteInvoices{repo: &mixedDeleteRepoStub{callCount: &callCount}, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	ids := []string{uuid.New().String(), uuid.New().String()}
	result, err := uc.Execute(context.Background(), actor, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 1 {
		t.Errorf("expected 1 deleted, got %d", len(result.Deleted))
	}
	if len(result.Blocked) != 1 {
		t.Errorf("expected 1 blocked, got %d", len(result.Blocked))
	}
	if result.Blocked[0].ErrorCode != "invoice_not_found" {
		t.Errorf("expected invoice_not_found, got %s", result.Blocked[0].ErrorCode)
	}
}

func TestBulkDeleteInvoices_Deduplication(t *testing.T) {
	repo := &deleteInvoiceRepoStub{deleteAffectedRows: 1, invoiceFound: false}
	uc := &BulkDeleteInvoices{repo: repo, auditW: nil, dispatcher: newTestDispatcher()}

	actor := testActor()
	id := uuid.New().String()
	ids := []string{id, id, id} // Same ID 3 times
	result, err := uc.Execute(context.Background(), actor, ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deleted) != 1 {
		t.Errorf("expected 1 deleted (deduplicated), got %d", len(result.Deleted))
	}
}

// --- Mixed delete stub ---

type mixedDeleteRepoStub struct {
	callCount *int
}

func (s *mixedDeleteRepoStub) DeleteInvoice(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	*s.callCount++
	if *s.callCount == 1 {
		return 1, nil // First call: deleted
	}
	return 0, nil // Second call: not found/ineligible
}

func (s *mixedDeleteRepoStub) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return domain.InvoiceReviewRow{}, false, nil // Not found
}

// Unused methods required by BillingRepository interface
func (s *mixedDeleteRepoStub) ListActiveBookingsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListActiveBookings(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) InsertInvoiceLine(_ context.Context, _ domain.Tx, _ domain.InvoiceLineCreateParams) error {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *mixedDeleteRepoStub) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("unused")
}
