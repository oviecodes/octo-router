package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Server() {
	router := gin.Default()
	router.GET("/health", health)

	router.Run("localhost:8000")
}

func health(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Healthy Endpoint Reched"})
}
