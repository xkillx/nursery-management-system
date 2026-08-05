package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

// genRegenRepoStub drives GenerateTermInvoices regeneration: it reports an
// existing draft invoice, returns pre-seeded system lines (with markers) via
// ListInvoiceLinesForManagerReviewTx, and captures every InsertInvoiceLine so
// the test can assert preserved descriptions and marker propagation (U3).
type genRegenRepoStub struct {
	invoiceFound       bool
	existingInvoice    domain.InvoiceRow
	existingLines      []domain.InvoiceReviewLineRow
	insertedLines      []domain.InvoiceLineCreateParams
	adHocRows          []domain.AdHocBookingRow
	hourlyRows         []domain.HourlyBookingRow
	funding            domain.FundedChildInfo
	bookingEntries     []domain.BookedPatternEntry
	recalculateMinimal bool
}

func (s *genRegenRepoStub) GetMonthlyInvoiceForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (domain.InvoiceRow, bool, error) {
	if !s.invoiceFound {
		return domain.InvoiceRow{}, false, nil
	}
	return s.existingInvoice, true, nil
}
func (s *genRegenRepoStub) ListDraftExtraLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.ExtraLineRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ListInvoiceLinesForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	return s.existingLines, nil
}
func (s *genRegenRepoStub) DeleteDraftSystemInvoiceLines(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	return nil
}
func (s *genRegenRepoStub) UpdateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceUpdateParams) error {
	return nil
}
func (s *genRegenRepoStub) CreateDraftInvoice(_ context.Context, _ domain.Tx, _ domain.DraftInvoiceCreateParams) error {
	return nil
}
func (s *genRegenRepoStub) InsertInvoiceLine(_ context.Context, _ domain.Tx, params domain.InvoiceLineCreateParams) error {
	s.insertedLines = append(s.insertedLines, params)
	return nil
}
func (s *genRegenRepoStub) ListActiveBookingsForGeneration(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ListActiveBookings(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.BillableChildRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ListActiveAdHocBookingsForChildInMonth(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.AdHocBookingRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) CreateInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCreateParams) error {
	return nil
}
func (s *genRegenRepoStub) CompleteInvoiceRun(_ context.Context, _ domain.Tx, _ domain.InvoiceRunCompleteParams) error {
	return nil
}
func (s *genRegenRepoStub) ListInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) ([]domain.InvoiceReviewRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) CountInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceReviewFilters) (int, error) {
	return 0, nil
}
func (s *genRegenRepoStub) GetInvoiceForManagerReview(_ context.Context, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return domain.InvoiceReviewRow{}, false, nil
}
func (s *genRegenRepoStub) GetInvoiceForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return domain.InvoiceReviewRow{}, false, nil
}
func (s *genRegenRepoStub) ListInvoiceLinesForManagerReview(_ context.Context, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	return s.existingLines, nil
}
func (s *genRegenRepoStub) ExportInvoicesForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceReviewRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ExportInvoiceDetailsForManagerReview(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceExportLineRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) InvoiceSummaryByMonth(_ context.Context, _, _ uuid.UUID, _ domain.InvoiceExportFilters) ([]domain.InvoiceMonthSummary, error) {
	return nil, nil
}
func (s *genRegenRepoStub) InvoiceOverdueSummary(_ context.Context, _, _ uuid.UUID) (domain.OverdueSummary, error) {
	return domain.OverdueSummary{}, nil
}
func (s *genRegenRepoStub) InvoiceOverdueTopItems(_ context.Context, _, _ uuid.UUID) ([]domain.OverdueSummaryItem, error) {
	return nil, nil
}
func (s *genRegenRepoStub) GetInvoiceForIssueForUpdate(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceIssueCandidateRow, bool, error) {
	return domain.InvoiceIssueCandidateRow{}, false, nil
}
func (s *genRegenRepoStub) ListDraftInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ time.Time) ([]domain.InvoiceIssueCandidateRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ListSelectedInvoicesForIssueForUpdate(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _ []uuid.UUID) ([]domain.InvoiceIssueCandidateRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) AllocateInvoiceNumberSequence(_ context.Context, _ domain.Tx, _, _ uuid.UUID, _, _ int) (int, error) {
	return 1, nil
}
func (s *genRegenRepoStub) MarkInvoiceIssued(_ context.Context, _ domain.Tx, _ domain.IssueInvoiceUpdateParams) (int64, error) {
	return 1, nil
}
func (s *genRegenRepoStub) GetInvoiceLine(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) (domain.InvoiceLine, bool, error) {
	return domain.InvoiceLine{}, false, nil
}
func (s *genRegenRepoStub) UpdateInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ int, _, _ domain.Money, _ []byte) (int64, error) {
	return 1, nil
}
func (s *genRegenRepoStub) DeleteInvoiceLine(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	return 1, nil
}
func (s *genRegenRepoStub) MarkInvoiceVoid(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string, _ time.Time) (int64, error) {
	return 1, nil
}
func (s *genRegenRepoStub) DeleteInvoice(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (int64, error) {
	return 1, nil
}
func (s *genRegenRepoStub) ListInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) ([]domain.ParentInvoiceRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) CountInvoicesForParent(_ context.Context, _, _, _ uuid.UUID, _ domain.ParentInvoiceFilters) (int, error) {
	return 0, nil
}
func (s *genRegenRepoStub) GetInvoiceForParent(_ context.Context, _, _, _, _ uuid.UUID) (domain.ParentInvoiceRow, bool, error) {
	return domain.ParentInvoiceRow{}, false, nil
}
func (s *genRegenRepoStub) ListInvoiceLinesForParent(_ context.Context, _, _, _, _ uuid.UUID) ([]domain.ParentInvoiceLineRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) TryAcquireOverdueTransitionJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	return true, nil
}
func (s *genRegenRepoStub) MarkIssuedInvoicesOverdue(_ context.Context, _ domain.Tx, _ time.Time) ([]domain.OverdueTransitionedInvoice, error) {
	return nil, nil
}
func (s *genRegenRepoStub) TryAcquireReminderJobLock(_ context.Context, _ domain.Tx) (bool, error) {
	return true, nil
}
func (s *genRegenRepoStub) ListInvoicesDueSoon(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) ListInvoicesDueToday(_ context.Context, _ domain.Tx) ([]domain.InvoiceReminderRow, error) {
	return nil, nil
}
func (s *genRegenRepoStub) InsertInvoiceReminderLog(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (s *genRegenRepoStub) CountRecentInvoiceResendsTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (int, error) {
	return 0, nil
}
func (s *genRegenRepoStub) LockInvoiceForResendTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (bool, error) {
	return true, nil
}
func (s *genRegenRepoStub) GetLatestResendAtTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (time.Time, error) {
	return time.Time{}, nil
}

type stubClosureLookup struct{ dates []time.Time }

func (s *stubClosureLookup) GetClosureDatesForBranchAndMonth(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]time.Time, error) {
	return s.dates, nil
}

type stubHolidayLookup struct {
	periods []domain.HolidayPeriodDateRange
}

func (s *stubHolidayLookup) GetHolidayPeriodsForBranchAndMonth(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.HolidayPeriodDateRange, error) {
	return s.periods, nil
}

type stubTermDateLookup struct{ ranges []domain.TermDateRange }

func (s *stubTermDateLookup) GetTermDateRangesForBranchAndMonth(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.TermDateRange, error) {
	return s.ranges, nil
}

type stubFundingLookup struct{ info domain.FundedChildInfo }

func (s *stubFundingLookup) GetChildFunding(_ context.Context, _, _, _ uuid.UUID, _ time.Time) (domain.FundedChildInfo, error) {
	return s.info, nil
}

type stubAdHocLookup struct{ rows []domain.AdHocBookingRow }

func (s *stubAdHocLookup) ListActiveBookingsForChildInMonth(_ context.Context, _, _, _ uuid.UUID, _ time.Time) ([]domain.AdHocBookingRow, error) {
	return s.rows, nil
}

type stubHourlyLookup struct{ rows []domain.HourlyBookingRow }

func (s *stubHourlyLookup) ListActiveByChildAndMonth(_ context.Context, _, _, _ uuid.UUID, _, _ time.Time) ([]domain.HourlyBookingRow, error) {
	return s.rows, nil
}

// newRegenUC builds a GenerateTermInvoices wired to a minimal advance-pay
// calculation: one Monday session of 540 minutes yields a deterministic core
// subtotal, with no funding/ad-hoc/hourly noise unless the test adds them.
func newRegenUC(repo *genRegenRepoStub) (*GenerateTermInvoices, domain.BillableChildRow) {
	bookingRow := makeBooking()
	bookingRow.SiteHourlyRateMinor = 600
	entries := []domain.BookedPatternEntry{
		{DayOfWeek: 1, SessionType: domain.BookedSessionType{ID: "st-full", Name: "Full Day", StartMinutes: 480, EndMinutes: 1020, DurationMinutes: 540}},
	}
	uc := &GenerateTermInvoices{
		repo:                 repo,
		auditW:               &stubAuditWriter{},
		termDateLookup:       &stubTermDateLookup{},
		adHocLookup:          &stubAdHocLookup{rows: repo.adHocRows},
		hourlyLookup:         &stubHourlyLookup{rows: repo.hourlyRows},
		closureDateLookup:    &stubClosureLookup{},
		holidayPeriodLookup:  &stubHolidayLookup{},
		fundingLookup:        &stubFundingLookup{info: repo.funding},
		bookingEntriesLookup: &stubBookingEntriesLookup{entries: entries},
	}
	return uc, bookingRow
}

func (uc *GenerateTermInvoices) run(t *testing.T, repo *genRegenRepoStub, bookingRow domain.BillableChildRow) {
	t.Helper()
	actor := tenant.ActorContext{
		TenantID: bookingRow.TenantID,
		BranchID: bookingRow.BranchID,
		UserID:   uuid.MustParse("b3000000-0000-0000-0000-000000000001"),
	}
	_, err := uc.Execute(context.Background(), GenerateTermInvoicesInput{
		Actor:           actor,
		BillingMonth:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		BillingMonthRaw: "2026-05",
		RunID:           uuid.New(),
	}, []domain.BillableChildRow{bookingRow}, map[uuid.UUID]struct{}{bookingRow.ChildID: {}}, false)
	if err != nil {
		t.Fatalf("generate execute: %v", err)
	}
}

func insertedLine(repo *genRegenRepoStub, kind string) (domain.InvoiceLineCreateParams, bool) {
	for _, l := range repo.insertedLines {
		if l.LineKind == kind {
			return l, true
		}
	}
	return domain.InvoiceLineCreateParams{}, false
}

func TestRegeneration_PreservesRenamedCoreDescription(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{
				LineKind:            domain.LineKindCoreChildcare,
				Description:         "Wrap-around care",
				DescriptionOverride: true,
			},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	core, ok := insertedLine(repo, domain.LineKindCoreChildcare)
	if !ok {
		t.Fatal("expected core line inserted on regeneration")
	}
	if core.Description != "Wrap-around care" {
		t.Errorf("core description = %q, want %q (preserved)", core.Description, "Wrap-around care")
	}
	if !domain.HasLineDescriptionOverride(core.Details) {
		t.Errorf("expected description_override marker carried into new core line, got %s", core.Details)
	}
}

func TestRegeneration_ResetsUntouchedCoreToDefault(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{
				LineKind:            domain.LineKindCoreChildcare,
				Description:         "May 2026 Recurring Booking",
				DescriptionOverride: false,
			},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	core, ok := insertedLine(repo, domain.LineKindCoreChildcare)
	if !ok {
		t.Fatal("expected core line inserted on regeneration")
	}
	if core.Description != "May 2026 Recurring Booking" {
		t.Errorf("core description = %q, want %q", core.Description, "May 2026 Recurring Booking")
	}
	if domain.HasLineDescriptionOverride(core.Details) {
		t.Errorf("expected no marker on untouched core line, got %s", core.Details)
	}
}

func TestRegeneration_ResetsLegacyCoreChildcareToDefault(t *testing.T) {
	// R8: a pre-feature draft labelled "Core childcare" (no marker) relabels
	// to the new default at its next regeneration.
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{
				LineKind:            domain.LineKindCoreChildcare,
				Description:         "Core childcare",
				DescriptionOverride: false,
			},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	core, ok := insertedLine(repo, domain.LineKindCoreChildcare)
	if !ok {
		t.Fatal("expected core line inserted on regeneration")
	}
	if core.Description != "May 2026 Recurring Booking" {
		t.Errorf("core description = %q, want %q", core.Description, "May 2026 Recurring Booking")
	}
}

func TestRegeneration_PreservesRenamedDeductionAndRecomputesUntouched(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{
				LineKind:            domain.LineKindFundedDeduction,
				Description:         "Nursery funding offset",
				DescriptionOverride: true,
			},
		},
		funding: domain.FundedChildInfo{
			HasFunding:             true,
			FundingType:            "stretched",
			FundedAllowanceMinutes: 570,
			FundedHourlyRateMinor:  600,
			FundedHoursPerWeek:     15,
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	ded, ok := insertedLine(repo, domain.LineKindFundedDeduction)
	if !ok {
		t.Fatal("expected deduction line inserted on regeneration")
	}
	if ded.Description != "Nursery funding offset" {
		t.Errorf("deduction description = %q, want %q (preserved)", ded.Description, "Nursery funding offset")
	}
	if !domain.HasLineDescriptionOverride(ded.Details) {
		t.Errorf("expected marker on preserved deduction line, got %s", ded.Details)
	}
}

func TestRegeneration_UntouchedDeductionRecomputesFundingDefault(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{
				LineKind:            domain.LineKindFundedDeduction,
				Description:         "Stretched funding (≈15.0h/week)",
				DescriptionOverride: false,
			},
		},
		funding: domain.FundedChildInfo{
			HasFunding:             true,
			FundingType:            "stretched",
			FundedAllowanceMinutes: 570,
			FundedHourlyRateMinor:  600,
			FundedHoursPerWeek:     15,
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	ded, ok := insertedLine(repo, domain.LineKindFundedDeduction)
	if !ok {
		t.Fatal("expected deduction line inserted on regeneration")
	}
	if ded.Description != "Stretched funding (≈15.0h/week)" {
		t.Errorf("deduction description = %q, want recomputed default", ded.Description)
	}
	if domain.HasLineDescriptionOverride(ded.Details) {
		t.Errorf("expected no marker on untouched deduction, got %s", ded.Details)
	}
}

func TestRegeneration_HourlyPreservedByBookingRefAndDroppedWhenRemoved(t *testing.T) {
	keptID := uuid.New()
	removedID := uuid.New()
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{LineKind: domain.LineKindHourly, Description: "Kept hourly", DescriptionOverride: true, HourlyBookingID: &keptID},
			{LineKind: domain.LineKindHourly, Description: "Removed hourly", DescriptionOverride: true, HourlyBookingID: &removedID},
		},
		hourlyRows: []domain.HourlyBookingRow{
			{ID: keptID, ChildID: bookingChildID, CalendarDate: time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), StartTimeMinutes: 480, DurationMinutes: 60},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.run(t, repo, bookingRow)

	var kept domain.InvoiceLineCreateParams
	var found bool
	for _, l := range repo.insertedLines {
		if l.LineKind == domain.LineKindHourly {
			kept = l
			found = true
		}
	}
	if !found {
		t.Fatal("expected hourly line inserted for the live booking")
	}
	if kept.Description != "Kept hourly" {
		t.Errorf("hourly description = %q, want %q (preserved by booking ref)", kept.Description, "Kept hourly")
	}
	if !domain.HasLineDescriptionOverride(kept.Details) {
		t.Errorf("expected marker on preserved hourly line, got %s", kept.Details)
	}
	// The removed booking's renamed line must not be re-inserted.
	count := 0
	for _, l := range repo.insertedLines {
		if l.LineKind == domain.LineKindHourly {
			count++
		}
	}
	if count != 1 {
		t.Errorf("hourly lines inserted = %d, want 1 (removed booking's renamed line dropped)", count)
	}
}

func TestRegeneration_AdHocSurvivesAndNoDuplicate(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: []domain.InvoiceReviewLineRow{
			{LineKind: domain.LineKindAdHoc, Description: "Ad-hoc session: Swim on 05 May (×1.50)", DescriptionOverride: false},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	// No ad-hoc lookup rows: the pre-existing ad-hoc line is left in place by
	// the delete step and nothing is re-inserted, so no duplicate can exist.
	uc.adHocLookup = &stubAdHocLookup{rows: nil}
	uc.run(t, repo, bookingRow)

	count := 0
	for _, l := range repo.insertedLines {
		if l.LineKind == domain.LineKindAdHoc {
			count++
		}
	}
	if count != 0 {
		t.Errorf("ad-hoc lines re-inserted on regeneration = %d, want 0 (left in place)", count)
	}
}

var bookingChildID = uuid.MustParse("33333333-3333-4333-8333-333333333003")

func TestRegeneration_AdHocInsertMatchesLiveLookup(t *testing.T) {
	repo := &genRegenRepoStub{
		invoiceFound: true,
		existingInvoice: domain.InvoiceRow{
			ID:     uuid.New(),
			Status: domain.InvoiceStatusDraft,
		},
		existingLines: nil,
		adHocRows: []domain.AdHocBookingRow{
			{ID: uuid.New(), ChildID: bookingChildID, CalendarDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC), SessionTypeName: "Swim", StartMinutes: 480, EndMinutes: 540},
		},
	}
	uc, bookingRow := newRegenUC(repo)
	uc.adHocLookup = &stubAdHocLookup{rows: repo.adHocRows}
	uc.run(t, repo, bookingRow)

	count := 0
	for _, l := range repo.insertedLines {
		if l.LineKind == domain.LineKindAdHoc {
			count++
		}
	}
	if count != 1 {
		t.Errorf("ad-hoc lines inserted = %d, want 1", count)
	}
}
