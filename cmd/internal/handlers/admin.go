package handlers

import (
	"llm-router/cmd/internal/app"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUsageHistory(resolver app.ConfigResolver, c *gin.Context) {
	ctx := c.Request.Context()
	date := c.Query("date")

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	router := resolver.GetRouter()
	usageHistory := router.GetUsageHistoryManager()

	if usageHistory == nil {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Usage history is not enabled",
		})
		return
	}

	stats, err := usageHistory.GetDailyUsage(ctx, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch usage history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  date,
		"usage": stats,
	})
}

func ReloadConfig(resolver app.ConfigResolver, c *gin.Context) {
	err := resolver.Reload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reload configuration",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "Configuration reloaded successfully",
		"timestamp": time.Now(),
	})
}

func GetSystemStatus(resolver app.ConfigResolver, c *gin.Context) {
	cbStates := make(map[string]string)

	circuits := resolver.GetCircuitBreaker()
	for name, cb := range circuits {
		cbStates[name] = cb.GetState()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"routing": gin.H{
			"strategy": resolver.GetConfig().Routing.Strategy,
		},
		"circuit_breakers": cbStates,
		"timestamp":        time.Now(),
	})
}

func ResetBudget(resolver app.ConfigResolver, c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider query parameter is required"})
		return
	}

	budgetManager := resolver.GetRouter().GetBudgetManager()
	if budgetManager == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Budget management is not enabled"})
		return
	}

	budgetManager.ResetUsage(provider)

	c.JSON(http.StatusOK, gin.H{
		"status":   "Budget reset successfully",
		"provider": provider,
	})
}
