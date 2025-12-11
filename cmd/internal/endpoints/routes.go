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

	ginRouter.POST("/admin/config", func(c *gin.Context) {
		handlers.AdminConfig(resolver, c)
	})
	ginRouter.POST("/admin/providers", func(c *gin.Context) {
		handlers.AdminProviders(resolver, c)
	})

}
