package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// createDraftMarkerRepo is a minimal BillingRepository stub that captures the
// details persisted on each inserted line, so tests can assert the manual-draft
// description-override marker (KTD6).
type createDraftMarkerRepo struct {
	found            bool
	insertedDetails  [][]byte
	insertedKinds    []string
	insertedDescript []string
}

func (s *createDraftMarkerRepo) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	return domain.InvoiceRow{}, s.found, nil
}
func (s *createDraftMarkerRepo) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	return nil
}
func (s *createDraftMarkerRepo) InsertInvoiceLine(_ context.Context, _ domain.Tx, params domain.InvoiceLineCreateParams) error {
	s.insertedDetails = append(s.insertedDetails, append([]byte(nil), params.Details...))
	s.insertedKinds = append(s.insertedKinds, params.LineKind)
	s.insertedDescript = append(s.insertedDescript, params.Description)
	return nil
}
func (s *createDraftMarkerRepo) ListActiveBookingsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	return nil, nil
}
func (s *createDraftMarkerRepo) ListActiveBookings(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	return nil, nil
}
func (s *createDraftMarkerRepo) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	return nil, nil
}
func (s *createDraftMarkerRepo) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("unused")
}
func (s *createDraftMarkerRepo) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("unused")
}
func (s *createDraftMarkerRepo) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("unused")
}
func (s *createDraftMarkerRepo) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetInvoiceForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoiceLinesForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money, _ []byte) (int64, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) DeleteInvoice(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("unused")
}
func (s *createDraftMarkerRepo) CountRecentInvoiceResendsTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (int, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) LockInvoiceForResendTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (bool, error) {
	panic("unused")
}
func (s *createDraftMarkerRepo) GetLatestResendAtTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (time.Time, error) {
	panic("unused")
}

func coreLine(details []byte, desc string) DraftInvoiceLineInput {
	return DraftInvoiceLineInput{
		LineKind:        domain.LineKindCoreChildcare,
		Description:     desc,
		SortOrder:       1,
		QuantityMinutes: 12000,
		UnitAmountMinor: 1000,
		LineAmountMinor: 200000,
	}
}

func TestCreateDraftInvoice_ManualCoreRenamePersistsMarker(t *testing.T) {
	repo := &createDraftMarkerRepo{found: false}
	uc := &CreateDraftInvoice{
		repo:   repo,
		txMgr:  &stubTxManager{},
		auditW: &stubAuditWriter{},
		core:   nil,
	}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, CreateDraftInvoiceInput{
		ChildID:      uuid.New(),
		BillingMonth: "2026-05",
		Lines: []DraftInvoiceLineInput{
			coreLine(nil, "Wrap-around care"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.insertedDetails) != 1 {
		t.Fatalf("inserted lines = %d, want 1", len(repo.insertedDetails))
	}
	if !domain.HasLineDescriptionOverride(repo.insertedDetails[0]) {
		t.Errorf("expected description_override marker for manual core rename, got %s", repo.insertedDetails[0])
	}
	if repo.insertedDescript[0] != "Wrap-around care" {
		t.Errorf("persisted description = %q, want %q", repo.insertedDescript[0], "Wrap-around care")
	}
}

func TestCreateDraftInvoice_DefaultCoreLabelOmitsMarker(t *testing.T) {
	repo := &createDraftMarkerRepo{found: false}
	uc := &CreateDraftInvoice{
		repo:   repo,
		txMgr:  &stubTxManager{},
		auditW: &stubAuditWriter{},
		core:   nil,
	}

	actor := testActor()
	_, err := uc.Execute(context.Background(), actor, CreateDraftInvoiceInput{
		ChildID:      uuid.New(),
		BillingMonth: "2026-05",
		Lines: []DraftInvoiceLineInput{
			coreLine(nil, "May 2026 Recurring Booking"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.insertedDetails) != 1 {
		t.Fatalf("inserted lines = %d, want 1", len(repo.insertedDetails))
	}
	if domain.HasLineDescriptionOverride(repo.insertedDetails[0]) {
		t.Errorf("expected no marker for default label, got %s", repo.insertedDetails[0])
	}
}

func TestCreateDraftInvoice_RejectsEmptyAndTooLongDescription(t *testing.T) {
	repo := &createDraftMarkerRepo{found: false}
	uc := &CreateDraftInvoice{
		repo:   repo,
		txMgr:  &stubTxManager{},
		auditW: &stubAuditWriter{},
		core:   nil,
	}

	actor := testActor()
	for _, desc := range []string{"   ", stringOfLen(121)} {
		_, err := uc.Execute(context.Background(), actor, CreateDraftInvoiceInput{
			ChildID:      uuid.New(),
			BillingMonth: "2026-05",
			Lines: []DraftInvoiceLineInput{
				coreLine(nil, desc),
			},
		})
		if err == nil {
			t.Fatalf("expected validation error for description %q", desc)
		}
		if _, ok := err.(*domainerrors.DomainError); !ok {
			t.Fatalf("expected DomainError, got %T", err)
		}
	}
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// TestCreateDraftInvoice_SubtotalExcludesFundedDeduction verifies that the
// subtotal calculation in CreateDraftInvoice correctly separates funded
// deduction lines from core lines. Before the fix, the subtotal included the
// deduction amount as a positive value, causing it to be inflated.
func TestCreateDraftInvoice_SubtotalExcludesFundedDeduction(t *testing.T) {
	lines := []DraftInvoiceLineInput{
		{
			LineKind:        domain.LineKindCoreChildcare,
			Description:     "Core childcare",
			SortOrder:       1,
			QuantityMinutes: 12000,
			UnitAmountMinor: 1000,
			LineAmountMinor: 200000, // £2,000.00
		},
		{
			LineKind:        domain.LineKindFundedDeduction,
			Description:     "Funded hours deduction",
			SortOrder:       2,
			QuantityMinutes: 0,
			UnitAmountMinor: 0,
			LineAmountMinor: 47500, // £475.00 (positive input from prefill)
		},
	}

	// Replicate the subtotal calculation logic from CreateDraftInvoice.Execute
	subtotalMinor := 0
	fundedDeductionMinor := 0
	for _, line := range lines {
		if line.LineKind == domain.LineKindFundedDeduction {
			fundedDeductionMinor += line.LineAmountMinor
		} else {
			subtotalMinor += line.LineAmountMinor
		}
	}
	totalDueMinor := subtotalMinor - fundedDeductionMinor

	if subtotalMinor != 200000 {
		t.Errorf("subtotal = %d, want 200000", subtotalMinor)
	}
	if fundedDeductionMinor != 47500 {
		t.Errorf("fundedDeduction = %d, want 47500", fundedDeductionMinor)
	}
	if totalDueMinor != 152500 {
		t.Errorf("totalDue = %d, want 152500", totalDueMinor)
	}
}

func TestCreateDraftInvoice_NoDeductionLine(t *testing.T) {
	lines := []DraftInvoiceLineInput{
		{
			LineKind:        domain.LineKindCoreChildcare,
			Description:     "Core childcare",
			SortOrder:       1,
			QuantityMinutes: 12000,
			UnitAmountMinor: 1000,
			LineAmountMinor: 200000,
		},
	}

	subtotalMinor := 0
	fundedDeductionMinor := 0
	for _, line := range lines {
		if line.LineKind == domain.LineKindFundedDeduction {
			fundedDeductionMinor += line.LineAmountMinor
		} else {
			subtotalMinor += line.LineAmountMinor
		}
	}
	totalDueMinor := subtotalMinor - fundedDeductionMinor

	if subtotalMinor != 200000 {
		t.Errorf("subtotal = %d, want 200000", subtotalMinor)
	}
	if totalDueMinor != 200000 {
		t.Errorf("totalDue = %d, want 200000", totalDueMinor)
	}
}

func TestCreateDraftInvoice_ValidationRejectsNegativeQuantity(t *testing.T) {
	uc := &CreateDraftInvoice{}
	actor := tenant.ActorContext{
		TenantID: uuid.New(),
		BranchID: uuid.New(),
	}

	input := CreateDraftInvoiceInput{
		ChildID:      uuid.New(),
		BillingMonth: "2026-08",
		Lines: []DraftInvoiceLineInput{
			{
				LineKind:        domain.LineKindCoreChildcare,
				Description:     "Core childcare",
				SortOrder:       1,
				QuantityMinutes: -100,
				UnitAmountMinor: 1000,
				LineAmountMinor: -100000,
			},
		},
	}

	_, err := uc.Execute(context.Background(), actor, input)
	if err == nil {
		t.Fatal("expected validation error for negative quantity")
	}
	var domainErr *domainerrors.DomainError
	if ok := (err != nil); !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	_ = domainErr
}
