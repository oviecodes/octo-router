package server

import (
	"fmt"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// App holds all dependencies that need to be injected into handlers
type App struct {
	Config *config.Config
	Router *router.RoundRobinRouter
	Logger *zap.Logger
}

func Server() {
	// Load config once at startup
	logger, _ := zap.NewDevelopment()
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

	ginRouter := gin.Default()

	ginRouter.GET("/health", app.health)
	ginRouter.POST("/v1/chat/completions", app.completions)

	ginRouter.POST("/admin/config", app.adminConfig)
	ginRouter.POST("/admin/providers", app.adminProviders)

	logger.Info("Starting server on localhost:8000")
	ginRouter.Run("localhost:8000")
}

// initializeRouter creates the LLM router with providers from config
func initializeRouter(cfg *config.Config) (*router.RoundRobinRouter, error) {
	enabled := cfg.GetEnabledProviders()
	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled providers found in config")
	}

	routerConfig := types.RouterConfig{
		Providers: enabled,
		MaxTokens: int64(cfg.MaxTokens),
		Model:     cfg.Model,
	}

	// Create router with config
	// round robin for now,
	// later (selectRouter) will determine what router type use based on config
	return router.NewRoundRobinRouter(routerConfig)
}

func (app *App) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"providers": len(app.Config.GetEnabledProviders()),
	})
}

func (app *App) completions(c *gin.Context) {
	provider := app.Router.SelectProvider(c.Request.Context())

	// TODO: Parse request, call provider, return response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider selected",
		"provider": fmt.Sprintf("%T", provider),
	})
}

func (app *App) adminConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": app.Config.Providers,
	})
}

func (app *App) adminProviders(c *gin.Context) {
	enabled := app.Config.GetEnabledProviders()
	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
		"count":   len(enabled),
	})
}
