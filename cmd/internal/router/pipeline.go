package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
)

type ProviderFilter interface {
	Filter(ctx context.Context, input *types.FilterInput) (*types.FilterOutput, error)
	Name() string
}

type PipelineRouter struct {
	baseRouter       Router
	filters          []ProviderFilter
	providerManager  *providers.ProviderManager
	budgetManager    BudgetManager
	rateLimitManager RateLimitManager
	usageHistory     UsageHistoryManager
}

func NewPipelineRouter(baseRouter Router, manager *providers.ProviderManager, budget BudgetManager, rateLimit RateLimitManager, history UsageHistoryManager) *PipelineRouter {
	return &PipelineRouter{
		baseRouter:       baseRouter,
		filters:          make([]ProviderFilter, 0),
		providerManager:  manager,
		budgetManager:    budget,
		rateLimitManager: rateLimit,
		usageHistory:     history,
	}
}

func (r *PipelineRouter) GetBudgetManager() BudgetManager {
	return r.budgetManager
}

func (r *PipelineRouter) GetRateLimitManager() RateLimitManager {
	return r.rateLimitManager
}

func (r *PipelineRouter) GetUsageHistoryManager() UsageHistoryManager {
	return r.usageHistory
}

func (r *PipelineRouter) AddFilter(filter ProviderFilter) {
	r.filters = append(r.filters, filter)
}

func (r *PipelineRouter) SelectProvider(ctx context.Context, input *types.SelectProviderInput) (*types.SelectedProviderOutput, error) {

	allProviders := r.providerManager.GetProviders()
	if len(allProviders) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var candidates []types.Provider
	for _, p := range allProviders {
		circuit, exists := input.Circuits[p.GetProviderName()]
		if !exists || circuit.CanExecute() {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}

	var err error
	for _, filter := range r.filters {
		filterOutput, err := filter.Filter(ctx, &types.FilterInput{
			Candidates: candidates,
			Messages:   input.Messages,
			Tier:       input.Tier,
		})
		if err != nil {
			return nil, fmt.Errorf("filter %s failed: %w", filter.Name(), err)
		}
		candidates = filterOutput.Candidates
		if len(candidates) == 0 {
			return nil, fmt.Errorf("filter %s filtered out all providers", filter.Name())
		}
	}

	input.Candidates = candidates
	output, err := r.baseRouter.SelectProvider(ctx, input)
	if err == nil {
		output.Candidates = candidates
	}
	return output, err
}

func (r *PipelineRouter) GetProviderManager() *providers.ProviderManager {
	return r.providerManager
}
