package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Invoice status constants.
const (
	InvoiceStatusDraft         = "draft"
	InvoiceStatusIssued        = "issued"
	InvoiceStatusPaymentFailed = "payment_failed"
	InvoiceStatusPaid          = "paid"
	InvoiceStatusOverdue       = "overdue"
)

// Invoice kind constants.
const (
	InvoiceKindMonthly = "monthly"
)

// Invoice line kind constants.
const (
	LineKindCoreChildcare   = "core_childcare"
	LineKindFundedDeduction = "funded_deduction"
	LineKindExtra           = "extra"
)

// Invoice run status constants.
const (
	InvoiceRunStatusStarted                 = "started"
	InvoiceRunStatusCompleted               = "completed"
	InvoiceRunStatusCompletedWithExceptions = "completed_with_exceptions"
)

// Invoice run type constants.
const (
	InvoiceRunTypeDraftGeneration = "draft_generation"
	InvoiceRunTypeIssue           = "issue"
)

// Generation-selected child blocker codes (added by API-17).
const (
	BlockerChildNotFound          BlockerCode = "child_not_found"
	BlockerChildNotInBillingMonth BlockerCode = "child_not_in_billing_month"
)

// DraftInvoiceAction indicates whether a draft was newly created or updated.
type DraftInvoiceAction string

const (
	DraftCreated DraftInvoiceAction = "created"
	DraftUpdated DraftInvoiceAction = "updated"
)

// GenerateDraftInvoicesParams is the validated request for draft generation.
type GenerateDraftInvoicesParams struct {
	BillingMonth time.Time
	ChildIDs     []uuid.UUID // nil = full-month, empty = no-op
	IsFullMonth  bool
}

// DraftGenerationChildResult is the per-child outcome of draft generation.
type DraftGenerationChildResult struct {
	ChildID              uuid.UUID
	ChildFirstName       string
	ChildMiddleName      *string
	ChildLastName        *string
	Action               DraftInvoiceAction
	InvoiceID            uuid.UUID
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
}

// DraftGenerationBlockedChild represents a child-month that could not be generated.
type DraftGenerationBlockedChild struct {
	ChildID         uuid.UUID
	ChildFirstName  string
	ChildMiddleName *string
	ChildLastName   *string
	Blockers        []PreflightBlocker
}

// DraftGenerationResult is the application-level result of a generation run.
type DraftGenerationResult struct {
	RunID        uuid.UUID
	BillingMonth string
	RunStatus    string
	Generated    []DraftGenerationChildResult
	Blocked      []DraftGenerationBlockedChild
	Summary      DraftGenerationSummary
}

// DraftGenerationSummary holds aggregate counts and totals for the run.
type DraftGenerationSummary struct {
	EligibleCount int
	SuccessCount  int
	BlockedCount  int
	TotalDueMinor int
}

// InvoiceCalculationDetails is stored as JSON in invoice.calculation_details.
//
// The advance-pay shape records the booking-driven inputs:
//   - term_id, booking_pattern_id: the term + pattern that drove the calculation
//   - booked_core_minutes: the sum over all pattern entries in the month
//   - booked_sessions: per-session explainability (one row per occurrence)
//   - booked_per_entry: per-(day, session) subtotals
//
// The legacy attendance-driven fields are kept as zero values to keep the
// downstream consumers (parent invoice view, manager invoice review) able to
// render the same high-level shape.
type InvoiceCalculationDetails struct {
	BillingMonth           string     `json:"billing_month"`
	ChildID                uuid.UUID  `json:"child_id"`
	CoreHourlyRateMinor    int        `json:"core_hourly_rate_minor"`
	CoreSubtotalMinor      int        `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int        `json:"extras_total_minor"`
	ManualExtrasSupported  bool       `json:"manual_extras_supported"`
	FundingProfileID       *uuid.UUID `json:"funding_profile_id,omitempty"`
	FundedAllowanceMinutes int        `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int        `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int        `json:"core_billable_minutes"`

	// Advance-pay source.
	TermID            uuid.UUID              `json:"term_id"`
	BookingPatternID  uuid.UUID              `json:"booking_pattern_id"`
	BookedCoreMinutes int                    `json:"booked_core_minutes"`
	BookedSessions    []BookedSession        `json:"booked_sessions"`
	BookedPerEntry    []BookedEntryBreakdown `json:"booked_per_entry"`
}

// SourceSessionSnapshot is retained for backward-compatible JSON unmarshaling
// of any historical invoices that used the attendance-driven shape. New
// invoices never populate this field.
type SourceSessionSnapshot struct {
	SessionID              uuid.UUID  `json:"session_id"`
	Status                 string     `json:"status"`
	CheckInAt              time.Time  `json:"check_in_at"`
	CheckOutAt             *time.Time `json:"check_out_at,omitempty"`
	RawElapsedMinutes      int        `json:"raw_elapsed_minutes"`
	RoundedBillableMinutes int        `json:"rounded_billable_minutes"`
}

// MarshalCalculationDetails serializes calculation details to JSON bytes.
func MarshalCalculationDetails(d InvoiceCalculationDetails) ([]byte, error) {
	return json.Marshal(d)
}

// CoreLineDetails is stored as JSON in the core_childcare invoice line details.
type CoreLineDetails struct {
	BookedCoreMinutes int                    `json:"booked_core_minutes"`
	BookedSessions    []BookedSession        `json:"booked_sessions"`
	BookedPerEntry    []BookedEntryBreakdown `json:"booked_per_entry"`
}

// FundedDeductionLineDetails is stored as JSON in the funded_deduction invoice line details.
type FundedDeductionLineDetails struct {
	FundingProfileID       uuid.UUID `json:"funding_profile_id"`
	FundedAllowanceMinutes int       `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int       `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int       `json:"core_billable_minutes"`
}

// Audit action types for draft invoice generation.
const (
	AuditInvoiceDraftGenerated   = "invoice_draft_generated"
	AuditInvoiceDraftRegenerated = "invoice_draft_regenerated"
	AuditInvoiceIssued           = "invoice_issued"
	AuditEntityInvoice           = "invoice"
)

// Issue blocker codes.
const (
	IssueBlockerInvoiceNotFound          = "invoice_not_found"
	IssueBlockerInvoiceNotInBillingMonth = "invoice_not_in_billing_month"
	IssueBlockerInvoiceNotDraft          = "invoice_not_draft"
	IssueBlockerInvoiceNotMonthly        = "invoice_not_monthly"
)

// IssueInvoiceResult is the result of a single invoice issue.
type IssueInvoiceResult struct {
	InvoiceID     uuid.UUID
	InvoiceNumber string
	Status        string
	IssuedAt      time.Time
	LockedAt      time.Time
	DueAt         time.Time
	IssuedRunID   uuid.UUID
	TotalDueMinor int
}

// IssuedInvoiceResult is the per-invoice outcome of a bulk issue.
type IssuedInvoiceResult struct {
	InvoiceID       uuid.UUID
	ChildID         uuid.UUID
	ChildFirstName  string
	ChildMiddleName *string
	ChildLastName   *string
	InvoiceNumber   string
	IssuedAt        time.Time
	DueAt           time.Time
	TotalDueMinor   int
}

// InvoiceIssueBlocked represents an invoice that could not be issued.
type InvoiceIssueBlocked struct {
	InvoiceID       uuid.UUID
	ChildID         *uuid.UUID
	ChildFirstName  string
	ChildMiddleName *string
	ChildLastName   *string
	Blockers        []InvoiceIssueBlocker
}

// InvoiceIssueBlocker is a single blocker reason for an invoice issue.
type InvoiceIssueBlocker struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BulkIssueInvoicesResult is the application-level result of a bulk issue run.
type BulkIssueInvoicesResult struct {
	RunID        uuid.UUID
	BillingMonth string
	Status       string
	Summary      InvoiceIssueSummary
	Issued       []IssuedInvoiceResult
	Blocked      []InvoiceIssueBlocked
}

// InvoiceIssueSummary holds aggregate counts and totals for an issue run.
type InvoiceIssueSummary struct {
	EligibleCount int
	SuccessCount  int
	BlockedCount  int
	TotalDueMinor int
}

// InvoiceIssueCandidateRow maps a row from the invoices table for issue.
type InvoiceIssueCandidateRow struct {
	ID              uuid.UUID
	ChildID         uuid.UUID
	ChildFirstName  string
	ChildMiddleName *string
	ChildLastName   *string
	BillingMonth    time.Time
	InvoiceKind     string
	Status          string
	TotalDueMinor   int
}

// IssueInvoiceUpdateParams holds fields needed to mark an invoice as issued.
type IssueInvoiceUpdateParams struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	BranchID             uuid.UUID
	InvoiceNumber        string
	IssuedSequence       int
	IssuedRunID          uuid.UUID
	IssuedAt             time.Time
	IssuedByUserID       uuid.UUID
	IssuedByMembershipID uuid.UUID
	DueAt                time.Time
}

// InvoiceRunCreateParams holds fields needed to create an invoice_runs row.
type InvoiceRunCreateParams struct {
	ID                      uuid.UUID
	TenantID                uuid.UUID
	BranchID                uuid.UUID
	BillingMonth            time.Time
	RunType                 string
	Status                  string
	RequestedByUserID       uuid.UUID
	RequestedByMembershipID uuid.UUID
	RequestID               string
}

// InvoiceRunCompleteParams holds fields needed to complete an invoice_runs row.
type InvoiceRunCompleteParams struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	Status        string
	EligibleCount int
	SuccessCount  int
	BlockedCount  int
	Details       json.RawMessage
}

// DraftInvoiceCreateParams holds fields needed to create a draft invoice.
type DraftInvoiceCreateParams struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	BranchID             uuid.UUID
	ChildID              uuid.UUID
	BillingMonth         time.Time
	GeneratedRunID       uuid.UUID
	CurrencyCode         string
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
	PeriodStartDate      time.Time
	PeriodEndDate        time.Time
	CalculationDetails   []byte
}

// DraftInvoiceUpdateParams holds fields needed to update a draft invoice header.
type DraftInvoiceUpdateParams struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	BranchID             uuid.UUID
	GeneratedRunID       uuid.UUID
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
	CalculationDetails   []byte
}

// InvoiceLineCreateParams holds fields needed to insert an invoice line.
type InvoiceLineCreateParams struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	InvoiceID              uuid.UUID
	LineKind               string
	Description            string
	SortOrder              int
	QuantityMinutes        int
	UnitAmountMinor        int
	LineAmountMinor        int
	RawAttendedMinutes     int
	RoundedAttendedMinutes int
	FundedAllowanceMinutes int
	FundedDeductionMinutes int
	CoreBillableMinutes    int
	SessionCount           int
	Details                []byte
}

// ExtraLineSnapshot represents a preserved extra line for regeneration totals.
type ExtraLineSnapshot struct {
	LineID          uuid.UUID
	LineAmountMinor int
}

// OverdueTransitionedInvoice represents an invoice transitioned to overdue.
type OverdueTransitionedInvoice struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
}

// OverdueTransitionResult is the result of an overdue transition job run.
type OverdueTransitionResult struct {
	LockAcquired      bool
	CurrentLondonDate time.Time
	CutoffUTC         time.Time
	Transitioned      []OverdueTransitionedInvoice
}
