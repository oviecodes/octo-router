package providers

import (
	"fmt"
	"strings"
	"sync"

	"llm-router/types"

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
	ModelGeminiPro3        = "gemini/gemini-3-pro"
	ModelGeminiFlash3      = "gemini/gemini-3-flash"
	ModelGeminiPro25       = "gemini/gemini-2.5-pro"
	ModelGeminiFlash25     = "gemini/gemini-2.5-flash"
	ModelGeminiFlash25Lite = "gemini/gemini-2.5-flash-lite"
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
	InputCostPer1M  float64
	OutputCostPer1M float64
	ContextWindow   int
	Tier            ModelTier
}

var (
	modelRegistry = make(map[string]ModelInfo)
	registryMu    sync.RWMutex
)

func InitializeModelRegistry(defaults []types.ModelConfig, overrides []types.ModelConfig) {
	registryMu.Lock()
	defer registryMu.Unlock()

	// clear existing
	modelRegistry = make(map[string]ModelInfo)

	// Load defaults first
	for _, cfg := range defaults {
		addToRegistry(cfg)
	}

	// Apply overrides (will overwrite existing IDs)
	for _, cfg := range overrides {
		addToRegistry(cfg)
	}
}

func addToRegistry(cfg types.ModelConfig) {
	info := ModelInfo{
		ID:              cfg.ID,
		Provider:        cfg.Provider,
		Name:            cfg.Name,
		InputCostPer1M:  cfg.InputCostPer1M,
		OutputCostPer1M: cfg.OutputCostPer1M,
		ContextWindow:   cfg.ContextWindow,
		Tier:            ModelTier(cfg.Tier),
	}
	modelRegistry[cfg.ID] = info
}

func ParseModelID(modelID string) (provider, model string, err error) {
	parts := strings.SplitN(modelID, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model ID format: %s (expected provider/model)", modelID)
	}
	return parts[0], parts[1], nil
}

func GetModelInfo(modelID string) (ModelInfo, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	info, exists := modelRegistry[modelID]
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
	case ModelGeminiPro3:
		return "gemini-3-pro-preview", nil
	case ModelGeminiFlash3:
		return "gemini-3-flash-preview", nil
	case ModelGeminiPro25:
		return "gemini-2.5-pro", nil
	case ModelGeminiFlash25:
		return "gemini-2.5-flash", nil
	case ModelGeminiFlash25Lite:
		return "gemini-2.5-flash-lite", nil
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
	registryMu.RLock()
	defer registryMu.RUnlock()

	var models []ModelInfo
	for _, model := range modelRegistry {
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
	registryMu.RLock()
	defer registryMu.RUnlock()

	var models []ModelInfo
	for _, model := range modelRegistry {
		if model.Tier == tier {
			models = append(models, model)
		}
	}
	return models
}

func ListModelsByProviderAndTier(providerName string, tier ModelTier) []ModelInfo {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var models []ModelInfo
	for _, model := range modelRegistry {
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
