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
	GetBudgetManager() BudgetManager
}

var logger = utils.SetUpLogger()

func ConfigureRouterStrategy(
	routingData *types.RoutingData,
	providerManager *providers.ProviderManager,
	tracker *LatencyTracker,
	budgetManager BudgetManager,
) (Router, []string, error) {

	var routerStrategy Router
	var err error

	switch routingData.Strategy {
	case "round-robin":
		routerStrategy, err = NewRoundRobinRouter(providerManager, budgetManager)
		if err != nil {
			logger.Error("Could not set up the round-robin router", zap.Error(err))
			return nil, nil, err
		}

	case "cost-based":
		routerStrategy, err = NewCostRouter(providerManager, routingData.CostOptions, budgetManager)
		if err != nil {
			logger.Error("Could not set up the cost-based router", zap.Error(err))
			return nil, nil, err
		}

	case "latency-based":
		routerStrategy, err = NewLatencyRouter(providerManager, tracker, budgetManager)
		if err != nil {
			logger.Error("Could not set up the latency-based router", zap.Error(err))
			return nil, nil, err
		}

	case "weighted":
		routerStrategy, err = NewWeightedRouter(providerManager, routingData.Weights, budgetManager)
		if err != nil {
			logger.Error("Could not set up the weighted router", zap.Error(err))
			return nil, nil, err
		}

	default:
		return nil, nil, fmt.Errorf("unsupported routing strategy: %s (supported: round-robin, cost-based, latency-based, weighted)", routingData.Strategy)
	}

	pipeline := NewPipelineRouter(routerStrategy, providerManager, budgetManager)

	if budgetManager != nil {
		pipeline.AddFilter(filters.NewBudgetFilter(budgetManager, logger))
		logger.Info("Enabled Budget Filter in Routing Pipeline")
	}

	if routingData.Policies != nil && routingData.Policies.Semantic != nil && routingData.Policies.Semantic.Enabled {
		var semanticFilter ProviderFilter
		var filterErr error

		if routingData.Policies.Semantic.Engine == "embedding" {
			semanticFilter, filterErr = filters.NewEmbeddingFilter(routingData.Policies.Semantic, logger)
			if filterErr != nil {
				logger.Error("Could not set up embedding filter, falling back to keywords", zap.Error(filterErr))
				semanticFilter = filters.NewKeywordFilter(routingData.Policies.Semantic, logger)
			} else {
				logger.Info("Enabled Semantic (Embedding) Filter in Routing Pipeline")
			}
		} else {
			semanticFilter = filters.NewKeywordFilter(routingData.Policies.Semantic, logger)
			logger.Info("Enabled Semantic (Keyword) Filter in Routing Pipeline")
		}

		pipeline.AddFilter(semanticFilter)
	}

	routerStrategy = pipeline

	return routerStrategy, routingData.Fallbacks, nil
}
