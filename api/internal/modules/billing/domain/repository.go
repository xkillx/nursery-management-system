package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Tx = any

var ErrSiteNotFound = errors.New("site not found")

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// SiteRateRepository is the interface for updating a site's core hourly rate.
// It follows the adapter pattern (KTD-2): defined in billing domain, implemented
// in bootstrap by wrapping the owner module's repository.
type SiteRateRepository interface {
	GetCoreHourlyRate(ctx context.Context, tenantID, branchID uuid.UUID) (int, bool, error)
	UpdateCoreHourlyRate(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, rateMinor int) error
}

type BillingRepository interface {
	// Advance-pay generation: list active terms covering the billing month,
	// joined with child + funding data, locked FOR UPDATE.
	ListActiveTermsForGeneration(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]AdvancePayTermRow, error)
	// ListActiveTerms is the non-tx variant used by the preflight preview.
	ListActiveTerms(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]AdvancePayTermRow, error)
	ListBookingPatternEntries(ctx context.Context, tx Tx, tenantID, branchID, patternID uuid.UUID) ([]BookingPatternEntryRow, error)
	// ListActiveAdHocBookingsForChildInMonth returns enriched ad-hoc booking rows
	// (joined with session_types) used for invoice line generation.
	ListActiveAdHocBookingsForChildInMonth(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, from, to time.Time) ([]AdHocBookingRow, error)

	CreateInvoiceRun(ctx context.Context, tx Tx, params InvoiceRunCreateParams) error
	CompleteInvoiceRun(ctx context.Context, tx Tx, params InvoiceRunCompleteParams) error
	GetMonthlyInvoiceForUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (InvoiceRow, bool, error)
	CreateDraftInvoice(ctx context.Context, tx Tx, params DraftInvoiceCreateParams) error
	UpdateDraftInvoice(ctx context.Context, tx Tx, params DraftInvoiceUpdateParams) error
	DeleteDraftSystemInvoiceLines(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) error
	ListDraftExtraLines(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) ([]ExtraLineRow, error)
	InsertInvoiceLine(ctx context.Context, tx Tx, params InvoiceLineCreateParams) error

	// Manager Invoice Review (API-18) — read-only, no transaction required.
	ListInvoicesForManagerReview(ctx context.Context, tenantID, branchID uuid.UUID, filters InvoiceReviewFilters) ([]InvoiceReviewRow, error)
	GetInvoiceForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) (InvoiceReviewRow, bool, error)
	ListInvoiceLinesForManagerReview(ctx context.Context, tenantID, branchID, invoiceID uuid.UUID) ([]InvoiceReviewLineRow, error)

	// Invoice Issue (API-19) — transactional methods using Tx.
	GetInvoiceForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID, invoiceID uuid.UUID) (InvoiceIssueCandidateRow, bool, error)
	ListDraftInvoicesForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]InvoiceIssueCandidateRow, error)
	ListSelectedInvoicesForIssueForUpdate(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, invoiceIDs []uuid.UUID) ([]InvoiceIssueCandidateRow, error)
	AllocateInvoiceNumberSequence(ctx context.Context, tx Tx, tenantID, branchID uuid.UUID, year, month int) (int, error)
	MarkInvoiceIssued(ctx context.Context, tx Tx, params IssueInvoiceUpdateParams) (int64, error)

	// Parent Invoice View (API-21) — read-only, no transaction required.
	ListInvoicesForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID, filters ParentInvoiceFilters) ([]ParentInvoiceRow, error)
	GetInvoiceForParent(ctx context.Context, tenantID, branchID, membershipID, invoiceID uuid.UUID) (ParentInvoiceRow, bool, error)
	ListInvoiceLinesForParent(ctx context.Context, tenantID, branchID, membershipID, invoiceID uuid.UUID) ([]ParentInvoiceLineRow, error)

	// Overdue Transition (API-20) — transactional methods using Tx.
	TryAcquireOverdueTransitionJobLock(ctx context.Context, tx Tx) (bool, error)
	MarkIssuedInvoicesOverdue(ctx context.Context, tx Tx, cutoffUTC time.Time) ([]OverdueTransitionedInvoice, error)
}

// InvoiceRow maps a row from the invoices table.
type InvoiceRow struct {
	ID                   uuid.UUID
	Status               string
	InvoiceKind          string
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
	CalculationDetails   json.RawMessage
}

// AdvancePayTermRow is the per-term row from BillingListActiveTermsForGeneration.
// The application layer joins this with booking-pattern entries to compute the
// per-term monthly invoice.
type AdvancePayTermRow struct {
	TermID                 uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	ChildID                uuid.UUID
	TermStartDate          time.Time
	TermEndDate            time.Time
	BookingPatternID       uuid.UUID
	SiteHourlyRateMinor    int
	Status                 string
	FirstName              string
	MiddleName             *string
	LastName               *string
	DateOfBirth            time.Time
	StartDate              time.Time
	EndDate                *time.Time
	HasParentCarerContact  bool
	FundingProfileID       *uuid.UUID
	FundedAllowanceMinutes *int
	TermTimeOnly           bool
	FundingModel           string
	FundedHoursPerWeek     *float64
	AdHocRateMultiplier    float64
}

// BookingPatternEntryRow is the per-entry row from ListBookingPatternEntries.
type BookingPatternEntryRow struct {
	DayOfWeek               int
	SessionTypeID           uuid.UUID
	SessionTypeName         string
	StartMinutes            int
	EndMinutes              int
	SessionTypeKind         string
	SessionTypeFlatFeeMinor *int
}

// ExtraLineRow maps an extra invoice line row.
type ExtraLineRow struct {
	ID              uuid.UUID
	LineKind        string
	LineAmountMinor int
	Details         json.RawMessage
}
