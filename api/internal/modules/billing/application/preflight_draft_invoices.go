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

// Execute previews the advance-pay generation: for each active Term covering
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

	var terms []domain.AdvancePayTermRow
	terms, err = uc.repo.ListActiveTerms(ctx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return domain.PreflightResult{}, domainerrors.Internal(fmt.Errorf("list active terms: %w", err))
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

	for _, t := range terms {
		blockers := preflightBlockers(t)
		if len(blockers) > 0 {
			result.BlockedChildren = append(result.BlockedChildren, domain.BlockedChild{
				ChildID:         t.ChildID,
				ChildFirstName:  t.FirstName,
				ChildMiddleName: t.MiddleName,
				ChildLastName:   t.LastName,
				Blockers:        blockers,
			})
			for _, b := range blockers {
				if blockerChildSet[b.Code] == nil {
					blockerChildSet[b.Code] = make(map[uuid.UUID]struct{})
				}
				blockerChildSet[b.Code][t.ChildID] = struct{}{}
			}
			continue
		}

		// Eligible: compute the booking-driven total. We don't run the full
		// per-entry arithmetic in the preflight, but the summary uses the
		// Term's site_hourly_rate for display purposes. Per-child amounts are
		// filled in by the generator.
		fundedAllowance := 0
		if t.FundedAllowanceMinutes != nil {
			fundedAllowance = *t.FundedAllowanceMinutes
		}
		result.EligibleChildren = append(result.EligibleChildren, domain.EligibleChild{
			ChildID:                t.ChildID,
			ChildFirstName:         t.FirstName,
			ChildMiddleName:        t.MiddleName,
			ChildLastName:          t.LastName,
			CoreHourlyRateMinor:    t.SiteHourlyRateMinor,
			FundingProfileID:       t.FundingProfileID,
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

// preflightBlockers returns the blockers for one term row in the preflight.
func preflightBlockers(t domain.AdvancePayTermRow) []domain.PreflightBlocker {
	var blockers []domain.PreflightBlocker
	if t.FirstName == "" {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildName, Message: "Child first name is missing.",
		})
	}
	if t.DateOfBirth.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildDateOfBirth, Message: "Child date of birth is missing.",
		})
	}
	if t.StartDate.IsZero() {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingChildStartDate, Message: "Child start date is missing.",
		})
	}
	if !t.HasGuardianLink {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingGuardianLink, Message: "No active guardian linked to this child.",
		})
	}
	if t.SiteHourlyRateMinor <= 0 {
		blockers = append(blockers, domain.PreflightBlocker{
			Code: domain.BlockerMissingBillingRate, Message: "Site billing rate is missing or invalid.",
		})
	}
	if t.FundingProfileID == nil {
		blockers = append(blockers, domain.PreflightBlocker{
			Code:    domain.BlockerMissingFundingProfile,
			Message: "Funding profile is missing for this billing month.",
			Field:   strPtr("funding_profile"),
		})
	}
	return blockers
}

func strPtr(s string) *string { return &s }

var _ = time.Now
