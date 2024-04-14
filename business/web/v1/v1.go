package v1

import (
	"os"

	"github.com/Zanda256/rate-limiter-go/business/data/cache"
	"github.com/Zanda256/rate-limiter-go/business/web/v1/mid"
	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Tiers    map[string]ratelimiter.Tier
	RedisKv  *cache.RedisCache
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
}

// RouteAdder defines behavior that sets the routes to bind for an instance
// of the service.
type RouteAdder interface {
	Add(app *web.App, cfg APIMuxConfig)
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig, routeAdder RouteAdder) *web.App {
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Errors(cfg.Log))

	routeAdder.Add(app, cfg)

	return app
}
