package handlers

import (
	"llm-router/cmd/internal/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(resolver app.ConfigResolver, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"providers": len(resolver.GetConfig().GetEnabledProviders()),
	})
}
