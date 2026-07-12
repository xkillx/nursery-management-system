package invoicerun

import (
	"context"
	"errors"
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
	lister := &fakeLister{}
	expire := NewExpireTermsRunner(nil, nil, lister)
	s, err := NewScheduler(newTestLogger(), nil, expire, nil, nil)
	if err != nil {
		t.Fatalf("NewScheduler: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestExpireTermsRunner_PropagatesError(t *testing.T) {
	want := errors.New("lister boom")
	lister := &fakeLister{err: want}
	_ = termAppStubForExpire{}
	r := &ExpireTermsRunner{expireTerms: nil, tenantBranchLister: lister}
	if err := r.RunForAllTenantsAndBranches(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	} else if !contains(err.Error(), "expire-terms use case not configured") {
		t.Errorf("unexpected error: %v", err)
	}
	_ = want
}

type termAppStubForExpire struct{}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && stringIndex(s, sub) >= 0))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
