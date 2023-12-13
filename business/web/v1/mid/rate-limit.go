package mid

import (
	"context"
	"errors"
	"net/http"

	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

// take in limiter config
func RateLimit(rlmt *ratelimiter.RateLimiterImpl) web.Middleware {
	f := func(h web.Handler) web.Handler {
		m := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Do what ever you want with the request here
			q := r.URL.Query()
			// fmt.Println(q.Get("b"))
			user := q.Get("user")
			if !rlmt.CheckUserLimit(user) {
				return errors.New("limit exceeded")
			}
			// Call h()
			return h(ctx, w, r)
		}
		return m
	}
	return f
}
