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
	providerManager *providers.ProviderManager
	mu              sync.Mutex
	current         int
}

func (r *RoundRobinRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (types.Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	providerList := r.providerManager.GetProviders()

	for range providerList {
		provider := providerList[r.current]
		r.current = (r.current + 1) % len(providerList)

		breaker, ok := deps.Circuits[provider.GetProviderName()]

		if !ok || !breaker.CanExecute() {
			continue
		}

		return provider, nil
	}

	return nil, fmt.Errorf("no available providers")
}

func NewRoundRobinRouter(providerManager *providers.ProviderManager) (*RoundRobinRouter, error) {
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
		current:         0,
		providerManager: providerManager,
	}, nil
}

func (r *RoundRobinRouter) GetProviderManager() *providers.ProviderManager {
	return r.providerManager
}
