package router

import (
	"context"
	"fmt"
	"llm-router/types"
	"llm-router/utils"

	"go.uber.org/zap"
)

type Router interface {
	SelectProvider(ctx context.Context, circuits map[string]types.CircuitBreaker) (types.Provider, error)
}

var logger = utils.SetUpLogger()

func ConfigureRouterStrategy(routingData *types.RoutingData, config *types.RouterConfig) (Router, error) {
	switch routingData.Strategy {
	case "round-robin":
		router, err := NewRoundRobinRouter(*config)

		if err != nil {
			logger.Error("Could not set up the round-robin router %v", zap.Error(err))
			return nil, err
		}

		return router, nil
	default:
		return nil, fmt.Errorf("unsupported routing strategy: %s (supported: round-robin)", routingData.Strategy)
	}
}
