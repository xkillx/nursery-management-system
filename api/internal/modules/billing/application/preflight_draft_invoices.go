package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

var londonLoc *time.Location

func init() {
	var err error
	londonLoc, err = time.LoadLocation("Europe/London")
	if err != nil {
		panic("billing app: failed to load Europe/London: " + err.Error())
	}
}

type PreflightDraftInvoices struct {
	repo domain.BillingRepository
}

func NewPreflightDraftInvoices(repo domain.BillingRepository) *PreflightDraftInvoices {
	return &PreflightDraftInvoices{repo: repo}
}

func (uc *PreflightDraftInvoices) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string) (domain.PreflightResult, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	year := billingMonth.Year()
	month := billingMonth.Month()

	period, err := domain.NewBillingPeriod(year, month)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("billing period: %w", err))
	}

	nextMonth := month + 1
	nextYear := year
	if nextMonth > time.December {
		nextMonth = time.January
		nextYear++
	}
	nextBillingMonth := time.Date(nextYear, nextMonth, 1, 0, 0, 0, 0, time.UTC)

	children, err := uc.repo.ListPreflightChildren(ctx, actor.TenantID, actor.BranchID, billingMonth, nextBillingMonth)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("list preflight children: %w", err))
	}

	sessions, err := uc.repo.ListPreflightAttendanceSessions(ctx, actor.TenantID, actor.BranchID, period.StartLocal, period.EndExclusiveLocal)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("list preflight attendance: %w", err))
	}

	sessionsByChild := make(map[uuid.UUID][]domain.AttendanceSessionInput)
	for _, s := range sessions {
		sessionsByChild[s.ChildID] = append(sessionsByChild[s.ChildID], domain.AttendanceSessionInput{
			SessionID:  s.SessionID,
			Status:     s.Status,
			CheckInAt:  s.CheckInAt,
			CheckOutAt: s.CheckOutAt,
		})
	}

	result := domain.PreflightResult{
		BillingMonth: billingMonthRaw,
		CurrencyCode: "GBP",
		Period: domain.PreflightPeriod{
			StartDate:        period.StartLocal.Format("2006-01-02"),
			EndDate:          period.EndExclusiveLocal.AddDate(0, 0, -1).Format("2006-01-02"),
			EndExclusiveDate: period.EndExclusiveLocal.Format("2006-01-02"),
		},
	}

	blockerChildSet := make(map[domain.BlockerCode]map[uuid.UUID]struct{})

	for _, child := range children {
		var blockers []domain.PreflightBlocker

		if child.FullName == "" {
			blockers = append(blockers, domain.PreflightBlocker{
				Code:    domain.BlockerMissingChildName,
				Message: "Child full name is missing.",
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
		if child.CoreHourlyRateMinor < 0 {
			blockers = append(blockers, domain.PreflightBlocker{
				Code:    domain.BlockerMissingBillingRate,
				Message: "Billing rate is missing or invalid.",
			})
		}

		if child.FundingProfileID == nil {
			blockers = append(blockers, domain.PreflightBlocker{
				Code:    domain.BlockerMissingFundingProfile,
				Message: "Funding profile is missing for this billing month.",
				Field:   strPtr("funding_profile"),
			})
		}

		childSessions := sessionsByChild[child.ChildID]
		attendanceCalc, calcErr := domain.CalculateAttendanceMinutes(year, month, childSessions)
		if calcErr != nil {
			return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("attendance calc for child %s: %w", child.ChildID, calcErr))
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
			result.BlockedChildren = append(result.BlockedChildren, domain.BlockedChild{
				ChildID:   child.ChildID,
				ChildName: child.FullName,
				Blockers:  blockers,
			})
			for _, b := range blockers {
				if blockerChildSet[b.Code] == nil {
					blockerChildSet[b.Code] = make(map[uuid.UUID]struct{})
				}
				blockerChildSet[b.Code][child.ChildID] = struct{}{}
			}
			continue
		}

		fundedAllowance := 0
		if child.FundedAllowanceMinutes != nil {
			fundedAllowance = *child.FundedAllowanceMinutes
		}

		fundingCalc, fundErr := domain.CalculateFundingDeduction(attendanceCalc, fundedAllowance)
		if fundErr != nil {
			return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("funding deduction for child %s: %w", child.ChildID, fundErr))
		}

		subtotalMinor, subErr := domain.CalculateHourlyAmountMinor(fundingCalc.CoreAttendedMinutes, child.CoreHourlyRateMinor)
		if subErr != nil {
			return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("subtotal calc: %w", subErr))
		}
		fundedDeductionMinor, dedErr := domain.CalculateHourlyAmountMinor(fundingCalc.FundedDeductionMinutes, child.CoreHourlyRateMinor)
		if dedErr != nil {
			return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("deduction calc: %w", dedErr))
		}
		totalDueMinor := max(0, subtotalMinor-fundedDeductionMinor)

		var existingInvoice *domain.ExistingInvoiceRef
		if child.ExistingInvoiceID != nil && child.ExistingInvoiceStatus != nil && *child.ExistingInvoiceStatus == "draft" {
			existingInvoice = &domain.ExistingInvoiceRef{
				ID:     *child.ExistingInvoiceID,
				Status: *child.ExistingInvoiceStatus,
			}
		}

		eligible := domain.EligibleChild{
			ChildID:                child.ChildID,
			ChildName:              child.FullName,
			CoreHourlyRateMinor:    child.CoreHourlyRateMinor,
			FundingProfileID:       child.FundingProfileID,
			FundedAllowanceMinutes: fundedAllowance,
			RawAttendedMinutes:     attendanceCalc.RawElapsedMinutes,
			RoundedAttendedMinutes: attendanceCalc.RoundedBillableMinutes,
			IncludedSessionCount:   attendanceCalc.IncludedSessionCount,
			FundedDeductionMinutes: fundingCalc.FundedDeductionMinutes,
			CoreBillableMinutes:    fundingCalc.CoreBillableMinutes,
			SubtotalMinor:          subtotalMinor,
			FundedDeductionMinor:   fundedDeductionMinor,
			TotalDueMinor:          totalDueMinor,
			ExistingInvoice:        existingInvoice,
		}

		result.EligibleChildren = append(result.EligibleChildren, eligible)

		result.Summary.TotalChildrenCount++
		result.Summary.EligibleChildrenCount++
		result.Summary.IncludedSessionCount += eligible.IncludedSessionCount
		result.Summary.RawAttendedMinutes += eligible.RawAttendedMinutes
		result.Summary.RoundedAttendedMinutes += eligible.RoundedAttendedMinutes
		result.Summary.FundedAllowanceMinutes += eligible.FundedAllowanceMinutes
		result.Summary.FundedDeductionMinutes += eligible.FundedDeductionMinutes
		result.Summary.CoreBillableMinutes += eligible.CoreBillableMinutes
		result.Summary.SubtotalMinor += eligible.SubtotalMinor
		result.Summary.FundedDeductionMinor += eligible.FundedDeductionMinor
		result.Summary.TotalDueMinor += eligible.TotalDueMinor
	}

	for _, blocked := range result.BlockedChildren {
		result.Summary.TotalChildrenCount++
		result.Summary.BlockedChildrenCount++
		_ = blocked
	}

	for _, code := range domain.BlockerPriority {
		childSet, ok := blockerChildSet[code]
		if !ok || len(childSet) == 0 {
			continue
		}
		result.Summary.BlockerCounts = append(result.Summary.BlockerCounts, domain.BlockerCount{
			Code:          code,
			ChildrenCount: len(childSet),
		})
	}

	return result, nil
}

func strPtr(s string) *string { return &s }
