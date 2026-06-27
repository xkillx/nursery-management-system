package domain

import (
	"time"

	"github.com/google/uuid"
)

type PreflightChildRow struct {
	ChildID                uuid.UUID
	FirstName              string
	MiddleName             *string
	LastName               *string
	DateOfBirth            time.Time
	StartDate              time.Time
	EndDate                *time.Time
	CoreHourlyRateMinor    *int
	HasParentCarerContact  bool
	FundingProfileID       *uuid.UUID
	FundedAllowanceMinutes *int
	ExistingInvoiceID      *uuid.UUID
	ExistingInvoiceStatus  *string
}

type PreflightAttendanceSessionRow struct {
	SessionID         uuid.UUID
	ChildID           uuid.UUID
	Status            AttendanceSessionStatus
	CheckInAt         time.Time
	CheckOutAt        *time.Time
	CheckInLocalDate  time.Time
	CheckOutLocalDate *time.Time
}

type BlockerCode string

const (
	BlockerMissingChildName        BlockerCode = "missing_child_name"
	BlockerMissingChildDateOfBirth BlockerCode = "missing_child_date_of_birth"
	BlockerMissingChildStartDate   BlockerCode = "missing_child_start_date"
	BlockerMissingGuardianLink     BlockerCode = "missing_guardian_link"
	BlockerMissingBillingRate      BlockerCode = "missing_billing_rate"
	BlockerMissingFundingProfile   BlockerCode = "missing_funding_profile"
	BlockerIncompleteAttendance    BlockerCode = "incomplete_attendance"
	BlockerInvoiceAlreadyIssued    BlockerCode = "invoice_already_issued"
)

var BlockerPriority = []BlockerCode{
	BlockerMissingChildName,
	BlockerMissingChildDateOfBirth,
	BlockerMissingChildStartDate,
	BlockerMissingGuardianLink,
	BlockerMissingBillingRate,
	BlockerMissingFundingProfile,
	BlockerIncompleteAttendance,
	BlockerInvoiceAlreadyIssued,
}

type PreflightBlocker struct {
	Code             BlockerCode `json:"code"`
	Message          string      `json:"message"`
	SessionID        *uuid.UUID  `json:"session_id,omitempty"`
	CheckInAt        *time.Time  `json:"check_in_at,omitempty"`
	CheckInLocalDate *string     `json:"check_in_local_date,omitempty"`
	InvoiceID        *uuid.UUID  `json:"invoice_id,omitempty"`
	InvoiceStatus    *string     `json:"invoice_status,omitempty"`
	Field            *string     `json:"field,omitempty"`
}

type ExistingInvoiceRef struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

type EligibleChild struct {
	ChildID                uuid.UUID           `json:"child_id"`
	ChildFirstName         string              `json:"child_first_name"`
	ChildMiddleName        *string             `json:"child_middle_name"`
	ChildLastName          *string             `json:"child_last_name"`
	CoreHourlyRateMinor    int                 `json:"core_hourly_rate_minor"`
	FundingProfileID       *uuid.UUID          `json:"funding_profile_id"`
	FundedAllowanceMinutes int                 `json:"funded_allowance_minutes"`
	ExistingInvoice        *ExistingInvoiceRef `json:"existing_invoice,omitempty"`
}

type BlockedChild struct {
	ChildID         uuid.UUID          `json:"child_id"`
	ChildFirstName  string             `json:"child_first_name"`
	ChildMiddleName *string            `json:"child_middle_name"`
	ChildLastName   *string            `json:"child_last_name"`
	Blockers        []PreflightBlocker `json:"blockers"`
}

type BlockerCount struct {
	Code          BlockerCode `json:"code"`
	ChildrenCount int         `json:"children_count"`
}

type PreflightResult struct {
	BillingMonth     string           `json:"billing_month"`
	CurrencyCode     string           `json:"currency_code"`
	Period           PreflightPeriod  `json:"period"`
	Summary          PreflightSummary `json:"summary"`
	EligibleChildren []EligibleChild  `json:"eligible_children"`
	BlockedChildren  []BlockedChild   `json:"blocked_children"`
}

type PreflightPeriod struct {
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	EndExclusiveDate string `json:"end_exclusive_date"`
}

type PreflightSummary struct {
	TotalChildrenCount     int            `json:"total_children_count"`
	EligibleChildrenCount  int            `json:"eligible_children_count"`
	BlockedChildrenCount   int            `json:"blocked_children_count"`
	FundedAllowanceMinutes int            `json:"funded_allowance_minutes"`
	BlockerCounts          []BlockerCount `json:"blocker_counts"`
}
