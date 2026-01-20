package app

import (
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"
	"llm-router/utils"
	"os"

	"go.uber.org/zap"
)

type ConfigResolver interface {
	GetConfig() *config.Config
	GetRouter() router.Router
	GetLogger() *zap.Logger
	GetCache() cache.Cache
	GetRetry() *resilience.Retry
	GetCircuitBreaker() map[string]types.CircuitBreaker
	GetProviderManager() *providers.ProviderManager
	GetFallbackChain() []string
}

type App struct {
	Config          *config.Config
	Router          router.Router
	Logger          *zap.Logger
	Cache           cache.Cache
	Retry           *resilience.Retry
	Circuit         map[string]types.CircuitBreaker
	ProviderManager *providers.ProviderManager
	FallbackChain   []string
}

var logger = utils.SetUpLogger()

func SetUpApp() *App {
	defer logger.Sync()

	logger.Info("Loading configs")
	cfg, err := config.LoadConfig()

	if err != nil {
		logger.Error("Failed to load config", zap.Error(err))
		os.Exit(1)
	}

	providers.InitializeModelRegistry(providers.GetDefaultCatalog(), cfg.Models.Catalog)
	latencyTracker := router.NewLatencyTracker()
	providerFactory := providers.NewProviderFactory()

	rawProviders := providerFactory.CreateProviders(cfg.GetProviderConfigWithExtras())

	var wrappedProviders []types.Provider
	for _, p := range rawProviders {
		wrappedProviders = append(wrappedProviders, router.NewLatencyMonitoringProvider(p, latencyTracker))
	}

	providerManager := providers.NewProviderManager(providerFactory)
	providerManager.SetProviders(wrappedProviders)

	llmRouter, fallback, err := initializeRouter(cfg, providerManager, latencyTracker)
	if err != nil {
		logger.Error("Failed to initialize router", zap.Error(err))
		os.Exit(1)
	}

	var cacheInstance cache.Cache

	if cfg.CacheConfig.Enabled {
		cacheInstance, err = cache.NewCacheClient(cfg.CacheConfig)

		if err != nil {
			logger.Error("Failed to initialize cache", zap.Error(err))
		}
	}

	resillienceConfig := cfg.GetResilienceConfigData()
	retry := resilience.NewRetryHandler(resillienceConfig.RetriesConfig, logger)
	circuit := initializeCircuitBreakers(cfg)

	// Create app with all dependencies
	app := &App{
		Config:          cfg,
		Router:          llmRouter,
		Logger:          logger,
		Cache:           cacheInstance,
		Retry:           retry,
		Circuit:         circuit,
		ProviderManager: providerManager,
		FallbackChain:   fallback,
	}

	return app
}

func initializeRouter(cfg *config.Config, providerManager *providers.ProviderManager, tracker *router.LatencyTracker) (router.Router, []string, error) {
	routerStrategy := cfg.GetRouterStrategy()

	logger.Info("Initializing router", zap.String("strategy", routerStrategy.Strategy))

	budgets := make(map[string]float64)
	for name, limits := range cfg.Limits.Providers {
		if limits.Budget > 0 {
			budgets[name] = limits.Budget
		}
	}

	llmRouter, fallback, err := router.ConfigureRouterStrategy(routerStrategy, providerManager, tracker, budgets)

	return llmRouter, fallback, err
}

func initializeCircuitBreakers(cfg *config.Config) map[string]types.CircuitBreaker {
	enabled := cfg.GetEnabledProviders()
	resillienceConfig := cfg.GetResilienceConfigData()
	providerNames := make([]string, len(enabled))

	for _, provider := range enabled {
		providerNames = append(providerNames, provider.Name)
	}

	circuit := resilience.NewCircuitBreakers(providerNames, resillienceConfig.CircuitBreakerConfig)
	return circuit
}
