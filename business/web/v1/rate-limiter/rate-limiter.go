package ratelimiter

import (
	"github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter/tokenbucket"
	"github.com/Zanda256/rate-limiter-go/foundation/cache"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

// DefaultRateLimitPeriod Default Rate Limit Period in seconds
var DefaultRateLimitPeriod = 30
var DefaultRateLimitCapacity = 5

// Write the supported types of rate limiting and their implementations in this file
// Start with Leaky bucket
type RateLimiterImpl struct {
	*tokenbucket.BucketController
}

// type RateLimiter interface {
// 	CheckUserLimit(string) bool
// }

type Algo int

const (
	// LeakyBucket   = "LeakyBucket"
	TokenBucket = "TokenBucket"
	// FixedWindow   = "FixedWindow"
	// SlidingLog    = "SlidingLog"
	// SlidingWindow = "SlidingWindow"
)

type Tier struct {
	algo     string
	Period   int
	Capacity int
}

type RateLimiterConfig struct {
	Tier    *Tier
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
	})
	return &RateLimiterImpl{
		BucketController: t,
	}
}

func (rl *RateLimiterImpl) CheckUserLimit(userID string) bool {
	return rl.Accept(userID)
}
