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
	ChildName            string
	Action               DraftInvoiceAction
	InvoiceID            uuid.UUID
	SubtotalMinor        int
	FundedDeductionMinor int
	TotalDueMinor        int
}

// DraftGenerationBlockedChild represents a child-month that could not be generated.
type DraftGenerationBlockedChild struct {
	ChildID   uuid.UUID
	ChildName string
	Blockers  []PreflightBlocker
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
type InvoiceCalculationDetails struct {
	BillingMonth           string                  `json:"billing_month"`
	ChildID                uuid.UUID               `json:"child_id"`
	CoreHourlyRateMinor    int                     `json:"core_hourly_rate_minor"`
	CoreSubtotalMinor      int                     `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int                     `json:"extras_total_minor"`
	ManualExtrasSupported  bool                    `json:"manual_extras_supported"`
	FundingProfileID       *uuid.UUID              `json:"funding_profile_id,omitempty"`
	FundedAllowanceMinutes int                     `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                     `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                     `json:"core_billable_minutes"`
	RawAttendedMinutes     int                     `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                     `json:"rounded_attended_minutes"`
	IncludedSessionCount   int                     `json:"included_session_count"`
	SourceSessions         []SourceSessionSnapshot `json:"source_sessions"`
}

// SourceSessionSnapshot captures attendance session data used for invoice calculation.
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
	RawAttendedMinutes     int                     `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                     `json:"rounded_attended_minutes"`
	IncludedSessionCount   int                     `json:"included_session_count"`
	CoreBillableMinutes    int                     `json:"core_billable_minutes"`
	SourceSessions         []SourceSessionSnapshot `json:"source_sessions"`
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
	AuditEntityInvoice           = "invoice"
)

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
