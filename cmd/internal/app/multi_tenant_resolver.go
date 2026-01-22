package app

import (
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"

	"go.uber.org/zap"
)

type MultiTenantResolver struct {
	Logger *zap.Logger
}

func (m *MultiTenantResolver) GetConfig() *config.Config {
	// Extract API key from request
	// apiKey := c.GetHeader("X-API-Key")
	// Fetch tenant config from DB/cache
	// return fetchTenantConfig(apiKey)
	return nil
}

// When I implement multi-tenancy
// it might not be initializeRouter,
// it might be some other function that checks
// if a router is already cached for said user, then retrieve
// if not fetch cfg from database and configure router properly
func (m *MultiTenantResolver) GetRouter() router.Router {
	// cfg := m.GetConfig(c)
	// providerManager, _ := initializeProviderManager(cfg)
	// router, _ := initializeRouter(cfg, providerManager)
	// return router

	return nil
}

func (m *MultiTenantResolver) GetLogger() *zap.Logger {
	return m.Logger
}

func (m *MultiTenantResolver) GetCache() cache.Cache {
	return nil
}

func (m *MultiTenantResolver) GetRetry() *resilience.Retry {
	return nil
}

func (m *MultiTenantResolver) GetCircuitBreaker() map[string]types.CircuitBreaker {
	return nil
}

func (m *MultiTenantResolver) GetProviderManager() *providers.ProviderManager {
	return nil
}

func (m *MultiTenantResolver) GetFallbackChain() []string {
	return nil
}

func (m *MultiTenantResolver) Reload() error {
	return nil
}
