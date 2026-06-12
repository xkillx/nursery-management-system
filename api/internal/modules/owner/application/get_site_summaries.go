package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/domain"
)

type GetSiteSummariesUseCase struct {
	repo domain.SummaryRepository
}

func NewGetSiteSummariesUseCase(repo domain.SummaryRepository) *GetSiteSummariesUseCase {
	return &GetSiteSummariesUseCase{repo: repo}
}

func (uc *GetSiteSummariesUseCase) Execute(ctx context.Context, actor domain.OwnerActor, billingMonthStr string, siteID *uuid.UUID) (domain.SiteSummariesResult, error) {
	billingMonth, err := parseBillingMonth(billingMonthStr)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}

	sites, err := uc.resolveSites(ctx, actor.TenantID, siteID)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}

	branchIDs := siteIDs(sites)
	localDate := londonToday()

	managers, err := uc.repo.CountActiveManagers(ctx, actor.TenantID, branchIDs)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	invites, err := uc.repo.CountPendingManagerInvites(ctx, actor.TenantID, branchIDs)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	children, err := uc.repo.CountActiveChildren(ctx, actor.TenantID, branchIDs)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	attendance, err := uc.repo.CountAttendanceToday(ctx, actor.TenantID, branchIDs, localDate)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	periodStart, periodEnd := domain.BillingMonthRange(billingMonth)
	incomplete, err := uc.repo.CountIncompleteAttendance(ctx, actor.TenantID, branchIDs, periodStart, periodEnd)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	funding, err := uc.repo.GetFundingReadiness(ctx, actor.TenantID, branchIDs, billingMonth)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}
	invoices, err := uc.repo.GetInvoicePaymentHealth(ctx, actor.TenantID, branchIDs, billingMonth)
	if err != nil {
		return domain.SiteSummariesResult{}, err
	}

	summaries := make([]domain.SiteSummary, 0, len(sites))
	var totals domain.SiteSummaryTotals
	var currencyCode string

	for _, s := range sites {
		mc := managers[s.ID]
		ic := invites[s.ID]
		cc := children[s.ID]
		at := attendance[s.ID]
		ia := incomplete[s.ID]
		fr := funding[s.ID]
		iph := invoices[s.ID]

		setupIssues := make([]string, 0, 2)
		if mc == 0 {
			setupIssues = append(setupIssues, "missing_manager")
		}
		if s.CoreHourlyRateMinor == nil || *s.CoreHourlyRateMinor <= 0 {
			setupIssues = append(setupIssues, "missing_site_core_hourly_rate")
		}
		setupStatus := "incomplete_setup"
		if len(setupIssues) == 0 {
			setupStatus = "complete"
		}

		summaries = append(summaries, domain.SiteSummary{
			SiteID:                    s.ID,
			SiteName:                  s.Name,
			SetupStatus:               setupStatus,
			SetupIssues:               setupIssues,
			SiteCoreHourlyRateMinor:   s.CoreHourlyRateMinor,
			ActiveManagerCount:        mc,
			PendingManagerInviteCount: ic,
			ActiveChildrenCount:       cc,
			Attendance: domain.AttendanceSummary{
				CheckedInTodayCount:      at,
				IncompleteAttendanceCount: ia,
			},
			FundingReadiness:     fr,
			InvoicePaymentHealth: iph,
		})

		if iph.CurrencyCode != "" {
			currencyCode = iph.CurrencyCode
		}

		totals.ActiveManagerCount += mc
		totals.PendingManagerInviteCount += ic
		totals.ActiveChildrenCount += cc
		totals.CheckedInTodayCount += at
		totals.IncompleteAttendanceCount += ia
		totals.DraftCount += iph.DraftCount
		totals.IssuedCount += iph.IssuedCount
		totals.OverdueCount += iph.OverdueCount
		totals.PaymentFailedCount += iph.PaymentFailedCount
		totals.PaidCount += iph.PaidCount
		totals.TotalIssuedMinor += iph.TotalIssuedMinor
		totals.TotalPaidMinor += iph.TotalPaidMinor
		totals.OutstandingMinor += iph.OutstandingMinor
		totals.OverdueOutstandingMinor += iph.OverdueOutstandingMinor
	}

	return domain.SiteSummariesResult{
		BillingMonth:        billingMonth.Format("2006-01"),
		AttendanceLocalDate: localDate.Format("2006-01-02"),
		CurrencyCode:        currencyCode,
		Totals:              totals,
		Sites:               summaries,
	}, nil
}

func (uc *GetSiteSummariesUseCase) resolveSites(ctx context.Context, tenantID uuid.UUID, siteID *uuid.UUID) ([]domain.Site, error) {
	if siteID != nil {
		site, err := uc.repo.GetActiveSite(ctx, tenantID, *siteID)
		if err != nil {
			return nil, domain.ErrSiteNotFound
		}
		return []domain.Site{site}, nil
	}
	return uc.repo.GetActiveSites(ctx, tenantID)
}

func parseBillingMonth(s string) (time.Time, error) {
	if s == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	}
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return time.Time{}, &domain.ValidationError{Field: "billing_month", Message: "billing_month must be in YYYY-MM format"}
	}
	return t, nil
}

func siteIDs(sites []domain.Site) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(sites))
	for _, s := range sites {
		ids = append(ids, s.ID)
	}
	return ids
}

func londonToday() time.Time {
	loc, _ := time.LoadLocation("Europe/London")
	return time.Now().In(loc).Truncate(24 * time.Hour)
}

func ResolveBillingMonth(s string) (time.Time, error) {
	return parseBillingMonth(s)
}
