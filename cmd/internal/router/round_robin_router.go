package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"os"
	"sync"
)

type RoundRobinRouter struct {
	providers []types.Provider
	mu        sync.Mutex
	current   int
}

// var logger = utils.SetUpLogger()

func (r *RoundRobinRouter) SelectProvider(ctx context.Context, circuits map[string]types.CircuitBreaker) (types.Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for range r.providers {
		provider := r.providers[r.current]
		r.current = (r.current + 1) % len(r.providers)

		if _, ok := circuits[provider.GetProviderName()]; !ok {
			continue
		}

		return provider, nil
	}

	return nil, fmt.Errorf("No available providers")
}

func NewRoundRobinRouter(config types.RouterConfig) (*RoundRobinRouter, error) {
	providers := providers.ConfigureProviders(config.Providers)

	// check length  of providers if == 0; throw error and exit the application
	if len(providers) == 0 {
		logger.Error("Could not set up any providers")
		os.Exit(1)
	}

	return &RoundRobinRouter{
		current:   0,
		providers: providers,
	}, nil
}
