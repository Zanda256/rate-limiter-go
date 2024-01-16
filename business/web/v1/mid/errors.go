package mid

import (
	"context"
	"net/http"

	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/business/web/v1/response"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *logger.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Error(ctx, "message", "msg", err)

				var er response.ErrorDocument
				var status int

				switch {
				case ratelimiter.IsRateLimitError(err):
					er = response.ErrorDocument{
						Error: http.StatusText(http.StatusTooManyRequests),
					}
					status = http.StatusTooManyRequests

				default:
					er = response.ErrorDocument{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				// if web.IsShutdown(err) {
				// 	return err
				// }
			}

			return nil
		}

		return h
	}

	return m
}
