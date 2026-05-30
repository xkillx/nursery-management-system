package stripe

import (
	"context"
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"

	"nursery-management-system/api/internal/modules/payments/domain"
)

type Client struct{}

func NewClient(secretKey string) *Client {
	stripe.Key = secretKey
	return &Client{}
}

func (c *Client) CreateCheckoutSession(ctx context.Context, params domain.CheckoutSessionCreateParams) (domain.CheckoutSessionResult, error) {
	lineItem := &stripe.CheckoutSessionLineItemParams{
		PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
			Currency:   stripe.String(params.Currency),
			UnitAmount: stripe.Int64(int64(params.AmountMinor)),
			ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
				Name: stripe.String(params.ProductName),
			},
		},
		Quantity: stripe.Int64(1),
	}

	if params.ProductDesc != "" {
		lineItem.PriceData.ProductData.Description = stripe.String(params.ProductDesc)
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			lineItem,
		},
		ClientReferenceID: stripe.String(params.PaymentAttemptID),
		Metadata: map[string]string{
			"tenant_id":          params.TenantID,
			"branch_id":         params.BranchID,
			"invoice_id":        params.InvoiceID,
			"payment_attempt_id": params.PaymentAttemptID,
			"invoice_number":    params.InvoiceNumber,
		},
	}

	checkoutParams.SetIdempotencyKey(params.PaymentAttemptID)

	s, err := session.New(checkoutParams)
	if err != nil {
		return domain.CheckoutSessionResult{}, fmt.Errorf("stripe checkout session create: %w", err)
	}

	result := domain.CheckoutSessionResult{
		CheckoutSessionID: s.ID,
		CheckoutURL:       s.URL,
	}

	if s.PaymentIntent != nil {
		result.PaymentIntentID = s.PaymentIntent.ID
	}

	if s.ExpiresAt != 0 {
		result.ExpiresAt = strconv.FormatInt(s.ExpiresAt, 10)
	}

	return result, nil
}
