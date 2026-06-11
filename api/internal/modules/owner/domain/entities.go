package domain

import (
	"time"

	"github.com/google/uuid"
)

type OwnerActor struct {
	UserID       uuid.UUID
	MembershipID uuid.UUID
	TenantID     uuid.UUID
}

type Site struct {
	ID   uuid.UUID
	Name string
}

type AttendanceSummary struct {
	CheckedInTodayCount    int
	IncompleteAttendanceCount int
}

type FundingReadiness struct {
	IncludedChildCount int
	MissingProfileCount int
	ExplicitZeroCount   int
	UnderOneHourCount   int
	Above160HoursCount  int
}

func (f FundingReadiness) FlaggedCount() int {
	return f.MissingProfileCount + f.ExplicitZeroCount + f.UnderOneHourCount + f.Above160HoursCount
}

type InvoicePaymentHealth struct {
	CurrencyCode          string
	DraftCount            int
	IssuedCount           int
	OverdueCount          int
	PaymentFailedCount    int
	PaidCount             int
	TotalIssuedMinor      int64
	TotalPaidMinor        int64
	OutstandingMinor      int64
	OverdueOutstandingMinor int64
}

type SiteSummary struct {
	SiteID                     uuid.UUID
	SiteName                   string
	SetupStatus                string
	ActiveManagerCount         int
	PendingManagerInviteCount  int
	ActiveChildrenCount        int
	Attendance                 AttendanceSummary
	FundingReadiness           FundingReadiness
	InvoicePaymentHealth       InvoicePaymentHealth
}

type SiteSummariesResult struct {
	BillingMonth        string
	AttendanceLocalDate string
	CurrencyCode        string
	Totals              SiteSummaryTotals
	Sites               []SiteSummary
}

type SiteSummaryTotals struct {
	ActiveManagerCount        int
	PendingManagerInviteCount int
	ActiveChildrenCount       int
	CheckedInTodayCount       int
	IncompleteAttendanceCount int
	DraftCount                int
	IssuedCount               int
	OverdueCount              int
	PaymentFailedCount        int
	PaidCount                 int
	TotalIssuedMinor          int64
	TotalPaidMinor            int64
	OutstandingMinor          int64
	OverdueOutstandingMinor   int64
}

func BillingMonthRange(billingMonth time.Time) (start, end time.Time) {
	start = time.Date(billingMonth.Year(), billingMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, 0)
	return
}
