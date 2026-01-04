package providers

import (
	"context"
	"fmt"
	"llm-router/types"
	"llm-router/utils"
	"sync"

	"go.uber.org/zap"
)

var managerLogger = utils.SetUpLogger()

type ProviderManager struct {
	providers []types.Provider
	factory   *ProviderFactory
	mu        sync.RWMutex
}

func NewProviderManager(factory *ProviderFactory) *ProviderManager {
	return &ProviderManager{
		providers: make([]types.Provider, 0),
		factory:   factory,
	}
}

func (pm *ProviderManager) Initialize(configs []types.ProviderConfigWithExtras) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	providers := pm.factory.CreateProviders(configs)

	if len(providers) == 0 {
		return fmt.Errorf("no providers could be initialized")
	}

	pm.providers = providers

	managerLogger.Info("Provider manager initialized",
		zap.Int("total_providers", len(pm.providers)),
	)

	return nil
}

func (pm *ProviderManager) GetProviders() []types.Provider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent external modification
	providersCopy := make([]types.Provider, len(pm.providers))
	copy(providersCopy, pm.providers)

	return providersCopy
}

func (pm *ProviderManager) GetProvider(name string) (types.Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, provider := range pm.providers {
		if provider.GetProviderName() == name {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("provider %s not found", name)
}

func (pm *ProviderManager) GetProviderCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.providers)
}

func (pm *ProviderManager) AddProvider(config types.ProviderConfigWithExtras) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	provider, err := pm.factory.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Check if provider already exists
	for _, p := range pm.providers {
		if p.GetProviderName() == provider.GetProviderName() {
			return fmt.Errorf("provider %s already exists", provider.GetProviderName())
		}
	}

	pm.providers = append(pm.providers, provider)

	managerLogger.Info("Provider added",
		zap.String("provider", provider.GetProviderName()),
	)

	return nil
}

func (pm *ProviderManager) RemoveProvider(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for i, provider := range pm.providers {
		if provider.GetProviderName() == name {
			// Remove provider from slice
			pm.providers = append(pm.providers[:i], pm.providers[i+1:]...)

			managerLogger.Info("Provider removed",
				zap.String("provider", name),
			)

			return nil
		}
	}

	return fmt.Errorf("provider %s not found", name)
}

// HealthCheck performs a basic health check on all providers
func (pm *ProviderManager) HealthCheck(ctx context.Context) map[string]bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	health := make(map[string]bool)

	for _, provider := range pm.providers {
		// Simple health check: just verify provider exists and has a name
		// More sophisticated checks could be added here
		name := provider.GetProviderName()
		health[name] = name != ""
	}

	return health
}

func (pm *ProviderManager) ListProviderNames() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, len(pm.providers))
	for i, provider := range pm.providers {
		names[i] = provider.GetProviderName()
	}

	return names
}

// SetProvidersForTest sets providers directly (for testing only)
// This bypasses the factory and is only intended for unit tests
func (pm *ProviderManager) SetProvidersForTest(providers []types.Provider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.providers = providers
}
