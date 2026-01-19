package router

import (
	"llm-router/cmd/internal/providers"
)

type WeightedRouter struct {
	providerManager *providers.ProviderManager
	// weightsConfig
}

func NewWeightedRouter(providerManager *providers.ProviderManager, routerConfig map[string]int) (*WeightedRouter, error) {

	return &WeightedRouter{
		providerManager: providerManager,
	}, nil

}
