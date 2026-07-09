package repository

import (
	"context"
	"fmt"

	"pickup-helper/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedis constructs a redis.Client using cfg.
func NewRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	// Eagerly verify connectivity so failures surface at startup.
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

// PingRedis verifies the redis connection is alive.
func PingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return fmt.Errorf("nil redis")
	}
	return rdb.Ping(ctx).Err()
}
