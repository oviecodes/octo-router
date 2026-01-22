package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"sync"

	"go.uber.org/zap"
)

type RoundRobinRouter struct {
	providerManager  *providers.ProviderManager
	budgetManager    BudgetManager
	rateLimitManager RateLimitManager
	mu               sync.Mutex
	current          int
}

func (r *RoundRobinRouter) GetBudgetManager() BudgetManager {
	return r.budgetManager
}

func (r *RoundRobinRouter) GetRateLimitManager() RateLimitManager {
	return r.rateLimitManager
}

func (r *RoundRobinRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var providerList []types.Provider
	if len(deps.Candidates) > 0 {
		providerList = deps.Candidates
	} else {
		providerList = r.providerManager.GetProviders()
	}

	if len(providerList) == 0 {
		return nil, fmt.Errorf("no available providers")
	}

	for range providerList {
		idx := r.current % len(providerList)
		provider := providerList[idx]
		r.current = (idx + 1) % len(providerList)

		breaker, ok := deps.Circuits[provider.GetProviderName()]

		if !ok || breaker.CanExecute() {
			return &types.SelectedProviderOutput{
				Provider: provider,
			}, nil
		}
	}

	return nil, fmt.Errorf("no healthy providers available")
}

func NewRoundRobinRouter(providerManager *providers.ProviderManager, budget BudgetManager, rateLimit RateLimitManager) (*RoundRobinRouter, error) {
	if providerManager == nil {
		return nil, fmt.Errorf("provider manager cannot be nil")
	}

	if providerManager.GetProviderCount() == 0 {
		return nil, fmt.Errorf("provider manager has no providers")
	}

	logger.Info("Round-robin router initialized",
		zap.Int("provider_count", providerManager.GetProviderCount()),
		zap.Strings("providers", providerManager.ListProviderNames()),
	)

	return &RoundRobinRouter{
		current:          0,
		providerManager:  providerManager,
		budgetManager:    budget,
		rateLimitManager: rateLimit,
	}, nil
}

func (r *RoundRobinRouter) GetProviderManager() *providers.ProviderManager {
	return r.providerManager
}
