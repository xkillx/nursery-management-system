package ratelimit

import (
	"testing"
	"time"
)

func TestAllowsUnderLimit(t *testing.T) {
	limiter := NewFixedWindowLimiter(3, 15*time.Minute)
	for i := 0; i < 3; i++ {
		if !limiter.Allow("key") {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
}

func TestBlocksAtLimit(t *testing.T) {
	limiter := NewFixedWindowLimiter(2, 15*time.Minute)
	limiter.Allow("key")
	limiter.Allow("key")
	if limiter.Allow("key") {
		t.Fatal("expected block on third request")
	}
}

func TestIndependentKeys(t *testing.T) {
	limiter := NewFixedWindowLimiter(1, 15*time.Minute)
	if !limiter.Allow("key-a") {
		t.Fatal("expected allow for key-a")
	}
	if !limiter.Allow("key-b") {
		t.Fatal("expected allow for key-b (independent key)")
	}
}

func TestWindowResetAfterExpiry(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := NewFixedWindowLimiterWithClock(1, 15*time.Minute, func() time.Time {
		return base
	})

	if !limiter.Allow("key") {
		t.Fatal("expected allow first request")
	}
	if limiter.Allow("key") {
		t.Fatal("expected block second request")
	}

	base = base.Add(16 * time.Minute)
	if !limiter.Allow("key") {
		t.Fatal("expected allow after window expiry")
	}
}
