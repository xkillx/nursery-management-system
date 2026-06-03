package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func newTestRecorder() (*Recorder, *prometheus.Registry) {
	registry := NewRegistry()
	return NewRecorder(registry), registry
}

func TestNilRecorderNoop(t *testing.T) {
	var r *Recorder
	r.ObserveHTTPRequest("/health", "GET", 200, 0.01)
	r.AuthFailure("op", "reason")
	r.AuthorizationDenial("op", "code")
	r.WebhookOutcome("stripe", "evt", "ok", "")
	r.InvoiceGenerationRun("mode", "ok", 1)
	r.InvoiceGenerationBlocker("code", 1)
	r.SchedulerRun("job", "trigger", "ok", 1)
	r.PaymentStateTransition("op", "type", "old", "new", "reason")
}

func TestObserveHTTPRequest(t *testing.T) {
	r, _ := newTestRecorder()
	r.ObserveHTTPRequest("/api/v1/children", "GET", 200, 0.05)

	val := readCounter(t, r.httpRequestsTotal, "/api/v1/children", "GET", "200", "2xx")
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestAuthFailure(t *testing.T) {
	r, _ := newTestRecorder()
	r.AuthFailure("bearer_auth", "missing_bearer_token")

	val := readCounter(t, r.authFailures, "bearer_auth", "missing_bearer_token")
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestAuthorizationDenial(t *testing.T) {
	r, _ := newTestRecorder()
	r.AuthorizationDenial("require_roles", "forbidden_role")

	val := readCounter(t, r.authzDenials, "require_roles", "forbidden_role")
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestWebhookOutcomeSanitizesEmptyReason(t *testing.T) {
	r, _ := newTestRecorder()
	r.WebhookOutcome("stripe", "checkout.session.completed", "processed", "")

	val := readCounter(t, r.webhookOutcomes, "stripe", "checkout.session.completed", "processed", "unknown")
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestInvoiceGenerationBlocker(t *testing.T) {
	r, _ := newTestRecorder()
	r.InvoiceGenerationBlocker("child_inactive", 3)

	val := readCounter(t, r.invoiceGenBlockers, "child_inactive")
	if val != 3 {
		t.Fatalf("expected 3, got %d", val)
	}
}

func TestPaymentStateTransition(t *testing.T) {
	r, _ := newTestRecorder()
	r.PaymentStateTransition("webhook", "payment_attempt", "checkout_created", "paid", "paid")

	val := readCounter(t, r.paymentTransitions, "webhook", "payment_attempt", "checkout_created", "paid", "paid")
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestSanitize(t *testing.T) {
	if got := sanitize(""); got != "unknown" {
		t.Fatalf("expected 'unknown', got %q", got)
	}
	if got := sanitize("test"); got != "test" {
		t.Fatalf("expected 'test', got %q", got)
	}
}

func TestStatusClass(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{100, "1xx"}, {200, "2xx"}, {301, "3xx"},
		{404, "4xx"}, {500, "5xx"}, {0, "unknown"},
	}
	for _, tc := range tests {
		if got := statusClass(tc.code); got != tc.want {
			t.Fatalf("statusClass(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func readCounter(t *testing.T, cv *prometheus.CounterVec, labels ...string) uint64 {
	t.Helper()
	m, err := cv.GetMetricWithLabelValues(labels...)
	if err != nil {
		t.Fatalf("metric not found: %v", err)
	}
	var mDto dto.Metric
	if err := m.Write(&mDto); err != nil {
		t.Fatalf("write metric: %v", err)
	}
	return uint64(mDto.GetCounter().GetValue())
}
