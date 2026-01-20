package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
)

type ProviderFilter interface {
	Filter(ctx context.Context, candidates []types.Provider, input *types.SelectProviderInput) ([]types.Provider, error)
	Name() string
}

type PipelineRouter struct {
	baseRouter      Router
	filters         []ProviderFilter
	providerManager *providers.ProviderManager
}

func NewPipelineRouter(baseRouter Router, manager *providers.ProviderManager) *PipelineRouter {
	return &PipelineRouter{
		baseRouter:      baseRouter,
		filters:         make([]ProviderFilter, 0),
		providerManager: manager,
	}
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
		candidates, err = filter.Filter(ctx, candidates, input)
		if err != nil {
			return nil, fmt.Errorf("filter %s failed: %w", filter.Name(), err)
		}
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
