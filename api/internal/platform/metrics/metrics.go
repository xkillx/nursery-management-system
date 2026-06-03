package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Recorder struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	authFailures         *prometheus.CounterVec
	authzDenials         *prometheus.CounterVec
	webhookOutcomes      *prometheus.CounterVec
	invoiceGenRuns       *prometheus.CounterVec
	invoiceGenDuration   *prometheus.HistogramVec
	invoiceGenBlockers   *prometheus.CounterVec
	schedulerRuns        *prometheus.CounterVec
	schedulerRunDuration *prometheus.HistogramVec
	paymentTransitions   *prometheus.CounterVec
}

func NewRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return registry
}

func NewRecorder(r *prometheus.Registry) *Recorder {
	rec := &Recorder{
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_http_requests_total",
		}, []string{"route", "method", "status_code", "status_class"}),
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "nursery_api_http_request_duration_seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"route", "method", "status_class"}),
		authFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_auth_failures_total",
		}, []string{"operation", "reason"}),
		authzDenials: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_authorization_denials_total",
		}, []string{"operation", "denial_code"}),
		webhookOutcomes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_webhook_outcomes_total",
		}, []string{"provider", "event_type", "outcome", "reason"}),
		invoiceGenRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_invoice_generation_runs_total",
		}, []string{"mode", "outcome"}),
		invoiceGenDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "nursery_api_invoice_generation_duration_seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		}, []string{"mode", "outcome"}),
		invoiceGenBlockers: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_invoice_generation_blockers_total",
		}, []string{"blocker_code"}),
		schedulerRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_scheduler_runs_total",
		}, []string{"job", "trigger", "outcome"}),
		schedulerRunDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "nursery_api_scheduler_run_duration_seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
		}, []string{"job", "trigger", "outcome"}),
		paymentTransitions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "nursery_api_payment_state_transitions_total",
		}, []string{"operation", "entity_type", "previous_status", "new_status", "reason"}),
	}

	r.MustRegister(
		rec.httpRequestsTotal,
		rec.httpRequestDuration,
		rec.authFailures,
		rec.authzDenials,
		rec.webhookOutcomes,
		rec.invoiceGenRuns,
		rec.invoiceGenDuration,
		rec.invoiceGenBlockers,
		rec.schedulerRuns,
		rec.schedulerRunDuration,
		rec.paymentTransitions,
	)

	return rec
}

func (r *Recorder) ObserveHTTPRequest(route, method string, status int, seconds float64) {
	if r == nil {
		return
	}
	sc := statusClass(status)
	r.httpRequestsTotal.WithLabelValues(sanitize(route), method, strconv.Itoa(status), sc).Inc()
	r.httpRequestDuration.WithLabelValues(sanitize(route), method, sc).Observe(seconds)
}

func (r *Recorder) AuthFailure(operation, reason string) {
	if r == nil {
		return
	}
	r.authFailures.WithLabelValues(operation, reason).Inc()
}

func (r *Recorder) AuthorizationDenial(operation, denialCode string) {
	if r == nil {
		return
	}
	r.authzDenials.WithLabelValues(operation, denialCode).Inc()
}

func (r *Recorder) WebhookOutcome(provider, eventType, outcome, reason string) {
	if r == nil {
		return
	}
	r.webhookOutcomes.WithLabelValues(provider, sanitize(eventType), outcome, sanitize(reason)).Inc()
}

func (r *Recorder) InvoiceGenerationRun(mode, outcome string, seconds float64) {
	if r == nil {
		return
	}
	r.invoiceGenRuns.WithLabelValues(mode, outcome).Inc()
	r.invoiceGenDuration.WithLabelValues(mode, outcome).Observe(seconds)
}

func (r *Recorder) InvoiceGenerationBlocker(blockerCode string, count int) {
	if r == nil {
		return
	}
	r.invoiceGenBlockers.WithLabelValues(blockerCode).Add(float64(count))
}

func (r *Recorder) SchedulerRun(job, trigger, outcome string, seconds float64) {
	if r == nil {
		return
	}
	r.schedulerRuns.WithLabelValues(job, trigger, outcome).Inc()
	r.schedulerRunDuration.WithLabelValues(job, trigger, outcome).Observe(seconds)
}

func (r *Recorder) PaymentStateTransition(operation, entityType, previousStatus, newStatus, reason string) {
	if r == nil {
		return
	}
	r.paymentTransitions.WithLabelValues(operation, sanitize(entityType), sanitize(previousStatus), sanitize(newStatus), sanitize(reason)).Inc()
}

func statusClass(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx"
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func sanitize(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
