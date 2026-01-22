package app

import (
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"
	"sync/atomic"

	"go.uber.org/zap"
)

type SingleTenantResolver struct {
	App atomic.Pointer[App]
}

func (s *SingleTenantResolver) GetConfig() *config.Config {
	return s.App.Load().Config
}

func (s *SingleTenantResolver) GetRouter() router.Router {
	return s.App.Load().Router
}

func (s *SingleTenantResolver) GetLogger() *zap.Logger {
	return s.App.Load().Logger
}

func (s *SingleTenantResolver) GetCache() cache.Cache {
	return s.App.Load().Cache
}

func (s *SingleTenantResolver) GetRetry() *resilience.Retry {
	return s.App.Load().Retry
}

func (s *SingleTenantResolver) GetCircuitBreaker() map[string]types.CircuitBreaker {
	return s.App.Load().Circuit
}

func (s *SingleTenantResolver) GetProviderManager() *providers.ProviderManager {
	return s.App.Load().ProviderManager
}

func (s *SingleTenantResolver) GetFallbackChain() []string {
	return s.App.Load().FallbackChain
}

func (s *SingleTenantResolver) Reload() error {
	newApp, err := SetUpApp()
	if err != nil {
		return err
	}
	s.App.Store(newApp)
	return nil
}
