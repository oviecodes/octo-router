package handlers

import (
	"llm-router/cmd/internal/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminConfig(resolver app.ConfigResolver, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": resolver.GetConfig().Providers,
	})
}

func AdminProviders(resolver app.ConfigResolver, c *gin.Context) {
	enabled := resolver.GetConfig().GetEnabledProviders()

	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
		"count":   len(enabled),
	})
}
