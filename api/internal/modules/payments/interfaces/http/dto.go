package httpayment

type createCheckoutSessionResponse struct {
	CheckoutSessionID string `json:"checkout_session_id"`
	CheckoutURL       string `json:"checkout_url"`
	PaymentAttemptID  string `json:"payment_attempt_id"`
}

type webhookResponse struct {
	Status string `json:"status"`
}

type managerPaymentStatusResponse struct {
	InvoiceID               string                       `json:"invoice_id"`
	InvoiceKind             string                       `json:"invoice_kind"`
	InvoiceNumber           string                       `json:"invoice_number"`
	InvoiceNumberDisplay    string                       `json:"invoice_number_display"`
	ChildID                 string                       `json:"child_id"`
	ChildFirstName          string                       `json:"child_first_name"`
	ChildMiddleName         *string                      `json:"child_middle_name"`
	ChildLastName           *string                      `json:"child_last_name"`
	BillingMonth            string                       `json:"billing_month"`
	Status                  string                       `json:"status"`
	DueStatus               string                       `json:"due_status"`
	CurrencyCode            string                       `json:"currency_code"`
	TotalDueMinor           int                          `json:"total_due_minor"`
	AmountPaidMinor         int                          `json:"amount_paid_minor"`
	IssuedAt                *string                      `json:"issued_at"`
	DueAt                   *string                      `json:"due_at"`
	PaidAt                  *string                      `json:"paid_at"`
	PaymentFailedAt         *string                      `json:"payment_failed_at"`
	PaymentStatusUpdatedAt  *string                      `json:"payment_status_updated_at"`
	CheckoutRetryAvailable  bool                         `json:"checkout_retry_available"`
	CheckoutRetryReasonCode string                       `json:"checkout_retry_reason_code"`
	LatestPaymentAttempt    *paymentAttemptDiagnosticDTO `json:"latest_payment_attempt"`
	LatestPaymentEvent      *paymentEventDiagnosticDTO   `json:"latest_payment_event"`
}

type paymentAttemptDiagnosticDTO struct {
	PaymentAttemptID        string  `json:"payment_attempt_id"`
	Status                  string  `json:"status"`
	AmountMinor             int     `json:"amount_minor"`
	CurrencyCode            string  `json:"currency_code"`
	StripeCheckoutSessionID *string `json:"stripe_checkout_session_id"`
	StripePaymentIntentID   *string `json:"stripe_payment_intent_id"`
	StripeExpiresAt         *string `json:"stripe_expires_at"`
	FailureReason           *string `json:"failure_reason"`
	ProviderErrorCode       *string `json:"provider_error_code"`
	ProviderErrorMessage    *string `json:"provider_error_message"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}

type paymentEventDiagnosticDTO struct {
	PaymentEventID          string  `json:"payment_event_id"`
	PaymentAttemptID        string  `json:"payment_attempt_id"`
	StripeEventID           string  `json:"stripe_event_id"`
	StripeEventType         string  `json:"stripe_event_type"`
	StripeCheckoutSessionID string  `json:"stripe_checkout_session_id"`
	StripePaymentIntentID   string  `json:"stripe_payment_intent_id"`
	Outcome                 string  `json:"outcome"`
	ReasonCode              string  `json:"reason_code"`
	PreviousInvoiceStatus   string  `json:"previous_invoice_status"`
	NewInvoiceStatus        string  `json:"new_invoice_status"`
	AttemptPreviousStatus   string  `json:"attempt_previous_status"`
	AttemptNewStatus        string  `json:"attempt_new_status"`
	AmountMinor             int     `json:"amount_minor"`
	CurrencyCode            string  `json:"currency_code"`
	WebhookProcessingStatus string  `json:"webhook_processing_status"`
	WebhookProcessingReason string  `json:"webhook_processing_reason"`
	WebhookReceivedAt       *string `json:"webhook_received_at"`
	WebhookProcessedAt      *string `json:"webhook_processed_at"`
	CreatedAt               string  `json:"created_at"`
}
