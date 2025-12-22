package app

import (
	"fmt"
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"
	"llm-router/utils"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ConfigResolver interface {
	GetConfig(c *gin.Context) *config.Config
	GetRouter(c *gin.Context) router.Router
	GetLogger(c *gin.Context) *zap.Logger
	GetCache(c *gin.Context) cache.Cache
	GetRetry(c *gin.Context) *resilience.Retry
}

type App struct {
	Config  *config.Config
	Router  router.Router
	Logger  *zap.Logger
	Cache   cache.Cache
	Retry   *resilience.Retry
	Circuit map[string]*resilience.Circuit
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

	// Initialize router once at startup
	llmRouter, err := initializeRouter(cfg)
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

	fmt.Printf("Current cache instance %v \n", cacheInstance)

	resillienceConfig := cfg.GetResilienceConfigData()
	retry := resilience.NewRetryHandler(resillienceConfig.RetriesConfig, logger)
	circuit := initializeCircuitBreakers(cfg)

	// Create app with all dependencies
	app := &App{
		Config:  cfg,
		Router:  llmRouter,
		Logger:  logger,
		Cache:   cacheInstance,
		Retry:   retry,
		Circuit: circuit,
	}

	return app
}

func initializeRouter(cfg *config.Config) (router.Router, error) {
	enabled := cfg.GetEnabledProviders()
	routerStrategy := cfg.GetRouterStrategy()

	fmt.Printf("All Routing configs %v \n", *cfg)

	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled providers found in config")
	}

	routerConfig := types.RouterConfig{
		Providers: enabled,
	}

	router, err := router.ConfigureRouterStrategy(routerStrategy, &routerConfig)

	return router, err
}

func initializeCircuitBreakers(cfg *config.Config) map[string]*resilience.Circuit {
	enabled := cfg.GetEnabledProviders()
	resillienceConfig := cfg.GetResilienceConfigData()
	providers := make([]string, len(enabled))

	for _, provider := range enabled {
		providers = append(providers, provider.Name)
	}

	circuit := resilience.NewCircuitBreakers(providers, resillienceConfig.CircuitBreakerConfig)
	return circuit
}
