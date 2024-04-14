package ratelimiter

import (
	"github.com/Zanda256/rate-limiter-go/business/data/cache"
	fixedwindowcounter "github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter/FixedWindowCounter"
	"github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter/tokenbucket"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

// DefaultRateLimitPeriod Default Rate Limit Period in seconds
var DefaultRateLimitPeriod = 30
var DefaultRateLimitCapacity = 5

// Write the supported types of rate limiting and their implementations in this file
// Start with Leaky bucket
type RateLimiterImpl struct {
	*tokenbucket.BucketController
	*fixedwindowcounter.WindowController
}

// type RateLimiter interface {
// 	CheckUserLimit(string) bool
// }

type Algo int

const (
	// LeakyBucket   = "LeakyBucket"
	TokenBucket = "TokenBucket"
	FixedWindow = "FixedWindow"
	// SlidingLog    = "SlidingLog"
	// SlidingWindow = "SlidingWindow"
)

type Tier struct {
	Algo     string `json:"algo"`
	Period   int    `json:"period"`
	Capacity int    `json:"capacity"`
}

// {
// 	"basic":"{
// 		"algo":"",
// 		"period":"",
// 		"capacity":""
// 	}"
// 	"premium":"{
// 		"algo":"",
// 		"period":"",
// 		"capacity":""
// 	}"
// }

type RateLimiterConfig struct {
	Tier    Tier
	KvStore *cache.RedisCache
	Log     *logger.Logger
}

func NewRateLimiter(cfg RateLimiterConfig) *RateLimiterImpl {
	t := tokenbucket.NewBucketController(tokenbucket.BucketControllerConfig{
		Store: cfg.KvStore,
		Period: func() int {
			if cfg.Tier.Period == 0 {
				return DefaultRateLimitPeriod
			}
			return cfg.Tier.Period
		}(),
		Capacity: func() int {
			if cfg.Tier.Capacity == 0 {
				return DefaultRateLimitPeriod
			}
			return cfg.Tier.Capacity
		}(),
		Log: cfg.Log,
	})
	fxW := fixedwindowcounter.NewWindowController(fixedwindowcounter.WindowControllerConfig{
		Store: cfg.KvStore,
		Log:   cfg.Log,
		MaxTokens: func() int {
			if cfg.Tier.Capacity == 0 {
				return DefaultRateLimitPeriod
			}
			return cfg.Tier.Capacity
		}(),
		WindowSize: func() int64 {
			if cfg.Tier.Period == 0 {
				return int64(DefaultRateLimitPeriod)
			}
			return int64(cfg.Tier.Period)
		}(),
	})
	return &RateLimiterImpl{
		BucketController: t,
		WindowController: fxW,
	}
}

func (rl *RateLimiterImpl) CheckUserLimit(userID string) bool {
	return rl.WindowController.Accept(userID)
}
