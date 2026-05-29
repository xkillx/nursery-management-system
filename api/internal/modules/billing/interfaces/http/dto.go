package httpbilling

type preflightResponse struct {
	BillingMonth     string                  `json:"billing_month"`
	CurrencyCode     string                  `json:"currency_code"`
	Period           periodResponse          `json:"period"`
	Summary          summaryResponse         `json:"summary"`
	EligibleChildren []eligibleChildResponse `json:"eligible_children"`
	BlockedChildren  []blockedChildResponse  `json:"blocked_children"`
}

type periodResponse struct {
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	EndExclusiveDate string `json:"end_exclusive_date"`
}

type summaryResponse struct {
	TotalChildrenCount     int                    `json:"total_children_count"`
	EligibleChildrenCount  int                    `json:"eligible_children_count"`
	BlockedChildrenCount   int                    `json:"blocked_children_count"`
	IncludedSessionCount   int                    `json:"included_session_count"`
	RawAttendedMinutes     int                    `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                    `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int                    `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                    `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                    `json:"core_billable_minutes"`
	SubtotalMinor          int                    `json:"subtotal_minor"`
	FundedDeductionMinor   int                    `json:"funded_deduction_minor"`
	TotalDueMinor          int                    `json:"total_due_minor"`
	BlockerCounts          []blockerCountResponse `json:"blocker_counts"`
}

type eligibleChildResponse struct {
	ChildID                string              `json:"child_id"`
	ChildName              string              `json:"child_name"`
	CoreHourlyRateMinor    int                 `json:"core_hourly_rate_minor"`
	FundingProfileID       *string             `json:"funding_profile_id"`
	FundedAllowanceMinutes int                 `json:"funded_allowance_minutes"`
	RawAttendedMinutes     int                 `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                 `json:"rounded_attended_minutes"`
	IncludedSessionCount   int                 `json:"included_session_count"`
	FundedDeductionMinutes int                 `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                 `json:"core_billable_minutes"`
	SubtotalMinor          int                 `json:"subtotal_minor"`
	FundedDeductionMinor   int                 `json:"funded_deduction_minor"`
	TotalDueMinor          int                 `json:"total_due_minor"`
	ExistingInvoice        *existingInvoiceRef `json:"existing_invoice"`
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
	Code             string  `json:"code"`
	Message          string  `json:"message"`
	SessionID        *string `json:"session_id,omitempty"`
	CheckInAt        *string `json:"check_in_at,omitempty"`
	CheckInLocalDate *string `json:"check_in_local_date,omitempty"`
	InvoiceID        *string `json:"invoice_id,omitempty"`
	InvoiceStatus    *string `json:"invoice_status,omitempty"`
	Field            *string `json:"field,omitempty"`
}

type blockerCountResponse struct {
	Code          string `json:"code"`
	ChildrenCount int    `json:"children_count"`
}

// --- Draft generation DTOs (API-17) ---

type generateDraftsRequest struct {
	BillingMonth string   `json:"billing_month" binding:"required"`
	ChildIDs     []string `json:"child_ids"`
}

type generateDraftsResponse struct {
	RunID        string                         `json:"run_id"`
	BillingMonth string                         `json:"billing_month"`
	Status       string                         `json:"status"`
	Summary      generateDraftsSummary          `json:"summary"`
	Generated    []generatedDraftResponse       `json:"generated"`
	Blocked      []generateBlockedChildResponse `json:"blocked"`
}

type generateDraftsSummary struct {
	EligibleCount int `json:"eligible_count"`
	SuccessCount  int `json:"success_count"`
	BlockedCount  int `json:"blocked_count"`
	TotalDueMinor int `json:"total_due_minor"`
}

type generatedDraftResponse struct {
	ChildID              string `json:"child_id"`
	ChildName            string `json:"child_name"`
	Action               string `json:"action"`
	InvoiceID            string `json:"invoice_id"`
	SubtotalMinor        int    `json:"subtotal_minor"`
	FundedDeductionMinor int    `json:"funded_deduction_minor"`
	TotalDueMinor        int    `json:"total_due_minor"`
}

type generateBlockedChildResponse struct {
	ChildID   string                    `json:"child_id"`
	ChildName string                    `json:"child_name,omitempty"`
	Blockers  []generateBlockerResponse `json:"blockers"`
}

type generateBlockerResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Manager Invoice Review DTOs (API-18) ---

type invoiceListResponse struct {
	Items  []invoiceListItemResponse `json:"items"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type invoiceListItemResponse struct {
	InvoiceID                  string  `json:"invoice_id"`
	InvoiceKind                string  `json:"invoice_kind"`
	InvoiceNumber              *string `json:"invoice_number"`
	InvoiceNumberDisplay       string  `json:"invoice_number_display"`
	ChildID                    string  `json:"child_id"`
	ChildName                  string  `json:"child_name"`
	BillingMonth               string  `json:"billing_month"`
	Period                     struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	Status                     string  `json:"status"`
	DueStatus                  string  `json:"due_status"`
	CurrencyCode               string  `json:"currency_code"`
	SubtotalMinor              int     `json:"subtotal_minor"`
	FundedDeductionMinor       int     `json:"funded_deduction_minor"`
	TotalDueMinor              int     `json:"total_due_minor"`
	AmountPaidMinor            int     `json:"amount_paid_minor"`
	DueAt                      *string `json:"due_at"`
	IssuedAt                   *string `json:"issued_at"`
	PaidAt                     *string `json:"paid_at"`
	PaymentFailedAt            *string `json:"payment_failed_at"`
	PaymentStatusUpdatedAt     *string `json:"payment_status_updated_at"`
	GeneratedRunID             *string `json:"generated_run_id"`
	GeneratedRunStatus         *string `json:"generated_run_status"`
	GeneratedRunStartedAt      *string `json:"generated_run_started_at"`
	GeneratedRunCompletedAt    *string `json:"generated_run_completed_at"`
	GeneratedRunExceptionCount int     `json:"generated_run_exception_count"`
	CreatedAt                  string  `json:"created_at"`
	UpdatedAt                  string  `json:"updated_at"`
}

type invoiceDetailResponse struct {
	InvoiceID                  string                       `json:"invoice_id"`
	InvoiceKind                string                       `json:"invoice_kind"`
	InvoiceNumber              *string                      `json:"invoice_number"`
	InvoiceNumberDisplay       string                       `json:"invoice_number_display"`
	ChildID                    string                       `json:"child_id"`
	ChildName                  string                       `json:"child_name"`
	BillingMonth               string                       `json:"billing_month"`
	Period                     struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	Status                     string                       `json:"status"`
	DueStatus                  string                       `json:"due_status"`
	CurrencyCode               string                       `json:"currency_code"`
	SubtotalMinor              int                          `json:"subtotal_minor"`
	FundedDeductionMinor       int                          `json:"funded_deduction_minor"`
	TotalDueMinor              int                          `json:"total_due_minor"`
	AmountPaidMinor            int                          `json:"amount_paid_minor"`
	IssuedAt                   *string                      `json:"issued_at"`
	LockedAt                   *string                      `json:"locked_at"`
	DueAt                      *string                      `json:"due_at"`
	PaidAt                     *string                      `json:"paid_at"`
	PaymentFailedAt            *string                      `json:"payment_failed_at"`
	PaymentStatusUpdatedAt     *string                      `json:"payment_status_updated_at"`
	AdjustsInvoiceID           *string                      `json:"adjusts_invoice_id"`
	AdjustmentReasonCode       *string                      `json:"adjustment_reason_code"`
	AdjustmentReasonNote       *string                      `json:"adjustment_reason_note"`
	GeneratedRunID             *string                      `json:"generated_run_id"`
	GeneratedRunStatus         *string                      `json:"generated_run_status"`
	GeneratedRunStartedAt      *string                      `json:"generated_run_started_at"`
	GeneratedRunCompletedAt    *string                      `json:"generated_run_completed_at"`
	GeneratedRunExceptionCount int                          `json:"generated_run_exception_count"`
	GeneratedRunExceptions     []invoiceRunExceptionResponse `json:"generated_run_exceptions"`
	Calculation                invoiceCalculationResponse    `json:"calculation"`
	Lines                      []invoiceLineResponse         `json:"lines"`
	CreatedAt                  string                        `json:"created_at"`
	UpdatedAt                  string                        `json:"updated_at"`
}

type invoiceLineResponse struct {
	LineID                 string  `json:"line_id"`
	LineKind               string  `json:"line_kind"`
	Description            string  `json:"description"`
	SortOrder              int     `json:"sort_order"`
	QuantityMinutes        *int    `json:"quantity_minutes"`
	UnitAmountMinor        *int    `json:"unit_amount_minor"`
	LineAmountMinor        int     `json:"line_amount_minor"`
	RawAttendedMinutes     *int    `json:"raw_attended_minutes"`
	RoundedAttendedMinutes *int    `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes *int    `json:"funded_allowance_minutes"`
	FundedDeductionMinutes *int    `json:"funded_deduction_minutes"`
	CoreBillableMinutes    *int    `json:"core_billable_minutes"`
	SessionCount           *int    `json:"session_count"`
}

type invoiceCalculationResponse struct {
	CoreHourlyRateMinor    int                      `json:"core_hourly_rate_minor"`
	RawAttendedMinutes     int                      `json:"raw_attended_minutes"`
	RoundedAttendedMinutes int                      `json:"rounded_attended_minutes"`
	FundedAllowanceMinutes int                      `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                      `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                      `json:"core_billable_minutes"`
	IncludedSessionCount   int                      `json:"included_session_count"`
	CoreSubtotalMinor      int                      `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int                      `json:"extras_total_minor"`
	SourceSessions         []sourceSessionResponse  `json:"source_sessions"`
}

type sourceSessionResponse struct {
	SessionID              string  `json:"session_id"`
	Status                 string  `json:"status"`
	CheckInAt              string  `json:"check_in_at"`
	CheckOutAt             *string `json:"check_out_at,omitempty"`
	RawElapsedMinutes      int     `json:"raw_elapsed_minutes"`
	RoundedBillableMinutes int     `json:"rounded_billable_minutes"`
}

type invoiceRunExceptionResponse struct {
	ChildID      string   `json:"child_id"`
	ChildName    string   `json:"child_name"`
	BlockerCodes []string `json:"blocker_codes"`
}

// --- Invoice Issue DTOs (API-19) ---

type issueInvoiceRequest struct {
	Confirm bool `json:"confirm"`
}

type issueInvoiceResponse struct {
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
	Status        string `json:"status"`
	IssuedAt      string `json:"issued_at"`
	LockedAt      string `json:"locked_at"`
	DueAt         string `json:"due_at"`
	IssuedRunID   string `json:"issued_run_id"`
	TotalDueMinor int    `json:"total_due_minor"`
}

type bulkIssueInvoicesRequest struct {
	BillingMonth string   `json:"billing_month" binding:"required"`
	InvoiceIDs   []string `json:"invoice_ids"`
	Confirm      bool     `json:"confirm"`
}

type bulkIssueInvoicesResponse struct {
	RunID         string                   `json:"run_id"`
	BillingMonth  string                   `json:"billing_month"`
	Status        string                   `json:"status"`
	Summary       bulkIssueSummary         `json:"summary"`
	Issued        []issuedInvoiceResponse  `json:"issued"`
	Blocked       []blockedInvoiceResponse `json:"blocked"`
}

type bulkIssueSummary struct {
	EligibleCount int `json:"eligible_count"`
	SuccessCount  int `json:"success_count"`
	BlockedCount  int `json:"blocked_count"`
	TotalDueMinor int `json:"total_due_minor"`
}

type issuedInvoiceResponse struct {
	InvoiceID     string `json:"invoice_id"`
	ChildID       string `json:"child_id"`
	ChildName     string `json:"child_name"`
	InvoiceNumber string `json:"invoice_number"`
	IssuedAt      string `json:"issued_at"`
	DueAt         string `json:"due_at"`
	TotalDueMinor int    `json:"total_due_minor"`
}

type blockedInvoiceResponse struct {
	InvoiceID string                      `json:"invoice_id"`
	ChildID   *string                     `json:"child_id,omitempty"`
	ChildName string                      `json:"child_name,omitempty"`
	Blockers  []issueBlockerResponse      `json:"blockers"`
}

type issueBlockerResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
