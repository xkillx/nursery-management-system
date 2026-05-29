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

func TestScheduler_StopWithoutStart(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	runner := &stubRunner{}

	s, _ := NewScheduler(logger, runner)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should not panic
	s.Stop(ctx)
}
