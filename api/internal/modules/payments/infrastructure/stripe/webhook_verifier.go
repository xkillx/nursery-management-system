package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/webhook"

	"nursery-management-system/api/internal/modules/payments/domain"
)

type WebhookVerifier struct {
	secret string
}

func NewWebhookVerifier(webhookSecret string) *WebhookVerifier {
	return &WebhookVerifier{secret: webhookSecret}
}

func (v *WebhookVerifier) VerifyAndParse(_ context.Context, payload []byte, signatureHeader string) (*domain.StripeWebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signatureHeader, v.secret)
	if err != nil {
		return nil, fmt.Errorf("stripe webhook signature verification: %w", err)
	}

	result := &domain.StripeWebhookEvent{
		StripeEventID: event.ID,
		EventType:     string(event.Type),
		Livemode:      event.Livemode,
		APIVersion:    string(event.APIVersion),
		RawPayload:    payload,
	}

	if event.Created > 0 {
		t := time.Unix(event.Created, 0)
		result.ProviderCreatedAt = &t
	}

	if domain.CheckoutMutatingEventTypes[string(event.Type)] {
		var cs stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
			return nil, fmt.Errorf("unmarshal checkout session: %w", err)
		}

		metadata := make(map[string]string)
		for k, v := range cs.Metadata {
			metadata[k] = v
		}

		result.CheckoutSession = &domain.CheckoutSessionWebhookData{
			CheckoutSessionID: cs.ID,
			PaymentStatus:     string(cs.PaymentStatus),
			AmountTotal:       cs.AmountTotal,
			Currency:          string(cs.Currency),
			Metadata:          metadata,
		}
		if cs.PaymentIntent != nil {
			result.CheckoutSession.PaymentIntentID = cs.PaymentIntent.ID
		}
	}

	return result, nil
}
