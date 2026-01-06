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

type SingleTenantResolver struct {
	App *App
}

func (s *SingleTenantResolver) GetConfig() *config.Config {
	return s.App.Config
}

func (s *SingleTenantResolver) GetRouter() router.Router {
	return s.App.Router
}

func (s *SingleTenantResolver) GetLogger() *zap.Logger {
	return s.App.Logger
}

func (s *SingleTenantResolver) GetCache() cache.Cache {
	return s.App.Cache
}

func (s *SingleTenantResolver) GetRetry() *resilience.Retry {
	return s.App.Retry
}

func (s *SingleTenantResolver) GetCircuitBreaker() map[string]types.CircuitBreaker {
	return s.App.Circuit
}

func (s *SingleTenantResolver) GetProviderManager() *providers.ProviderManager {
	return s.App.ProviderManager
}

func (s *SingleTenantResolver) GetFallbackChain() []string {
	return s.App.FallbackChain
}
