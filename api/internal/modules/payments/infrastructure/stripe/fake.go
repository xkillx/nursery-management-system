package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"nursery-management-system/api/internal/modules/payments/domain"
)

type FakeCheckoutProvider struct {
	apiPort string
}

func NewFakeCheckoutProvider(apiPort string) *FakeCheckoutProvider {
	return &FakeCheckoutProvider{apiPort: apiPort}
}

func (f *FakeCheckoutProvider) CreateCheckoutSession(_ context.Context, params domain.CheckoutSessionCreateParams) (domain.CheckoutSessionResult, error) {
	return domain.CheckoutSessionResult{
		CheckoutSessionID: "cs_test_" + params.PaymentAttemptID,
		CheckoutURL:       fmt.Sprintf("http://localhost:%s/api/v1/test/pay/%s", f.apiPort, params.PaymentAttemptID),
		PaymentIntentID:   "pi_test_" + params.PaymentAttemptID,
	}, nil
}

type FakeWebhookVerifier struct{}

func NewFakeWebhookVerifier() *FakeWebhookVerifier {
	return &FakeWebhookVerifier{}
}

func (v *FakeWebhookVerifier) VerifyAndParse(_ context.Context, payload []byte, _ string) (*domain.StripeWebhookEvent, error) {
	var raw struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Created int64  `json:"created"`
		Data    struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal webhook payload: %w", err)
	}

	result := &domain.StripeWebhookEvent{
		StripeEventID: raw.ID,
		EventType:     raw.Type,
		RawPayload:    payload,
	}

	if domain.CheckoutMutatingEventTypes[raw.Type] {
		var cs domain.CheckoutSessionWebhookData
		if err := json.Unmarshal(raw.Data.Object, &cs); err != nil {
			return nil, fmt.Errorf("unmarshal checkout session: %w", err)
		}
		result.CheckoutSession = &cs
	}

	return result, nil
}
