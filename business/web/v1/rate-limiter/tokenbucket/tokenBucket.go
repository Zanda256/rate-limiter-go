package tokenbucket

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Zanda256/rate-limiter-go/business/web/v1/ratelimiter"
	"github.com/Zanda256/rate-limiter-go/foundation/cache"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

// TokenBucketConfig stores configuration for the token bucket
type TokenBucketConfig struct {
	Period   int
	UserID   string
	Capacity int
}

// TokenBucket is the data representation of a bucket
type TokenBucket struct {
	userID   string
	tokens   int
	lastSeen int64
	capacity int
	period   int
}

func (bc *BucketController) NewBucket(cfg TokenBucketConfig) TokenBucket {
	return TokenBucket{
		userID:   cfg.UserID,
		tokens:   cfg.Capacity,
		capacity: cfg.Capacity,
		lastSeen: time.Now().Unix(),
		period:   cfg.Period,
	}
}

// BucketController manages bucket creation, and state of individual buckets
type BucketController struct {
	Store *cache.RedisCache
	log   *logger.Logger
}

type BucketControllerConfig struct {
	Store    *cache.RedisCache
	log      *logger.Logger
	Period   int
	Capacity int
}

func NewBucketController(cfg BucketControllerConfig) *BucketController {
	return &BucketController{
		Store: cfg.Store,
		log:   cfg.log,
	}
}

func (bc *BucketController) Accept(userID string) bool {
	v, err := bc.checkBucket(userID)
	if err != nil {
		// If the key isn't present, create a new one
		if err.Error() == "key not found" { // move to string constatnt later
			buckt := bc.NewBucket(TokenBucketConfig{
				Period:   ratelimiter.DefaultRateLimitPeriod,
				UserID:   userID,
				Capacity: ratelimiter.DefaultRateLimitCapacity,
			})
			if err = bc.Store.StoreValue(context.Background(), userID, buckt, 30); err != nil {
				bc.log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
				return false
			}
			//If no error accept the request
			return true
		}
		// Other error type we log it and return false
		bc.log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
		return false
	}
	buckt, _ := v.(TokenBucket)
	if (time.Now().Unix() - buckt.lastSeen) > int64(buckt.period) {
		bc.refreshTokens(buckt)
		return true
	}
	if buckt.tokens == 0 {
		return false
	}
	buckt.tokens -= 1
	return true
}

//var store = map[string]int{}

func (bc *BucketController) checkBucket(UserID string) (any, error) {
	v, err := bc.Store.RetrieveValue(context.Background(), UserID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		//tb.Store.StoreValue(context.Background(), UserID)
		return nil, errors.New("key not found")
	}
	return v, nil
}

func (bc *BucketController) refreshTokens(bucket TokenBucket) {
	bucket.tokens = bucket.capacity
	bc.Store.StoreValue(context.Background(), bucket.userID, bucket, 30)
}
