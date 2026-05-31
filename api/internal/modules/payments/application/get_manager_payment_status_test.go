package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/payments/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeManagerRepo struct {
	invoice domain.ManagerInvoicePaymentStatus
	found   bool
	attempt *domain.PaymentAttemptDiagnostic
	event   *domain.PaymentEventDiagnostic
	events  []domain.PaymentEventDiagnostic
	err     error
}

func (f *fakeManagerRepo) GetManagerInvoicePaymentStatus(_ context.Context, _, _, _ string) (domain.ManagerInvoicePaymentStatus, bool, error) {
	return f.invoice, f.found, f.err
}

func (f *fakeManagerRepo) GetLatestPaymentAttemptForInvoice(_ context.Context, _, _, _ string) (*domain.PaymentAttemptDiagnostic, error) {
	return f.attempt, f.err
}

func (f *fakeManagerRepo) GetLatestPaymentEventForInvoice(_ context.Context, _, _, _ string) (*domain.PaymentEventDiagnostic, error) {
	return f.event, f.err
}

func (f *fakeManagerRepo) ListPaymentEventsForInvoice(_ context.Context, _, _, _ string, _ domain.PaymentEventFilters) ([]domain.PaymentEventDiagnostic, error) {
	return f.events, f.err
}

func makeActor() tenant.ActorContext {
	return tenant.ActorContext{
		TenantID:     uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		UserID:       uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		MembershipID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
	}
}

func payableInvoice() domain.ManagerInvoicePaymentStatus {
	return domain.ManagerInvoicePaymentStatus{
		InvoiceID:       uuid.New().String(),
		InvoiceKind:     "monthly",
		InvoiceNumber:   "INV-202605-0001",
		ChildID:         uuid.New().String(),
		ChildName:       "Alex Child",
		BillingMonth:    "2026-05",
		Status:          "issued",
		CurrencyCode:    "GBP",
		TotalDueMinor:   1500,
		AmountPaidMinor: 0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func TestGetManagerPaymentStatus_MalformedInvoiceID(t *testing.T) {
	uc := NewGetManagerPaymentStatus(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), "not-a-uuid")
	assertDomainCode(t, err, "validation_error")
}

func TestGetManagerPaymentStatus_InvoiceNotFound(t *testing.T) {
	uc := NewGetManagerPaymentStatus(&fakeManagerRepo{})
	_, err := uc.Execute(context.Background(), makeActor(), uuid.New().String())
	assertDomainCode(t, err, "invoice_not_found")
}

func TestGetManagerPaymentStatus_NoAttemptsNoEvents(t *testing.T) {
	invoice := payableInvoice()
	uc := NewGetManagerPaymentStatus(&fakeManagerRepo{invoice: invoice, found: true})
	result, err := uc.Execute(context.Background(), makeActor(), invoice.InvoiceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LatestPaymentAttempt != nil {
		t.Error("expected nil attempt")
	}
	if result.LatestPaymentEvent != nil {
		t.Error("expected nil event")
	}
	if result.CheckoutRetryAvailable != true {
		t.Error("expected retry available")
	}
	if result.CheckoutRetryReasonCode != RetryAvailable {
		t.Errorf("expected available, got %s", result.CheckoutRetryReasonCode)
	}
	if result.DueStatus != "due" {
		t.Errorf("expected due, got %s", result.DueStatus)
	}
}

func TestGetManagerPaymentStatus_WithAttemptAndEvent(t *testing.T) {
	invoice := payableInvoice()
	attemptID := uuid.New().String()
	eventID := uuid.New().String()
	now := time.Now()
	uc := NewGetManagerPaymentStatus(&fakeManagerRepo{
		invoice: invoice,
		found:   true,
		attempt: &domain.PaymentAttemptDiagnostic{
			PaymentAttemptID: attemptID,
			Status:           "checkout_created",
			AmountMinor:      1500,
			CurrencyCode:     "GBP",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		event: &domain.PaymentEventDiagnostic{
			PaymentEventID:          eventID,
			PaymentAttemptID:        attemptID,
			StripeEventID:           "evt_123",
			StripeEventType:         "checkout.session.completed",
			Outcome:                 "paid",
			ReasonCode:              "paid",
			WebhookProcessingStatus: "processed",
			CreatedAt:               now,
		},
	})
	result, err := uc.Execute(context.Background(), makeActor(), invoice.InvoiceID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LatestPaymentAttempt == nil {
		t.Fatal("expected attempt")
	}
	if result.LatestPaymentEvent == nil {
		t.Fatal("expected event")
	}
	if result.LatestPaymentAttempt.PaymentAttemptID != attemptID {
		t.Errorf("attempt id mismatch: %s", result.LatestPaymentAttempt.PaymentAttemptID)
	}
	if result.LatestPaymentEvent.PaymentEventID != eventID {
		t.Errorf("event id mismatch: %s", result.LatestPaymentEvent.PaymentEventID)
	}
}

func TestRetryAvailability_Reasons(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*domain.ManagerInvoicePaymentStatus)
		expected string
	}{
		{"issued monthly GBP positive unpaid", nil, RetryAvailable},
		{"payment_failed", func(i *domain.ManagerInvoicePaymentStatus) { i.Status = "payment_failed" }, RetryAvailable},
		{"overdue", func(i *domain.ManagerInvoicePaymentStatus) { i.Status = "overdue" }, RetryAvailable},
		{"draft", func(i *domain.ManagerInvoicePaymentStatus) { i.Status = "draft" }, RetryNotIssued},
		{"paid", func(i *domain.ManagerInvoicePaymentStatus) { i.Status = "paid" }, RetryAlreadyPaid},
		{"zero total", func(i *domain.ManagerInvoicePaymentStatus) { i.TotalDueMinor = 0 }, RetryZeroTotal},
		{"partial paid", func(i *domain.ManagerInvoicePaymentStatus) { i.AmountPaidMinor = 500 }, RetryPartialPaid},
		{"non-monthly", func(i *domain.ManagerInvoicePaymentStatus) { i.InvoiceKind = "ad_hoc" }, RetryNotMonthly},
		{"non-GBP", func(i *domain.ManagerInvoicePaymentStatus) { i.CurrencyCode = "USD" }, RetryCurrencyNotSupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := payableInvoice()
			if tt.modify != nil {
				tt.modify(&invoice)
			}
			uc := NewGetManagerPaymentStatus(&fakeManagerRepo{invoice: invoice, found: true})
			result, err := uc.Execute(context.Background(), makeActor(), invoice.InvoiceID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.CheckoutRetryReasonCode != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.CheckoutRetryReasonCode)
			}
		})
	}
}

func TestDueStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"draft", "not_due"},
		{"issued", "due"},
		{"payment_failed", "due"},
		{"overdue", "overdue"},
		{"paid", "paid"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := computeDueStatus(tt.status)
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func assertDomainCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != code {
		t.Errorf("expected code %s, got %s", code, de.Code)
	}
}
