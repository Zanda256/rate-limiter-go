package ratelimiter

import (
	"github.com/Zanda256/rate-limiter-go/business/web/v1/rate-limiter/tokenbucket"
	"github.com/Zanda256/rate-limiter-go/foundation/cache"
)

// DefaultRateLimitPeriod Default Rate Limit Period in seconds
var DefaultRateLimitPeriod = 30
var DefaultRateLimitCapacity = 5

// Write the supported types of rate limiting and their implementations in this file
// Start with Leaky bucket
type RateLimiterImpl struct {
	tokenbucket.BucketController
}
type RateLimiter interface {
	Accept(string) bool
}

type Algo int

const (
	LeakyBucket   = "LeakyBucket"
	TokenBucket   = "TokenBucket"
	FixedWindow   = "FixedWindow"
	SlidingLog    = "SlidingLog"
	SlidingWindow = "SlidingWindow"
)

type RateLimiterConfig struct {
	kvStore *cache.RedisCache
}

func NewRateLimiter(cfg RateLimiterConfig) RateLimiter {
	return tokenbucket.NewBucketController(tokenbucket.BucketControllerConfig{
		Store:    cfg.kvStore,
		Period:   DefaultRateLimitPeriod,
		Capacity: DefaultRateLimitCapacity,
	})
}
