package router

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type BudgetManager interface {
	TrackUsage(provider string, cost float64)
	IsWithinBudget(provider string) bool
	GetUsage(provider string) float64
	ResetUsage(provider string)
}

type InMemoryBudgetManager struct {
	mu     sync.RWMutex
	usage  map[string]float64
	limits map[string]float64
	logger *zap.Logger
}

func NewInMemoryBudgetManager(limits map[string]float64, logger *zap.Logger) BudgetManager {
	if limits == nil {
		limits = make(map[string]float64)
	}
	return &InMemoryBudgetManager{
		usage:  make(map[string]float64),
		limits: limits,
		logger: logger,
	}
}

func (bm *InMemoryBudgetManager) TrackUsage(provider string, cost float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.usage[provider] += cost

	limit, exists := bm.limits[provider]
	if exists && bm.usage[provider] >= limit {
		bm.logger.Warn("Provider has exceeded budget limit",
			zap.String("provider", provider),
			zap.Float64("usage", bm.usage[provider]),
			zap.Float64("limit", limit),
		)
	}
}

func (bm *InMemoryBudgetManager) IsWithinBudget(provider string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	limit, exists := bm.limits[provider]
	if !exists {
		return true
	}

	return bm.usage[provider] < limit
}

func (bm *InMemoryBudgetManager) GetUsage(provider string) float64 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.usage[provider]
}

func (bm *InMemoryBudgetManager) ResetUsage(provider string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.usage, provider)
}

type RedisBudgetManager struct {
	client *redis.Client
	limits map[string]float64
	logger *zap.Logger
	ctx    context.Context
}

func NewRedisBudgetManager(client *redis.Client, limits map[string]float64, logger *zap.Logger) BudgetManager {
	return &RedisBudgetManager{
		client: client,
		limits: limits,
		logger: logger,
		ctx:    context.Background(),
	}
}

func (bm *RedisBudgetManager) getRedisKey(provider string) string {
	return fmt.Sprintf("budget:total:%s", provider)
}

func (bm *RedisBudgetManager) TrackUsage(provider string, cost float64) {
	key := bm.getRedisKey(provider)
	newUsage, err := bm.client.IncrByFloat(bm.ctx, key, cost).Result()
	if err != nil {
		bm.logger.Error("Failed to track usage in Redis", zap.Error(err), zap.String("provider", provider))
		return
	}

	limit, exists := bm.limits[provider]
	if exists && newUsage >= limit {
		bm.logger.Warn("Provider has exceeded budget limit (Redis)",
			zap.String("provider", provider),
			zap.Float64("usage", newUsage),
			zap.Float64("limit", limit),
		)
	}
}

func (bm *RedisBudgetManager) IsWithinBudget(provider string) bool {
	limit, exists := bm.limits[provider]
	if !exists {
		return true
	}

	usage := bm.GetUsage(provider)
	return usage < limit
}

func (bm *RedisBudgetManager) GetUsage(provider string) float64 {
	key := bm.getRedisKey(provider)
	val, err := bm.client.Get(bm.ctx, key).Float64()
	if err != nil {
		if err != redis.Nil {
			bm.logger.Error("Failed to get usage from Redis", zap.Error(err), zap.String("provider", provider))
		}
		return 0
	}
	return val
}

func (bm *RedisBudgetManager) ResetUsage(provider string) {
	key := bm.getRedisKey(provider)
	bm.client.Del(bm.ctx, key)
}
