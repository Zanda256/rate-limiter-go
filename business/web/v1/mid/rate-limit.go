package mid

import "net/http"

// take in limiter config
func RateLimiter() web.Middleware {
	f := func(h web.Handler) web.Handler {
		m := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Do what ever you want with the request here
			// Call h()
			return h(ctx, w, r)
		}
		return m
	}
	return f
}
