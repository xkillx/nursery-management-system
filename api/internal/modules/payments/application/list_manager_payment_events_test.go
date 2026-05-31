package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
)

func TestListManagerPaymentEvents_MalformedInvoiceID(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), "not-a-uuid", "", "")
	assertDomainCode(t, err, "validation_error")
}

func TestListManagerPaymentEvents_InvoiceNotFound(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "", "")
	assertDomainCode(t, err, "invoice_not_found")
}

func TestListManagerPaymentEvents_DefaultPagination(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Limit != 50 {
		t.Errorf("expected limit 50, got %d", result.Limit)
	}
	if result.Offset != 0 {
		t.Errorf("expected offset 0, got %d", result.Offset)
	}
}

func TestListManagerPaymentEvents_CustomPagination(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "100", "25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Limit != 100 {
		t.Errorf("expected limit 100, got %d", result.Limit)
	}
	if result.Offset != 25 {
		t.Errorf("expected offset 25, got %d", result.Offset)
	}
}

func TestListManagerPaymentEvents_InvalidLimit(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "0", "")
	assertDomainCode(t, err, "validation_error")
}

func TestListManagerPaymentEvents_InvalidLimitTooLarge(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "201", "")
	assertDomainCode(t, err, "validation_error")
}

func TestListManagerPaymentEvents_InvalidOffset(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "", "-1")
	assertDomainCode(t, err, "validation_error")
}

func TestListManagerPaymentEvents_EventsReturned(t *testing.T) {
	eventID := uuid.New().String()
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{
		found: true,
		events: []domain.PaymentEventDiagnostic{
			{
				PaymentEventID:          eventID,
				StripeEventID:           "evt_123",
				StripeEventType:         "checkout.session.completed",
				Outcome:                 "paid",
				WebhookProcessingStatus: "processed",
			},
		},
	})
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].PaymentEventID != eventID {
		t.Errorf("event id mismatch: %s", result.Items[0].PaymentEventID)
	}
}
