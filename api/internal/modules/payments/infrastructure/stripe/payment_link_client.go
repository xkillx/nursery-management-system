package stripe

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/paymentlink"

	"nursery-management-system/api/internal/modules/payments/domain"
)

func (c *Client) CreatePaymentLink(ctx context.Context, params domain.PaymentLinkCreateParams) (domain.PaymentLinkResult, error) {
	lineItem := &stripe.PaymentLinkLineItemParams{
		PriceData: &stripe.PaymentLinkLineItemPriceDataParams{
			Currency:   stripe.String(params.Currency),
			UnitAmount: stripe.Int64(int64(params.AmountMinor)),
			ProductData: &stripe.PaymentLinkLineItemPriceDataProductDataParams{
				Name: stripe.String(params.ProductName),
			},
		},
		Quantity: stripe.Int64(1),
	}

	if params.Description != "" {
		lineItem.PriceData.ProductData.Description = stripe.String(params.Description)
	}

	linkParams := &stripe.PaymentLinkParams{
		LineItems: []*stripe.PaymentLinkLineItemParams{lineItem},
		AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
			Type: stripe.String(string(stripe.PaymentLinkAfterCompletionTypeHostedConfirmation)),
			HostedConfirmation: &stripe.PaymentLinkAfterCompletionHostedConfirmationParams{
				CustomMessage: stripe.String("Payment received. You may close this page."),
			},
		},
		Metadata: map[string]string{
			"tenant_id":      params.TenantID,
			"branch_id":      params.BranchID,
			"invoice_id":     params.InvoiceID,
			"invoice_number": params.InvoiceNumber,
		},
	}

	linkParams.SetIdempotencyKey(fmt.Sprintf("plink_%s", params.InvoiceID))

	pl, err := paymentlink.New(linkParams)
	if err != nil {
		return domain.PaymentLinkResult{}, fmt.Errorf("stripe payment link create: %w", err)
	}

	return domain.PaymentLinkResult{
		ID:  pl.ID,
		URL: pl.URL,
	}, nil
}
