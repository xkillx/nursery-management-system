package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/funding/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ConsumedMinutesProvider interface {
	GetConsumedMinutes(ctx context.Context, tenantID, branchID uuid.UUID, childIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]int, error)
}

type ListOverview struct {
	repo         domain.Repository
	consumedProv ConsumedMinutesProvider
}

func NewListOverview(repo domain.Repository, consumedProv ConsumedMinutesProvider) *ListOverview {
	return &ListOverview{repo: repo, consumedProv: consumedProv}
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

	// Collect child IDs for consumed minutes lookup
	childIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		childIDs = append(childIDs, row.ChildID)
	}

	// Get consumed minutes for all children
	consumedMap, err := uc.consumedProv.GetConsumedMinutes(ctx, actor.TenantID, actor.BranchID, childIDs, billingMonth)
	if err != nil {
		return domain.OverviewResult{}, domainerrors.Internal(err)
	}

	var summary domain.OverviewSummary
	var items []domain.OverviewItem

	summary.IncludedChildCount = len(rows)

	for _, row := range rows {
		flags := computeFlags(row)

		// Compute remaining minutes
		var remaining *int
		if row.FundedAllowanceMinutes != nil {
			consumed := consumedMap[row.ChildID]
			r := max(0, *row.FundedAllowanceMinutes-consumed)
			remaining = &r
		}

		if len(flags) > 0 {
			items = append(items, domain.OverviewItem{Row: row, Flags: flags, RemainingMinutes: remaining})
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

func (uc *ListOverview) ExecutePaginated(ctx context.Context, actor tenant.ActorContext, billingMonthRaw string, limit, offset int) (domain.OverviewResult, int, error) {
	billingMonth, err := ParseBillingMonth(billingMonthRaw)
	if err != nil {
		return domain.OverviewResult{}, 0, domainerrors.Validation("Invalid billing month. Must be YYYY-MM.", "billing_month")
	}

	rows, err := uc.repo.ListOverviewPaginated(ctx, actor.TenantID, actor.BranchID, billingMonth, limit, offset)
	if err != nil {
		return domain.OverviewResult{}, 0, domainerrors.Internal(err)
	}

	total, err := uc.repo.CountOverview(ctx, actor.TenantID, actor.BranchID, billingMonth)
	if err != nil {
		return domain.OverviewResult{}, 0, domainerrors.Internal(err)
	}

	// Collect child IDs for consumed minutes lookup
	childIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		childIDs = append(childIDs, row.ChildID)
	}

	// Get consumed minutes for all children
	consumedMap, err := uc.consumedProv.GetConsumedMinutes(ctx, actor.TenantID, actor.BranchID, childIDs, billingMonth)
	if err != nil {
		return domain.OverviewResult{}, 0, domainerrors.Internal(err)
	}

	var summary domain.OverviewSummary
	var items []domain.OverviewItem

	summary.IncludedChildCount = total

	for _, row := range rows {
		flags := computeFlags(row)

		// Compute remaining minutes
		var remaining *int
		if row.FundedAllowanceMinutes != nil {
			consumed := consumedMap[row.ChildID]
			r := max(0, *row.FundedAllowanceMinutes-consumed)
			remaining = &r
		}

		if len(flags) > 0 {
			items = append(items, domain.OverviewItem{Row: row, Flags: flags, RemainingMinutes: remaining})
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
	}, total, nil
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
