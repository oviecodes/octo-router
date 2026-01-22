package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKeys []string) gin.HandlerFunc {
	validKeys := make(map[string]bool)
	for _, key := range apiKeys {
		if key != "" {
			validKeys[key] = true
		}
	}

	return func(c *gin.Context) {
		if len(validKeys) == 0 {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected 'Bearer <token>'",
			})
			return
		}

		apiKey := parts[1]
		if !validKeys[apiKey] {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}

		c.Next()
	}
}
