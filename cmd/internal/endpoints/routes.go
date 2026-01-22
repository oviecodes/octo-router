package endpoints

import (
	"llm-router/cmd/internal/app"
	"llm-router/cmd/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetUpRoutes(resolver app.ConfigResolver, ginRouter *gin.Engine) {

	ginRouter.GET("/health", func(c *gin.Context) {
		handlers.Health(resolver, c)
	})

	ginRouter.POST("/v1/chat/completions", func(c *gin.Context) {
		handlers.Completions(resolver, c)
	})

	ginRouter.GET("/admin/usage", func(c *gin.Context) {
		handlers.GetUsageHistory(resolver, c)
	})
}
