package cache

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisCache struct {
	*redis.Client
}

func NewRedisCache(redisAddress string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &RedisCache{
		rdb,
	}
}

// defer rdb.Close()

// status, err := rdb.Ping(ctx).Result()
//     if err != nil {
//         log.Fatalln("Redis connection was refused")
//     }
//     fmt.Println(status)

func (rc *RedisCache) StoreValue(ctx context.Context, key string, value any, ttl int) error {
	err := rc.Set(ctx, "key", "value", time.Minute*time.Duration(ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) RetrieveValue(ctx context.Context, key string, value any) (any, error) {
	val, err := rc.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return val, nil

	// val2, err := rc.Get(ctx, "key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }
}
