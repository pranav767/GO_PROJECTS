package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection parameters with pooling settings.
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedis creates a new Redis client with connection pooling.
func NewRedis(cfg RedisConfig) (*redis.Client, error) {
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 20
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 5
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}
