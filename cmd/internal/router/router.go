package router

import (
	"context"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"llm-router/utils"
	"os"
	"sync"
)

type RoundRobinRouter struct {
	providers []providers.Provider
	mu        sync.Mutex
	current   int
}

var logger = utils.SetUpLogger()

func (r *RoundRobinRouter) SelectProvider(ctx context.Context) providers.Provider {
	r.mu.Lock()
	defer r.mu.Unlock()

	provider := r.providers[r.current]
	r.current = (r.current + 1) % len(r.providers)

	return provider
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
