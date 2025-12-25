# ng-contrib/ratelimit

`ng-contrib/ratelimit` is a flexible rate-limiting middleware for [foxie-io/ng](https://github.com/foxie-io/ng), supporting multiple algorithms, per-route and global limits, bursts, and pluggable stores including in-memory and Redis.

---

## Features

- ✅ Token Bucket, Fixed Window, Sliding Window algorithms
- ✅ Burst and smooth rate limiting
- ✅ Per-route and global guards
- ✅ Custom error handling
- ✅ Pluggable storage: in-memory, Redis
- ✅ Compatible with `foxie-io/ng`

---

## Algorithms Comparison

| Algorithm      | Burst | Smooth | Complexity      | Memory          | DoS / Abuse Resistance                                                         |
| -------------- | ----- | ------ | --------------- | --------------- | ------------------------------------------------------------------------------ |
| Fixed Window   | ❌    | ❌     | O(1)            | O(1)            | ❌ prone to spikes at window boundary                                          |
| Sliding Window | ❌    | ✅     | O(n per client) | O(n per client) | ⚠️ high memory per client; can be exploited by many clients (high cardinality) |
| Token Bucket   | ✅    | ✅     | O(1)            | O(1)            | ✅ handles bursts well; constant memory; better for abuse traffic              |

> **Note:** Use Token Bucket for high-throughput public APIs; Sliding Window for accurate, low-volume scenarios.

---

## Installation

```bash
go get github.com/foxie-io/ng-contrib/ratelimit
```

---

## Usage

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/foxie-io/ng"
    nghttp "github.com/foxie-io/ng/http"
    "github.com/foxie-io/ng-contrib/ratelimit"
)

func main() {
    globalGuard := ratelimit.New(&ratelimit.Config{
        Limit:  1000,
        Burst:  20,
        Window: time.Second,
        Identifier: func(ctx context.Context) string { return "global-client" },
        ErrorHandler: func(ctx context.Context) error { return nghttp.NewErrTooManyRequests() },
        SetHeaderHandler: func(ctx context.Context, key, value string) {
            w := ng.MustLoad[http.ResponseWriter](ctx)
            w.Header().Set(fmt.Sprintf("X-G-RateLimit-%s", key), value)
        },
    })

    perRouteGuard := ratelimit.New(&ratelimit.Config{
        Limit:  5,
        Burst:  2,
        Window: time.Second,
        Identifier: func(ctx context.Context) string {
            r := ng.GetContext(ctx)
            route := r.Route()
            return fmt.Sprintf("route_%s_%s", route.Method(), route.Path())
        },
        ErrorHandler: func(ctx context.Context) error { return nghttp.NewErrResourceExhausted() },
        SetHeaderHandler: func(ctx context.Context, key, value string) {
            w := ng.MustLoad[http.ResponseWriter](ctx)
            w.Header().Set(fmt.Sprintf("X-RateLimit-%s", key), value)
        },
    })

    app := ng.NewApp(
        ng.WithGuards(globalGuard, perRouteGuard),
    )

    type TestController struct{ ng.DefaultControllerInitializer }

    func (c *TestController) Simple() ng.Route {
        return ng.NewRoute(http.MethodGet, "/simple",
            ng.WithHandler(func(ctx context.Context) error {
                return ng.Respond(ctx, nghttp.NewRawResponse(200, []byte("simple")))
            }),
        )
    }

    func (c *TestController) Limited() ng.Route {
        return ng.NewRoute(http.MethodGet, "/limited",
            ratelimit.WithConfig(&ratelimit.Config{
                Limit:  3,
                Window: time.Second,
                Identifier: func(ctx context.Context) string { return "test-client" },
            }),
            ng.WithHandler(func(ctx context.Context) error {
                return ng.Respond(ctx, nghttp.NewRawResponse(200, []byte("limited")))
            }),
        )
    }

    app.AddController(&TestController{})
    app.Build()

    mux := http.NewServeMux()
    ngadapter.ServeMuxRegisterRoutes(app, mux)
    fmt.Println("Server running at http://localhost:8080")
    http.ListenAndServe(":8080", mux)
}

```

### Replace JSON Serialzer

```go
ratelimit.DefaultSerializer = // your custom serializer implementing ratelimit.Serializer
```

### Example of using Redis as Store

```go
package ratelimit

import (
	"context"
	"time"

	"github.com/foxie-io/ng-contrib/ratelimit/limiter"
	"github.com/redis/go-redis/v9"
)


// RedisStore implements Store using Redis.
type RedisStore struct {
	client *redis.UniversalClient
	prefix string
	ctx    context.Context
}

// NewRedisStore creates a new Redis-backed store.
func NewRedisStore(client *redis.UniversalClient, prefix string) *RedisStore {
	return &RedisStore{
		client: client,
		prefix: prefix,
		ctx:    context.Background(),
	}
}

// getKey returns Redis key for a client ID
func (r *RedisStore) getKey(id string) string {
	return r.prefix + ":" + id
}

// Get retrieves a limiter from Redis. Returns nil if not found.
func (r *RedisStore) Get(id string) limiter.Limiter {
	data, err := r.client.Get(r.ctx, r.getKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		panic(err) // unexpected Redis error
	}

	var wrapper limiter.LimiterJSON
	if err := limiter.DefaultSerializer.Unmarshal(data, &wrapper); err != nil {
		panic(err)
	}

	return wrapper.Limiter
}

// Set stores/updates a limiter in Redis.
func (r *RedisStore) Set(id string, l limiter.Limiter) {
	wrapper := limiter.LimiterJSON{Limiter: l}
	data, err := limiter.DefaultSerializer.Marshal(wrapper)
	if err != nil {
		panic(err)
	}

	// Optional: set a TTL to clean up old clients automatically
	ttl := time.Hour * 24
	if err := r.client.Set(r.ctx, r.getKey(id), data, ttl).Err(); err != nil {
		panic(err)
	}
}

// Delete removes client limiter data.
func (r *RedisStore) Delete(id string) {
	if err := r.client.Del(r.ctx, r.getKey(id)).Err(); err != nil {
		panic(err)
	}
}

// Keys returns all client IDs stored in Redis.
func (r *RedisStore) Keys() []string {
	keys, err := r.client.Keys(r.ctx, r.prefix+":*").Result()
	if err != nil {
		panic(err)
	}

	ids := make([]string, len(keys))
	for i, k := range keys {
		ids[i] = k[len(r.prefix)+1:] // strip prefix
	}
	return ids
}
```

```go
ratelimit.New(&ratelimit.Config{
    Store:  ratelimit.NewRedisStore(redisClient, "ratelimit"),
    ... // other config
})
```
