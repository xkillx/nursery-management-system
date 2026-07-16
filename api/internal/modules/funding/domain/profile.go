package domain

import (
	"time"

	"github.com/google/uuid"
)

type FundingProfile struct {
	ID                     uuid.UUID
	TenantID               uuid.UUID
	BranchID               uuid.UUID
	ChildID                uuid.UUID
	BillingMonth           time.Time
	FundedAllowanceMinutes int
	FundingType            *string
	FundingModel           *string
	FundedHoursPerWeek     *float64
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ChildEnrollment struct {
	ID        uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

type OverviewFlag string

const (
	FlagMissingProfile OverviewFlag = "missing_profile"
	FlagExplicitZero   OverviewFlag = "explicit_zero_allowance"
	FlagUnderOneHour   OverviewFlag = "under_one_hour_allowance"
	FlagAbove160Hours  OverviewFlag = "above_160_hours_allowance"
)

type OverviewRow struct {
	ChildID                uuid.UUID
	ChildFirstName         string
	ChildMiddleName        *string
	ChildLastName          *string
	IsActive               bool
	StartDate              time.Time
	EndDate                *time.Time
	FundingProfileID       *uuid.UUID
	FundedAllowanceMinutes *int
	FundingUpdatedAt       *time.Time
	ChildPhotoPath         *string
}

type OverviewSummary struct {
	IncludedChildCount  int
	FlaggedChildCount   int
	MissingProfileCount int
	ExplicitZeroCount   int
	UnderOneHourCount   int
	Above160HoursCount  int
}

type OverviewResult struct {
	BillingMonth time.Time
	Summary      OverviewSummary
	Items        []OverviewItem
}

type OverviewItem struct {
	Row              OverviewRow
	Flags            []OverviewFlag
	RemainingMinutes *int
}

type ExpiringFundingRecord struct {
	FundingRecordID    uuid.UUID
	ChildID            uuid.UUID
	ChildFirstName     string
	ChildMiddleName    *string
	ChildLastName      *string
	FundingType        *string
	FundedHoursPerWeek *float64
	FundingEndDate     time.Time
}

type EnhancedOverviewMetrics struct {
	TotalFundedChildren int
	FifteenHourCount    int
	ThirtyHourCount     int
	BookedHoursThisWeek float64
	ExpiringSoonCount   int
}

type AllocationEntry struct {
	BookingID              uuid.UUID
	EffectiveStartDate     time.Time
	EffectiveEndDate       *time.Time
	DaysOfWeek             []int32
	SessionTypeName        string
	SessionDurationMinutes int
}

type EnhancedChildDetail struct {
	Profile    FundingProfile
	Allocation []AllocationEntry
	History    []FundingHistory
}
