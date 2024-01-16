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

// TokenBucketConfig stores configuration for the token bucket
type TokenBucketConfig struct {
	Period   int
	UserID   string
	Capacity int
}

// TokenBucket is the data representation of a bucket
type TokenBucket struct {
	UserID   string `json:"userID"`    // redis:"userID",
	Tokens   int    ` json:"tokens"`   // redis:"tokens",
	LastSeen int64  ` json:"lastSeen"` // redis:"lastSeen",
	Capacity int    `json:"capacity"`  // redis:"capacity",
	Period   int    ` json:"period"`   // redis:"period",
}

// To be stored in redis, we need to implement this interface.
func (t TokenBucket) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(t)
	return
}

func UnmarshalBinarytoTB(data []byte, t *TokenBucket) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		fmt.Printf("unmarshal to TokenBucket{} failed %+v", err.Error())
		return errors.New(fmt.Sprintf("unmarshal to TokenBucket{} failed %+v", err.Error()))
	}
	fmt.Printf("\nsuccessfully converted value %+v from %T to %T with value %+v\n", data, data, t, t)
	return nil
}

func (t TokenBucket) UnmarshalString(data string) error {
	tmp := TokenBucket{}
	err := json.Unmarshal([]byte(data), &tmp)
	if err != nil {
		fmt.Printf("unmarshal to TokenBucket{} failed %+v", err.Error())
		return errors.New(fmt.Sprintf("unmarshal to TokenBucket{} failed %+v", err.Error()))
	}
	fmt.Printf("\nsuccessfully converted value %+v from %T to string %s\n", data, data, t)
	return nil
}

func (bc *BucketController) NewBucket(cfg TokenBucketConfig) TokenBucket {
	return TokenBucket{
		UserID:   cfg.UserID,
		Tokens:   cfg.Capacity,
		Capacity: cfg.Capacity,
		LastSeen: time.Now().Unix(),
		Period:   cfg.Period * 60,
	}
}

//	type KvStore interface{
//		StoreValue(ctx context.Context, key string, value any)
//	}
//
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
		if err.Error() == "key not found" { // move to string constatnt later
			buckt := bc.NewBucket(TokenBucketConfig{
				Period:   bc.Period,
				UserID:   userID,
				Capacity: bc.Cap,
			})
			if err = bc.Store.StoreValue(context.Background(), userID, buckt, 30); err != nil {
				bc.Log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
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
	fmt.Printf("\ntype after retrieval: %T\nvalue after retrieval: %+v\n", v, v)
	buckt := TokenBucket{}
	var (
		t  string
		ok bool
	)
	switch v.(type) {
	case string:
		fmt.Printf("\ntype is string\n")
		t, ok = v.(string)
		if t, ok = v.(string); !ok {
			fmt.Printf("\ncannot marshal retrieved value into string\n")
		}
		fmt.Printf("\nsuccessfully converted value %+v from %T to string %s\n", v, v, t)
	default:
		fmt.Printf("\nunkwown type\n")
	}

	if err = UnmarshalBinarytoTB([]byte(t), &buckt); err != nil {
		fmt.Printf("\ncannot marshal retrieved value into TokenBucket\n")
	}
	bc.Log.Info(context.Background(), "successfully retrieved value", "buckt", buckt)
	fmt.Printf("\ntime.Now().Unix() = %d\nbuckt.LastSeen = %d\n", time.Now().Unix(), buckt.LastSeen)
	fmt.Printf("\ntime diff = %d\nbuckt.Period = %d\n", (time.Now().Unix() - buckt.LastSeen), buckt.Period)
	// User hasn't visited in a long time or new user, create them a token bucket
	if (time.Now().Unix() - buckt.LastSeen) > int64(buckt.Period) {
		bc.Log.Info(context.Background(), "refreshing tokens")
		if err = bc.refreshTokens(buckt); err != nil {
			bc.Log.Error(context.Background(), "refresh tokens failed:", err.Error())
			return false
		}
		return true
	}

	// User has no tokens left, return false
	if buckt.Tokens == 0 {
		return false
	}

	// Decrement tokens by 1 and persist the result. If successful, accept and process the request
	buckt.Tokens -= 1
	buckt.LastSeen = time.Now().Unix()
	if err = bc.Store.StoreValue(context.Background(), userID, buckt, 30); err != nil {
		bc.Log.Error(context.Background(), fmt.Sprintf("Store bucket value failed: %s", err.Error()))
		return false
	}
	fmt.Printf("\nbucket in Accept%+v\n", buckt)
	return true
}

//var store = map[string]int{}

func (bc *BucketController) checkBucket(UserID string) (any, error) {
	fmt.Printf("\nUserID: %s\n", UserID)
	v, err := bc.Store.RetrieveValue(context.Background(), UserID)
	if err != nil {
		fmt.Printf("\ncheckbucket: %s\n", err.Error())
		//fmt.Printf("\nerr: %+v\n", err)
		return nil, err
	}
	if v == nil {
		fmt.Printf("\nkey not found\n")
		//tb.Store.StoreValue(context.Background(), UserID)
		return nil, errors.New("key not found")
	}
	fmt.Printf("\nbucket in check bucket: %+v\n", v)
	return v, nil
}

func (bc *BucketController) refreshTokens(bucket TokenBucket) error {
	bucket.Tokens = bucket.Capacity
	return bc.Store.StoreValue(context.Background(), bucket.UserID, bucket, 30)
}
