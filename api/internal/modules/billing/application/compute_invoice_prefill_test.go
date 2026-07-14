package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type stubPrefillRepo struct {
	terms      []domain.AdvancePayTermRow
	termsErr   error
	entries    []domain.BookingPatternEntryRow
	entriesErr error
}

func (s *stubPrefillRepo) ListActiveTermsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	return s.terms, s.termsErr
}

func (s *stubPrefillRepo) ListBookingPatternEntries(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.BookingPatternEntryRow, error) {
	return s.entries, s.entriesErr
}

func (s *stubPrefillRepo) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	return nil, nil
}

func (s *stubPrefillRepo) ListActiveTerms(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.AdvancePayTermRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	panic("unused")
}

func (s *stubPrefillRepo) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	panic("unused")
}

func (s *stubPrefillRepo) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	panic("unused")
}

func (s *stubPrefillRepo) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	panic("unused")
}

func (s *stubPrefillRepo) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	panic("unused")
}

func (s *stubPrefillRepo) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	panic("unused")
}

func (s *stubPrefillRepo) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) InsertInvoiceLine(_ context.Context, _ domain.Tx, _ domain.InvoiceLineCreateParams) error {
	panic("unused")
}

func (s *stubPrefillRepo) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	panic("unused")
}

func (s *stubPrefillRepo) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	panic("unused")
}

func (s *stubPrefillRepo) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	panic("unused")
}

func (s *stubPrefillRepo) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	panic("unused")
}

func (s *stubPrefillRepo) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	panic("unused")
}

func (s *stubPrefillRepo) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	panic("unused")
}

func (s *stubPrefillRepo) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	panic("unused")
}

func (s *stubPrefillRepo) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	panic("unused")
}

func (s *stubPrefillRepo) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	panic("unused")
}

func (s *stubPrefillRepo) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}

func (s *stubPrefillRepo) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	panic("unused")
}

func (s *stubPrefillRepo) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	panic("unused")
}
func (s *stubPrefillRepo) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *stubPrefillRepo) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	panic("unused")
}
func (s *stubPrefillRepo) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	panic("unused")
}
func (s *stubPrefillRepo) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	panic("unused")
}
func (s *stubPrefillRepo) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money) (int64, error) {
	panic("unused")
}
func (s *stubPrefillRepo) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	panic("unused")
}
func (s *stubPrefillRepo) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	panic("unused")
}
func (s *stubPrefillRepo) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	panic("unused")
}
func (s *stubPrefillRepo) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	panic("unused")
}
func (s *stubPrefillRepo) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	panic("unused")
}
func (s *stubPrefillRepo) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	panic("unused")
}

type stubPrefillTxMgr struct {
	repo *stubPrefillRepo
}

func (m *stubPrefillTxMgr) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func makeTerm() domain.AdvancePayTermRow {
	fpID := uuid.MustParse("22222222-2222-4222-8222-222222222002")
	allowance := 570
	lastName := "Doe"
	return domain.AdvancePayTermRow{
		TermID:                 uuid.MustParse("11111111-1111-4111-8111-111111111001"),
		TenantID:               uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		BranchID:               uuid.MustParse("00000000-0000-4000-8000-000000000002"),
		ChildID:                uuid.MustParse("33333333-3333-4333-8333-333333333003"),
		TermStartDate:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TermEndDate:            time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		BookingPatternID:       uuid.MustParse("44444444-4444-4444-8444-444444444004"),
		SiteHourlyRateMinor:    600,
		Status:                 "active",
		FirstName:              "Jane",
		LastName:               &lastName,
		DateOfBirth:            time.Date(2021, 5, 10, 0, 0, 0, 0, time.UTC),
		StartDate:              time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		HasParentCarerContact:  true,
		FundingProfileID:       &fpID,
		FundedAllowanceMinutes: &allowance,
	}
}

func TestComputeInvoicePrefill_HappyPath(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	termRow := makeTerm()
	entries := []domain.BookingPatternEntryRow{
		{DayOfWeek: 1, SessionTypeID: uuid.New(), SessionTypeName: "Full Day", StartMinutes: 480, EndMinutes: 1020},
		{DayOfWeek: 2, SessionTypeID: uuid.New(), SessionTypeName: "Full Day", StartMinutes: 480, EndMinutes: 1020},
	}

	repo := &stubPrefillRepo{
		terms:   []domain.AdvancePayTermRow{termRow},
		entries: entries,
	}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{
		TenantID: termRow.TenantID,
		BranchID: termRow.BranchID,
	}
	result, err := uc.Execute(context.Background(), actor, childID.String(), "2026-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ChildID != childID {
		t.Fatalf("child_id: got %v, want %v", result.ChildID, childID)
	}
	if result.ChildFirstName != "Jane" {
		t.Fatalf("first_name: got %q, want %q", result.ChildFirstName, "Jane")
	}
	if result.SubtotalMinor <= 0 {
		t.Fatal("expected positive subtotal")
	}
	if result.FundedDeductionMinor <= 0 {
		t.Fatal("expected positive funded deduction for a child with funding profile")
	}
	if result.TotalDueMinor <= 0 {
		t.Fatal("expected positive total due")
	}
	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines (core + deduction), got %d", len(result.Lines))
	}
	if result.Lines[0].LineKind != domain.LineKindCoreChildcare {
		t.Fatalf("line[0] kind: got %q, want %q", result.Lines[0].LineKind, domain.LineKindCoreChildcare)
	}
	if result.Lines[1].LineKind != domain.LineKindFundedDeduction {
		t.Fatalf("line[1] kind: got %q, want %q", result.Lines[1].LineKind, domain.LineKindFundedDeduction)
	}
	if result.FundingProfileID == nil {
		t.Fatal("expected funding profile ID")
	}
	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", result.Warnings)
	}
}

func TestComputeInvoicePrefill_MissingFundingProfile(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	termRow := makeTerm()
	termRow.FundingProfileID = nil
	termRow.FundedAllowanceMinutes = nil

	repo := &stubPrefillRepo{
		terms:   []domain.AdvancePayTermRow{termRow},
		entries: []domain.BookingPatternEntryRow{},
	}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{
		TenantID: termRow.TenantID,
		BranchID: termRow.BranchID,
	}
	result, err := uc.Execute(context.Background(), actor, childID.String(), "2026-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FundingProfileID != nil {
		t.Fatal("expected nil funding profile ID")
	}
	if result.FundedDeductionMinor != 0 {
		t.Fatalf("expected zero funded deduction, got %d", result.FundedDeductionMinor)
	}
	if len(result.Lines) != 1 {
		t.Fatalf("expected 1 line (core only, no deduction), got %d", len(result.Lines))
	}
	found := false
	for _, w := range result.Warnings {
		if w == "missing_funding_profile" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected missing_funding_profile warning")
	}
}

func TestComputeInvoicePrefill_ZeroAttendance(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	termRow := makeTerm()

	repo := &stubPrefillRepo{
		terms:   []domain.AdvancePayTermRow{termRow},
		entries: []domain.BookingPatternEntryRow{},
	}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{
		TenantID: termRow.TenantID,
		BranchID: termRow.BranchID,
	}
	result, err := uc.Execute(context.Background(), actor, childID.String(), "2026-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SubtotalMinor != 0 {
		t.Fatalf("expected zero subtotal, got %d", result.SubtotalMinor)
	}
	if result.TotalDueMinor != 0 {
		t.Fatalf("expected zero total due, got %d", result.TotalDueMinor)
	}
}

func TestComputeInvoicePrefill_ChildNotFound(t *testing.T) {
	childID := uuid.New()
	repo := &stubPrefillRepo{
		terms:   []domain.AdvancePayTermRow{},
		entries: []domain.BookingPatternEntryRow{},
	}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{
		TenantID: uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		BranchID: uuid.MustParse("00000000-0000-4000-8000-000000000002"),
	}
	_, err := uc.Execute(context.Background(), actor, childID.String(), "2026-06")
	if err == nil {
		t.Fatal("expected error for child not found")
	}
}

func TestComputeInvoicePrefill_InvalidBillingMonth(t *testing.T) {
	childID := uuid.New()
	repo := &stubPrefillRepo{}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{}
	_, err := uc.Execute(context.Background(), actor, childID.String(), "invalid-month")
	if err == nil {
		t.Fatal("expected error for invalid billing month format")
	}
}

func TestComputeInvoicePrefill_SiteRateNotSet(t *testing.T) {
	childID := uuid.MustParse("33333333-3333-4333-8333-333333333003")
	termRow := makeTerm()
	termRow.SiteHourlyRateMinor = 0

	repo := &stubPrefillRepo{
		terms:   []domain.AdvancePayTermRow{termRow},
		entries: []domain.BookingPatternEntryRow{},
	}
	uc := NewComputeInvoicePrefill(repo, &stubPrefillTxMgr{repo: repo})

	actor := tenant.ActorContext{
		TenantID: termRow.TenantID,
		BranchID: termRow.BranchID,
	}
	result, err := uc.Execute(context.Background(), actor, childID.String(), "2026-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, w := range result.Warnings {
		if w == "site_rate_not_set" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected site_rate_not_set warning")
	}
	if result.SubtotalMinor != 0 {
		t.Fatalf("expected zero subtotal when no site rate, got %d", result.SubtotalMinor)
	}
}
