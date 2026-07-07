package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InvoiceReviewFilters holds optional list filters for manager invoice review.
type InvoiceReviewFilters struct {
	BillingMonth     *time.Time
	BillingMonthFrom *time.Time
	BillingMonthTo   *time.Time
	Status           *string
	ChildID          *uuid.UUID
	Limit            int
	Offset           int
	SortField        string
	SortDir          string
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
	Subtotal                Money
	FundedDeduction         Money
	TotalDue                Money
	AmountPaid              Money
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
	RoomName                *string
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
	UnitAmount             *Money
	LineAmount             Money
	FundedAllowanceMinutes *int
	FundedDeductionMinutes *int
	CoreBillableMinutes    *int
	SessionCount           *int
	FundingModel           *string
}

// InvoiceReviewCalculation is the curated calculation object for the API response.
// Under advance-pay this reflects the booking-driven inputs.
type InvoiceReviewCalculation struct {
	CoreHourlyRate         Money                  `json:"core_hourly_rate_minor"`
	BookedCoreMinutes      int                    `json:"booked_core_minutes"`
	BookedSessionCount     int                    `json:"booked_session_count"`
	FundedAllowanceMinutes int                    `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                    `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                    `json:"core_billable_minutes"`
	CoreSubtotal           Money                  `json:"core_subtotal_minor"`
	ExtrasTotal            Money                  `json:"extras_total_minor"`
	TermID                 uuid.UUID              `json:"term_id"`
	BookingPatternID       uuid.UUID              `json:"booking_pattern_id"`
	BookedSessions         []BookedSession        `json:"booked_sessions"`
	BookedPerEntry         []BookedEntryBreakdown `json:"booked_per_entry"`
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
