package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math/rand"
	"sync"
)

type WeightedRouter struct {
	providerManager  *providers.ProviderManager
	weights          map[string]int
	budgetManager    BudgetManager
	rateLimitManager RateLimitManager
	mu               sync.RWMutex
}

func (r *WeightedRouter) GetBudgetManager() BudgetManager {
	return r.budgetManager
}

func (r *WeightedRouter) GetRateLimitManager() RateLimitManager {
	return r.rateLimitManager
}

func NewWeightedRouter(providerManager *providers.ProviderManager, weights map[string]int, budget BudgetManager, rateLimit RateLimitManager) (*WeightedRouter, error) {
	if len(weights) == 0 {
		return nil, fmt.Errorf("weighted router requires at least one weight definition")
	}

	return &WeightedRouter{
		providerManager:  providerManager,
		weights:          weights,
		budgetManager:    budget,
		rateLimitManager: rateLimit,
	}, nil
}

func (r *WeightedRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allProviders []types.Provider
	if len(deps.Candidates) > 0 {
		allProviders = deps.Candidates
	} else {
		allProviders = r.providerManager.GetProviders()
	}
	var candidates []types.Provider

	for _, p := range allProviders {
		name := p.GetProviderName()
		weight, hasWeight := r.weights[name]

		if !hasWeight || weight <= 0 {
			continue
		}

		circuit, exists := deps.Circuits[name]
		if !exists || circuit.CanExecute() {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy providers available with positive weights")
	}

	totalWeight := 0
	for _, p := range candidates {
		totalWeight += r.weights[p.GetProviderName()]
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("total weight of available providers is zero")
	}

	target := rand.Intn(totalWeight)

	currentSum := 0
	for _, p := range candidates {
		weight := r.weights[p.GetProviderName()]
		currentSum += weight
		if currentSum > target {
			return &types.SelectedProviderOutput{
				Provider: p,
			}, nil
		}
	}

	return &types.SelectedProviderOutput{
		Provider: candidates[len(candidates)-1],
	}, nil
}

func (r *WeightedRouter) GetProviderManager() *providers.ProviderManager {
	return r.providerManager
}
