package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InvoiceReviewFilters holds optional list filters for manager invoice review.
type InvoiceReviewFilters struct {
	BillingMonth *time.Time
	Status       *string
	ChildID      *uuid.UUID
	Limit        int
	Offset       int
}

// InvoiceReviewRow maps a joined invoice + child + run row for manager review.
type InvoiceReviewRow struct {
	ID                      uuid.UUID
	InvoiceKind             string
	InvoiceNumber           *string
	Status                  string
	ChildID                 uuid.UUID
	ChildFirstName          string
	ChildMiddleName         *string
	ChildLastName           *string
	BillingMonth            time.Time
	PeriodStartDate         time.Time
	PeriodEndDate           time.Time
	CurrencyCode            string
	SubtotalMinor           int
	FundedDeductionMinor    int
	TotalDueMinor           int
	AmountPaidMinor         int
	DueAt                   *time.Time
	IssuedAt                *time.Time
	LockedAt                *time.Time
	PaidAt                  *time.Time
	PaymentFailedAt         *time.Time
	PaymentStatusUpdatedAt  *time.Time
	AdjustsInvoiceID        *uuid.UUID
	AdjustmentReasonCode    *string
	AdjustmentReasonNote    *string
	GeneratedRunID          *uuid.UUID
	GeneratedRunStatus      *string
	GeneratedRunStartedAt   *time.Time
	GeneratedRunCompletedAt *time.Time
	GeneratedRunDetails     json.RawMessage
	CalculationDetails      json.RawMessage
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// InvoiceReviewLineRow maps an invoice line for manager review.
type InvoiceReviewLineRow struct {
	ID                     uuid.UUID
	LineKind               string
	Description            string
	SortOrder              int
	QuantityMinutes        *int
	UnitAmountMinor        *int
	LineAmountMinor        int
	RawAttendedMinutes     *int
	RoundedAttendedMinutes *int
	FundedAllowanceMinutes *int
	FundedDeductionMinutes *int
	CoreBillableMinutes    *int
	SessionCount           *int
}

// InvoiceReviewCalculation is the curated calculation object for the API response.
type InvoiceReviewCalculation struct {
	CoreHourlyRateMinor    int                     `json:"core_hourly_rate_minor"`
	RawAttendedMinutes     int                     `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                     `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int                     `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                     `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                     `json:"core_billable_minutes"`
	IncludedSessionCount   int                     `json:"included_session_count"`
	CoreSubtotalMinor      int                     `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int                     `json:"extras_total_minor"`
	SourceSessions         []SourceSessionSnapshot `json:"source_sessions"`
}

// InvoiceRunExceptionReference is a compact exception from an invoice run.
type InvoiceRunExceptionReference struct {
	ChildID         string   `json:"child_id"`
	ChildFirstName  string   `json:"child_first_name"`
	ChildMiddleName *string  `json:"child_middle_name"`
	ChildLastName   *string  `json:"child_last_name"`
	BlockerCodes    []string `json:"blocker_codes"`
}

// InvoiceReviewDetail is the full detail response for a single invoice.
type InvoiceReviewDetail struct {
	Invoice                    InvoiceReviewRow
	Lines                      []InvoiceReviewLineRow
	Calculation                InvoiceReviewCalculation
	GeneratedRunExceptions     []InvoiceRunExceptionReference
	GeneratedRunExceptionCount int
}
