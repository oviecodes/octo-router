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
	// fmt.Printf("enabled providers %v", enabled)
	modelData := cfg.GetModelData()

	if len(enabled) == 0 {
		return nil, fmt.Errorf("no enabled providers found in config")
	}

	routerConfig := types.RouterConfig{
		Providers: enabled,
		MaxTokens: int64(modelData.MaxToken),
		Model:     modelData.Model,
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
	var request types.Completion

	// Bind and validate JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Additional business logic validation
	if err := app.validateCompletionRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log the request
	app.Logger.Info("Completion request received",
		zap.Int("message_count", len(request.Messages)),
		zap.String("model", request.Model),
	)

	// Select provider
	provider := app.Router.SelectProvider(c.Request.Context())

	// TODO: Call provider and return response
	response, err := provider.Complete(c.Request.Context(), request.Messages)
	if err != nil {
		app.Logger.Error("Provider completion failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate completion",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  response.Content,
		"role":     response.Role,
		"provider": fmt.Sprintf("%T", provider),
	})
}

// validateCompletionRequest performs additional business logic validation
func (app *App) validateCompletionRequest(req *types.Completion) error {

	if len(req.Messages) > 0 {
		// First message should typically be from user
		if req.Messages[0].Role != "user" && req.Messages[0].Role != "system" {
			return fmt.Errorf("first message must be from user or system")
		}
	}

	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2")
		}
	}

	// Check total message length (approximate)
	totalLength := 0
	for _, msg := range req.Messages {
		totalLength += len(msg.Content)
	}
	if totalLength > 1000000 { // 1MB limit
		return fmt.Errorf("total message content too large (max 1MB)")
	}

	return nil
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
