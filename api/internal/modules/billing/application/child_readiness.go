package application

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
)

// ChildReadinessResult holds the per-child blockers and calculation output.
type ChildReadinessResult struct {
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	CoreHourlyRateMinor    *int
	FundingProfileID       *uuid.UUID
	FundedAllowanceMinutes int
	HasGuardianLink        bool
	ExistingInvoiceID      *uuid.UUID
	ExistingInvoiceStatus  *string

	Blockers []domain.PreflightBlocker

	// Calculation fields (set only when no blockers).
	AttendanceCalc         *domain.AttendanceMinuteCalculation
	FundingCalc            *domain.FundingDeductionCalculation
	SubtotalMinor          int
	FundedDeductionMinor   int
	TotalDueMinor          int
	RawAttendedMinutes     int
	RoundedAttendedMinutes int
	IncludedSessionCount   int
}

// EvaluateChildReadiness evaluates blockers and performs calculations for one child.
// Returns a result with Blockers populated if ineligible, or calculation fields populated if eligible.
func EvaluateChildReadiness(
	child domain.PreflightChildRow,
	sessions []domain.AttendanceSessionInput,
	year int,
	month int,
) (ChildReadinessResult, error) {
	result := ChildReadinessResult{
		ChildID:               child.ChildID,
		ChildFirstName:        child.FirstName,
		ChildMiddleName:       child.MiddleName,
		ChildLastName:         child.LastName,
		CoreHourlyRateMinor:   child.CoreHourlyRateMinor,
		FundingProfileID:      child.FundingProfileID,
		HasGuardianLink:       child.HasGuardianLink,
		ExistingInvoiceID:     child.ExistingInvoiceID,
		ExistingInvoiceStatus: child.ExistingInvoiceStatus,
	}

	var blockers []domain.PreflightBlocker

	if child.FirstName == "" {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingChildName,
			Message: "Child first name is missing.",
		})
	}
	if child.DateOfBirth.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingChildDateOfBirth,
			Message: "Child date of birth is missing.",
		})
	}
	if child.StartDate.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingChildStartDate,
			Message: "Child start date is missing.",
		})
	}
	if !child.HasGuardianLink {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingGuardianLink,
			Message: "No active guardian linked to this child.",
		})
	}
	if child.CoreHourlyRateMinor == nil || *child.CoreHourlyRateMinor <= 0 {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingBillingRate,
			Message: "Site billing rate is missing or invalid.",
		})
	}
	if child.FundingProfileID == nil {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingFundingProfile,
			Message: "Funding profile is missing for this billing month.",
			Field:   strPtr("funding_profile"),
		})
	}

	attendanceCalc, calcErr := domain.CalculateAttendanceMinutes(year, time.Month(month), sessions)
	if calcErr != nil {
		return result, fmt.Errorf("attendance calc for child %s: %w", child.ChildID, calcErr)
	}
	for _, excl := range attendanceCalc.ExcludedIncompleteSessions {
		localDateStr := excl.CheckInAt.In(londonLoc).Format("2006-01-02")
		blockers = append(blockers, domain.PreflightBlocker{
			Code:             domain.BlockerIncompleteAttendance,
			Message:          "Attendance session is missing check-out.",
			SessionID:        &excl.SessionID,
			CheckInAt:        &excl.CheckInAt,
			CheckInLocalDate: &localDateStr,
		})
	}

	if child.ExistingInvoiceStatus != nil && *child.ExistingInvoiceStatus != "draft" {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:          domain.BlockerInvoiceAlreadyIssued,
			Message:       "A monthly invoice has already been issued for this child and billing month.",
			InvoiceID:     child.ExistingInvoiceID,
			InvoiceStatus: child.ExistingInvoiceStatus,
		})
	}

	if len(blockers) > 0 {
		result.Blockers = blockers
		return result, nil
	}

	fundedAllowance := 0
	if child.FundedAllowanceMinutes != nil {
		fundedAllowance = *child.FundedAllowanceMinutes
	}
	result.FundedAllowanceMinutes = fundedAllowance

	fundingCalc, fundErr := domain.CalculateFundingDeduction(attendanceCalc, fundedAllowance)
	if fundErr != nil {
		return result, fmt.Errorf("funding deduction for child %s: %w", child.ChildID, fundErr)
	}

	subtotalMinor, subErr := domain.CalculateHourlyAmountMinor(fundingCalc.CoreAttendedMinutes, *child.CoreHourlyRateMinor)
	if subErr != nil {
		return result, fmt.Errorf("subtotal calc: %w", subErr)
	}
	fundedDeductionMinor, dedErr := domain.CalculateHourlyAmountMinor(fundingCalc.FundedDeductionMinutes, *child.CoreHourlyRateMinor)
	if dedErr != nil {
		return result, fmt.Errorf("deduction calc: %w", dedErr)
	}
	totalDueMinor := max(0, subtotalMinor-fundedDeductionMinor)

	result.AttendanceCalc = &attendanceCalc
	result.FundingCalc = &fundingCalc
	result.SubtotalMinor = subtotalMinor
	result.FundedDeductionMinor = fundedDeductionMinor
	result.TotalDueMinor = totalDueMinor
	result.RawAttendedMinutes = attendanceCalc.RawElapsedMinutes
	result.RoundedAttendedMinutes = attendanceCalc.RoundedBillableMinutes
	result.IncludedSessionCount = attendanceCalc.IncludedSessionCount

	return result, nil
}
