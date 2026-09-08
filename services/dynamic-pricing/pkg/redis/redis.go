// Package redis menyediakan Redis Client initialization.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kaffein/dynamic-pricing-service/utils/constants"
	"time"

	"github.com/redis/go-redis/v9"

	"kaffein/dynamic-pricing-service/config"
)

type Cache struct {
	Client *redis.Client
	prefix string
	ttl    time.Duration
}

// NewCache membuat Redis Client baru dari konfigurasi.
func NewCache(cfg *config.Config, prefix string, ttl time.Duration) (*Cache, error) {
	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf(constants.ErrRedisAddrEmpty)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &Cache{
		Client: client,
		prefix: prefix,
		ttl:    ttl,
	}, nil
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.Client.Get(ctx, fmt.Sprintf("%s:%s", c.prefix, key)).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf(constants.ErrRedisGet, key, err)
	}

	return val, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf(constants.ErrRedisMarshal, key, err)
	}

	return c.Client.Set(ctx, fmt.Sprintf("%s:%s", c.prefix, key), data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, fmt.Sprintf("%s:%s", c.prefix, key)).Err()
}

func (c *Cache) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error)) ([]byte, error) {
	data, err := c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return data, nil
	}

	value, err := fn()
	if err != nil {
		return nil, err
	}

	if err := c.Set(ctx, key, value); err != nil {
		return nil, err
	}

	return json.Marshal(value)
}
