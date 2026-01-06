package router

import (
	"context"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"sync"
)

type CostRouter struct {
	providerManager *providers.ProviderManager
	mu              sync.Mutex
}

func (c *CostRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (types.Provider, error) {

	allProviders := c.providerManager.GetProviders()

	for _, provider := range allProviders {
		providerName := provider.GetProviderName()
		providers.ListModelsByProvider(providerName)

		// provider.CountTokens()
	}

	return nil, nil
}

func NewCostRouter() *CostRouter {
	return nil
}

func (c *CostRouter) GetProviderManager() *providers.ProviderManager {
	return c.providerManager
}
