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
}
