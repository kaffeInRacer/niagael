package redis

import (
	"context"
	"fmt"
	"kaffein/order-service/config"
	"kaffein/order-service/utils/constants"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func New(cfg config.RedisConfig) (*Redis, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf(constants.ErrRedisAddrEmpty)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf(constants.ErrRedisPing, err)
	}

	return &Redis{Client: client}, nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf(constants.ErrRedisGet, key, err)
	}
	return val, nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf(constants.ErrRedisMarshal, key, err)
	}
	return nil
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}
