package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"nursery-management-system/api/internal/modules/billing/domain"
)

type OverdueTransitionRunner interface {
	Execute(ctx context.Context) (domain.OverdueTransitionResult, error)
}

type Scheduler struct {
	logger  *slog.Logger
	cron    *cron.Cron
	runner  OverdueTransitionRunner
	ctx     context.Context
	started bool
}

func NewScheduler(logger *slog.Logger, runner OverdueTransitionRunner) (*Scheduler, error) {
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("load Europe/London: %w", err)
	}

	c := cron.New(cron.WithLocation(london), cron.WithLogger(cron.DiscardLogger))

	s := &Scheduler{
		logger: logger,
		cron:   c,
		runner: runner,
	}

	if _, addErr := c.AddFunc("0 0 * * *", func() { s.run("schedule") }); addErr != nil {
		return nil, fmt.Errorf("register overdue job: %w", addErr)
	}

	return s, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx = ctx
	s.cron.Start()
	s.started = true
	go s.run("startup")
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

func (s *Scheduler) run(trigger string) {
	runCtx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	s.logger.Info("overdue_job_started", "trigger", trigger)

	result, err := s.runner.Execute(runCtx)
	if err != nil {
		s.logger.Error("overdue_job_failed", "trigger", trigger, "error", err)
		return
	}

	if !result.LockAcquired {
		s.logger.Info("overdue_job_skipped_lock_held", "trigger", trigger)
		return
	}

	s.logger.Info("overdue_job_completed",
		"trigger", trigger,
		"transitioned_count", len(result.Transitioned),
		"current_london_date", result.CurrentLondonDate.Format("2006-01-02"),
		"cutoff_utc", result.CutoffUTC.Format(time.RFC3339),
	)
}
