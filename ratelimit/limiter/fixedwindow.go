package limiter

import (
	"time"
)

var _ Limiter = (*FixedWindow)(nil)

// =========================
// Fixed Window implementation of Limiter
// =========================

type FixedWindow struct {
	Count     int       `json:"count"`
	ResetAt   time.Time `json:"resetAt"`
	Limit     int       `json:"limit"`
	WindowDur float64   `json:"windowDur"` // in seconds
}

func (f *FixedWindow) Algorithm() Algorithm {
	return AlgorithmFixedWindow
}

func (f *FixedWindow) Window() time.Duration {
	return time.Duration(f.WindowDur * float64(time.Second))
}

func (f *FixedWindow) Allow(now time.Time) bool {
	if now.After(f.ResetAt) {
		f.Count = 0
		f.ResetAt = now.Add(f.Window())
	}
	f.Count++
	return f.Count <= f.Limit
}

func (f *FixedWindow) Remaining() int {
	return max(0, f.Limit-f.Count)
}

func (f *FixedWindow) LastSeen() time.Time {
	return f.ResetAt
}
