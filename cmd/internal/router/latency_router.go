package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math/rand"
)

type LatencyRouter struct {
	providerManager *providers.ProviderManager
	tracker         *LatencyTracker
	budgetManager   BudgetManager
}

func NewLatencyRouter(providerManager *providers.ProviderManager, tracker *LatencyTracker, budget BudgetManager) (*LatencyRouter, error) {
	return &LatencyRouter{
		providerManager: providerManager,
		tracker:         tracker,
		budgetManager:   budget,
	}, nil
}

func (r *LatencyRouter) GetBudgetManager() BudgetManager {
	return r.budgetManager
}

func (r *LatencyRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error) {
	var allProviders []types.Provider
	if len(deps.Candidates) > 0 {
		allProviders = deps.Candidates
	} else {
		allProviders = r.providerManager.GetProviders()
	}

	if len(allProviders) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var candidates []types.Provider
	for _, p := range allProviders {
		circuit, exists := deps.Circuits[p.GetProviderName()]
		if !exists || circuit.CanExecute() {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	var bestProvider types.Provider
	var bestScore float64 = -1

	var unknownProviders []types.Provider

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for _, p := range candidates {
		score := r.tracker.GetLatencyScore(p.GetProviderName())

		if score == 0 {
			unknownProviders = append(unknownProviders, p)
			continue
		}

		if bestScore == -1 || score < bestScore {
			bestScore = score
			bestProvider = p
		}
	}

	if len(unknownProviders) > 0 {
		idx := rand.Intn(len(unknownProviders))
		return &types.SelectedProviderOutput{
			Provider: unknownProviders[idx],
		}, nil
	}

	if bestProvider != nil {
		return &types.SelectedProviderOutput{
			Provider: bestProvider,
		}, nil
	}

	// Fallback: This should technically not be reached unless all scores > 0 AND no best found (impossible logic)
	// But safe fallback is first candidate
	return &types.SelectedProviderOutput{
		Provider: candidates[0],
	}, nil
}

func (r *LatencyRouter) GetProviderManager() *providers.ProviderManager {
	return r.providerManager
}
