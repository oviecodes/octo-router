package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/cmd/internal/router/filters"
	"llm-router/types"
	"llm-router/utils"

	"go.uber.org/zap"
)

type Router interface {
	SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error)
	GetProviderManager() *providers.ProviderManager
}

var logger = utils.SetUpLogger()

func ConfigureRouterStrategy(routingData *types.RoutingData, providerManager *providers.ProviderManager, tracker *LatencyTracker) (Router, []string, error) {

	var routerStrategy Router
	var err error

	switch routingData.Strategy {
	case "round-robin":
		routerStrategy, err = NewRoundRobinRouter(providerManager)
		if err != nil {
			logger.Error("Could not set up the round-robin router", zap.Error(err))
			return nil, nil, err
		}

	case "cost-based":
		routerStrategy, err = NewCostRouter(providerManager, routingData.CostOptions)
		if err != nil {
			logger.Error("Could not set up the cost-based router", zap.Error(err))
			return nil, nil, err
		}

	case "latency-based":
		routerStrategy, err = NewLatencyRouter(providerManager, tracker)
		if err != nil {
			logger.Error("Could not set up the latency-based router", zap.Error(err))
			return nil, nil, err
		}

	case "weighted":
		routerStrategy, err = NewWeightedRouter(providerManager, routingData.Weights)
		if err != nil {
			logger.Error("Could not set up the weighted router", zap.Error(err))
			return nil, nil, err
		}

	default:
		return nil, nil, fmt.Errorf("unsupported routing strategy: %s (supported: round-robin, cost-based, latency-based, weighted)", routingData.Strategy)
	}

	if routingData.Policies != nil {
		pipeline := NewPipelineRouter(routerStrategy, providerManager)

		if routingData.Policies.Semantic != nil && routingData.Policies.Semantic.Enabled {
			var semanticFilter ProviderFilter
			var filterErr error

			if routingData.Policies.Semantic.Engine == "embedding" {
				semanticFilter, filterErr = filters.NewEmbeddingFilter(routingData.Policies.Semantic)
				if filterErr != nil {
					logger.Error("Could not set up embedding filter, falling back to keywords", zap.Error(filterErr))
					semanticFilter = filters.NewKeywordFilter(routingData.Policies.Semantic)
				} else {
					logger.Info("Enabled Semantic (Embedding) Filter in Routing Pipeline")
				}
			} else {
				semanticFilter = filters.NewKeywordFilter(routingData.Policies.Semantic)
				logger.Info("Enabled Semantic (Keyword) Filter in Routing Pipeline")
			}

			pipeline.AddFilter(semanticFilter)
		}

		// If filters were added, use pipeline as the strategy.
		// Even if no specific filter added (e.g. semantic disabled), wrapping in pipeline is harmless or we can skip.
		// For now, let's wrap it to be safe if policies struct exists.
		routerStrategy = pipeline
	}

	return routerStrategy, routingData.Fallbacks, nil
}
