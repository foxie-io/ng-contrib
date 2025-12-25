package limiter

import (
	"time"
)

var _ Limiter = (*SlidingWindow)(nil)

// =========================
// Sliding Window implementation of Limiter
// =========================

type SlidingWindow struct {
	Log       []time.Time `json:"log"`
	Limit     int         `json:"limit"`
	WindowDur float64     `json:"windowDur"` // in seconds
}

func (s *SlidingWindow) Algorithm() Algorithm {
	return AlgorithmSlidingWindow
}

func (s *SlidingWindow) Window() time.Duration {
	return time.Duration(s.WindowDur * float64(time.Second))
}

func (s *SlidingWindow) Allow(now time.Time) bool {
	cutoff := now.Add(-s.Window())
	i := 0
	for ; i < len(s.Log); i++ {
		if s.Log[i].After(cutoff) {
			break
		}
	}
	s.Log = s.Log[i:]

	if len(s.Log) >= s.Limit {
		return false
	}

	s.Log = append(s.Log, now)
	return true
}

func (s *SlidingWindow) Remaining() int {
	return max(0, s.Limit-len(s.Log))
}

func (s *SlidingWindow) LastSeen() time.Time {
	if len(s.Log) == 0 {
		return time.Time{}
	}
	return s.Log[len(s.Log)-1]
}
