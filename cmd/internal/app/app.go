package app

import (
	"fmt"
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
	GetRouter(c *gin.Context) *router.Router
	GetLogger(c *gin.Context) *zap.Logger
}

type App struct {
	Config *config.Config
	Router *router.Router
	Logger *zap.Logger
}

type SingleTenantResolver struct {
	App *App
}

type MultiTenantResolver struct {
	Logger *zap.Logger
}

func (m *MultiTenantResolver) GetConfig(c *gin.Context) *config.Config {
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
func (m *MultiTenantResolver) GetRouter(c *gin.Context) *router.Router {
	cfg := m.GetConfig(c)
	router, _ := initializeRouter(cfg)
	return router
}

func (m *MultiTenantResolver) GetLogger(c *gin.Context) *zap.Logger {
	return m.Logger
}

func (s *SingleTenantResolver) GetConfig(c *gin.Context) *config.Config {
	return s.App.Config
}

func (s *SingleTenantResolver) GetRouter(c *gin.Context) *router.Router {
	return s.App.Router
}

func (s *SingleTenantResolver) GetLogger(c *gin.Context) *zap.Logger {
	return s.App.Logger
}

var logger = utils.SetUpLogger()

func SetUpApp() *App {
	defer logger.Sync()

	// logger.Info("This is an info message", zap.String("key", "value"), zap.Int("number", 123))

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

	// Create app with all dependencies
	app := &App{
		Config: cfg,
		Router: llmRouter,
		Logger: logger,
	}

	return app
}

func initializeRouter(cfg *config.Config) (*router.Router, error) {
	enabled := cfg.GetEnabledProviders()
	routerStrategy := cfg.GetRouterStrategy()

	fmt.Printf("All Routing configs %v \n", cfg.GetRouterStrategy())

	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled providers found in config")
	}

	routerConfig := types.RouterConfig{
		Providers: enabled,
	}

	router, err := router.ConfigureRouterStrategy(routerStrategy, &routerConfig)

	return &router, err
}
