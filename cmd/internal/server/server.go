package server

import (
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/endpoints"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/utils"
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

var logger = utils.SetUpLogger()

func Server() {
	// decide start up type
	var resolver app.ConfigResolver

	if os.Getenv("MULTI_TENANT") == "true" {
		resolver = &app.MultiTenantResolver{Logger: logger}
	} else {
		singleTenant := app.SetUpApp()
		resolver = &app.SingleTenantResolver{App: singleTenant}
	}

	ginRouter := gin.Default()

	endpoints.SetUpRoutes(resolver, ginRouter)

	// Same handler works for both modes!
	// multi-tenant reasoning: all the same functions only difference is how config is gotten,
	// validated per request API key, user config fetched and used on per request basis - normal web app standard

	logger.Info("Starting server on localhost:8000")
	ginRouter.Run("localhost:8000")
}
