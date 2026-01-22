package router

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UsageStats struct {
	CostUSD      float64 `json:"cost_usd"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	RequestCount int     `json:"request_count"`
}

type UsageHistoryManager interface {
	RecordUsage(ctx context.Context, provider string, cost float64, inputTokens, outputTokens int) error
	GetDailyUsage(ctx context.Context, date string) (map[string]*UsageStats, error)
}

type RedisUsageHistoryManager struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisUsageHistoryManager(client *redis.Client, logger *zap.Logger) *RedisUsageHistoryManager {
	return &RedisUsageHistoryManager{
		client: client,
		logger: logger,
	}
}

func (m *RedisUsageHistoryManager) RecordUsage(ctx context.Context, provider string, cost float64, inputTokens, outputTokens int) error {
	date := time.Now().Format("2006-01-02")

	keys := []string{
		fmt.Sprintf("usage:v1:%s:%s", date, provider),
		fmt.Sprintf("usage:v1:%s:global", date),
	}

	for _, key := range keys {
		pipe := m.client.Pipeline()
		pipe.HIncrByFloat(ctx, key, "cost", cost)
		pipe.HIncrBy(ctx, key, "input_tokens", int64(inputTokens))
		pipe.HIncrBy(ctx, key, "output_tokens", int64(outputTokens))
		pipe.HIncrBy(ctx, key, "requests", 1)
		pipe.Expire(ctx, key, time.Hour*24*90) // Keep for 90 days

		_, err := pipe.Exec(ctx)
		if err != nil {
			m.logger.Error("Failed to record usage in Redis", zap.Error(err), zap.String("key", key))
		}
	}

	return nil
}

func (m *RedisUsageHistoryManager) GetDailyUsage(ctx context.Context, date string) (map[string]*UsageStats, error) {
	// Identify all provider usage keys for that date
	pattern := fmt.Sprintf("usage:v1:%s:*", date)
	keys, err := m.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	results := make(map[string]*UsageStats)
	for _, key := range keys {
		data, err := m.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		parts := append([]string{}, strings.Split(key, ":")...)
		provider := parts[len(parts)-1]

		stats := &UsageStats{}
		if val, ok := data["cost"]; ok {
			stats.CostUSD, _ = strconv.ParseFloat(val, 64)
		}
		if val, ok := data["input_tokens"]; ok {
			stats.InputTokens, _ = strconv.Atoi(val)
		}
		if val, ok := data["output_tokens"]; ok {
			stats.OutputTokens, _ = strconv.Atoi(val)
		}
		if val, ok := data["requests"]; ok {
			stats.RequestCount, _ = strconv.Atoi(val)
		}

		results[provider] = stats
	}

	return results, nil
}

type InMemoryUsageHistoryManager struct {
	// Not implemented for now, just a placeholder
}

func NewInMemoryUsageHistoryManager() *InMemoryUsageHistoryManager {
	return &InMemoryUsageHistoryManager{}
}

func (m *InMemoryUsageHistoryManager) RecordUsage(ctx context.Context, provider string, cost float64, inputTokens, outputTokens int) error {
	return nil
}

func (m *InMemoryUsageHistoryManager) GetDailyUsage(ctx context.Context, date string) (map[string]*UsageStats, error) {
	return make(map[string]*UsageStats), nil
}
