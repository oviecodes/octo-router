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
