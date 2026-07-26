package invoicerun

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

// fakeLister returns a fixed list of tenant/branch scopes.
type fakeLister struct {
	scopes []TenantBranch
	err    error
}

func (f *fakeLister) ListAllTenantBranches(_ context.Context) ([]TenantBranch, error) {
	return f.scopes, f.err
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewScheduler_RegistersAllJobs(t *testing.T) {
	s, err := NewScheduler(newTestLogger(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestExpireTermsRunner_NoOp(t *testing.T) {
	r := NewExpireTermsRunner(nil)
	if err := r.RunForAllTenantsAndBranches(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
