package router

import (
	"sync"

	"go.uber.org/zap"
)

type BudgetManager struct {
	mu     sync.RWMutex
	usage  map[string]float64
	limits map[string]float64
	logger *zap.Logger
}

func NewBudgetManager(limits map[string]float64, logger *zap.Logger) *BudgetManager {
	if limits == nil {
		limits = make(map[string]float64)
	}
	return &BudgetManager{
		usage:  make(map[string]float64),
		limits: limits,
		logger: logger,
	}
}

func (bm *BudgetManager) TrackUsage(provider string, cost float64) {
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

func (bm *BudgetManager) IsWithinBudget(provider string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	limit, exists := bm.limits[provider]
	if !exists {
		return true
	}

	return bm.usage[provider] < limit
}

func (bm *BudgetManager) GetUsage(provider string) float64 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.usage[provider]
}

func (bm *BudgetManager) ResetUsage(provider string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.usage, provider)
}
