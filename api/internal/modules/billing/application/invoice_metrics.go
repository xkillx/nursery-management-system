package application

import (
	"log/slog"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/tenant"
)

type InvoiceMetrics struct {
	recorder *metrics.Recorder
	logger   *slog.Logger
}

func NewInvoiceMetrics(recorder *metrics.Recorder, logger *slog.Logger) *InvoiceMetrics {
	return &InvoiceMetrics{recorder: recorder, logger: logger}
}

func (m *InvoiceMetrics) RecordRun(mode, outcome string, startedAt time.Time, result domain.DraftGenerationResult, actor tenant.ActorContext) {
	elapsed := time.Since(startedAt).Seconds()
	if m.recorder != nil {
		m.recorder.InvoiceGenerationRun(mode, outcome, elapsed)
		for _, b := range result.Blocked {
			for _, bl := range b.Blockers {
				m.recorder.InvoiceGenerationBlocker(string(bl.Code), 1)
			}
		}
	}
	if m.logger != nil {
		args := []any{
			"operation", "advance_pay_draft_generation",
			"outcome", outcome,
			"run_id", result.RunID.String(),
			"billing_month", result.BillingMonth,
			"mode", mode,
			"eligible_count", result.Summary.EligibleCount,
			"success_count", result.Summary.SuccessCount,
			"blocked_count", result.Summary.BlockedCount,
			"total_due_minor", result.Summary.TotalDue.Minor(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"request_id", actor.RequestID,
			"correlation_id", actor.CorrelationID,
		}
		if actor.TraceID != "" {
			args = append(args, "trace_id", actor.TraceID)
		}
		m.logger.Info("advance_pay_draft_generation", args...)
	}
}
