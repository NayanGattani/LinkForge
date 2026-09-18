package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis config: %w", err)
	}

	opt.PoolSize = 16
	opt.MinIdleConns = 2
	opt.ReadTimeout = 500 * time.Millisecond
	opt.WriteTimeout = 500 * time.Millisecond

	client := redis.NewClient(opt)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
