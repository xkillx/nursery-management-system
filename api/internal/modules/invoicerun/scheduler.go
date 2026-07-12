package invoicerun

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"nursery-management-system/api/internal/modules/billing/domain"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
)

// OverdueRunner is the existing billing overdue-job runner.
type OverdueRunner interface {
	Execute(ctx context.Context) (domain.OverdueTransitionResult, error)
}

// ReminderRunner is the pre-overdue reminder job runner.
type ReminderRunner interface {
	Execute(ctx context.Context) (domain.ReminderJobResult, error)
}

// Scheduler is the cron-style runner for the three advance-pay jobs:
//
//   - 02:00 Europe/London daily — mark_overdue_advance_invoices
//     (8th-of-month rule; today >= billing_month_start + 7 days).
//   - 02:00 Europe/London daily — expire_terms (mark pending_renewal
//     at T-30, mark ended at T+1, write audit events for soft warning
//     at T-45 and renewal prompt at T-30).
//   - 25th 00:05 Europe/London monthly — generate_advance_invoices
//     (one per active Term for the next billing month). Idempotent;
//     re-running for the same month is a no-op.
//
// All jobs run in their own goroutine on the cron tick. Job locks
// (advisory transactions) keep concurrent invocations safe.
type Scheduler struct {
	logger   *slog.Logger
	cron     *cron.Cron
	overdue  OverdueRunner
	reminder ReminderRunner
	expire   *ExpireTermsRunner
	recorder *metrics.Recorder
	ctx      context.Context
	started  bool
}

func NewScheduler(
	logger *slog.Logger,
	overdue OverdueRunner,
	expire *ExpireTermsRunner,
	recorder *metrics.Recorder,
	reminder ReminderRunner,
) (*Scheduler, error) {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("load Europe/London: %w", err)
	}
	c := cron.New(cron.WithLocation(london), cron.WithLogger(cron.DiscardLogger))

	s := &Scheduler{
		logger:   logger,
		cron:     c,
		overdue:  overdue,
		reminder: reminder,
		expire:   expire,
		recorder: recorder,
	}

	// Daily 02:00 — overdue transition.
	if _, addErr := c.AddFunc("0 2 * * *", s.runOverdue); addErr != nil {
		return nil, fmt.Errorf("register overdue job: %w", addErr)
	}
	// Daily 02:00 — term lifecycle.
	if _, addErr := c.AddFunc("0 2 * * *", s.runExpireTerms); addErr != nil {
		return nil, fmt.Errorf("register expire-terms job: %w", addErr)
	}
	// Daily 08:00 — pre-overdue reminders.
	if _, addErr := c.AddFunc("0 8 * * *", s.runReminders); addErr != nil {
		return nil, fmt.Errorf("register reminders job: %w", addErr)
	}
	return s, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx = ctx
	s.cron.Start()
	s.started = true
	go s.runOverdueStartup()
	go s.runExpireTermsStartup()
	go s.runRemindersStartup()
}

func (s *Scheduler) Stop(ctx context.Context) {
	if !s.started {
		return
	}
	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
}

// runOverdue is the cron callback for the daily 02:00 overdue job.
func (s *Scheduler) runOverdue() {
	s.runOverdueWithTrigger("schedule")
}

func (s *Scheduler) runOverdueStartup() {
	s.runOverdueWithTrigger("startup")
}

func (s *Scheduler) runOverdueWithTrigger(trigger string) {
	if s.overdue == nil {
		return
	}
	startedAt := time.Now()
	runCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	requestID := httpserver.NewRequestID()
	s.logger.Info("overdue_job_started",
		"trigger", trigger,
		"request_id", requestID,
	)

	result, err := s.overdue.Execute(runCtx)
	elapsed := time.Since(startedAt).Seconds()

	if err != nil {
		s.logger.Error("overdue_job_failed",
			"request_id", requestID,
			"error", err,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		s.recorder.SchedulerRun("overdue_transition_job", trigger, "error", elapsed)
		return
	}

	if !result.LockAcquired {
		s.logger.Info("overdue_job_skipped_lock_held",
			"request_id", requestID,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		s.recorder.SchedulerRun("overdue_transition_job", trigger, "skipped_lock_held", elapsed)
		return
	}

	s.logger.Info("overdue_job_completed",
		"request_id", requestID,
		"transitioned_count", len(result.Transitioned),
		"current_london_date", result.CurrentLondonDate.Format("2006-01-02"),
		"cutoff_utc", result.CutoffUTC.Format(time.RFC3339),
		"latency_ms", time.Since(startedAt).Milliseconds(),
	)
	s.recorder.SchedulerRun("overdue_transition_job", trigger, "completed", elapsed)

	for range result.Transitioned {
		s.recorder.PaymentStateTransition("overdue_transition_job", "invoice", "issued", "overdue", "due_at_cutoff")
	}
}

// runExpireTerms is the cron callback for the daily 02:00 term-lifecycle job.
func (s *Scheduler) runExpireTerms() {
	s.runExpireTermsWithTrigger("schedule")
}

func (s *Scheduler) runExpireTermsStartup() {
	s.runExpireTermsWithTrigger("startup")
}

func (s *Scheduler) runExpireTermsWithTrigger(trigger string) {
	if s.expire == nil {
		return
	}
	startedAt := time.Now()
	runCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	requestID := httpserver.NewRequestID()
	s.logger.Info("expire_terms_job_started",
		"trigger", trigger,
		"request_id", requestID,
	)

	err := s.expire.RunForAllTenantsAndBranches(runCtx)
	elapsed := time.Since(startedAt).Seconds()

	if err != nil {
		s.logger.Error("expire_terms_job_failed",
			"request_id", requestID,
			"error", err,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		s.recorder.SchedulerRun("expire_terms_job", trigger, "error", elapsed)
		return
	}

	s.logger.Info("expire_terms_job_completed",
		"request_id", requestID,
		"latency_ms", time.Since(startedAt).Milliseconds(),
	)
	s.recorder.SchedulerRun("expire_terms_job", trigger, "completed", elapsed)
}

// runReminders is the cron callback for the daily 08:00 reminder job.
func (s *Scheduler) runReminders() {
	s.runRemindersWithTrigger("schedule")
}

func (s *Scheduler) runRemindersStartup() {
	s.runRemindersWithTrigger("startup")
}

func (s *Scheduler) runRemindersWithTrigger(trigger string) {
	if s.reminder == nil {
		return
	}
	startedAt := time.Now()
	runCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	requestID := httpserver.NewRequestID()
	s.logger.Info("reminder_job_started",
		"trigger", trigger,
		"request_id", requestID,
	)

	result, err := s.reminder.Execute(runCtx)
	elapsed := time.Since(startedAt).Seconds()

	if err != nil {
		s.logger.Error("reminder_job_failed",
			"request_id", requestID,
			"error", err,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		s.recorder.SchedulerRun("reminder_job", trigger, "error", elapsed)
		return
	}

	if !result.LockAcquired {
		s.logger.Info("reminder_job_skipped_lock_held",
			"request_id", requestID,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		s.recorder.SchedulerRun("reminder_job", trigger, "skipped_lock_held", elapsed)
		return
	}

	s.logger.Info("reminder_job_completed",
		"request_id", requestID,
		"due_soon_count", len(result.DueSoon),
		"due_today_count", len(result.DueToday),
		"latency_ms", time.Since(startedAt).Milliseconds(),
	)
	s.recorder.SchedulerRun("reminder_job", trigger, "completed", elapsed)
}
