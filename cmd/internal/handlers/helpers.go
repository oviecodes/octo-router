package handlers

import (
	"llm-router/cmd/internal/providers"
	"llm-router/types"

	"go.uber.org/zap"
)

func buildProviderChain(primaryProvider types.Provider, fallbackNames []string, manager *providers.ProviderManager, candidates []types.Provider) []types.Provider {
	providerChain := make([]types.Provider, 0, len(fallbackNames)+1)
	seen := make(map[string]bool)

	// Create a map for fast candidate lookup if candidates are provided
	allowed := make(map[string]bool)
	useFilter := len(candidates) > 0
	if useFilter {
		for _, c := range candidates {
			allowed[c.GetProviderName()] = true
		}
	}

	primaryName := primaryProvider.GetProviderName()
	providerChain = append(providerChain, primaryProvider)
	seen[primaryName] = true

	for _, fallbackName := range fallbackNames {
		if seen[fallbackName] {
			continue
		}

		if useFilter && !allowed[fallbackName] {
			continue
		}

		fallbackProvider, err := manager.GetProvider(fallbackName)
		if err != nil {
			continue
		}

		providerChain = append(providerChain, fallbackProvider)
		seen[fallbackName] = true
	}

	return providerChain
}

func buildProviderChainWithModels(
	primaryModel string,
	primaryProvider types.Provider,
	fallbackNames []string,
	manager *providers.ProviderManager,
	candidates []types.Provider,
	logger *zap.Logger,
) []types.ProviderWithModel {
	chain := make([]types.ProviderWithModel, 0, len(fallbackNames)+1)
	seen := make(map[string]bool)

	// Create lookup table for allowed candidates
	allowed := make(map[string]bool)
	useFilter := len(candidates) > 0
	if useFilter {
		for _, c := range candidates {
			allowed[c.GetProviderName()] = true
		}
	}

	primaryModelInfo, err := providers.GetModelInfo(primaryModel)
	if err != nil {
		logger.Warn("Failed to get primary model info, building simple chain",
			zap.String("model", primaryModel),
			zap.Error(err),
		)

		return buildSimpleChainWithModels(primaryModel, primaryProvider, fallbackNames, manager, candidates)
	}

	primaryTier := primaryModelInfo.Tier

	primaryName := primaryProvider.GetProviderName()
	chain = append(chain, types.ProviderWithModel{
		Provider: primaryProvider,
		Model:    primaryModel,
	})
	seen[primaryName] = true

	logger.Debug("Building tier-aware fallback chain",
		zap.String("primary_tier", string(primaryTier)),
		zap.String("primary_model", primaryModel),
	)

	for _, fallbackName := range fallbackNames {
		if seen[fallbackName] {
			continue
		}

		if useFilter && !allowed[fallbackName] {
			continue
		}

		fallbackProvider, err := manager.GetProvider(fallbackName)
		if err != nil {
			continue
		}

		models := providers.ListModelsByProviderAndTier(fallbackName, primaryTier)
		if len(models) == 0 {
			logger.Debug("No models in tier for provider, skipping",
				zap.String("provider", fallbackName),
				zap.String("tier", string(primaryTier)),
			)
			continue
		}

		cheapestModel, err := providers.FindCheapestModel(models)
		if err != nil {
			continue
		}

		logger.Debug("Adding fallback provider",
			zap.String("provider", fallbackName),
			zap.String("model", cheapestModel.ID),
			zap.String("tier", string(cheapestModel.Tier)),
		)

		chain = append(chain, types.ProviderWithModel{
			Provider: fallbackProvider,
			Model:    cheapestModel.ID,
		})
		seen[fallbackName] = true
	}

	return chain
}

func buildSimpleChainWithModels(
	primaryModel string,
	primaryProvider types.Provider,
	fallbackNames []string,
	manager *providers.ProviderManager,
	candidates []types.Provider,
) []types.ProviderWithModel {
	chain := make([]types.ProviderWithModel, 0, len(fallbackNames)+1)
	seen := make(map[string]bool)

	// Create lookup table for allowed candidates
	allowed := make(map[string]bool)
	useFilter := len(candidates) > 0
	if useFilter {
		for _, c := range candidates {
			allowed[c.GetProviderName()] = true
		}
	}

	primaryName := primaryProvider.GetProviderName()
	chain = append(chain, types.ProviderWithModel{
		Provider: primaryProvider,
		Model:    primaryModel,
	})
	seen[primaryName] = true

	for _, fallbackName := range fallbackNames {
		if seen[fallbackName] {
			continue
		}

		if useFilter && !allowed[fallbackName] {
			continue
		}

		fallbackProvider, err := manager.GetProvider(fallbackName)
		if err != nil {
			continue
		}

		models := providers.ListModelsByProvider(fallbackName)
		if len(models) == 0 {
			continue
		}

		cheapestModel, err := providers.FindCheapestModel(models)
		if err != nil {
			continue
		}

		chain = append(chain, types.ProviderWithModel{
			Provider: fallbackProvider,
			Model:    cheapestModel.ID,
		})
		seen[fallbackName] = true
	}

	return chain
}
