package router

import (
	"context"
	"fmt"
	"llm-router/cmd/internal/providers"
	"llm-router/types"
	"math"
	"sync"

	"go.uber.org/zap"
)

type CostRouter struct {
	providerManager *providers.ProviderManager
	costOptions     *types.CostOptions
	budgetManager   *BudgetManager
	mu              sync.RWMutex
}

func (c *CostRouter) GetBudgetManager() *BudgetManager {
	return c.budgetManager
}

func NewCostRouter(providerManager *providers.ProviderManager, costOptions *types.CostOptions, budget *BudgetManager) (*CostRouter, error) {
	if providerManager == nil {
		return nil, fmt.Errorf("provider manager cannot be nil")
	}

	if providerManager.GetProviderCount() == 0 {
		return nil, fmt.Errorf("provider manager has no providers")
	}

	if costOptions == nil {
		costOptions = &types.CostOptions{
			DefaultTier:  "",
			MinimumTier:  "",
			TierStrategy: "same-tier",
		}
	}

	logger.Info("Cost-based router initialized",
		zap.Int("provider_count", providerManager.GetProviderCount()),
		zap.Strings("providers", providerManager.ListProviderNames()),
		zap.String("default_tier", costOptions.DefaultTier),
		zap.String("minimum_tier", costOptions.MinimumTier),
	)

	return &CostRouter{
		providerManager: providerManager,
		costOptions:     costOptions,
		budgetManager:   budget,
	}, nil
}

func (c *CostRouter) SelectProvider(ctx context.Context, deps *types.SelectProviderInput) (*types.SelectedProviderOutput, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tierConstraint := deps.Tier
	if tierConstraint == "" && c.costOptions.DefaultTier != "" {
		tierConstraint = c.costOptions.DefaultTier
	}

	cheapestModel, cheapestProvider, err := c.findCheapestModel(ctx, deps.Messages, deps.Circuits, tierConstraint)
	if err != nil {
		return nil, err
	}

	logger.Info("Cost-based router selected provider",
		zap.String("provider", cheapestProvider.GetProviderName()),
		zap.String("model", cheapestModel.ID),
		zap.String("tier", string(cheapestModel.Tier)),
		zap.String("requested_tier", tierConstraint),
		zap.Float64("input_cost_per_1m", cheapestModel.InputCostPer1M),
		zap.Float64("output_cost_per_1m", cheapestModel.OutputCostPer1M),
	)

	return &types.SelectedProviderOutput{
		Provider: cheapestProvider,
		Model:    cheapestModel.ID,
	}, nil
}

func (c *CostRouter) findCheapestModel(
	ctx context.Context,
	messages []types.Message,
	circuits map[string]types.CircuitBreaker,
	tierConstraint string,
) (providers.ModelInfo, types.Provider, error) {
	allProviders := c.providerManager.GetProviders()

	var cheapestModel providers.ModelInfo
	var selectedProvider types.Provider
	cheapestPrice := math.Inf(1)

	for _, provider := range allProviders {
		providerName := provider.GetProviderName()

		if circuit, exists := circuits[providerName]; exists {
			if circuit.GetState() == "open" {
				logger.Debug("Skipping provider with open circuit",
					zap.String("provider", providerName),
				)
				continue
			}
		}

		tokens, err := provider.CountTokens(ctx, messages)
		if err != nil {
			logger.Warn("Failed to count tokens for provider",
				zap.String("provider", providerName),
				zap.Error(err),
			)
			continue
		}

		var modelsToCheck []providers.ModelInfo
		if tierConstraint != "" {

			modelsToCheck = providers.ListModelsByProviderAndTier(providerName, providers.ModelTier(tierConstraint))
			if len(modelsToCheck) == 0 {
				logger.Debug("No models found for provider in requested tier",
					zap.String("provider", providerName),
					zap.String("tier", tierConstraint),
				)
				continue
			}
		} else {

			allModels := providers.ListModelsByProvider(providerName)
			if c.costOptions.MinimumTier != "" {
				modelsToCheck = c.filterByMinimumTier(allModels, providers.ModelTier(c.costOptions.MinimumTier))
			} else {
				modelsToCheck = allModels
			}
		}

		for _, model := range modelsToCheck {

			cost, err := providers.CalculateCost(model.ID, tokens, tokens)
			if err != nil {
				continue
			}

			if cost < cheapestPrice {
				cheapestPrice = cost
				cheapestModel = model
				selectedProvider = provider
			}
		}
	}

	if selectedProvider == nil {
		if tierConstraint != "" {
			return providers.ModelInfo{}, nil, fmt.Errorf("no available providers in tier: %s", tierConstraint)
		}
		return providers.ModelInfo{}, nil, fmt.Errorf("no available providers")
	}

	return cheapestModel, selectedProvider, nil
}

func (c *CostRouter) filterByMinimumTier(models []providers.ModelInfo, minimumTier providers.ModelTier) []providers.ModelInfo {
	tierOrder := map[providers.ModelTier]int{
		providers.TierBudget:       1,
		providers.TierStandard:     2,
		providers.TierPremium:      3,
		providers.TierUltraPremium: 4,
	}

	minTierLevel := tierOrder[minimumTier]
	var filtered []providers.ModelInfo

	for _, model := range models {
		if tierOrder[model.Tier] >= minTierLevel {
			filtered = append(filtered, model)
		}
	}

	return filtered
}

func (c *CostRouter) GetProviderManager() *providers.ProviderManager {
	return c.providerManager
}
