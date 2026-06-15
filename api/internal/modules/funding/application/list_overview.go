package application

import (
	"context"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ListOverview struct {
	repo domain.Repository
}

func NewListOverview(repo domain.Repository) *ListOverview {
	return &ListOverview{repo: repo}
}

func (uc *ListOverview) Execute(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string) (domain.OverviewResult, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.OverviewResult{}, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	rows, err := uc.repo.ListOverview(ctx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return domain.OverviewResult{}, domainerrors.Internal(err)
	}

	var summary domain.OverviewSummary
	var items []domain.OverviewItem

	summary.IncludedChildCount = len(rows)

	for _, row := range rows {
		flags := computeFlags(row)
		if len(flags) > 0 {
			items = append(items, domain.OverviewItem{Row: row, Flags: flags})
			summary.FlaggedChildCount++
			for _, f := range flags {
				switch f {
				case domain.FlagMissingProfile:
					summary.MissingProfileCount++
				case domain.FlagExplicitZero:
					summary.ExplicitZeroCount++
				case domain.FlagUnderOneHour:
					summary.UnderOneHourCount++
				case domain.FlagAbove160Hours:
					summary.Above160HoursCount++
				}
			}
		}
	}

	if items == nil {
		items = []domain.OverviewItem{}
	}

	return domain.OverviewResult{
		BillingMonth: billingMonth,
		Summary:      summary,
		Items:        items,
	}, nil
}

func computeFlags(row domain.OverviewRow) []domain.OverviewFlag {
	if row.FundingProfileID == nil {
		return []domain.OverviewFlag{domain.FlagMissingProfile}
	}

	var flags []domain.OverviewFlag
	allowance := *row.FundedAllowanceMinutes

	if allowance == 0 {
		flags = append(flags, domain.FlagExplicitZero)
	} else if allowance > 0 && allowance < 60 {
		flags = append(flags, domain.FlagUnderOneHour)
	}

	if allowance > 9600 {
		flags = append(flags, domain.FlagAbove160Hours)
	}

	return flags
}
