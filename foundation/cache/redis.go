package cache

import (
	"context"
	// redis "github.com/redis/go-redis/v9"
)

type RedisCache struct {
	m map[string]any
	// *redis.Client
}

func NewRedisCache(redisAddress string) *RedisCache {
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })
	return &RedisCache{
		m: make(map[string]any),
	}

	// return &RedisCache{
	// 	rdb,
	// }
}

func (rc *RedisCache) StoreValue(ctx context.Context, key string, value any, ttl int) error {
	rc.m[key] = value
	return nil
}

func (rc *RedisCache) RetrieveValue(ctx context.Context, key string) (any, error) {
	val, found := rc.m[key]
	if !found {
		return nil, nil
	}
	return val, nil
}

// defer rdb.Close()

// status, err := rdb.Ping(ctx).Result()
//     if err != nil {
//         log.Fatalln("Redis connection was refused")
//     }
//     fmt.Println(status)

// func (rc *RedisCache) StoreValue(ctx context.Context, key string, value any, ttl int) error {
// 	err := rc.Set(ctx, "key", value, time.Minute*time.Duration(ttl)).Err()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (rc *RedisCache) RetrieveValue(ctx context.Context, key string) (any, error) {
// 	val, err := rc.Get(ctx, key).Result()
// 	if err == redis.Nil {
// 		return nil, nil
// 	} else if err != nil {
// 		return nil, err
// 	}
// 	return val, nil
// }
