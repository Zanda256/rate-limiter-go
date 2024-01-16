package handlers

import (
	rlgroup "github.com/Zanda256/rate-limiter-go/app/services/rate-limiter/handlers/rl-group"
	v1 "github.com/Zanda256/rate-limiter-go/business/web/v1"
	ratelimiter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter"
	"github.com/Zanda256/rate-limiter-go/foundation/web"
)

type Routes struct{}

// Add implements the RouterAdder interface to add all routes.
func (Routes) Add(app *web.App, apiCfg v1.APIMuxConfig) {
	rlgroup.Routes(app, rlgroup.Config{
		TierConfig: map[string]*ratelimiter.Tier{
			"basic": {
				Algo:     ratelimiter.TokenBucket,
				Period:   1,
				Capacity: 5,
			},
		},
		Log:     apiCfg.Log,
		KvStore: apiCfg.RedisKv,
	})
}
