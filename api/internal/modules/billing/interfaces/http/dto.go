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
	FundedAllowanceMinutes int                    `json:"funded_allowance_minutes"`
	BlockerCounts          []blockerCountResponse `json:"blocker_counts"`
}

type eligibleChildResponse struct {
	ChildID                string              `json:"child_id"`
	ChildFirstName         string              `json:"child_first_name"`
	ChildMiddleName        *string             `json:"child_middle_name"`
	ChildLastName          *string             `json:"child_last_name"`
	CoreHourlyRateMinor    int                 `json:"core_hourly_rate_minor"`
	FundingProfileID       *string             `json:"funding_profile_id"`
	FundedAllowanceMinutes int                 `json:"funded_allowance_minutes"`
	ExistingInvoice        *existingInvoiceRef `json:"existing_invoice,omitempty"`
}

type existingInvoiceRef struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type blockedChildResponse struct {
	ChildID         string            `json:"child_id"`
	ChildFirstName  string            `json:"child_first_name"`
	ChildMiddleName *string           `json:"child_middle_name"`
	ChildLastName   *string           `json:"child_last_name"`
	Blockers        []blockerResponse `json:"blockers"`
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
	ChildID              string  `json:"child_id"`
	ChildFirstName       string  `json:"child_first_name"`
	ChildMiddleName      *string `json:"child_middle_name"`
	ChildLastName        *string `json:"child_last_name"`
	Action               string  `json:"action"`
	InvoiceID            string  `json:"invoice_id"`
	SubtotalMinor        int     `json:"subtotal_minor"`
	FundedDeductionMinor int     `json:"funded_deduction_minor"`
	TotalDueMinor        int     `json:"total_due_minor"`
}

type generateBlockedChildResponse struct {
	ChildID         string                    `json:"child_id"`
	ChildFirstName  string                    `json:"child_first_name,omitempty"`
	ChildMiddleName *string                   `json:"child_middle_name"`
	ChildLastName   *string                   `json:"child_last_name"`
	Blockers        []generateBlockerResponse `json:"blockers"`
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
	InvoiceID            string  `json:"invoice_id"`
	InvoiceKind          string  `json:"invoice_kind"`
	InvoiceNumber        *string `json:"invoice_number"`
	InvoiceNumberDisplay string  `json:"invoice_number_display"`
	ChildID              string  `json:"child_id"`
	ChildFirstName       string  `json:"child_first_name"`
	ChildMiddleName      *string `json:"child_middle_name"`
	ChildLastName        *string `json:"child_last_name"`
	BillingMonth         string  `json:"billing_month"`
	Period               struct {
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
	InvoiceID            string  `json:"invoice_id"`
	InvoiceKind          string  `json:"invoice_kind"`
	InvoiceNumber        *string `json:"invoice_number"`
	InvoiceNumberDisplay string  `json:"invoice_number_display"`
	ChildID              string  `json:"child_id"`
	ChildFirstName       string  `json:"child_first_name"`
	ChildMiddleName      *string `json:"child_middle_name"`
	ChildLastName        *string `json:"child_last_name"`
	BillingMonth         string  `json:"billing_month"`
	Period               struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	Status                     string                        `json:"status"`
	DueStatus                  string                        `json:"due_status"`
	CurrencyCode               string                        `json:"currency_code"`
	SubtotalMinor              int                           `json:"subtotal_minor"`
	FundedDeductionMinor       int                           `json:"funded_deduction_minor"`
	TotalDueMinor              int                           `json:"total_due_minor"`
	AmountPaidMinor            int                           `json:"amount_paid_minor"`
	IssuedAt                   *string                       `json:"issued_at"`
	LockedAt                   *string                       `json:"locked_at"`
	DueAt                      *string                       `json:"due_at"`
	PaidAt                     *string                       `json:"paid_at"`
	PaymentFailedAt            *string                       `json:"payment_failed_at"`
	PaymentStatusUpdatedAt     *string                       `json:"payment_status_updated_at"`
	AdjustsInvoiceID           *string                       `json:"adjusts_invoice_id"`
	AdjustmentReasonCode       *string                       `json:"adjustment_reason_code"`
	AdjustmentReasonNote       *string                       `json:"adjustment_reason_note"`
	GeneratedRunID             *string                       `json:"generated_run_id"`
	GeneratedRunStatus         *string                       `json:"generated_run_status"`
	GeneratedRunStartedAt      *string                       `json:"generated_run_started_at"`
	GeneratedRunCompletedAt    *string                       `json:"generated_run_completed_at"`
	GeneratedRunExceptionCount int                           `json:"generated_run_exception_count"`
	GeneratedRunExceptions     []invoiceRunExceptionResponse `json:"generated_run_exceptions"`
	Calculation                invoiceCalculationResponse    `json:"calculation"`
	Lines                      []invoiceLineResponse         `json:"lines"`
	CreatedAt                  string                        `json:"created_at"`
	UpdatedAt                  string                        `json:"updated_at"`
}

type invoiceLineResponse struct {
	LineID                 string `json:"line_id"`
	LineKind               string `json:"line_kind"`
	Description            string `json:"description"`
	SortOrder              int    `json:"sort_order"`
	QuantityMinutes        *int   `json:"quantity_minutes"`
	UnitAmountMinor        *int   `json:"unit_amount_minor"`
	LineAmountMinor        int    `json:"line_amount_minor"`
	FundedAllowanceMinutes *int   `json:"funded_allowance_minutes"`
	FundedDeductionMinutes *int   `json:"funded_deduction_minutes"`
	CoreBillableMinutes    *int   `json:"core_billable_minutes"`
	SessionCount           *int   `json:"session_count"`
}

type invoiceCalculationResponse struct {
	CoreHourlyRateMinor    int                     `json:"core_hourly_rate_minor"`
	BookedCoreMinutes      int                     `json:"booked_core_minutes"`
	BookedSessionCount     int                     `json:"booked_session_count"`
	FundedAllowanceMinutes int                     `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int                     `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int                     `json:"core_billable_minutes"`
	CoreSubtotalMinor      int                     `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int                     `json:"extras_total_minor"`
	TermID                 string                  `json:"term_id"`
	BookingPatternID       string                  `json:"booking_pattern_id"`
	BookedSessions         []bookedSessionResponse `json:"booked_sessions"`
	BookedPerEntry         []bookedEntryResponse   `json:"booked_per_entry"`
}

type bookedSessionResponse struct {
	DayOfWeek       int    `json:"day_of_week"`
	OccurrenceDate  string `json:"occurrence_date"`
	DurationMinutes int    `json:"duration_minutes"`
	SessionTypeID   string `json:"session_type_id"`
	SessionTypeName string `json:"session_type_name"`
}

type bookedEntryResponse struct {
	DayOfWeek          int    `json:"day_of_week"`
	SessionTypeID      string `json:"session_type_id"`
	SessionTypeName    string `json:"session_type_name"`
	DurationMinutes    int    `json:"duration_minutes"`
	OccurrencesInMonth int    `json:"occurrences_in_month"`
	TotalMinutes       int    `json:"total_minutes"`
}

type invoiceRunExceptionResponse struct {
	ChildID         string   `json:"child_id"`
	ChildFirstName  string   `json:"child_first_name"`
	ChildMiddleName *string  `json:"child_middle_name"`
	ChildLastName   *string  `json:"child_last_name"`
	BlockerCodes    []string `json:"blocker_codes"`
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
	RunID        string                   `json:"run_id"`
	BillingMonth string                   `json:"billing_month"`
	Status       string                   `json:"status"`
	Summary      bulkIssueSummary         `json:"summary"`
	Issued       []issuedInvoiceResponse  `json:"issued"`
	Blocked      []blockedInvoiceResponse `json:"blocked"`
}

type bulkIssueSummary struct {
	EligibleCount int `json:"eligible_count"`
	SuccessCount  int `json:"success_count"`
	BlockedCount  int `json:"blocked_count"`
	TotalDueMinor int `json:"total_due_minor"`
}

type issuedInvoiceResponse struct {
	InvoiceID       string  `json:"invoice_id"`
	ChildID         string  `json:"child_id"`
	ChildFirstName  string  `json:"child_first_name"`
	ChildMiddleName *string `json:"child_middle_name"`
	ChildLastName   *string `json:"child_last_name"`
	InvoiceNumber   string  `json:"invoice_number"`
	IssuedAt        string  `json:"issued_at"`
	DueAt           string  `json:"due_at"`
	TotalDueMinor   int     `json:"total_due_minor"`
}

type blockedInvoiceResponse struct {
	InvoiceID       string                 `json:"invoice_id"`
	ChildID         *string                `json:"child_id,omitempty"`
	ChildFirstName  string                 `json:"child_first_name,omitempty"`
	ChildMiddleName *string                `json:"child_middle_name"`
	ChildLastName   *string                `json:"child_last_name"`
	Blockers        []issueBlockerResponse `json:"blockers"`
}

type issueBlockerResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Parent Invoice View DTOs (API-21) ---

type parentInvoiceListResponse struct {
	Items  []parentInvoiceListItemResponse `json:"items"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
}

type parentInvoiceListItemResponse struct {
	InvoiceID            string  `json:"invoice_id"`
	InvoiceKind          string  `json:"invoice_kind"`
	InvoiceNumber        *string `json:"invoice_number"`
	InvoiceNumberDisplay string  `json:"invoice_number_display"`
	ChildID              string  `json:"child_id"`
	ChildFirstName       string  `json:"child_first_name"`
	ChildMiddleName      *string `json:"child_middle_name"`
	ChildLastName        *string `json:"child_last_name"`
	BillingMonth         string  `json:"billing_month"`
	Period               struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	Status                 string  `json:"status"`
	DueStatus              string  `json:"due_status"`
	CurrencyCode           string  `json:"currency_code"`
	SubtotalMinor          int     `json:"subtotal_minor"`
	FundedDeductionMinor   int     `json:"funded_deduction_minor"`
	TotalDueMinor          int     `json:"total_due_minor"`
	AmountPaidMinor        int     `json:"amount_paid_minor"`
	IssuedAt               *string `json:"issued_at"`
	DueAt                  *string `json:"due_at"`
	PaidAt                 *string `json:"paid_at"`
	PaymentFailedAt        *string `json:"payment_failed_at"`
	PaymentStatusUpdatedAt *string `json:"payment_status_updated_at"`
}

type parentInvoiceDetailResponse struct {
	InvoiceID            string  `json:"invoice_id"`
	InvoiceKind          string  `json:"invoice_kind"`
	InvoiceNumber        *string `json:"invoice_number"`
	InvoiceNumberDisplay string  `json:"invoice_number_display"`
	ChildID              string  `json:"child_id"`
	ChildFirstName       string  `json:"child_first_name"`
	ChildMiddleName      *string `json:"child_middle_name"`
	ChildLastName        *string `json:"child_last_name"`
	BillingMonth         string  `json:"billing_month"`
	Period               struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	} `json:"period"`
	Status                 string                           `json:"status"`
	DueStatus              string                           `json:"due_status"`
	CurrencyCode           string                           `json:"currency_code"`
	SubtotalMinor          int                              `json:"subtotal_minor"`
	FundedDeductionMinor   int                              `json:"funded_deduction_minor"`
	TotalDueMinor          int                              `json:"total_due_minor"`
	AmountPaidMinor        int                              `json:"amount_paid_minor"`
	IssuedAt               *string                          `json:"issued_at"`
	DueAt                  *string                          `json:"due_at"`
	PaidAt                 *string                          `json:"paid_at"`
	PaymentFailedAt        *string                          `json:"payment_failed_at"`
	PaymentStatusUpdatedAt *string                          `json:"payment_status_updated_at"`
	Calculation            parentInvoiceCalculationResponse `json:"calculation"`
	Lines                  []parentInvoiceLineResponse      `json:"lines"`
}

type parentInvoiceLineResponse struct {
	LineKind        string `json:"line_kind"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
	QuantityMinutes *int   `json:"quantity_minutes"`
	UnitAmountMinor *int   `json:"unit_amount_minor"`
	LineAmountMinor int    `json:"line_amount_minor"`
}

type parentInvoiceCalculationResponse struct {
	CoreHourlyRateMinor    int `json:"core_hourly_rate_minor"`
	BookedCoreMinutes      int `json:"booked_core_minutes"`
	BookedSessionCount     int `json:"booked_session_count"`
	FundedAllowanceMinutes int `json:"funded_allowance_minutes"`
	FundedDeductionMinutes int `json:"funded_deduction_minutes"`
	CoreBillableMinutes    int `json:"core_billable_minutes"`
	CoreSubtotalMinor      int `json:"core_subtotal_minor"`
	ExtrasTotalMinor       int `json:"extras_total_minor"`
}
