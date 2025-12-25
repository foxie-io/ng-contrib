package limiter

import (
	"time"
)

var _ Limiter = (*TokenBucket)(nil)

// =========================
// Token Bucket implementation of Limiter
// =========================

type TokenBucket struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"lastRefill"`
	Rate       float64   `json:"rate"`
	Burst      float64   `json:"burst"`
	WindowDur  float64   `json:"windowDur"` // in seconds
}

func (t *TokenBucket) Algorithm() Algorithm {
	return AlgorithmTokenBucket
}

func (t *TokenBucket) Allow(now time.Time) bool {
	elapsed := now.Sub(t.LastRefill).Seconds()
	t.Tokens += elapsed * t.Rate
	if t.Tokens > t.Burst {
		t.Tokens = t.Burst
	}
	t.LastRefill = now

	if t.Tokens < 1 {
		return false
	}
	t.Tokens--
	return true
}

func (t *TokenBucket) Remaining() int {
	return int(t.Tokens)
}

func (t *TokenBucket) LastSeen() time.Time {
	return t.LastRefill
}

func (t *TokenBucket) Window() time.Duration {
	return time.Duration(t.WindowDur * float64(time.Second))
}
