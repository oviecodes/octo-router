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

	ginRouter.GET("/admin/status", func(c *gin.Context) {
		handlers.GetSystemStatus(resolver, c)
	})

	ginRouter.POST("/admin/budgets/reset", func(c *gin.Context) {
		handlers.ResetBudget(resolver, c)
	})

	ginRouter.POST("/admin/config/reload", func(c *gin.Context) {
		handlers.ReloadConfig(resolver, c)
	})
}
