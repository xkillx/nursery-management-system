# Stripe Checkout Session webhook authority for MVP payments

For month 1, invoice payment reconciliation is driven by Stripe Checkout Session webhook events, not by parent browser return URLs or PaymentIntent events. We chose this because Checkout Sessions are the local payment-attempt boundary created by the API, carry the invoice and payment-attempt metadata needed for reconciliation, and avoid double-applying outcomes when Stripe emits multiple event families for the same payment; PaymentIntent events may be stored for operational visibility but must not mutate invoice status in the MVP.
