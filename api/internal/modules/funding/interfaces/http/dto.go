package httpfunding

import "time"

type fundingProfileRequest struct {
	BillingMonth           string `json:"billing_month" binding:"required"`
	FundedAllowanceMinutes int    `json:"funded_allowance_minutes" binding:"min=0"`
}

type fundingProfileResponse struct {
	ID                     string    `json:"id"`
	ChildID                string    `json:"child_id"`
	BillingMonth           string    `json:"billing_month"`
	FundedAllowanceMinutes int       `json:"funded_allowance_minutes"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type overviewResponse struct {
	BillingMonth string                  `json:"billing_month"`
	Summary      overviewSummaryResponse `json:"summary"`
	Items        []overviewItemResponse  `json:"items"`
}

type overviewSummaryResponse struct {
	IncludedChildCount  int `json:"included_child_count"`
	FlaggedChildCount   int `json:"flagged_child_count"`
	MissingProfileCount int `json:"missing_profile_count"`
	ExplicitZeroCount   int `json:"explicit_zero_count"`
	UnderOneHourCount   int `json:"under_one_hour_count"`
	Above160HoursCount  int `json:"above_160_hours_count"`
}

type overviewItemResponse struct {
	ChildID                string     `json:"child_id"`
	ChildFirstName         string     `json:"child_first_name"`
	ChildMiddleName        *string    `json:"child_middle_name"`
	ChildLastName          *string    `json:"child_last_name"`
	IsActive               bool       `json:"is_active"`
	StartDate              time.Time  `json:"start_date"`
	EndDate                *time.Time `json:"end_date,omitempty"`
	FundingProfileID       string     `json:"funding_profile_id,omitempty"`
	FundedAllowanceMinutes *int       `json:"funded_allowance_minutes"`
	FundingUpdatedAt       *time.Time `json:"funding_updated_at"`
	ChildPhotoURL          *string    `json:"photo_url,omitempty"`
	Flags                  []string   `json:"flags"`
	RemainingMinutes       *int       `json:"remaining_minutes"`
}

type enhancedOverviewMetricsResponse struct {
	TotalFundedChildren int     `json:"total_funded_children"`
	FifteenHourCount    int     `json:"fifteen_hour_count"`
	ThirtyHourCount     int     `json:"thirty_hour_count"`
	BookedHoursThisWeek float64 `json:"booked_hours_this_week"`
	ExpiringSoonCount   int     `json:"expiring_soon_count"`
}

type expiringFundingResponse struct {
	FundingRecordID    string   `json:"funding_record_id"`
	ChildID            string   `json:"child_id"`
	ChildFirstName     string   `json:"child_first_name"`
	ChildMiddleName    *string  `json:"child_middle_name,omitempty"`
	ChildLastName      *string  `json:"child_last_name,omitempty"`
	FundingType        *string  `json:"funding_type,omitempty"`
	FundedHoursPerWeek *float64 `json:"funded_hours_per_week,omitempty"`
	FundingEndDate     string   `json:"funding_end_date"`
}

type allocationEntryResponse struct {
	BookingID              string  `json:"booking_id"`
	EffectiveStartDate     string  `json:"effective_start_date"`
	EffectiveEndDate       *string `json:"effective_end_date,omitempty"`
	DaysOfWeek             []int32 `json:"days_of_week"`
	SessionTypeName        string  `json:"session_type_name"`
	SessionDurationMinutes int     `json:"session_duration_minutes"`
}

type fundingHistoryResponse struct {
	ID                 string    `json:"id"`
	FundingType        *string   `json:"funding_type,omitempty"`
	FundingModel       *string   `json:"funding_model,omitempty"`
	FundedHoursPerWeek *float64  `json:"funded_hours_per_week,omitempty"`
	FundingStartDate   *string   `json:"funding_start_date,omitempty"`
	FundingEndDate     *string   `json:"funding_end_date,omitempty"`
	ChangedAt          time.Time `json:"changed_at"`
}

type enhancedChildDetailResponse struct {
	Profile    fundingProfileResponse    `json:"profile"`
	Allocation []allocationEntryResponse `json:"allocation"`
	History    []fundingHistoryResponse  `json:"history"`
}
