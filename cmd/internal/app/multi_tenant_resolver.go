package app

import (
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"
	"llm-router/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MultiTenantResolver struct {
	Logger *zap.Logger
}

func (m *MultiTenantResolver) GetConfig(c *gin.Context) *config.Config {
	// Extract API key from request
	// apiKey := c.GetHeader("X-API-Key")
	// Fetch tenant config from DB/cache
	// return fetchTenantConfig(apiKey)
	return nil
}

// When I implement multi-tenancy
// it might not be initializeRouter,
// it might be some other function that checks
// if a router is already cached for said user, then retrieve
// if not fetch cfg from database and configure router properly
func (m *MultiTenantResolver) GetRouter(c *gin.Context) router.Router {
	// cfg := m.GetConfig(c)
	// router, _ := initializeRouter(cfg)
	// return router

	return nil
}

func (m *MultiTenantResolver) GetLogger(c *gin.Context) *zap.Logger {
	return m.Logger
}

func (m *MultiTenantResolver) GetCache(c *gin.Context) cache.Cache {
	return nil
}

func (m *MultiTenantResolver) GetRetry(c *gin.Context) *resilience.Retry {
	return nil
}

func (m *MultiTenantResolver) GetCircuitBreaker(c *gin.Context) map[string]types.CircuitBreaker {
	return nil
}
