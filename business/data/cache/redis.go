package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisAddress string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress, //"localhost:6379",
		Password: "",           // no password set
		DB:       0,            // use default DB
	})

	status, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("Redis connection was refused")
	}
	fmt.Println(status)

	return &RedisCache{
		client: rdb,
	}
}

// defer rdb.Close()

// status, err := rdb.Ping(ctx).Result()
//     if err != nil {
//         log.Fatalln("Redis connection was refused")
//     }
//     fmt.Println(status)

func (rc *RedisCache) StoreValue(ctx context.Context, key string, value any, ttl int) error {
	err := rc.client.Set(ctx, key, value, time.Minute*time.Duration(ttl)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisCache) RetrieveValue(ctx context.Context, key string) (any, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	fmt.Printf("\nbucket in RetrieveValue: %+v\n", val)
	return val, nil
}
