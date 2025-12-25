package limiter

import (
	"time"
)

type Limiter interface {
	Algorithm() Algorithm
	Allow(now time.Time) bool
	Remaining() int
	LastSeen() time.Time
	Window() time.Duration
}

func NewTokenBucket(burst, fillRate int, window time.Duration, now time.Time) *TokenBucket {
	return &TokenBucket{
		Tokens:     float64(burst),
		Rate:       float64(fillRate) / window.Seconds(),
		Burst:      float64(burst),
		WindowDur:  window.Seconds(),
		LastRefill: now,
	}
}

func NewFixedWindow(limit int, window time.Duration, now time.Time) *FixedWindow {
	return &FixedWindow{
		Limit:     limit,
		WindowDur: window.Seconds(),
		ResetAt:   now.Add(window),
	}
}

func NewSlidingWindow(limit int, window time.Duration, now time.Time) *SlidingWindow {
	return &SlidingWindow{
		Limit:     limit,
		WindowDur: window.Seconds(),
		Log:       []time.Time{now},
	}
}
