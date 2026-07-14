package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// --- Minimal stub for ManageInvoiceLines tests ---

type manageLinesRepoStub struct {
	invoice     domain.InvoiceReviewRow
	invoiceOK   bool
	lines       []domain.InvoiceReviewLineRow
	line        domain.InvoiceLine
	lineOK      bool
	insertedIDs []uuid.UUID
	updatedIDs  []uuid.UUID
	deletedIDs  []uuid.UUID
	updateErr   error
	deleteErr   error
}

func (s *manageLinesRepoStub) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return s.invoice, s.invoiceOK, nil
}
func (s *manageLinesRepoStub) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	return s.lines, nil
}
func (s *manageLinesRepoStub) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	return s.line, s.lineOK, nil
}
func (s *manageLinesRepoStub) InsertInvoiceLine(_ context.Context, _ domain.Tx, params domain.InvoiceLineCreateParams) error {
	s.insertedIDs = append(s.insertedIDs, params.ID)
	return nil
}
func (s *manageLinesRepoStub) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	if s.updateErr != nil {
		return 0, s.updateErr
	}
	return 1, nil
}
func (s *manageLinesRepoStub) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	if s.deleteErr != nil {
		return 0, s.deleteErr
	}
	return 1, nil
}
func (s *manageLinesRepoStub) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	return nil
}

// Satisfy the rest of BillingRepository with panics (unused in these tests).
func (s *manageLinesRepoStub) ListActiveTermsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListActiveTerms(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListBookingPatternEntries(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.BookingPatternEntryRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("stub")
}
func (s *manageLinesRepoStub) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("stub")
}
func (s *manageLinesRepoStub) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("stub")
}
func (s *manageLinesRepoStub) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("stub")
}
func (s *manageLinesRepoStub) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("stub")
}
func (s *manageLinesRepoStub) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("stub")
}
func (s *manageLinesRepoStub) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("stub")
}

// --- Stub TxManager ---

type stubTxManager struct{}

func (s *stubTxManager) ExecTx(_ context.Context, fn func(pgx.Tx) error) error {
	return fn(nil)
}

// --- Stub AuditWriter ---

type stubAuditWriter struct{}

func (s *stubAuditWriter) WriteWithTx(_ context.Context, _ pgx.Tx, _ tenant.ActorContext, _ interface{}) error {
	return nil
}

// --- Tests ---

func TestAddLine_RejectsImmutableKind(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   draftInvoice(),
		invoiceOK: true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: nil, auditW: nil}

	actor := testActor()
	_, err := uc.AddLine(context.Background(), actor, uuid.New().String(), AddLineInput{
		LineKind:    domain.LineKindCoreChildcare,
		Description: "test",
	})
	if err == nil {
		t.Fatal("expected error for core_childcare line kind")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_line_kind_immutable" {
		t.Errorf("expected code invoice_line_kind_immutable, got %s", de.Code)
	}
}

func TestAddLine_RejectsNonDraftInvoice(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   issuedInvoice(),
		invoiceOK: true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: &stubTxManager{}, auditW: nil}

	actor := testActor()
	_, err := uc.AddLine(context.Background(), actor, uuid.New().String(), AddLineInput{
		LineKind:    domain.LineKindExtra,
		Description: "test",
	})
	if err == nil {
		t.Fatal("expected error for non-draft invoice")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_not_draft" {
		t.Errorf("expected code invoice_not_draft, got %s", de.Code)
	}
}

func TestAddLine_RejectsEmptyDescription(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   draftInvoice(),
		invoiceOK: true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: nil, auditW: nil}

	actor := testActor()
	_, err := uc.AddLine(context.Background(), actor, uuid.New().String(), AddLineInput{
		LineKind:    domain.LineKindExtra,
		Description: "",
	})
	if err == nil {
		t.Fatal("expected error for empty description")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "validation_error" {
		t.Errorf("expected code validation_error, got %s", de.Code)
	}
}

func TestUpdateLine_RejectsImmutableKind(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   draftInvoice(),
		invoiceOK: true,
		line:      systemLine(),
		lineOK:    true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: &stubTxManager{}, auditW: nil}

	actor := testActor()
	_, err := uc.UpdateLine(context.Background(), actor, uuid.New().String(), uuid.New().String(), UpdateLineInput{
		Description:     "updated",
		QuantityMinutes: 30,
		UnitAmountMinor: 1000,
		LineAmountMinor: 500,
	})
	if err == nil {
		t.Fatal("expected error for updating system line")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_line_kind_immutable" {
		t.Errorf("expected code invoice_line_kind_immutable, got %s", de.Code)
	}
}

func TestDeleteLine_RejectsImmutableKind(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   draftInvoice(),
		invoiceOK: true,
		line:      systemLine(),
		lineOK:    true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: &stubTxManager{}, auditW: nil}

	actor := testActor()
	_, err := uc.DeleteLine(context.Background(), actor, uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error for deleting system line")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_line_kind_immutable" {
		t.Errorf("expected code invoice_line_kind_immutable, got %s", de.Code)
	}
}

func TestDeleteLine_RejectsNonDraftInvoice(t *testing.T) {
	repo := &manageLinesRepoStub{
		invoice:   issuedInvoice(),
		invoiceOK: true,
		line:      extraLine(),
		lineOK:    true,
	}
	uc := &ManageInvoiceLines{repo: repo, txMgr: &stubTxManager{}, auditW: nil}

	actor := testActor()
	_, err := uc.DeleteLine(context.Background(), actor, uuid.New().String(), uuid.New().String())
	if err == nil {
		t.Fatal("expected error for non-draft invoice")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != "invoice_not_draft" {
		t.Errorf("expected code invoice_not_draft, got %s", de.Code)
	}
}

// --- Helpers ---

func testActor() tenant.ActorContext {
	return tenant.ActorContext{
		TenantID: uuid.New(),
		BranchID: uuid.New(),
	}
}

func draftInvoice() domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:       uuid.New(),
		Status:   domain.InvoiceStatusDraft,
		Subtotal: domain.MustGBP(5000),
		TotalDue: domain.MustGBP(5000),
	}
}

func issuedInvoice() domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:     uuid.New(),
		Status: domain.InvoiceStatusIssued,
	}
}

func systemLine() domain.InvoiceLine {
	return domain.InvoiceLine{
		ID:       uuid.New(),
		LineKind: domain.LineKindCoreChildcare,
	}
}

func extraLine() domain.InvoiceLine {
	return domain.InvoiceLine{
		ID:       uuid.New(),
		LineKind: domain.LineKindExtra,
	}
}
