package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/foxie-io/ng"
	"github.com/foxie-io/ng-contrib/ratelimit/limiter"
)

var _ interface {
	ng.ID
	ng.Guard
} = (*Guard)(nil)

/* =========================
   Config
========================= */

type Config struct {

	// Limit is the maximum number of requests allowed within the window.
	Limit int

	// Burst is the maximum burst size for token bucket algorithm.
	Burst int

	// Window is the rate limit window duration
	Window time.Duration // rate limit window

	// Identifier extracts client ID from context (e.g., IP, API key).
	Identifier func(ctx context.Context) string

	// ErrorHandler is called when rate limit is exceeded.
	ErrorHandler func(ctx context.Context) error

	// SetHeaderHandler sets rate limit headers (optional)
	//	 key := "LIMIT", "REMAINING", "RESET", "BURST"
	SetHeaderHandler func(ctx context.Context, key, value string)

	// Optional metadata key for per-route config
	MetadataKey string

	// Optional Guard ID (for skipping)
	GuardSkipperID ng.ID

	// Algorithm specifies the rate limiting algorithm to use.
	Algorithm limiter.Algorithm

	// Store default is in-memory store
	Store Store
}

/* =========================
   Client state
========================= */

type clientData struct {
	limiter limiter.Limiter
}

func newLimiter(cfg *Config, now time.Time) limiter.Limiter {
	switch cfg.Algorithm {
	case limiter.AlgorithmFixedWindow:
		return &limiter.FixedWindow{
			Limit:     cfg.Limit,
			WindowDur: cfg.Window.Seconds(),
			ResetAt:   now.Add(cfg.Window),
		}
	case limiter.AlgorithmSlidingWindow:
		return &limiter.SlidingWindow{
			Limit:     cfg.Limit,
			WindowDur: cfg.Window.Seconds(),
			Log:       []time.Time{},
		}
	default:
		return &limiter.TokenBucket{
			Rate:       float64(cfg.Limit) / cfg.Window.Seconds(),
			Burst:      float64(cfg.Burst),
			WindowDur:  cfg.Window.Seconds(),
			Tokens:     float64(cfg.Burst),
			LastRefill: now,
		}
	}
}

/* =========================
   Guard
========================= */

type Guard struct {
	defaultLimitGuardID
	clients map[string]*clientData
	mutex   sync.Mutex
	config  *Config
}

type defaultLimitGuardID struct {
	ng.DefaultID[defaultLimitGuardID]
}

func (g *Guard) NgID() string {
	if g.config == nil || g.config.GuardSkipperID == nil {
		return g.defaultLimitGuardID.NgID()
	}
	return g.config.GuardSkipperID.NgID()
}

/* =========================
   Defaults
========================= */

var DefaultConfig = &Config{
	Limit:  100,
	Burst:  20,
	Window: time.Second,
	Identifier: func(ctx context.Context) string {
		return "default-client"
	},
	ErrorHandler: func(ctx context.Context) error {
		return errors.New("rate limit exceeded")
	},
}

/* =========================
   Constructor
========================= */

func New(config *Config) *Guard {
	if config == nil {
		config = DefaultConfig
	}

	config = overrideOptional(config, DefaultConfig)

	// Provide default store if none provided
	if config.Store == nil {
		config.Store = NewInMemoryStore()
	}

	guard := &Guard{
		config: config,
	}

	guard.startCleanup(2 * time.Minute)
	return guard
}

func (g *Guard) Allow(ctx context.Context) error {
	cfg := g.config

	if routeCfg, ok := GetConfig(ctx, cfg.MetadataKey); ok {
		cfg = overrideOptional(routeCfg, g.config)
	}

	id := cfg.Identifier(ctx)
	now := time.Now()

	// Get limiter from store
	lim := cfg.Store.Get(id)
	if lim == nil {
		lim = newLimiter(cfg, now)
		cfg.Store.Set(id, lim)
	}

	allowed := lim.Allow(now)
	remaining := lim.Remaining()

	if !allowed {
		return cfg.ErrorHandler(ctx)
	}

	if cfg.SetHeaderHandler != nil {
		cfg.SetHeaderHandler(ctx, "Limit", fmt.Sprintf("%d", cfg.Limit))
		cfg.SetHeaderHandler(ctx, "Remaining", fmt.Sprintf("%d", remaining))
		resetTime := time.Now().Add(lim.Window())
		cfg.SetHeaderHandler(ctx, "Reset", resetTime.Format(time.RFC3339))

		// Optional burst info for token bucket
		if tb, ok := lim.(*limiter.TokenBucket); ok {
			cfg.SetHeaderHandler(ctx, "Burst", fmt.Sprintf("%d", int(tb.Burst)))
		}
	}

	return nil
}

func (g *Guard) cleanupExpiredClients() {
	now := time.Now()
	for _, id := range g.config.Store.Keys() {
		lim := g.config.Store.Get(id)
		if lim == nil || lim.Window() == 0 {
			continue
		}
		if now.Sub(lim.LastSeen()) > 2*lim.Window() {
			g.config.Store.Delete(id)
		}
	}
}

func (g *Guard) startCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			g.cleanupExpiredClients()
		}
	}()
}

/* =========================
   Helpers
========================= */

func overrideOptional(config, def *Config) *Config {
	if config == nil {
		return def
	}
	if config.Limit == 0 {
		config.Limit = def.Limit
	}
	if config.Burst == 0 {
		config.Burst = def.Burst
	}
	if config.Window == 0 {
		config.Window = def.Window
	}
	if config.Identifier == nil {
		config.Identifier = def.Identifier
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = def.ErrorHandler
	}
	if config.SetHeaderHandler == nil {
		config.SetHeaderHandler = def.SetHeaderHandler
	}
	if config.MetadataKey == "" {
		config.MetadataKey = def.MetadataKey
	}
	if config.Algorithm == "" {
		config.Algorithm = def.Algorithm
	}
	if config.Store == nil {
		config.Store = def.Store
	}
	return config
}
