package router

import (
	"context"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"sync"
)

type RoundRobinRouter struct {
	providers []providers.Provider
	mu        sync.Mutex
	current   int
}

func (r *RoundRobinRouter) SelectProvider(ctx context.Context) providers.Provider {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider := r.providers[r.current]
	r.current = (r.current + 1) % len(r.providers)

	return provider
}

func NewRoundRobinRouter(config types.RouterConfig) (*RoundRobinRouter, error) {
	providers := providers.ConfigureProviders(config.Providers)

	return &RoundRobinRouter{
		current:   0,
		providers: providers,
	}, nil
}
