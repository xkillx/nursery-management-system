package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count    int
	windowStart time.Time
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
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()

	e, exists := l.windows[key]
	if !exists || now.Sub(e.windowStart) >= l.windowSize {
		l.windows[key] = &entry{count: 1, windowStart: now}
		return true
	}

	if e.count >= l.maxCount {
		return false
	}

	e.count++
	return true
}
