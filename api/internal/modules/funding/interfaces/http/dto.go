package httpfunding

import "time"

type fundingProfileRequest struct {
	BillingMonth           string `json:"billing_month" binding:"required"`
	FundedAllowanceMinutes int    `json:"funded_allowance_minutes" binding:"required"`
}

type fundingProfileResponse struct {
	ID                     string    `json:"id"`
	ChildID                string    `json:"child_id"`
	BillingMonth           string    `json:"billing_month"`
	FundedAllowanceMinutes int       `json:"funded_allowance_minutes"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
