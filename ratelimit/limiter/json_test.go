package limiter_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/foxie-io/ng-contrib/ratelimit/limiter"
)

func TestLimiterJSON_MarshalUnmarshal(t *testing.T) {
	now := time.Now()

	// Create one of each limiter type
	limiters := []limiter.Limiter{
		limiter.NewTokenBucket(10, 5, time.Second, now),
		limiter.NewFixedWindow(3, time.Second, now),
		limiter.NewSlidingWindow(4, time.Second, now),
	}

	for _, l := range limiters {
		t.Run(l.Algorithm().String(), func(t *testing.T) {
			wrapped := limiter.LimiterJSON{Limiter: l}

			// Marshal
			data, err := json.Marshal(wrapped)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			// Unmarshal
			var unmarshalled limiter.LimiterJSON
			if err := json.Unmarshal(data, &unmarshalled); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			// Check type
			if unmarshalled.Limiter.Algorithm() != l.Algorithm() {
				t.Fatalf("expected algorithm %v, got %v", l.Algorithm(), unmarshalled.Limiter.Algorithm())
			}

			// Check remaining tokens/count roughly equal
			if unmarshalled.Limiter.Remaining() != l.Remaining() {
				t.Fatalf("expected remaining %d, got %d", l.Remaining(), unmarshalled.Limiter.Remaining())
			}

			if unmarshalled.Limiter.Window() != l.Window() {
				t.Fatalf("expected window %v, got %v", l.Window(), unmarshalled.Limiter.Window())
			}

			// Note: Time fields (LastSeen) may differ due to serialization precision loss
			if unmarshalled.Limiter.LastSeen().Sub(l.LastSeen()) > time.Millisecond {
				t.Fatalf("expected last seen approx %v, got %v", l.LastSeen(), unmarshalled.Limiter.LastSeen())
			}
		})
	}
}

func TestLimiterJSON_UnknownType(t *testing.T) {
	data := []byte(`{"type":"unknown","data":{}}`)
	var w limiter.LimiterJSON
	if err := json.Unmarshal(data, &w); err != limiter.ErrUnknownLimiter {
		t.Fatalf("expected ErrUnknownLimiter, got %v", err)
	}
}

func BenchmarkLimiterJSON_MarshalUnmarshal(b *testing.B) {
	now := time.Now()

	limiters := []limiter.Limiter{
		limiter.NewTokenBucket(100, 50, time.Second, now),
		limiter.NewFixedWindow(100, time.Second, now),
		limiter.NewSlidingWindow(100, time.Second, now),
	}

	for _, l := range limiters {
		b.Run(l.Algorithm().String(), func(b *testing.B) {
			wrapped := limiter.LimiterJSON{Limiter: l}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				data, err := json.Marshal(wrapped)
				if err != nil {
					b.Fatal(err)
				}

				var unmarshalled limiter.LimiterJSON
				if err := json.Unmarshal(data, &unmarshalled); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
