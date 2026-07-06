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

func TestAllowWithInfoFirstRequest(t *testing.T) {
	limiter := NewFixedWindowLimiter(10, 15*time.Minute)
	result := limiter.AllowWithInfo("key")

	if !result.Allowed {
		t.Fatal("expected first request to be allowed")
	}
	if result.Remaining != 9 {
		t.Fatalf("expected remaining=9, got %d", result.Remaining)
	}
	if result.Limit != 10 {
		t.Fatalf("expected limit=10, got %d", result.Limit)
	}
	if result.ResetAt.IsZero() {
		t.Fatal("expected reset time to be set")
	}
}

func TestAllowWithInfoAtLimit(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := NewFixedWindowLimiterWithClock(10, 15*time.Minute, func() time.Time {
		return base
	})

	for i := 0; i < 10; i++ {
		result := limiter.AllowWithInfo("key")
		if !result.Allowed {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}

	result := limiter.AllowWithInfo("key")
	if result.Allowed {
		t.Fatal("expected block after limit reached")
	}
	if result.Remaining != 0 {
		t.Fatalf("expected remaining=0, got %d", result.Remaining)
	}
	if result.Limit != 10 {
		t.Fatalf("expected limit=10, got %d", result.Limit)
	}
	expectedReset := base.Add(15 * time.Minute)
	if !result.ResetAt.Equal(expectedReset) {
		t.Fatalf("expected reset at %v, got %v", expectedReset, result.ResetAt)
	}
}

func TestAllowWithInfoWindowReset(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := NewFixedWindowLimiterWithClock(10, 15*time.Minute, func() time.Time {
		return base
	})

	for i := 0; i < 10; i++ {
		limiter.AllowWithInfo("key")
	}

	blocked := limiter.AllowWithInfo("key")
	if blocked.Allowed {
		t.Fatal("expected block before window reset")
	}

	base = base.Add(16 * time.Minute)
	result := limiter.AllowWithInfo("key")
	if !result.Allowed {
		t.Fatal("expected allow after window reset")
	}
	if result.Remaining != 9 {
		t.Fatalf("expected remaining=9 after reset, got %d", result.Remaining)
	}
	if result.Limit != 10 {
		t.Fatalf("expected limit=10, got %d", result.Limit)
	}
}
