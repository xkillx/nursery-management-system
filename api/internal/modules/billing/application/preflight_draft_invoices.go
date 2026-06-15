package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
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
		childSessions := sessionsByChild[child.ChildID]
		readiness, readinessErr := EvaluateChildReadiness(child, childSessions, year, int(month))
		if readinessErr != nil {
			return domain.PreflightResult{}, domainerrors.Internal(readinessErr)
		}

		if len(readiness.Blockers) > 0 {
			result.BlockedChildren = append(result.BlockedChildren, domain.BlockedChild{
				ChildID:         child.ChildID,
				ChildFirstName:  child.FirstName,
				ChildMiddleName: child.MiddleName,
				ChildLastName:   child.LastName,
				Blockers:        readiness.Blockers,
			})
			for _, b := range readiness.Blockers {
				if blockerChildSet[b.Code] == nil {
					blockerChildSet[b.Code] = make(map[uuid.UUID]struct{})
				}
				blockerChildSet[b.Code][child.ChildID] = struct{}{}
			}
			continue
		}

		var existingInvoice *domain.ExistingInvoiceRef
		if child.ExistingInvoiceID != nil && child.ExistingInvoiceStatus != nil && *child.ExistingInvoiceStatus == "draft" {
			existingInvoice = &domain.ExistingInvoiceRef{
				ID:     *child.ExistingInvoiceID,
				Status: *child.ExistingInvoiceStatus,
			}
		}

		eligible := domain.EligibleChild{
			ChildID:                child.ChildID,
			ChildFirstName:         child.FirstName,
			ChildMiddleName:        child.MiddleName,
			ChildLastName:          child.LastName,
			CoreHourlyRateMinor:    *child.CoreHourlyRateMinor,
			FundingProfileID:       child.FundingProfileID,
			FundedAllowanceMinutes: readiness.FundedAllowanceMinutes,
			RawAttendedMinutes:     readiness.RawAttendedMinutes,
			RoundedAttendedMinutes: readiness.RoundedAttendedMinutes,
			IncludedSessionCount:   readiness.IncludedSessionCount,
			FundedDeductionMinutes: readiness.FundingCalc.FundedDeductionMinutes,
			CoreBillableMinutes:    readiness.FundingCalc.CoreBillableMinutes,
			SubtotalMinor:          readiness.SubtotalMinor,
			FundedDeductionMinor:   readiness.FundedDeductionMinor,
			TotalDueMinor:          readiness.TotalDueMinor,
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
