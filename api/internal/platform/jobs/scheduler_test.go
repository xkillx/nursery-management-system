package jobs

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
)

type stubRunner struct {
	mu     sync.Mutex
	result domain.OverdueTransitionResult
	err    error
	calls  int
}

func (r *stubRunner) Execute(ctx context.Context) (domain.OverdueTransitionResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	return r.result, r.err
}

func (r *stubRunner) Calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func TestScheduler_CreationRegistersJob(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{}

	s, err := NewScheduler(logger, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduler_StartTriggersStartupRun(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{
		result: domain.OverdueTransitionResult{
			LockAcquired: true,
			Transitioned: []domain.OverdueTransitionedInvoice{},
		},
	}

	s, err := NewScheduler(logger, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)

	// Wait for startup run
	time.Sleep(500 * time.Millisecond)

	s.Stop(ctx)

	calls := runner.Calls()
	if calls < 1 {
		t.Fatalf("expected at least 1 call, got %d", calls)
	}
}

func TestScheduler_SkippedWhenLockHeld(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{
		result: domain.OverdueTransitionResult{
			LockAcquired: false,
		},
	}

	s, _ := NewScheduler(logger, runner)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Start(ctx)
	time.Sleep(500 * time.Millisecond)
	s.Stop(ctx)

	// Should not panic or fail — lock not acquired is treated as skip
	if runner.Calls() < 1 {
		t.Fatal("expected at least one call")
	}
}

func TestScheduler_OverdueJobNextRunUsesLondonMidnight(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{}

	s, err := NewScheduler(logger, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	london, _ := time.LoadLocation("Europe/London")

	s.Start(context.Background())
	defer s.Stop(context.Background())

	entries := s.cron.Entries()
	if len(entries) == 0 {
		t.Fatal("expected at least one cron entry")
	}
	sched := entries[0].Schedule

	t.Run("BST next run from 2026-07-14 22:30 UTC", func(t *testing.T) {
		from := time.Date(2026, 7, 14, 22, 30, 0, 0, time.UTC).In(london)
		next := sched.Next(from)
		want := time.Date(2026, 7, 14, 23, 0, 0, 0, time.UTC)
		if !next.Equal(want) {
			t.Fatalf("BST next run = %v, want %v (London midnight during BST = 23:00 UTC)", next, want)
		}
	})

	t.Run("GMT next run from 2026-01-14 23:30 UTC", func(t *testing.T) {
		from := time.Date(2026, 1, 14, 23, 30, 0, 0, time.UTC).In(london)
		next := sched.Next(from)
		want := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		if !next.Equal(want) {
			t.Fatalf("GMT next run = %v, want %v (London midnight during GMT = 00:00 UTC)", next, want)
		}
	})
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{}

	s, _ := NewScheduler(logger, runner)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should not panic
	s.Stop(ctx)
}
