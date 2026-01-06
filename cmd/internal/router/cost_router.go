package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math"
	"sync"
)

type CostRouter struct {
	providerManager *providers.ProviderManager
	mu              sync.Mutex
}

func (c *CostRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (types.Provider, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cheapestModel providers.ModelInfo
	var selectedProvider types.Provider
	cheapestPrice := math.Inf(1)

	allProviders := c.providerManager.GetProviders()

	for _, provider := range allProviders {
		providerName := provider.GetProviderName()
		allModels := providers.ListModelsByProvider(providerName)

		tokens, err := provider.CountTokens(ctx, deps.Messages)

		if err != nil {
			continue
		}

		for _, model := range allModels {
			cost, err := providers.CalculateCost(model.ID, tokens, 0)

			if err != nil {
				continue
			}

			if cost < cheapestPrice {
				cheapestPrice = cost
				cheapestModel = model
				selectedProvider = provider
			}
		}

		fmt.Print(cheapestModel)
	}

	return selectedProvider, nil
}

func NewCostRouter() *CostRouter {
	return nil
}

func (c *CostRouter) GetProviderManager() *providers.ProviderManager {
	return c.providerManager
}
