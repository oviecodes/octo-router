package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"llm-router/utils"

	"go.uber.org/zap"
)

type Router interface {
	SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error)
	GetProviderManager() *providers.ProviderManager
}

var logger = utils.SetUpLogger()

func ConfigureRouterStrategy(routingData *types.RoutingData, providerManager *providers.ProviderManager) (Router, []string, error) {

	var routerStrategy Router

	switch routingData.Strategy {
	case "round-robin":
		router, err := NewRoundRobinRouter(providerManager)

		if err != nil {
			logger.Error("Could not set up the round-robin router", zap.Error(err))
			return nil, nil, err
		}

		routerStrategy = router
	default:
		return nil, nil, fmt.Errorf("unsupported routing strategy: %s (supported: round-robin)", routingData.Strategy)
	}

	return routerStrategy, routingData.Fallbacks, nil
}
