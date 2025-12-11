package server

import (
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/endpoints"
	"llm-router/utils"
	"os"

	"github.com/gin-gonic/gin"
)

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
