package server

import (
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/endpoints"
	"llm-router/cmd/internal/metrics"
	"llm-router/cmd/internal/middleware"
	"llm-router/utils"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger = utils.SetUpLogger()

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Observe(duration)
	}
}

func Server() {
	// decide start up type
	var resolver app.ConfigResolver

	if os.Getenv("MULTI_TENANT") == "true" {
		resolver = &app.MultiTenantResolver{Logger: logger}
	} else {
		singleTenant, err := app.SetUpApp()
		if err != nil {
			logger.Error("Failed to initialize app", zap.Error(err))
			os.Exit(1)
		}
		resolverInstance := &app.SingleTenantResolver{}
		resolverInstance.App.Store(singleTenant)
		resolver = resolverInstance
	}

	ginRouter := gin.Default()
	ginRouter.Use(MetricsMiddleware())

	if config := resolver.GetConfig(); config != nil && len(config.Security.APIKeys) > 0 {
		ginRouter.Use(middleware.APIKeyAuth(config.Security.APIKeys))
	}

	endpoints.SetUpRoutes(resolver, ginRouter)

	// Same handler works for both modes
	// multi-tenant reasoning: all the same functions only difference is how config is gotten,
	// validated per request API key, user config fetched and used on per request basis - normal web app standard

	logger.Info("Starting server on localhost:8000")
	ginRouter.Run("localhost:8000")
}
