package main

import (
	"context"
	"fmt"

	"pickup-helper/internal/config"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Container holds all long-lived application dependencies wired at startup.
type Container struct {
	Cfg *config.Config
	DB  *sqlx.DB
	RDB *redis.Client
}

// BuildContainer manually wires dependencies (no DI framework).
// Callers must defer container.Close() to release resources.
func BuildContainer(cfg *config.Config) (*Container, error) {
	db, err := repository.NewMySQL(cfg)
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}
	rdb, err := repository.NewRedis(cfg)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}
	return &Container{Cfg: cfg, DB: db, RDB: rdb}, nil
}

// Close releases all resources held by the container.
func (c *Container) Close() error {
	var firstErr error
	if c.DB != nil {
		if err := c.DB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if c.RDB != nil {
		if err := c.RDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return fmt.Errorf("container close: %w", firstErr)
	}
	return nil
}

// PingAll checks DB and Redis reachability. Used by health readiness checks.
func (c *Container) PingAll(ctx context.Context) error {
	if err := repository.PingMySQL(ctx, c.DB); err != nil {
		return fmt.Errorf("mysql ping: %w", err)
	}
	if err := repository.PingRedis(ctx, c.RDB); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}
