package httpbilling

type preflightResponse struct {
	BillingMonth     string                   `json:"billing_month"`
	CurrencyCode     string                   `json:"currency_code"`
	Period           periodResponse           `json:"period"`
	Summary          summaryResponse          `json:"summary"`
	EligibleChildren []eligibleChildResponse  `json:"eligible_children"`
	BlockedChildren  []blockedChildResponse   `json:"blocked_children"`
}

type periodResponse struct {
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	EndExclusiveDate string `json:"end_exclusive_date"`
}

type summaryResponse struct {
	TotalChildrenCount     int                   `json:"total_children_count"`
	EligibleChildrenCount  int                   `json:"eligible_children_count"`
	BlockedChildrenCount   int                   `json:"blocked_children_count"`
	IncludedSessionCount   int                   `json:"included_session_count"`
	RawAttendedMinutes     int                   `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                   `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int                   `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                   `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                   `json:"core_billable_minutes"`
	SubtotalMinor          int                   `json:"subtotal_minor"`
	FundedDeductionMinor   int                   `json:"funded_deduction_minor"`
	TotalDueMinor          int                   `json:"total_due_minor"`
	BlockerCounts          []blockerCountResponse `json:"blocker_counts"`
}

type eligibleChildResponse struct {
	ChildID                string               `json:"child_id"`
	ChildName              string               `json:"child_name"`
	CoreHourlyRateMinor    int                  `json:"core_hourly_rate_minor"`
	FundingProfileID       *string              `json:"funding_profile_id"`
	FundedAllowanceMinutes int                  `json:"funded_allowance_minutes"`
	RawAttendedMinutes     int                  `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                  `json:"rounded_attended_minutes"`
	IncludedSessionCount   int                  `json:"included_session_count"`
	FundedDeductionMinutes int                  `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                  `json:"core_billable_minutes"`
	SubtotalMinor          int                  `json:"subtotal_minor"`
	FundedDeductionMinor   int                  `json:"funded_deduction_minor"`
	TotalDueMinor          int                  `json:"total_due_minor"`
	ExistingInvoice        *existingInvoiceRef  `json:"existing_invoice"`
}

type existingInvoiceRef struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type blockedChildResponse struct {
	ChildID   string            `json:"child_id"`
	ChildName string            `json:"child_name"`
	Blockers  []blockerResponse `json:"blockers"`
}

type blockerResponse struct {
	Code             string   `json:"code"`
	Message          string   `json:"message"`
	SessionID        *string  `json:"session_id,omitempty"`
	CheckInAt        *string  `json:"check_in_at,omitempty"`
	CheckInLocalDate *string  `json:"check_in_local_date,omitempty"`
	InvoiceID        *string  `json:"invoice_id,omitempty"`
	InvoiceStatus    *string  `json:"invoice_status,omitempty"`
	Field            *string  `json:"field,omitempty"`
}

type blockerCountResponse struct {
	Code          string `json:"code"`
	ChildrenCount int    `json:"children_count"`
}
