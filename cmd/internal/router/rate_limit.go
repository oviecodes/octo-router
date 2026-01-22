package router

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimitManager interface {
	Allow(ctx context.Context, key string, limit int) (bool, error)
}

type RedisRateLimitManager struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisRateLimitManager(client *redis.Client, logger *zap.Logger) RateLimitManager {
	return &RedisRateLimitManager{
		client: client,
		logger: logger,
	}
}

func (m *RedisRateLimitManager) Allow(ctx context.Context, key string, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil
	}

	now := time.Now()
	bucket := now.Format("2006-01-02 15:04")
	redisKey := fmt.Sprintf("ratelimit:%s:%s", key, bucket)

	count, err := m.client.Incr(ctx, redisKey).Result()
	if err != nil {
		m.logger.Error("Failed to increment rate limit counter", zap.Error(err), zap.String("key", redisKey))
		return true, err
	}

	if count == 1 {
		m.client.Expire(ctx, redisKey, time.Minute*2)
	}

	if count > int64(limit) {
		m.logger.Warn("Rate limit exceeded", zap.String("key", key), zap.Int64("count", count), zap.Int("limit", limit))
		return false, nil
	}

	return true, nil
}

type InMemoryRateLimitManager struct {
	// For MVP, we can just say everything is allowed or implement a simple map
}

func NewInMemoryRateLimitManager() RateLimitManager {
	return &InMemoryRateLimitManager{}
}

func (m *InMemoryRateLimitManager) Allow(ctx context.Context, key string, limit int) (bool, error) {
	return true, nil
}
