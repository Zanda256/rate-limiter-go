package rlgroup

import (
	"net/http"

	"github.com/Zanda256/rate-limiter-go/business/web/v1/mid"
	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/foundation/cache"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

type Config struct {
	TierConfig map[string]*ratelimiter.Tier
	KvStore    *cache.RedisCache
	Log        *logger.Logger
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"

	rateLmt := ratelimiter.NewRateLimiter(ratelimiter.RateLimiterConfig{
		Tier:    cfg.TierConfig["basic"],
		KvStore: cfg.KvStore,
		Log:     cfg.Log,
	})
	rateLmtMiddleware := mid.RateLimit(rateLmt)

	hdl := New(cfg.Log)
	app.Handle(http.MethodPost, version, "/limited", hdl.Limited, rateLmtMiddleware)
	app.Handle(http.MethodPost, version, "/unlimited", hdl.UnLimited)
}
