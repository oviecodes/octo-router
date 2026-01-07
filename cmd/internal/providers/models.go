package providers

import (
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go/v3"
)

const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
)

const (
	ModelOpenAIGPT5       = "openai/gpt-5"
	ModelOpenAIGPT51      = "openai/gpt-5.1"
	ModelOpenAIGPT4o      = "openai/gpt-4o"
	ModelOpenAIGPT4oMini  = "openai/gpt-4o-mini"
	ModelOpenAIGPT35Turbo = "openai/gpt-3.5-turbo"
)

const (
	ModelAnthropicOpus45  = "anthropic/claude-opus-4.5"
	ModelAnthropicSonnet4 = "anthropic/claude-sonnet-4"
	ModelAnthropicHaiku45 = "anthropic/claude-haiku-4.5"
	ModelAnthropicHaiku3  = "anthropic/claude-haiku-3"
)

const (
	ModelGeminiFlash25     = "gemini/gemini-2.5-flash"
	ModelGeminiFlash25Lite = "gemini/gemini-2.5-flash-lite"
	ModelGeminiPro20       = "gemini/gemini-2.0-pro"
)

type ModelTier string

const (
	TierUltraPremium ModelTier = "ultra-premium"
	TierPremium      ModelTier = "premium"
	TierStandard     ModelTier = "standard"
	TierBudget       ModelTier = "budget"
)

type ModelInfo struct {
	ID              string
	Provider        string
	Name            string
	InputCostPer1M  float64 // USD per 1M input tokens
	OutputCostPer1M float64 // USD per 1M output tokens
	ContextWindow   int     // Max context tokens
	Tier            ModelTier
}

// Models catalog with pricing and metadata
var Models = map[string]ModelInfo{
	// OpenAI Models - Ultra-Premium Tier
	ModelOpenAIGPT51: {
		ID:              ModelOpenAIGPT51,
		Provider:        ProviderOpenAI,
		Name:            "GPT-5.1",
		InputCostPer1M:  7.50,
		OutputCostPer1M: 22.50,
		ContextWindow:   200000,
		Tier:            TierUltraPremium,
	},
	ModelOpenAIGPT5: {
		ID:              ModelOpenAIGPT5,
		Provider:        ProviderOpenAI,
		Name:            "GPT-5",
		InputCostPer1M:  5.00,
		OutputCostPer1M: 15.00,
		ContextWindow:   200000,
		Tier:            TierUltraPremium,
	},

	// OpenAI Models - Premium Tier
	ModelOpenAIGPT4o: {
		ID:              ModelOpenAIGPT4o,
		Provider:        ProviderOpenAI,
		Name:            "GPT-4o",
		InputCostPer1M:  2.50,
		OutputCostPer1M: 10.00,
		ContextWindow:   128000,
		Tier:            TierPremium,
	},

	// OpenAI Models - Standard Tier
	ModelOpenAIGPT35Turbo: {
		ID:              ModelOpenAIGPT35Turbo,
		Provider:        ProviderOpenAI,
		Name:            "GPT-3.5 Turbo",
		InputCostPer1M:  0.50,
		OutputCostPer1M: 1.50,
		ContextWindow:   16385,
		Tier:            TierStandard,
	},

	// OpenAI Models - Budget Tier
	ModelOpenAIGPT4oMini: {
		ID:              ModelOpenAIGPT4oMini,
		Provider:        ProviderOpenAI,
		Name:            "GPT-4o Mini",
		InputCostPer1M:  0.15,
		OutputCostPer1M: 0.60,
		ContextWindow:   128000,
		Tier:            TierBudget,
	},

	// Anthropic Models - Ultra-Premium Tier
	ModelAnthropicOpus45: {
		ID:              ModelAnthropicOpus45,
		Provider:        ProviderAnthropic,
		Name:            "Claude Opus 4.5",
		InputCostPer1M:  15.00,
		OutputCostPer1M: 75.00,
		ContextWindow:   200000,
		Tier:            TierUltraPremium,
	},

	// Anthropic Models - Premium Tier
	ModelAnthropicSonnet4: {
		ID:              ModelAnthropicSonnet4,
		Provider:        ProviderAnthropic,
		Name:            "Claude Sonnet 4",
		InputCostPer1M:  3.00,
		OutputCostPer1M: 15.00,
		ContextWindow:   200000,
		Tier:            TierPremium,
	},

	// Anthropic Models - Standard Tier
	ModelAnthropicHaiku45: {
		ID:              ModelAnthropicHaiku45,
		Provider:        ProviderAnthropic,
		Name:            "Claude Haiku 4.5",
		InputCostPer1M:  0.80,
		OutputCostPer1M: 4.00,
		ContextWindow:   200000,
		Tier:            TierStandard,
	},
	ModelAnthropicHaiku3: {
		ID:              ModelAnthropicHaiku3,
		Provider:        ProviderAnthropic,
		Name:            "Claude Haiku 3",
		InputCostPer1M:  0.25,
		OutputCostPer1M: 1.25,
		ContextWindow:   200000,
		Tier:            TierStandard,
	},

	// Gemini Models - Premium Tier
	ModelGeminiPro20: {
		ID:              ModelGeminiPro20,
		Provider:        ProviderGemini,
		Name:            "Gemini 2.0 Pro",
		InputCostPer1M:  1.25,
		OutputCostPer1M: 5.00,
		ContextWindow:   2000000,
		Tier:            TierPremium,
	},

	// Gemini Models - Budget Tier
	ModelGeminiFlash25: {
		ID:              ModelGeminiFlash25,
		Provider:        ProviderGemini,
		Name:            "Gemini 2.5 Flash",
		InputCostPer1M:  0.075,
		OutputCostPer1M: 0.30,
		ContextWindow:   1000000,
		Tier:            TierBudget,
	},
	ModelGeminiFlash25Lite: {
		ID:              ModelGeminiFlash25Lite,
		Provider:        ProviderGemini,
		Name:            "Gemini 2.5 Flash Lite",
		InputCostPer1M:  0.0375,
		OutputCostPer1M: 0.15,
		ContextWindow:   1000000,
		Tier:            TierBudget,
	},
}

func ParseModelID(modelID string) (provider, model string, err error) {
	parts := strings.SplitN(modelID, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model ID format: %s (expected provider/model)", modelID)
	}
	return parts[0], parts[1], nil
}

func GetModelInfo(modelID string) (ModelInfo, error) {
	info, exists := Models[modelID]
	if !exists {
		return ModelInfo{}, fmt.Errorf("unknown model: %s", modelID)
	}
	return info, nil
}

// MapToOpenAIModel maps standardized model ID to OpenAI SDK model
func MapToOpenAIModel(modelID string) (string, error) {
	switch modelID {
	case ModelOpenAIGPT5:
		return openai.ChatModelGPT5_2025_08_07, nil
	case ModelOpenAIGPT51:
		return openai.ChatModelGPT5_1ChatLatest, nil
	case ModelOpenAIGPT4o:
		return openai.ChatModelChatgpt4oLatest, nil
	case ModelOpenAIGPT4oMini:
		return openai.ChatModelGPT4oMini, nil
	case ModelOpenAIGPT35Turbo:
		return openai.ChatModelGPT3_5Turbo, nil
	default:
		return "", fmt.Errorf("unknown OpenAI model: %s", modelID)
	}
}

func MapToAnthropicModel(modelID string) (anthropic.Model, error) {
	switch modelID {
	case ModelAnthropicOpus45:
		return anthropic.ModelClaudeOpus4_5_20251101, nil
	case ModelAnthropicSonnet4:
		return anthropic.ModelClaude4Sonnet20250514, nil
	case ModelAnthropicHaiku45:
		return anthropic.ModelClaudeHaiku4_5_20251001, nil
	case ModelAnthropicHaiku3:
		return anthropic.ModelClaude_3_Haiku_20240307, nil
	default:
		return "", fmt.Errorf("unknown Anthropic model: %s", modelID)
	}
}

func MapToGeminiModel(modelID string) (string, error) {
	switch modelID {
	case ModelGeminiFlash25:
		return "gemini-2.5-flash", nil
	case ModelGeminiFlash25Lite:
		return "gemini-2.5-flash-lite", nil
	case ModelGeminiPro20:
		return "gemini-2.0-pro", nil
	default:
		return "", fmt.Errorf("unknown Gemini model: %s", modelID)
	}
}

func CalculateCost(modelID string, inputTokens, outputTokens int) (float64, error) {
	info, err := GetModelInfo(modelID)
	if err != nil {
		return 0, err
	}

	inputCost := float64(inputTokens) * info.InputCostPer1M / 1_000_000
	outputCost := float64(outputTokens) * info.OutputCostPer1M / 1_000_000

	return inputCost + outputCost, nil
}

func ListModelsByProvider(providerName string) []ModelInfo {
	var models []ModelInfo
	for _, model := range Models {
		if model.Provider == providerName {
			models = append(models, model)
		}
	}
	return models
}

func ValidateModelID(modelID string) (string, error) {
	provider, _, err := ParseModelID(modelID)
	if err != nil {
		return "", err
	}

	if _, err := GetModelInfo(modelID); err != nil {
		return "", err
	}

	return provider, nil
}

func ListModelsByTier(tier ModelTier) []ModelInfo {
	var models []ModelInfo
	for _, model := range Models {
		if model.Tier == tier {
			models = append(models, model)
		}
	}
	return models
}

func ListModelsByProviderAndTier(providerName string, tier ModelTier) []ModelInfo {
	var models []ModelInfo
	for _, model := range Models {
		if model.Provider == providerName && model.Tier == tier {
			models = append(models, model)
		}
	}
	return models
}

func FindCheapestModel(models []ModelInfo) (ModelInfo, error) {
	if len(models) == 0 {
		return ModelInfo{}, fmt.Errorf("no models provided")
	}

	cheapest := models[0]
	cheapestAvgCost := (cheapest.InputCostPer1M + cheapest.OutputCostPer1M) / 2

	for _, model := range models[1:] {
		avgCost := (model.InputCostPer1M + model.OutputCostPer1M) / 2
		if avgCost < cheapestAvgCost {
			cheapest = model
			cheapestAvgCost = avgCost
		}
	}

	return cheapest, nil
}
