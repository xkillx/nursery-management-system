package application

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
)

func TestListManagerPaymentEvents_MalformedInvoiceID(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), "not-a-uuid", 1, 50)
	assertDomainCode(t, err, "validation_error")
}

func TestListManagerPaymentEvents_InvoiceNotFound(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), 1, 50)
	assertDomainCode(t, err, "invoice_not_found")
}

func TestListManagerPaymentEvents_DefaultPagination(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.PageSize != 50 {
		t.Errorf("expected page_size 50, got %d", result.PageSize)
	}
}

func TestListManagerPaymentEvents_CustomPagination(t *testing.T) {
	uc := NewListManagerPaymentEvents(&fakeManagerRepo{found: true})
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), 3, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 3 {
		t.Errorf("expected page 3, got %d", result.Page)
	}
	if result.PageSize != 25 {
		t.Errorf("expected page_size 25, got %d", result.PageSize)
	}
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
	result, err := uc.Execute(context.Background(), makeActor(), uuid.New().String(), 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].PaymentEventID != eventID {
		t.Errorf("event id mismatch: %s", result.Items[0].PaymentEventID)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}
