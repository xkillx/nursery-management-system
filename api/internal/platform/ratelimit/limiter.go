package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count       int
	windowStart time.Time
}

type AllowResult struct {
	Allowed   bool
	Remaining int
	Limit     int
	ResetAt   time.Time
}

type FixedWindowLimiter struct {
	mu         sync.Mutex
	windows    map[string]*entry
	maxCount   int
	windowSize time.Duration
	nowFunc    func() time.Time
}

func NewFixedWindowLimiter(maxCount int, windowSize time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows:    make(map[string]*entry),
		maxCount:   maxCount,
		windowSize: windowSize,
		nowFunc:    func() time.Time { return time.Now().UTC() },
	}
}

func NewFixedWindowLimiterWithClock(maxCount int, windowSize time.Duration, nowFunc func() time.Time) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		windows:    make(map[string]*entry),
		maxCount:   maxCount,
		windowSize: windowSize,
		nowFunc:    nowFunc,
	}
}

func (l *FixedWindowLimiter) Allow(key string) bool {
	return l.AllowWithInfo(key).Allowed
}

func (l *FixedWindowLimiter) AllowWithInfo(key string) AllowResult {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()

	e, exists := l.windows[key]
	if !exists || now.Sub(e.windowStart) >= l.windowSize {
		l.windows[key] = &entry{count: 1, windowStart: now}
		return AllowResult{
			Allowed:   true,
			Remaining: l.maxCount - 1,
			Limit:     l.maxCount,
			ResetAt:   now.Add(l.windowSize),
		}
	}

	if e.count >= l.maxCount {
		return AllowResult{
			Allowed:   false,
			Remaining: 0,
			Limit:     l.maxCount,
			ResetAt:   e.windowStart.Add(l.windowSize),
		}
	}

	e.count++
	return AllowResult{
		Allowed:   true,
		Remaining: l.maxCount - e.count,
		Limit:     l.maxCount,
		ResetAt:   e.windowStart.Add(l.windowSize),
	}
}
