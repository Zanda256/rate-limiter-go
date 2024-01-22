package tokenbucket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Zanda256/rate-limiter-go/business/data/cache"
	"github.com/Zanda256/rate-limiter-go/foundation/logger"
)

var ErrKeyNotFound = errors.New("key not found")

const timeFormat = time.RFC3339

// TokenBucketConfig stores configuration for the token bucket
type TokenBucketConfig struct {
	Period   int
	UserID   string
	Capacity int
}

// TokenBucket is the data representation of a bucket
type TokenBucket struct {
	UserID      string `json:"userID"`       // redis:"userID",
	Tokens      int    ` json:"tokens"`      // redis:"tokens",
	NextRefresh string ` json:"nextRefresh"` // redis:"lastSeen",
	Capacity    int    `json:"capacity"`     // redis:"capacity",
	Period      int    ` json:"period"`      // redis:"period",
}

// To be stored in redis, we need to implement this interface.
func (t TokenBucket) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(t)
	return
}

func UnmarshalBinarytoTB(data []byte, t *TokenBucket) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("unmarshal to TokenBucket{} failed %+v", err.Error()))
	}
	return nil
}

func (bc *BucketController) NewBucket(cfg TokenBucketConfig) TokenBucket {
	return TokenBucket{
		UserID:      cfg.UserID,
		Tokens:      cfg.Capacity,
		Capacity:    cfg.Capacity,
		NextRefresh: time.Now().Add(time.Duration(cfg.Period) * time.Second).Format(time.RFC3339),
		Period:      cfg.Period,
	}
}

// BucketController manages bucket creation, and state of individual buckets
type BucketController struct {
	Period, Cap int
	Store       *cache.RedisCache
	Log         *logger.Logger
}

type BucketControllerConfig struct {
	Store    *cache.RedisCache
	Log      *logger.Logger
	Period   int
	Capacity int
}

func NewBucketController(cfg BucketControllerConfig) *BucketController {
	return &BucketController{
		Period: cfg.Period,
		Cap:    cfg.Capacity,
		Store:  cfg.Store,
		Log:    cfg.Log,
	}
}

func (bc *BucketController) Accept(userID string) bool {
	v, err := bc.checkBucket(userID)
	if err != nil {
		// If the key isn't present, create a new one
		if errors.Is(err, ErrKeyNotFound) { // move to string constatnt later
			buckt := bc.NewBucket(TokenBucketConfig{
				Period:   bc.Period,
				UserID:   userID,
				Capacity: bc.Cap,
			})
			if _, err = bc.updateTokens(buckt); err != nil {
				bc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
				return false
			}
			bc.Log.Info(context.Background(), "successfully stored new value", buckt)
			//If no error accept the request
			return true
		}
		// Other error type we log it and return false
		bc.Log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
		return false
	}
	buckt := TokenBucket{}
	var (
		t  string
		ok bool
	)

	if t, ok = v.(string); !ok {
		bc.Log.Info(context.Background(), "cannot marshal retrieved value into string")
		return false
	}

	if err = UnmarshalBinarytoTB([]byte(t), &buckt); err != nil {
		bc.Log.Info(context.Background(), "cannot marshal retrieved value into TokenBucket")
		return false
	}
	bc.Log.Info(context.Background(), "successfully retrieved value", "buckt", buckt)

	nextRefresh, err := time.Parse(timeFormat, buckt.NextRefresh)
	if err != nil {
		bc.Log.Info(context.Background(), "cannot Parse time value")
		return false
	}

	// If user is ready for next token refresh, refill bucket
	if time.Now().After(nextRefresh) {
		if err = bc.refreshTokens(&buckt); err != nil {
			bc.Log.Error(context.Background(), "refresh tokens failed:", err.Error())
			return false
		}
		if _, err = bc.updateTokens(buckt); err != nil {
			bc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
			return false
		}
		return true
	}
	// User has no tokens left, return false
	if buckt.Tokens < 1 {
		return false
	} else {
		if _, err = bc.updateTokens(buckt); err != nil {
			bc.Log.Error(context.Background(), fmt.Sprintf("updateTokens: %s", err.Error()))
			return false
		}
		return true
	}
}

func (bc *BucketController) checkBucket(UserID string) (any, error) {
	v, err := bc.Store.RetrieveValue(context.Background(), UserID)
	if err != nil {
		bc.Log.Error(context.Background(), "checkbucket: %s", err.Error())
		return nil, err
	}
	if v == nil {
		bc.Log.Warn(context.Background(), "key not found")
		return nil, ErrKeyNotFound
	}
	bc.Log.Info(context.Background(), "bucket in check bucket: %+v", v)
	return v, nil
}

func (bc *BucketController) updateTokens(b TokenBucket) (TokenBucket, error) {
	// Decrement tokens by 1 and persist the result. If successful,accept and process the request.
	b.Tokens -= 1
	res, err := bc.Store.StoreValue(context.Background(), b.UserID, b, 30)
	if err != nil {
		bc.Log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
		return TokenBucket{}, err
	}
	buckt := res.(TokenBucket)
	return buckt, nil
}

func (bc *BucketController) refreshTokens(bucket *TokenBucket) error {
	bucket.Tokens = bucket.Capacity
	bucket.NextRefresh = time.Now().Add(time.Duration(bucket.Period) * time.Second).Format(timeFormat)
	bc.Log.Info(context.Background(), fmt.Sprintf("refreshing tokens: %+v", bucket))
	_, err := bc.Store.StoreValue(context.Background(), bucket.UserID, *bucket, 30)
	if err != nil {
		return err
	}
	return nil
}
