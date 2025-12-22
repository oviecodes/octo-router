package app

import (
	"llm-router/cmd/internal/cache"
	"llm-router/cmd/internal/resilience"
	"llm-router/cmd/internal/router"
	"llm-router/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SingleTenantResolver struct {
	App *App
}

func (s *SingleTenantResolver) GetConfig(c *gin.Context) *config.Config {
	return s.App.Config
}

func (s *SingleTenantResolver) GetRouter(c *gin.Context) router.Router {
	return s.App.Router
}

func (s *SingleTenantResolver) GetLogger(c *gin.Context) *zap.Logger {
	return s.App.Logger
}

func (s *SingleTenantResolver) GetCache(c *gin.Context) cache.Cache {
	return s.App.Cache
}

func (s *SingleTenantResolver) GetRetry(c *gin.Context) *resilience.Retry {
	return s.App.Retry
}

func (s *SingleTenantResolver) GetCircuitBreaker(c *gin.Context, name string) *resilience.Circuit {
	return s.App.Circuit[name]
}
