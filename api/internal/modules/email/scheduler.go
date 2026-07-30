package email

import (
	"context"
	"log/slog"
	"time"

	"nursery-management-system/api/internal/modules/email/application"
	"nursery-management-system/api/internal/platform/metrics"
)

type Scheduler struct {
	logger       *slog.Logger
	sendPending  *application.SendPendingEmails
	recorder     *metrics.Recorder
	pollInterval time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	started      bool
}

func NewScheduler(
	logger *slog.Logger,
	sendPending *application.SendPendingEmails,
	recorder *metrics.Recorder,
	pollIntervalSeconds int,
) *Scheduler {
	return &Scheduler{
		logger:       logger,
		sendPending:  sendPending,
		recorder:     recorder,
		pollInterval: time.Duration(pollIntervalSeconds) * time.Second,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.started = true

	go s.run()
	s.logger.Info("email_scheduler_started",
		"poll_interval", s.pollInterval,
	)
}

func (s *Scheduler) Stop(ctx context.Context) {
	if !s.started {
		return
	}
	s.cancel()
	s.logger.Info("email_scheduler_stopped")
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

func (s *Scheduler) runOnce() {
	startedAt := time.Now()
	runCtx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
	defer cancel()

	sent, failed, err := s.sendPending.Execute(runCtx)
	elapsed := time.Since(startedAt).Seconds()

	if err != nil {
		s.logger.Error("email_worker_failed",
			"error", err,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
		if s.recorder != nil {
			s.recorder.SchedulerRun("email_worker", "poll", "error", elapsed)
		}
		return
	}

	if sent > 0 || failed > 0 {
		s.logger.Info("email_worker_completed",
			"sent", sent,
			"failed", failed,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
	}

	if s.recorder != nil {
		s.recorder.SchedulerRun("email_worker", "poll", "completed", elapsed)
	}
}
