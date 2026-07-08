package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	domainerrors "nursery-management-system/api/internal/platform/errors"
)

// Invoice status constants.
const (
	InvoiceStatusDraft         = "draft"
	InvoiceStatusIssued        = "issued"
	InvoiceStatusPaymentFailed = "payment_failed"
	InvoiceStatusPaid          = "paid"
	InvoiceStatusOverdue       = "overdue"
	InvoiceStatusVoid          = "void"
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
	LineKindAdHoc           = "ad_hoc"
	LineKindHourly          = "hourly"
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
	ChildID         uuid.UUID
	ChildFirstName  string
	ChildMiddleName *string
	ChildLastName   *string
	Action          DraftInvoiceAction
	InvoiceID       uuid.UUID
	Subtotal        Money
	FundedDeduction Money
	TotalDue        Money
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
	TotalDue      Money
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
	CoreHourlyRate         Money      `json:"core_hourly_rate_minor"`
	CoreSubtotal           Money      `json:"core_subtotal_minor"`
	ExtrasTotal            Money      `json:"extras_total_minor"`
	ManualExtrasSupported  bool       `json:"manual_extras_supported"`
	FundingProfileID       *uuid.UUID `json:"funding_profile_id,omitempty"`
	FundedAllowanceMinutes int        `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int        `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int        `json:"core_billable_minutes"`
	TermTimeOnly           bool       `json:"term_time_only"`
	FundingModel           string     `json:"funding_model,omitempty"`
	TermDatesUsed          []string   `json:"term_dates_used,omitempty"`
	ClosureDaysExcluded    []string   `json:"closure_days_excluded,omitempty"`

	// Advance-pay source.
	TermID            uuid.UUID                 `json:"term_id"`
	BookingPatternID  uuid.UUID                 `json:"booking_pattern_id"`
	BookedCoreMinutes int                       `json:"booked_core_minutes"`
	BookedSessions    []BookedSession           `json:"booked_sessions"`
	BookedPerEntry    []BookedEntryBreakdown    `json:"booked_per_entry"`
	HourlyBookings    []HourlyBookingLineDetail `json:"hourly_bookings,omitempty"`
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
	FundingModel           string    `json:"funding_model,omitempty"`
}

// HourlyBookingLineDetail is stored as JSON in the hourly invoice line details.
type HourlyBookingLineDetail struct {
	HourlyBookingID  uuid.UUID `json:"hourly_booking_id"`
	CalendarDate     string    `json:"calendar_date"`
	StartTimeMinutes int       `json:"start_time_minutes"`
	DurationMinutes  int       `json:"duration_minutes"`
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
	TotalDue      Money
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
	TotalDue        Money
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
	TotalDue      Money
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
	TotalDue        Money
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
	ID                 uuid.UUID
	TenantID           uuid.UUID
	BranchID           uuid.UUID
	ChildID            uuid.UUID
	BillingMonth       time.Time
	GeneratedRunID     uuid.UUID
	CurrencyCode       string
	Subtotal           Money
	FundedDeduction    Money
	TotalDue           Money
	PeriodStartDate    time.Time
	PeriodEndDate      time.Time
	CalculationDetails []byte
}

// DraftInvoiceUpdateParams holds fields needed to update a draft invoice header.
type DraftInvoiceUpdateParams struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	BranchID           uuid.UUID
	GeneratedRunID     uuid.UUID
	Subtotal           Money
	FundedDeduction    Money
	TotalDue           Money
	CalculationDetails []byte
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
	UnitAmount             Money
	LineAmount             Money
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
	LineID     uuid.UUID
	LineAmount Money
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

// ── Invoice entity ────────────────────────────────────────────────────────

// InvoiceLine represents a line item on an Invoice entity.
type InvoiceLine struct {
	LineKind               string
	Description            string
	SortOrder              int
	QuantityMinutes        int
	UnitAmount             Money
	LineAmount             Money
	FundedAllowanceMinutes int
	FundedDeductionMinutes int
	CoreBillableMinutes    int
	SessionCount           int
}

// Invoice is the domain entity representing an invoice with lifecycle state.
// DTOs (GenerateDraftInvoicesParams, InvoiceCalculationDetails, etc.) remain
// unchanged for API transport — this entity is the domain model.
type Invoice struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	ChildID       uuid.UUID
	TermID        uuid.UUID
	Status        string
	InvoiceNumber string
	Lines         []InvoiceLine
	TotalMinor    int
	IssuedAt      time.Time
	DueDate       time.Time
	PaidAt        *time.Time
	OverdueAt     *time.Time
	VoidedAt      *time.Time
	VoidReason    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Issue transitions the invoice from draft to issued, assigning an invoice number.
func (inv *Invoice) Issue(invoiceNumber string, issuedAt time.Time) error {
	if inv.Status != InvoiceStatusDraft {
		return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be issued.")
	}
	if invoiceNumber == "" {
		return domainerrors.Validation("Invoice number must not be empty.", "invoice_number")
	}
	inv.InvoiceNumber = invoiceNumber
	inv.IssuedAt = issuedAt
	inv.Status = InvoiceStatusIssued
	return nil
}

// MarkOverdue transitions the invoice from issued to overdue.
func (inv *Invoice) MarkOverdue(overdueAt time.Time) error {
	if inv.Status != InvoiceStatusIssued {
		return domainerrors.Conflict("invoice_not_issued", "Only issued invoices can be marked overdue.")
	}
	inv.OverdueAt = &overdueAt
	inv.Status = InvoiceStatusOverdue
	return nil
}

// Void transitions the invoice from draft to void (cancellation).
func (inv *Invoice) Void(reason string, voidedAt time.Time) error {
	if inv.Status != InvoiceStatusDraft {
		return domainerrors.Conflict("invoice_not_draft", "Only draft invoices can be voided.")
	}
	inv.VoidedAt = &voidedAt
	inv.VoidReason = reason
	inv.Status = InvoiceStatusVoid
	return nil
}

// AddLine appends an invoice line and recalculates the total. Only draft invoices accept lines.
func (inv *Invoice) AddLine(line InvoiceLine) error {
	if inv.Status != InvoiceStatusDraft {
		return domainerrors.Conflict("invoice_not_draft", "Lines can only be added to draft invoices.")
	}
	if line.QuantityMinutes < 0 {
		return domainerrors.Validation("Line quantity minutes must not be negative.", "quantity_minutes")
	}
	inv.Lines = append(inv.Lines, line)
	inv.recalculateTotal()
	return nil
}

// IsDraft returns true if the invoice is in draft status.
func (inv Invoice) IsDraft() bool {
	return inv.Status == InvoiceStatusDraft
}

// IsIssued returns true if the invoice is in issued status.
func (inv Invoice) IsIssued() bool {
	return inv.Status == InvoiceStatusIssued
}

func (inv *Invoice) recalculateTotal() {
	total := 0
	for _, line := range inv.Lines {
		total += line.LineAmount.Minor()
	}
	inv.TotalMinor = total
}
