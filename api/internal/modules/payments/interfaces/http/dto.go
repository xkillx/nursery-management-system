package httpayment

type createCheckoutSessionResponse struct {
	CheckoutSessionID string `json:"checkout_session_id"`
	CheckoutURL       string `json:"checkout_url"`
	PaymentAttemptID  string `json:"payment_attempt_id"`
}
