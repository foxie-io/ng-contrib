Quick Start

```go


// optional
type GlobalLimitID struct {
	ng.DefaultID[GlobalLimitID]
}

// create global limiter allowing 1000 requests per second
func createGlobalGuard() *ratelimit.Guard {
	return ratelimit.New(&ratelimit.Config{
		Limit:  1000,
		Window: time.Second,
		Identifier: func(ctx context.Context) string {
			return GLOBAL
		},
		ErrorHandler: func(ctx context.Context) error {
			return nghttp.NewErrTooManyRequests()
		},
		SetHeaderHandler: func(ctx context.Context, key, value string) {
			w := ng.MustLoad[http.ResponseWriter](ctx)
			key = fmt.Sprintf("X-G-RateLimit-%s", key)
			w.Header().Set(key, value)
		},
		MetadataKey: "GLOBAL", // optional
        GuardSkipperID: GlobalLimitID{}, // optional
	})
}


// optional
type VisitorLimiterID struct {
	ng.DefaultID[VisitorLimiterID]
}

// create per-visitor limiter allowing 60 requests per minute
func createVisitorGuard() *ratelimit.Guard {
	return ratelimit.New(&ratelimit.Config{
		Limit:  60,
		Window: time.Minute,
		Identifier: func(ctx context.Context) string {
			route := ng.GetContext(ctx).Route()
            fingerprint := // eg: ip address, user id, api key, etc.
            return fmt.Sprintf("PER_ROUTE_%s_%s_%s_%s", prefix, fingerprint, route.Method(), route.Path())
		},
		ErrorHandler: func(ctx context.Context) error {
			return nghttp.NewErrResourceExhausted()
		},
		SetHeaderHandler: func(ctx context.Context, key, value string) {
			w := ng.MustLoad[http.ResponseWriter](ctx)
			key = fmt.Sprintf("X-RateLimit-%s", key)
			w.Header().Set(key, value)
		},
        // MetadataKey: "PER_VISITOR", optional
        GuardSkipperID: VisitorLimiterID{}, // optional
	})
}

// App
func createApp() {
    app := ng.NewApp(
        ng.WithGuards(
            createGlobalGuard(), // global guard
            createVisitorGuard(), // per-visitor guard
        ),
        // ... other options
    )
}


// FreeTierController
type FreeTierController struct {
    ng.Controller
}

// FreeTierController initializer with 1000 requests per day limit overriding PER_ROUTE guard
func (c *FreeTierController) InitializeController() ng.Controller {
	return ng.NewController(
        ng.WithPrefix("/free-tier"),
        ratelimit.WithConfig(&ratelimit.Config{
            // 1000 requests per day
            Limit:  1000,
            Window: time.Hour * 24,

            // optional fallback to PER_VISITOR
            Identifier: func(ctx context.Context) string {
                return "free-tier-client-id"
            },
            // MetadataKey: "PER_VISITOR", optional
        }),
    )
}

// 1000 requests per day limited route
func (c *FreeTierController) LimitedAccess() ng.Route {
	return ng.NewRoute(http.MethodGet, "/limited-access",
		ng.WithHandler(func(ctx context.Context) error {
			return ng.Respond(ctx, nghttp.NewRawResponse(200, []byte("limited-access")))
		}),
	)
}


func (c *FreeTierController) Prices() ng.Route {
	return ng.NewRoute(http.MethodGet, "/prices",
        ratelimit.SkipRateLimit(), // skip all rate limits, includes createGlobalGuard and createVisitorGuard
        // or
        // ratelimit.WithSkip(VisitorLimiterID{}), // skip PER_VISITOR, so only GLOBAL applies
		ng.WithHandler(func(ctx context.Context) error {
			return ng.Respond(ctx, nghttp.NewRawResponse(200, []byte("prices")))
		}),
	)
}

// 100 requests per day limited route overriding controller limit
func (c *FreeTierController) OveriteLimitedAccess() ng.Route {
	return ng.NewRoute(http.MethodGet, "/special-limited-access",
		ratelimit.WithConfig(&ratelimit.Config{
		    // 100 requests per day
            Limit:  100,
            Window: time.Hour * 24,

            // optional fallback to config in FreeTierController
            Identifier: func(ctx context.Context) string {
                return "free-tier-client-id"
            },
            // ng.WithSkip(VisitorLimiterID{}), optional for skipping PER_VISITOR
            // ng.WithSkip(GlobalLimitID{}), optional for skipping GLOBAL
            // MetadataKey: "PER_VISITOR", optional override config with metadata same key
		}),
		ng.WithHandler(func(ctx context.Context) error {
			return ng.Respond(ctx, nghttp.NewRawResponse(200, []byte("special-limited-access")))
		}),
	)
}



```
