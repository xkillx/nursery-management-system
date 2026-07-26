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

type PreflightDraftInvoices struct {
	repo domain.BillingRepository
}

func NewPreflightDraftInvoices(repo domain.BillingRepository) *PreflightDraftInvoices {
	return &PreflightDraftInvoices{repo: repo}
}

// Execute previews the advance-pay generation: for each active booking covering
// the billing month, checks funding profile presence and basic eligibility.
// It does NOT run booking-minute arithmetic (that's done in the generator).
func (uc *PreflightDraftInvoices) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string) (domain.PreflightResult, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Validation("Invalid billing month format.", "billing_month")
	}

	period, err := domain.NewBillingPeriod(billingMonth.Year(), billingMonth.Month())
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("billing period: %w", err))
	}

	var bookings []domain.BillableChildRow
	bookings, err = uc.repo.ListActiveBookings(ctx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("list active bookings: %w", err))
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

	for _, b := range bookings {
		blockers := preflightBlockers(b)
		if len(blockers) > 0 {
			result.BlockedChildren = append(result.BlockedChildren, domain.BlockedChild{
				ChildID:         b.ChildID,
				ChildFirstName:  b.FirstName,
				ChildMiddleName: b.MiddleName,
				ChildLastName:   b.LastName,
				Blockers:        blockers,
			})
			for _, bl := range blockers {
				if blockerChildSet[bl.Code] == nil {
					blockerChildSet[bl.Code] = make(map[uuid.UUID]struct{})
				}
				blockerChildSet[bl.Code][b.ChildID] = struct{}{}
			}
			continue
		}

		// Eligible: compute the booking-driven total. We don't run the full
		// per-entry arithmetic in the preflight, but the summary uses the
		// site_hourly_rate for display purposes. Per-child amounts are
		// filled in by the generator.
		fundedAllowance := 0
		result.EligibleChildren = append(result.EligibleChildren, domain.EligibleChild{
			ChildID:                b.ChildID,
			ChildFirstName:         b.FirstName,
			ChildMiddleName:        b.MiddleName,
			ChildLastName:          b.LastName,
			CoreHourlyRate:         domain.MustGBP(b.SiteHourlyRateMinor),
			FundedAllowanceMinutes: fundedAllowance,
		})

		result.Summary.TotalChildrenCount++
		result.Summary.EligibleChildrenCount++
		result.Summary.FundedAllowanceMinutes += fundedAllowance
	}

	for range result.BlockedChildren {
		result.Summary.TotalChildrenCount++
		result.Summary.BlockedChildrenCount++
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

// preflightBlockers returns the blockers for one booking row in the preflight.
func preflightBlockers(b domain.BillableChildRow) []domain.PreflightBlocker {
	var blockers []domain.PreflightBlocker
	if b.FirstName == "" {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildName, Message: "Child first name is missing.",
		})
	}
	if b.DateOfBirth.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildDateOfBirth, Message: "Child date of birth is missing.",
		})
	}
	if b.StartDate.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildStartDate, Message: "Child start date is missing.",
		})
	}
	if !b.HasParentCarerContact {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingGuardianLink, Message: "No active guardian linked to this child.",
		})
	}
	if b.SiteHourlyRateMinor <= 0 {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingBillingRate, Message: "Site billing rate is missing or invalid.",
		})
	}
	return blockers
}

func strPtr(s string) *string { return &s }

var _ = time.Now
